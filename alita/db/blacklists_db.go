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

	// Invalidate cache after adding blacklist
	deleteCache(blacklistCacheKey(chatId))
}

func RemoveBlacklist(chatId int64, trigger string) {
	result := DB.Where("chat_id = ? AND word = ?", chatId, strings.ToLower(trigger)).Delete(&BlacklistSettings{})
	if result.Error != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", result.Error, chatId)
	}

	// Invalidate cache if something was deleted
	if result.RowsAffected > 0 {
		deleteCache(blacklistCacheKey(chatId))
	}
}

func RemoveAllBlacklist(chatId int64) {
	err := DB.Where("chat_id = ?", chatId).Delete(&BlacklistSettings{}).Error
	if err != nil {
		log.Errorf("[Database] RemoveAllBlacklist: %v - %d", err, chatId)
	}

	// Invalidate cache after removing all blacklist entries
	deleteCache(blacklistCacheKey(chatId))
}

func SetBlacklistAction(chatId int64, action string) {
	err := DB.Model(&BlacklistSettings{}).Where("chat_id = ?", chatId).Update("action", strings.ToLower(action)).Error
	if err != nil {
		log.Errorf("[Database] SetBlacklistAction: %v - %d", err, chatId)
	}

	// Invalidate cache after updating action
	deleteCache(blacklistCacheKey(chatId))
}

func GetBlacklistSettings(chatId int64) BlacklistSettingsSlice {
	// Try to get from cache first
	cacheKey := blacklistCacheKey(chatId)
	result, err := getFromCacheOrLoad(cacheKey, CacheTTLBlacklist, func() (BlacklistSettingsSlice, error) {
		var blacklists []*BlacklistSettings
		err := GetRecords(&blacklists, BlacklistSettings{ChatId: chatId})
		if err != nil {
			log.Errorf("[Database] GetBlacklistSettings: %v - %d", err, chatId)
			return BlacklistSettingsSlice{}, err
		}
		return BlacklistSettingsSlice(blacklists), nil
	})

	if err != nil {
		return BlacklistSettingsSlice{}
	}
	return result
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
