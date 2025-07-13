package modules

import (
	"fmt"
	"html"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

/*
HelpModule provides logic for the bot's help and about system.

Implements commands and handlers for help menus, about info, configuration, and donation instructions.
*/
var HelpModule = moduleStruct{
	moduleName:     autoModuleName(),
	AbleMap:        moduleEnabled{},
	AltHelpOptions: make(map[string][]string),
	helpableKb:     make(map[string][][]gotgbot.InlineKeyboardButton),
}

const (
	aboutText = "Info & About\n\nHere are some of the FAQs about Alita."
	startHelp = "Hey there! My name is Alita ✨.\n" +
		"I'm here to help you manage your groups!\n" +
		"Hit /help to find out more about how to use me to my full potential.\n" +
		"Join my <a href='https://t.me/AlitaRobotUpdates'>News Channel</a> to get information on all the latest updates."
	mainhlp = "Hey %s!\n" +
		"My name is Alita ✨.\n\n" +
		"I am a group management bot, here to help you get around and keep the order in your groups!\n" +
		"I have lots of handy features, such as flood control, a warning system, a note keeping system, " +
		"and even predetermined replies on certain keywords.\n\n" +
		"<b>Helpful commands</b>:\n" +
		" - /start: Starts me! You've probably already used this!\n" +
		" - /help: Sends this message; I'll tell you more about myself!\n" +
		" - /donate: Gives you info on how to support me and my creator.\n\n" +
		"All commands can be used with the following: / or !"
)

// getBackBtnSuffix returns the back button suffix with proper translations
func getBackBtnSuffix(tr *i18n.I18n) []gotgbot.InlineKeyboardButton {
	homeText, homeErr := tr.GetStringWithError("strings.CommonStrings.buttons.home")
	if homeErr != nil {
		log.Errorf("[help] missing translation for buttons.home: %v", homeErr)
		homeText = "Home"
	}
	return []gotgbot.InlineKeyboardButton{
		{
			Text:         "« Back",
			CallbackData: "helpq.Help",
		},
		{
			Text:         homeText,
			CallbackData: "help.home",
		},
	}
}

// getAboutKb returns the about keyboard with proper translations
func getAboutKb(tr *i18n.I18n) gotgbot.InlineKeyboardMarkup {
	aboutMeText, aboutMeErr := tr.GetStringWithError("strings.Help.about_me")
	if aboutMeErr != nil {
		log.Errorf("[help] missing translation for about_me: %v", aboutMeErr)
		aboutMeText = "About Me"
	}

	newsChannelText, newsChannelErr := tr.GetStringWithError("strings.Help.about.news_channel_button")
	if newsChannelErr != nil {
		log.Errorf("[help] missing translation for about.news_channel_button: %v", newsChannelErr)
		newsChannelText = "News Channel"
	}

	supportGroupText, supportGroupErr := tr.GetStringWithError("strings.Help.start.support_group_button")
	if supportGroupErr != nil {
		log.Errorf("[help] missing translation for start.support_group_button: %v", supportGroupErr)
		supportGroupText = "Support Group"
	}

	configurationText, configurationErr := tr.GetStringWithError("strings.Help.about.configuration_button")
	if configurationErr != nil {
		log.Errorf("[help] missing translation for about.configuration_button: %v", configurationErr)
		configurationText = "Configuration"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         aboutMeText,
					CallbackData: "about.me",
				},
			},
			{
				{
					Text: newsChannelText,
					Url:  "https://t.me/DivideProjects",
				},
				{
					Text: supportGroupText,
					Url:  "https://t.me/DivideSupport",
				},
			},
			{
				{
					Text:         configurationText,
					CallbackData: "help.config",
				},
			},
			{
				// custom back button
				{
					Text:         "⬅ Back",
					CallbackData: "helpq.BackStart",
				},
			},
		},
	}
}

