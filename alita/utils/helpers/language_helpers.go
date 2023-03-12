package helpers

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/Divkix/Alita_Robot/alita/config"
	"github.com/Divkix/Alita_Robot/alita/i18n"
)

func MakeLanguageKeyboard() [][]gotgbot.InlineKeyboardButton {
	var kb []gotgbot.InlineKeyboardButton

	for _, langCode := range config.ValidLangCodes {
		properLang := GetLangFormat(langCode)
		if properLang == "" || properLang == " " {
			continue
		}

		kb = append(
			kb,
			gotgbot.InlineKeyboardButton{
				Text:         properLang,
				CallbackData: fmt.Sprintf("change_language.%s", langCode),
			},
		)
	}

	return ChunkKeyboardSlices(kb, 2)
}

func GetLangFormat(langCode string) string {
	return i18n.I18n{LangCode: langCode}.GetString("main.language_name") +
		" " +
		i18n.I18n{LangCode: langCode}.GetString("main.language_flag")
}
