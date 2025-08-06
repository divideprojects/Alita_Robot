package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetChannelSettings retrieves channel settings from cache or database.
// Returns nil if the channel is not found or an error occurs.
func GetChannelSettings(channelId int64) (channelSrc *ChannelSettings) {
	// Use optimized cached query instead of SELECT *
	channelSrc, err := GetOptimizedQueries().GetChannelSettingsCached(channelId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Errorf("[Database] GetChannelSettings: %v - %d", err, channelId)
		return nil
	}
	return channelSrc
}

// EnsureChatExists ensures a chat record exists before creating related records.
// Creates a minimal chat record with default settings if it doesn't exist.
func EnsureChatExists(chatId int64, chatName string) error {
	if ChatExists(chatId) {
		return nil
	}

	// Create minimal chat record
	chat := &Chat{
		ChatId:     chatId,
		ChatName:   chatName,
		Language:   "en", // default language
		Users:      Int64Array{},
		IsInactive: false,
	}

	err := CreateRecord(chat)
	if err != nil {
		log.Errorf("[Database] EnsureChatExists: Failed to create chat %d: %v", chatId, err)
		return err
	}

	log.Infof("[Database] EnsureChatExists: Created chat record for %d", chatId)
	return nil
}

// UpdateChannel updates or creates a channel record.
// Ensures the chat exists before creating channel settings and invalidates cache after updates.
func UpdateChannel(channelId int64, channelName, username string) {
	// Check if channel already exists
	channelSrc := GetChannelSettings(channelId)

	if channelSrc != nil && channelSrc.ChannelId == channelId {
		// Channel already exists with same ID, no update needed
		return
	}

	// Ensure the chat exists before creating/updating channel
	if err := EnsureChatExists(channelId, channelName); err != nil {
		log.Errorf("[Database] UpdateChannel: Failed to ensure chat exists for %d (%s): %v", channelId, username, err)
		return
	}

	if channelSrc == nil {
		// Create new channel - Note: The original Channel struct doesn't map well to ChannelSettings
		// ChannelSettings is for chat->channel mapping, not channel info storage
		channelSrc = &ChannelSettings{
			ChatId:    channelId,
			ChannelId: channelId, // Assuming this is the mapping
		}
		err := CreateRecord(channelSrc)
		if err != nil {
			log.Errorf("[Database] UpdateChannel: %v - %d (%s)", err, channelId, username)
			return
		}
		// Invalidate cache after create
		deleteCache(channelCacheKey(channelId))
		log.Debugf("[Database] UpdateChannel: created channel %d", channelId)
	}
}

// GetChannelIdByUserName attempts to find a channel ID by username.
// Returns 0 as this function is not supported with the current model structure.
func GetChannelIdByUserName(username string) int64 {
	// Note: The new ChannelSettings model doesn't store username
	// This function cannot be implemented with the current model structure
	log.Warnf("[Database] GetChannelIdByUserName: Function not supported with current model structure")
	return 0
}

// GetChannelInfoById retrieves channel information by user ID.
// Returns empty strings for username and name as the current model doesn't store this data.
func GetChannelInfoById(userId int64) (username, name string, found bool) {
	channel := GetChannelSettings(userId)
	if channel != nil {
		// Note: The new model doesn't store username/name, only IDs
		username = ""
		name = ""
		found = true
	}
	return
}

// LoadChannelStats returns the total count of channel settings records.
func LoadChannelStats() (count int64) {
	err := DB.Model(&ChannelSettings{}).Count(&count).Error
	if err != nil {
		log.Errorf("[Database] loadChannelStats: %v", err)
		return 0
	}
	return
}
