package concurrent_processing

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
)

// ProcessingStage represents different stages of message processing
type ProcessingStage string

const (
	StageAntiflood  ProcessingStage = "antiflood"
	StageBlacklist  ProcessingStage = "blacklist"
	StageFilter     ProcessingStage = "filter"
	StagePermission ProcessingStage = "permission"
	StageValidation ProcessingStage = "validation"
)

// MessageProcessingJob represents a message processing job
type MessageProcessingJob struct {
	Bot     *gotgbot.Bot
	Context *ext.Context
	Stage   ProcessingStage
	UserID  int64
	ChatID  int64
}

// ProcessingResult represents the result of message processing
type ProcessingResult struct {
	Stage     ProcessingStage
	Success   bool
	Blocked   bool // Whether message should be blocked
	Error     error
	Duration  time.Duration
	UserID    int64
	ChatID    int64
	Timestamp time.Time
}

// ProcessorFunc defines the signature for processing functions
// The context parameter allows for cancellation and timeout handling
type ProcessorFunc func(ctx context.Context, bot *gotgbot.Bot, extCtx *ext.Context) (blocked bool, err error)

// MessageProcessingPipeline handles concurrent message processing
type MessageProcessingPipeline struct {
	stages         map[ProcessingStage]ProcessorFunc
	stagesLock     sync.RWMutex  // Protects stages map
	workers        int
	jobs           chan MessageProcessingJob
	results        chan ProcessingResult
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	statsLock      sync.RWMutex
	stats          map[ProcessingStage]*StageStats
	rateLimiter    chan struct{}
	maxConcurrency int
}

// StageStats tracks performance statistics for each stage
type StageStats struct {
	TotalProcessed int64
	TotalBlocked   int64
	TotalErrors    int64
	AverageTime    time.Duration
	LastProcessed  time.Time
}

// NewMessageProcessingPipeline creates a new message processing pipeline
func NewMessageProcessingPipeline(workers int) *MessageProcessingPipeline {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// Limit maximum concurrency based on configuration
	maxConcurrency := workers * 2
	if maxConcurrency > config.MaxConcurrentOperations {
		maxConcurrency = config.MaxConcurrentOperations
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MessageProcessingPipeline{
		stages:         make(map[ProcessingStage]ProcessorFunc),
		workers:        workers,
		jobs:           make(chan MessageProcessingJob, workers*4),
		results:        make(chan ProcessingResult, workers*4),
		ctx:            ctx,
		cancel:         cancel,
		stats:          make(map[ProcessingStage]*StageStats),
		rateLimiter:    make(chan struct{}, maxConcurrency),
		maxConcurrency: maxConcurrency,
	}
}

// RegisterStage registers a processing stage with its processor function
func (p *MessageProcessingPipeline) RegisterStage(stage ProcessingStage, processor ProcessorFunc) {
	p.stagesLock.Lock()
	p.stages[stage] = processor
	p.stagesLock.Unlock()
	
	p.statsLock.Lock()
	p.stats[stage] = &StageStats{}
	p.statsLock.Unlock()
}

// Start begins processing jobs with the configured number of workers
func (p *MessageProcessingPipeline) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start statistics collector
	go p.statsCollector()
}

// worker processes message jobs
func (p *MessageProcessingPipeline) worker(workerID int) {
	defer p.wg.Done()

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			// Rate limiting
			select {
			case p.rateLimiter <- struct{}{}:
				// Acquired rate limit token
			case <-p.ctx.Done():
				return
			}

			// Process the job
			result := p.processJob(job, workerID)

			// Release rate limit token
			<-p.rateLimiter

			// Send result
			select {
			case p.results <- result:
			case <-p.ctx.Done():
				return
			}

		case <-p.ctx.Done():
			log.WithField("worker_id", workerID).Debug("Message processing worker shutting down")
			return
		}
	}
}

// processJob processes a single message job
func (p *MessageProcessingPipeline) processJob(job MessageProcessingJob, workerID int) ProcessingResult {
	startTime := time.Now()

	result := ProcessingResult{
		Stage:     job.Stage,
		UserID:    job.UserID,
		ChatID:    job.ChatID,
		Timestamp: startTime,
	}

	// Find and execute the processor
	p.stagesLock.RLock()
	processor, exists := p.stages[job.Stage]
	p.stagesLock.RUnlock()
	
	if !exists {
		result.Error = fmt.Errorf("no processor registered for stage: %s", job.Stage)
		result.Duration = time.Since(startTime)
		return result
	}

	// Check if pipeline is shutting down before launching goroutine
	select {
	case <-p.ctx.Done():
		result.Error = fmt.Errorf("pipeline shutting down, skipping stage: %s", job.Stage)
		result.Duration = time.Since(startTime)
		return result
	default:
		// Continue with processing
	}
	
	// Execute with timeout
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	var blocked bool
	var err error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("processor panic: %v", r)
				log.WithFields(log.Fields{
					"worker_id": workerID,
					"stage":     job.Stage,
					"user_id":   job.UserID,
					"chat_id":   job.ChatID,
					"panic":     r,
				}).Error("Message processor panicked")
			}
			close(done)
		}()

		blocked, err = processor(ctx, job.Bot, job.Context)
	}()

	select {
	case <-done:
		result.Success = err == nil
		result.Blocked = blocked
		result.Error = err
	case <-ctx.Done():
		result.Error = fmt.Errorf("processor timeout for stage: %s", job.Stage)
		log.WithFields(log.Fields{
			"worker_id": workerID,
			"stage":     job.Stage,
			"user_id":   job.UserID,
			"chat_id":   job.ChatID,
		}).Warn("Message processor timeout")
	}

	result.Duration = time.Since(startTime)

	// Update statistics
	p.updateStats(result)

	return result
}

