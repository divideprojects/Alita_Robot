package db

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

func AddBlacklist(chatId int64, trigger string) {
	// Create a new blacklist entry
	blacklist := &BlacklistSettings{
		ChatId: chatId,
		Word:   strings.ToLower(trigger),
		Action: "warn", // default action
	}

	err := CreateRecord(blacklist)
	if err != nil {
		log.Errorf("[Database] AddBlacklist: %v - %d", err, chatId)
	}
}

func RemoveBlacklist(chatId int64, trigger string) {
	err := DB.Where("chat_id = ? AND word = ?", chatId, strings.ToLower(trigger)).Delete(&BlacklistSettings{}).Error
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

func RemoveAllBlacklist(chatId int64) {
	err := DB.Where("chat_id = ?", chatId).Delete(&BlacklistSettings{}).Error
	if err != nil {
		log.Errorf("[Database] RemoveAllBlacklist: %v - %d", err, chatId)
	}
}

func SetBlacklistAction(chatId int64, action string) {
	err := DB.Model(&BlacklistSettings{}).Where("chat_id = ?", chatId).Update("action", strings.ToLower(action)).Error
	if err != nil {
		log.Errorf("[Database] SetBlacklistAction: %v - %d", err, chatId)
	}
}

func GetBlacklistSettings(chatId int64) BlacklistSettingsSlice {
	var blacklists []*BlacklistSettings
	err := GetRecords(&blacklists, BlacklistSettings{ChatId: chatId})
	if err != nil {
		log.Errorf("[Database] GetBlacklistSettings: %v - %d", err, chatId)
		return BlacklistSettingsSlice{}
	}
	return BlacklistSettingsSlice(blacklists)
}

func GetBlacklistWords(chatId int64) []string {
	var blacklists []*BlacklistSettings
	err := GetRecords(&blacklists, BlacklistSettings{ChatId: chatId})
	if err != nil {
		log.Errorf("[Database] GetBlacklistWords: %v - %d", err, chatId)
		return []string{}
	}

	words := make([]string, len(blacklists))
	for i, bl := range blacklists {
		words[i] = bl.Word
	}
	return words
}

func LoadBlacklistsStats() (blacklistTriggers, blacklistChats int64) {
	// Count total blacklist entries
	err := DB.Model(&BlacklistSettings{}).Count(&blacklistTriggers).Error
	if err != nil {
		log.Errorf("[Database] LoadBlacklistsStats (triggers): %v", err)
		return 0, 0
	}

	// Count distinct chats with blacklists
	err = DB.Model(&BlacklistSettings{}).Distinct("chat_id").Count(&blacklistChats).Error
	if err != nil {
		log.Errorf("[Database] LoadBlacklistsStats (chats): %v", err)
		return blacklistTriggers, 0
	}

	return
}
