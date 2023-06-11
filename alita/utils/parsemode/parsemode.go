package parsemode

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
)

const (
	Markdown = "Markdown"
	HTML     = "HTML"
	None     = "None"
)

// Shtml is a shortcut for SendMessageOpts with HTML parse mode.
func Shtml() *gotgbot.SendMessageOpts {
	return &gotgbot.SendMessageOpts{
		ParseMode:                HTML,
		DisableWebPagePreview:    true,
		AllowSendingWithoutReply: true,
	}
}

// Smarkdown is a shortcut for SendMessageOpts with Markdown parse mode.
func Smarkdown() *gotgbot.SendMessageOpts {
	return &gotgbot.SendMessageOpts{
		ParseMode:                Markdown,
		DisableWebPagePreview:    true,
		AllowSendingWithoutReply: true,
	}
}
