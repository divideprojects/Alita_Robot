package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetOrCreateSettings is a generic helper function to get or create settings for any model type.
// Uses Go generics to work with any settings struct and ensures chat exists before creating records.
func GetOrCreateSettings[T any](chatID int64, defaultSettings T, tableName string) (T, error) {
	var settings T

	err := DB.Where("chat_id = ?", chatID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Ensure chat exists before creating settings
			if !ChatExists(chatID) {
				log.Warnf("[Database][GetOrCreateSettings]: Chat %d doesn't exist for %s", chatID, tableName)
				return defaultSettings, nil
			}

			// Create default settings
			settings = defaultSettings
			if err := CreateRecord(&settings); err != nil {
				log.Errorf("[Database][GetOrCreateSettings] creating %s: %d - %v", tableName, chatID, err)
				return defaultSettings, err
			}
		} else {
			log.Errorf("[Database][GetOrCreateSettings] fetching %s: %d - %v", tableName, chatID, err)
			return defaultSettings, err
		}
	}

	return settings, nil
}

// UpdateSettings is a generic helper function to update settings for any model type.
// Supports cache invalidation after successful updates.
func UpdateSettings[T any](chatID int64, updates interface{}, tableName string, cacheKey string) error {
	result := DB.Model(new(T)).Where("chat_id = ?", chatID).Updates(updates)
	if result.Error != nil {
		log.Errorf("[Database][UpdateSettings] updating %s: %d - %v", tableName, chatID, result.Error)
		return fmt.Errorf("failed to update %s for chat %d: %w", tableName, chatID, result.Error)
	}

	// Invalidate cache if a key is provided
	if cacheKey != "" {
		deleteCache(cacheKey)
	}

	return nil
}

// CountRecords is a generic helper function to count records matching a condition.
// Works with any model type using Go generics.
func CountRecords[T any](condition interface{}) (int64, error) {
	var count int64
	err := DB.Model(new(T)).Where(condition).Count(&count).Error
	if err != nil {
		log.Errorf("[Database][CountRecords]: %v", err)
		return 0, fmt.Errorf("failed to count records: %w", err)
	}
	return count, nil
}

// ExistsRecord is a generic helper function to check if any record exists matching a condition.
// Returns true if at least one record matches, false otherwise.
// Uses LIMIT 1 optimization for better performance than COUNT.
func ExistsRecord[T any](condition interface{}) bool {
	var record T
	err := DB.Where(condition).Take(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false
		}
		log.Errorf("[Database][ExistsRecord]: %v", err)
		return false
	}
	return true
}

// DeleteRecords is a generic helper function to delete records matching a condition.
// Supports cache invalidation if records were successfully deleted.
func DeleteRecords[T any](condition interface{}, cacheKey string) error {
	result := DB.Where(condition).Delete(new(T))
	if result.Error != nil {
		log.Errorf("[Database][DeleteRecords]: %v", result.Error)
		return fmt.Errorf("failed to delete records: %w", result.Error)
	}

	// Invalidate cache if affected and key provided
	if result.RowsAffected > 0 && cacheKey != "" {
		deleteCache(cacheKey)
	}

	return nil
}

// GetFirstRecord is a generic helper function to get the first record matching a condition.
// Returns gorm.ErrRecordNotFound if no record matches the condition.
func GetFirstRecord[T any](dest interface{}, condition interface{}) error {
	err := DB.Where(condition).First(dest).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[Database][GetFirstRecord]: %v", err)
		return fmt.Errorf("failed to get first record: %w", err)
	}
	return err
}

// SaveRecord is a generic helper function to save (create or update) a record.
// Uses GORM's Save method which creates if not exists or updates if exists.
// Supports cache invalidation after successful save.
func SaveRecord(record interface{}, cacheKey string) error {
	err := DB.Save(record).Error
	if err != nil {
		log.Errorf("[Database][SaveRecord]: %v", err)
		return fmt.Errorf("failed to save record: %w", err)
	}

	// Invalidate cache if key provided
	if cacheKey != "" {
		deleteCache(cacheKey)
	}

	return nil
}

// BatchCreate creates multiple records in batches for improved performance.
// Uses configurable batch size with fallback to default BulkBatchSize if not specified.
func BatchCreate[T any](records []T, batchSize int) error {
	if len(records) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = BulkBatchSize
	}

	err := DB.CreateInBatches(records, batchSize).Error
	if err != nil {
		log.Errorf("[Database][BatchCreate]: %v", err)
		return fmt.Errorf("failed to batch create %d records: %w", len(records), err)
	}

	return nil
}

// TransactionWrapper wraps a function in a database transaction.
// Automatically commits on success or rolls back on error.
func TransactionWrapper(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// TransactionWrapperWithContext wraps a function in a database transaction with context support.
// Allows for context cancellation during transaction execution.
func TransactionWrapperWithContext(ctx context.Context, fn func(*gorm.DB) error) error {
	return DB.WithContext(ctx).Transaction(fn)
}

// GetWithCache is a generic cached getter with automatic cache management.
// Loads data from cache if available, otherwise calls the loader function and caches the result.
func GetWithCache[T any](key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	return getFromCacheOrLoad(key, ttl, loader)
}

// InvalidateMultipleCache invalidates multiple cache keys at once for batch cache cleanup.
// Useful when multiple related cache entries need to be cleared simultaneously.
func InvalidateMultipleCache(keys ...string) {
	for _, key := range keys {
		deleteCache(key)
	}
}

// CountDistinct counts distinct values in a specified column for a model type.
// Supports optional WHERE conditions and works with any model type using Go generics.
func CountDistinct[T any](column string, condition interface{}) (int64, error) {
	var count int64
	query := DB.Model(new(T))

	if condition != nil {
		query = query.Where(condition)
	}

	err := query.Select(fmt.Sprintf("COUNT(DISTINCT %s)", column)).Scan(&count).Error
	if err != nil {
		log.Errorf("[Database][CountDistinct]: %v", err)
		return 0, fmt.Errorf("failed to count distinct values for column %s: %w", column, err)
	}

	return count, nil
}

// BulkUpdate performs bulk update on multiple records matching a condition.
// Supports cache invalidation for multiple keys if records were actually updated.
func BulkUpdate[T any](condition interface{}, updates interface{}, cacheKeys ...string) error {
	result := DB.Model(new(T)).Where(condition).Updates(updates)
	if result.Error != nil {
		log.Errorf("[Database][BulkUpdate]: %v", result.Error)
		return fmt.Errorf("failed to bulk update records: %w", result.Error)
	}

	// Invalidate all provided cache keys if records were updated
	if result.RowsAffected > 0 {
		InvalidateMultipleCache(cacheKeys...)
	}

	return nil
}

// GetOrDefault returns a value from a getter function or a default if not found or error occurs.
// Handles gorm.ErrRecordNotFound specifically and falls back to default for other errors.
func GetOrDefault[T any](getter func() (T, error), defaultValue T) T {
	value, err := getter()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultValue
		}
		log.Debugf("[Database][GetOrDefault]: Using default due to error: %v", err)
		return defaultValue
	}
	return value
}
