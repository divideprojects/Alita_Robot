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
	groupc := GetChatSettings(GroupID)
	if groupc.Language == "" {
		return "en"
	}
	return groupc.Language
}

func getUserLanguage(UserID int64) string {
	userc := checkUserInfo(UserID)
	if userc == nil {
		return "en"
	} else if userc.Language == "" {
		return "en"
	}
	return userc.Language
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
	log.Infof("[Database] ChangeGroupLanguage: %d", GroupID)
}
