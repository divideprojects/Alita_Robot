package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func ToggleAllowConnect(chatID int64, pref bool) {
	err := UpdateRecord(&ConnectionChatSettings{}, ConnectionChatSettings{ChatId: chatID}, ConnectionChatSettings{Enabled: pref})
	if err != nil {
		log.Errorf("[Database] ToggleAllowConnect: %d - %v", chatID, err)
	}
}

func GetChatConnectionSetting(chatID int64) (connectionSrc *ConnectionChatSettings) {
	connectionSrc = &ConnectionChatSettings{}
	err := GetRecord(connectionSrc, ConnectionChatSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		connectionSrc = &ConnectionChatSettings{ChatId: chatID, Enabled: false}
		err := CreateRecord(connectionSrc)
		if err != nil {
			log.Errorf("[Database] GetChatConnectionSetting: %d - %v", chatID, err)
		}
	} else if err != nil {
		// Return default on error
		connectionSrc = &ConnectionChatSettings{ChatId: chatID, Enabled: false}
		log.Errorf("[Database] GetChatConnectionSetting: %d - %v", chatID, err)
	}
	return connectionSrc
}

func getUserConnectionSetting(userID int64) (connectionSrc *ConnectionSettings) {
	connectionSrc = &ConnectionSettings{}
	err := GetRecord(connectionSrc, ConnectionSettings{UserId: userID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		connectionSrc = &ConnectionSettings{UserId: userID, Connected: false}
		err := CreateRecord(connectionSrc)
		if err != nil {
			log.Errorf("[Database] getUserConnectionSetting: %d - %v", userID, err)
		}
	} else if err != nil {
		// Return default on error
		connectionSrc = &ConnectionSettings{UserId: userID, Connected: false}
		log.Errorf("[Database] getUserConnectionSetting: %d - %v", userID, err)
	}

	return connectionSrc
}

func Connection(UserID int64) *ConnectionSettings {
	return getUserConnectionSetting(UserID)
}

func ConnectId(UserID, chatID int64) {
	err := UpdateRecord(&ConnectionSettings{}, ConnectionSettings{UserId: UserID}, ConnectionSettings{Connected: true, ChatId: chatID})
	if err != nil {
		log.Errorf("[Database] ConnectId: %v - %d", err, chatID)
	}
}

func DisconnectId(UserID int64) {
	err := UpdateRecord(&ConnectionSettings{}, ConnectionSettings{UserId: UserID}, ConnectionSettings{Connected: false})
	if err != nil {
		log.Errorf("[Database] DisconnectId: %v - %d", err, UserID)
	}
}

func ReconnectId(UserID int64) int64 {
	connectionUpdate := Connection(UserID)
	err := UpdateRecord(&ConnectionSettings{}, ConnectionSettings{UserId: UserID}, ConnectionSettings{Connected: true})
	if err != nil {
		log.Errorf("[Database] ReconnectId: %v - %d", err, UserID)
		return 0
	}
	return connectionUpdate.ChatId
}

func LoadConnectionStats() (connectedUsers, connectedChats int64) {
	// Count chats that allow connections
	err := DB.Model(&ConnectionChatSettings{}).Where("enabled = ?", true).Count(&connectedChats).Error
	if err != nil {
		log.Error(err)
		return
	}

	// Count connected users
	err = DB.Model(&ConnectionSettings{}).Where("connected = ?", true).Count(&connectedUsers).Error
	if err != nil {
		log.Error(err)
		return
	}

	return
}
