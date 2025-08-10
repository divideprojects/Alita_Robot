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

var (
	miscModule = moduleStruct{moduleName: "Misc"}
	// HTTP client with timeout and connection pooling for external requests
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: true,
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
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		text, _ := tr.GetString("misc_reply_to_someone")
		_, _ = msg.Reply(b, text, nil)
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
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		text, _ := tr.GetString("misc_provide_content")
		_, _ = msg.Reply(b, text, nil)
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
			tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
			temp, _ := tr.GetString("misc_chat_id")
			text := fmt.Sprintf(temp, msg.Chat.Id)
			builder.WriteString(text + "\n")
			if msg.IsTopicMessage {
				temp2, _ := tr.GetString("misc_thread_id")
				text = fmt.Sprintf(temp2, msg.MessageThreadId)
				builder.WriteString(text + "\n")
			}
			if msg.ReplyToMessage.From != nil {
				originalId := msg.ReplyToMessage.From.Id
				_, user1Name, _ := extraction.GetUserInfo(originalId)
				temp3, _ := tr.GetString("misc_user_id")
				text = fmt.Sprintf(temp3, user1Name, originalId)
				builder.WriteString(text + "\n")
			}

			if rpm := msg.ReplyToMessage; rpm != nil {
				if frpm := rpm.ForwardOrigin; frpm != nil {
					if frpm.GetDate() != 0 {
						fwdd := frpm.MergeMessageOrigin()

						if fwdc := fwdd.SenderUser; fwdc != nil {
							user1Id := fwdc.Id
							_, user1Name, _ := extraction.GetUserInfo(user1Id)
							temp4, _ := tr.GetString("misc_forwarded_from_user")
							text = fmt.Sprintf(temp4, user1Name, user1Id)
							builder.WriteString(text + "\n")
						}

						if fwdc := fwdd.Chat; fwdc != nil {
							temp5, _ := tr.GetString("misc_forwarded_from_chat")
							text = fmt.Sprintf(temp5, fwdc.Title, fwdc.Id)
							builder.WriteString(text + "\n")
						}
					}
				}
			}
			if msg.ReplyToMessage.Animation != nil {
				temp6, _ := tr.GetString("misc_gif_id")
				text = fmt.Sprintf(temp6, msg.ReplyToMessage.Animation.FileId)
				builder.WriteString(text + "\n")
			}
			if msg.ReplyToMessage.Sticker != nil {
				temp7, _ := tr.GetString("misc_sticker_id")
				text = fmt.Sprintf(temp7, msg.ReplyToMessage.Sticker.FileId)
				builder.WriteString(text + "\n")
			}
		} else {
			_, name, _ := extraction.GetUserInfo(userId)
			tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
			temp, _ := tr.GetString("misc_user_id_is")
			text := fmt.Sprintf(temp, name, userId)
			builder.WriteString(text)
		}
	} else {
		chat := ctx.EffectiveChat
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		if ctx.Message.Chat.Type == "private" {
			temp, _ := tr.GetString("misc_your_id_private")
			text := fmt.Sprintf(temp, chat.Id)
			builder.WriteString(text)
		} else {
			temp, _ := tr.GetString("misc_your_id_group")
			text := fmt.Sprintf(temp, msg.From.Id, chat.Id)
			builder.WriteString(text)
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
// detailed latency breakdown including user-to-Telegram, processing, and API response times.
// Advanced implementation with location estimation and comprehensive metrics.
func (moduleStruct) ping(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	webhookReceivedTime := time.Now()

	// When user sent the message (Unix timestamp from Telegram)
	userSentTime := time.Unix(int64(msg.Date), 0)

	// Calculate incoming latency (user -> Telegram -> webhook)
	incomingLatency := webhookReceivedTime.Sub(userSentTime)

	// Prefetch for accurate processing measurement
	prefetchStart := time.Now()
	prefetched, err := db.PrefetchCommandContext(ctx)
	prefetchTime := time.Since(prefetchStart)

	if err != nil {
		// Fallback without timing details
		_, _ = msg.Reply(b, "🏓 Pong!", nil)
		return ext.EndGroups
	}

	// Check if command is disabled
	if msg.Chat.Type != "private" && prefetched.IsCommandDisabled("ping") {
		return ext.EndGroups
	}

	// Total processing time before API call
	processingTime := time.Since(webhookReceivedTime)

	// Send initial message and measure API call time
	apiStart := time.Now()
	rmsg, err := msg.Reply(b, "🏓 Calculating latency...", &gotgbot.SendMessageOpts{
		ParseMode: helpers.HTML,
	})
	if err != nil {
		return err
	}
	apiSendTime := time.Since(apiStart)

	// Edit message to measure round-trip
	editStart := time.Now()

	// In webhook mode, we can calculate accurate components:
	// 1. User -> Telegram: Unknown, but we can estimate from total
	// 2. Telegram -> Webhook: Very fast (direct push)
	// 3. Bot processing: What we measured
	// 4. Bot -> Telegram API: What we measured

	// Estimate user's latency to Telegram
	// In webhook mode, Telegram -> Webhook is typically 5-10ms
	telegramToWebhook := 10 * time.Millisecond
	userToTelegram := incomingLatency - telegramToWebhook
	if userToTelegram < 0 {
		userToTelegram = incomingLatency // Fallback if estimate is off
	}

	// Calculate total round-trip estimate
	totalRoundTrip := userToTelegram*2 + processingTime + apiSendTime

	// Determine user's approximate location based on latency
	userLocation := getLocationFromLatency(userToTelegram)

	// Build detailed response
	text := fmt.Sprintf(
		"🏓 <b>Pong!</b>\n\n"+
			"📊 <b>Latency Breakdown:</b>\n"+
			"├─ 📤 <b>Incoming Path:</b>\n"+
			"│  ├ 👤 You → Telegram: ~%dms %s\n"+
			"│  ├ ⚡ Telegram → Bot: ~%dms\n"+
			"│  └ 📥 Total incoming: %dms\n"+
			"│\n"+
			"├─ ⚙️ <b>Bot Processing:</b>\n"+
			"│  ├ 🗄️ Database: %dms\n"+
			"│  └ 🤖 Total: %dms\n"+
			"│\n"+
			"├─ 📡 <b>Bot → Telegram:</b> %dms\n"+
			"│\n"+
			"└─ ⏱️ <b>Est. Round-trip:</b> %dms\n\n"+
			"<i>Webhook mode • Server: Zurich</i>",
		userToTelegram.Milliseconds(),
		userLocation,
		telegramToWebhook.Milliseconds(),
		incomingLatency.Milliseconds(),
		prefetchTime.Milliseconds(),
		processingTime.Milliseconds(),
		apiSendTime.Milliseconds(),
		totalRoundTrip.Milliseconds(),
	)

	_, _, err = rmsg.EditText(b, text, &gotgbot.EditMessageTextOpts{
		ParseMode: helpers.HTML,
	})
	editTime := time.Since(editStart)

	if err != nil {
		log.WithError(err).Error("[Ping] Failed to edit message with latency details")
		return err
	}

	// Log detailed metrics for debugging
	log.WithFields(log.Fields{
		"user_id":          msg.From.Id,
		"user_to_telegram": userToTelegram.Milliseconds(),
		"incoming_total":   incomingLatency.Milliseconds(),
		"processing":       processingTime.Milliseconds(),
		"api_send":         apiSendTime.Milliseconds(),
		"api_edit":         editTime.Milliseconds(),
		"total_round_trip": totalRoundTrip.Milliseconds(),
	}).Debug("[Ping] Detailed latency metrics")

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
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		text, _ = tr.GetString("misc_user_not_found")
	} else {

		user := &gotgbot.User{
			Id:        userId,
			Username:  username,
			FirstName: name,
		}

		// If channel then this info
		if helpers.IsChannelID(userId) {
			tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
			textTemplate, _ := tr.GetString("misc_channel_info_header")
			text = fmt.Sprintf(textTemplate, userId, html.EscapeString(user.FirstName))

			if user.Username != "" {
				usernameTemplate, _ := tr.GetString("misc_username")
				text += fmt.Sprintf("\n"+usernameTemplate, user.Username)
				linkTemplate, _ := tr.GetString("misc_channel_link")
				text += fmt.Sprintf("\n"+linkTemplate, user.Username)
			}
		} else {
			tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
			textTemplate, _ := tr.GetString("misc_user_info_header")
			text = fmt.Sprintf(textTemplate, userId, html.EscapeString(user.FirstName))
			if user.Username != "" {
				usernameTemplate, _ := tr.GetString("misc_username")
				text += fmt.Sprintf("\n"+usernameTemplate, user.Username)
			}
			linkTemplate, _ := tr.GetString("misc_user_link")
			text += fmt.Sprintf("\n"+linkTemplate, helpers.MentionHtml(user.Id, "link"))
			if user.Id == config.OwnerId {
				ownerText, _ := tr.GetString("misc_owner_info")
				text += "\n" + ownerText
			}
			if db.GetTeamMemInfo(user.Id).Dev {
				devText, _ := tr.GetString("misc_dev_info")
				text += "\n" + devText
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
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		text, _ := tr.GetString("misc_translate_need_text")
		_, err := msg.Reply(b, text, helpers.Shtml())
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
			tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
			text, _ := tr.GetString("misc_translate_no_text")
			_, _ = msg.Reply(b, text, helpers.Shtml())
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
			tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
			text, _ := tr.GetString("misc_translate_provide_text")
			_, _ = msg.Reply(b, text, helpers.Shtml())
			return ext.EndGroups
		}
		// args[0] is the language code
		toLang = args[0]
		origText = strings.Join(args[1:], " ")
	}
	req, err := httpClient.Get(fmt.Sprintf("https://clients5.google.com/translate_a/t?client=dict-chrome-ex&sl=auto&tl=%s&q=%s", toLang, url.QueryEscape(strings.TrimSpace(origText))))
	if err != nil {
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		text, _ := tr.GetString("misc_translation_error")
		_, _ = msg.Reply(b, text, nil)
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
		tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
		text, _ := tr.GetString("misc_translate_read_error")
		_, _ = msg.Reply(b, text+": "+err.Error(), nil)
		return ext.EndGroups
	}
	data := strings.Split(strings.Trim(string(all), `"][`), `","`)
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	textTemplate, _ := tr.GetString("misc_translate_result")
	text := fmt.Sprintf(textTemplate, data[1], data[0])
	_, _ = msg.Reply(b, text, helpers.Shtml())
	return ext.EndGroups
}

// removeBotKeyboard handles the /removebotkeyboard command to
// remove stuck bot keyboards from the chat interface.
func (moduleStruct) removeBotKeyboard(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	text, _ := tr.GetString("misc_removing_keyboard")
	rMsg, err := msg.Reply(b,
		text,
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
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	textTemplate, _ := tr.GetString("misc_total_messages")
	text := fmt.Sprintf(textTemplate, msg.Chat.Title, msg.MessageId+1)
	_, err := msg.Reply(b, text, nil)
	if err != nil {
		log.Error(err)
	}
	return ext.EndGroups
}

// getLocationFromLatency estimates user location from latency to Telegram servers
func getLocationFromLatency(latency time.Duration) string {
	ms := latency.Milliseconds()

	// Telegram servers are primarily in:
	// - Amsterdam (Europe)
	// - Singapore (Asia)
	// - Miami (Americas)

	switch {
	case ms < 20:
		return "🇪🇺" // Very close to EU servers
	case ms < 40:
		return "🇪🇺/🇬🇧" // Europe
	case ms < 60:
		return "🇹🇷/🇦🇪" // Middle East
	case ms < 80:
		return "🇺🇸" // Eastern US
	case ms < 100:
		return "🇮🇳/🇷🇺" // South Asia/Russia
	case ms < 120:
		return "🇯🇵/🇰🇷" // East Asia
	case ms < 150:
		return "🇧🇷/🇦🇺" // South America/Oceania
	default:
		return "🌍" // Far region
	}
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
