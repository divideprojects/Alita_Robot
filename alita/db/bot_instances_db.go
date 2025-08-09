package db

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CreateBotInstance creates a new bot instance record in the database
func CreateBotInstance(botID, ownerID int64, tokenHash, username, name string) (*BotInstance, error) {
	instance := &BotInstance{
		BotID:       botID,
		OwnerID:     ownerID,
		TokenHash:   tokenHash,
		BotUsername: username,
		BotName:     name,
		IsActive:    true,
		WebhookURL:  "", // Will be set when webhook is configured
	}

	err := DB.Create(instance).Error
	if err != nil {
		log.Errorf("[Database] CreateBotInstance: %v", err)
		return nil, err
	}

	log.Infof("[Database] Created bot instance for bot_id %d, owner_id %d", botID, ownerID)
	return instance, nil
}

// GetBotInstance retrieves a bot instance by bot ID
func GetBotInstance(botID int64) (*BotInstance, error) {
	var instance BotInstance
	err := DB.Where("bot_id = ?", botID).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil instead of error for not found
		}
		log.Errorf("[Database] GetBotInstance: %v", err)
		return nil, err
	}
	return &instance, nil
}

// GetUserBotInstances retrieves all bot instances owned by a specific user
func GetUserBotInstances(ownerID int64) ([]BotInstance, error) {
	var instances []BotInstance
	err := DB.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&instances).Error
	if err != nil {
		log.Errorf("[Database] GetUserBotInstances: %v", err)
		return nil, err
	}
	return instances, nil
}

// GetActiveBotInstances retrieves all active bot instances
func GetActiveBotInstances() ([]BotInstance, error) {
	var instances []BotInstance
	err := DB.Where("is_active = ?", true).Order("created_at DESC").Find(&instances).Error
	if err != nil {
		log.Errorf("[Database] GetActiveBotInstances: %v", err)
		return nil, err
	}
	return instances, nil
}

// GetAllBotInstances retrieves all bot instances with optional pagination
func GetAllBotInstances(limit, offset int) ([]BotInstance, error) {
	var instances []BotInstance
	query := DB.Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&instances).Error
	if err != nil {
		log.Errorf("[Database] GetAllBotInstances: %v", err)
		return nil, err
	}
	return instances, nil
}

// UpdateBotInstanceActivity updates the last activity timestamp for a bot instance
func UpdateBotInstanceActivity(botID int64) error {
	now := time.Now()
	err := DB.Model(&BotInstance{}).Where("bot_id = ?", botID).Updates(map[string]interface{}{
		"last_activity": now,
		"updated_at":    now,
	}).Error
	
	if err != nil {
		log.Errorf("[Database] UpdateBotInstanceActivity: %v", err)
		return err
	}
	return nil
}

// UpdateBotInstanceStatus updates the active status of a bot instance
func UpdateBotInstanceStatus(botID int64, isActive bool) error {
	err := DB.Model(&BotInstance{}).Where("bot_id = ?", botID).Updates(map[string]interface{}{
		"is_active":  isActive,
		"updated_at": time.Now(),
	}).Error
	
	if err != nil {
		log.Errorf("[Database] UpdateBotInstanceStatus: %v", err)
		return err
	}
	return nil
}

// UpdateBotInstanceWebhook updates the webhook URL for a bot instance
func UpdateBotInstanceWebhook(botID int64, webhookURL string) error {
	err := DB.Model(&BotInstance{}).Where("bot_id = ?", botID).Updates(map[string]interface{}{
		"webhook_url": webhookURL,
		"updated_at":  time.Now(),
	}).Error
	
	if err != nil {
		log.Errorf("[Database] UpdateBotInstanceWebhook: %v", err)
		return err
	}
	return nil
}

// DeleteBotInstance deletes a bot instance from the database
func DeleteBotInstance(botID int64) error {
	err := DB.Where("bot_id = ?", botID).Delete(&BotInstance{}).Error
	if err != nil {
		log.Errorf("[Database] DeleteBotInstance: %v", err)
		return err
	}
	
	log.Infof("[Database] Deleted bot instance %d", botID)
	return nil
}

