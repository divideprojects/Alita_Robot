package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// default mode is 'mute'
const defaultFloodsettingsMode string = "mute"

// GetFlood Get flood settings for a chat
func GetFlood(chatID int64) *AntifloodSettings {
	return checkFloodSetting(chatID)
}

// check Chat Flood Settings, used to get data before performing any operation
func checkFloodSetting(chatID int64) (floodSrc *AntifloodSettings) {
	floodSrc = &AntifloodSettings{}

	err := GetRecord(floodSrc, AntifloodSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		floodSrc = &AntifloodSettings{ChatId: chatID, Limit: 0, Action: defaultFloodsettingsMode}
		err := CreateRecord(floodSrc)
		if err != nil {
			log.Errorf("[Database][checkFloodSetting]: %v ", err)
		}
	} else if err != nil {
		// Return default on error
		floodSrc = &AntifloodSettings{ChatId: chatID, Limit: 0, Action: defaultFloodsettingsMode}
		log.Errorf("[Database][checkFloodSetting]: %v ", err)
	}
	return floodSrc
}

// SetFlood set Flood Setting for a Chat
func SetFlood(chatID int64, limit int) {
	floodSrc := checkFloodSetting(chatID)

	if floodSrc.Action == "" {
		floodSrc.Action = defaultFloodsettingsMode
	}
	floodSrc.Limit = limit

	// update the value in db
	err := UpdateRecord(&AntifloodSettings{}, AntifloodSettings{ChatId: chatID}, AntifloodSettings{Limit: limit, Action: floodSrc.Action})
	if err != nil {
		log.Errorf("[Database] SetFlood: %v - %d", err, chatID)
	}
}

// SetFloodMode Set flood mode for a chat
func SetFloodMode(chatID int64, mode string) {
	err := UpdateRecord(&AntifloodSettings{}, AntifloodSettings{ChatId: chatID}, AntifloodSettings{Action: mode})
	if err != nil {
		log.Errorf("[Database] SetFloodMode: %v - %d", err, chatID)
	}
}

// SetFloodMsgDel Set flood message deletion setting for a chat
func SetFloodMsgDel(chatID int64, val bool) {
	floodSrc := checkFloodSetting(chatID)
	err := UpdateRecord(&AntifloodSettings{}, AntifloodSettings{ChatId: chatID}, AntifloodSettings{DeleteAntifloodMessage: val})
	if err != nil {
		log.Errorf("[Database] SetFloodMsgDel: %v", err)
		return
	}
	floodSrc.DeleteAntifloodMessage = val
}

func LoadAntifloodStats() (antiCount int64) {
	var totalCount int64
	var noAntiCount int64

	// Count total antiflood settings
	err := DB.Model(&AntifloodSettings{}).Count(&totalCount).Error
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
		return 0
	}

	// Count settings with limit 0 (disabled)
	err = DB.Model(&AntifloodSettings{}).Where("flood_limit = ?", 0).Count(&noAntiCount).Error
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
		return 0
	}

	antiCount = totalCount - noAntiCount // gives chats which have enabled anti flood

	return
}