// getStartMarkup returns the start markup with proper translations
func getStartMarkup(tr *i18n.I18n) gotgbot.InlineKeyboardMarkup {
	aboutButtonText, aboutButtonErr := tr.GetStringWithError("strings.Help.start.about_button")
	if aboutButtonErr != nil {
		log.Errorf("[help] missing translation for start.about_button: %v", aboutButtonErr)
		aboutButtonText = "About"
	}

	supportGroupText, supportGroupErr := tr.GetStringWithError("strings.Help.start.support_group_button")
	if supportGroupErr != nil {
		log.Errorf("[help] missing translation for start.support_group_button: %v", supportGroupErr)
		supportGroupText = "Support Group"
	}

	languageButtonText, languageButtonErr := tr.GetStringWithError("strings.Help.start.language_button")
	if languageButtonErr != nil {
		log.Errorf("[help] missing translation for start.language_button: %v", languageButtonErr)
		languageButtonText = "Language"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         aboutButtonText,
					CallbackData: "help.about",
				},
			},
			{
				{
					Text: "➕ Add me to chat!",
					Url:  "https://t.me/Alita_Robot?startgroup=botstart",
				},
				{
					Text: supportGroupText,
					Url:  "https://t.me/DivideSupport",
				},
			},
			{
				{
					Text:         "📚 Commands & Help",
					CallbackData: "helpq.Help",
				},
			},
			{
				{
					Text:         languageButtonText,
					CallbackData: "chlang.start",
				},
			},
		},
	}
}

/*
moduleEnabled tracks which modules are enabled for help and configuration.

Used internally by the help system.
*/
type moduleEnabled struct {
	modules map[string]bool
}

/*
Init initializes the moduleEnabled map.
*/
func (m *moduleEnabled) Init() {
	m.modules = make(map[string]bool)
}

/*
Store sets the enabled state for a module.
*/
func (m *moduleEnabled) Store(module string, enabled bool) {
	m.modules[module] = enabled
}

/*
Load retrieves the enabled state for a module.

Returns the module name and whether it is enabled.
*/
func (m *moduleEnabled) Load(module string) (string, bool) {
	log.Info(fmt.Sprintf("[Module] Loading %s module", module))
	return module, m.modules[module]
}

/*
LoadModules returns a list of all enabled module names.
*/
func (m *moduleEnabled) LoadModules() []string {
	modules := make([]string, 0)
	for module := range m.modules {
		moduleName, enabled := m.Load(module)
		if enabled {
			modules = append(modules, moduleName)
		}
	}
	return modules
}

// helpModule holds the configuration for the help module
var helpModule = moduleStruct{
	moduleName: autoModuleName(),
	cfg:        nil, // will be set during LoadHelp
}

// getHelpButtonText is a helper function to safely get button text with fallback
func getHelpButtonText(tr *i18n.I18n, key, fallback string) string {
	text, err := tr.GetStringWithError(key)
	if err != nil {
		log.Error(err)
		return fallback
	}
	return text
}

