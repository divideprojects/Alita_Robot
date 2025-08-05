package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func checkWarnSettings(chatID int64) (warnrc *WarnSettings) {
	defaultWarnSettings := &WarnSettings{ChatId: chatID, WarnLimit: 3, WarnMode: "mute"}
	warnrc = &WarnSettings{}
	err := DB.Where("chat_id = ?", chatID).First(warnrc)
	if errors.Is(err.Error, gorm.ErrRecordNotFound) {
		warnrc = defaultWarnSettings
		err := DB.Create(warnrc)
		if err.Error != nil {
			log.Errorf("[Database] checkWarnSettings: %v", err.Error)
		}
	} else if err.Error != nil {
		log.Errorf("[Database][checkWarnSettings]: %d - %v", chatID, err.Error)
		warnrc = defaultWarnSettings
	}
	return
}

func checkWarns(userId, chatId int64) (warnrc *Warns) {
	defaultWarnSrc := &Warns{UserId: userId, ChatId: chatId, NumWarns: 0, Reasons: make(StringArray, 0)}
	warnrc = &Warns{}
	err := DB.Where("user_id = ? AND chat_id = ?", userId, chatId).First(warnrc)
	if errors.Is(err.Error, gorm.ErrRecordNotFound) {
		warnrc = defaultWarnSrc
		err := DB.Create(warnrc)
		if err.Error != nil {
			log.Errorf("[Database] checkWarns: %v", err.Error)
		}
	} else if err.Error != nil {
		log.Errorf("[Database][checkUserWarns]: %d - %v", userId, err.Error)
		warnrc = defaultWarnSrc
	}
	return
}

func WarnUser(userId, chatId int64, reason string) (int, []string) {
	warnrc := checkWarns(userId, chatId)

	warnrc.NumWarns++ // Increment warns - Add 1 warn

	// Add reason if it exists
	if reason != "" {
		if len(reason) >= 3001 {
			reason = reason[:3000]
		}
		warnrc.Reasons = append(warnrc.Reasons, reason)
	} else {
		warnrc.Reasons = append(warnrc.Reasons, "No Reason")
	}

	err := DB.Save(warnrc)
	if err.Error != nil {
		log.Errorf("[Database] WarnUser: %v", err.Error)
	}

	return warnrc.NumWarns, []string(warnrc.Reasons)
}

func RemoveWarn(userId, chatId int64) bool {
	removed := false
	warnrc := checkWarns(userId, chatId)

	// only remove if user has warns
	if warnrc.NumWarns > 0 {
		warnrc.NumWarns--                                       // Remove last warn num
		warnrc.Reasons = warnrc.Reasons[:len(warnrc.Reasons)-1] // Remove last warn reason
		removed = true
	}

	// update record in db
	err := DB.Save(warnrc)
	if err.Error != nil {
		log.Errorf("[Database] RemoveWarn: %v", err.Error)
		return false // force return false to show error
	}

	return removed
}

func ResetUserWarns(userId, chatId int64) (removed bool) {
	removed = true
	err := DB.Where("user_id = ? AND chat_id = ?", userId, chatId).Delete(&Warns{})
	if err.Error != nil {
		log.Errorf("[Database] ResetUserWarns: %v", err.Error)
		removed = false
	}
	return removed
}

func GetWarns(userId, chatId int64) (int, []string) {
	warnrc := checkWarns(userId, chatId)
	return warnrc.NumWarns, []string(warnrc.Reasons)
}

func SetWarnLimit(chatId int64, warnLimit int) {
	warnrc := checkWarnSettings(chatId)
	warnrc.WarnLimit = warnLimit
	err := DB.Save(warnrc)
	if err.Error != nil {
		log.Errorf("[Database] SetWarnLimit: %v", err.Error)
	}
}

func SetWarnMode(chatId int64, warnMode string) {
	warnrc := checkWarnSettings(chatId)
	warnrc.WarnMode = warnMode
	err := DB.Save(warnrc)
	if err.Error != nil {
		log.Errorf("[Database] SetWarnMode: %v", err.Error)
	}
}

func GetWarnSetting(chatId int64) *WarnSettings {
	return checkWarnSettings(chatId)
}

func GetAllChatWarns(chatId int64) int {
	var count int64
	err := DB.Model(&Warns{}).Where("chat_id = ?", chatId).Count(&count)
	if err.Error != nil {
		log.Errorf("[Database] GetAllChatWarns: %v", err.Error)
		return 0
	}
	return int(count)
}

func ResetAllChatWarns(chatId int64) bool {
	err := DB.Where("chat_id = ?", chatId).Delete(&Warns{})
	if err.Error != nil {
		log.Errorf("[Database] ResetAllChatWarns: %v", err.Error)
		return false
	}
	return true
}
