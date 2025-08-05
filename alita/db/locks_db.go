package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetChatLocks(chatID int64) map[string]bool {
	var lockSettings []*LockSettings
	err := GetRecords(&lockSettings, LockSettings{ChatId: chatID})
	if err != nil {
		log.Errorf("[Database] GetChatLocks: %v - %d", err, chatID)
		return make(map[string]bool)
	}

	locks := make(map[string]bool)
	for _, setting := range lockSettings {
		locks[setting.LockType] = setting.Locked
	}

	return locks
}

func MapLockType(chatID int64) map[string]bool {
	return GetChatLocks(chatID)
}

// UpdateLock Modify the value of lock setting and update it in database
func UpdateLock(chatID int64, perm string, val bool) {
	lockSetting := &LockSettings{
		ChatId:   chatID,
		LockType: perm,
		Locked:   val,
	}

	// Try to update existing record first
	err := UpdateRecord(&LockSettings{}, LockSettings{ChatId: chatID, LockType: perm}, LockSettings{Locked: val})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new record if not exists
		err = CreateRecord(lockSetting)
	}

	if err != nil {
		log.Errorf("[Database] UpdateLock: %v", err)
	}
}

func IsPermLocked(chatID int64, perm string) bool {
	var lockSetting LockSettings
	err := GetRecord(&lockSetting, LockSettings{ChatId: chatID, LockType: perm})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false // Default to unlocked if not found
	} else if err != nil {
		log.Errorf("[Database] IsPermLocked: %v - %d", err, chatID)
		return false
	}

	return lockSetting.Locked
}
