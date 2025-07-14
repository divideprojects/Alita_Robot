package db

import (
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
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
// If no settings exist, it returns a new Chat struct with default values.
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
// If both the chat name and user are unchanged, no update is performed.
func UpdateChat(chatId int64, chatname string, userid int64) {
	chatr := GetChatSettings(chatId)
	foundUser := string_handling.FindInInt64Slice(chatr.Users, userid)
	if chatr.ChatName == chatname && foundUser {
		return
	} else {
		newUsers := chatr.Users
		newUsers = append(newUsers, userid)
		usersUpdate := &Chat{ChatId: chatId, ChatName: chatname, Users: newUsers, IsInactive: false}
		err2 := updateOne(chatColl, bson.M{"_id": chatId}, usersUpdate)
		if err2 != nil {
			log.Errorf("[Database] UpdateChat: %v - %d (%d)", err2, chatId, userid)
			return
		}
		// Update cache
		_ = cache.Marshal.Set(cache.Context, chatId, usersUpdate, store.WithExpiration(10*time.Minute))
	}
}

// GetAllChats returns a map of all chats, keyed by ChatId.
func GetAllChats() map[int64]Chat {
	var (
		chatArray []*Chat
		chatMap   = make(map[int64]Chat)
	)
	cursor := findAll(chatColl, bson.M{})
	cursor.All(bgCtx, &chatArray)

	for _, i := range chatArray {
		chatMap[i.ChatId] = *i
	}

	return chatMap
}

// LoadChatStats returns the number of active and inactive chats.
// Uses MongoDB aggregation pipeline for optimal performance.
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

	cursor, err := chatColl.Aggregate(bgCtx, pipeline)
	if err != nil {
		log.Error("Failed to aggregate chat stats:", err)
		// Fallback to manual method if aggregation fails
		return loadChatStatsManual()
	}
	defer cursor.Close(bgCtx)

	var result struct {
		ActiveChats   int `bson:"activeChats"`
		InactiveChats int `bson:"inactiveChats"`
	}

	if cursor.Next(bgCtx) {
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

/*
loadChatStatsManual is the fallback manual implementation.

Used when MongoDB aggregation fails for any reason.
*/
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
