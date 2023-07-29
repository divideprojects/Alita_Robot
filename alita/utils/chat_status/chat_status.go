package chat_status

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/error_handling"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// 1087968824 - Group Anonymous Bot (For anonymous users)
// 777000 - Telegram
// 136817688 - SendAsChannel Bot (For users that send messages as channel)
const (
	groupAnonymousBot = 1087968824
	tgUserId          = 777000
)

var (
	tgAdminList            = []int64{groupAnonymousBot}
	anonChatMapExpirartion = 20 * time.Second
)

// GetChat So that we can getchat with username also
func GetChat(bot *gotgbot.Bot, chatId string) (*gotgbot.Chat, error) {
	r, err := bot.Request("getChat", map[string]string{"chat_id": chatId}, nil, nil)
	if err != nil {
		return nil, err
	}

	var c gotgbot.Chat
	return &c, json.Unmarshal(r, &c)
}

func CheckDisabledCmd(bot *gotgbot.Bot, msg *gotgbot.Message, cmd string) bool {
	// Placing this first would not make additional queries if check is success!
	// if chat is private, return false
	if msg.Chat.Type == "private" {
		return false
	}

	// check if command is disabled
	if !db.IsCommandDisabled(msg.Chat.Id, cmd) {
		return false
	}

	// check if user is admin or creator, can bypass disabled commands
	member, _ := bot.GetChatMember(msg.Chat.Id, msg.From.Id, nil)
	if member.MergeChatMember().Status == "administrator" && member.MergeChatMember().Status == "creator" {
		return false
	}

	// check if ShouldDel is enabled, if not, just return true
	if !db.ShouldDel(msg.Chat.Id) {
		return false
	}

	// delete the message and return false
	_, err := msg.Delete(bot, nil)
	if err != nil {
		log.Error(err)
	}

	return true
}

func IsUserAdmin(b *gotgbot.Bot, chatID, userId int64) bool {
	// Placing this first would not make additional queries if check is success!
	if string_handling.FindInInt64Slice(tgAdminList, userId) {
		return true
	}

	chat, err := b.GetChat(chatID, nil)
	if err != nil {
		log.Error(err)
		return false
	}

	if chat.Type == "private" {
		return true
	}

	adminlist := make([]int64, 0)

	adminsAvail, admins := cache.GetAdminCacheList(chat.Id)
	if !adminsAvail {
		admins = cache.LoadAdminCache(b, chat)
	}

	if !admins.Cached {
		adminList, err := chat.GetAdministrators(b, nil)
		if err != nil {
			log.Error(err)
			return false
		}
		for _, admin := range adminList {
			adminlist = append(adminlist, admin.MergeChatMember().User.Id)
		}
	} else {
		for i := range admins.UserInfo {
			admin := &admins.UserInfo[i]
			adminlist = append(adminlist, admin.User.Id)
		}
	}

	return string_handling.FindInInt64Slice(adminlist, userId)
}

func IsBotAdmin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	if chat.Type == "private" {
		return true
	}

	mem, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)

	return mem.MergeChatMember().Status == "administrator"
}

