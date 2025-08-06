package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetChatLocks retrieves all lock settings for a specific chat ID.
// Uses optimized queries with caching for better performance.
// Returns an empty map if no locks are found or an error occurs.
func GetChatLocks(chatID int64) map[string]bool {
	// Use optimized query with caching
	locks, err := GetOptimizedQueries().lockQueries.GetChatLocksOptimized(chatID)
	if err != nil {
		log.Errorf("[Database] GetChatLocks: %v - %d", err, chatID)
		return make(map[string]bool)
	}

	return locks
}

// MapLockType is an alias for GetChatLocks that returns the lock settings for a chat.
// Provided for backward compatibility and alternative naming convention.
func MapLockType(chatID int64) map[string]bool {
	return GetChatLocks(chatID)
}

// UpdateLock modifies the value of a specific lock setting and updates it in the database.
// Creates a new lock record if one doesn't exist for the given chat and permission type.
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

// IsPermLocked checks whether a specific permission type is locked in the given chat.
// Uses optimized cached queries for better performance.
// Returns false if the permission is not locked or an error occurs.
func IsPermLocked(chatID int64, perm string) bool {
	// Use optimized cached query
	locked, err := GetOptimizedQueries().GetLockStatusCached(chatID, perm)
	if err != nil {
		log.Errorf("[Database] IsPermLocked: %v - %d", err, chatID)
		return false
	}

	return locked
}
