package modules

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
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

var miscModule = moduleStruct{moduleName: "Misc"}

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
				ReplyToMessageId: replyMsg.MessageId,
				ParseMode:        helpers.Shtml().ParseMode,
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

func (moduleStruct) getId(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	}
	var replyText string

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "id") {
		return ext.EndGroups
	}

	if userId != 0 {
		if msg.ReplyToMessage != nil {
			replyText = fmt.Sprintf(
				"<b>Chat ID:</b> <code>%d</code>\n",
				msg.Chat.Id,
			)
			if msg.IsTopicMessage {
				replyText += fmt.Sprintf("Thread Id: <code>%d</code>\n", msg.MessageThreadId)
			}
			if msg.ReplyToMessage.From != nil {
				originalId := msg.ReplyToMessage.From.Id
				_, user1Name, _ := extraction.GetUserInfo(originalId)
				replyText += fmt.Sprintf(
					"<b>%s's ID:</b> <code>%d</code>\n",
					user1Name,
					originalId,
				)
			}
			if msg.ReplyToMessage.ForwardFrom != nil {
				user1Id := msg.ReplyToMessage.ForwardFrom.Id
				_, user1Name, _ := extraction.GetUserInfo(user1Id)
				replyText += fmt.Sprintf(
					"<b>Forwarded from %s's ID:</b> <code>%d</code>\n",
					user1Name, user1Id,
				)
			}
			if msg.ReplyToMessage.ForwardFromChat != nil {
				replyText += fmt.Sprintf("<b>Forwarded from chat %s's ID:</b> <code>%d</code>\n",
					msg.ReplyToMessage.ForwardFromChat.Title, msg.ReplyToMessage.ForwardFromChat.Id,
				)
			}
			if msg.ReplyToMessage.Animation != nil {
				replyText += fmt.Sprintf("<b>GIF ID:</b> <code>%s</code>\n",
					msg.ReplyToMessage.Animation.FileId,
				)
			}
			if msg.ReplyToMessage.Sticker != nil {
				replyText += fmt.Sprintf("<b>Sticker ID:</b> <code>%s</code>\n",
					msg.ReplyToMessage.Sticker.FileId,
				)
			}
		} else {
			_, name, _ := extraction.GetUserInfo(userId)
			replyText = fmt.Sprintf("%s's ID is <code>%d</code>", name, userId)
		}
	} else {
		chat := ctx.EffectiveChat
		if ctx.Update.Message.Chat.Type == "private" {
			replyText = fmt.Sprintf("Your ID is <code>%d</code>", chat.Id)
		} else {
			replyText = fmt.Sprintf("Your ID is <code>%d</code>\nThis group's ID is <code>%d</code>",
				msg.From.Id, chat.Id,
			)
		}
	}

	_, err := msg.Reply(b,
		replyText,
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func (moduleStruct) paste(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "paste") {
		return ext.EndGroups
	}

	var (
		err  error
		text string
	)

	if len(args) == 1 && msg.ReplyToMessage == nil {
		_, err = msg.Reply(b, "Please give text to paste or reply to a document!", nil)
		if err != nil {
			log.Error(err)
		}
		return ext.EndGroups
	}
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.Text == "" && msg.ReplyToMessage.Document == nil && msg.ReplyToMessage.Caption == "" {
		_, err = msg.Reply(b, "Please give text to paste or reply to a document!", nil)
		if err != nil {
			log.Error(err)
		}
		return ext.EndGroups
	}

	edited, _ := msg.Reply(b, "Pasting ...", nil)
	extention := "txt"
	if len(args) >= 2 {
		text = strings.Join(args[1:], " ")
	} else if len(args) != 2 && msg.ReplyToMessage.Text != "" {
		text = msg.ReplyToMessage.Text
	} else if len(args) != 2 && msg.ReplyToMessage.Caption != "" && msg.ReplyToMessage.Document == nil {
		text = msg.ReplyToMessage.Caption
	} else if msg.ReplyToMessage.Document != nil {
		if strings.Contains(msg.ReplyToMessage.Document.FileName, ".") {
			extention = strings.SplitN(msg.ReplyToMessage.Document.FileName, ".", 2)[1]
		}
		f, err := b.GetFile(msg.ReplyToMessage.Document.FileId, nil)
		if err != nil {
			_, _, _ = edited.EditText(b, "BadRequest on GetFile!", nil)
			return ext.EndGroups
		}
		if f.FileSize > 600000 {
			_, _, _ = edited.EditText(b, "File too big to paste; Max. file size that can be pasted is 600 kb!", nil)
			return ext.EndGroups
		}
		fileName := fmt.Sprintf("paste_%d_%d.txt", msg.Chat.Id, msg.MessageId)
		raw, err := http.Get(config.ApiServer + "/file/bot" + b.GetToken() + "/" + f.FilePath)
		if err != nil {
			log.Error(err)
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(raw.Body)
		out, err := os.Create(fileName)
		if err != nil {
			log.Error(err)
		}
		_, err = io.Copy(out, raw.Body)
		if err != nil {
			log.Error(err)
			err = os.Remove(fileName)
			if err != nil {
				log.Error(err)
			}
			return ext.EndGroups
		}
		data, er := os.ReadFile(fileName)
		if er != nil {
			log.Error(er)
			return ext.EndGroups
		}
		text = string(data)
		err = os.Remove(fileName)
		if err != nil {
			log.Error(err)
		}
	}
	pasted, key := helpers.PasteToNekoBin(text)

	if pasted {
		_, _, err = edited.EditText(b, fmt.Sprintf("<b>Pasted Successfully!</b>\nhttps://www.nekobin.com/%s.%s", key, extention),
			&gotgbot.EditMessageTextOpts{
				ParseMode:             helpers.HTML,
				DisableWebPagePreview: true,
			},
		)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, _, err = edited.EditText(b, "Can't paste the provided data!", nil)
		if err != nil {
			log.Error(err)
		}
	}
	return ext.EndGroups
}

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

func (moduleStruct) info(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	sender := ctx.EffectiveSender
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if userId == 0 {
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
	req, err := http.Get(fmt.Sprintf("https://clients5.google.com/translate_a/t?client=dict-chrome-ex&sl=auto&tl=%s&q=%s", toLang, url.QueryEscape(strings.TrimSpace(origText))))
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

// This function removes the stuck bot keyboard from your chat!
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

func LoadMisc(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(miscModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("stat", miscModule.stat))
	misc.AddCmdToDisableable("stat")
	dispatcher.AddHandler(handlers.NewCommand("paste", miscModule.paste))
	misc.AddCmdToDisableable("paste")
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
