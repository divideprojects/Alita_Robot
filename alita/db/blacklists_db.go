package db

import (
	"context"
	"strings"
	"time"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
BlacklistSettings represents blacklist configuration for a chat.

Fields:
  - ChatId: Unique identifier for the chat.
  - Action: Action to take when a blacklisted word is triggered (e.g., "ban", "mute", "none").
  - Triggers: List of blacklisted words or phrases.
  - Reason: Default reason for blacklist actions.
*/
type BlacklistSettings struct {
	ChatId   int64    `bson:"_id,omitempty" json:"_id,omitempty"`
	Action   string   `bson:"action,omitempty" json:"action,omitempty"`
	Triggers []string `bson:"triggers,omitempty" json:"triggers,omitempty"`
	Reason   string   `bson:"reason,omitempty" json:"reason,omitempty"`
}

// checkBlacklistSetting fetches blacklist settings for a chat from the database.
// If no document exists, it creates one with default values. Uses cache for performance.
// Returns a pointer to the BlacklistSettings struct with either existing or default values.
func checkBlacklistSetting(chatID int64) (blSrc *BlacklistSettings) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, chatID, new(BlacklistSettings)); err == nil && cached != nil {
		return cached.(*BlacklistSettings)
	}
	defaultBlacklistSrc := &BlacklistSettings{
		ChatId:   chatID,
		Action:   "none",
		Triggers: make([]string, 0),
		Reason:   "Automated Blacklisted word %s",
	}
	errS := findOne(blacklistsColl, bson.M{"_id": chatID}).Decode(&blSrc)
	if errS == mongo.ErrNoDocuments {
		blSrc = defaultBlacklistSrc
		err := updateOne(blacklistsColl, bson.M{"_id": chatID}, defaultBlacklistSrc)
		if err != nil {
			log.Errorf("[Database][GetBlacklistSettings]: %v ", err)
		}
	} else if errS != nil {
		log.Errorf("[Database][GetBlacklistSettings]: %v - %d", errS, chatID)
		blSrc = defaultBlacklistSrc
	}
	// Cache the result
	if blSrc != nil {
		_ = cache.Marshal.Set(cache.Context, chatID, blSrc, store.WithExpiration(10*time.Minute))
	}
	return blSrc
}

// AddBlacklist adds a new trigger word to the blacklist for the specified chat.
// The trigger is converted to lowercase before being stored to ensure case-insensitive matching.
// Updates both database and cache with the new trigger.
func AddBlacklist(chatId int64, trigger string) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Triggers = append(blSrc.Triggers, strings.ToLower(trigger))
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] AddBlacklist: %v - %d", err, chatId)
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, chatId, blSrc, store.WithExpiration(10*time.Minute))
}

// RemoveBlacklist removes a trigger word from the blacklist for the specified chat.
// The trigger is matched in lowercase. If the trigger doesn't exist, no action is taken.
// Updates both database and cache after removal.
func RemoveBlacklist(chatId int64, trigger string) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Triggers = removeStrfromStr(blSrc.Triggers, strings.ToLower(trigger))
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

// RemoveAllBlacklist clears all blacklist triggers for the specified chat.
// This removes all blocked words/phrases, effectively disabling blacklist filtering.
// The blacklist settings and action remain but the triggers list becomes empty.
func RemoveAllBlacklist(chatId int64) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Triggers = make([]string, 0)
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

// SetBlacklistAction sets the action to be taken when a blacklist trigger is matched.
// Valid actions include: ban, mute, kick, warn, none. Action is converted to lowercase.
// Updates both database and cache with the new action setting.
func SetBlacklistAction(chatId int64, action string) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Action = strings.ToLower(action)
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] ChangeBlacklistAction: %v - %d", err, chatId)
	}
}

// GetBlacklistSettings retrieves the blacklist settings for a given chat ID.
// If no settings exist, it initializes them with default values (action: none, empty triggers).
// This is the main function for accessing blacklist settings with caching support.
func GetBlacklistSettings(chatId int64) *BlacklistSettings {
	return checkBlacklistSetting(chatId)
}

// LoadBlacklistsStats returns the total number of blacklist triggers and the number of chats with at least one blacklist trigger.
// Uses MongoDB aggregation pipeline for optimal performance instead of manual loops.
// Falls back to manual counting if aggregation fails for any reason.
func LoadBlacklistsStats() (blacklistTriggers, blacklistChats int64) {
	// Use MongoDB aggregation for optimal performance
	pipeline := []bson.M{
		{
			"$project": bson.M{
				"_id":          1,
				"triggerCount": bson.M{"$size": "$triggers"},
				"hasTriggers":  bson.M{"$gt": []interface{}{bson.M{"$size": "$triggers"}, 0}},
			},
		},
		{
			"$group": bson.M{
				"_id":           nil,
				"totalTriggers": bson.M{"$sum": "$triggerCount"},
				"chatsWithTriggers": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{"$hasTriggers", 1, 0},
					},
				},
			},
		},
	}

	ctx := context.Background()
	cursor, err := blacklistsColl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Failed to aggregate blacklist stats:", err)
		// Fallback to manual method if aggregation fails
		return loadBlacklistsStatsManual()
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error("Failed to close blacklists cursor:", err)
		}
	}()

	var result struct {
		TotalTriggers     int64 `bson:"totalTriggers"`
		ChatsWithTriggers int64 `bson:"chatsWithTriggers"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			log.Error("Failed to decode blacklist stats:", err)
			// Fallback to manual method if decode fails
			return loadBlacklistsStatsManual()
		}
		return result.TotalTriggers, result.ChatsWithTriggers
	}

	// No results found, return zeros
	return 0, 0
}

// loadBlacklistsStatsManual is the fallback manual implementation for loading blacklist statistics.
// Used when MongoDB aggregation fails for any reason. Iterates through all blacklist documents
// and counts triggers and chats manually. Less efficient but more reliable.
func loadBlacklistsStatsManual() (blacklistTriggers, blacklistChats int64) {
	var blacklistStruct []*BlacklistSettings
	ctx := context.Background()

	cursor := findAll(blacklistsColl, bson.M{})
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Error(err)
		}
	}(cursor, ctx)

	for cursor.Next(ctx) {
		var blacklistSetting BlacklistSettings
		if err := cursor.Decode(&blacklistSetting); err != nil {
			log.Error("Failed to decode blacklist setting:", err)
			continue
		}
		blacklistStruct = append(blacklistStruct, &blacklistSetting)
	}

	for _, i := range blacklistStruct {
		lenBl := len(i.Triggers)
		blacklistTriggers += int64(lenBl)
		if lenBl > 0 {
			blacklistChats++
		}
	}

	return
}
