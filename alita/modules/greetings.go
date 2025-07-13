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

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

/*
greetingsModule provides logic for managing welcome and goodbye messages in group chats.

Implements commands to set, reset, and configure greetings, as well as handlers for join/leave events and join requests.
*/
var greetingsModule = moduleStruct{
	moduleName: autoModuleName(),
	cfg:        nil, // will be set during LoadGreetings
}

// getGreetingMsg is a helper function to safely get greeting messages with fallback
func getGreetingMsg(tr *i18n.I18n, key, fallback string) string {
	text, err := tr.GetStringWithError(key)
	if err != nil {
		log.Error(err)
		return fallback
	}
	return text
}

/*
welcome displays or toggles the welcome message settings for the chat.

Admins can view current settings, toggle welcoming on/off, or display the current welcome message.
*/
func (m moduleStruct) welcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
		welcomeStatusMsg, welcomeStatusErr := tr.GetStringWithError("strings.greetings.welcome.status")
		if welcomeStatusErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", welcomeStatusErr)
			welcomeStatusMsg = "I am currently welcoming users: `%s`"
		}
		_, err := msg.Reply(bot, fmt.Sprintf(welcomeStatusMsg,
			welcPrefs.WelcomeSettings.WelcomeText), helpers.Smarkdown())
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
			welcomeEnabledMsg, welcomeEnabledErr := tr.GetStringWithError("strings.greetings.welcome.enabled")
			if welcomeEnabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", welcomeEnabledErr)
				welcomeEnabledMsg = "I'll now welcome users!"
			}
			_, err = msg.Reply(bot, welcomeEnabledMsg, helpers.Shtml())
		case "off", "no":
			db.SetWelcomeToggle(chat.Id, false)
			welcomeDisabledMsg, welcomeDisabledErr := tr.GetStringWithError("strings.greetings.welcome.disabled")
			if welcomeDisabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", welcomeDisabledErr)
				welcomeDisabledMsg = "I'll no longer welcome users."
			}
			_, err = msg.Reply(bot, welcomeDisabledMsg, helpers.Shtml())
		default:
			invalidOptionMsg, invalidErr := tr.GetStringWithError("strings.commonstrings.errors.invalid_option_yes_no")
			if invalidErr != nil {
				log.Error(invalidErr)
				invalidOptionMsg = "Invalid option. Please use yes/no or on/off"
			}
			_, err = msg.Reply(bot, invalidOptionMsg, helpers.Shtml())
		}

		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

