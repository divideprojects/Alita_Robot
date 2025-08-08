package db

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// Batch size for bulk operations
	BulkBatchSize      = 100
	FilterBatchSize    = 50
	BlacklistBatchSize = 50

	// Timeout for bulk operations
	BulkOperationTimeout = 30 * time.Second

	// Worker pool configuration
	DefaultWorkerCount = 4
	MaxWorkerCount     = 10
)

// BulkOperationStats tracks bulk operation performance
type BulkOperationStats struct {
	ProcessedCount int
	FailedCount    int
	Duration       time.Duration
	StartTime      time.Time
}

// LogBulkOperationStats logs the performance stats of bulk operations.
// Includes processed count, failed count, duration, and records per second.
func LogBulkOperationStats(operation string, stats *BulkOperationStats) {
	log.WithFields(log.Fields{
		"operation":     operation,
		"processed":     stats.ProcessedCount,
		"failed":        stats.FailedCount,
		"duration_ms":   stats.Duration.Milliseconds(),
		"records_per_s": float64(stats.ProcessedCount) / stats.Duration.Seconds(),
	}).Info("Bulk operation completed")
}

// BulkProcessingJob represents a generic bulk processing job
type BulkProcessingJob[T any] struct {
	Data  []T
	Index int
}

// BulkProcessingResult represents the result of bulk processing
type BulkProcessingResult struct {
	Index          int
	ProcessedCount int
	Error          error
}

// ParallelBulkProcessor handles concurrent bulk database operations
type ParallelBulkProcessor[T any] struct {
	workerCount int
	jobs        chan BulkProcessingJob[T]
	results     chan BulkProcessingResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	processor   func([]T) error
}

// NewParallelBulkProcessor creates a new parallel bulk processor with specified worker count.
// Worker count is limited by CPU cores and maximum configured workers.
func NewParallelBulkProcessor[T any](workerCount int, processor func([]T) error) *ParallelBulkProcessor[T] {
	if workerCount <= 0 {
		workerCount = DefaultWorkerCount
	}
	if workerCount > MaxWorkerCount {
		workerCount = MaxWorkerCount
	}

	// Limit based on available CPU cores
	maxCores := runtime.NumCPU()
	if workerCount > maxCores {
		workerCount = maxCores
	}

	ctx, cancel := context.WithTimeout(context.Background(), BulkOperationTimeout)

	return &ParallelBulkProcessor[T]{
		workerCount: workerCount,
		jobs:        make(chan BulkProcessingJob[T], workerCount*2),
		results:     make(chan BulkProcessingResult, workerCount*2),
		ctx:         ctx,
		cancel:      cancel,
		processor:   processor,
	}
}

// Start begins processing jobs with the configured number of workers.
// Each worker runs in a separate goroutine and processes jobs from the job channel.
func (p *ParallelBulkProcessor[T]) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processes bulk jobs from the jobs channel.
// Runs until the jobs channel is closed or context is cancelled.
func (p *ParallelBulkProcessor[T]) worker(workerID int) {
	defer p.wg.Done()
	
	// Add panic recovery to prevent worker death
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{
				"worker_id": workerID,
				"panic":     r,
			}).Error("Bulk processor worker panicked")
		}
	}()

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			result := BulkProcessingResult{
				Index: job.Index,
			}
			
			// Wrap processor call in function with its own panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						result.Error = fmt.Errorf("processor panic: %v", r)
						log.WithFields(log.Fields{
							"worker_id": workerID,
							"job_index": job.Index,
							"panic":     r,
						}).Error("Bulk processor job panicked")
					}
				}()
				
				if err := p.processor(job.Data); err != nil {
					result.Error = err
					log.WithFields(log.Fields{
						"worker_id": workerID,
						"job_index": job.Index,
						"error":     err,
					}).Error("Bulk processing job failed")
				} else {
					result.ProcessedCount = len(job.Data)
				}
			}()

			p.results <- result

		case <-p.ctx.Done():
			log.WithField("worker_id", workerID).Warn("Bulk processor worker context cancelled")
			return
		}
	}
}