// AddJob adds a processing job to the pipeline
func (p *MessageProcessingPipeline) AddJob(bot *gotgbot.Bot, ctx *ext.Context, stage ProcessingStage) bool {
	user := ctx.EffectiveSender.User
	chat := ctx.EffectiveChat

	if user == nil || chat == nil {
		return false
	}

	job := MessageProcessingJob{
		Bot:     bot,
		Context: ctx,
		Stage:   stage,
		UserID:  user.Id,
		ChatID:  chat.Id,
	}

	select {
	case p.jobs <- job:
		return true
	case <-p.ctx.Done():
		return false
	default:
		// Channel is full, log warning but don't block
		log.WithFields(log.Fields{
			"stage":   stage,
			"user_id": user.Id,
			"chat_id": chat.Id,
		}).Warn("Message processing pipeline full, dropping job")
		return false
	}
}

// ProcessConcurrently processes multiple stages concurrently and returns if any stage blocked the message
func (p *MessageProcessingPipeline) ProcessConcurrently(bot *gotgbot.Bot, ctx *ext.Context, stages []ProcessingStage) (bool, error) {
	if len(stages) == 0 {
		return false, nil
	}

	processedCount := 0

	// Submit all jobs
	for _, stage := range stages {
		if p.AddJob(bot, ctx, stage) {
			processedCount++
		}
	}

	// Collect results
	blocked := false
	var firstError error
	timeout := time.After(10 * time.Second)

	for i := 0; i < processedCount; i++ {
		select {
		case result := <-p.results:
			if result.Error != nil && firstError == nil {
				firstError = result.Error
			}
			if result.Blocked {
				blocked = true
			}
		case <-timeout:
			return blocked, fmt.Errorf("timeout waiting for processing results")
		}
	}

	return blocked, firstError
}

// updateStats updates performance statistics for a stage
func (p *MessageProcessingPipeline) updateStats(result ProcessingResult) {
	p.statsLock.Lock()
	defer p.statsLock.Unlock()

	stats := p.stats[result.Stage]
	if stats == nil {
		return
	}

	stats.TotalProcessed++
	stats.LastProcessed = result.Timestamp

	if result.Blocked {
		stats.TotalBlocked++
	}
	if result.Error != nil {
		stats.TotalErrors++
	}

	// Update average time using exponential moving average
	alpha := 0.1 // Smoothing factor
	if stats.AverageTime == 0 {
		stats.AverageTime = result.Duration
	} else {
		stats.AverageTime = time.Duration(float64(stats.AverageTime)*(1-alpha) + float64(result.Duration)*alpha)
	}
}

// GetStats returns performance statistics for all stages
func (p *MessageProcessingPipeline) GetStats() map[ProcessingStage]*StageStats {
	p.statsLock.RLock()
	defer p.statsLock.RUnlock()

	result := make(map[ProcessingStage]*StageStats)
	for stage, stats := range p.stats {
		// Create a copy to avoid data races
		result[stage] = &StageStats{
			TotalProcessed: stats.TotalProcessed,
			TotalBlocked:   stats.TotalBlocked,
			TotalErrors:    stats.TotalErrors,
			AverageTime:    stats.AverageTime,
			LastProcessed:  stats.LastProcessed,
		}
	}

	return result
}

// statsCollector periodically logs performance statistics
func (p *MessageProcessingPipeline) statsCollector() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.logStats()
		case <-p.ctx.Done():
			return
		}
	}
}

// logStats logs current performance statistics
func (p *MessageProcessingPipeline) logStats() {
	stats := p.GetStats()

	for stage, stat := range stats {
		if stat.TotalProcessed > 0 {
			log.WithFields(log.Fields{
				"stage":           stage,
				"total_processed": stat.TotalProcessed,
				"total_blocked":   stat.TotalBlocked,
				"total_errors":    stat.TotalErrors,
				"avg_time_ms":     stat.AverageTime.Milliseconds(),
				"last_processed":  stat.LastProcessed.Format(time.RFC3339),
				"block_rate":      fmt.Sprintf("%.2f%%", float64(stat.TotalBlocked)/float64(stat.TotalProcessed)*100),
				"error_rate":      fmt.Sprintf("%.2f%%", float64(stat.TotalErrors)/float64(stat.TotalProcessed)*100),
			}).Info("Message processing pipeline statistics")
		}
	}
}

// Stop gracefully shuts down the pipeline
func (p *MessageProcessingPipeline) Stop() {
	p.cancel()
	close(p.jobs)
	p.wg.Wait()
	close(p.results)

	log.Info("Message processing pipeline stopped")
}

// Health returns the health status of the pipeline
func (p *MessageProcessingPipeline) Health() map[string]any {
	return map[string]any{
		"workers":           p.workers,
		"max_concurrency":   p.maxConcurrency,
		"registered_stages": len(p.stages),
		"job_queue_size":    len(p.jobs),
		"result_queue_size": len(p.results),
		"rate_limit_usage":  len(p.rateLimiter),
	}
}
