package db

import (
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
	log "github.com/sirupsen/logrus"
)

type Chat struct {
	ChatId     int64   `bson:"_id,omitempty" json:"_id,omitempty"`
	ChatName   string  `bson:"chat_name" json:"chat_name" default:"nil"`
	Language   string  `bson:"language" json:"language" default:"nil"`
	Users      []int64 `bson:"users" json:"users" default:"nil"`
	IsInactive bool    `bson:"is_inactive" json:"is_inactive" default:"false"`
}

var chatSettingsHandler = &SettingsHandler[Chat]{
	Collection: chatColl,
	Default: func(chatId int64) *Chat {
		return &Chat{
			ChatId:     chatId,
			ChatName:   "",
			Language:   "",
			Users:      []int64{},
			IsInactive: false,
		}
	},
}

func CheckChatSetting(chatId int64) *Chat {
	return chatSettingsHandler.CheckOrInit(chatId)
}

func GetChatSettings(chatId int64) (chatSrc *Chat) {
	return CheckChatSetting(chatId)
}

func ToggleInactiveChat(chatId int64, toggle bool) {
	chat := CheckChatSetting(chatId)
	chat.IsInactive = toggle
	err := updateOne(chatColl, map[string]interface{}{"_id": chatId}, chat)
	if err != nil {
		log.Errorf("[Database] ToggleInactiveChat: %d - %v", chatId, err)
		return
	}
}

func UpdateChat(chatId int64, chatname string, userid int64) {
	chatr := CheckChatSetting(chatId)
	foundUser := string_handling.FindInInt64Slice(chatr.Users, userid)
	if chatr.ChatName == chatname && foundUser {
		return
	} else {
		newUsers := chatr.Users
		newUsers = append(newUsers, userid)
		usersUpdate := &Chat{ChatId: chatId, ChatName: chatname, Users: newUsers, IsInactive: false}
		err2 := updateOne(chatColl, map[string]interface{}{"_id": chatId}, usersUpdate)
		if err2 != nil {
			log.Errorf("[Database] UpdateChat: %v - %d (%d)", err2, chatId, userid)
			return
		}
	}
}

func GetAllChats() map[int64]Chat {
	var (
		chatArray []*Chat
		chatMap   = make(map[int64]Chat)
	)
	cursor := findAll(chatColl, map[string]interface{}{})
	cursor.All(bgCtx, &chatArray)

	for _, i := range chatArray {
		chatMap[i.ChatId] = *i
	}

	return chatMap
}

func LoadChatStats() (activeChats, inactiveChats int) {
	chats := GetAllChats()
	for _, i := range chats {
		if i.IsInactive {
			inactiveChats++
		}
	}
	activeChats = len(chats) - inactiveChats
	return
}
