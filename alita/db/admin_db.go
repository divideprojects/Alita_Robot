package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetAdminSettings Get admin settings for a chat
func GetAdminSettings(chatID int64) *AdminSettings {
	return checkAdminSetting(chatID)
}

// check Chat Admin Settings, used to get data before performing any operation
func checkAdminSetting(chatID int64) (adminSrc *AdminSettings) {
	adminSrc = &AdminSettings{}

	err := GetRecord(adminSrc, AdminSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		adminSrc = &AdminSettings{ChatId: chatID, AnonAdmin: false}
		err := CreateRecord(adminSrc)
		if err != nil {
			log.Errorf("[Database][checkAdminSetting]: %v ", err)
		}
	} else if err != nil {
		// Return default on error
		adminSrc = &AdminSettings{ChatId: chatID, AnonAdmin: false}
		log.Errorf("[Database][checkAdminSetting]: %v ", err)
	}
	return adminSrc
}

// SetAnonAdminMode Set anon admin mode for a chat
func SetAnonAdminMode(chatID int64, val bool) {
	adminSrc := checkAdminSetting(chatID)
	adminSrc.AnonAdmin = val

	err := UpdateRecord(&AdminSettings{}, AdminSettings{ChatId: chatID}, AdminSettings{AnonAdmin: val})
	if err != nil {
		log.Errorf("[Database] SetAnonAdminMode: %v - %d", err, chatID)
	}
}