func CanUserChangeInfo(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	var userMember gotgbot.MergedChatMember

	if db.GetAdminSettings(chat.Id).AnonAdmin && sender.IsAnonymousAdmin() {
		return true
	}

	// group anonymous bot
	if sender.IsAnonymousAdmin() {
		setAnonAdminCache(chat.Id, msg)
		_, err := sendAnonAdminKeyboard(b, msg, chat)
		if err != nil {
			log.Error(err)
		}
		return false
	}

	found, userMember := cache.GetAdminCacheUser(chat.Id, userId)
	if !found {
		tmpUserMember, err := chat.GetMember(b, userId, nil)
		userMember = tmpUserMember.MergeChatMember()
		error_handling.HandleErr(err)
	}

	if !userMember.CanChangeInfo && userMember.Status != "creator" {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "You don't have permissions to change info!!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "You don't have permission to change info in this group!", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanUserRestrict(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	var userMember gotgbot.MergedChatMember

	if db.GetAdminSettings(chat.Id).AnonAdmin && sender.IsAnonymousAdmin() {
		return true
	}

	// group anonymous bot
	if sender.IsAnonymousAdmin() {
		setAnonAdminCache(chat.Id, msg)
		_, err := sendAnonAdminKeyboard(b, msg, chat)
		if err != nil {
			log.Error(err)
		}
		return false
	}

	found, userMember := cache.GetAdminCacheUser(chat.Id, userId)
	if !found {
		tmpUserMember, err := chat.GetMember(b, userId, nil)
		userMember = tmpUserMember.MergeChatMember()
		error_handling.HandleErr(err)
	}
	if !userMember.CanRestrictMembers && userMember.Status != "creator" {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "You don't have permissions to restrict members!!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "You don't have permission to restrict users in this group!", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanBotRestrict(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	botMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botMember.MergeChatMember().CanRestrictMembers {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "I don't have permissions to restrict members!!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I can't restrict people here! Make sure I'm admin and can restrict other members.", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanUserPromote(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	var userMember gotgbot.MergedChatMember

	if db.GetAdminSettings(chat.Id).AnonAdmin && sender.IsAnonymousAdmin() {
		return true
	}

	// group anonymous bot
	if sender.IsAnonymousAdmin() {
		setAnonAdminCache(chat.Id, msg)
		_, err := sendAnonAdminKeyboard(b, msg, chat)
		if err != nil {
			log.Error(err)
		}
		return false
	}

	found, userMember := cache.GetAdminCacheUser(chat.Id, userId)
	if !found {
		tmpUserMember, err := chat.GetMember(b, userId, nil)
		userMember = tmpUserMember.MergeChatMember()
		error_handling.HandleErr(err)
	}
	if !userMember.CanPromoteMembers && userMember.Status != "creator" {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "You don't have permissions to promote/demote members!!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "You can't promote/demote people here! Make sure you have appropriate rights!", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanBotPromote(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	botChatMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botChatMember.MergeChatMember().CanPromoteMembers {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I can't promote/demote people here! Make sure I'm admin and can appoint new admins.", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanUserPin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	var userMember gotgbot.MergedChatMember

	if db.GetAdminSettings(chat.Id).AnonAdmin && sender.IsAnonymousAdmin() {
		return true
	}

	// group anonymous bot
	if sender.IsAnonymousAdmin() {
		setAnonAdminCache(chat.Id, msg)
		_, err := sendAnonAdminKeyboard(b, msg, chat)
		if err != nil {
			log.Error(err)
		}
		return false
	}

	found, userMember := cache.GetAdminCacheUser(chat.Id, userId)
	if !found {
		tmpUserMember, err := chat.GetMember(b, userId, nil)
		userMember = tmpUserMember.MergeChatMember()
		error_handling.HandleErr(err)
	}
	if !userMember.CanPinMessages && userMember.Status != "creator" {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "You can't pin messages here! Make sure you're admin and can pin messages.", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanBotPin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	botChatMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botChatMember.MergeChatMember().CanPinMessages {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I can't pin messages here! Make sure I'm admin and can pin messages.", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func Caninvite(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, msg *gotgbot.Message, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}
	if chat.Username != "" {
		return true
	}
	botChatMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botChatMember.MergeChatMember().CanInviteUsers {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I don't have access to invite links! Make sure I'm admin and can invite users.", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	sender := ctx.EffectiveSender
	if db.GetAdminSettings(chat.Id).AnonAdmin && sender.IsAnonymousAdmin() {
		return true
	}
	var userMember gotgbot.MergedChatMember

	// group anonymous bot
	if sender.IsAnonymousAdmin() {
		setAnonAdminCache(chat.Id, msg)
		_, err := sendAnonAdminKeyboard(b, msg, chat)
		if err != nil {
			log.Error(err)
		}
		return false
	}
	userid := msg.From.Id
	found, userMember := cache.GetAdminCacheUser(chat.Id, userid)
	if !found {
		tmpUserMember, err := chat.GetMember(b, userid, nil)
		userMember = tmpUserMember.MergeChatMember()
		error_handling.HandleErr(err)
	}
	if !userMember.CanInviteUsers && userMember.Status != "creator" {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "You don't have access to invite links; You need to be admin to get this!", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func CanUserDelete(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	var userMember gotgbot.MergedChatMember

	if db.GetAdminSettings(chat.Id).AnonAdmin && sender.IsAnonymousAdmin() {
		return true
	}

	// group anonymous bot
	if sender.IsAnonymousAdmin() {
		setAnonAdminCache(chat.Id, msg)
		_, err := sendAnonAdminKeyboard(b, msg, chat)
		if err != nil {
			log.Error(err)
		}
		return false
	}

	found, userMember := cache.GetAdminCacheUser(chat.Id, userId)
	if !found {
		tmpUserMember, err := chat.GetMember(b, userId, nil)
		userMember = tmpUserMember.MergeChatMember()
		error_handling.HandleErr(err)
	}

	if !userMember.CanDeleteMessages && userMember.Status != "creator" {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "You don't have permissions to delete messages!!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := msg.Reply(b, "You don't have Permissions to Delete Messages!", nil)
			if err != nil {
				log.Error(err)
				return false
			}
		}
		return false
	}
	return true
}

func CanBotDelete(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	botChatMember, err := chat.GetMember(b, b.Id, nil)
	if err != nil {
		log.Error(err)
		return false
	}

	if !botChatMember.MergeChatMember().CanDeleteMessages {
		if !justCheck {
			_, err := msg.Reply(b, "I don't have Permissions to Delete Messages!", nil)
			if err != nil {
				log.Error(err)
				return false
			}
		}
		return false
	}
	return true
}

func RequireBotAdmin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	if !IsBotAdmin(b, ctx, chat) {
		if !justCheck {
			_, err := msg.Reply(b, "I'm not admin!", nil)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func IsUserInChat(b *gotgbot.Bot, chat *gotgbot.Chat, userId int64) bool {
	// telegram cannot be in chat, will need to fix this later
	if userId == tgUserId {
		return false
	}
	member, err := chat.GetMember(b, userId, nil)
	error_handling.HandleErr(err)
	userStatus := member.MergeChatMember().Status
	return !string_handling.FindInStringSlice([]string{"left", "kicked"}, userStatus)
}

func IsUserBanProtected(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	if chat.Type == "private" {
		return true
	}

	return IsUserAdmin(b, ctx.EffectiveChat.Id, userId) || string_handling.FindInInt64Slice(tgAdminList, userId)
}

func RequireUserAdmin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	if !IsUserAdmin(b, chat.Id, userId) {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "You need to be an admin to do this!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := msg.Reply(b, "Only admins can execute this command!", nil)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func RequireUserOwner(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}

	msg := ctx.EffectiveMessage
	mem, err := chat.GetMember(b, userId, nil)
	error_handling.HandleErr(err)
	if mem == nil {
		return false
	}

	if mem.GetStatus() != "creator" {
		query := ctx.Update.CallbackQuery
		if query != nil {
			if !justCheck {
				_, err := query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "You need to be the group creator to do this!"})
				if err != nil {
					log.Error(err)
					return false
				}
			}
			return false
		}
		if !justCheck {
			_, err := msg.Reply(b, "Only group creator to do this!", nil)
			error_handling.HandleErr(err)
		}
		return false
	}

	return true
}

func RequirePrivate(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}
	msg := ctx.EffectiveMessage
	if chat.Type != "private" {
		if !justCheck {
			_, err := msg.Reply(b, "This command is made for pm, not group chat!",
				&gotgbot.SendMessageOpts{
					ReplyToMessageId: msg.MessageId,
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func RequireGroup(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		if ctx.CallbackQuery != nil {
			chat = &ctx.CallbackQuery.Message.Chat
		} else {
			chat = &ctx.Update.Message.Chat
		}
	}
	msg := ctx.EffectiveMessage
	if chat.Type == "private" {
		if !justCheck {
			_, err := msg.Reply(b, "This command is made to be used in group chats, not in pm!",
				&gotgbot.SendMessageOpts{
					ReplyToMessageId: msg.MessageId,
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

func setAnonAdminCache(chatId int64, msg *gotgbot.Message) {
	cache.Marshal.Set(cache.Context, fmt.Sprintf("anonAdmin.%d.%d", chatId, msg.MessageId), msg, store.WithExpiration(anonChatMapExpirartion))
}
