package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

func GetChatSettings(chatId int64) (chatSrc *Chat) {
	chatSrc = &Chat{}
	err := DB.Where("chat_id = ?", chatId).First(chatSrc)
	if errors.Is(err.Error, gorm.ErrRecordNotFound) {
		chatSrc = &Chat{}
	} else if err.Error != nil {
		log.Errorf("[Database] getChatSettings: %v - %d ", err.Error, chatId)
		return
	}
	return
}

func ToggleInactiveChat(chatId int64, toggle bool) {
	chat := GetChatSettings(chatId)
	chat.IsInactive = toggle
	err := DB.Where("chat_id = ?", chatId).Assign(chat).FirstOrCreate(&Chat{})
	if err.Error != nil {
		log.Errorf("[Database] ToggleInactiveChat: %d - %v", chatId, err.Error)
		return
	}
}

func UpdateChat(chatId int64, chatname string, userid int64) {
	chatr := GetChatSettings(chatId)
	foundUser := string_handling.FindInInt64Slice(chatr.Users, userid)
	if chatr.ChatName == chatname && foundUser {
		return
	} else {
		newUsers := chatr.Users
		newUsers = append(newUsers, userid)
		usersUpdate := &Chat{ChatId: chatId, ChatName: chatname, Users: newUsers, IsInactive: false}
		err := DB.Where("chat_id = ?", chatId).Assign(usersUpdate).FirstOrCreate(&Chat{})
		if err.Error != nil {
			log.Errorf("[Database] UpdateChat: %v - %d (%d)", err.Error, chatId, userid)
			return
		}
	}
}

func GetAllChats() map[int64]Chat {
	var (
		chatArray []Chat
		chatMap   = make(map[int64]Chat)
	)
	err := DB.Find(&chatArray)
	if err.Error != nil {
		log.Errorf("[Database] GetAllChats: %v", err.Error)
		return chatMap
	}

	for _, i := range chatArray {
		chatMap[i.ChatId] = i
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
