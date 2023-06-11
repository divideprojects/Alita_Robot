package chat_status

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// sendAnonAdminKeyboard
// sends a keyboard with button to verify anonymous admin status
func sendAnonAdminKeyboard(b *gotgbot.Bot, msg *gotgbot.Message, chat *gotgbot.Chat) (*gotgbot.Message, error) {
	return msg.Reply(b,
		"It looks like you're anonymous. Tap this button to confirm your identity.",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{{
						Text:         "Click to prove admin",
						CallbackData: fmt.Sprintf("anonAdmin.%d.%d", chat.Id, msg.MessageId),
					}},
				},
			},
		},
	)
}
