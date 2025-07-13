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
	"github.com/divideprojects/Alita_Robot/alita/i18n"

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

// miscModule holds the configuration for the misc module
var miscModule = moduleStruct{
	moduleName: autoModuleName(),
	cfg:        nil, // will be set during LoadMisc
}

func (moduleStruct) echomsg(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
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
		needTargetMsg, needTargetErr := tr.GetStringWithError("strings.misc.reply.need_target")
		if needTargetErr != nil {
			log.Errorf("[misc] missing translation for Misc.reply.need_target: %v", needTargetErr)
			needTargetMsg = "Reply to a message to echo it."
		}
		_, _ = msg.Reply(b, needTargetMsg, nil)
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
		needContentMsg, needContentErr := tr.GetStringWithError("strings.misc.reply.need_content")
		if needContentErr != nil {
			log.Errorf("[misc] missing translation for Misc.reply.need_content: %v", needContentErr)
			needContentMsg = "Provide some content to echo."
		}
		_, _ = msg.Reply(b, needContentMsg, nil)
	}

	return ext.EndGroups
}

func (moduleStruct) getId(b *gotgbot.Bot, ctx *ext.Context) error {
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

			if rpm := msg.ReplyToMessage; rpm != nil {
				if frpm := rpm.ForwardOrigin; frpm != nil {
					if frpm.GetDate() != 0 {
						fwdd := frpm.MergeMessageOrigin()

						if fwdc := fwdd.SenderUser; fwdc != nil {
							user1Id := fwdc.Id
							_, user1Name, _ := extraction.GetUserInfo(user1Id)
							replyText += fmt.Sprintf(
								"<b>Forwarded from %s's ID:</b> <code>%d</code>\n",
								user1Name, user1Id,
							)
						}

						if fwdc := fwdd.Chat; fwdc != nil {
							replyText += fmt.Sprintf("<b>Forwarded from chat %s's ID:</b> <code>%d</code>\n",
								fwdc.Title, fwdc.Id,
							)
						}
					}
				}
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

func (m moduleStruct) paste(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
	msg := ctx.EffectiveMessage
	args := ctx.Args()

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "paste") {
		return ext.EndGroups
	}

	processingMsg, processingErr := tr.GetStringWithError("strings.misc.paste.processing")
	if processingErr != nil {
		log.Errorf("[misc] missing translation for Misc.paste.processing: %v", processingErr)
		processingMsg = "Processing..."
	}
	edited, err := msg.Reply(b, processingMsg, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	var (
		text      string
		extention string
	)

	if len(args) >= 2 {
		text = strings.Join(args[1:], " ")
		extention = "txt"
	} else if msg.ReplyToMessage != nil {
		if msg.ReplyToMessage.Text != "" {
			text = msg.ReplyToMessage.Text
			extention = "txt"
		} else if msg.ReplyToMessage.Caption != "" {
			text = msg.ReplyToMessage.Caption
			extention = "txt"
		} else if msg.ReplyToMessage.Document != nil {
			f, err := b.GetFile(msg.ReplyToMessage.Document.FileId, nil)
			if err != nil {
				log.Error(err)
				return err
			}
			if f.FileSize > 600000 {
				fileTooBigMsg, fileTooBigErr := tr.GetStringWithError("strings.misc.paste.file_too_big")
				if fileTooBigErr != nil {
					log.Errorf("[misc] missing translation for Misc.paste.file_too_big: %v", fileTooBigErr)
					fileTooBigMsg = "File is too big to paste."
				}
				_, _, _ = edited.EditText(b, fileTooBigMsg, nil)
				return ext.EndGroups
			}
			fileName := fmt.Sprintf("paste_%d_%d.txt", msg.Chat.Id, msg.MessageId)
			cfg := m.cfg
			raw, err := http.Get(cfg.ApiServer + "/file/bot" + cfg.BotToken + "/" + f.FilePath)
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
	}
	pasted, key := helpers.PasteToNekoBin(text)

	if pasted {
		_, _, err = edited.EditText(b, fmt.Sprintf("<b>Pasted Successfully!</b>\nhttps://www.nekobin.com/%s.%s", key, extention),
			&gotgbot.EditMessageTextOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
			},
		)
		if err != nil {
			log.Error(err)
		}
	} else {
		pasteErrorMsg, pasteErrorErr := tr.GetStringWithError("strings.misc.paste.paste_error")
		if pasteErrorErr != nil {
			log.Errorf("[misc] missing translation for Misc.paste.paste_error: %v", pasteErrorErr)
			pasteErrorMsg = "Failed to paste content."
		}
		_, _, err = edited.EditText(b, pasteErrorMsg, nil)
		if err != nil {
			log.Error(err)
		}
	}
	return ext.EndGroups
}

func (moduleStruct) ping(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
	msg := ctx.EffectiveMessage
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "ping") {
		return ext.EndGroups
	}
	stime := time.Now()
	rmsg, _ := msg.Reply(b, "<code>Pinging</code>", &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeHTML})
	pingedMsg, pingedErr := tr.GetStringWithError("strings.misc.pinged_in_percent_ms")
	if pingedErr != nil {
		log.Errorf("[misc] missing translation for Misc.pinged_in_percent_ms: %v", pingedErr)
		pingedMsg = "Pinged in %d ms"
	}
	_, _, err := rmsg.EditText(b, fmt.Sprintf(pingedMsg, int64(time.Since(stime)/time.Millisecond)), nil)
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

func (m moduleStruct) info(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
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
		userNotFoundMsg, userNotFoundErr := tr.GetStringWithError("strings.misc.info.user_not_found")
		if userNotFoundErr != nil {
			log.Errorf("[misc] missing translation for Misc.info.user_not_found: %v", userNotFoundErr)
			userNotFoundMsg = "User not found."
		}
		text = userNotFoundMsg
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
			cfg := m.cfg
			if user.Id == cfg.OwnerId {
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
	tr := i18n.New(db.GetLanguage(ctx))
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
		needTextAndLangMsg, needTextAndLangErr := tr.GetStringWithError("strings.misc.translate.need_text_and_lang")
		if needTextAndLangErr != nil {
			log.Errorf("[misc] missing translation for Misc.translate.need_text_and_lang: %v", needTextAndLangErr)
			needTextAndLangMsg = "Provide text and language to translate."
		}
		_, err := msg.Reply(b, needTextAndLangMsg, helpers.Shtml())
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
			noTextInReplyMsg, noTextInReplyErr := tr.GetStringWithError("strings.misc.translate.no_text_in_reply")
			if noTextInReplyErr != nil {
				log.Errorf("[misc] missing translation for Misc.translate.no_text_in_reply: %v", noTextInReplyErr)
				noTextInReplyMsg = "No text found in the replied message."
			}
			_, _ = msg.Reply(b, noTextInReplyMsg, helpers.Shtml())
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
			needTextMsg, needTextErr := tr.GetStringWithError("strings.misc.translate.need_text")
			if needTextErr != nil {
				log.Errorf("[misc] missing translation for Misc.translate.need_text: %v", needTextErr)
				needTextMsg = "Provide text to translate."
			}
			_, _ = msg.Reply(b, needTextMsg, helpers.Shtml())
			return ext.EndGroups
		}
		// args[0] is the language code
		toLang = args[0]
		origText = strings.Join(args[1:], " ")
	}
	req, err := http.Get(fmt.Sprintf("https://clients5.google.com/translate_a/t?client=dict-chrome-ex&sl=auto&tl=%s&q=%s", toLang, url.QueryEscape(strings.TrimSpace(origText))))
	if err != nil {
		requestErrorMsg, requestErrorErr := tr.GetStringWithError("strings.misc.translate.request_error")
		if requestErrorErr != nil {
			log.Errorf("[misc] missing translation for Misc.translate.request_error: %v", requestErrorErr)
			requestErrorMsg = "Failed to make translation request."
		}
		_, _ = msg.Reply(b, requestErrorMsg, nil)
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
		readingErrorMsg, readingErrorErr := tr.GetStringWithError("strings.misc.reading_error")
		if readingErrorErr != nil {
			log.Errorf("[misc] missing translation for Misc.reading_error: %v", readingErrorErr)
			readingErrorMsg = "Error reading response:"
		}
		_, _ = msg.Reply(b, readingErrorMsg+" "+err.Error(), nil)
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
	tr := i18n.New(db.GetLanguage(ctx))
	msg := ctx.EffectiveMessage

	removingKeyboardMsg, removingKeyboardErr := tr.GetStringWithError("strings.misc.keyboard.removing")
	if removingKeyboardErr != nil {
		log.Errorf("[misc] missing translation for Misc.keyboard.removing: %v", removingKeyboardErr)
		removingKeyboardMsg = "Removing keyboard..."
	}

	rMsg, err := msg.Reply(b,
		removingKeyboardMsg,
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
	tr := i18n.New(db.GetLanguage(ctx))
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	if !chat_status.RequireGroup(b, ctx, chat, false) {
		return ext.EndGroups
	}
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "stat") {
		return ext.EndGroups
	}

	totalMessagesMsg, totalMessagesErr := tr.GetStringWithError("strings.misc.total_messages_in_percent_are_percent")
	if totalMessagesErr != nil {
		log.Errorf("[misc] missing translation for Misc.total_messages_in_percent_are_percent: %v", totalMessagesErr)
		totalMessagesMsg = "Total messages in %s are %d"
	}

	_, err := msg.Reply(b, fmt.Sprintf(totalMessagesMsg, msg.Chat.Title, msg.MessageId+1), nil)
	if err != nil {
		log.Error(err)
	}
	return ext.EndGroups
}

func LoadMisc(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	miscModule.cfg = cfg

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
