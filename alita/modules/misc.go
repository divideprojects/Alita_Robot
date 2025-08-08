package modules

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/divideprojects/Alita_Robot/alita/db"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

var (
	miscModule = moduleStruct{moduleName: "Misc"}
	// HTTP client with timeout and connection pooling for external requests
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  true,
		},
	}
)

// echomsg handles the /tell command to make the bot echo a message
// as a reply to another message, requiring admin permissions.
func (moduleStruct) echomsg(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]

	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.IsUserAdmin(b, msg.Chat.Id, msg.From.Id) {
		return ext.EndGroups
	}

	replyMsg := msg.ReplyToMessage
	if replyMsg == nil {
		_, _ = msg.Reply(b, "Reply to someone.", nil)
		return ext.EndGroups
	}

	if len(args) > 0 {
		_, _ = msg.Delete(b, nil)
		_, err := msg.Reply(b,
			strings.Join(
				strings.Split(msg.OriginalHTML(), " ")[1:], " ",
			),
			&gotgbot.SendMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId: replyMsg.MessageId,
				},
				ParseMode: helpers.Shtml().ParseMode,
			},
		)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, _ = msg.Reply(b, "Provide some content to reply!", nil)
	}

	return ext.EndGroups
}

// getId handles the /id command to display IDs of users, chats,
// files, and forwarded messages with detailed information.
func (moduleStruct) getId(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	}
	var builder strings.Builder
	builder.Grow(512) // Pre-allocate capacity for better performance

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "id") {
		return ext.EndGroups
	}

	if userId != 0 {
		if msg.ReplyToMessage != nil {
			builder.WriteString(fmt.Sprintf(
				"<b>Chat ID:</b> <code>%d</code>\n",
				msg.Chat.Id,
			))
			if msg.IsTopicMessage {
				builder.WriteString(fmt.Sprintf("Thread Id: <code>%d</code>\n", msg.MessageThreadId))
			}
			if msg.ReplyToMessage.From != nil {
				originalId := msg.ReplyToMessage.From.Id
				_, user1Name, _ := extraction.GetUserInfo(originalId)
				builder.WriteString(fmt.Sprintf(
					"<b>%s's ID:</b> <code>%d</code>\n",
					user1Name,
					originalId,
				))
			}

			if rpm := msg.ReplyToMessage; rpm != nil {
				if frpm := rpm.ForwardOrigin; frpm != nil {
					if frpm.GetDate() != 0 {
						fwdd := frpm.MergeMessageOrigin()

						if fwdc := fwdd.SenderUser; fwdc != nil {
							user1Id := fwdc.Id
							_, user1Name, _ := extraction.GetUserInfo(user1Id)
							builder.WriteString(fmt.Sprintf(
								"<b>Forwarded from %s's ID:</b> <code>%d</code>\n",
								user1Name, user1Id,
							))
						}

						if fwdc := fwdd.Chat; fwdc != nil {
							builder.WriteString(fmt.Sprintf("<b>Forwarded from chat %s's ID:</b> <code>%d</code>\n",
								fwdc.Title, fwdc.Id,
							))
						}
					}
				}
			}
			if msg.ReplyToMessage.Animation != nil {
				builder.WriteString(fmt.Sprintf("<b>GIF ID:</b> <code>%s</code>\n",
					msg.ReplyToMessage.Animation.FileId,
				))
			}
			if msg.ReplyToMessage.Sticker != nil {
				builder.WriteString(fmt.Sprintf("<b>Sticker ID:</b> <code>%s</code>\n",
					msg.ReplyToMessage.Sticker.FileId,
				))
			}
		} else {
			_, name, _ := extraction.GetUserInfo(userId)
			builder.WriteString(fmt.Sprintf("%s's ID is <code>%d</code>", name, userId))
		}
	} else {
		chat := ctx.EffectiveChat
		if ctx.Message.Chat.Type == "private" {
			builder.WriteString(fmt.Sprintf("Your ID is <code>%d</code>", chat.Id))
		} else {
			builder.WriteString(fmt.Sprintf("Your ID is <code>%d</code>\nThis group's ID is <code>%d</code>",
				msg.From.Id, chat.Id,
			))
		}
	}

	_, err := msg.Reply(b,
		builder.String(),
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// ping handles the /ping command to measure and display
// the bot's response time in milliseconds.
func (moduleStruct) ping(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "ping") {
		return ext.EndGroups
	}
	stime := time.Now()
	rmsg, _ := msg.Reply(b, "<code>Pinging</code>", &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})
	_, _, err := rmsg.EditText(b, fmt.Sprintf("Pinged in %d ms", int64(time.Since(stime)/time.Millisecond)), nil)
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// info handles the /info command to display detailed information
// about a user or channel including ID, name, and special roles.
func (moduleStruct) info(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	userId := extraction.ExtractUser(b, ctx)
	switch userId {
	case -1:
		return ext.EndGroups
	case 0:
		// 0 id is for self
		userId = sender.Id()
	}

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "info") {
		return ext.EndGroups
	}

	username, name, found := extraction.GetUserInfo(userId)
	var text string

	if !found {
		text = "Could not find the any information about this user."
	} else {

		user := &gotgbot.User{
			Id:        userId,
			Username:  username,
			FirstName: name,
		}

		// If channel then this info
		if strings.HasPrefix(fmt.Sprint(userId), "-100") {
			text = fmt.Sprintf(
				"<b>Channel Info:</b>"+
					"\nID: <code>%d</code>"+
					"\nChannel Name: %s", userId, html.EscapeString(user.FirstName),
			)

			if user.Username != "" {
				text += fmt.Sprintf("\nUsername: @%s", user.Username)
				text += fmt.Sprintf("\nChannel link: @%s", user.Username)
			}
		} else {
			text = fmt.Sprintf(
				"<b>User Info:</b>"+
					"\nID: <code>%d</code>"+
					"\nName: %s", userId, html.EscapeString(user.FirstName),
			)
			if user.Username != "" {
				text += fmt.Sprintf("\nUsername: @%s", user.Username)
			}
			text += fmt.Sprintf("\nUser link: %s", helpers.MentionHtml(user.Id, "link"))
			if user.Id == config.OwnerId {
				text += "\nHe is my owner!"
			}
			if db.GetTeamMemInfo(user.Id).Dev {
				text += "\nHe is one of my dev users!"
			}
		}
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// translate handles the /tr command to translate text using
// Google Translate API with automatic language detection.
func (moduleStruct) translate(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "tr") {
		return ext.EndGroups
	}

	var (
		origText string
		toLang   string
	)

	if len(args) == 0 && msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, "I need some text and a language code to translate.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if reply := msg.ReplyToMessage; reply != nil {
		if reply.Text != "" {
			origText = reply.Text
		} else if reply.Caption != "" {
			origText = reply.Caption
		} else {
			_, _ = msg.Reply(b, "The replied message does not contain any text to translate.", helpers.Shtml())
			return ext.EndGroups
		}
		if len(args) == 0 {
			toLang = "en"
		} else {
			toLang = args[0]
		}
	} else {
		// args[1:] leaves the language code and takes rest of the text
		if len(args[1:]) < 1 {
			_, _ = msg.Reply(b, "Please provide some text to translate.", helpers.Shtml())
			return ext.EndGroups
		}
		// args[0] is the language code
		toLang = args[0]
		origText = strings.Join(args[1:], " ")
	}
	req, err := httpClient.Get(fmt.Sprintf("https://clients5.google.com/translate_a/t?client=dict-chrome-ex&sl=auto&tl=%s&q=%s", toLang, url.QueryEscape(strings.TrimSpace(origText))))
	if err != nil {
		_, _ = msg.Reply(b, "Error making a translation request!", nil)
		return ext.EndGroups
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(req.Body)
	all, err := io.ReadAll(req.Body)
	if err != nil {
		_, _ = msg.Reply(b, "Reading Error: "+err.Error(), nil)
		return ext.EndGroups
	}
	data := strings.Split(strings.Trim(string(all), `"][`), `","`)
	_, _ = msg.Reply(b,
		fmt.Sprintf("<b>Detected Language:</b> <code>%s</code>\n<b>Translation:</b> <code>%s</code>", data[1], data[0]),
		helpers.Shtml(),
	)
	return ext.EndGroups
}

