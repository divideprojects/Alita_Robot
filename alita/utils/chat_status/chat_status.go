// Package chat_status provides utilities for checking user and bot permissions in Telegram chats.
//
// This package contains functions for verifying admin status, bot capabilities,
// user permissions, and enforcing access control for various bot commands and features.
// All functions include proper error handling and user feedback mechanisms.
package chat_status

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
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

// safeGetChat safely retrieves chat information from context, handling different update types.
// It checks various update types (callback query, message, chat member) and returns the appropriate chat.
// Falls back to EffectiveChat if other methods fail.
func safeGetChat(ctx *ext.Context) *gotgbot.Chat {
	if ctx.CallbackQuery != nil {
		_chatValue := ctx.CallbackQuery.Message.GetChat()
		return &_chatValue
	} else if ctx.Message != nil {
		return &ctx.Message.Chat
	} else if ctx.ChatMember != nil {
		return &ctx.ChatMember.Chat
	} else {
		// Fallback to EffectiveChat if available
		return ctx.EffectiveChat
	}
}

// GetChat So that we can getchat with username also
/*
GetChat retrieves chat information for the given chat ID.

This function allows fetching chat details using either a username or numeric ID.
Returns a pointer to gotgbot.Chat and an error if the request fails.
*/
func GetChat(bot *gotgbot.Bot, chatId string) (*gotgbot.Chat, error) {
	r, err := bot.Request("getChat", map[string]string{"chat_id": chatId}, nil, nil)
	if err != nil {
		return nil, err
	}

	var c gotgbot.Chat
	return &c, json.Unmarshal(r, &c)
}

/*
CheckDisabledCmd determines if a command is disabled in the current chat context.

Returns true if the command is disabled and should be blocked, otherwise false.
Handles admin bypass and message deletion if required.
*/
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

/*
IsUserAdmin checks if the specified user is an admin in the given chat.

Uses cached admin data if available, otherwise fetches from Telegram.
Returns true if the user is an admin, otherwise false.
Optimized with map-based lookups for O(1) performance.
*/
func IsUserAdmin(b *gotgbot.Bot, chatID, userId int64) bool {
	// Check global admin list first with optimized lookup
	tgAdminMap := string_handling.Int64SliceToMap(tgAdminList)
	if string_handling.FindInInt64Map(tgAdminMap, userId) {
		return true
	}

	// Create context with timeout to prevent blocking indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check cache first - avoid GetChat call if possible
	adminList, cached := cache.GetAdmins(b, chatID)
	if cached {
		// Use optimized map-based lookup for cached data
		adminIds := make([]int64, len(adminList))
		for i, admin := range adminList {
			adminIds[i] = admin.User.Id
		}
		adminMap := string_handling.Int64SliceToMap(adminIds)
		return string_handling.FindInInt64Map(adminMap, userId)
	}

	// Only make GetChat call if cache miss - use context with timeout
	chat, err := b.GetChatWithContext(ctx, chatID, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"chatID": chatID,
			"userID": userId,
			"error":  err,
		}).Warning("IsUserAdmin: Failed to get chat, treating as non-admin")
		return false
	}

	// Don't allow check if not a group/supergroup
	if chat.Type != "group" && chat.Type != "supergroup" {
		return false
	}

	// Check if user is in admin list (adminList already contains fresh data from GetAdmins)
	for i := range adminList {
		admin := &adminList[i]
		if admin.User.Id == userId {
			return true
		}
	}

	return false
}

/*
IsBotAdmin checks if the bot is an admin in the specified chat.

Returns true if the bot has administrator status, otherwise false.
*/
func IsBotAdmin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	if chat.Type == "private" {
		return true
	}

	mem, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)

	return mem.MergeChatMember().Status == "administrator"
}

