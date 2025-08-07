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

var HelpModule = moduleStruct{
	moduleName:     "Help",
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

var (
	backBtnSuffix = []gotgbot.InlineKeyboardButton{
		{
			Text:         "« Back",
			CallbackData: "helpq.Help",
		},
		{
			Text:         "Home",
			CallbackData: "helpq.BackStart",
		},
	}
	aboutKb = gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "About me 👨\u200d💻",
					CallbackData: "about.me",
				},
			},
			{
				{
					Text: "News Channel 📢",
					Url:  "https://t.me/AlitaRobotUpdates",
				},
				{
					Text: "Support Group 👥",
					Url:  "https://t.me/DivideSupport",
				},
			},
			{
				{
					Text:         "Configuration ⚙️",
					CallbackData: "configuration.step1",
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
	startMarkup = gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "About ✨",
					CallbackData: "about.main",
				},
			},
			{
				{
					Text: "➕ Add me to chat!",
					Url:  "https://t.me/Alita_Robot?startgroup=botstart",
				},
				{
					Text: "Support Group 👥",
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
					Text:         "Language 🌏",
					CallbackData: "helpq.Languages",
				},
			},
		},
	}
)

type moduleEnabled struct {
	modules map[string]bool
}

// Init initializes the module enabled map for tracking enabled bot modules.
// Sets up the internal map structure for storing module activation states.
func (m *moduleEnabled) Init() {
	m.modules = make(map[string]bool)
}

// Store sets the enabled status for a specific module in the bot.
// Records whether a module is active and available for use.
func (m *moduleEnabled) Store(module string, enabled bool) {
	m.modules[module] = enabled
}

// Load retrieves the enabled status for a specific module.
// Returns the module name and whether it's currently enabled and accessible.
func (m *moduleEnabled) Load(module string) (string, bool) {
	log.Info(fmt.Sprintf("[Module] Loading %s module", module))
	return module, m.modules[module]
}

// LoadModules returns a slice of all currently enabled module names.
// Provides a list of active modules that users can access and get help for.
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

// about displays information about the bot including version and features.
// Shows bot details, links to support channels, and configuration options.
func (moduleStruct) about(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	var (
		currText string
		currKb   gotgbot.InlineKeyboardMarkup
	)

	if query := ctx.CallbackQuery; query != nil {
		args := strings.Split(query.Data, ".")
		response := args[1]

		switch response {
		case "main":
			currText = aboutText
			currKb = aboutKb
		case "me":
			temp, _ := tr.GetString("strings.Help.About")
			currText = fmt.Sprintf(temp, b.Username, config.BotVersion)
			currKb = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text:         "Back",
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
	} else {
		if ctx.Message.Chat.Type == "private" {
			currText = aboutText
			currKb = aboutKb
		} else {
			currText = "Click on the button below to get info about me!"
			currKb = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text: "About",
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
				ParseMode: helpers.HTML,
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

// helpButtonHandler processes callback queries from help menu button interactions.
// Navigates between help sections and displays appropriate help content for modules.
func (moduleStruct) helpButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	args := strings.Split(query.Data, ".")
	module := args[1]

	var (
		parsemode, helpText string
		replyKb             gotgbot.InlineKeyboardMarkup
	)

	// Sort the module names
	if string_handling.FindInStringSlice([]string{"BackStart", "Help"}, module) {
		parsemode = helpers.HTML
		switch module {
		case "Help":
			// This shows the main start menu
			helpText = fmt.Sprintf(mainhlp, html.EscapeString(query.From.FirstName))
			replyKb = markup
		case "BackStart":
			// This shows the modules menu
			helpText = startHelp
			replyKb = startMarkup
		}
	} else {
		// For all remaining modules
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
// start handles the /start command and displays welcome message with navigation options.
// Shows different content in private vs group chats and handles start parameters.
func (moduleStruct) start(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	args := ctx.Args()

	if ctx.Message.Chat.Type == "private" {
		if len(args) == 1 {
			_, err := msg.Reply(b,
				startHelp,
				&gotgbot.SendMessageOpts{
					ParseMode: helpers.HTML,
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
		_, err := msg.Reply(b, "Hey :) PM me if you have any questions on how to use me!", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// donate displays information about supporting the bot and its development.
// Shows donation links and ways users can contribute to bot maintenance.
func (moduleStruct) donate(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	_, err := b.SendMessage(chat.Id,
		func() string {
			tr := i18n.MustNewTranslator("en")
			text, _ := tr.GetString("strings.Help.DonateText")
			return text
		}(),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
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

// botConfig provides step-by-step configuration guidance for new users.
// Walks users through adding the bot to chats and basic setup procedures.
func (moduleStruct) botConfig(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	msg := query.Message

	// just in case
	if msg.GetChat().Type != "private" {
		_, _, err := msg.EditText(b, "Configuration only works in private", nil)
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

	tr := i18n.MustNewTranslator("en")

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
					Text:         "Done ✅",
					CallbackData: "configuration.step2",
				},
			},
		}
		text, _ = tr.GetString("strings.Help.Configuration.Step-1")
	case "step2":
		iKeyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Done ✅",
					CallbackData: "configuration.step3",
				},
			},
		}
		temp, _ := tr.GetString("strings.Help.Configuration.Step-2")
		text = fmt.Sprintf(temp, b.Username)
	case "step3":
		iKeyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "✅ Continue ✅",
					CallbackData: "helpq.Help",
				},
			},
		}
		text, _ = tr.GetString("strings.Help.Configuration.Step-3")
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

// help displays the main help menu or specific module help information.
// Shows module list in private messages or provides links to PM help in groups.
func (moduleStruct) help(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	args := ctx.Args()

	if ctx.Message.Chat.Type == "private" {
		if len(args) == 1 {
			_, err := b.SendMessage(chat.Id,
				fmt.Sprintf(
					mainhlp,
					html.EscapeString(msg.From.FirstName),
				),
				&gotgbot.SendMessageOpts{
					ParseMode:   helpers.HTML,
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
			_, err := b.SendMessage(
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
		pmMeKbText := "Click here for help!"
		pmMeKbUri := fmt.Sprintf("https://t.me/%s?start=help_help", b.Username)
		moduleHelpString := "Contact me in PM for help!"
		replyMsgId := msg.MessageId
		var lowerModName string

		if len(args) == 2 {
			helpModName := args[1]
			lowerModName = strings.ToLower(helpModName)
			originalModuleName := getModuleNameFromAltName(lowerModName)
			if originalModuleName != "" && string_handling.FindInStringSlice(getAltNamesOfModule(originalModuleName), lowerModName) {
				moduleHelpString = fmt.Sprintf("Contact me in PM for help regarding <code>%s</code>!", originalModuleName)
				pmMeKbUri = fmt.Sprintf("https://t.me/%s?start=help_%s", b.Username, lowerModName)
			}
		}

		if msg.ReplyToMessage != nil {
			replyMsgId = msg.ReplyToMessage.MessageId
		}

		_, err := msg.Reply(b,
			moduleHelpString,
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
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

// LoadHelp registers all help-related command and callback handlers.
// Sets up the help system including start, about, donate, and configuration commands.
func LoadHelp(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("start", HelpModule.start))
	dispatcher.AddHandler(handlers.NewCommand("help", HelpModule.help))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("helpq"), HelpModule.helpButtonHandler))
	dispatcher.AddHandler(handlers.NewCommand("donate", HelpModule.donate))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("configuration"), HelpModule.botConfig))
	dispatcher.AddHandler(handlers.NewCommand("about", HelpModule.about))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("about"), HelpModule.about))
	initHelpButtons()
}
