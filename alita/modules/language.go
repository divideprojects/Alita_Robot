package modules

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

var languagesModule = moduleStruct{moduleName: "Languages"}

func (moduleStruct) genFullLanguageKb() [][]gotgbot.InlineKeyboardButton {
	keyboard := helpers.MakeLanguageKeyboard()
	keyboard = append(
		keyboard,
		[]gotgbot.InlineKeyboardButton{
			{
				Text: "Help Us Translate 🌎",
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
	chat := query.Message.Chat
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
			ParseMode:             helpers.HTML,
			DisableWebPagePreview: true,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func LoadLanguage(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(languagesModule.moduleName, true)
	HelpModule.helpableKb[languagesModule.moduleName] = languagesModule.genFullLanguageKb()

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("change_language."), languagesModule.langBtnHandler))
	dispatcher.AddHandler(handlers.NewCommand("lang", languagesModule.changeLanguage))
}
