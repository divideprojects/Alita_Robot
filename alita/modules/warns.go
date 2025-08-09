package modules

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var warnsModule = moduleStruct{moduleName: "Warns"}

// setWarnMode handles the /setwarnmode command to configure the action
// taken when users reach the warning limit (ban, kick, or mute).
func (moduleStruct) setWarnMode(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]
	user := ctx.EffectiveSender.User

	// permissions check
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	var replyText string

	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "ban":
			go db.SetWarnMode(chat.Id, "ban")
			replyText = tr.Message("warns_warn_mode_updated", i18n.Params{"mode": "ban"})
		case "kick":
			go db.SetWarnMode(chat.Id, "kick")
			replyText = tr.Message("warns_warn_mode_updated", i18n.Params{"mode": "kick"})
		case "mute":
			go db.SetWarnMode(chat.Id, "mute")
			replyText = tr.Message("warns_warn_mode_updated", i18n.Params{"mode": "mute"})
		default:
			replyText = tr.Message("warns_warn_mode_unknown", i18n.Params{"type": args[0]})
		}
	} else {
		replyText = tr.Message("warns_warn_mode_specify", nil)
	}

	_, err := msg.Reply(b, replyText, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// warnThisUser is a helper function that performs the actual warning process,
// including limit checking and enforcement of warn mode actions.
func (moduleStruct) warnThisUser(b *gotgbot.Bot, ctx *ext.Context, userId int64, reason, warnType string) (err error) {
	var (
		reply    string
		keyboard gotgbot.InlineKeyboardMarkup
	)

	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// permissions check
	if chat_status.IsUserAdmin(b, chat.Id, userId) {
		_, err = msg.Reply(b, tr.Message("warns_cannot_warn_admin", nil), nil)
		return err
	}

	switch warnType {
	case "dwarn":
		_, err := msg.ReplyToMessage.Delete(b, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	case "swarn":
		_, err := msg.Delete(b, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	chatMember, err := chat.GetMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	u := chatMember.MergeChatMember().User
	warnrc := db.GetWarnSetting(chat.Id)
	numWarns, reasons := db.WarnUser(userId, chat.Id, reason)

	if numWarns >= warnrc.WarnLimit {
		db.ResetUserWarns(userId, chat.Id)
		switch warnrc.WarnMode {
		case "kick":
			_, err = chat.BanMember(b, userId, nil)
			reply = tr.Message("warns_limit_exceeded_kick", i18n.Params{"user": helpers.MentionHtml(u.Id, u.FirstName), "current": numWarns, "limit": warnrc.WarnLimit})
			if err != nil {
				log.Errorf("[warn] warnlimit: kick (%d) - %s", userId, err)
				return err
			}
		case "mute":
			_, err = chat.RestrictMember(b, userId,
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
			reply = tr.Message("warns_limit_exceeded_mute", i18n.Params{"user": helpers.MentionHtml(u.Id, u.FirstName), "current": numWarns, "limit": warnrc.WarnLimit})
			if err != nil {
				log.Errorf("[warn] warnlimit: mute (%d) - %s", userId, err)
				return err
			}
		case "ban":
			_, err = chat.BanMember(b, userId, nil)
			reply = tr.Message("warns_limit_exceeded_ban", i18n.Params{"user": helpers.MentionHtml(u.Id, u.FirstName), "current": numWarns, "limit": warnrc.WarnLimit})
			if err != nil {
				log.Errorf("[warn] warnlimit: ban (%d) - %s", userId, err)
				return err
			}
		}
		for _, warnReason := range reasons {
			reply += fmt.Sprintf("\n - %s", html.EscapeString(warnReason))
		}
	} else {
		rules := db.GetChatRulesInfo(chat.Id)
		if len(rules.Rules) >= 1 {
			keyboard = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         "❌ Remove warn",
							CallbackData: fmt.Sprintf("rmWarn.%d", u.Id),
						},
						{
							Text: "Rules 📝",
							Url:  fmt.Sprintf("t.me/%s?start=rules_%d", b.Username, chat.Id),
						},
					},
				},
			}
		} else {
			keyboard = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         "❌ Remove warn",
							CallbackData: fmt.Sprintf("rmWarn.%d", u.Id),
						},
					},
				},
			}
		}

		reply = tr.Message("warns_user_warned", i18n.Params{"user": helpers.MentionHtml(u.Id, u.FirstName), "current": numWarns, "limit": warnrc.WarnLimit})

		if reason != "" {
			reply = tr.Message("warns_user_warned_reason", i18n.Params{"user": helpers.MentionHtml(u.Id, u.FirstName), "current": numWarns, "limit": warnrc.WarnLimit, "reason": html.EscapeString(reason)})
		}
	}
	_, err = b.SendMessage(chat.Id, reply,
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                msg.MessageId,
				AllowSendingWithoutReply: true,
			},
			ReplyMarkup: &keyboard,
		},
	)
	if err != nil {
		log.Errorf("[warn] sendMessage (%d) - %s", userId, err)
		return err
	}

	return ext.EndGroups
}

