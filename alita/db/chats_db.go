package db

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Chat represents a chat's metadata and settings stored in the database.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - ChatName: Human-readable name of the chat.
//   - Language: Language code for the chat.
//   - Users: List of user IDs associated with the chat.
//   - IsInactive: Indicates if the chat is marked as inactive.
type Chat struct {
	ChatId     int64   `bson:"_id,omitempty" json:"_id,omitempty"`
	ChatName   string  `bson:"chat_name" json:"chat_name" default:"nil"`
	Language   string  `bson:"language" json:"language" default:"nil"`
	Users      []int64 `bson:"users" json:"users" default:"nil"`
	IsInactive bool    `bson:"is_inactive" json:"is_inactive" default:"false"`
}

// GetChatSettings retrieves the chat settings for a given chat ID.
// If no settings exist, it returns an empty Chat struct. Uses cache for performance.
// This is the main function for accessing chat metadata with caching support.
func GetChatSettings(chatId int64) (chatSrc *Chat) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, chatId, new(Chat)); err == nil && cached != nil {
		return cached.(*Chat)
	}
	err := findOne(chatColl, bson.M{"_id": chatId}).Decode(&chatSrc)
	if err == mongo.ErrNoDocuments {
		chatSrc = &Chat{}
	} else if err != nil {
		log.Errorf("[Database] getChatSettings: %v - %d ", err, chatId)
		return
	}
	// Cache the result
	if chatSrc != nil {
		_ = cache.Marshal.Set(cache.Context, chatId, chatSrc, store.WithExpiration(10*time.Minute))
	}
	return
}

// ToggleInactiveChat sets the IsInactive flag for a chat to the specified value.
// Inactive chats are typically those where the bot is no longer active or banned.
// Updates the database but does not update the cache automatically.
func ToggleInactiveChat(chatId int64, toggle bool) {
	chat := GetChatSettings(chatId)
	chat.IsInactive = toggle
	err := updateOne(chatColl, bson.M{"_id": chatId}, chat)
	if err != nil {
		log.Errorf("[Database] ToggleInactiveChat: %d - %v", chatId, err)
		return
	}
}

// UpdateChat updates the chat name and adds a user ID to the chat's user list if not already present.
// Also marks the chat as active and sets default language to English for new chats.
// Uses atomic upsert operations with $addToSet to prevent duplicate users and race conditions.
func UpdateChat(chatId int64, chatname string, userid int64) {
	// Use atomic upsert with $addToSet to prevent duplicate users
	filter := bson.M{"_id": chatId}
	update := bson.M{
		"$set": bson.M{
			"chat_name":   chatname,
			"is_inactive": false,
		},
		"$addToSet": bson.M{
			"users": userid,
		},
		"$setOnInsert": bson.M{
			"_id":      chatId,
			"language": "en",
		},
	}

	result := &Chat{}
	err := findOneAndUpsert(chatColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database] UpdateChat: %v - %d (%d)", err, chatId, userid)
		return
	}

	// Update cache with the actual result from database
	_ = cache.Marshal.Set(cache.Context, chatId, result, store.WithExpiration(10*time.Minute))
}

// GetAllChats returns a map of all chats in the database, keyed by ChatId.
// This function loads all chat data into memory and should be used carefully.
// Primarily used for statistics and bulk operations.
func GetAllChats() map[int64]Chat {
	var (
		chatArray []*Chat
		chatMap   = make(map[int64]Chat)
	)
	cursor := findAll(chatColl, bson.M{})
	ctx := context.Background()
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error("Failed to close chats cursor:", err)
		}
	}()
	if err := cursor.All(ctx, &chatArray); err != nil {
		log.Error("Failed to load all chats:", err)
		return chatMap
	}

	for _, i := range chatArray {
		chatMap[i.ChatId] = *i
	}

	return chatMap
}

// LoadChatStats returns the number of active and inactive chats.
// Uses MongoDB aggregation pipeline for optimal performance, with manual fallback.
// Used for bot statistics and monitoring purposes.
func LoadChatStats() (activeChats, inactiveChats int) {
	// Use MongoDB aggregation for optimal performance
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": nil,
				"activeChats": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$is_inactive", false}},
							1,
							0,
						},
					},
				},
				"inactiveChats": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$is_inactive", true}},
							1,
							0,
						},
					},
				},
			},
		},
	}

	ctx := context.Background()
	cursor, err := chatColl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Failed to aggregate chat stats:", err)
		// Fallback to manual method if aggregation fails
		return loadChatStatsManual()
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error("Failed to close chat stats cursor:", err)
		}
	}()

	var result struct {
		ActiveChats   int `bson:"activeChats"`
		InactiveChats int `bson:"inactiveChats"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			log.Error("Failed to decode chat stats:", err)
			// Fallback to manual method if decode fails
			return loadChatStatsManual()
		}
		return result.ActiveChats, result.InactiveChats
	}

	// No results found, return zeros
	return 0, 0
}

// loadChatStatsManual is the fallback manual implementation for loading chat statistics.
// Used when MongoDB aggregation fails for any reason. Loads all chats into memory
// and counts them manually. Less efficient but more reliable.
func loadChatStatsManual() (activeChats, inactiveChats int) {
	chats := GetAllChats()
	for _, i := range chats {
		if i.IsInactive {
			inactiveChats++
		}
	}
	activeChats = len(chats) - inactiveChats
	return
}
