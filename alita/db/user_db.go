package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// User represents a user's information stored in the database.
//
// Fields:
//   - UserId: Unique identifier for the user.
//   - UserName: Telegram username (without @).
//   - Name: Full name or first name of the user.
//   - Language: Preferred language code (defaults to "en").
type User struct {
	UserId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	UserName string `bson:"username" json:"username" default:"nil"`
	Name     string `bson:"name" json:"name" default:"nil"`
	Language string `bson:"language" json:"language" default:"en"`
}

// EnsureBotInDb ensures the bot's information is stored in the user database.
// Updates the bot's username and name in the database for consistency.
// Should be called during bot initialization.
func EnsureBotInDb(b *gotgbot.Bot) {
	usersUpdate := &User{UserId: b.Id, UserName: b.Username, Name: b.FirstName}
	err := updateOne(userColl, bson.M{"_id": b.Id}, usersUpdate)
	if err != nil {
		log.Errorf("[Database] EnsureBotInDb: %v", err)
	}
	log.Infof("[Database] Bot Updated in Database!")
}

// checkUserInfo fetches user information from the database with caching.
// Returns nil if the user doesn't exist in the database.
// Uses cache for performance optimization with 10-minute expiration.
func checkUserInfo(userId int64) (userc *User) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, userId, new(User)); err == nil && cached != nil {
		return cached.(*User)
	}
	defaultUser := &User{UserId: userId}
	errS := findOne(userColl, bson.M{"_id": userId}).Decode(&userc)
	if errS == mongo.ErrNoDocuments {
		userc = nil
	} else if errS != nil {
		log.Errorf("[Database] checkUserInfo: %v - %d", errS, userId)
		userc = defaultUser
	}
	// Cache the result (even if nil, to avoid repeated DB hits)
	if userc != nil {
		_ = cache.Marshal.Set(cache.Context, userId, userc, store.WithExpiration(10*time.Minute))
	}
	return userc
}

// UpdateUser creates or updates a user's information in the database.
// Uses atomic upsert operations to prevent race conditions.
// Updates both database and cache with the new user information.
func UpdateUser(userId int64, username, name string) {
	// Use atomic upsert to avoid race conditions
	filter := bson.M{"_id": userId}
	update := bson.M{
		"$set": bson.M{
			"username": username,
			"name":     name,
		},
		"$setOnInsert": bson.M{
			"_id":      userId,
			"language": "en",
		},
	}

	result := &User{}
	err := findOneAndUpsert(userColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database] UpdateUser: %v - %d", err, userId)
		return
	}

	// Update cache with the actual result from database
	_ = cache.Marshal.Set(cache.Context, userId, result, store.WithExpiration(10*time.Minute))
	log.Infof("[Database] UpdateUser: %d", userId)
}

// GetUserIdByUserName retrieves a user ID by their username.
// Returns 0 if the username is not found in the database.
// Username lookup is case-sensitive.
func GetUserIdByUserName(username string) int64 {
	var guids *User
	err := findOne(userColl, bson.M{"username": username}).Decode(&guids)
	if err == mongo.ErrNoDocuments {
		return 0
	} else if err != nil {
		log.Errorf("[Database] GetUserIdByUserName: %v - %d", err, guids.UserId)
		return 0
	}
	log.Infof("[Database] GetUserIdByUserName: %d", guids.UserId)
	return guids.UserId
}

// GetUserInfoById retrieves user information by user ID.
// Returns username, name, and whether the user was found in the database.
// Uses caching for improved performance.
func GetUserInfoById(userId int64) (username, name string, found bool) {
	user := checkUserInfo(userId)
	if user != nil {
		username = user.UserName
		name = user.Name
		found = true
		log.Debugf("%+v", user)
	}
	return
}

// LoadUsersStats returns the total number of users in the database.
// Used for bot statistics and monitoring purposes.
func LoadUsersStats() (count int64) {
	count, err := countDocs(userColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] loadStats: %v", err)
		return
	}
	return
}
