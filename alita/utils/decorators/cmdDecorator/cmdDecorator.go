package cmdDecorator

import (
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

// MultiCommand adds multiple commands to the dispatcher.
func MultiCommand(dispatcher *ext.Dispatcher, alias []string, r handlers.Response) {
	for _, i := range alias {
		dispatcher.AddHandler(handlers.NewCommand(i, r))
	}
}
