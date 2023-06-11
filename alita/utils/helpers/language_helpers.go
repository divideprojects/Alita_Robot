package helpers

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
)

// MakeLanguageKeyboard makes a keyboard with all the languages in it.
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

// GetLangFormat returns the language name and flag.
func GetLangFormat(langCode string) string {
	return i18n.I18n{LangCode: langCode}.GetString("main.language_name") +
		" " +
		i18n.I18n{LangCode: langCode}.GetString("main.language_flag")
}
