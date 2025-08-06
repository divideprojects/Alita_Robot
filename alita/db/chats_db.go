package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

func GetChatSettings(chatId int64) (chatSrc *Chat) {
	// Use optimized cached query instead of SELECT *
	chat, err := OptimizedQueries.GetChatBasicInfoCached(chatId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &Chat{}
		}
		log.Errorf("[Database] GetChatSettings: %v - %d", err, chatId)
		return &Chat{}
	}
	return chat
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
	
	// Check if update is actually needed
	if chatr.ChatName == chatname && foundUser {
		return
	}
	
	// Prepare updates only for changed fields
	updates := make(map[string]interface{})
	if chatr.ChatName != chatname {
		updates["chat_name"] = chatname
	}
	if !foundUser {
		newUsers := chatr.Users
		newUsers = append(newUsers, userid)
		updates["users"] = newUsers
	}
	updates["is_inactive"] = false
	
	if chatr.ChatId == 0 {
		// Create new chat
		newChat := &Chat{
			ChatId:     chatId,
			ChatName:   chatname,
			Users:      Int64Array{userid},
			IsInactive: false,
		}
		err := DB.Create(newChat).Error
		if err != nil {
			log.Errorf("[Database] UpdateChat: %v - %d (%d)", err, chatId, userid)
			return
		}
	} else if len(updates) > 0 {
		// Update existing chat only if there are changes
		err := DB.Model(&Chat{}).Where("chat_id = ?", chatId).Updates(updates).Error
		if err != nil {
			log.Errorf("[Database] UpdateChat: %v - %d (%d)", err, chatId, userid)
			return
		}
	}
	
	// Invalidate cache after update
	deleteCache(chatCacheKey(chatId))
	log.Debugf("[Database] UpdateChat: %d", chatId)
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
