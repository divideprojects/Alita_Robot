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

var bansModule = moduleStruct{moduleName: "Bans"}

/* Used to Kick a user from group

The Bot, Kicker should be admin with ban permissions in order to use this */

func (m moduleStruct) dkick(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, _ = msg.Reply(b, "Reply to a user's message to delete and kick him!", nil)
		return ext.EndGroups
	}

	_, reason := extraction.ExtractUserAndText(b, ctx)
	userId := msg.ReplyToMessage.From.Id
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, "This command cannot be used on anonymous user, these user can only be banned/unbanned.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
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

func (m moduleStruct) kick(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, err := msg.Reply(b, "This command cannot be used on anonymous user, these user can only be banned/unbanned.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

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
func (m moduleStruct) kickme(b *gotgbot.Bot, ctx *ext.Context) error {
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

func (m moduleStruct) tBan(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, err := msg.Reply(b, "This command cannot be used on anonymous user, these user can only be banned/unbanned.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
			helpers.Shtml())
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

func (m moduleStruct) ban(b *gotgbot.Bot, ctx *ext.Context) error {
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
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
			helpers.Shtml())
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

		baseStr := tr.GetString("strings." + m.moduleName + ".ban.normal_ban")
		if reason != "" {
			baseStr += fmt.Sprintf(tr.GetString("strings."+m.moduleName+".ban.ban_reason"), reason)
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

/* Used to Silently Ban a user from group

This deletes the command of Banner and also does not reply.

The Bot, Banner should be admin with ban permissions in order to use this */

func (m moduleStruct) sBan(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, err := msg.Reply(b, "This command cannot be used on anonymous user, these user can only be banned/unbanned.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
			helpers.Shtml())
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

func (m moduleStruct) dBan(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, err := msg.Reply(b, "This command cannot be used on anonymous user, these user can only be banned/unbanned.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
			helpers.Shtml())
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

/* Used to Unban a user from group

The Bot, Unbanner should be admin with ban permissions in order to use this */

func (m moduleStruct) unban(b *gotgbot.Bot, ctx *ext.Context) error {
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
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
			helpers.Shtml())
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
			tr.GetString("strings."+m.moduleName+".unban.unbanned_user"),
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

func (moduleStruct) restrict(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage

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
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
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

// Handles the queries fore restrict command
func (moduleStruct) restrictButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
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

/* Used to Unrestrict members from a chat
Shows an inline keyboard menu which shows options to unban and unmute */

func (moduleStruct) unrestrict(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage

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
		_, err := msg.Reply(b, "I don't know who you're talking about, you're going to need to specify a user...!",
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

// Handles queries for unrestrict command
func (moduleStruct) unrestrictButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
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

	// only strikethrough if msg.Text is non-empty
	if msg.Text != "" {
		msg.Text = fmt.Sprint("<s>", msg.Text, "</s>", "\n\n")
	}

	_, _, err := msg.EditText(
		b,
		fmt.Sprint(msg.Text, helpText),
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
