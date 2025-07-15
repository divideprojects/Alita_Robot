package db

import (
	"sync"
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

var floodSettingsMutex sync.Mutex

// GetFlood retrieves the flood settings for a given chat ID.
// If no settings exist, it initializes them with default values.
func GetFlood(chatID int64) *FloodSettings {
	return checkFloodSetting(chatID)
}

// checkFloodSetting fetches flood settings for a chat from the database.
// If no document exists, it creates one with default values.
// Returns a pointer to the FloodSettings struct.
func checkFloodSetting(chatID int64) (floodSrc *FloodSettings) {
	floodSettingsMutex.Lock()
	defer floodSettingsMutex.Unlock()

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

// SetFlood updates the flood limit for a specific chat.
// If the mode is not set, it initializes the FloodSettings with the default mode.
func SetFlood(chatID int64, limit int) {
	floodSrc := checkFloodSetting(chatID)

	if floodSrc.Mode == "" {
		floodSrc = &FloodSettings{ChatId: chatID, Limit: limit, Mode: defaultFloodsettingsMode}
	} else {
		floodSrc.Limit = limit // update floodSrc.limit
	}

	// update the value in db
	err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFlood: %v - %d", err, chatID)
	}
	// Invalidate cache
	_ = cache.Marshal.Delete(cache.Context, chatID)
}

// SetFloodMode updates the flood action mode for a specific chat.
// The mode determines what action is taken when flooding is detected.
func SetFloodMode(chatID int64, mode string) {
	floodSrc := checkFloodSetting(chatID)
	floodSrc.Mode = mode

	err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFloodMode: %v - %d", err, chatID)
	}
	// Invalidate cache
	_ = cache.Marshal.Delete(cache.Context, chatID)
}

// SetFloodMsgDel sets whether messages that trigger the flood filter should be deleted for a chat.
func SetFloodMsgDel(chatID int64, val bool) {
	floodSrc := checkFloodSetting(chatID)
	floodSrc.DeleteAntifloodMessage = val

	err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFloodMsgDel: %v - %d", err, chatID)
	}
	// Invalidate cache
	_ = cache.Marshal.Delete(cache.Context, chatID)
}

/*
LoadAntifloodStats returns the number of chats that have anti-flood enabled.

It calculates the total number of chat documents and subtracts those with a flood limit of zero,
indicating anti-flood is disabled.
*/
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