/*
setWelcome sets a custom welcome message for the chat.

Admins can set the message content, type, and buttons. Handles input validation and replies with the result.
*/
func (m moduleStruct) setWelcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
	setWelcomeSuccessMsg, setWelcomeSuccessErr := tr.GetStringWithError("strings.greetings.set_welcome.success")
	if setWelcomeSuccessErr != nil {
		log.Errorf("[greetings] missing translation for key: %v", setWelcomeSuccessErr)
		setWelcomeSuccessMsg = "I'll now welcome users with that message!"
	}
	_, err := msg.Reply(bot, setWelcomeSuccessMsg, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
resetWelcome resets the welcome message to the default.

Admins can use this to remove any custom welcome message.
*/
func (m moduleStruct) resetWelcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
	resetWelcomeSuccessMsg, resetWelcomeSuccessErr := tr.GetStringWithError("strings.greetings.reset_welcome.success")
	if resetWelcomeSuccessErr != nil {
		log.Errorf("[greetings] missing translation for key: %v", resetWelcomeSuccessErr)
		resetWelcomeSuccessMsg = "I have reset the welcome message back to default!"
	}
	_, err := msg.Reply(bot, resetWelcomeSuccessMsg, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
goodbye displays or toggles the goodbye message settings for the chat.

Admins can view current settings, toggle goodbyes on/off, or display the current goodbye message.
*/
func (moduleStruct) goodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
		goodbyeStatusMsg, goodbyeStatusErr := tr.GetStringWithError("strings.greetings.i_am_currently_goodbying_users_t")
		if goodbyeStatusErr != nil {
			log.Errorf("[greetings] missing translation for i_am_currently_goodbying_users_t: %v", goodbyeStatusErr)
			goodbyeStatusMsg = "I am currently goodbying users: <code>%t</code>"
		}
		_, err := msg.Reply(bot, fmt.Sprintf(goodbyeStatusMsg+
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
			_, err = msg.Reply(bot, getGreetingMsg(tr, "strings.greetings.goodbye.enabled", "Goodbye messages enabled"), helpers.Shtml())
		case "off", "no":
			db.SetGoodbyeToggle(chat.Id, false)
			_, err = msg.Reply(bot, getGreetingMsg(tr, "strings.greetings.goodbye.disabled", "Goodbye messages disabled"), helpers.Shtml())
		default:
			_, err = msg.Reply(bot, getGreetingMsg(tr, "strings.greetings.i_understand_on_yes_or_off_no_only", "I understand only on/yes or off/no"), helpers.Shtml())
		}
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

/*
setGoodbye sets a custom goodbye message for the chat.

Admins can set the message content, type, and buttons. Handles input validation and replies with the result.
*/
func (moduleStruct) setGoodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
	setGoodbyeSuccessMsg, setGoodbyeSuccessErr := tr.GetStringWithError("strings.greetings.set_goodbye.success")
	if setGoodbyeSuccessErr != nil {
		log.Errorf("[greetings] missing translation for key: %v", setGoodbyeSuccessErr)
		setGoodbyeSuccessMsg = "Goodbye message set successfully."
	}
	_, err := msg.Reply(bot, setGoodbyeSuccessMsg, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
resetGoodbye resets the goodbye message to the default.

Admins can use this to remove any custom goodbye message.
*/
func (moduleStruct) resetGoodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
	resetGoodbyeSuccessMsg, resetGoodbyeSuccessErr := tr.GetStringWithError("strings.greetings.reset_goodbye.success")
	if resetGoodbyeSuccessErr != nil {
		log.Errorf("[greetings] missing translation for key: %v", resetGoodbyeSuccessErr)
		resetGoodbyeSuccessMsg = "Goodbye message reset successfully."
	}
	_, err := msg.Reply(bot, resetGoodbyeSuccessMsg, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
cleanWelcome toggles or displays the setting for deleting old welcome messages.

Admins can enable or disable automatic deletion of old welcome messages.
*/
func (moduleStruct) cleanWelcome(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
			statusEnabledMsg, statusEnabledErr := tr.GetStringWithError("strings.greetings.clean_welcome.status_enabled")
			if statusEnabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", statusEnabledErr)
				statusEnabledMsg = "Clean welcome is currently enabled."
			}
			_, err = msg.Reply(bot, statusEnabledMsg, helpers.Shtml())
		} else {
			statusDisabledMsg, statusDisabledErr := tr.GetStringWithError("strings.greetings.clean_welcome.status_disabled")
			if statusDisabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", statusDisabledErr)
				statusDisabledMsg = "Clean welcome is currently disabled."
			}
			_, err = msg.Reply(bot, statusDisabledMsg, helpers.Shtml())
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
		disabledMsg, disabledErr := tr.GetStringWithError("strings.greetings.clean_welcome.disabled")
		if disabledErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", disabledErr)
			disabledMsg = "Clean welcome has been disabled."
		}
		_, err = msg.Reply(bot, disabledMsg, helpers.Shtml())
	case "on", "yes":
		db.SetCleanWelcomeSetting(chat.Id, true)
		enabledMsg, enabledErr := tr.GetStringWithError("strings.greetings.clean_welcome.enabled")
		if enabledErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", enabledErr)
			enabledMsg = "Clean welcome has been enabled."
		}
		_, err = msg.Reply(bot, enabledMsg, helpers.Shtml())
	default:
		invalidOptionMsg, invalidOptionErr := tr.GetStringWithError("strings.greetings.i_understand_on_yes_or_off_no_only")
		if invalidOptionErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", invalidOptionErr)
			invalidOptionMsg = "I understand only on/yes or off/no."
		}
		_, err = msg.Reply(bot, invalidOptionMsg, helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
cleanGoodbye toggles or displays the setting for deleting old goodbye messages.

Admins can enable or disable automatic deletion of old goodbye messages.
*/
func (moduleStruct) cleanGoodbye(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
			statusEnabledMsg, statusEnabledErr := tr.GetStringWithError("strings.greetings.clean_goodbye.status_enabled")
			if statusEnabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", statusEnabledErr)
				statusEnabledMsg = "Clean goodbye is currently enabled."
			}
			_, err = msg.Reply(bot, statusEnabledMsg, helpers.Shtml())
		} else {
			statusDisabledMsg, statusDisabledErr := tr.GetStringWithError("strings.greetings.clean_goodbye.status_disabled")
			if statusDisabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", statusDisabledErr)
				statusDisabledMsg = "Clean goodbye is currently disabled."
			}
			_, err = msg.Reply(bot, statusDisabledMsg, helpers.Shtml())
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
		disabledMsg, disabledErr := tr.GetStringWithError("strings.greetings.clean_goodbye.disabled")
		if disabledErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", disabledErr)
			disabledMsg = "Clean goodbye has been disabled."
		}
		_, err = msg.Reply(bot, disabledMsg, helpers.Shtml())
	case "on", "yes":
		db.SetCleanGoodbyeSetting(chat.Id, true)
		enabledMsg, enabledErr := tr.GetStringWithError("strings.greetings.clean_goodbye.enabled")
		if enabledErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", enabledErr)
			enabledMsg = "Clean goodbye has been enabled."
		}
		_, err = msg.Reply(bot, enabledMsg, helpers.Shtml())
	default:
		invalidOptionMsg, invalidOptionErr := tr.GetStringWithError("strings.greetings.i_understand_on_yes_or_off_no_only")
		if invalidOptionErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", invalidOptionErr)
			invalidOptionMsg = "I understand only on/yes or off/no."
		}
		_, err = msg.Reply(bot, invalidOptionMsg, helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
delJoined toggles or displays the setting for deleting "user joined" service messages.

Admins can enable or disable automatic deletion of join messages.
*/
func (moduleStruct) delJoined(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
			statusEnabledMsg, statusEnabledErr := tr.GetStringWithError("strings.greetings.del_joined.status_enabled")
			if statusEnabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", statusEnabledErr)
				statusEnabledMsg = "Delete joined messages is currently enabled."
			}
			_, err = msg.Reply(bot, statusEnabledMsg, helpers.Smarkdown())
		} else {
			statusDisabledMsg, statusDisabledErr := tr.GetStringWithError("strings.greetings.del_joined.status_disabled")
			if statusDisabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", statusDisabledErr)
				statusDisabledMsg = "Delete joined messages is currently disabled."
			}
			_, err = msg.Reply(bot, statusDisabledMsg, helpers.Shtml())
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
		disabledMsg, disabledErr := tr.GetStringWithError("strings.greetings.del_joined.disabled")
		if disabledErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", disabledErr)
			disabledMsg = "Delete joined messages has been disabled."
		}
		_, err = msg.Reply(bot, disabledMsg, helpers.Shtml())
	case "on", "yes":
		db.SetShouldCleanService(chat.Id, true)
		enabledMsg, enabledErr := tr.GetStringWithError("strings.greetings.del_joined.enabled")
		if enabledErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", enabledErr)
			enabledMsg = "Delete joined messages has been enabled."
		}
		_, err = msg.Reply(bot, enabledMsg, helpers.Shtml())
	default:
		invalidOptionMsg, invalidOptionErr := tr.GetStringWithError("strings.greetings.i_understand_on_yes_or_off_no_only")
		if invalidOptionErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", invalidOptionErr)
			invalidOptionMsg = "I understand only on/yes or off/no."
		}
		_, err = msg.Reply(bot, invalidOptionMsg, helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
newMember handles the event when a new member joins the chat.

Sends a welcome message if enabled and manages deletion of previous welcome messages if configured.
*/
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

/*
leftMember handles the event when a member leaves the chat.

Sends a goodbye message if enabled and manages deletion of previous goodbye messages if configured.
*/
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

/*
cleanService deletes service messages if the setting is enabled.

Used for cleaning up join/leave notifications and other service messages.
*/
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

/*
pendingJoins handles new chat join requests.

If auto-approve is enabled, approves the request. Otherwise, notifies admins and tracks pending requests.
*/
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
				ParseMode: gotgbot.ParseModeHTML,
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

/*
joinRequestHandler handles callback queries for join requests.

Admins can approve, decline, or ban users requesting to join the chat.
*/
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
			ParseMode: gotgbot.ParseModeHTML,
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

/*
autoApprove toggles or displays the setting for auto-approving join requests.

Admins can enable or disable automatic approval of new join requests.
*/
func (moduleStruct) autoApprove(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
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
			enabledMsg, enabledErr := tr.GetStringWithError("strings.greetings.auto_approve.enabled")
			if enabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", enabledErr)
				enabledMsg = "Auto approve is currently enabled."
			}
			_, err = msg.Reply(bot, enabledMsg, helpers.Smarkdown())
		} else {
			disabledMsg, disabledErr := tr.GetStringWithError("strings.greetings.auto_approve.disabled")
			if disabledErr != nil {
				log.Errorf("[greetings] missing translation for key: %v", disabledErr)
				disabledMsg = "Auto approve is currently disabled."
			}
			_, err = msg.Reply(bot, disabledMsg, helpers.Shtml())
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
		disabledNowMsg, disabledNowErr := tr.GetStringWithError("strings.greetings.auto_approve.disabled_now")
		if disabledNowErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", disabledNowErr)
			disabledNowMsg = "Auto approve has been disabled."
		}
		_, err = msg.Reply(bot, disabledNowMsg, helpers.Shtml())
	case "on", "yes":
		db.SetShouldAutoApprove(chat.Id, true)
		enabledNowMsg, enabledNowErr := tr.GetStringWithError("strings.greetings.auto_approve.enabled_now")
		if enabledNowErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", enabledNowErr)
			enabledNowMsg = "Auto approve has been enabled."
		}
		_, err = msg.Reply(bot, enabledNowMsg, helpers.Shtml())
	default:
		invalidOptionMsg, invalidOptionErr := tr.GetStringWithError("strings.greetings.i_understand_on_yes_or_off_no_only")
		if invalidOptionErr != nil {
			log.Errorf("[greetings] missing translation for key: %v", invalidOptionErr)
			invalidOptionMsg = "I understand only on/yes or off/no."
		}
		_, err = msg.Reply(bot, invalidOptionMsg, helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
loadPendingJoins checks if a join request for a user in a chat is already pending.

Returns true if the request is pending, otherwise false.
*/
func (moduleStruct) loadPendingJoins(chatId, userId int64) bool {
	alreadyAsked, _ := cache.Marshal.Get(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chatId, userId), new(bool))
	if alreadyAsked == nil || !alreadyAsked.(bool) {
		return false
	}
	return true
}

/*
setPendingJoins marks a join request as pending for a user in a chat.

Stores the pending state with a 5-minute expiration.
*/
func (moduleStruct) setPendingJoins(chatId, userId int64) {
	_ = cache.Marshal.Set(cache.Context, fmt.Sprintf("pendingJoins.%d.%d", chatId, userId), true, store.WithExpiration(5*time.Minute))
}

/*
LoadGreetings registers all greeting-related command handlers with the dispatcher.

Enables the greetings module and adds handlers for welcome/goodbye messages, join/leave events, and join requests.
*/
func LoadGreetings(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	greetingsModule.cfg = cfg

	HelpModule.AbleMap.Store(greetingsModule.moduleName, true)

	// Adds Formatting kb button to Greetings Menu
	HelpModule.helpableKb[greetingsModule.moduleName] = [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         "Formatting", // Note: tr is not available in LoadGreetings, using fallback
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