/*
about displays information about the bot, including FAQs and about text.

Handles both command and callback query contexts.
*/
func (m moduleStruct) about(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
	msg := ctx.EffectiveMessage
	var (
		currText string
		currKb   gotgbot.InlineKeyboardMarkup
	)

	aboutText, aboutErr := tr.GetStringWithError("strings.Help.about_text")
	if aboutErr != nil {
		log.Errorf("[help] missing translation for about_text: %v", aboutErr)
		aboutText = "I'm Alita, a group management bot built to help you manage your groups effectively!"
	}
	aboutKb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         getHelpButtonText(tr, "strings.Help.about_me", "About Me"),
					CallbackData: "about.me",
				},
			},
		},
	}

	if query := ctx.Update.CallbackQuery; query != nil {
		response := strings.Split(query.Data, ".")[1]

		switch response {
		case "main":
			currText = aboutText
			currKb = aboutKb
		case "me":
			cfg := m.cfg
			aboutMeText, aboutMeErr := tr.GetStringWithError("strings.Help.About")
			if aboutMeErr != nil {
				log.Errorf("[help] missing translation for Help.About: %v", aboutMeErr)
				aboutMeText = "I'm @%s, version %s. I'm here to help you manage your groups!"
			}
			currText = fmt.Sprintf(aboutMeText, b.Username, cfg.BotVersion)
			currKb = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         getHelpButtonText(tr, "strings.CommonStrings.buttons.back", "Back"),
							CallbackData: "about.main",
						},
					},
				},
			}
		}
		_, _, err := query.Message.EditText(b,
			currText,
			&gotgbot.EditMessageTextOpts{
				ReplyMarkup: currKb,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ParseMode: gotgbot.ParseModeHTML,
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
	} else {
		if ctx.Update.Message.Chat.Type == "private" {
			currText = aboutText
			currKb = aboutKb
		} else {
			aboutButtonText, aboutButtonErr := tr.GetStringWithError("strings.Help.about.button")
			if aboutButtonErr != nil {
				log.Errorf("[help] missing translation for about.button: %v", aboutButtonErr)
				aboutButtonText = "About"
			}
			currText = "Click on the button below to get info about me!"
			currKb = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text: aboutButtonText,
							Url:  fmt.Sprintf("https://t.me/%s?start=about", b.Username),
						},
					},
				},
			}
		}
		_, err := msg.Reply(
			b,
			currText,
			&gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyMarkup: &currKb,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

/*
helpButtonHandler handles callback queries for the help menu.

Displays help text and navigation for modules and main help options.
*/
func (moduleStruct) helpButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	args := strings.Split(query.Data, ".")
	module := args[1]

	var (
		parsemode, helpText string
		replyKb             gotgbot.InlineKeyboardMarkup
	)

	// Sort the module names
	if string_handling.FindInStringSlice([]string{"BackStart", "Help"}, module) {
		parsemode = gotgbot.ParseModeHTML
		switch module {
		case "Help":
			// This shows the main start menu
			helpText = fmt.Sprintf(mainhlp, html.EscapeString(query.From.FirstName))
			replyKb = markup
		case "BackStart":
			// This shows the modules menu
			tr := i18n.New(db.GetLanguage(ctx))
			helpText = startHelp
			replyKb = getStartMarkup(tr)
		}
	} else {
		// For all remainging modules
		// FIXME: error for pins, purges, reports, rules, warns
		helpText, replyKb, parsemode = getHelpTextAndMarkup(ctx, strings.ToLower(module))
	}

	// Edit the main message, the main querymessage
	_, _, err := query.Message.EditText(
		b,
		helpText,
		&gotgbot.EditMessageTextOpts{
			ParseMode:   parsemode,
			ReplyMarkup: replyKb,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
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

// start introduces the bot
/*
start introduces the bot and handles /start commands.

Displays the main help menu or processes special start arguments for help, connection, rules, or notes.
*/
func (moduleStruct) start(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	args := ctx.Args()

	if ctx.Update.Message.Chat.Type == "private" {
		if len(args) == 1 {
			tr := i18n.New(db.GetLanguage(ctx))
			startMarkup := getStartMarkup(tr)
			_, err := msg.Reply(b,
				startHelp,
				&gotgbot.SendMessageOpts{
					ParseMode: gotgbot.ParseModeHTML,
					LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
						IsDisabled: true,
					},
					ReplyMarkup: &startMarkup,
				},
			)
			if err != nil {
				log.Error(err)
				return err
			}
		} else if len(args) == 2 {
			err := startHelpPrefixHandler(b, ctx, user, args[1])
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			log.Info("sed")
		}
	} else {
		tr := i18n.New(db.GetLanguage(ctx))
		groupPromptMsg, groupPromptErr := tr.GetStringWithError("strings.Help.start.group_prompt")
		if groupPromptErr != nil {
			log.Errorf("[help] missing translation for start.group_prompt: %v", groupPromptErr)
			groupPromptMsg = tr.GetString("strings.Help.start.group_prompt")
		}
		_, err := msg.Reply(b, groupPromptMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

/*
donate displays information on how to support the bot and its creator.
*/
func (moduleStruct) donate(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	tr := i18n.New("en")
	donateText, donateErr := tr.GetStringWithError("strings.Help.DonateText")
	if donateErr != nil {
		log.Errorf("[help] missing translation for DonateText: %v", donateErr)
		donateText = "Support the development of this bot by donating!"
	}
	_, err := b.SendMessage(chat.Id,
		donateText,
		&gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                msg.MessageId,
				AllowSendingWithoutReply: true,
			},
		},
	)
	if err != nil {
		log.Error(err)
	}

	return ext.EndGroups
}

/*
botConfig handles the interactive configuration menu for the bot.

Only works in private chat. Guides users through configuration steps.
*/
func (moduleStruct) botConfig(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	msg := query.Message
	tr := i18n.New(db.GetLanguage(ctx))

	// just in case
	if msg.GetChat().Type != "private" {
		privateOnlyMsg, privateOnlyErr := tr.GetStringWithError("strings.Help.configuration.private_only")
		if privateOnlyErr != nil {
			log.Errorf("[help] missing translation for configuration.private_only: %v", privateOnlyErr)
			privateOnlyMsg = "This command can only be used in private chat."
		}
		_, _, err := msg.EditText(b, privateOnlyMsg, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	args := strings.Split(query.Data, ".")
	response := args[1]

	var (
		iKeyboard [][]gotgbot.InlineKeyboardButton
		text      string
	)

	tr = i18n.New("en")

	switch response {
	case "step1":
		iKeyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text: "➕ Add Alita to chat!",
					Url:  fmt.Sprintf("https://t.me/%s?startgroup=botstart", b.Username),
				},
			},
			{
				{
					Text:         getHelpButtonText(tr, "strings.CommonStrings.buttons.done", "Done"),
					CallbackData: "configuration.step2",
				},
			},
		}
		step1Text, step1Err := tr.GetStringWithError("strings.Help.Configuration.Step-1")
		if step1Err != nil {
			log.Errorf("[help] missing translation for Configuration.Step-1: %v", step1Err)
			step1Text = "First, add me to your group!"
		}
		text = step1Text
	case "step2":
		iKeyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         getHelpButtonText(tr, "strings.CommonStrings.buttons.done", "Done"),
					CallbackData: "configuration.step3",
				},
			},
		}
		step2Text, step2Err := tr.GetStringWithError("strings.Help.Configuration.Step-2")
		if step2Err != nil {
			log.Errorf("[help] missing translation for Configuration.Step-2: %v", step2Err)
			step2Text = "Now, make me an admin in your group by typing /promote @%s"
		}
		text = fmt.Sprintf(step2Text, b.Username)
	case "step3":
		iKeyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "✅ Continue ✅",
					CallbackData: "helpq.Help",
				},
			},
		}
		step3Text, step3Err := tr.GetStringWithError("strings.Help.Configuration.Step-3")
		if step3Err != nil {
			log.Errorf("[help] missing translation for Configuration.Step-3: %v", step3Err)
			step3Text = "Great! Now you can start using me. Type /help to see all available commands."
		}
		text = step3Text
	}
	_, _, err := msg.EditText(
		b,
		text,
		&gotgbot.EditMessageTextOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: iKeyboard,
			},
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
help displays the help menu or module-specific help.

