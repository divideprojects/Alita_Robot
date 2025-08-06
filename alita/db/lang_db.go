package db

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"
)

func GetLanguage(ctx *ext.Context) string {
	var chat gotgbot.Chat
	if ctx.CallbackQuery != nil {
		chat = ctx.CallbackQuery.Message.GetChat()
	} else {
		chat = ctx.EffectiveMessage.Chat
	}
	// FIXME: this is a hack
	// if ctx.Update.Message.Chat.Type == "private" || ctx.CallbackQuery.Message.Chat.Type == "private" {
	// debug_bot.PrettyPrintStruct(ctx)
	if chat.Type == "private" {
		user := ctx.EffectiveSender.User
		return getUserLanguage(user.Id)
	}
	return getGroupLanguage(chat.Id)
}

func getGroupLanguage(GroupID int64) string {
	// Try to get from cache first
	cacheKey := chatLanguageCacheKey(GroupID)
	lang, err := getFromCacheOrLoad(cacheKey, CacheTTLLanguage, func() (string, error) {
		groupc := GetChatSettings(GroupID)
		if groupc.Language == "" {
			return "en", nil
		}
		return groupc.Language, nil
	})

	if err != nil {
		return "en"
	}
	return lang
}

func getUserLanguage(UserID int64) string {
	// Try to get from cache first
	cacheKey := userLanguageCacheKey(UserID)
	lang, err := getFromCacheOrLoad(cacheKey, CacheTTLLanguage, func() (string, error) {
		userc := checkUserInfo(UserID)
		if userc == nil {
			return "en", nil
		} else if userc.Language == "" {
			return "en", nil
		}
		return userc.Language, nil
	})

	if err != nil {
		return "en"
	}
	return lang
}

func ChangeUserLanguage(UserID int64, lang string) {
	userc := checkUserInfo(UserID)
	if userc == nil {
		return
	} else if userc.Language == lang {
		return
	}

	err := UpdateRecord(&User{}, User{UserId: UserID}, User{Language: lang})
	if err != nil {
		log.Errorf("[Database] ChangeUserLanguage: %v - %d", err, UserID)
		return
	}
	// Invalidate cache after update
	deleteCache(userLanguageCacheKey(UserID))
	log.Infof("[Database] ChangeUserLanguage: %d", UserID)
}

func ChangeGroupLanguage(GroupID int64, lang string) {
	groupc := GetChatSettings(GroupID)
	if groupc.Language == lang {
		return
	}

	err := UpdateRecord(&Chat{}, Chat{ChatId: GroupID}, Chat{Language: lang})
	if err != nil {
		log.Errorf("[Database] ChangeGroupLanguage: %v - %d", err, GroupID)
		return
	}
	// Invalidate both caches after update
	deleteCache(chatLanguageCacheKey(GroupID))
	deleteCache(chatSettingsCacheKey(GroupID)) // Also invalidate chat settings cache since language is part of it
	log.Infof("[Database] ChangeGroupLanguage: %d", GroupID)
}
