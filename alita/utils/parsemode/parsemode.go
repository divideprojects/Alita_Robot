package parsemode

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
)

const (
	Markdown = "Markdown"
	HTML     = "HTML"
	None     = "None"
)

func Shtml() *gotgbot.SendMessageOpts {
	return &gotgbot.SendMessageOpts{
		ParseMode:                HTML,
		DisableWebPagePreview:    true,
		AllowSendingWithoutReply: true,
	}
}

func Smarkdown() *gotgbot.SendMessageOpts {
	return &gotgbot.SendMessageOpts{
		ParseMode:                Markdown,
		DisableWebPagePreview:    true,
		AllowSendingWithoutReply: true,
	}
}
