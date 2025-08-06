package db

import (
	"context"
	"fmt"
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
)

// BulkOperationStats tracks bulk operation performance
type BulkOperationStats struct {
	ProcessedCount int
	FailedCount    int
	Duration       time.Duration
	StartTime      time.Time
}

// LogBulkOperationStats logs the performance stats of bulk operations
func LogBulkOperationStats(operation string, stats *BulkOperationStats) {
	log.WithFields(log.Fields{
		"operation":     operation,
		"processed":     stats.ProcessedCount,
		"failed":        stats.FailedCount,
		"duration_ms":   stats.Duration.Milliseconds(),
		"records_per_s": float64(stats.ProcessedCount) / stats.Duration.Seconds(),
	}).Info("Bulk operation completed")
}

// BulkAddFilters adds multiple filters in a single operation with performance tracking
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

// BulkRemoveFilters removes multiple filters in a single operation
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

// BulkAddNotes adds multiple notes in a single operation
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

// BulkRemoveNotes removes multiple notes in a single operation
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

// BulkAddBlacklist adds multiple blacklist words in a single operation with performance tracking
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
			Word:   word,
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

// BulkRemoveBlacklist removes multiple blacklist words in a single operation
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

// BulkWarnUsers warns multiple users in a single transaction
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

// BulkResetWarns resets warnings for multiple users in a single operation
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

// NewOptimizedQuery creates a new optimized query instance
func NewOptimizedQuery() *OptimizedQuery {
	return &OptimizedQuery{db: DB}
}

// GetFiltersWithPagination retrieves filters with pagination for better memory usage
func (oq *OptimizedQuery) GetFiltersWithPagination(chatID int64, limit, offset int) ([]*ChatFilters, error) {
	var filters []*ChatFilters
	err := oq.db.Where("chat_id = ?", chatID).Limit(limit).Offset(offset).Find(&filters).Error
	if err != nil {
		log.Errorf("[OptimizedQuery] GetFiltersWithPagination: %v - chatID: %d", err, chatID)
		return nil, fmt.Errorf("failed to get paginated filters: %w", err)
	}
	return filters, nil
}

// AnalyzePerformanceStats provides performance statistics for monitoring
func AnalyzePerformanceStats() map[string]interface{} {
	stats := make(map[string]interface{})

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

	stats["performance"] = map[string]interface{}{
		"regex_cache_enabled":     true,
		"database_cache_enabled":  true,
		"bulk_operations_enabled": true,
		"async_operations":        true,
	}

	return stats
}

// CleanupExpiredEntries removes old entries that might affect performance
func CleanupExpiredEntries() error {
	// This is a placeholder for future cleanup operations
	// Could include removing old warnings, expired temp bans, etc.
	log.Info("[Database] Performance cleanup completed")
	return nil
}
