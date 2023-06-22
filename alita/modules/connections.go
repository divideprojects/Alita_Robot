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

var ConnectionsModule = moduleStruct{moduleName: "Connections"}

/*
	Check the status of connection of a user in their PM

User can check if they are connected to a chat and can also bring up the keyboard for it.
Normal use will have just one option with 'User Commands' and admin will have "Admin Commands" along the earlier as
well.
*/
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
	_text := fmt.Sprintf(tr.GetString("strings."+m.moduleName+".connected"), chat.Title)
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

/*
	Allow users to connect to your chat

You can give a word such as on/off/yes/no to toggle options

Also, if no word is given, you will get your current setting.
*/
func (m moduleStruct) allowConnect(b *gotgbot.Bot, ctx *ext.Context) error {
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
			text = tr.GetString("strings." + m.moduleName + ".allow_connect.turned_on")
			go db.ToggleAllowConnect(chat.Id, true)
		case "off", "false", "no":
			text = tr.GetString("strings." + m.moduleName + ".allow_connect.turned_off")
			go db.ToggleAllowConnect(chat.Id, false)
		default:
			text = "Please give me a vaid option from <yes/on/no/off>"
		}
	} else {
		currSetting := db.GetChatConnectionSetting(chat.Id).AllowConnect
		if currSetting {
			text = tr.GetString("strings." + m.moduleName + ".allow_connect.currently_on")
		} else {
			text = tr.GetString("strings." + m.moduleName + ".allow_connect.currently_off")
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
func (m moduleStruct) connect(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	var text string
	var replyMarkup gotgbot.ReplyMarkup

	if ctx.Update.Message.Chat.Type == "private" {
		chat := extraction.ExtractChat(b, ctx)
		if chat == nil {
			return ext.EndGroups
		}

		if !db.GetChatConnectionSetting(chat.Id).AllowConnect && !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			text = tr.GetString("strings." + m.moduleName + ".connect.connection_disabled")
		} else {
			go db.ConnectId(user.Id, chat.Id)
			text = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".connect.connected"), chat.Title)
			replyMarkup = helpers.InitButtons(b, chat.Id, user.Id)
		}
	} else {
		if !db.GetChatConnectionSetting(chat.Id).AllowConnect && !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			text = tr.GetString("strings." + m.moduleName + ".connect.connection_disabled")
		} else {
			text = tr.GetString("strings." + m.moduleName + ".connect.tap_btn_connect")
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
func (m moduleStruct) connectionButtons(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
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
		replyText = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".connections_btns.admin_conn_cmds"), m.adminCmdConnString())
	case "User":
		replyText = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".connections_btns.user_conn_cmds"), m.userCmdConnString())
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

		replyText = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".connected"), pchat.Title)
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

/*
	Disconnect from a chat

Used to disconnect from currently connected chat
*/
func (m moduleStruct) disconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	var text string

	if ctx.Update.Message.Chat.Type == "private" {
		chatId := m.isConnected(b, ctx, user.Id)
		if chatId == 0 {
			return ext.EndGroups
		}

		go db.DisconnectId(user.Id)

		text = tr.GetString("strings." + m.moduleName + ".disconnect.disconnected")
	} else {
		text = tr.GetString("strings." + m.moduleName + ".disconnect.need_pm")
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
func (m moduleStruct) isConnected(b *gotgbot.Bot, ctx *ext.Context, userId int64) int64 {
	conn := db.Connection(userId)
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	if conn.Connected {
		return conn.ChatId
	}

	_, err := ctx.EffectiveMessage.Reply(b, tr.GetString("strings."+m.moduleName+".not_connected"), nil)
	if err != nil {
		log.Error(err)
	}

	return 0
}

/*
	Used to reconnect to last chat connected by user

Both user and admin can use this command to connect to the previous chat
*/
func (m moduleStruct) reconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
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

			if !chat_status.IsUserInChat(b, gchat, user.Id) {
				return ext.EndGroups
			}

			text = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".reconnect.reconnected"), gchat.Title)
			connKeyboard = helpers.InitButtons(b, gchat.Id, user.Id)
		} else {
			text = tr.GetString("strings." + m.moduleName + ".reconnect.no_last_chat")
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
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".reconnect.need_pm"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

func (moduleStruct) adminCmdConnString() string {
	return "\n - /" + strings.Join(misc.AdminCmds, "\n - /")
}

func (moduleStruct) userCmdConnString() string {
	return "\n - /" + strings.Join(misc.UserCmds, "\n - /")
}

func LoadConnections(dispatcher *ext.Dispatcher) {
	// modules.helpModule.ableMap.Store(m.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("connect", ConnectionsModule.connect))
	dispatcher.AddHandler(handlers.NewCommand("disconnect", ConnectionsModule.disconnect))
	dispatcher.AddHandler(handlers.NewCommand("connection", ConnectionsModule.connection))
	dispatcher.AddHandler(handlers.NewCommand("reconnect", ConnectionsModule.reconnect))
	dispatcher.AddHandler(handlers.NewCommand("allowconnect", ConnectionsModule.allowConnect))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("connbtns."), ConnectionsModule.connectionButtons))
}
