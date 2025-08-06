package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

func GetChatSettings(chatId int64) (chatSrc *Chat) {
	// Try to get from cache first
	cacheKey := chatSettingsCacheKey(chatId)
	result, err := getFromCacheOrLoad(cacheKey, CacheTTLChatSettings, func() (*Chat, error) {
		chat := &Chat{}
		dbErr := DB.Where("chat_id = ?", chatId).First(chat).Error
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return &Chat{}, nil
		} else if dbErr != nil {
			log.Errorf("[Database] getChatSettings: %v - %d ", dbErr, chatId)
			return &Chat{}, dbErr
		}
		return chat, nil
	})

	if err != nil {
		return &Chat{}
	}
	return result
}

func ToggleInactiveChat(chatId int64, toggle bool) {
	chat := GetChatSettings(chatId)
	chat.IsInactive = toggle
	err := DB.Where("chat_id = ?", chatId).Assign(chat).FirstOrCreate(&Chat{}).Error
	if err != nil {
		log.Errorf("[Database] ToggleInactiveChat: %d - %v", chatId, err)
		return
	}
	// Invalidate cache after update
	deleteCache(chatSettingsCacheKey(chatId))
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
		err := DB.Where("chat_id = ?", chatId).Assign(usersUpdate).FirstOrCreate(&Chat{}).Error
		if err != nil {
			log.Errorf("[Database] UpdateChat: %v - %d (%d)", err, chatId, userid)
			return
		}
		// Invalidate cache after update
		deleteCache(chatSettingsCacheKey(chatId))
	}
}

func GetAllChats() map[int64]Chat {
	var (
		chatArray []Chat
		chatMap   = make(map[int64]Chat)
	)
	err := DB.Find(&chatArray).Error
	if err != nil {
		log.Errorf("[Database] GetAllChats: %v", err)
		return chatMap
	}

	for _, i := range chatArray {
		chatMap[i.ChatId] = i
	}

	return chatMap
}

func LoadChatStats() (activeChats, inactiveChats int) {
	var activeCount, inactiveCount int64

	// Count active chats
	err := DB.Model(&Chat{}).Where("is_inactive = ?", false).Count(&activeCount).Error
	if err != nil {
		log.Errorf("[Database][LoadChatStats] counting active chats: %v", err)
	}

	// Count inactive chats
	err = DB.Model(&Chat{}).Where("is_inactive = ?", true).Count(&inactiveCount).Error
	if err != nil {
		log.Errorf("[Database][LoadChatStats] counting inactive chats: %v", err)
	}

	activeChats = int(activeCount)
	inactiveChats = int(inactiveCount)
	return
}
