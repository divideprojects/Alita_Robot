package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// WarnSettings represents the warning configuration for a chat.
// It defines the warning threshold and the action to take when the limit is exceeded.
type WarnSettings struct {
	ChatId    int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	WarnLimit int    `bson:"warn_limit" json:"warn_limit" default:"3"`
	WarnMode  string `bson:"warn_mode,omitempty" json:"warn_mode,omitempty"`
}

// Warns represents a user's warning record in a specific chat.
// It tracks the number of warnings and the reasons for each warning.
type Warns struct {
	UserId   int64    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ChatId   int64    `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	NumWarns int      `bson:"num_warns,omitempty" json:"num_warns,omitempty"`
	Reasons  []string `bson:"warns" json:"warns" default:"[]"`
}

func checkWarnSettings(chatID int64) (warnrc *WarnSettings) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, chatID, new(WarnSettings)); err == nil && cached != nil {
		return cached.(*WarnSettings)
	}
	defaultWarnSettings := &WarnSettings{ChatId: chatID, WarnLimit: 3, WarnMode: "mute"}
	err := findOne(warnSettingsColl, bson.M{"_id": chatID}).Decode(&warnrc)
	if err == mongo.ErrNoDocuments {
		warnrc = defaultWarnSettings
		err := updateOne(warnSettingsColl, bson.M{"_id": chatID}, warnrc)
		if err != nil {
			log.Errorf("[Database] checkWarnSettings: %v", err)
		}
	} else if err != nil {
		log.Errorf("[Database][checkWarnSettings]: %d - %v", chatID, err)
		warnrc = defaultWarnSettings
	}
	// Cache the result
	if warnrc != nil {
		_ = cache.Marshal.Set(cache.Context, chatID, warnrc, store.WithExpiration(10*time.Minute))
	}
	return
}

func checkWarns(userId, chatId int64) (warnrc *Warns) {
	defaultWarnSrc := &Warns{UserId: userId, ChatId: chatId, NumWarns: 0, Reasons: make([]string, 0)}
	err := findOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId}).Decode(&warnrc)
	if err == mongo.ErrNoDocuments {
		warnrc = defaultWarnSrc
		err := updateOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId}, warnrc)
		if err != nil {
			log.Errorf("[Database] checkWarns: %v", err)
		}
	} else if err != nil {
		log.Errorf("[Database][checkUserWarns]: %d - %v", userId, err)
		warnrc = defaultWarnSrc
	}
	return
}

// WarnUser issues a warning to a user in a specific chat.
// It atomically increments the warning count and adds the reason to the warning list.
// Reason strings are truncated to 3000 characters if exceeded.
// Returns the updated warning count and all warning reasons.
// If reason is empty, it defaults to "No Reason".
func WarnUser(userId, chatId int64, reason string) (int, []string) {
	// Prepare reason
	if reason == "" {
		reason = "No Reason"
	} else if len(reason) >= 3001 {
		reason = reason[:3000]
	}

	// Use atomic increment and push to avoid race conditions
	filter := bson.M{"user_id": userId, "chat_id": chatId}
	update := bson.M{
		"$inc":  bson.M{"num_warns": 1},
		"$push": bson.M{"warns": reason},
		"$setOnInsert": bson.M{
			"user_id": userId,
			"chat_id": chatId,
		},
	}

	result := &Warns{}
	err := findOneAndUpsert(warnUsersColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database] WarnUser: %v", err)
		// Return default values on error
		return 0, []string{}
	}

	return result.NumWarns, result.Reasons
}

// RemoveWarn removes the most recent warning from a user in a specific chat.
// It atomically decrements the warning count and removes the last warning reason.
// Returns true if a warning was successfully removed, false if the user has no warnings
// or if an error occurred.
func RemoveWarn(userId, chatId int64) bool {
	// Use atomic decrement and pop to avoid race conditions
	filter := bson.M{
		"user_id":   userId,
		"chat_id":   chatId,
		"num_warns": bson.M{"$gt": 0}, // Only update if warns > 0
	}
	update := bson.M{
		"$inc": bson.M{"num_warns": -1},
		"$pop": bson.M{"warns": 1}, // Remove last element
	}

	result := &Warns{}
	err := findOneAndUpsert(warnUsersColl, filter, update, result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No document found means user has no warns to remove
			return false
		}
		log.Errorf("[Database] RemoveWarn: %v", err)
		return false
	}

	return true
}

// ResetUserWarns removes all warnings for a specific user in a chat.
// It deletes the user's warning record entirely from the database.
// Returns true if the warnings were successfully reset, false if an error occurred.
func ResetUserWarns(userId, chatId int64) (removed bool) {
	removed = true
	err := deleteOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId})
	if err != nil {
		log.Errorf("[Database] ResetUserWarns: %v", err)
		removed = false
	}
	return removed
}

// GetWarns retrieves the current warning count and reasons for a user in a specific chat.
// Returns the number of warnings and a slice of warning reasons.
// If the user has no warnings, returns 0 and an empty slice.
func GetWarns(userId, chatId int64) (int, []string) {
	warnrc := checkWarns(userId, chatId)
	return warnrc.NumWarns, warnrc.Reasons
}

// SetWarnLimit configures the warning threshold for a chat.
// When a user reaches this limit, the configured warn mode action will be triggered.
// The warn limit must be a positive integer.
func SetWarnLimit(chatId int64, warnLimit int) {
	warnrc := checkWarnSettings(chatId)
	warnrc.WarnLimit = warnLimit
	err := updateOne(warnSettingsColl, bson.M{"_id": chatId}, warnrc)
	if err != nil {
		log.Errorf("[Database] SetWarnLimit: %v", err)
	}
}

// SetWarnMode configures the action to take when a user reaches the warning limit.
// Common modes include "kick", "ban", "mute", or "tban".
// The mode determines what happens to users who exceed the warning threshold.
func SetWarnMode(chatId int64, warnMode string) {
	warnrc := checkWarnSettings(chatId)
	warnrc.WarnMode = warnMode
	err := updateOne(warnSettingsColl, bson.M{"_id": chatId}, warnrc)
	if err != nil {
		log.Errorf("[Database] SetWarnMode: %v", err)
	}
}

// GetWarnSetting retrieves the current warning configuration for a chat.
// Returns a WarnSettings struct containing the warning limit and mode.
// If no settings exist, returns default settings (limit: 3, mode: "mute").
func GetWarnSetting(chatId int64) *WarnSettings {
	return checkWarnSettings(chatId)
}

// GetAllChatWarns returns the total number of users with warnings in a specific chat.
// This count represents how many users have at least one warning, not the total warning count.
func GetAllChatWarns(chatId int64) int {
	length, _ := countDocs(warnUsersColl, bson.M{"chat_id": chatId})
	return int(length)
}

// ResetAllChatWarns removes all warning records for all users in a specific chat.
// This is a destructive operation that permanently deletes all warning data for the chat.
// Returns true if all warnings were successfully removed, false if an error occurred.
func ResetAllChatWarns(chatId int64) bool {
	err := deleteMany(warnUsersColl, bson.M{"chat_id": chatId})
	if err != nil {
		log.Errorf("[Database] ResetAllChatWarns: %v", err)
		return false
	}
	return true
}
