package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetOrCreateSettings is a generic helper to get or create settings for a model
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

// UpdateSettings is a generic helper to update settings
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

// CountRecords is a generic helper to count records with a condition
func CountRecords[T any](condition interface{}) (int64, error) {
	var count int64
	err := DB.Model(new(T)).Where(condition).Count(&count).Error
	if err != nil {
		log.Errorf("[Database][CountRecords]: %v", err)
		return 0, fmt.Errorf("failed to count records: %w", err)
	}
	return count, nil
}

// ExistsRecord is a generic helper to check if a record exists
func ExistsRecord[T any](condition interface{}) bool {
	count, err := CountRecords[T](condition)
	if err != nil {
		return false
	}
	return count > 0
}

// DeleteRecords is a generic helper to delete records
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

// GetFirstRecord is a generic helper to get the first matching record
func GetFirstRecord[T any](dest interface{}, condition interface{}) error {
	err := DB.Where(condition).First(dest).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[Database][GetFirstRecord]: %v", err)
		return fmt.Errorf("failed to get first record: %w", err)
	}
	return err
}

// SaveRecord is a generic helper to save (create or update) a record
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

// BatchCreate creates multiple records in batches
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

// TransactionWrapper wraps a function in a database transaction
func TransactionWrapper(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// TransactionWrapperWithContext wraps a function in a database transaction with context
func TransactionWrapperWithContext(ctx context.Context, fn func(*gorm.DB) error) error {
	return DB.WithContext(ctx).Transaction(fn)
}

// GetWithCache is a generic cached getter with automatic cache management
func GetWithCache[T any](key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	return getFromCacheOrLoad(key, ttl, loader)
}

// InvalidateMultipleCache invalidates multiple cache keys at once
func InvalidateMultipleCache(keys ...string) {
	for _, key := range keys {
		deleteCache(key)
	}
}

// CountDistinct counts distinct values in a column
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

// BulkUpdate performs bulk update on multiple records
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

// GetOrDefault returns a value or a default if not found
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