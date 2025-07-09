package chat_status

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

/*
sendAnonAdminKeyboard sends an inline keyboard to verify anonymous admin status.

It replies to the given message with a button that allows the user to confirm their admin identity.
Returns the sent message and any error encountered.
*/
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
