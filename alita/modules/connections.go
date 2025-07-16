package modules

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

// ConnectionsModule provides logic for managing user-to-chat connections.
//
// Implements commands to connect, disconnect, and manage chat connections for users and admins.
var ConnectionsModule = moduleStruct{moduleName: "Connections"}

// connection checks the status of a user's connection to a chat in their private messages.
//
// Check the status of connection of a user in their PM
// User can check if they are connected to a chat and can also bring up the keyboard for it.
// Normal use will have just one option with 'User Commands' and admin will have "Admin Commands" along the earlier as well.
//
// Displays connection status and provides a keyboard for user/admin commands.
func (m moduleStruct) connection(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// permission checks
	if !chat_status.RequirePrivate(b, ctx, nil, false) {
		return ext.EndGroups
	}

	chatId := m.isConnected(b, ctx, user.Id)
	if chatId == 0 {
		return ext.EndGroups
	}

	chat, err := b.GetChat(chatId, nil)
	if err != nil {
		go db.DisconnectId(user.Id)
		log.Error(err)
		return err
	}
	_text := fmt.Sprintf(tr.GetString("strings.connections.connected"), chat.Title)
	connKeyboard := helpers.InitButtons(b, chat.Id, user.Id)
	_, err = msg.Reply(b,
		_text,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: connKeyboard,
			ParseMode:   helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// allowConnect allows admins to enable or disable user connections to the chat.
//
// Allow users to connect to your chat
// You can give a word such as on/off/yes/no to toggle options
// Also, if no word is given, you will get your current setting.
//
// Admins can toggle the setting or view the current status.
func (moduleStruct) allowConnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	var text string

	// permission checks
	if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		return ext.EndGroups
	}

	if len(args) >= 2 {
		toogleOption := args[1]
		switch toogleOption {
		case "on", "true", "yes":
			text = tr.GetString("strings.connections.allow_connect.turned_on")
			go db.ToggleAllowConnect(chat.Id, true)
		case "off", "false", "no":
			text = tr.GetString("strings.connections.allow_connect.turned_off")
			go db.ToggleAllowConnect(chat.Id, false)
		default:
			text = "Please give me a vaid option from <yes/on/no/off>"
		}
	} else {
		currSetting := db.GetChatConnectionSetting(chat.Id).AllowConnect
		if currSetting {
			text = tr.GetString("strings.connections.allow_connect.currently_on")
		} else {
			text = tr.GetString("strings.connections.allow_connect.currently_off")
		}
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// connect allows a user or admin to connect to a chat.
//
// Connect to a chat
// Use this command to connect to your chat!
// Admins and Users both can use this.
//
// Handles both private and group chat contexts, checks permissions, and updates the connection status.
func (moduleStruct) connect(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	var text string
	var replyMarkup gotgbot.ReplyMarkup

	if ctx.Message.Chat.Type == "private" {
		chat := extraction.ExtractChat(b, ctx)
		if chat == nil {
			return ext.EndGroups
		}

		if !db.GetChatConnectionSetting(chat.Id).AllowConnect && !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			text = tr.GetString("strings.connections.connect.connection_disabled")
		} else {
			go db.ConnectId(user.Id, chat.Id)
			text = fmt.Sprintf(tr.GetString("strings.connections.connect.connected"), chat.Title)
			replyMarkup = helpers.InitButtons(b, chat.Id, user.Id)
		}
	} else {
		if !db.GetChatConnectionSetting(chat.Id).AllowConnect && !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			text = tr.GetString("strings.connections.connect.connection_disabled")
		} else {
			text = tr.GetString("strings.connections.connect.tap_btn_connect")
			replyMarkup = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text: "Connect to chat",
							Url:  fmt.Sprintf("https://t.me/%s?start=connect_%d", b.Username, chat.Id),
						},
					},
				},
			}
		}
	}

	_, err := msg.Reply(b,
		text,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: replyMarkup,
			ParseMode:   helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// Handler for Connection buttons
// connectionButtons handles callback queries for connection-related buttons.
//
// Displays appropriate command options based on user type and connection status.
func (m moduleStruct) connectionButtons(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From
	msg := query.Message
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	args := strings.Split(query.Data, ".")
	userType := args[1]

	var (
		replyText string
		replyKb   = gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "Back",
						CallbackData: "connbtns.Main",
					},
				},
			},
		}
	)

	chatStat := m.isConnected(b, ctx, user.Id)
	if chatStat == 0 {
		return ext.EndGroups
	}

	switch userType {
	case "Admin":
		replyText = fmt.Sprintf(tr.GetString("strings.connections.connections_btns.admin_conn_cmds"), m.adminCmdConnString())
	case "User":
		replyText = fmt.Sprintf(tr.GetString("strings.connections.connections_btns.user_conn_cmds"), m.userCmdConnString())
	case "Main":
		chatId := m.isConnected(b, ctx, user.Id)
		if chatId == 0 {
			return ext.EndGroups
		}
		pchat, err := b.GetChat(chatId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		replyText = fmt.Sprintf(tr.GetString("strings.connections.connected"), pchat.Title)
		replyKb = helpers.InitButtons(b, pchat.Id, user.Id)
	}

	_, _, err := msg.EditText(b,
		replyText,
		&gotgbot.EditMessageTextOpts{
			ReplyMarkup: replyKb,
			ParseMode:   helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// disconnect disconnects a user from their currently connected chat.
//
// Disconnect from a chat
// Used to disconnect from currently connected chat
//
// Can only be used in private messages. Updates the database and replies with the result.
func (m moduleStruct) disconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	var text string

	if ctx.Message.Chat.Type == "private" {
		chatId := m.isConnected(b, ctx, user.Id)
		if chatId == 0 {
			return ext.EndGroups
		}

		go db.DisconnectId(user.Id)

		text = tr.GetString("strings.connections.disconnect.disconnected")
	} else {
		text = tr.GetString("strings.connections.disconnect.need_pm")
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// isConnected checks if a user is connected to a chat.
//
// Function used to check if user is connected to a chat or not
// If user is connected, chatId is returned else 0
//
// Returns the chat ID if connected, otherwise replies with a message and returns 0.
func (moduleStruct) isConnected(b *gotgbot.Bot, ctx *ext.Context, userId int64) int64 {
	conn := db.Connection(userId)
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	if conn.Connected {
		return conn.ChatId
	}

	_, err := ctx.EffectiveMessage.Reply(b, tr.GetString("strings.connections.not_connected"), nil)
	if err != nil {
		log.Error(err)
	}

	return 0
}

// reconnect reconnects a user to their last connected chat.
//
// Used to reconnect to last chat connected by user
// Both user and admin can use this command to connect to the previous chat
//
// Handles both user and admin contexts, checks permissions, and updates the connection status.
func (moduleStruct) reconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	var (
		connKeyboard gotgbot.InlineKeyboardMarkup
		text         string
	)

	if ctx.Message.Chat.Type == "private" {
		user := ctx.EffectiveSender.User
		chatId := db.ReconnectId(user.Id)

		if chatId != 0 {
			gchat, err := b.GetChat(chatId, nil)
			if err != nil {
				log.Error(err)
				return err
			}

			// need to convert to chat type
			_chat := gchat.ToChat()

			if !chat_status.IsUserInChat(b, &_chat, user.Id) {
				return ext.EndGroups
			}

			text = fmt.Sprintf(tr.GetString("strings.connections.reconnect.reconnected"), gchat.Title)
			connKeyboard = helpers.InitButtons(b, gchat.Id, user.Id)
		} else {
			text = tr.GetString("strings.connections.reconnect.no_last_chat")
		}
		_, err := msg.Reply(b, text,
			&gotgbot.SendMessageOpts{
				ReplyMarkup: connKeyboard,
				ParseMode:   helpers.HTML,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}

	} else {
		_, err := msg.Reply(b, tr.GetString("strings.connections.reconnect.need_pm"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// adminCmdConnString returns a formatted string of admin connection commands.
func (moduleStruct) adminCmdConnString() string {
	return "\n - /" + strings.Join(misc.AdminCmds, "\n - /")
}

// userCmdConnString returns a formatted string of user connection commands.
func (moduleStruct) userCmdConnString() string {
	return "\n - /" + strings.Join(misc.UserCmds, "\n - /")
}

// LoadConnections registers all connection-related command handlers with the dispatcher.
//
// Enables the connections module and adds handlers for connect, disconnect, and related commands.
func LoadConnections(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(ConnectionsModule.moduleName, true)
	dispatcher.AddHandler(handlers.NewCommand("connect", ConnectionsModule.connect))
	dispatcher.AddHandler(handlers.NewCommand("disconnect", ConnectionsModule.disconnect))
	dispatcher.AddHandler(handlers.NewCommand("connection", ConnectionsModule.connection))
	dispatcher.AddHandler(handlers.NewCommand("reconnect", ConnectionsModule.reconnect))
	dispatcher.AddHandler(handlers.NewCommand("allowconnect", ConnectionsModule.allowConnect))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("connbtns."), ConnectionsModule.connectionButtons))
}
