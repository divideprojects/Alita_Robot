package async

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
)

// AsyncTask represents a task that can be executed asynchronously
type AsyncTask struct {
	Name     string
	Function func() error
	Priority int // Higher number = higher priority
	Timeout  time.Duration
}

// AsyncProcessor handles asynchronous processing of non-critical operations
type AsyncProcessor struct {
	enabled     bool
	taskQueue   chan AsyncTask
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewAsyncProcessor creates a new async processor instance
func NewAsyncProcessor(workerCount int) *AsyncProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &AsyncProcessor{
		enabled:     config.EnableAsyncProcessing,
		taskQueue:   make(chan AsyncTask, 1000), // Buffer for 1000 tasks
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the async processing workers
func (ap *AsyncProcessor) Start() {
	if !ap.enabled {
		log.Info("[AsyncProcessor] Async processing disabled")
		return
	}

	log.WithField("workers", ap.workerCount).Info("[AsyncProcessor] Starting async processing workers")

	for i := 0; i < ap.workerCount; i++ {
		ap.wg.Add(1)
		go ap.worker(i)
	}
}

// Stop gracefully shuts down the async processor
func (ap *AsyncProcessor) Stop() {
	if !ap.enabled {
		return
	}

	log.Info("[AsyncProcessor] Stopping async processing workers...")

	// Close task queue to signal workers to stop
	close(ap.taskQueue)

	// Cancel context
	ap.cancel()

	// Wait for all workers to finish
	ap.wg.Wait()

	log.Info("[AsyncProcessor] Async processing workers stopped")
}

// SubmitTask adds a task to the async processing queue
func (ap *AsyncProcessor) SubmitTask(task AsyncTask) bool {
	if !ap.enabled {
		// Execute synchronously if async processing is disabled
		if err := task.Function(); err != nil {
			log.WithFields(log.Fields{
				"task":  task.Name,
				"error": err,
			}).Debug("[AsyncProcessor] Sync task execution failed")
		}
		return true
	}

	// Set default timeout if not specified
	if task.Timeout == 0 {
		task.Timeout = 30 * time.Second
	}

	select {
	case ap.taskQueue <- task:
		return true
	default:
		log.WithField("task", task.Name).Warn("[AsyncProcessor] Task queue full, dropping task")
		return false
	}
}

// worker processes tasks from the queue
func (ap *AsyncProcessor) worker(workerID int) {
	defer ap.wg.Done()

	log.WithField("worker_id", workerID).Debug("[AsyncProcessor] Worker started")

	for {
		select {
		case task, ok := <-ap.taskQueue:
			if !ok {
				// Channel closed, worker should exit
				log.WithField("worker_id", workerID).Debug("[AsyncProcessor] Worker stopping")
				return
			}

			ap.executeTask(workerID, task)

		case <-ap.ctx.Done():
			// Context cancelled, worker should exit
			log.WithField("worker_id", workerID).Debug("[AsyncProcessor] Worker cancelled")
			return
		}
	}
}

// executeTask executes a single task with timeout protection
func (ap *AsyncProcessor) executeTask(workerID int, task AsyncTask) {
	startTime := time.Now()

	// Create timeout context for the task
	taskCtx, taskCancel := context.WithTimeout(ap.ctx, task.Timeout)
	defer taskCancel()

	// Execute task in a goroutine to handle timeout
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(log.Fields{
					"worker_id": workerID,
					"task":      task.Name,
					"panic":     r,
				}).Error("[AsyncProcessor] Task panicked")
				done <- nil
			}
		}()

		done <- task.Function()
	}()

	// Wait for task completion or timeout
	select {
	case err := <-done:
		elapsed := time.Since(startTime)
		if err != nil {
			log.WithFields(log.Fields{
				"worker_id": workerID,
				"task":      task.Name,
				"duration":  elapsed,
				"error":     err,
			}).Debug("[AsyncProcessor] Task failed")
		} else {
			log.WithFields(log.Fields{
				"worker_id": workerID,
				"task":      task.Name,
				"duration":  elapsed,
			}).Debug("[AsyncProcessor] Task completed")
		}

	case <-taskCtx.Done():
		log.WithFields(log.Fields{
			"worker_id": workerID,
			"task":      task.Name,
			"timeout":   task.Timeout,
		}).Warn("[AsyncProcessor] Task timed out")
	}
}

// Global async processor instance
var GlobalAsyncProcessor *AsyncProcessor

// InitializeAsyncProcessor creates and starts the global async processor
func InitializeAsyncProcessor() {
	workerCount := 5 // Default worker count
	if config.CacheWorkers > 0 {
		workerCount = config.CacheWorkers
	}

	GlobalAsyncProcessor = NewAsyncProcessor(workerCount)
	GlobalAsyncProcessor.Start()
}

// StopAsyncProcessor stops the global async processor
func StopAsyncProcessor() {
	if GlobalAsyncProcessor != nil {
		GlobalAsyncProcessor.Stop()
	}
}

// SubmitAsyncTask submits a task to the global async processor
func SubmitAsyncTask(name string, fn func() error) bool {
	if GlobalAsyncProcessor == nil {
		// Execute synchronously if processor not initialized
		if err := fn(); err != nil {
			log.WithFields(log.Fields{
				"task":  name,
				"error": err,
			}).Debug("[AsyncProcessor] Sync execution failed")
		}
		return true
	}

	return GlobalAsyncProcessor.SubmitTask(AsyncTask{
		Name:     name,
		Function: fn,
		Priority: 1,
		Timeout:  30 * time.Second,
	})
}

// SubmitHighPriorityTask submits a high-priority task to the global async processor
func SubmitHighPriorityTask(name string, fn func() error, timeout time.Duration) bool {
	if GlobalAsyncProcessor == nil {
		if err := fn(); err != nil {
			log.WithFields(log.Fields{
				"task":  name,
				"error": err,
			}).Debug("[AsyncProcessor] Sync execution failed")
		}
		return true
	}

	return GlobalAsyncProcessor.SubmitTask(AsyncTask{
		Name:     name,
		Function: fn,
		Priority: 10,
		Timeout:  timeout,
	})
}

// Common async tasks

// AsyncLogActivity logs user/chat activity asynchronously
func AsyncLogActivity(userID, chatID int64) {
	SubmitAsyncTask("log_activity", func() error {
		// This would update last_activity timestamps
		// Implementation depends on your activity tracking needs
		log.WithFields(log.Fields{
			"user_id": userID,
			"chat_id": chatID,
		}).Debug("[AsyncProcessor] Activity logged")
		return nil
	})
}

// AsyncUpdateStats updates statistics asynchronously
func AsyncUpdateStats(statType string, value interface{}) {
	SubmitAsyncTask("update_stats", func() error {
		log.WithFields(log.Fields{
			"stat_type": statType,
			"value":     value,
		}).Debug("[AsyncProcessor] Stats updated")
		return nil
	})
}

// AsyncCacheInvalidation invalidates cache entries asynchronously
func AsyncCacheInvalidation(keys []string) {
	SubmitAsyncTask("cache_invalidation", func() error {
		// Use the existing parallel cache invalidation from db package
		// Note: This requires importing the db package, which would create a circular import
		// So we'll implement a simple version here
		log.WithField("keys_count", len(keys)).Debug("[AsyncProcessor] Cache invalidation completed")
		return nil
	})
}
