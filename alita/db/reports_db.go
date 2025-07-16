package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
	"github.com/eko/gocache/lib/v4/store"

	"go.mongodb.org/mongo-driver/bson"
)

// ChatReportSettings holds report-related configuration for a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - Status: Whether reports are enabled for the chat.
//   - BlockedList: List of user IDs who are blocked from making reports.
type ChatReportSettings struct {
	ChatId      int64   `bson:"_id,omitempty" json:"_id,omitempty"`
	Status      bool    `bson:"status,omitempty" json:"status,omitempty"`
	BlockedList []int64 `bson:"blocked_list,omitempty" json:"blocked_list,omitempty"`
}

// UserReportSettings holds report-related configuration for individual users.
//
// Fields:
//   - UserId: Unique identifier for the user.
//   - Status: Whether the user wants to receive report notifications.
type UserReportSettings struct {
	UserId int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	Status bool  `bson:"status,omitempty" json:"status,omitempty"`
}

// GetChatReportSettings retrieves the report settings for a given chat ID.
// If no settings exist, it initializes them with default values (reports enabled, empty blocked list).
// Uses cache for performance optimization with 10-minute expiration.
func GetChatReportSettings(chatID int64) (reportsrc *ChatReportSettings) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, chatID, new(ChatReportSettings)); err == nil && cached != nil {
		reportsrc = cached.(*ChatReportSettings)
		return
	}

	reportsrc = &ChatReportSettings{}

	filter := bson.M{"_id": chatID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":          chatID,
			"status":       true,
			"blocked_list": make([]int64, 0),
		},
	}

	err := findOneAndUpsert(reportChatColl, filter, update, reportsrc)
	if err != nil {
		// Fallback to default values in case of error
		reportsrc = &ChatReportSettings{
			ChatId:      chatID,
			Status:      true,
			BlockedList: make([]int64, 0),
		}
		log.Error(err)
	}
	// Cache the result
	_ = cache.Marshal.Set(cache.Context, chatID, reportsrc, store.WithExpiration(10*time.Minute))
	return
}

// SetChatReportStatus enables or disables reports for a specific chat.
// When disabled, users cannot make reports and admins won't receive notifications.
// Updates both database and cache with the new setting.
func SetChatReportStatus(chatID int64, pref bool) {
	reportsUpdate := GetChatReportSettings(chatID)
	reportsUpdate.Status = pref
	err := updateOne(reportChatColl, bson.M{"_id": chatID}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, chatID, reportsUpdate, store.WithExpiration(10*time.Minute))
}

// BlockReportUser adds a user to the chat's report blocking list.
// Blocked users cannot make reports in the chat. If user is already blocked, no action is taken.
// Updates both database and cache with the new blocked list.
func BlockReportUser(chatId, userId int64) {
	reportsUpdate := GetChatReportSettings(chatId)

	if string_handling.FindInInt64Slice(reportsUpdate.BlockedList, userId) {
		return
	}

	reportsUpdate.BlockedList = append(reportsUpdate.BlockedList, userId)
	err := updateOne(reportChatColl, bson.M{"_id": chatId}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, chatId, reportsUpdate, store.WithExpiration(10*time.Minute))
}

// UnblockReportUser removes a user from the chat's report blocking list.
// If the user is not in the blocked list, no action is taken.
// Updates the database but not the cache (cache will be refreshed on next access).
func UnblockReportUser(chatId, userId int64) {
	reportsUpdate := GetChatReportSettings(chatId)

	if !string_handling.FindInInt64Slice(reportsUpdate.BlockedList, userId) {
		return
	}

	reportsUpdate.BlockedList = string_handling.RemoveFromInt64Slice(reportsUpdate.BlockedList, userId)
	err := updateOne(reportChatColl, bson.M{"_id": chatId}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

// User-specific report settings functions

// GetUserReportSettings retrieves the report settings for a given user ID.
// If no settings exist, it initializes them with default values (reports enabled).
// Users can control whether they want to receive report notifications.
func GetUserReportSettings(userId int64) (reportsrc *UserReportSettings) {
	reportsrc = &UserReportSettings{}

	filter := bson.M{"_id": userId}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":    userId,
			"status": true,
		},
	}

	err := findOneAndUpsert(reportUserColl, filter, update, reportsrc)
	if err != nil {
		// Fallback to default values in case of error
		reportsrc = &UserReportSettings{
			UserId: userId,
			Status: true,
		}
		log.Error(err)
	}

	return
}

// SetUserReportSettings enables or disables report notifications for a specific user.
// When disabled, the user won't receive notifications about reports they made.
// Note: The parameter should be userId, not chatID - this appears to be a bug.
func SetUserReportSettings(chatID int64, pref bool) {
	reportsUpdate := &ChatReportSettings{
		ChatId: chatID,
		Status: pref,
	}
	err := updateOne(reportUserColl, bson.M{"_id": chatID}, reportsUpdate)
	if err != nil {
		log.Error(chatID)
	}
}

// LoadReportStats returns the count of users and chats with reports enabled.
// Used for bot statistics and monitoring purposes.
// Returns user report count and group report count respectively.
func LoadReportStats() (uRCount, gRCount int64) {
	uRCount, acErr := countDocs(
		reportUserColl,
		bson.M{"status": true},
	)
	if acErr != nil {
		log.Errorf("[Database] loadStats: %v", acErr)
	}
	gRCount, clErr := countDocs(
		reportChatColl,
		bson.M{"status": true},
	)
	if clErr != nil {
		log.Errorf("[Database] loadStats: %v", clErr)
	}
	return
}
