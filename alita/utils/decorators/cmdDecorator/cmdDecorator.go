package cmdDecorator

import (
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

// MultiCommand registers multiple command aliases with the same handler function.
// Useful for creating command shortcuts and alternative names for the same functionality.
func MultiCommand(dispatcher *ext.Dispatcher, alias []string, r handlers.Response) {
	for _, i := range alias {
		dispatcher.AddHandler(handlers.NewCommand(i, r))
	}
}
