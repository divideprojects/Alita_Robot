package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetChannelSettings(channelId int64) (channelSrc *ChannelSettings) {
	channelSrc = &ChannelSettings{}
	err := GetRecord(channelSrc, ChannelSettings{ChatId: channelId})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		log.Errorf("[Database] getChannelSettings: %v - %d ", err, channelId)
		return nil
	}
	return channelSrc
}

func UpdateChannel(channelId int64, channelName, username string) {
	channelSrc := GetChannelSettings(channelId)

	if channelSrc != nil {
		// Update existing channel
		err := UpdateRecord(&ChannelSettings{}, ChannelSettings{ChatId: channelId}, ChannelSettings{ChannelId: channelSrc.ChannelId})
		if err != nil {
			log.Errorf("[Database] UpdateChannel: %v - %d (%s)", err, channelId, username)
			return
		}
	} else {
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
	}
	log.Infof("[Database] UpdateChannel: channel %d", channelId)
}

func GetChannelIdByUserName(username string) int64 {
	// Note: The new ChannelSettings model doesn't store username
	// This function cannot be implemented with the current model structure
	log.Warnf("[Database] GetChannelIdByUserName: Function not supported with current model structure")
	return 0
}

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

func LoadChannelStats() (count int64) {
	err := DB.Model(&ChannelSettings{}).Count(&count).Error
	if err != nil {
		log.Errorf("[Database] loadChannelStats: %v", err)
		return 0
	}
	return
}