// AddJob adds a job to the processor's job queue.
// Returns immediately if the context is cancelled.
func (p *ParallelBulkProcessor[T]) AddJob(data []T, index int) {
	select {
	case p.jobs <- BulkProcessingJob[T]{Data: data, Index: index}:
	case <-p.ctx.Done():
		log.Warn("Cannot add bulk processing job: context cancelled")
	}
}

// Close shuts down the processor gracefully.
// Closes channels, waits for workers to finish, and cancels the context.
func (p *ParallelBulkProcessor[T]) Close() {
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
	p.cancel()
}

// ProcessInParallel processes large datasets in parallel batches.
// Returns performance statistics including processed count, failed count, and duration.
func (p *ParallelBulkProcessor[T]) ProcessInParallel(data []T, batchSize int) *BulkOperationStats {
	stats := &BulkOperationStats{
		StartTime: time.Now(),
	}

	if len(data) == 0 {
		stats.Duration = time.Since(stats.StartTime)
		return stats
	}

	p.Start()

	// Split data into batches and submit jobs
	jobCount := 0
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]
		p.AddJob(batch, jobCount)
		jobCount++
	}

	// Signal no more jobs
	close(p.jobs)

	// Collect results
	for i := 0; i < jobCount; i++ {
		result := <-p.results
		if result.Error != nil {
			stats.FailedCount += len(data) / jobCount // Approximate failed count
		} else {
			stats.ProcessedCount += result.ProcessedCount
		}
	}

	// Wait for completion and cleanup
	p.wg.Wait()
	close(p.results)
	p.cancel()

	stats.Duration = time.Since(stats.StartTime)
	return stats
}

