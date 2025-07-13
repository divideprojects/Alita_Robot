package modules

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

var languagesModule = moduleStruct{
	moduleName: "Languages",
	cfg:        nil, // will be set during LoadLanguage
}

func (m moduleStruct) genFullLanguageKb() [][]gotgbot.InlineKeyboardButton {
	tr := i18n.New("en") // Use default language for this function
	keyboard := helpers.MakeLanguageKeyboard(m.cfg)

	helpTranslateText, helpTranslateErr := tr.GetStringWithError("strings.Languages.help_translate_button")
	if helpTranslateErr != nil {
		log.Errorf("[languages] missing translation for Languages.help_translate_button: %v", helpTranslateErr)
		helpTranslateText = "Help Translate"
	}

	keyboard = append(
		keyboard,
		[]gotgbot.InlineKeyboardButton{
			{
				Text: helpTranslateText,
				Url:  "https://crowdin.com/project/alita_robot",
			},
		},
	)
	return keyboard
}

func (m moduleStruct) changeLanguage(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage

	var replyString string

	cLang := db.GetLanguage(ctx)

	if ctx.Update.Message.Chat.Type == "private" {
		replyString = fmt.Sprintf("Your Current Language is %s\nChoose a language from keyboard below.", helpers.GetLangFormat(cLang))
	} else {

		// language won't be changed if user is not admin
		if !chat_status.RequireUserAdmin(b, ctx, chat, user.Id, false) {
			return ext.EndGroups
		}

		replyString = fmt.Sprintf("This Group's Current Language is %s\nChoose a language from keyboard below.", helpers.GetLangFormat(cLang))
	}

	_, err := msg.Reply(
		b,
		replyString,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: m.genFullLanguageKb(),
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func (moduleStruct) langBtnHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := query.Message.GetChat()
	user := query.From

	var replyString string
	language := strings.Split(query.Data, ".")[1]

	if ctx.Update.Message.Chat.Type == "private" {
		go db.ChangeUserLanguage(user.Id, language)
		replyString = fmt.Sprintf("Your language has been changed to %s", helpers.GetLangFormat(language))
	} else {
		go db.ChangeGroupLanguage(chat.Id, language)
		replyString = fmt.Sprintf("This group's language has been changed to %s", helpers.GetLangFormat(language))
	}

	_, _, err := query.Message.EditText(
		b,
		replyString,
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
langStartHandler handles callback queries with "chlang" prefix.

Shows the language selection menu when the language button is clicked from the start menu.
*/
func (moduleStruct) langStartHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	args := strings.Split(query.Data, ".")
	
	if len(args) < 2 {
		log.Warnf("[languages] Invalid callback data format: %s", query.Data)
		return ext.EndGroups
	}
	
	action := args[1]
	
	switch action {
	case "start":
		// Show language selection menu (same as /lang command)
		// Create a modified context that simulates a message instead of callback query
		user := query.From
		chat := query.Message.GetChat()
		
		var replyString string
		cLang := db.GetLanguage(ctx)
		
		if chat.Type == "private" {
			replyString = fmt.Sprintf("Your Current Language is %s\nChoose a language from keyboard below.", helpers.GetLangFormat(cLang))
		} else {
			// language won't be changed if user is not admin
			if !chat_status.RequireUserAdmin(b, ctx, &chat, user.Id, false) {
				return ext.EndGroups
			}
			replyString = fmt.Sprintf("This Group's Current Language is %s\nChoose a language from keyboard below.", helpers.GetLangFormat(cLang))
		}
		
		_, _, err := query.Message.EditText(
			b,
			replyString,
			&gotgbot.EditMessageTextOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: languagesModule.genFullLanguageKb(),
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
	default:
		log.Warnf("[languages] Unknown chlang action: %s", action)
		return ext.EndGroups
	}
}

func LoadLanguage(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	languagesModule.cfg = cfg

	HelpModule.AbleMap.Store(languagesModule.moduleName, true)
	HelpModule.helpableKb[languagesModule.moduleName] = languagesModule.genFullLanguageKb()

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("change_language."), languagesModule.langBtnHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("chlang"), languagesModule.langStartHandler))
	dispatcher.AddHandler(handlers.NewCommand("lang", languagesModule.changeLanguage))
}
