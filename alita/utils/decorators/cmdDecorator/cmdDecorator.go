package cmdDecorator

import (
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

/*
MultiCommand registers multiple command aliases with the dispatcher.

Each alias in the provided slice is added as a command handler using the given response handler.
This allows a single handler to respond to multiple command triggers.
*/
func MultiCommand(dispatcher *ext.Dispatcher, alias []string, r handlers.Response) {
	for _, i := range alias {
		dispatcher.AddHandler(handlers.NewCommand(i, r))
	}
}
