package modules

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

/*
ConnectionsModule provides logic for managing user-to-chat connections.

Implements commands to connect, disconnect, and manage chat connections for users and admins.
*/
var ConnectionsModule = moduleStruct{
	moduleName: "Connections",
	cfg:        nil, // will be set during LoadConnections
}

/*
	Check the status of connection of a user in their PM

User can check if they are connected to a chat and can also bring up the keyboard for it.
Normal use will have just one option with 'User Commands' and admin will have "Admin Commands" along the earlier as
well.
*/
/*
connection checks the status of a user's connection to a chat in their private messages.

Displays connection status and provides a keyboard for user/admin commands.
*/
func (m moduleStruct) connection(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.New(db.GetLanguage(ctx))

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
	connectedMsg, connectedErr := tr.GetStringWithError("strings." + m.moduleName + ".connected")
	if connectedErr != nil {
		log.Errorf("[connections] missing translation for connected: %v", connectedErr)
		connectedMsg = "You are currently connected to <b>%s</b>!"
	}
	_text := fmt.Sprintf(connectedMsg, chat.Title)
	connKeyboard := helpers.InitButtons(b, chat.Id, user.Id)
	_, err = msg.Reply(b,
		_text,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: connKeyboard,
			ParseMode:   gotgbot.ParseModeHTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Allow users to connect to your chat

You can give a word such as on/off/yes/no to toggle options

Also, if no word is given, you will get your current setting.
*/
/*
allowConnect allows admins to enable or disable user connections to the chat.

Admins can toggle the setting or view the current status.
*/
func (m moduleStruct) allowConnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()
	tr := i18n.New(db.GetLanguage(ctx))

	var text string

	// permission checks
	if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		return ext.EndGroups
	}

	if len(args) >= 2 {
		toogleOption := args[1]
		switch toogleOption {
		case "on", "true", "yes":
			allowConnectTurnedOnMsg, allowConnectTurnedOnErr := tr.GetStringWithError("strings." + m.moduleName + ".allow_connect.turned_on")
			if allowConnectTurnedOnErr != nil {
				log.Errorf("[connections] missing translation for allow_connect.turned_on: %v", allowConnectTurnedOnErr)
				allowConnectTurnedOnMsg = "Turned <b>on</b> User connections to this chat!\nUsers can now connect chat to their PMs!"
			}
			text = allowConnectTurnedOnMsg
			go db.ToggleAllowConnect(chat.Id, true)
		case "off", "false", "no":
			allowConnectTurnedOffMsg, allowConnectTurnedOffErr := tr.GetStringWithError("strings." + m.moduleName + ".allow_connect.turned_off")
			if allowConnectTurnedOffErr != nil {
				log.Errorf("[connections] missing translation for allow_connect.turned_off: %v", allowConnectTurnedOffErr)
				allowConnectTurnedOffMsg = "Turned <b>off</b> User connections to this chat!\nUsers can't connect chat to their PM's!"
			}
			text = allowConnectTurnedOffMsg
			go db.ToggleAllowConnect(chat.Id, false)
		default:
			text = "Please give me a vaid option from <yes/on/no/off>"
		}
	} else {
		currSetting := db.GetChatConnectionSetting(chat.Id).AllowConnect
		if currSetting {
			allowConnectCurrentlyOnMsg, allowConnectCurrentlyOnErr := tr.GetStringWithError("strings." + m.moduleName + ".allow_connect.currently_on")
			if allowConnectCurrentlyOnErr != nil {
				log.Errorf("[connections] missing translation for allow_connect.currently_on: %v", allowConnectCurrentlyOnErr)
				allowConnectCurrentlyOnMsg = "User connections are currently turned <b>on</b>.\nUsers can connect this chat to their PMs!"
			}
			text = allowConnectCurrentlyOnMsg
		} else {
			allowConnectCurrentlyOffMsg, allowConnectCurrentlyOffErr := tr.GetStringWithError("strings." + m.moduleName + ".allow_connect.currently_off")
			if allowConnectCurrentlyOffErr != nil {
				log.Errorf("[connections] missing translation for allow_connect.currently_off: %v", allowConnectCurrentlyOffErr)
				allowConnectCurrentlyOffMsg = "User connections are currently turned <b>off</b>.\nUsers can't connect this chat to their PMs!"
			}
			text = allowConnectCurrentlyOffMsg
		}
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Connect to a chat

Use this command to connect to your chat!

Admins and Users both can use this.
*/
/*
connect allows a user or admin to connect to a chat.

Handles both private and group chat contexts, checks permissions, and updates the connection status.
*/
func (m moduleStruct) connect(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.New(db.GetLanguage(ctx))
	var text string
	var replyMarkup gotgbot.ReplyMarkup

	if ctx.Update.Message.Chat.Type == "private" {
		chat := extraction.ExtractChat(b, ctx)
		if chat == nil {
			return ext.EndGroups
		}

		if !db.GetChatConnectionSetting(chat.Id).AllowConnect && !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			connectionDisabledMsg, connectionDisabledErr := tr.GetStringWithError("strings." + m.moduleName + ".connect.connection_disabled")
			if connectionDisabledErr != nil {
				log.Errorf("[connections] missing translation for connect.connection_disabled: %v", connectionDisabledErr)
				connectionDisabledMsg = "User connections are currently <b>disabled</b> to this chat.\nPlease ask admins to allow if you want to connect!"
			}
			text = connectionDisabledMsg
		} else {
			go db.ConnectId(user.Id, chat.Id)
			connectConnectedMsg, connectConnectedErr := tr.GetStringWithError("strings." + m.moduleName + ".connect.connected")
			if connectConnectedErr != nil {
				log.Errorf("[connections] missing translation for connect.connected: %v", connectConnectedErr)
				connectConnectedMsg = "You are now connected to <b>%s</b>!"
			}
			text = fmt.Sprintf(connectConnectedMsg, chat.Title)
			replyMarkup = helpers.InitButtons(b, chat.Id, user.Id)
		}
	} else {
		if !db.GetChatConnectionSetting(chat.Id).AllowConnect && !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			connectionDisabledMsg2, connectionDisabledErr2 := tr.GetStringWithError("strings." + m.moduleName + ".connect.connection_disabled")
			if connectionDisabledErr2 != nil {
				log.Errorf("[connections] missing translation for connect.connection_disabled: %v", connectionDisabledErr2)
				connectionDisabledMsg2 = "User connections are currently <b>disabled</b> to this chat.\nPlease ask admins to allow if you want to connect!"
			}
			text = connectionDisabledMsg2
		} else {
			tapBtnConnectMsg, tapBtnConnectErr := tr.GetStringWithError("strings." + m.moduleName + ".connect.tap_btn_connect")
			if tapBtnConnectErr != nil {
				log.Errorf("[connections] missing translation for connect.tap_btn_connect: %v", tapBtnConnectErr)
				tapBtnConnectMsg = "Please press the button below to connect this chat to your PM."
			}
			text = tapBtnConnectMsg
			connectToChatButtonMsg, connectToChatButtonErr := tr.GetStringWithError("strings.Connections.connect_to_chat_button")
			if connectToChatButtonErr != nil {
				log.Errorf("[connections] missing translation for connect_to_chat_button: %v", connectToChatButtonErr)
				connectToChatButtonMsg = "Connect to chat"
			}
			replyMarkup = gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text: connectToChatButtonMsg,
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
			ParseMode:   gotgbot.ParseModeHTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// Handler for Connection buttons
/*
connectionButtons handles callback queries for connection-related buttons.

Displays appropriate command options based on user type and connection status.
*/
func (m moduleStruct) connectionButtons(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	user := query.From
	msg := query.Message
	tr := i18n.New(db.GetLanguage(ctx))

	args := strings.Split(query.Data, ".")
	userType := args[1]

	backButtonMsg, backButtonErr := tr.GetStringWithError("strings.CommonStrings.buttons.back")
	if backButtonErr != nil {
		log.Errorf("[connections] missing translation for buttons.back: %v", backButtonErr)
		backButtonMsg = "Back"
	}
	var (
		replyText string
		replyKb   = gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         backButtonMsg,
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
		adminConnCmdsMsg, adminConnCmdsErr := tr.GetStringWithError("strings." + m.moduleName + ".connections_btns.admin_conn_cmds")
		if adminConnCmdsErr != nil {
			log.Errorf("[connections] missing translation for connections_btns.admin_conn_cmds: %v", adminConnCmdsErr)
			adminConnCmdsMsg = "Available Admin commands:%s"
		}
		replyText = fmt.Sprintf(adminConnCmdsMsg, m.adminCmdConnString())
	case "User":
		userConnCmdsMsg, userConnCmdsErr := tr.GetStringWithError("strings." + m.moduleName + ".connections_btns.user_conn_cmds")
		if userConnCmdsErr != nil {
			log.Errorf("[connections] missing translation for connections_btns.user_conn_cmds: %v", userConnCmdsErr)
			userConnCmdsMsg = "Available User commands:%s"
		}
		replyText = fmt.Sprintf(userConnCmdsMsg, m.userCmdConnString())
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

		connectedMsg2, connectedErr2 := tr.GetStringWithError("strings." + m.moduleName + ".connected")
		if connectedErr2 != nil {
			log.Errorf("[connections] missing translation for connected: %v", connectedErr2)
			connectedMsg2 = "You are currently connected to <b>%s</b>!"
		}
		replyText = fmt.Sprintf(connectedMsg2, pchat.Title)
		replyKb = helpers.InitButtons(b, pchat.Id, user.Id)
	}

	_, _, err := msg.EditText(b,
		replyText,
		&gotgbot.EditMessageTextOpts{
			ReplyMarkup: replyKb,
			ParseMode:   gotgbot.ParseModeHTML,
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

/*
	Disconnect from a chat

Used to disconnect from currently connected chat
*/
/*
disconnect disconnects a user from their currently connected chat.

Can only be used in private messages. Updates the database and replies with the result.
*/
func (m moduleStruct) disconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.New(db.GetLanguage(ctx))

	var text string

	if ctx.Update.Message.Chat.Type == "private" {
		chatId := m.isConnected(b, ctx, user.Id)
		if chatId == 0 {
			return ext.EndGroups
		}

		go db.DisconnectId(user.Id)

		disconnectedMsg, disconnectedErr := tr.GetStringWithError("strings." + m.moduleName + ".disconnect.disconnected")
		if disconnectedErr != nil {
			log.Errorf("[connections] missing translation for disconnect.disconnected: %v", disconnectedErr)
			disconnectedMsg = "Successfully disconnected from the connected chat."
		}
		text = disconnectedMsg
	} else {
		needPmMsg, needPmErr := tr.GetStringWithError("strings." + m.moduleName + ".disconnect.need_pm")
		if needPmErr != nil {
			log.Errorf("[connections] missing translation for disconnect.need_pm: %v", needPmErr)
			needPmMsg = "You need to send this in PM to me to disconnect from the chat!"
		}
		text = needPmMsg
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Function used to check if user is connected to a chat or not

If user is connected, chatId is returned else 0
*/
/*
isConnected checks if a user is connected to a chat.

Returns the chat ID if connected, otherwise replies with a message and returns 0.
*/
func (m moduleStruct) isConnected(b *gotgbot.Bot, ctx *ext.Context, userId int64) int64 {
	conn := db.Connection(userId)
	tr := i18n.New(db.GetLanguage(ctx))

	if conn.Connected {
		return conn.ChatId
	}

	notConnectedMsg, notConnectedErr := tr.GetStringWithError("strings." + m.moduleName + ".not_connected")
	if notConnectedErr != nil {
		log.Errorf("[connections] missing translation for not_connected: %v", notConnectedErr)
		notConnectedMsg = "You aren't connected to any chats."
	}
	_, err := ctx.EffectiveMessage.Reply(b, notConnectedMsg, nil)
	if err != nil {
		log.Error(err)
	}

	return 0
}

/*
	Used to reconnect to last chat connected by user

Both user and admin can use this command to connect to the previous chat
*/
/*
reconnect reconnects a user to their last connected chat.

Handles both user and admin contexts, checks permissions, and updates the connection status.
*/
func (m moduleStruct) reconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))
	var (
		connKeyboard gotgbot.InlineKeyboardMarkup
		text         string
	)

	if ctx.Update.Message.Chat.Type == "private" {
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

			reconnectedMsg, reconnectedErr := tr.GetStringWithError("strings." + m.moduleName + ".reconnect.reconnected")
			if reconnectedErr != nil {
				log.Errorf("[connections] missing translation for reconnect.reconnected: %v", reconnectedErr)
				reconnectedMsg = "You are now reconnected to <b>%s</b>!!"
			}
			text = fmt.Sprintf(reconnectedMsg, gchat.Title)
			connKeyboard = helpers.InitButtons(b, gchat.Id, user.Id)
		} else {
			noLastChatMsg, noLastChatErr := tr.GetStringWithError("strings." + m.moduleName + ".reconnect.no_last_chat")
			if noLastChatErr != nil {
				log.Errorf("[connections] missing translation for reconnect.no_last_chat: %v", noLastChatErr)
				noLastChatMsg = "You have no last chat to reconnect!"
			}
			text = noLastChatMsg
		}
		_, err := msg.Reply(b, text,
			&gotgbot.SendMessageOpts{
				ReplyMarkup: connKeyboard,
				ParseMode:   gotgbot.ParseModeHTML,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}

	} else {
		reconnectNeedPmMsg, reconnectNeedPmErr := tr.GetStringWithError("strings." + m.moduleName + ".reconnect.need_pm")
		if reconnectNeedPmErr != nil {
			log.Errorf("[connections] missing translation for reconnect.need_pm: %v", reconnectNeedPmErr)
			reconnectNeedPmMsg = "You need to be in a PM with me to reconnect to a chat!"
		}
		_, err := msg.Reply(b, reconnectNeedPmMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

/*
adminCmdConnString returns a formatted string of admin connection commands.
*/
func (moduleStruct) adminCmdConnString() string {
	return "\n - /" + strings.Join(misc.AdminCmds, "\n - /")
}

/*
userCmdConnString returns a formatted string of user connection commands.
*/
func (moduleStruct) userCmdConnString() string {
	return "\n - /" + strings.Join(misc.UserCmds, "\n - /")
}

/*
LoadConnections registers all connection-related command handlers with the dispatcher.

Enables the connections module and adds handlers for connect, disconnect, and related commands.
*/
func LoadConnections(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	ConnectionsModule.cfg = cfg

	// modules.helpModule.ableMap.Store(m.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("connect", ConnectionsModule.connect))
	dispatcher.AddHandler(handlers.NewCommand("disconnect", ConnectionsModule.disconnect))
	dispatcher.AddHandler(handlers.NewCommand("connection", ConnectionsModule.connection))
	dispatcher.AddHandler(handlers.NewCommand("reconnect", ConnectionsModule.reconnect))
	dispatcher.AddHandler(handlers.NewCommand("allowconnect", ConnectionsModule.allowConnect))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("connbtns."), ConnectionsModule.connectionButtons))
}
