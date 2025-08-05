package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetChatReportSettings(chatID int64) (reportsrc *ReportChatSettings) {
	reportsrc = &ReportChatSettings{}
	err := GetRecord(reportsrc, ReportChatSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		reportsrc = &ReportChatSettings{ChatId: chatID, Enabled: true}
		err := CreateRecord(reportsrc)
		if err != nil {
			log.Error(err)
		}
	} else if err != nil {
		// Return default on error
		reportsrc = &ReportChatSettings{ChatId: chatID, Enabled: true}
		log.Error(err)
	}
	return
}

func SetChatReportStatus(chatID int64, pref bool) {
	err := UpdateRecord(&ReportChatSettings{}, ReportChatSettings{ChatId: chatID}, ReportChatSettings{Enabled: pref})
	if err != nil {
		log.Error(err)
	}
}

func BlockReportUser(chatId, userId int64) {
	settings := GetChatReportSettings(chatId)

	// Check if user is already blocked
	for _, blockedId := range settings.BlockedList {
		if blockedId == userId {
			return // User already blocked
		}
	}

	// Add user to blocked list
	settings.BlockedList = append(settings.BlockedList, userId)
	err := UpdateRecord(&ReportChatSettings{}, ReportChatSettings{ChatId: chatId}, ReportChatSettings{BlockedList: settings.BlockedList})
	if err != nil {
		log.Errorf("[Database] BlockReportUser: %v", err)
	}
}

func UnblockReportUser(chatId, userId int64) {
	settings := GetChatReportSettings(chatId)

	// Remove user from blocked list
	var newBlockedList Int64Array
	for _, blockedId := range settings.BlockedList {
		if blockedId != userId {
			newBlockedList = append(newBlockedList, blockedId)
		}
	}

	err := UpdateRecord(&ReportChatSettings{}, ReportChatSettings{ChatId: chatId}, ReportChatSettings{BlockedList: newBlockedList})
	if err != nil {
		log.Errorf("[Database] UnblockReportUser: %v", err)
	}
}

/* user settings below */

func GetUserReportSettings(userId int64) (reportsrc *ReportUserSettings) {
	reportsrc = &ReportUserSettings{}
	err := GetRecord(reportsrc, ReportUserSettings{UserId: userId})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		reportsrc = &ReportUserSettings{UserId: userId, Enabled: true}
		err := CreateRecord(reportsrc)
		if err != nil {
			log.Error(err)
		}
	} else if err != nil {
		// Return default on error
		reportsrc = &ReportUserSettings{UserId: userId, Enabled: true}
		log.Error(err)
	}

	return
}

func SetUserReportSettings(userID int64, pref bool) {
	err := UpdateRecord(&ReportUserSettings{}, ReportUserSettings{UserId: userID}, ReportUserSettings{Enabled: pref})
	if err != nil {
		log.Error(userID)
	}
}

func LoadReportStats() (uRCount, gRCount int64) {
	// Count users with reports enabled
	err := DB.Model(&ReportUserSettings{}).Where("enabled = ?", true).Count(&uRCount).Error
	if err != nil {
		log.Errorf("[Database] LoadReportStats (users): %v", err)
	}

	// Count chats with reports enabled
	err = DB.Model(&ReportChatSettings{}).Where("enabled = ?", true).Count(&gRCount).Error
	if err != nil {
		log.Errorf("[Database] LoadReportStats (chats): %v", err)
	}

	return
}
