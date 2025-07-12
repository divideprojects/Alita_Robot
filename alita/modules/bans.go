package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/permissions"
)

/*
bansModule provides ban, kick, restrict, and unrestrict logic for group chats.

Implements all moderation actions related to user removal and restriction.
*/
var bansModule = moduleStruct{
	moduleName: "Bans",
	cfg:        nil, // will be set during LoadBans
}

// getActionText is a helper function to safely get action text with fallback
func getActionText(tr *i18n.I18n, key, fallback string) string {
	text, err := tr.GetStringWithError(key)
	if err != nil {
		log.Error(err)
		return fallback
	}
	return text
}

/* Used to Kick a user from group

The Bot, Kicker should be admin with ban permissions in order to use this */

/*
dkick deletes a user's message and kicks them from the group.

Performs permission checks, extracts the target user from a reply, deletes their message, and kicks them. Handles edge cases such as anonymous users and admins.

The Bot and the user issuing the command must have appropriate permissions.
*/
func (m moduleStruct) dkick(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

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
	if msg.ReplyToMessage == nil {
		tr := i18n.New(db.GetLanguage(ctx))
		noReplyMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_reply")
		if err != nil {
			log.Error(err)
			noReplyMsg = "Reply to a message to use this command"
		}
		_, _ = msg.Reply(b, noReplyMsg, nil)
		return ext.EndGroups
	}

	_, reason := extraction.ExtractUserAndText(b, ctx)
	userId := msg.ReplyToMessage.From.Id
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".errors.anon_user_kick"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		noUserMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_user_specified")
		if err != nil {
			log.Error(err)
			noUserMsg = "No user specified"
		}
		_, err = msg.Reply(b, noUserMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, _ = msg.ReplyToMessage.Delete(b, nil)

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".kick.user_not_in_chat"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".kick.cannot_kick_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".kick.is_bot_itself"), helpers.Shtml())
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

	kickedUserMsg, kickedUserErr := tr.GetStringWithError("strings." + m.moduleName + ".kick.kicked_user")
	if kickedUserErr != nil {
		log.Errorf("[bans] missing translation for kick.kicked_user: %v", kickedUserErr)
		kickedUserMsg = "User has been kicked."
	}
	baseStr := kickedUserMsg

	if reason != "" {
		kickedReasonMsg, kickedReasonErr := tr.GetStringWithError("strings." + m.moduleName + ".kick.kicked_reason")
		if kickedReasonErr != nil {
			log.Errorf("[bans] missing translation for kick.kicked_reason: %v", kickedReasonErr)
			kickedReasonMsg = "\n<b>Reason:</b> %s"
		}
		baseStr += fmt.Sprintf(kickedReasonMsg, reason)
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

/*
kick removes a user from the group.

Checks permissions, extracts the target user, and kicks them. Handles edge cases such as anonymous users, admins, and the bot itself.
*/
func (m moduleStruct) kick(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

	// Use helper for permission checks, user extraction, and protection validation
	userId, reason, ok := permissions.PerformCommonRestrictionChecks(b, ctx, permissions.CommonRestrictionPerms, true)
	if !ok {
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".kick.is_bot_itself"), helpers.Shtml())
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

	baseStr := tr.GetString("strings." + m.moduleName + ".kick.kicked_user")
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings."+m.moduleName+".kick.kicked_reason"), reason)
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

/*
	Used to kick a user from group

The Bot should be admin with ban permissions in order to use this
*/
/*
kickme allows a user to remove themselves from the group.

Admins are not allowed to use this command. The bot must have restriction permissions.
*/
func (m moduleStruct) kickme(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

	// Permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return ext.EndGroups
	}

	// Don't allow admins to use the command
	if chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".kickme.is_admin"), helpers.Shtml())
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

	_, err = msg.Reply(b, tr.GetString("strings."+m.moduleName+".kickme.ok_out"), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/* Used to temporarily ban a user from chat

The Bot, Kick should be admin with ban permissions in order to use this */

/*
tBan temporarily bans a user from the chat.

Performs permission checks, extracts the target user and ban duration, and bans them for the specified time. Handles edge cases such as anonymous users and admins.
*/
func (m moduleStruct) tBan(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

	// Use helper for permission checks, user extraction, and protection validation
	userId, reason, ok := permissions.PerformCommonRestrictionChecks(b, ctx, permissions.CommonRestrictionPerms, false)
	if !ok {
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".ban.is_bot_itself"), helpers.Shtml())
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
		tr.GetString("strings."+m.moduleName+".ban.tban"),
		helpers.MentionHtml(banUser.Id, banUser.FirstName),
		timeVal,
	)
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings."+m.moduleName+".ban.ban_reason"), reason)
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

/* Used to indefinitely ban a user from group

The Bot, Banner should be admin with ban permissions in order to use this */

/*
ban bans a user from the group indefinitely.

Checks permissions, extracts the target user, and bans them. Handles anonymous users, admins, and the bot itself. Provides an inline button for unbanning.
*/
func (m moduleStruct) ban(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
		noUserMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_user_specified")
		if err != nil {
			log.Error(err)
			noUserMsg = "No user specified"
		}
		_, err = msg.Reply(b, noUserMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".ban.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".ban.is_bot_itself"), helpers.Shtml())
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
			bannedUserMsg, err := tr.GetStringWithError("strings.Bans.banned_user")
			if err != nil {
				log.Error(err)
				bannedUserMsg = "User has been banned: "
			}
			text = bannedUserMsg + helpers.MentionHtml(userId, msg.ReplyToMessage.GetSender().Name())
		} else {
			text = tr.GetString("strings." + m.moduleName + ".errors.anon_user_ban_reply")
		}
		sendMsgOptns = helpers.Shtml()
	} else {
		_, err := chat.BanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		_, name, _ := extraction.GetUserInfo(userId)

		normalBanMsg, normalBanErr := tr.GetStringWithError("strings." + m.moduleName + ".ban.normal_ban")
		if normalBanErr != nil {
			log.Errorf("[bans] missing translation for ban.normal_ban: %v", normalBanErr)
			normalBanMsg = "User has been banned."
		}
		baseStr := normalBanMsg

		if reason != "" {
			banReasonMsg, banReasonErr := tr.GetStringWithError("strings." + m.moduleName + ".ban.ban_reason")
			if banReasonErr != nil {
				log.Errorf("[bans] missing translation for ban.ban_reason: %v", banReasonErr)
				banReasonMsg = "\n<b>Reason:</b> %s"
			}
			baseStr += fmt.Sprintf(banReasonMsg, reason)
		}

		text = fmt.Sprintf(baseStr, helpers.MentionHtml(userId, name))

		sendMsgOptns = &gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         getActionText(tr, "strings.CommonStrings.actions.unban", "Unban") + " (Admin Only)",
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

/* Used to Silently Ban a user from group

This deletes the command of Banner and also does not reply.

The Bot, Banner should be admin with ban permissions in order to use this */

/*
sBan silently bans a user from the group and deletes the command message.

Performs permission checks, extracts the target user, and bans them without sending a reply. Handles edge cases such as anonymous users and admins.
*/
func (moduleStruct) sBan(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage

	// Create permission config that also requires delete permissions
	banWithDeletePerms := permissions.CommonRestrictionPerms
	banWithDeletePerms.RequireBotDelete = true

	// Use helper for permission checks, user extraction, and protection validation
	userId, _, ok := permissions.PerformCommonRestrictionChecks(b, ctx, banWithDeletePerms, false)
	if !ok {
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

/* Used to Ban a user from group and delete their message

This deletes the message of replied user

The Bot, Banner should be admin with ban permissions in order to use this */

/*
dBan bans a user from the group and deletes their message.

Checks permissions, extracts the target user, deletes their message, and bans them. Handles anonymous users, admins, and edge cases.
*/
func (m moduleStruct) dBan(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

	// Create permission config that also requires delete permissions
	banWithDeletePerms := permissions.CommonRestrictionPerms
	banWithDeletePerms.RequireUserDelete = true
	banWithDeletePerms.RequireBotDelete = true

	// Use helper for permission checks, user extraction, and protection validation
	userId, reason, ok := permissions.PerformCommonRestrictionChecks(b, ctx, banWithDeletePerms, false)
	if !ok {
		return ext.EndGroups
	}

	if msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".ban.dban.no_reply"), helpers.Shtml())
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

	baseStr := tr.GetString("strings." + m.moduleName + ".ban.normal_ban")
	if reason != "" {
		baseStr += fmt.Sprintf(tr.GetString("strings."+m.moduleName+".ban.ban_reason"), reason)
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(baseStr, helpers.MentionHtml(banUser.Id, banUser.FirstName)),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         getActionText(tr, "strings.CommonStrings.actions.unban", "Unban") + " (Admin Only)",
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

/* Used to Unban a user from group

The Bot, Unbanner should be admin with ban permissions in order to use this */

/*
unban removes a ban from a user in the group.

Checks permissions, extracts the target user, and unbans them. Handles anonymous users, admins, and the bot itself.
*/
func (m moduleStruct) unban(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
		noUserMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_user_specified")
		if err != nil {
			log.Error(err)
			noUserMsg = "No user specified"
		}
		_, err = msg.Reply(b, noUserMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".unban.is_bot_itself"), helpers.Shtml())
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
			bannedUserMsg, err := tr.GetStringWithError("strings.Bans.banned_user")
			if err != nil {
				log.Error(err)
				bannedUserMsg = "User has been banned: "
			}
			text = bannedUserMsg + helpers.MentionHtml(userId, msg.ReplyToMessage.GetSender().Name())
		} else {
			text = tr.GetString("strings." + m.moduleName + ".errors.anon_user_ban_reply")
		}
	} else {
		var err error
		_, err = chat.UnbanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		banUser, err := b.GetChat(userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		unbannedUserMsg, unbannedUserErr := tr.GetStringWithError("strings." + m.moduleName + ".unban.unbanned_user")
		if unbannedUserErr != nil {
			log.Errorf("[bans] missing translation for unban.unbanned_user: %v", unbannedUserErr)
			unbannedUserMsg = "User %s has been unbanned."
		}

		text = fmt.Sprintf(
			unbannedUserMsg,
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

/* Used to Restrict members from a chat
Shows an inline keyboard menu which shows options to kick, ban and mute */

/*
restrict shows an inline keyboard to restrict a user (ban, kick, mute).

Checks permissions and displays options for restricting the target user.
*/
func (m moduleStruct) restrict(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

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
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
		noUserMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_user_specified")
		if err != nil {
			log.Error(err)
			noUserMsg = "No user specified"
		}
		_, err = msg.Reply(b, noUserMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		userNotInChatMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.user_not_in_chat")
		if err != nil {
			log.Error(err)
			userNotInChatMsg = "User not in chat"
		}
		_, err = msg.Reply(b, userNotInChatMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".errors.restrict_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		actionOnSelfMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.action_on_self")
		if err != nil {
			log.Error(err)
			actionOnSelfMsg = "Cannot perform action on self"
		}
		_, err = msg.Reply(b, actionOnSelfMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".restrict.question"),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: getActionText(tr, "strings.CommonStrings.actions.ban", "Ban"), CallbackData: fmt.Sprintf("restrict.ban.%d", userId)},
						{Text: getActionText(tr, "strings.CommonStrings.actions.kick", "Kick"), CallbackData: fmt.Sprintf("restrict.kick.%d", userId)},
					},
					{{Text: getActionText(tr, "strings.CommonStrings.actions.mute", "Mute"), CallbackData: fmt.Sprintf("restrict.mute.%d", userId)}},
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

// Handles the queries fore restrict command
/*
restrictButtonHandler handles callback queries for the restrict command.

Performs the selected restriction action (ban, kick, mute) on the target user.
*/
func (m moduleStruct) restrictButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	tr := i18n.New(db.GetLanguage(ctx))

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
		helpText = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".restrict.kick_success"),
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
		helpText = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".restrict.mute_success"),
			helpers.MentionHtml(user.Id, user.FirstName),
			helpers.MentionHtml(int64(userId), actionUser.FirstName),
		)
	case "ban":
		_, err := chat.BanMember(b, int64(userId), &gotgbot.BanChatMemberOpts{})
		if err != nil {
			log.Error(err)
			return err
		}
		helpText = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".restrict.ban_success"),
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

/* Used to Unrestrict members from a chat
Shows an inline keyboard menu which shows options to unban and unmute */

/*
unrestrict shows an inline keyboard to unrestrict a user (unban, unmute).

Checks permissions and displays options for unrestricting the target user.
*/
func (m moduleStruct) unrestrict(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

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
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
		noUserMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_user_specified")
		if err != nil {
			log.Error(err)
			noUserMsg = "No user specified"
		}
		_, err = msg.Reply(b, noUserMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		userNotInChatMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.user_not_in_chat")
		if err != nil {
			log.Error(err)
			userNotInChatMsg = "User not in chat"
		}
		_, err = msg.Reply(b, userNotInChatMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.IsUserBanProtected(b, ctx, nil, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".errors.restrict_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		actionOnSelfMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.action_on_self")
		if err != nil {
			log.Error(err)
			actionOnSelfMsg = "Cannot perform action on self"
		}
		_, err = msg.Reply(b, actionOnSelfMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".unrestrict.question"),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: getActionText(tr, "strings.CommonStrings.actions.unban", "Unban"), CallbackData: fmt.Sprintf("unrestrict.unban.%d", userId)},
						{Text: getActionText(tr, "strings.CommonStrings.actions.unmute", "Unmute"), CallbackData: fmt.Sprintf("unrestrict.unmute.%d", userId)},
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

// Handles queries for unrestrict command
/*
unrestrictButtonHandler handles callback queries for the unrestrict command.

Performs the selected unrestriction action (unban, unmute) on the target user and updates the message accordingly.
*/
func (m moduleStruct) unrestrictButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := query.Message
	tr := i18n.New(db.GetLanguage(ctx))

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
			tr.GetString("strings."+m.moduleName+".unrestrict.unmute_success"),
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
			tr.GetString("strings."+m.moduleName+".unrestrict.unban_success"),
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

/*
LoadBans registers all ban, kick, restrict, and unrestrict command handlers with the dispatcher.

Enables the bans module and adds handlers for all moderation-related commands and callbacks.
*/
func LoadBans(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	bansModule.cfg = cfg

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
