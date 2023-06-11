package helpers

import (
	"fmt"

	tgmd2html "github.com/PaulSonOfLars/gotg_md2html"
	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// BuildKeyboard is used to build a keyboard from a list of buttons provided by the database.
func BuildKeyboard(buttons []db.Button) [][]gotgbot.InlineKeyboardButton {
	keyb := make([][]gotgbot.InlineKeyboardButton, 0)
	for _, btn := range buttons {
		if btn.SameLine && len(keyb) > 0 {
			keyb[len(keyb)-1] = append(keyb[len(keyb)-1], gotgbot.InlineKeyboardButton{Text: btn.Name, Url: btn.Url})
		} else {
			k := make([]gotgbot.InlineKeyboardButton, 1)
			k[0] = gotgbot.InlineKeyboardButton{Text: btn.Name, Url: btn.Url}
			keyb = append(keyb, k)
		}
	}
	return keyb
}

// ConvertButtonV2ToDbButton is used to convert []tgmd2html.ButtonV2 to []db.Button
func ConvertButtonV2ToDbButton(buttons []tgmd2html.ButtonV2) (btns []db.Button) {
	btns = make([]db.Button, len(buttons))
	for i, btn := range buttons {
		btns[i] = db.Button{
			Name:     btn.Name,
			Url:      btn.Text,
			SameLine: btn.SameLine,
		}
	}
	return
}

// RevertButtons is used to convert []db.Button to string
func RevertButtons(buttons []db.Button) string {
	res := ""
	for _, btn := range buttons {
		if btn.SameLine {
			res += fmt.Sprintf("\n[%s](buttonurl://%s)", btn.Name, btn.Url)
		} else {
			res += fmt.Sprintf("\n[%s](buttonurl://%s:same)", btn.Name, btn.Url)
		}
	}
	return res
}

// InlineKeyboardMarkupToTgmd2htmlButtonV2 this func is used to convert gotgbot.InlineKeyboardarkup to []tgmd2html.ButtonV2
func InlineKeyboardMarkupToTgmd2htmlButtonV2(replyMarkup *gotgbot.InlineKeyboardMarkup) (btns []tgmd2html.ButtonV2) {
	btns = make([]tgmd2html.ButtonV2, 0)
	for _, inlineKeyboard := range replyMarkup.InlineKeyboard {
		if len(inlineKeyboard) > 1 {
			for i, button := range inlineKeyboard {
				// if any button has anything other than url, it's not a valid button
				// skip options such as CallbackData, CallbackUrl, etc.
				if button.Url == "" {
					continue
				}

				sameline := true
				if i == 0 {
					sameline = false
				}
				btns = append(
					btns,
					tgmd2html.ButtonV2{
						Name:     button.Text,
						Text:     button.Url,
						SameLine: sameline,
					},
				)
			}
		} else {
			btns = append(btns,
				tgmd2html.ButtonV2{
					Name:     inlineKeyboard[0].Text,
					Text:     inlineKeyboard[0].Url,
					SameLine: false,
				},
			)
		}
	}
	return
}

// ChunkKeyboardSlices function used in making the help menu keyboard
func ChunkKeyboardSlices(slice []gotgbot.InlineKeyboardButton, chunkSize int) (chunks [][]gotgbot.InlineKeyboardButton) {
	for {
		if len(slice) == 0 {
			break
		}
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]

	}
	return chunks
}