Handles both private and group chat contexts, and provides navigation to module help or PM help links.
*/
func (moduleStruct) help(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	args := ctx.Args()
	tr := i18n.New(db.GetLanguage(ctx))
	var err error

	if ctx.Update.Message.Chat.Type == "private" {
		if len(args) == 1 {
			_, err = b.SendMessage(chat.Id,
				fmt.Sprintf(
					mainhlp,
					html.EscapeString(msg.From.FirstName),
				),
				&gotgbot.SendMessageOpts{
					ParseMode:   gotgbot.ParseModeHTML,
					ReplyMarkup: &markup,
				},
			)
			if err != nil {
				log.Error(err)
				return err
			}
		} else if len(args) == 2 {
			module := strings.ToLower(args[1])
			helpText, replyMarkup, _parsemode := getHelpTextAndMarkup(ctx, module)
			_, err = b.SendMessage(
				chat.Id,
				helpText,
				&gotgbot.SendMessageOpts{
					ParseMode:   _parsemode,
					ReplyMarkup: &replyMarkup,
				},
			)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		var pmMeKbText string
		pmMeKbText, err = tr.GetStringWithError("strings.Help.help.button")
		if err != nil {
			log.Error(err)
			return err
		}
		pmMeKbUri := fmt.Sprintf("https://t.me/%s?start=help_help", b.Username)
		moduleHelpString := tr.GetString("strings.Help.start.group_prompt")
		replyMsgId := msg.MessageId
		var lowerModName string

		if len(args) == 2 {
			helpModName := args[1]
			lowerModName = strings.ToLower(helpModName)
			originalModuleName := getModuleNameFromAltName(lowerModName)
			if originalModuleName != "" && string_handling.FindInStringSlice(getAltNamesOfModule(originalModuleName), lowerModName) {
				moduleHelpString = fmt.Sprintf(tr.GetString("strings.Help.help.group_prompt_module"), originalModuleName)
				pmMeKbUri = fmt.Sprintf("https://t.me/%s?start=help_%s", b.Username, lowerModName)
			}
		}

		if msg.ReplyToMessage != nil {
			replyMsgId = msg.ReplyToMessage.MessageId
		}

		_, err = msg.Reply(b,
			moduleHelpString,
			&gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{Text: pmMeKbText, Url: pmMeKbUri},
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

/*
createModifiedCallbackContext creates a new context with modified callback data.

This is used to redirect help.* callbacks to their appropriate handlers.
*/
func createModifiedCallbackContext(ctx *ext.Context, newCallbackData string) *ext.Context {
	// Create a copy of the context with modified callback data
	newCtx := &ext.Context{
		Update: &gotgbot.Update{
			UpdateId: ctx.Update.UpdateId,
			CallbackQuery: &gotgbot.CallbackQuery{
				Id:      ctx.Update.CallbackQuery.Id,
				From:    ctx.Update.CallbackQuery.From,
				Message: ctx.Update.CallbackQuery.Message,
				Data:    newCallbackData,
			},
		},
	}
	return newCtx
}

/*
helpCallbackHandler handles callback queries with "help" prefix.

Routes help.* callbacks to their appropriate handlers by modifying the callback data.
*/
func (moduleStruct) helpCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	args := strings.Split(query.Data, ".")

	if len(args) < 2 {
		log.Warnf("[help] Invalid callback data format: %s", query.Data)
		return ext.EndGroups
	}

	action := args[1]

	switch action {
	case "about":
		// Redirect to about.main
		newCtx := createModifiedCallbackContext(ctx, "about.main")
		return helpModule.about(b, newCtx)
	case "config":
		// Redirect to configuration.step1
		newCtx := createModifiedCallbackContext(ctx, "configuration.step1")
		return HelpModule.botConfig(b, newCtx)
	case "home":
		// Redirect to start menu
		newCtx := createModifiedCallbackContext(ctx, "helpq.BackStart")
		return HelpModule.helpButtonHandler(b, newCtx)
	default:
		log.Warnf("[help] Unknown help action: %s", action)
		return ext.EndGroups
	}
}

/*
LoadHelp registers all help-related command handlers with the dispatcher.

Enables the help module and adds handlers for help, about, configuration, and donation commands.
*/
func LoadHelp(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	helpModule.cfg = cfg

	dispatcher.AddHandler(handlers.NewCommand("start", HelpModule.start))
	dispatcher.AddHandler(handlers.NewCommand("help", HelpModule.help))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("helpq"), HelpModule.helpButtonHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("help"), HelpModule.helpCallbackHandler))
	dispatcher.AddHandler(handlers.NewCommand("donate", HelpModule.donate))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("configuration"), HelpModule.botConfig))
	dispatcher.AddHandler(handlers.NewCommand("about", helpModule.about))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("about"), helpModule.about))
	initHelpButtons()
}
