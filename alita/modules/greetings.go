package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatjoinrequest"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

// greetingsModule provides logic for managing welcome and goodbye messages in group chats.
//
// Implements commands to set, reset, and configure greetings, as well as handlers for join/leave events and join requests.
var greetingsModule = moduleStruct{moduleName: "Greetings"}

// welcome displays or toggles the welcome message settings for the chat.
//
// Admins can view current settings, toggle welcoming on/off, or display the current welcome message.
func (moduleStruct) welcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// Save original context for replies to PM before connection check
	originalCtx := *ctx
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	var wlcmText string

	if len(args) == 0 || strings.ToLower(args[0]) == "noformat" {
		noformat := len(args) > 0 && strings.ToLower(args[0]) == "noformat"
		welcPrefs := db.GetGreetingSettings(chat.Id)
		wlcmText = welcPrefs.WelcomeSettings.WelcomeText
		_, err := msg.Reply(bot, fmt.Sprintf("I am currently welcoming users: <code>%t</code>"+
			"\nI am currently deleting old welcomes: <code>%t</code>"+
			"\nI am currently deleting service messages: <code>%t</code>"+
			"\nThe welcome message not filling the {} is:",
			welcPrefs.WelcomeSettings.ShouldWelcome,
			welcPrefs.WelcomeSettings.CleanWelcome,
			welcPrefs.ShouldCleanService), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		buttons := db.GetWelcomeButtons(chat.Id)

		if noformat {
			wlcmText += helpers.RevertButtons(buttons)
			_, err := helpers.GreetingsEnumFuncMap[welcPrefs.WelcomeSettings.WelcomeType](bot, &originalCtx, wlcmText, welcPrefs.WelcomeSettings.FileID, &gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil})
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			wlcmText, buttons = helpers.FormattingReplacer(bot, chat, user, wlcmText, buttons)
			keyb := helpers.BuildKeyboard(buttons)
			keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}
			_, err := helpers.GreetingsEnumFuncMap[welcPrefs.WelcomeSettings.WelcomeType](bot, &originalCtx, wlcmText, welcPrefs.WelcomeSettings.FileID, &keyboard)
			if err != nil {
				log.Error(err)
				return err
			}
		}

	} else if len(args) >= 1 {
		var err error
		switch strings.ToLower(args[0]) {
		case "on", "yes":
			db.SetWelcomeToggle(chat.Id, true)
			_, err = msg.Reply(bot, "I'll welcome users from now on.", helpers.Shtml())
		case "off", "no":
			db.SetWelcomeToggle(chat.Id, false)
			_, err = msg.Reply(bot, "I'll not welcome users from now on.", helpers.Shtml())
		default:
			_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
		}

		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// setWelcome sets a custom welcome message for the chat.
//
// Admins can set the message content, type, and buttons. Handles input validation and replies with the result.
func (moduleStruct) setWelcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	text, dataType, content, buttons, errorMsg := helpers.GetWelcomeType(msg, "welcome")
	if dataType == -1 {
		_, err := msg.Reply(bot, errorMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	db.SetWelcomeText(chat.Id, text, content, buttons, dataType)
	_, err := msg.Reply(bot, "Successfully set custom welcome message!", helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// resetWelcome resets the welcome message to the default.
//
// Admins can use this to remove any custom welcome message.
func (moduleStruct) resetWelcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	go db.SetWelcomeText(chat.Id, db.DefaultWelcome, "", nil, db.TEXT)
	_, err := msg.Reply(bot, "Successfully reset custom welcome message to default!", helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// goodbye displays or toggles the goodbye message settings for the chat.
//
// Admins can view current settings, toggle goodbyes on/off, or display the current goodbye message.
func (moduleStruct) goodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	var gdbyeText string

	if len(args) == 0 || strings.ToLower(args[0]) == "noformat" {
		noformat := len(args) > 0 && strings.ToLower(args[0]) == "noformat"
		gdbyePrefs := db.GetGreetingSettings(chat.Id)
		gdbyeText = gdbyePrefs.GoodbyeSettings.GoodbyeText
		_, err := msg.Reply(bot, fmt.Sprintf("I am currently goodbying users: <code>%t</code>"+
			"\nI am currently deleting old goodbyes: <code>%t</code>"+
			"\nI am currently deleting service messages: <code>%t</code>"+
			"\nThe goodbye message not filling the {} is:",
			gdbyePrefs.GoodbyeSettings.ShouldGoodbye,
			gdbyePrefs.GoodbyeSettings.CleanGoodbye,
			gdbyePrefs.ShouldCleanService), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		buttons := db.GetGoodbyeButtons(chat.Id)

		if noformat {
			gdbyeText += helpers.RevertButtons(buttons)
			_, err := helpers.GreetingsEnumFuncMap[gdbyePrefs.GoodbyeSettings.GoodbyeType](bot, ctx, gdbyeText, gdbyePrefs.GoodbyeSettings.FileID, &gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil})
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			gdbyeText, buttons = helpers.FormattingReplacer(bot, chat, user, gdbyeText, buttons)
			keyb := helpers.BuildKeyboard(buttons)
			keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}
			_, err := helpers.GreetingsEnumFuncMap[gdbyePrefs.GoodbyeSettings.GoodbyeType](bot, ctx, gdbyeText, gdbyePrefs.GoodbyeSettings.FileID, &keyboard)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else if len(args) >= 1 {
		var err error
		switch strings.ToLower(args[0]) {
		case "on", "yes":
			db.SetGoodbyeToggle(chat.Id, true)
			_, err = msg.Reply(bot, "I'll goodbye users from now on.", helpers.Shtml())
		case "off", "no":
			db.SetGoodbyeToggle(chat.Id, false)
			_, err = msg.Reply(bot, "I'll not goodbye users from now on.", helpers.Shtml())
		default:
			_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
		}
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// setGoodbye sets a custom goodbye message for the chat.
//
// Admins can set the message content, type, and buttons. Handles input validation and replies with the result.
func (moduleStruct) setGoodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	text, dataType, content, buttons, errorMsg := helpers.GetWelcomeType(msg, "goodbye")
	if dataType == -1 {
		_, err := msg.Reply(bot, errorMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	db.SetGoodbyeText(chat.Id, text, content, buttons, dataType)
	_, err := msg.Reply(bot, "Successfully set custom goodbye message!", helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// resetGoodbye resets the goodbye message to the default.
//
// Admins can use this to remove any custom goodbye message.
func (moduleStruct) resetGoodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if chat == nil {
		return ext.EndGroups
	}
	go db.SetGoodbyeText(chat.Id, db.DefaultGoodbye, "", nil, db.TEXT)
	_, err := msg.Reply(bot, "Successfully reset custom goodbye message to default!", helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// cleanWelcome toggles or displays the setting for deleting old welcome messages.
//
// Admins can enable or disable automatic deletion of old welcome messages.
func (moduleStruct) cleanWelcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]
	var err error
	user := ctx.EffectiveSender.User
	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		var err error
		cleanPref := db.GetGreetingSettings(chat.Id).WelcomeSettings.CleanWelcome
		if !cleanPref {
			_, err = msg.Reply(bot, "I should be deleting welcome messages up to two days old.", helpers.Shtml())
		} else {
			_, err = msg.Reply(bot, "I'm currently not deleting old welcome messages!", helpers.Shtml())
		}
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		db.SetCleanWelcomeSetting(chat.Id, false)
		_, err = msg.Reply(bot, "I'll not delete old welcome messages!", helpers.Shtml())
	case "on", "yes":
		db.SetCleanWelcomeSetting(chat.Id, true)
		_, err = msg.Reply(bot, "I'll try to delete old welcome messages!", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// cleanGoodbye toggles or displays the setting for deleting old goodbye messages.
//
// Admins can enable or disable automatic deletion of old goodbye messages.
func (moduleStruct) cleanGoodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]
	var err error
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		var err error
		cleanPref := db.GetGreetingSettings(chat.Id).GoodbyeSettings.CleanGoodbye
		if !cleanPref {
			_, err = msg.Reply(bot, "I should be deleting welcome messages up to two days old.", helpers.Shtml())
		} else {
			_, err = msg.Reply(bot, "I'm currently not deleting old welcome messages!", helpers.Shtml())
		}
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		db.SetCleanGoodbyeSetting(chat.Id, false)
		_, err = msg.Reply(bot, "I'll not delete old goodbye messages!", helpers.Shtml())
	case "on", "yes":
		db.SetCleanGoodbyeSetting(chat.Id, true)
		_, err = msg.Reply(bot, "I'll try to delete old goodbye messages!", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// delJoined toggles or displays the setting for deleting "user joined" service messages.
//
// Admins can enable or disable automatic deletion of join messages.
func (moduleStruct) delJoined(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]
	var err error
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		delPref := db.GetGreetingSettings(chat.Id).ShouldCleanService
		if delPref {
			_, err = msg.Reply(bot, "I should be deleting `user` joined the chat messages now.", helpers.Smarkdown())
		} else {
			_, err = msg.Reply(bot, "I'm currently not deleting joined messages.", helpers.Shtml())
		}
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		db.SetShouldCleanService(chat.Id, false)
		_, err = msg.Reply(bot, "I won't delete joined messages.", helpers.Shtml())
	case "on", "yes":
		db.SetShouldCleanService(chat.Id, true)
		_, err = msg.Reply(bot, "I'll try to delete joined messages!", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// newMember handles the event when a new member joins the chat.
//
// Sends a welcome message if enabled and manages deletion of previous welcome messages if configured.
func (moduleStruct) newMember(bot *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	newMember := ctx.ChatMember.NewChatMember.MergeChatMember().User
	greetPrefs := db.GetGreetingSettings(chat.Id)

	// when bot joins stop all updates of the groups
	// we use bot_updates for this
	if newMember.Id == bot.Id {
		return ext.EndGroups
	}

	if greetPrefs.WelcomeSettings.ShouldWelcome {
		buttons := db.GetWelcomeButtons(chat.Id)
		res, buttons := helpers.FormattingReplacer(bot, chat, &newMember,
			greetPrefs.WelcomeSettings.WelcomeText,
			buttons,
		)
		keyboard := &gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
		sent, err := helpers.GreetingsEnumFuncMap[greetPrefs.WelcomeSettings.WelcomeType](bot, ctx, res, greetPrefs.WelcomeSettings.FileID, keyboard)
		if err != nil {
			log.Error(err)
			return err
		}
		if greetPrefs.WelcomeSettings.CleanWelcome {
			_, _ = bot.DeleteMessage(chat.Id, greetPrefs.WelcomeSettings.LastMsgId, nil)
			db.SetCleanWelcomeMsgId(chat.Id, sent.MessageId)
			// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
			// 	log.WithFields(
			// 		log.Fields{
			// 			"chat": chat.Id,
			// 		},
			// 	).Error("error deleting message")
			// 	return ext.EndGroups
			// }
		}
	}
	return ext.EndGroups
}

// leftMember handles the event when a member leaves the chat.
//
// Sends a goodbye message if enabled and manages deletion of previous goodbye messages if configured.
func (moduleStruct) leftMember(bot *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	leftMember := ctx.ChatMember.OldChatMember.MergeChatMember().User
	greetPrefs := db.GetGreetingSettings(chat.Id)

	// when bot leaves stop all updates of the groups
	if leftMember.Id == bot.Id {
		return ext.EndGroups
	}

	if greetPrefs.GoodbyeSettings.ShouldGoodbye {
		buttons := db.GetGoodbyeButtons(chat.Id)
		res, buttons := helpers.FormattingReplacer(bot, chat, &leftMember, greetPrefs.GoodbyeSettings.GoodbyeText, buttons)
		keyboard := &gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
		sent, err := helpers.GreetingsEnumFuncMap[greetPrefs.GoodbyeSettings.GoodbyeType](bot, ctx, res, greetPrefs.GoodbyeSettings.FileID, keyboard)
		if err != nil {
			log.Error(err)
			return err
		}

		if greetPrefs.GoodbyeSettings.CleanGoodbye {
			_, _ = bot.DeleteMessage(chat.Id, greetPrefs.GoodbyeSettings.LastMsgId, nil)
			db.SetCleanGoodbyeMsgId(chat.Id, sent.MessageId)
			// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
			// 	log.WithFields(
			// 		log.Fields{
			// 			"chat": chat.Id,
			// 		},
			// 	).Error("error deleting message")
			// 	return ext.EndGroups
			// }
		}
	}
	return ext.EndGroups
}

// cleanService deletes service messages if the setting is enabled.
//
// Used for cleaning up join/leave notifications and other service messages.
func (moduleStruct) cleanService(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	if user.Id == bot.Id {
		return ext.EndGroups
	}

	greetPrefs := db.GetGreetingSettings(chat.Id)

	if greetPrefs.ShouldCleanService {
		_, err := msg.Delete(bot, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// pendingJoins handles new chat join requests.
//
// If auto-approve is enabled, approves the request. Otherwise, notifies admins and tracks pending requests.
func (m moduleStruct) pendingJoins(bot *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.ChatJoinRequest.Chat
	user := ctx.ChatJoinRequest.From
	joinReqStr := "join_request"

	if !m.loadPendingJoins(chat.Id, user.Id) {

		// auto approve join requests
		if db.GetGreetingSettings(chat.Id).ShouldAutoApprove {
			_, _ = bot.ApproveChatJoinRequest(chat.Id, user.Id, nil)
			return ext.ContinueGroups
		}

		_, err := bot.SendMessage(
			chat.Id,
			fmt.Sprint(
				"A new user has requested to join chat!",
				fmt.Sprintf("\nUser: %s", helpers.MentionHtml(user.Id, user.FirstName)),
				fmt.Sprintf("\nUser ID: %d", user.Id),
			),
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         "✅ Approve",
								CallbackData: fmt.Sprintf("%s.accept.%d", joinReqStr, user.Id),
							},
							{
								Text:         "❌ Decline",
								CallbackData: fmt.Sprintf("%s.decline.%d", joinReqStr, user.Id),
							},
						},
						{
							{
								Text:         "✅ Ban",
								CallbackData: fmt.Sprintf("%s.ban.%d", joinReqStr, user.Id),
							},
						},
					},
				},
			},
		)
		m.setPendingJoins(chat.Id, user.Id)

		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.ContinueGroups
}

// joinRequestHandler handles callback queries for join requests.
//
// Admins can approve, decline, or ban users requesting to join the chat.
func (moduleStruct) joinRequestHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From
	chat := ctx.EffectiveChat
	msg := query.Message

	// permission checks
	if !chat_status.RequireUserAdmin(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	response := args[1]
	joinUserId, _ := strconv.ParseInt(args[2], 10, 64)
	joinUser, err := b.GetChat(joinUserId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	var helpText string

	switch response {
	case "accept":
		_, _ = b.ApproveChatJoinRequest(chat.Id, joinUser.Id, nil)
		helpText = "Accepted %s in Chat ✅"
		_ = cache.Marshal.Delete(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chat.Id, joinUser.Id))
	case "decline":
		_, _ = b.DeclineChatJoinRequest(chat.Id, joinUser.Id, nil)
		helpText = "Declined %s to join chat ❌"
	case "ban":
		_, _ = chat.BanMember(b, joinUser.Id, nil)
		_, _ = b.DeclineChatJoinRequest(chat.Id, joinUser.Id, nil)
		helpText = "✅ Successfully Banned! %s"
	}

	_, _, err = msg.EditText(b,
		fmt.Sprintf(helpText, helpers.MentionHtml(joinUser.Id, joinUser.FirstName)),
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b,
		&gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf(helpText, joinUser.FirstName),
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// autoApprove toggles or displays the setting for auto-approving join requests.
//
// Admins can enable or disable automatic approval of new join requests.
func (moduleStruct) autoApprove(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]
	var err error
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		delPref := db.GetGreetingSettings(chat.Id).ShouldAutoApprove
		if delPref {
			_, err = msg.Reply(bot, "I'm auto-approving new chat join requests now.", helpers.Smarkdown())
		} else {
			_, err = msg.Reply(bot, "I'm not auto-approving new chat join requests now..", helpers.Shtml())
		}
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		db.SetShouldAutoApprove(chat.Id, false)
		_, err = msg.Reply(bot, "I won't auto-approve new join requests!", helpers.Shtml())
	case "on", "yes":
		db.SetShouldAutoApprove(chat.Id, true)
		_, err = msg.Reply(bot, "I'll try to auto-approve new join requests!", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// loadPendingJoins checks if a join request for a user in a chat is already pending.
//
// Returns true if the request is pending, otherwise false.
func (moduleStruct) loadPendingJoins(chatId, userId int64) bool {
	alreadyAsked, _ := cache.Marshal.Get(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chatId, userId), new(bool))
	if alreadyAsked == nil || !alreadyAsked.(bool) {
		return false
	}
	return true
}

// setPendingJoins marks a join request as pending for a user in a chat.
//
// Stores the pending state with a 5-minute expiration.
func (moduleStruct) setPendingJoins(chatId, userId int64) {
	_ = cache.Marshal.Set(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chatId, userId), true, store.WithExpiration(5*time.Minute))
}

// LoadGreetings registers all greeting-related command handlers with the dispatcher.
//
// This function enables the greetings module and adds handlers for welcome/goodbye
// messages, member join/leave events, and join request management. The module
// provides customizable greeting systems with rich formatting and media support.
//
// Registered commands:
//   - /welcome: Displays current welcome message settings
//   - /setwelcome: Sets custom welcome message for new members
//   - /resetwelcome: Resets welcome message to default
//   - /goodbye: Displays current goodbye message settings
//   - /setgoodbye: Sets custom goodbye message for leaving members
//   - /resetgoodbye: Resets goodbye message to default
//   - /cleanwelcome: Toggles automatic welcome message cleanup
//   - /cleangoodbye: Toggles automatic goodbye message cleanup
//   - /cleanservice: Toggles deletion of service messages
//   - /autoapprove: Toggles automatic approval of join requests
//
// Event handlers:
//   - Join requests: Handles approval/rejection with optional manual review
//   - Member joins: Sends welcome messages to new members
//   - Member leaves: Sends goodbye messages for departing members
//   - Service messages: Automatically cleans join/leave service messages
//
// Features:
//   - Rich message formatting with variables and buttons
//   - Media support for welcome/goodbye messages
//   - Automatic message cleanup with configurable timing
//   - Join request management with manual approval
//   - Service message deletion for cleaner chat experience
//   - Pending join tracking with expiration
//
// Requirements:
//   - Bot must be admin to delete service messages
//   - User must be admin to configure greeting settings
//   - Module supports remote configuration via connections
//   - Integrates with formatting module for rich messages
//
// The greetings system enhances user experience with personalized messages
// and automated chat management for member transitions.
func LoadGreetings(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(greetingsModule.moduleName, true)

	// Adds Formatting kb button to Greetings Menu
	HelpModule.helpableKb[greetingsModule.moduleName] = [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         "Formatting",
				CallbackData: fmt.Sprintf("helpq.%s", "Formatting"),
			},
		},
	}

	// this is used when user join, and creates a join request
	dispatcher.AddHandler(
		handlers.NewChatJoinRequest(
			chatjoinrequest.All, greetingsModule.pendingJoins,
		),
	)

	// this is for chat member joined the chat
	dispatcher.AddHandler(
		handlers.NewChatMember(
			func(u *gotgbot.ChatMemberUpdated) bool {
				wasMember, isMember := helpers.ExtractJoinLeftStatusChange(u)
				return !wasMember && isMember
			},
			greetingsModule.newMember,
		),
	)

	// this is for chat member left the chat
	dispatcher.AddHandler(
		handlers.NewChatMember(
			func(u *gotgbot.ChatMemberUpdated) bool {
				wasMember, isMember := helpers.ExtractJoinLeftStatusChange(u)
				return wasMember && !isMember
			},
			greetingsModule.leftMember,
		),
	)

	// for cleaning service messages
	dispatcher.AddHandler(
		handlers.NewMessage(
			func(msg *gotgbot.Message) bool {
				return msg.LeftChatMember != nil || msg.NewChatMembers != nil
			},
			greetingsModule.cleanService,
		),
	)

	dispatcher.AddHandler(handlers.NewCommand("welcome", greetingsModule.welcome))
	dispatcher.AddHandler(handlers.NewCommand("setwelcome", greetingsModule.setWelcome))
	dispatcher.AddHandler(handlers.NewCommand("resetwelcome", greetingsModule.resetWelcome))
	dispatcher.AddHandler(handlers.NewCommand("goodbye", greetingsModule.goodbye))
	dispatcher.AddHandler(handlers.NewCommand("setgoodbye", greetingsModule.setGoodbye))
	dispatcher.AddHandler(handlers.NewCommand("resetgoodbye", greetingsModule.resetGoodbye))
	dispatcher.AddHandler(handlers.NewCommand("cleanwelcome", greetingsModule.cleanWelcome))
	dispatcher.AddHandler(handlers.NewCommand("cleangoodbye", greetingsModule.cleanGoodbye))
	dispatcher.AddHandler(handlers.NewCommand("cleanservice", greetingsModule.delJoined))
	dispatcher.AddHandler(handlers.NewCommand("autoapprove", greetingsModule.autoApprove))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("join_request."), greetingsModule.joinRequestHandler))
}