// removeBotKeyboard handles the /removebotkeyboard command to
// remove stuck bot keyboards from the chat interface.
func (moduleStruct) removeBotKeyboard(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	rMsg, err := msg.Reply(b,
		"Removing the stuck bot keyboard...",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: &gotgbot.ReplyKeyboardRemove{
				RemoveKeyboard: true,
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	time.Sleep(1 * time.Second)
	_, err = rMsg.Delete(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// stat handles the /stat command to display the total number
// of messages in the current group chat.
func (moduleStruct) stat(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	if !chat_status.RequireGroup(b, ctx, chat, false) {
		return ext.EndGroups
	}
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "stat") {
		return ext.EndGroups
	}
	_, err := msg.Reply(b, fmt.Sprintf("Total Messages in %s are: %d", msg.Chat.Title, msg.MessageId+1), nil)
	if err != nil {
		log.Error(err)
	}
	return ext.EndGroups
}

// LoadMisc registers all miscellaneous module handlers with the dispatcher,
// including utility commands for IDs, ping, translation, and stats.
func LoadMisc(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(miscModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("stat", miscModule.stat))
	misc.AddCmdToDisableable("stat")
	dispatcher.AddHandler(handlers.NewCommand("id", miscModule.getId))
	misc.AddCmdToDisableable("id")
	dispatcher.AddHandler(handlers.NewCommand("tell", miscModule.echomsg))
	dispatcher.AddHandler(handlers.NewCommand("ping", miscModule.ping))
	misc.AddCmdToDisableable("ping")
	dispatcher.AddHandler(handlers.NewCommand("info", miscModule.info))
	misc.AddCmdToDisableable("info")
	dispatcher.AddHandler(handlers.NewCommand("tr", miscModule.translate))
	misc.AddCmdToDisableable("tr")
	dispatcher.AddHandler(handlers.NewCommand("removebotkeyboard", miscModule.removeBotKeyboard))
}
