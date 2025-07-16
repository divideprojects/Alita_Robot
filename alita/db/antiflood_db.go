package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// default mode is 'mute'
const defaultFloodsettingsMode string = "mute"

/*
FloodSettings represents anti-flood configuration for a chat.

Fields:
  - ChatId: Unique identifier for the chat.
  - Limit: Maximum allowed consecutive messages before triggering flood action.
  - Mode: Action to take when flood is detected (e.g., "mute", "ban").
  - DeleteAntifloodMessage: Whether to delete messages that trigger the flood filter.
*/
type FloodSettings struct {
	ChatId                 int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	Limit                  int    `bson:"limit" json:"limit"`
	Mode                   string `bson:"mode,omitempty" json:"mode,omitempty"`
	DeleteAntifloodMessage bool   `bson:"del_msg" json:"del_msg"`
}

// GetFlood retrieves the antiflood settings for a given chat ID.
// If no settings exist, it initializes them with default values (limit: 0, mode: mute).
// This is the main function for accessing antiflood settings with caching support.
func GetFlood(chatID int64) *FloodSettings {
	return checkFloodSetting(chatID)
}

// checkFloodSetting fetches antiflood settings for a chat from the database.
// If no document exists, it creates one with default values (limit: 0, mode: mute).
// Uses cache for performance optimization with 10-minute expiration.
// Returns a pointer to the FloodSettings struct with either existing or default values.
func checkFloodSetting(chatID int64) (floodSrc *FloodSettings) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, chatID, new(FloodSettings)); err == nil && cached != nil {
		return cached.(*FloodSettings)
	}
	defaultFloodSrc := &FloodSettings{ChatId: chatID, Limit: 0, Mode: defaultFloodsettingsMode}

	err := findOne(antifloodSettingsColl, bson.M{"_id": chatID}).Decode(&floodSrc)
	if err == mongo.ErrNoDocuments {
		floodSrc = defaultFloodSrc
		err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, defaultFloodSrc)
		if err != nil {
			log.Errorf("[Database][checkFloodSetting]: %v ", err)
		}
	} else if err != nil {
		floodSrc = defaultFloodSrc
		log.Errorf("[Database][checkGreetingSettings]: %v ", err)
	}
	// Cache the result
	if floodSrc != nil {
		_ = cache.Marshal.Set(cache.Context, chatID, floodSrc, store.WithExpiration(10*time.Minute))
	}
	return floodSrc
}

// SetFlood updates the antiflood message limit for a specific chat.
// Set to 0 to disable antiflood. Uses atomic upsert operations to prevent race conditions.
// Updates both database and cache with the new limit setting.
func SetFlood(chatID int64, limit int) {
	// Use atomic upsert to avoid race conditions
	filter := bson.M{"_id": chatID}
	update := bson.M{
		"$set": bson.M{
			"limit": limit,
		},
		"$setOnInsert": bson.M{
			"_id":     chatID,
			"mode":    defaultFloodsettingsMode,
			"del_msg": false,
		},
	}

	result := &FloodSettings{}
	err := findOneAndUpsert(antifloodSettingsColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database] SetFlood: %v - %d", err, chatID)
		return
	}

	// Update cache with actual result from database
	_ = cache.Marshal.Set(cache.Context, chatID, result, store.WithExpiration(10*time.Minute))
}

// SetFloodMode updates the antiflood action mode for a specific chat.
// Valid modes include: mute, ban, kick, warn. Uses atomic upsert operations to prevent race conditions.
// Updates both database and cache with the new mode setting.
func SetFloodMode(chatID int64, mode string) {
	// Use atomic upsert to avoid race conditions
	filter := bson.M{"_id": chatID}
	update := bson.M{
		"$set": bson.M{
			"mode": mode,
		},
		"$setOnInsert": bson.M{
			"_id":     chatID,
			"limit":   0,
			"del_msg": false,
		},
	}

	result := &FloodSettings{}
	err := findOneAndUpsert(antifloodSettingsColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database] SetFloodMode: %v - %d", err, chatID)
		return
	}

	// Update cache with actual result from database
	_ = cache.Marshal.Set(cache.Context, chatID, result, store.WithExpiration(10*time.Minute))
}

// SetFloodMsgDel sets whether messages that trigger the antiflood filter should be deleted.
// When enabled, offending messages are automatically deleted along with the flood action.
// Uses atomic upsert operations to prevent race conditions and updates cache.
func SetFloodMsgDel(chatID int64, val bool) {
	// Use atomic upsert to avoid race conditions
	filter := bson.M{"_id": chatID}
	update := bson.M{
		"$set": bson.M{
			"del_msg": val,
		},
		"$setOnInsert": bson.M{
			"_id":   chatID,
			"limit": 0,
			"mode":  defaultFloodsettingsMode,
		},
	}

	result := &FloodSettings{}
	err := findOneAndUpsert(antifloodSettingsColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database] SetFloodMsgDel: %v - %d", err, chatID)
		return
	}

	// Update cache with actual result from database
	_ = cache.Marshal.Set(cache.Context, chatID, result, store.WithExpiration(10*time.Minute))
}

// LoadAntifloodStats returns the number of chats that have antiflood enabled.
// Calculates the total by subtracting chats with limit=0 (disabled) from total documents.
// Used for bot statistics and monitoring purposes.
func LoadAntifloodStats() (antiCount int64) {
	totalCount, err := countDocs(antifloodSettingsColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
	}
	noAntiCount, err := countDocs(antifloodSettingsColl, bson.M{"limit": 0})
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
	}

	antiCount = totalCount - noAntiCount //  gives chats which have enabled anti flood

	return
}
