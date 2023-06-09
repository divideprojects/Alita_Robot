package modules

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/Divkix/Alita_Robot/alita/i18n"
	"github.com/Divkix/Alita_Robot/alita/utils/chat_status"
	"github.com/Divkix/Alita_Robot/alita/utils/decorators/cmdDecorator"
	"github.com/Divkix/Alita_Robot/alita/utils/parsemode"
)

type formattingModuleStruct struct {
	modname string
}

var formattingModule = formattingModuleStruct{modname: "Formatting"}

func (m formattingModuleStruct) markdownHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Check of group or pm
	if !chat_status.RequirePrivate(b, ctx, nil, true) {
		reply := msg.Reply
		if msg.ReplyToMessage != nil {
			reply = msg.ReplyToMessage.Reply
		}
		_, err := reply(b,
			"Press the button below to get Markdown Help!",
			&gotgbot.SendMessageOpts{
				ParseMode: parsemode.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text: "Markdown Help",
								Url:  fmt.Sprintf("https://t.me/%s?start=help_formatting", b.Username),
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
		Mkdkb := append(m.genFormattingKb(),
			[]gotgbot.InlineKeyboardButton{
				{
					Text: "Back", CallbackData: "helpq.Help",
				},
			},
		)

		_, err := msg.Reply(b,
			// help.HELPABLE[ModName],

			// TODO: Fix help msg here
			"Alita supports a large number of formatting options to make your messages more expressive. Take a look!",
			&gotgbot.SendMessageOpts{
				ParseMode: parsemode.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: Mkdkb,
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

func (m formattingModuleStruct) genFormattingKb() [][]gotgbot.InlineKeyboardButton {
	fxt := "formatting.%s"

	keyboard := [][]gotgbot.InlineKeyboardButton{
		make([]gotgbot.InlineKeyboardButton, 2),
		make([]gotgbot.InlineKeyboardButton, 1),
	}

	// First row
	keyboard[0][0] = gotgbot.InlineKeyboardButton{
		Text:         "Markdown Formatting",
		CallbackData: fmt.Sprintf(fxt, "md_formatting"),
	}
	keyboard[0][1] = gotgbot.InlineKeyboardButton{
		Text:         "Fillings",
		CallbackData: fmt.Sprintf(fxt, "fillings"),
	}

	// Second Row
	keyboard[1][0] = gotgbot.InlineKeyboardButton{
		Text:         "Random Content",
		CallbackData: fmt.Sprintf(fxt, "random"),
	}

	return keyboard
}

func (m formattingModuleStruct) getMarkdownHelp(module string) string {
	var helpTxt string
	tr := i18n.I18n{LangCode: "en"}
	if module == "md_formatting" {
		helpTxt = tr.GetString("strings.Formatting.Markdown")
	} else if module == "fillings" {
		helpTxt = tr.GetString("strings.Formatting.Fillings")
	} else if module == "random" {
		helpTxt = tr.GetString("strings.Formatting.Random")
	}
	return helpTxt
}

func (m formattingModuleStruct) formattingHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	msg := query.Message

	// Get the sub-module
	module := strings.Split(query.Data, ".")[1]

	// Edit the help as per sub-module selected in markdownhelp
	_, _, err := msg.EditText(b,
		m.getMarkdownHelp(module),
		&gotgbot.EditMessageTextOpts{
			MessageId: msg.MessageId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Back", CallbackData: "helpq.Formatting"},
					},
				},
			},
			ParseMode: parsemode.HTML,
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

func LoadMkdCmd(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(formattingModule.modname, true)
	HelpModule.helpableKb[formattingModule.modname] = formattingModule.genFormattingKb()
	cmdDecorator.MultiCommand(dispatcher, []string{"markdownhelp", "formatting"}, formattingModule.markdownHelp)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("formatting."), formattingModule.formattingHandler))
}
