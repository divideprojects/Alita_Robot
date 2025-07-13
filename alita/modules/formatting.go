package modules

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/cmdDecorator"
)

/*
formattingModule provides logic for formatting help and markdown support.

Implements commands and handlers for markdown help and formatting options.
*/
var formattingModule = moduleStruct{
	moduleName: "Formatting",
	cfg:        nil, // will be set during LoadMkdCmd
}

/*
markdownHelp provides markdown and formatting help to users.

Displays help in private chat or via a button in group chats, with a keyboard for navigation.
*/
func (m moduleStruct) markdownHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
	msg := ctx.EffectiveMessage

	// Check of group or pm
	if !chat_status.RequirePrivate(b, ctx, nil, true) {
		reply := msg.Reply
		if msg.ReplyToMessage != nil {
			reply = msg.ReplyToMessage.Reply
		}
		markdownHelpMsg, markdownHelpErr := tr.GetStringWithError("strings.Formatting.markdown_help")
		if markdownHelpErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", markdownHelpErr)
			markdownHelpMsg = "You can use markdown in your messages. This allows for more expressive formatting."
		}

		backButtonMsg, backButtonErr := tr.GetStringWithError("strings.CommonStrings.buttons.back")
		if backButtonErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", backButtonErr)
			backButtonMsg = "Back"
		}

		_, err := reply(b,
			markdownHelpMsg,
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text: backButtonMsg, CallbackData: "helpq.Help",
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
	} else {

		// Keyboard for markdown help menu
		markdownFormattingMsg, markdownFormattingErr := tr.GetStringWithError("strings.Formatting.markdown_formatting")
		if markdownFormattingErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", markdownFormattingErr)
			markdownFormattingMsg = "Markdown Formatting"
		}

		fillingsMsg, fillingsErr := tr.GetStringWithError("strings.Formatting.fillings")
		if fillingsErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", fillingsErr)
			fillingsMsg = "Fillings"
		}

		randomContentMsg, randomContentErr := tr.GetStringWithError("strings.Formatting.random_content")
		if randomContentErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", randomContentErr)
			randomContentMsg = "Random Content"
		}

		backButtonMsg, backButtonErr := tr.GetStringWithError("strings.CommonStrings.buttons.back")
		if backButtonErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", backButtonErr)
			backButtonMsg = "Back"
		}

		keyboard := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         markdownFormattingMsg,
						CallbackData: "helpq.Formatting.Markdown",
					},
				},
				{
					{
						Text:         fillingsMsg,
						CallbackData: "helpq.Formatting.Fillings",
					},
				},
				{
					{
						Text:         randomContentMsg,
						CallbackData: "helpq.Formatting.Random",
					},
				},
				{
					{Text: backButtonMsg, CallbackData: "helpq.Formatting"},
				},
			},
		}

		_, err := msg.Reply(b,
			// help.HELPABLE[ModName],

			// TODO: Fix help msg here
			"Alita supports a large number of formatting options to make your messages more expressive. Take a look!",
			&gotgbot.SendMessageOpts{
				ParseMode:   helpers.HTML,
				ReplyMarkup: keyboard,
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
genFormattingKb generates the inline keyboard for formatting help options.

Returns a keyboard with buttons for markdown, fillings, and random content.
*/
func (m moduleStruct) genFormattingKb() [][]gotgbot.InlineKeyboardButton {
	return [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         "Markdown Formatting",
				CallbackData: "helpq.Formatting.Markdown",
			},
		},
		{
			{
				Text:         "Fillings",
				CallbackData: "helpq.Formatting.Fillings",
			},
		},
		{
			{
				Text:         "Random Content",
				CallbackData: "helpq.Formatting.Random",
			},
		},
	}
}

/*
getMarkdownHelp returns the help text for a given formatting sub-module.

Supports markdown formatting, fillings, and random content.
*/
func (m moduleStruct) getFormattingHelp(formattingType string, tr *i18n.I18n) string {
	var helpTxt string

	switch formattingType {
	case "Markdown":
		markdownMsg, markdownErr := tr.GetStringWithError("strings.Formatting.Markdown")
		if markdownErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", markdownErr)
			markdownMsg = "Markdown formatting help"
		}
		helpTxt = markdownMsg
	case "Fillings":
		fillingsMsg, fillingsErr := tr.GetStringWithError("strings.Formatting.Fillings")
		if fillingsErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", fillingsErr)
			fillingsMsg = "Fillings help"
		}
		helpTxt = fillingsMsg
	case "Random":
		randomMsg, randomErr := tr.GetStringWithError("strings.Formatting.Random")
		if randomErr != nil {
			log.Errorf("[formatting] missing translation for key: %v", randomErr)
			randomMsg = "Random content help"
		}
		helpTxt = randomMsg
	}
	return helpTxt
}

/*
formattingHandler handles callback queries for formatting help sub-modules.

Edits the help message to display the selected formatting topic.
*/
func (m moduleStruct) formattingHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	tr := i18n.New(db.GetLanguage(ctx))
	args := strings.SplitN(query.Data, ".", 3)
	module := strings.ToLower(args[2])

	helpTxt := m.getFormattingHelp(module, tr)

	backButtonMsg, backButtonErr := tr.GetStringWithError("strings.CommonStrings.buttons.back")
	if backButtonErr != nil {
		log.Errorf("[formatting] missing translation for key: %v", backButtonErr)
		backButtonMsg = "Back"
	}

	_, _, err := query.Message.EditText(b,
		helpTxt,
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: backButtonMsg, CallbackData: "helpq.Formatting"},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{})
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
LoadMkdCmd registers formatting-related command handlers with the dispatcher.

Enables the formatting module and adds handlers for markdown help and formatting options.
*/
func LoadMkdCmd(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	formattingModule.cfg = cfg

	HelpModule.AbleMap.Store(formattingModule.moduleName, true)
	HelpModule.helpableKb[formattingModule.moduleName] = formattingModule.genFormattingKb()
	cmdDecorator.MultiCommand(dispatcher, []string{"markdownhelp", "formatting"}, formattingModule.markdownHelp)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("formatting."), formattingModule.formattingHandler))
}
