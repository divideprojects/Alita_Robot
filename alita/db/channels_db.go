package db

import (
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Channel represents a Telegram channel's metadata stored in the database.
//
// Fields:
//   - ChannelId: Unique identifier for the channel.
//   - ChannelName: Human-readable name of the channel.
//   - Username: Channel's username (if any).
type Channel struct {
	ChannelId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	ChannelName string `bson:"channel_name" json:"channel_name" default:"nil"`
	Username    string `bson:"username" json:"username" default:"nil"`
}

// GetChannelSettings retrieves the channel settings for a given channel ID.
// Returns nil if the channel does not exist in the database.
func GetChannelSettings(channelId int64) (channelSrc *Channel) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, channelId, new(Channel)); err == nil && cached != nil {
		return cached.(*Channel)
	}
	err := findOne(channelColl, bson.M{"_id": channelId}).Decode(&channelSrc)
	if err == mongo.ErrNoDocuments {
		channelSrc = nil
	} else if err != nil {
		log.Errorf("[Database] getChannelSettings: %v - %d ", err, channelId)
		return
	}
	// Cache the result
	if channelSrc != nil {
		_ = cache.Marshal.Set(cache.Context, channelId, channelSrc, store.WithExpiration(10*time.Minute))
	}
	return
}

// UpdateChannel updates the channel's name and username for the given channel ID.
// If the channel does not exist, it creates a new entry.
func UpdateChannel(channelId int64, channelName, username string) {
	channelSrc := GetChannelSettings(channelId)

	if channelSrc != nil {
		if channelSrc.ChannelName == channelName && channelSrc.Username == username {
			return
		}
		channelSrc.ChannelName = channelName
		channelSrc.Username = username
	} else {
		channelSrc = &Channel{
			ChannelId:   channelId,
			ChannelName: channelName,
			Username:    username,
		}
	}

	err2 := updateOne(channelColl, bson.M{"_id": channelId}, channelSrc)
	if err2 != nil {
		log.Errorf("[Database] UpdateChannel: %v - %d (%s)", err2, channelId, username)
		return
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, channelId, channelSrc, store.WithExpiration(10*time.Minute))
	log.Infof("[Database] UpdateChannel: %s", channelName)
}

// GetChannelIdByUserName returns the channel ID for a given username.
// Returns 0 if the channel is not found.
func GetChannelIdByUserName(username string) int64 {
	var cuids *Channel
	err := findOne(channelColl, bson.M{"username": username}).Decode(&cuids)
	if err == mongo.ErrNoDocuments {
		return 0
	} else if err != nil {
		log.Errorf("[Database] GetChannelByUserName: %v - %d", err, cuids.ChannelId)
		return 0
	}
	log.Infof("[Database] GetChannelByUserName: %d", cuids.ChannelId)
	return cuids.ChannelId
}

// GetChannelInfoById retrieves the username and channel name for a given channel ID.
// Returns found=false if the channel does not exist.
func GetChannelInfoById(userId int64) (username, name string, found bool) {
	channel := GetChannelSettings(userId)
	if channel != nil {
		username = channel.Username
		name = channel.ChannelName
		found = true
	}
	return
}

// LoadChannelStats returns the total number of channels stored in the database.
func LoadChannelStats() (count int64) {
	count, err := countDocs(channelColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] loadChannelStats: %v", err)
		return 0
	}
	return
}
