package db

import (
	log "github.com/sirupsen/logrus"
)

type Channel struct {
	ChannelId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	ChannelName string `bson:"channel_name" json:"channel_name" default:"nil"`
	Username    string `bson:"username" json:"username" default:"nil"`
}

var channelSettingsHandler = &SettingsHandler[Channel]{
	Collection: channelColl,
	Default: func(channelId int64) *Channel {
		return &Channel{
			ChannelId:   channelId,
			ChannelName: "",
			Username:    "",
		}
	},
}

func CheckChannelSetting(channelId int64) *Channel {
	return channelSettingsHandler.CheckOrInit(channelId)
}

func GetChannelSettings(channelId int64) (channelSrc *Channel) {
	channelSrc = CheckChannelSetting(channelId)
	if channelSrc.ChannelId == 0 {
		channelSrc = nil
	}
	return
}

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

	err2 := updateOne(channelColl, map[string]interface{}{"_id": channelId}, channelSrc)
	if err2 != nil {
		log.Errorf("[Database] UpdateChannel: %v - %d (%s)", err2, channelId, username)
		return
	}
	log.Infof("[Database] UpdateChannel: %s", channelName)
}

func GetChannelIdByUserName(username string) int64 {
	var cuids *Channel
	err := findOne(channelColl, map[string]interface{}{"username": username}).Decode(&cuids)
	if err != nil {
		return 0
	}
	log.Infof("[Database] GetChannelByUserName: %d", cuids.ChannelId)
	return cuids.ChannelId
}

func GetChannelInfoById(userId int64) (username, name string, found bool) {
	channel := GetChannelSettings(userId)
	if channel != nil {
		username = channel.Username
		name = channel.ChannelName
		found = true
	}
	return
}

func LoadChannelStats() (count int64) {
	count, err := countDocs(channelColl, nil)
	if err != nil {
		log.Errorf("[Database] loadChannelStats: %v", err)
		return 0
	}
	return
}
