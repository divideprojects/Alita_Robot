package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
)

// bansModule provides ban, kick, restrict, and unrestrict logic for group chats.
//
// Implements all moderation actions related to user removal and restriction.
var bansModule = moduleStruct{moduleName: "Bans"}

// dkick deletes a user's message and kicks them from the group.
//
// Performs permission checks, extracts the target user from a reply, deletes their message, and kicks them. Handles edge cases such as anonymous users and admins.
//
// The Bot and the user issuing the command must have appropriate permissions.
func (moduleStruct) dkick(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserDelete(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}
	if msg.ReplyToMessage != nil {
		_, _ = msg.Reply(b, tr.GetString("strings.bans.kick.reply_to_delete_kick"), nil)
		return ext.EndGroups
	}

	_, reason := extraction.ExtractUserAndText(b, ctx)
	userId := msg.ReplyToMessage.From.Id
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("strings.bans.errors.anon_user_restrict"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, _ = msg.ReplyToMessage.Delete(b, nil)

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kick.user_not_in_chat"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kick.cannot_kick_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kick.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := chat.BanMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	time.Sleep(2 * time.Second)

	_, err = chat.UnbanMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	kickuser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := tr.GetString("strings.bans.kick.kicked_user")
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings.bans.kick.kicked_reason"), reason)
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(baseStr, helpers.MentionHtml(kickuser.Id, kickuser.FirstName)),
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// kick removes a user from the group.
//
// Checks permissions, extracts the target user, and kicks them. Handles edge cases such as anonymous users, admins, and the bot itself.
func (moduleStruct) kick(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId, reason := extraction.ExtractUserAndText(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("strings.bans.errors.anon_user_restrict"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kick.user_not_in_chat"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kick.cannot_kick_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kick.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := chat.BanMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	time.Sleep(2 * time.Second)

	_, err = chat.UnbanMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	kickuser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := tr.GetString("strings.bans.kick.kicked_user")
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings.bans.kick.kicked_reason"), reason)
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(baseStr, helpers.MentionHtml(kickuser.Id, kickuser.FirstName)),
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// kickme allows a user to remove themselves from the group.
//
// Admins are not allowed to use this command. The bot must have restriction permissions.
func (moduleStruct) kickme(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}

	// Don't allow admins to use the command
	if chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.kickme.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// kick the member
	_, err := chat.UnbanMember(b, user.Id, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = msg.Reply(b, tr.GetString("strings.bans.kickme.ok_out"), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// tBan temporarily bans a user from the chat.
//
// Performs permission checks, extracts the target user and ban duration, and bans them for the specified time. Handles edge cases such as anonymous users and admins.
func (moduleStruct) tBan(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId, reason := extraction.ExtractUserAndText(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("strings.bans.errors.anon_user_restrict"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// Extract Time
	_time, timeVal, reason := extraction.ExtractTime(b, ctx, reason)
	if _time == -1 {
		return ext.EndGroups
	}

	_, err := chat.BanMember(b,
		userId,
		&gotgbot.BanChatMemberOpts{
			UntilDate: _time,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	banUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := fmt.Sprintf(
		tr.GetString("strings.bans.ban.tban"),
		helpers.MentionHtml(banUser.Id, banUser.FirstName),
		timeVal,
	)
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings.bans.ban.ban_reason"), reason)
	}

	_, err = msg.Reply(
		b,
		baseStr,
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// ban bans a user from the group indefinitely.
//
// Checks permissions, extracts the target user, and bans them. Handles anonymous users, admins, and the bot itself. Provides an inline button for unbanning.
func (moduleStruct) ban(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	var text string
	var sendMsgOptns *gotgbot.SendMessageOpts

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, true) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId, reason := extraction.ExtractUserAndText(b, ctx)
	switch userId {
	case -1:
		return ext.EndGroups
	case 0:
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		if msg.ReplyToMessage != nil {
			userId := msg.ReplyToMessage.GetSender().Id()
			_, err := b.BanChatSenderChat(chat.Id, userId, nil)
			if err != nil {
				log.Error(err)
				return err
			}
			text = "Banned user: " + helpers.MentionHtml(userId, msg.ReplyToMessage.GetSender().Name())
		} else {
			text = "You can only ban an anonymous user by replying to their message."
		}
		sendMsgOptns = helpers.Shtml()
	} else {
		_, err := chat.BanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		_, name, _ := extraction.GetUserInfo(userId)

		baseStr := tr.GetString("strings.bans.ban.normal_ban")
		if reason != "" {
			baseStr += fmt.Sprintf(tr.GetString("strings.bans.ban.ban_reason"), reason)
		}

		text = fmt.Sprintf(baseStr, helpers.MentionHtml(userId, name))

		sendMsgOptns = &gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         "Unban (Admin Only)",
							CallbackData: fmt.Sprintf("unrestrict.unban.%d", userId),
						},
					},
				},
			},
		}
	}

	_, err := msg.Reply(b, text, sendMsgOptns)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// sBan silently bans a user from the group and deletes the command message.
//
// Used to silently ban a user from group.
// This deletes the command of Banner and also does not reply.
// The Bot, Banner should be admin with ban permissions in order to use this.

// Performs permission checks, extracts the target user, and bans them without sending a reply. Handles edge cases such as anonymous users and admins.
func (moduleStruct) sBan(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("strings.bans.errors.anon_user_restrict"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := chat.BanMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = msg.Delete(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// dBan bans a user from the group and deletes their message.
//
// Used to ban a user from group and delete their message.
// This deletes the message of replied user.
// The Bot, Banner should be admin with ban permissions in order to use this.

// Checks permissions, extracts the target user, deletes their message, and bans them. Handles anonymous users, admins, and edge cases.
func (moduleStruct) dBan(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserDelete(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId, reason := extraction.ExtractUserAndText(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("strings.bans.errors.anon_user_restrict"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, tr.GetString("strings.bans.ban.dban.no_reply"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := msg.ReplyToMessage.Delete(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = chat.BanMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	banUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := tr.GetString("strings.bans.ban.normal_ban")
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings.bans.ban.ban_reason"), reason)
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(baseStr, helpers.MentionHtml(banUser.Id, banUser.FirstName)),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         "Unban (Admin Only)",
							CallbackData: fmt.Sprintf("unrestrict.unban.%d", userId),
						},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// unban removes a ban from a user in the group.
//
// Used to unban a user from group.
// The Bot, Unbanner should be admin with ban permissions in order to use this.

// Checks permissions, extracts the target user, and unbans them. Handles anonymous users, admins, and the bot itself.
func (moduleStruct) unban(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	var text string

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId := extraction.ExtractUser(b, ctx)
	switch userId {
	case -1:
		return ext.EndGroups
	case 0:
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings.bans.unban.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		if msg.ReplyToMessage != nil {
			userId := msg.ReplyToMessage.GetSender().Id()
			_, err := b.UnbanChatSenderChat(chat.Id, userId, nil)
			if err != nil {
				log.Error(err)
				return err
			}
			text = "Banned user: " + helpers.MentionHtml(userId, msg.ReplyToMessage.GetSender().Name())
		} else {
			text = "You can only unban an anonymous user by replying to their message."
		}
	} else {
		_, err := chat.UnbanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		banUser, err := b.GetChat(userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		text = fmt.Sprintf(
			tr.GetString("strings.bans.unban.unbanned_user"),
			helpers.MentionHtml(banUser.Id, banUser.FirstName),
		)
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// restrict shows an inline keyboard to restrict a user (ban, kick, mute).
//
// Used to restrict members from a chat.
// Shows an inline keyboard menu which shows options to kick, ban and mute.

// Checks permissions and displays options for restricting the target user.
func (moduleStruct) restrict(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, chat, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, chat, false) {
		return ext.EndGroups
	}

	userId := extraction.ExtractUser(b, ctx)
	switch userId {
	case -1:
		return ext.EndGroups
	case 0:
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		_, err := msg.Reply(b, "This user is not in this chat, how can I restrict them?", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, "Why would I restrict an admin? That sounds like a pretty dumb idea.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, "Why would I restrict myself?", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := msg.Reply(b, "How can I restrict this user?",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Ban", CallbackData: fmt.Sprintf("restrict.ban.%d", userId)},
						{Text: "Kick", CallbackData: fmt.Sprintf("restrict.kick.%d", userId)},
					},
					{{Text: "Mute", CallbackData: fmt.Sprintf("restrict.mute.%d", userId)}},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// restrictButtonHandler handles callback queries for the restrict command.
//
// Handles the queries for restrict command.
// Performs the selected restriction action (ban, kick, mute) on the target user.
func (moduleStruct) restrictButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// permissions check
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	var helpText string

	action := args[0]
	userId, _ := strconv.Atoi(args[1])

	actionUser, err := b.GetChat(int64(userId), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	switch action {
	case "kick":
		_, err := chat.BanMember(b, int64(userId), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		helpText = fmt.Sprintf("Admin %s kicked %s from this chat!",
			helpers.MentionHtml(user.Id, user.FirstName),
			helpers.MentionHtml(int64(userId), actionUser.FirstName),
		)
		// unban the member
		time.Sleep(3 * time.Second)
		_, err = chat.UnbanMember(b, int64(userId), nil)
		if err != nil {
			log.Error(err)
			return err
		}
	case "mute":
		_, err := chat.RestrictMember(b, int64(userId),
			gotgbot.ChatPermissions{
				CanSendMessages:       false,
				CanSendPhotos:         false,
				CanSendVideos:         false,
				CanSendAudios:         false,
				CanSendDocuments:      false,
				CanSendVideoNotes:     false,
				CanSendVoiceNotes:     false,
				CanAddWebPagePreviews: false,
				CanChangeInfo:         false,
				CanInviteUsers:        false,
				CanPinMessages:        false,
				CanManageTopics:       false,
				CanSendPolls:          false,
				CanSendOtherMessages:  false,
			},
			nil,
		)
		if err != nil {
			log.Error(err)
			return err
		}
		helpText = fmt.Sprintf("Admin %s muted %s in chat!",
			helpers.MentionHtml(user.Id, user.FirstName),
			helpers.MentionHtml(int64(userId), actionUser.FirstName),
		)
	case "ban":
		_, err := chat.BanMember(b, int64(userId), &gotgbot.BanChatMemberOpts{})
		if err != nil {
			log.Error(err)
			return err
		}
		helpText = fmt.Sprintf("Admin %s banned %s from this chat!",
			helpers.MentionHtml(user.Id, user.FirstName),
			helpers.MentionHtml(int64(userId), actionUser.FirstName),
		)
	}

	_, _, err = query.Message.EditText(b,
		helpText,
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = query.Answer(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// unrestrict shows an inline keyboard to unrestrict a user (unban, unmute).
//
// Used to unrestrict members from a chat.
// Shows an inline keyboard menu which shows options to unban and unmute.

// Checks permissions and displays options for unrestricting the target user.
func (moduleStruct) unrestrict(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, chat, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, chat, false) {
		return ext.EndGroups
	}

	userId := extraction.ExtractUser(b, ctx)
	switch userId {
	case -1:
		return ext.EndGroups
	case 0:
		_, err := msg.Reply(b, tr.GetString("strings.common.errors.specify_user"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		_, err := msg.Reply(b, "This user is not in this chat, how can I restrict them?", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, "Why would I kick an admin? That sounds like a pretty dumb idea.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, "No u", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := msg.Reply(b, "How can I unrestrict this user?",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Unban", CallbackData: fmt.Sprintf("unrestrict.unban.%d", userId)},
						{Text: "Unmute", CallbackData: fmt.Sprintf("unrestrict.unmute.%d", userId)},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// unrestrictButtonHandler handles callback queries for the unrestrict command.
//
// Handles queries for unrestrict command.
// Performs the selected unrestriction action (unban, unmute) on the target user and updates the message accordingly.
func (moduleStruct) unrestrictButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := query.Message

	// permissions check
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	var helpText string

	action := args[0]
	userId, _ := strconv.Atoi(args[1])

	switch action {
	case "unmute":

		c, err := b.GetChat(chat.Id, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		_, err = chat.RestrictMember(b, int64(userId),
			*c.Permissions,
			nil,
		)
		if err != nil {
			log.Error(err)
			return err
		}

		helpText = fmt.Sprintf(
			"Unmuted by %s!",
			helpers.MentionHtml(user.Id, user.FirstName),
		)
	case "unban":
		_, err := chat.Unban(b,
			int64(userId),
			&gotgbot.UnbanChatMemberOpts{
				OnlyIfBanned: true,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}

		helpText = fmt.Sprintf(
			"Unbanned by %s !",
			helpers.MentionHtml(user.Id, user.FirstName),
		)
	}

	// type assertion to get the message
	_updatedMsg := msg.(*gotgbot.Message)

	// only strikethrough if msg.Text is non-empty
	if _updatedMsg.Text != "" {
		_updatedMsg.Text = fmt.Sprint("<s>", _updatedMsg.Text, "</s>", "\n\n")
	}

	_, _, err := msg.EditText(
		b,
		fmt.Sprint(_updatedMsg.Text, helpText),
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// LoadBans registers all ban, kick, restrict, and unrestrict command handlers with the dispatcher.
//
// This function enables the bans module and adds handlers for all moderation-related
// commands and callbacks. The module provides comprehensive user management with
// support for temporary restrictions and detailed enforcement actions.
//
// Registered commands:
//   - /ban: Permanently bans a user from the chat
//   - /sban: Silent ban (deletes command message)
//   - /tban: Temporary ban with configurable duration
//   - /dban: Delete ban (removes user's message and bans)
//   - /unban: Removes a ban and allows user to rejoin
//   - /kick: Kicks a user from the chat (they can rejoin)
//   - /dkick: Delete kick (removes user's message and kicks)
//   - /kickme: Allows users to kick themselves from the chat
//   - /restrict: Restricts user permissions without full ban
//   - /unrestrict: Removes restrictions and restores permissions
//
// The module includes comprehensive permission checking to ensure only
// authorized users can perform moderation actions. It handles various
// edge cases including anonymous users, bot self-protection, and admin immunity.
//
// Features:
//   - Permanent and temporary bans with duration support
//   - Silent moderation commands that auto-delete
//   - Message deletion combined with user actions
//   - Interactive restriction menus with callback buttons
//   - Flexible restriction system with granular permissions
//   - Comprehensive permission and safety checks
//   - Integration with warning and logging systems
//
// Requirements:
//   - Bot must be admin with ban/restrict permissions
//   - User must be admin with appropriate permissions
//   - Module respects admin hierarchy and bot limitations
//   - Integrates with i18n for multilingual support
//
// The bans system provides robust moderation tools while maintaining
// safety checks and proper permission enforcement.
func LoadBans(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(bansModule.moduleName, true)

	// ban cmds
	dispatcher.AddHandler(handlers.NewCommand("ban", bansModule.ban))
	dispatcher.AddHandler(handlers.NewCommand("sban", bansModule.sBan))
	dispatcher.AddHandler(handlers.NewCommand("tban", bansModule.tBan))
	dispatcher.AddHandler(handlers.NewCommand("dban", bansModule.dBan))
	dispatcher.AddHandler(handlers.NewCommand("unban", bansModule.unban))

	// kick cmds
	dispatcher.AddHandler(handlers.NewCommand("kick", bansModule.kick))
	dispatcher.AddHandler(handlers.NewCommand("dkick", bansModule.dkick))
	dispatcher.AddHandler(handlers.NewCommand("kickme", bansModule.kickme))

	// special commands
	dispatcher.AddHandler(handlers.NewCommand("restrict", bansModule.restrict))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("restrict."), bansModule.restrictButtonHandler))
	dispatcher.AddHandler(handlers.NewCommand("unrestrict", bansModule.unrestrict))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("unrestrict."), bansModule.unrestrictButtonHandler))
}