// warnUser handles the /warn command to issue warnings to users
// with optional reasons, requiring admin permissions.
func (m moduleStruct) warnUser(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check permissions
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
		_, err := msg.Reply(b, tr.Message("error_anonymous_user_not_supported", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.Message("error_specify_user", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if !chat_status.IsUserInChat(b, chat, userId) {
		return ext.EndGroups
	}
	var warnusr int64
	if msg.ReplyToMessage != nil {
		warnusr = msg.ReplyToMessage.From.Id
	} else {
		warnusr = userId
	}

	return m.warnThisUser(b, ctx, warnusr, reason, "warn")
}

// sWarnUser handles the /swarn command to silently warn users
// by deleting the command message, requiring admin permissions.
func (m moduleStruct) sWarnUser(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check permissions
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
		_, err := msg.Reply(b, tr.Message("error_anonymous_user_not_supported", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.Message("error_specify_user", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if !chat_status.IsUserInChat(b, chat, userId) {
		return ext.EndGroups
	}
	var warnusr int64
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From.Id == userId {
		warnusr = msg.ReplyToMessage.From.Id
	} else {
		warnusr = userId
	}

	return m.warnThisUser(b, ctx, warnusr, reason, "swarn")
}

// dWarnUser handles the /dwarn command to warn users and delete
// the message they replied to, requiring admin permissions.
func (m moduleStruct) dWarnUser(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check permissions
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
		_, err := msg.Reply(b, tr.Message("error_anonymous_user_not_supported", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.Message("error_specify_user", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if !chat_status.IsUserInChat(b, chat, userId) {
		return ext.EndGroups
	}
	var warnusr int64
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From.Id == userId {
		warnusr = msg.ReplyToMessage.From.Id
	} else {
		warnusr = userId
	}

	return m.warnThisUser(b, ctx, warnusr, reason, "dwarn")
}

// warnings handles the /warnings command to display current
// warning settings including limit and enforcement mode.
func (moduleStruct) warnings(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check permissions
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	warnrc := db.GetWarnSetting(chat.Id)
	text := tr.Message("warns_current_settings", i18n.Params{"limit": warnrc.WarnLimit, "mode": warnrc.WarnMode})
	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// warns handles the /warns command to check the warning count
// and reasons for a specific user or the command sender.
func (moduleStruct) warns(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "warns") {
		return ext.EndGroups
	}

	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		userId = ctx.EffectiveUser.Id
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.Message("error_anonymous_user_not_supported", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.Message("error_specify_user", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	numWarns, reasons := db.GetWarns(userId, chat.Id)
	text := ""

	if numWarns != 0 {
		warnrc := db.GetWarnSetting(chat.Id)
		if len(reasons) > 0 {
			text = tr.Message("warns_user_warnings_with_reasons", i18n.Params{"current": numWarns, "limit": warnrc.WarnLimit})
			for _, reason := range reasons {
				text += fmt.Sprintf("\n - %s", reason)
			}
			msgs := helpers.SplitMessage(text)
			for _, msgText := range msgs {
				_, err := msg.Reply(b, msgText, nil)
				if err != nil {
					log.Error(err)
					return err
				}
			}
		} else {
			_, err := msg.Reply(b, tr.Message("warns_user_warnings_no_reasons", i18n.Params{"current": numWarns, "limit": warnrc.WarnLimit}), nil)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		_, err := msg.Reply(b, tr.Message("warns_no_warnings", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// rmWarnButton processes callback queries from remove warning buttons
// to remove the latest warning from a user, requiring admin permissions.
func (moduleStruct) rmWarnButton(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := ctx.EffectiveSender.User
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check permissions
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	userMatch := args[1]
	userId, _ := strconv.Atoi(userMatch)
	var replyText string

	res := db.RemoveWarn(int64(userId), chat.Id)
	if res {
		replyText = tr.Message("warns_warning_removed_by", i18n.Params{"admin": helpers.MentionHtml(user.Id, user.FirstName)})
	} else {
		replyText = tr.Message("warns_no_warnings_to_remove", nil)
	}

	_, _, err := query.Message.EditText(
		b,
		replyText,
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

// setWarnLimit handles the /setwarnlimit command to configure
// the maximum number of warnings before enforcement action.
func (moduleStruct) setWarnLimit(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// Check permissions
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	var replyText string

	if len(args) == 0 {
		replyText = tr.Message("warns_limit_specify", nil)
	} else {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			replyText = tr.Message("warns_limit_not_integer", i18n.Params{"value": args[0]})
		} else {
			if num < 1 || num > 100 {
				replyText = tr.Message("warns_limit_range_error", nil)
			} else {
				go db.SetWarnLimit(chat.Id, num)
				replyText = tr.Message("warns_limit_updated", i18n.Params{"limit": num})
			}
		}
	}

	_, err := msg.Reply(b, replyText, helpers.Smarkdown())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// resetWarns handles the /resetwarns command to clear all warnings
// for a specific user, requiring admin permissions.
func (moduleStruct) resetWarns(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check permissions
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.Message("error_anonymous_user_not_supported", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.Message("error_specify_user", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	db.ResetUserWarns(userId, chat.Id)
	_, err := msg.Reply(b, tr.Message("warns_warnings_reset", nil), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// resetAllWarns handles the /resetallwarns command to clear all warnings
// for all users in the chat with confirmation, restricted to owners.
func (moduleStruct) resetAllWarns(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// Check if group or not
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	warnrc := db.GetAllChatWarns(chat.Id)
	if warnrc == 0 {
		_, err := msg.Reply(b, tr.Message("warns_no_users_warned", nil), helpers.Shtml())
		return err
	}

	if chat_status.RequireUserOwner(b, ctx, chat, user.Id, false) {
		_, err := msg.Reply(b, tr.Message("warns_confirm_reset_all", nil),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{Text: "Yes", CallbackData: "rmAllChatWarns.yes"},
							{Text: "No", CallbackData: "rmAllChatWarns.no"},
						},
					},
				},
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

// warnsButtonHandler processes callback queries for the reset all warnings
// confirmation dialog, restricted to chat owners.
func (moduleStruct) warnsButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From

	if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	response := args[1]
	var helpText string

	switch response {
	case "yes":
		go db.ResetAllChatWarns(query.Message.GetChat().Id)
		helpText = "Removed all warns of all the users in this chat !"
	case "no":
		helpText = "Cancelled the removal of all the warns of all the users in this chat !"
	}

	_, _, err := query.Message.EditText(
		b,
		"Removed all warns of all users in this chat.",
		nil,
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b,
		&gotgbot.AnswerCallbackQueryOpts{
			Text: helpText,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// LoadWarns registers all warns module handlers with the dispatcher,
// including warning commands and callback handlers.
func LoadWarns(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(warnsModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("warn", warnsModule.warnUser))
	dispatcher.AddHandler(handlers.NewCommand("swarn", warnsModule.sWarnUser))
	dispatcher.AddHandler(handlers.NewCommand("dwarn", warnsModule.dWarnUser))
	dispatcher.AddHandler(handlers.NewCommand("resetwarns", warnsModule.resetWarns))
	dispatcher.AddHandler(handlers.NewCommand("warns", warnsModule.warns))
	misc.AddCmdToDisableable("warns")
	dispatcher.AddHandler(handlers.NewCommand("setwarnlimit", warnsModule.setWarnLimit))
	dispatcher.AddHandler(handlers.NewCommand("setwarnmode", warnsModule.setWarnMode))
	dispatcher.AddHandler(handlers.NewCommand("resetallwarns", warnsModule.resetAllWarns))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("rmAllChatWarns"), warnsModule.warnsButtonHandler))
	dispatcher.AddHandler(handlers.NewCommand("warnings", warnsModule.warnings))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("rmWarn"), warnsModule.rmWarnButton))
}
