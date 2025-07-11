package modules

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/permissions"
)

var mutesModule = moduleStruct{moduleName: "Mutes"}

/* Used to temporarily mute a user from group

The Bot, Muter should be admin with restrict permissions in order to use this */

func (m moduleStruct) tMute(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage

	// Use helper for permission checks, user extraction, and protection validation
	userId, reason, ok := permissions.PerformCommonRestrictionChecks(b, ctx, permissions.CommonRestrictionPerms, true)
	if !ok {
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("Mutes.errors.restrict_self"), helpers.Shtml())
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

	_, err := chat.RestrictMember(b, userId, gotgbot.ChatPermissions{
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
		&gotgbot.RestrictChatMemberOpts{
			UntilDate: _time,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	muteUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := fmt.Sprintf(
		"Shh...\nMuted %s for %s",
		helpers.MentionHtml(muteUser.Id, muteUser.FirstName),
		timeVal,
	)
	if reason != "" {
		baseStr += "\n<b>Reason: </b>" + reason
	}

	_, err = msg.Reply(b,
		baseStr,
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/* Used to mute a user from group

The Bot, Muter should be admin with restrict permissions in order to use this */

func (m moduleStruct) mute(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Use helper for permission checks, user extraction, and protection validation
	userId, reason, ok := permissions.PerformCommonRestrictionChecks(b, ctx, permissions.CommonRestrictionPerms, true)
	if !ok {
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("Mutes.errors.restrict_self"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	_, err := chat.RestrictMember(b, userId, gotgbot.ChatPermissions{
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
	}, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	muteUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := "Shh...\nMuted %s."
	if reason != "" {
		baseStr += "\n<b>Reason: </b>" + reason
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(baseStr, helpers.MentionHtml(muteUser.Id, muteUser.FirstName)),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         tr.GetString("strings.Mute.unmute_admin_only"),
							CallbackData: fmt.Sprintf("unrestrict.unmute.%d", userId),
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

/* Used to silently mute a user from group

The Bot, Muter should be admin with restrict permissions in order to use this

The message of muter will be deleted after sending this command */

func (m moduleStruct) sMute(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage

	// Create permission config that also requires delete permissions
	muteWithDeletePerms := permissions.CommonRestrictionPerms
	muteWithDeletePerms.RequireBotDelete = true

	// Use helper for permission checks, user extraction, and protection validation
	userId, _, ok := permissions.PerformCommonRestrictionChecks(b, ctx, muteWithDeletePerms, true)
	if !ok {
		return ext.EndGroups
	}

	_, err := chat.RestrictMember(b, userId, gotgbot.ChatPermissions{
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
	}, nil)
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

/* Used to mute a user from group and delete their message

The Bot, Muter should be admin with restrict permissions in order to use this

Used as a reply to a message and delete the replied message*/

func (m moduleStruct) dMute(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Create permission config that also requires delete permissions
	muteWithDeletePerms := permissions.CommonRestrictionPerms
	muteWithDeletePerms.RequireBotDelete = true

	// Use helper for permission checks, user extraction, and protection validation
	userId, reason, ok := permissions.PerformCommonRestrictionChecks(b, ctx, muteWithDeletePerms, true)
	if !ok {
		return ext.EndGroups
	}

	if msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, tr.GetString("Mutes.dmute.no_reply"), helpers.Shtml())
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

	_, err = chat.RestrictMember(b, userId, gotgbot.ChatPermissions{
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
	}, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	muteUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	baseStr := "Shh...\nMuted %s."
	if reason != "" {
		baseStr += "\n<b>Reason: </b>" + reason
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(baseStr, helpers.MentionHtml(muteUser.Id, muteUser.FirstName)),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         tr.GetString("strings.Mute.unmute_admin_only"),
							CallbackData: fmt.Sprintf("unrestrict.unmute.%d", userId),
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

/* Used to Unmute a user from group

The Bot, Unmuter should be admin with restrict permissions in order to use this */

func (m moduleStruct) unmute(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage

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
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("Warns.errors.anon_user"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.CommonStrings.errors.no_user_specified"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// User should be in chat for getting restricted
	if !chat_status.IsUserInChat(b, chat, userId) {
		_, err := msg.Reply(b, tr.GetString("CommonStrings.errors.user_not_in_chat"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("Mutes.errors.restrict_self"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	c, err := b.GetChat(chat.Id, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// should give the current chat permissions to the users who is unmuted
	_, err = chat.RestrictMember(
		b,
		userId,
		*c.Permissions,
		nil,
	)
	if err != nil {
		log.Error(err)
		return err
	}

	muteUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = msg.Reply(b,
		fmt.Sprintf(
			"Alright!\nI'll allow %s to speak again.",
			helpers.MentionHtml(muteUser.Id, muteUser.FirstName),
		),
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func LoadMutes(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(mutesModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("mute", mutesModule.mute))
	dispatcher.AddHandler(handlers.NewCommand("smute", mutesModule.sMute))
	dispatcher.AddHandler(handlers.NewCommand("tmute", mutesModule.tMute))
	dispatcher.AddHandler(handlers.NewCommand("dmute", mutesModule.dMute))
	dispatcher.AddHandler(handlers.NewCommand("unmute", mutesModule.unmute))
}