// BulkAddFilters adds multiple filters in a single operation with performance tracking.
// Uses batch inserts for efficiency and invalidates cache after successful operations.
func BulkAddFilters(chatID int64, filters map[string]string) (*BulkOperationStats, error) {
	stats := &BulkOperationStats{
		StartTime: time.Now(),
	}

	if len(filters) == 0 {
		stats.Duration = time.Since(stats.StartTime)
		return stats, nil
	}

	filterRecords := make([]ChatFilters, 0, len(filters))
	for keyword, reply := range filters {
		filterRecords = append(filterRecords, ChatFilters{
			ChatId:      chatID,
			KeyWord:     keyword,
			FilterReply: reply,
			MsgType:     TEXT, // default to text
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), BulkOperationTimeout)
	defer cancel()

	// Use CreateInBatches for efficient bulk insert with context
	err := DB.WithContext(ctx).CreateInBatches(filterRecords, FilterBatchSize).Error
	if err != nil {
		stats.FailedCount = len(filters)
		log.Errorf("[Database][BulkAddFilters]: %d - %v", chatID, err)
	} else {
		stats.ProcessedCount = len(filters)
		// Invalidate cache after bulk add
		deleteCache(filterListCacheKey(chatID))
	}

	stats.Duration = time.Since(stats.StartTime)
	return stats, err
}

// BulkRemoveFilters removes multiple filters in a single operation.
// Invalidates cache after successful removal.
func BulkRemoveFilters(chatID int64, keywords []string) error {
	if len(keywords) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND keyword IN ?", chatID, keywords).Delete(&ChatFilters{}).Error
	if err != nil {
		log.Errorf("[Database][BulkRemoveFilters]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk remove
	deleteCache(filterListCacheKey(chatID))
	return nil
}

// BulkAddNotes adds multiple notes in a single operation with default text type and web preview enabled.
// Invalidates related cache entries after successful operations.
func BulkAddNotes(chatID int64, notes map[string]string) error {
	if len(notes) == 0 {
		return nil
	}

	noteRecords := make([]Notes, 0, len(notes))
	for name, content := range notes {
		noteRecords = append(noteRecords, Notes{
			ChatId:      chatID,
			NoteName:    name,
			NoteContent: content,
			MsgType:     TEXT, // default to text
			WebPreview:  true, // default web preview enabled
		})
	}

	// Use CreateInBatches for efficient bulk insert
	err := DB.CreateInBatches(noteRecords, BulkBatchSize).Error
	if err != nil {
		log.Errorf("[Database][BulkAddNotes]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk add
	deleteCache(notesListCacheKey(chatID, true))
	deleteCache(notesListCacheKey(chatID, false))
	return nil
}

// BulkRemoveNotes removes multiple notes in a single operation.
// Invalidates related cache entries after successful removal.
func BulkRemoveNotes(chatID int64, noteNames []string) error {
	if len(noteNames) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND note_name IN ?", chatID, noteNames).Delete(&Notes{}).Error
	if err != nil {
		log.Errorf("[Database][BulkRemoveNotes]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk remove
	deleteCache(notesListCacheKey(chatID, true))
	deleteCache(notesListCacheKey(chatID, false))
	return nil
}

// BulkAddBlacklist adds multiple blacklist words in a single operation with performance tracking.
// Uses batch inserts for efficiency and invalidates cache after successful operations.
func BulkAddBlacklist(chatID int64, words []string, action string) (*BulkOperationStats, error) {
	stats := &BulkOperationStats{
		StartTime: time.Now(),
	}

	if len(words) == 0 {
		stats.Duration = time.Since(stats.StartTime)
		return stats, nil
	}

	blacklistRecords := make([]BlacklistSettings, 0, len(words))
	for _, word := range words {
		blacklistRecords = append(blacklistRecords, BlacklistSettings{
			ChatId: chatID,
			Word:   strings.ToLower(word),
			Action: action,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), BulkOperationTimeout)
	defer cancel()

	// Use CreateInBatches for efficient bulk insert with context
	err := DB.WithContext(ctx).CreateInBatches(blacklistRecords, BlacklistBatchSize).Error
	if err != nil {
		stats.FailedCount = len(words)
		log.Errorf("[Database][BulkAddBlacklist]: %d - %v", chatID, err)
	} else {
		stats.ProcessedCount = len(words)
		// Invalidate cache after bulk add
		deleteCache(blacklistCacheKey(chatID))
	}

	stats.Duration = time.Since(stats.StartTime)
	return stats, err
}

// BulkRemoveBlacklist removes multiple blacklist words in a single operation.
// Invalidates cache after successful removal.
func BulkRemoveBlacklist(chatID int64, words []string) error {
	if len(words) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND word IN ?", chatID, words).Delete(&BlacklistSettings{}).Error
	if err != nil {
		log.Errorf("[Database][BulkRemoveBlacklist]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk remove
	deleteCache(blacklistCacheKey(chatID))
	return nil
}

// BulkWarnUsers warns multiple users in a single transaction.
// Creates warn settings with default limit if not found, and updates user warn counts.
func BulkWarnUsers(chatID int64, userIDs []int64, reason string) error {
	if len(userIDs) == 0 {
		return nil
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		// Get warn settings
		warnSettings := &WarnSettings{}
		err := tx.Where("chat_id = ?", chatID).First(warnSettings).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create default warn settings if not found
				warnSettings = &WarnSettings{
					ChatId:    chatID,
					WarnLimit: 3,
				}
				if err := tx.Create(warnSettings).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Process each user
		for _, userID := range userIDs {
			warns := &Warns{}
			err := tx.Where("user_id = ? AND chat_id = ?", userID, chatID).First(warns).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// Create new warn record
					warns = &Warns{
						UserId:   userID,
						ChatId:   chatID,
						NumWarns: 1,
						Reasons:  StringArray{reason},
					}
					if err := tx.Create(warns).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				// Update existing warn record
				warns.NumWarns++
				warns.Reasons = append(warns.Reasons, reason)
				if err := tx.Save(warns).Error; err != nil {
					return err
				}
			}
		}

		// Invalidate cache after bulk warn
		deleteCache(warnSettingsCacheKey(chatID))
		return nil
	})
}

// BulkResetWarns resets warnings for multiple users in a single operation.
// Removes all warn records for the specified users and invalidates cache.
func BulkResetWarns(chatID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND user_id IN ?", chatID, userIDs).Delete(&Warns{}).Error
	if err != nil {
		log.Errorf("[Database][BulkResetWarns]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk reset
	deleteCache(warnSettingsCacheKey(chatID))
	return nil
}

// OptimizedQuery provides query optimization utilities
type OptimizedQuery struct {
	db *gorm.DB
}

// NewOptimizedQuery creates a new optimized query instance.
// Uses the global database connection for query operations.
func NewOptimizedQuery() *OptimizedQuery {
	return &OptimizedQuery{db: DB}
}

// GetFiltersWithPagination retrieves filters with pagination for better memory usage.
// Returns a subset of filters based on limit and offset parameters.
func (oq *OptimizedQuery) GetFiltersWithPagination(chatID int64, limit, offset int) ([]*ChatFilters, error) {
	var filters []*ChatFilters
	err := oq.db.Where("chat_id = ?", chatID).Limit(limit).Offset(offset).Find(&filters).Error
	if err != nil {
		log.Errorf("[OptimizedQuery] GetFiltersWithPagination: %v - chatID: %d", err, chatID)
		return nil, fmt.Errorf("failed to get paginated filters: %w", err)
	}
	return filters, nil
}

// AnalyzePerformanceStats provides performance statistics for monitoring.
// Returns comprehensive statistics about filters, blacklists, and system performance features.
func AnalyzePerformanceStats() map[string]any {
	stats := make(map[string]any)

	// Get filter statistics
	var filterCount int64
	var filterChats int64
	DB.Model(&ChatFilters{}).Count(&filterCount)
	DB.Model(&ChatFilters{}).Select("COUNT(DISTINCT chat_id)").Scan(&filterChats)

	// Get blacklist statistics
	var blacklistCount int64
	var blacklistChats int64
	DB.Model(&BlacklistSettings{}).Count(&blacklistCount)
	DB.Model(&BlacklistSettings{}).Select("COUNT(DISTINCT chat_id)").Scan(&blacklistChats)

	stats["filters"] = map[string]int64{
		"total_filters": filterCount,
		"chats_using":   filterChats,
		"avg_per_chat":  0,
	}

	if filterChats > 0 {
		stats["filters"].(map[string]int64)["avg_per_chat"] = filterCount / filterChats
	}

	stats["blacklists"] = map[string]int64{
		"total_entries": blacklistCount,
		"chats_using":   blacklistChats,
		"avg_per_chat":  0,
	}

	if blacklistChats > 0 {
		stats["blacklists"].(map[string]int64)["avg_per_chat"] = blacklistCount / blacklistChats
	}

	stats["performance"] = map[string]any{
		"regex_cache_enabled":     true,
		"database_cache_enabled":  true,
		"bulk_operations_enabled": true,
		"async_operations":        true,
	}

	return stats
}

// ParallelBulkAddFilters adds multiple filters concurrently with improved performance.
// Uses parallel processing with worker pools and invalidates cache after successful operations.
func ParallelBulkAddFilters(chatID int64, filters map[string]string) (*BulkOperationStats, error) {
	if len(filters) == 0 {
		return &BulkOperationStats{StartTime: time.Now(), Duration: 0}, nil
	}

	// Convert map to slice for processing
	filterRecords := make([]ChatFilters, 0, len(filters))
	for keyword, reply := range filters {
		filterRecords = append(filterRecords, ChatFilters{
			ChatId:      chatID,
			KeyWord:     keyword,
			FilterReply: reply,
			MsgType:     TEXT,
		})
	}

	// Create processor function
	processor := func(batch []ChatFilters) error {
		ctx, cancel := context.WithTimeout(context.Background(), BulkOperationTimeout)
		defer cancel()

		return DB.WithContext(ctx).CreateInBatches(batch, FilterBatchSize).Error
	}

	// Process in parallel
	bulkProcessor := NewParallelBulkProcessor(4, processor)
	stats := bulkProcessor.ProcessInParallel(filterRecords, FilterBatchSize)

	if stats.ProcessedCount > 0 {
		// Invalidate cache after successful bulk add
		go func() {
			deleteCache(filterListCacheKey(chatID))
		}()
	}

	LogBulkOperationStats("ParallelBulkAddFilters", stats)

	var err error
	if stats.FailedCount > 0 {
		err = fmt.Errorf("failed to process %d out of %d filter records", stats.FailedCount, len(filterRecords))
		log.Errorf("[Database][ParallelBulkAddFilters]: %d - %v", chatID, err)
	}

	return stats, err
}

// ParallelBulkAddBlacklist adds multiple blacklist words concurrently.
// Uses parallel processing with worker pools and invalidates cache after successful operations.
func ParallelBulkAddBlacklist(chatID int64, words []string, action string) (*BulkOperationStats, error) {
	if len(words) == 0 {
		return &BulkOperationStats{StartTime: time.Now(), Duration: 0}, nil
	}

	// Convert to blacklist records
	blacklistRecords := make([]BlacklistSettings, 0, len(words))
	for _, word := range words {
		blacklistRecords = append(blacklistRecords, BlacklistSettings{
			ChatId: chatID,
			Word:   strings.ToLower(word),
			Action: action,
		})
	}

	// Create processor function
	processor := func(batch []BlacklistSettings) error {
		ctx, cancel := context.WithTimeout(context.Background(), BulkOperationTimeout)
		defer cancel()

		return DB.WithContext(ctx).CreateInBatches(batch, BlacklistBatchSize).Error
	}

	// Process in parallel
	bulkProcessor := NewParallelBulkProcessor(4, processor)
	stats := bulkProcessor.ProcessInParallel(blacklistRecords, BlacklistBatchSize)

	if stats.ProcessedCount > 0 {
		// Invalidate cache after successful bulk add
		go func() {
			deleteCache(blacklistCacheKey(chatID))
		}()
	}

	LogBulkOperationStats("ParallelBulkAddBlacklist", stats)

	var err error
	if stats.FailedCount > 0 {
		err = fmt.Errorf("failed to process %d out of %d blacklist records", stats.FailedCount, len(blacklistRecords))
		log.Errorf("[Database][ParallelBulkAddBlacklist]: %d - %v", chatID, err)
	}

	return stats, err
}

// ParallelCacheInvalidation invalidates multiple cache keys concurrently.
// Uses a worker pool to invalidate cache keys with small delays to avoid overwhelming the cache system.
func ParallelCacheInvalidation(cacheKeys []string) {
	if len(cacheKeys) == 0 {
		return
	}

	// Use worker pool for cache invalidation
	workerCount := 3
	if len(cacheKeys) < workerCount {
		workerCount = len(cacheKeys)
	}

	jobs := make(chan string, len(cacheKeys))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for cacheKey := range jobs {
				deleteCache(cacheKey)
				// Small delay to avoid overwhelming cache system
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// Submit jobs
	for _, key := range cacheKeys {
		jobs <- key
	}
	close(jobs)

	// Wait for completion
	wg.Wait()

	log.WithField("cache_keys_invalidated", len(cacheKeys)).Info("Parallel cache invalidation completed")
}

// CleanupExpiredEntries removes old entries that might affect performance.
// Currently a placeholder for future cleanup operations like expired warnings and temp bans.
func CleanupExpiredEntries() error {
	// This is a placeholder for future cleanup operations
	// Could include removing old warnings, expired temp bans, etc.
	log.Info("[Database] Performance cleanup completed")
	return nil
}
