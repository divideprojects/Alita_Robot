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

var greetingsModule = moduleStruct{moduleName: "Greetings"}

func (moduleStruct) welcome(bot *gotgbot.Bot, ctx *ext.Context) error {
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
			_, err := helpers.GreetingsEnumFuncMap[welcPrefs.WelcomeSettings.WelcomeType](bot, ctx, wlcmText, welcPrefs.WelcomeSettings.FileID, &gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil})
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			wlcmText, buttons = helpers.FormattingReplacer(bot, chat, user, wlcmText, buttons)
			keyb := helpers.BuildKeyboard(buttons)
			keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}
			_, err := helpers.GreetingsEnumFuncMap[welcPrefs.WelcomeSettings.WelcomeType](bot, ctx, wlcmText, welcPrefs.WelcomeSettings.FileID, &keyboard)
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

func (moduleStruct) joinRequestHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
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

func (moduleStruct) loadPendingJoins(chatId, userId int64) bool {
	alreadyAsked, _ := cache.Marshal.Get(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chatId, userId), new(bool))
	if alreadyAsked == nil || !alreadyAsked.(bool) {
		return false
	}
	return true
}

func (moduleStruct) setPendingJoins(chatId, userId int64) {
	_ = cache.Marshal.Set(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chatId, userId), true, store.WithExpiration(5*time.Minute))
}

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
