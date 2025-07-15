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

type User struct {
	UserId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	UserName string `bson:"username" json:"username" default:"nil"`
	Name     string `bson:"name" json:"name" default:"nil"`
	Language string `bson:"language" json:"language" default:"en"`
}

func EnsureBotInDb(b *gotgbot.Bot) {
	usersUpdate := &User{UserId: b.Id, UserName: b.Username, Name: b.FirstName}
	err := updateOne(userColl, bson.M{"_id": b.Id}, usersUpdate)
	if err != nil {
		log.Errorf("[Database] EnsureBotInDb: %v", err)
	}
	log.Infof("[Database] Bot Updated in Database!")
}

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

func LoadUsersStats() (count int64) {
	count, err := countDocs(userColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] loadStats: %v", err)
		return
	}
	return
}