/*
CanUserChangeInfo checks if a user has permission to change chat info.

Returns true if the user can change info, otherwise false.
Handles anonymous admin cases and provides feedback if permissions are lacking.
*/
func CanUserChangeInfo(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
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
		query := ctx.CallbackQuery
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
			_, err := b.SendMessage(chat.Id, tr.GetString("strings.utils.chat_status.user.no_permission_change_info_cmd"),
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanUserRestrict checks if a user has permission to restrict members in the chat.

Returns true if the user can restrict members, otherwise false.
Handles anonymous admin and provides feedback if permissions are lacking.
*/
func CanUserRestrict(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
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
		query := ctx.CallbackQuery
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
			_, err := b.SendMessage(chat.Id, "You don't have permission to restrict users in this group!",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanBotRestrict checks if the bot has permission to restrict members in the chat.

Returns true if the bot can restrict members, otherwise false.
Provides feedback if permissions are lacking.
*/
func CanBotRestrict(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	botMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botMember.MergeChatMember().CanRestrictMembers {
		query := ctx.CallbackQuery
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
			_, err := b.SendMessage(chat.Id, "I can't restrict people here! Make sure I'm admin and can restrict other members.",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanUserPromote checks if a user has permission to promote or demote members in the chat.

Returns true if the user can promote/demote, otherwise false.
Handles anonymous admin and provides feedback if permissions are lacking.
*/
func CanUserPromote(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
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
		query := ctx.CallbackQuery
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
			_, err := b.SendMessage(chat.Id, "You can't promote/demote people here! Make sure you have appropriate rights!",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanBotPromote checks if the bot has permission to promote or demote members in the chat.

Returns true if the bot can promote/demote, otherwise false.
Provides feedback if permissions are lacking.
*/
func CanBotPromote(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	botChatMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botChatMember.MergeChatMember().CanPromoteMembers {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I can't promote/demote people here! Make sure I'm admin and can appoint new admins.",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanUserPin checks if a user has permission to pin messages in the chat.

Returns true if the user can pin messages, otherwise false.
Handles anonymous admin and provides feedback if permissions are lacking.
*/
func CanUserPin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
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
			_, err := b.SendMessage(chat.Id, "You can't pin messages here! Make sure you're admin and can pin messages.", &gotgbot.SendMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                ctx.EffectiveMessage.MessageId,
					AllowSendingWithoutReply: true,
				},
			},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanBotPin checks if the bot has permission to pin messages in the chat.

Returns true if the bot can pin messages, otherwise false.
Provides feedback if permissions are lacking.
*/
func CanBotPin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	botChatMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botChatMember.MergeChatMember().CanPinMessages {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I can't pin messages here! Make sure I'm admin and can pin messages.",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
Caninvite checks if the bot or user has permission to generate invite links for the chat.

Returns true if invite links can be generated, otherwise false.
Handles anonymous admin and provides feedback if permissions are lacking.
*/
func Caninvite(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, msg *gotgbot.Message, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}
	if chat.Username != "" {
		return true
	}
	botChatMember, err := chat.GetMember(b, b.Id, nil)
	error_handling.HandleErr(err)
	if !botChatMember.MergeChatMember().CanInviteUsers {
		if !justCheck {
			_, err := b.SendMessage(chat.Id, "I don't have access to invite links! Make sure I'm admin and can invite users.",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
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
			_, err := b.SendMessage(chat.Id, "You don't have access to invite links; You need to be admin to get this!",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                ctx.EffectiveMessage.MessageId,
						AllowSendingWithoutReply: true,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
CanUserDelete checks if a user has permission to delete messages in the chat.

Returns true if the user can delete messages, otherwise false.
Handles anonymous admin and provides feedback if permissions are lacking.
*/
func CanUserDelete(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
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
		query := ctx.CallbackQuery
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

/*
CanBotDelete checks if the bot has permission to delete messages in the chat.

Returns true if the bot can delete messages, otherwise false.
Provides feedback if permissions are lacking.
*/
func CanBotDelete(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
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

/*
RequireBotAdmin ensures the bot is an admin in the chat.

Returns true if the bot is admin, otherwise false.
Provides feedback if the bot lacks admin rights.
*/
func RequireBotAdmin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
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

/*
IsUserInChat checks if a user is currently a member of the chat.

Returns true if the user is present and not left or kicked, otherwise false.
*/
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

/*
IsUserBanProtected checks if a user is protected from being banned in the chat.

Returns true if the user is an admin or a Telegram special user, otherwise false.
*/
func IsUserBanProtected(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	if chat.Type == "private" {
		return true
	}

	return IsUserAdmin(b, ctx.EffectiveChat.Id, userId) || string_handling.FindInInt64Slice(tgAdminList, userId)
}

/*
RequireUserAdmin ensures the user is an admin in the chat.

Returns true if the user is admin, otherwise false.
Provides feedback if the user lacks admin rights.
*/
func RequireUserAdmin(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	msg := ctx.EffectiveMessage
	if !IsUserAdmin(b, chat.Id, userId) {
		query := ctx.CallbackQuery
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

/*
RequireUserOwner ensures the user is the creator (owner) of the chat.

Returns true if the user is the group creator, otherwise false.
Provides feedback if the user lacks ownership rights.
*/
func RequireUserOwner(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, userId int64, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}

	msg := ctx.EffectiveMessage
	mem, err := chat.GetMember(b, userId, nil)
	error_handling.HandleErr(err)
	if mem == nil {
		return false
	}

	if mem.GetStatus() != "creator" {
		query := ctx.CallbackQuery
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

/*
RequirePrivate ensures the command is executed in a private chat.

Returns true if the chat is private, otherwise false.
Provides feedback if used in a group context.
*/
func RequirePrivate(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}
	msg := ctx.EffectiveMessage
	if chat.Type != "private" {
		if !justCheck {
			_, err := msg.Reply(b, "This command is made for pm, not group chat!",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId: msg.MessageId,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
RequireGroup ensures the command is executed in a group chat.

Returns true if the chat is a group, otherwise false.
Provides feedback if used in a private context.
*/
func RequireGroup(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, justCheck bool) bool {
	if chat == nil {
		chat = safeGetChat(ctx)
	}
	msg := ctx.EffectiveMessage
	if chat.Type == "private" {
		if !justCheck {
			_, err := msg.Reply(b, "This command is made to be used in group chats, not in pm!",
				&gotgbot.SendMessageOpts{
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId: msg.MessageId,
					},
				},
			)
			error_handling.HandleErr(err)
		}
		return false
	}
	return true
}

/*
setAnonAdminCache stores the anonymous admin message in the cache for a limited time.

Used to track anonymous admin actions for verification purposes.
*/
func setAnonAdminCache(chatId int64, msg *gotgbot.Message) {
	if err := cache.Marshal.Set(cache.Context, fmt.Sprintf("anonAdmin.%d.%d", chatId, msg.MessageId), msg, store.WithExpiration(anonChatMapExpirartion)); err != nil {
		log.Error("Failed to set anonymous admin cache:", err)
	}
}