// CountUserBotInstances returns the number of bot instances owned by a user
func CountUserBotInstances(ownerID int64) (int64, error) {
	var count int64
	err := DB.Model(&BotInstance{}).Where("owner_id = ?", ownerID).Count(&count).Error
	if err != nil {
		log.Errorf("[Database] CountUserBotInstances: %v", err)
		return 0, err
	}
	return count, nil
}

// CountTotalBotInstances returns the total number of bot instances
func CountTotalBotInstances() (int64, error) {
	var count int64
	err := DB.Model(&BotInstance{}).Count(&count).Error
	if err != nil {
		log.Errorf("[Database] CountTotalBotInstances: %v", err)
		return 0, err
	}
	return count, nil
}

// CountActiveBotInstances returns the number of active bot instances
func CountActiveBotInstances() (int64, error) {
	var count int64
	err := DB.Model(&BotInstance{}).Where("is_active = ?", true).Count(&count).Error
	if err != nil {
		log.Errorf("[Database] CountActiveBotInstances: %v", err)
		return 0, err
	}
	return count, nil
}

// GetBotInstancesByActivity returns bot instances sorted by last activity
func GetBotInstancesByActivity(limit int) ([]BotInstance, error) {
	var instances []BotInstance
	query := DB.Where("last_activity IS NOT NULL").Order("last_activity DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&instances).Error
	if err != nil {
		log.Errorf("[Database] GetBotInstancesByActivity: %v", err)
		return nil, err
	}
	return instances, nil
}

// GetInactiveBotInstances returns bot instances that haven't been active for a specified duration
func GetInactiveBotInstances(inactiveDuration time.Duration) ([]BotInstance, error) {
	cutoffTime := time.Now().Add(-inactiveDuration)
	
	var instances []BotInstance
	err := DB.Where("last_activity < ? OR last_activity IS NULL", cutoffTime).Find(&instances).Error
	if err != nil {
		log.Errorf("[Database] GetInactiveBotInstances: %v", err)
		return nil, err
	}
	return instances, nil
}

// BotInstanceExists checks if a bot instance exists with the given bot ID
func BotInstanceExists(botID int64) bool {
	var count int64
	err := DB.Model(&BotInstance{}).Where("bot_id = ?", botID).Count(&count).Error
	if err != nil {
		log.Errorf("[Database] BotInstanceExists: %v", err)
		return false
	}
	return count > 0
}

// ValidateTokenHash checks if the provided token hash matches the stored hash for a bot instance
func ValidateTokenHash(botID int64, tokenHash string) (bool, error) {
	var instance BotInstance
	err := DB.Select("token_hash").Where("bot_id = ?", botID).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		log.Errorf("[Database] ValidateTokenHash: %v", err)
		return false, err
	}
	
	return instance.TokenHash == tokenHash, nil
}

// GetBotInstanceStats returns comprehensive statistics about bot instances
func GetBotInstanceStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total count
	totalCount, err := CountTotalBotInstances()
	if err != nil {
		return nil, err
	}
	stats["total"] = totalCount
	
	// Active count
	activeCount, err := CountActiveBotInstances()
	if err != nil {
		return nil, err
	}
	stats["active"] = activeCount
	stats["inactive"] = totalCount - activeCount
	
	// Users with bots
	var userCount int64
	err = DB.Model(&BotInstance{}).Select("DISTINCT owner_id").Count(&userCount).Error
	if err != nil {
		log.Errorf("[Database] GetBotInstanceStats - user count: %v", err)
		return nil, err
	}
	stats["users_with_bots"] = userCount
	
	// Recently active (last 24 hours)
	var recentlyActive int64
	cutoff24h := time.Now().Add(-24 * time.Hour)
	err = DB.Model(&BotInstance{}).Where("last_activity > ?", cutoff24h).Count(&recentlyActive).Error
	if err != nil {
		log.Errorf("[Database] GetBotInstanceStats - recently active: %v", err)
		return nil, err
	}
	stats["recently_active_24h"] = recentlyActive
	
	return stats, nil
}