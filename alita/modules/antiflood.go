package modules

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

type antifloodStruct struct {
	moduleStruct  // inheritance
	syncHelperMap sync.Map
	// Add semaphore to limit concurrent admin checks
	adminCheckSemaphore chan struct{}
}

type floodControl struct {
	userId       int64
	messageCount int
	messageIDs   []int64
	lastActivity int64 // Unix timestamp for cleanup
}

var _normalAntifloodModule = moduleStruct{
	moduleName:   "Antiflood",
	handlerGroup: 4,
}

var antifloodModule = antifloodStruct{
	moduleStruct:        _normalAntifloodModule,
	syncHelperMap:       sync.Map{},
	adminCheckSemaphore: make(chan struct{}, 50), // Limit to 50 concurrent admin checks
}

// init starts cleanup goroutine for antiflood cache
func init() {
	go antifloodModule.cleanupLoop()
}

// cleanupLoop periodically cleans up old entries from the flood cache
// cleanupLoop periodically removes old flood control entries from memory.
// Runs every 5 minutes to clean entries older than 10 minutes.
func (a *antifloodStruct) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		currentTime := time.Now().Unix()
		a.syncHelperMap.Range(func(key, value interface{}) bool {
			if floodData, ok := value.(floodControl); ok {
				// Remove entries older than 10 minutes
				if currentTime-floodData.lastActivity > 600 {
					a.syncHelperMap.Delete(key)
				}
			}
			return true
		})
	}
}

// updateFlood tracks message counts per user and determines if flood limit exceeded.
// Returns true if user has exceeded flood limit and should be restricted.
func (*moduleStruct) updateFlood(chatId, userId, msgId int64) (returnVar bool, floodCrc floodControl) {
	floodSrc := db.GetFlood(chatId)

	if floodSrc.Limit != 0 {
		currentTime := time.Now().Unix()

		// Key by composite (chatId, userId)
		key := fmt.Sprintf("%d:%d", chatId, userId)
		tmpInterface, valExists := antifloodModule.syncHelperMap.Load(key)
		if valExists && tmpInterface != nil {
			floodCrc = tmpInterface.(floodControl)

			// Clean up old entries (older than 1 minute)
			if currentTime-floodCrc.lastActivity > 60 {
				floodCrc = floodControl{}
			}
		}

		// No need to check userId mismatch since key includes userId
		if floodCrc.userId == 0 {
			floodCrc.userId = userId
			floodCrc.messageCount = 0
			floodCrc.messageIDs = make([]int64, 0, floodSrc.Limit+5) // Pre-allocate with capacity
		}

		floodCrc.messageCount++
		floodCrc.lastActivity = currentTime

		// Use efficient prepend with pre-allocated slice
		if len(floodCrc.messageIDs) >= cap(floodCrc.messageIDs) {
			// Resize if needed, keep only recent messages
			newIDs := make([]int64, 1, floodSrc.Limit+5)
			newIDs[0] = msgId
			if len(floodCrc.messageIDs) > 0 {
				copyCount := floodSrc.Limit
				if copyCount > len(floodCrc.messageIDs) {
					copyCount = len(floodCrc.messageIDs)
				}
				newIDs = append(newIDs, floodCrc.messageIDs[:copyCount]...)
			}
			floodCrc.messageIDs = newIDs
		} else {
			floodCrc.messageIDs = append([]int64{msgId}, floodCrc.messageIDs...)
		}

		if floodCrc.messageCount > floodSrc.Limit {
			antifloodModule.syncHelperMap.Store(key,
				floodControl{
					userId:       0,
					messageCount: 0,
					messageIDs:   make([]int64, 0),
					lastActivity: currentTime,
				},
			)
			returnVar = true
		} else {
			antifloodModule.syncHelperMap.Store(key, floodCrc)
		}
	}

	return
}

// checkFlood monitors incoming messages for flood violations.
// Applies configured flood actions (mute/kick/ban) when limits are exceeded.
func (m *moduleStruct) checkFlood(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender
	if user.IsAnonymousAdmin() {
		return ext.ContinueGroups
	}
	msg := ctx.EffectiveMessage
	if msg.MediaGroupId != "" {
		return ext.ContinueGroups
	}

	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	var (
		fmode    string
		keyboard [][]gotgbot.InlineKeyboardButton
	)
	userId := user.Id()

	// Use semaphore to limit concurrent admin checks and add timeout
	select {
	case antifloodModule.adminCheckSemaphore <- struct{}{}:
		// Got semaphore, proceed with admin check
		defer func() { <-antifloodModule.adminCheckSemaphore }()

		// Create context with timeout for admin check
		ctx_timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Check if user is admin with timeout and proper goroutine cleanup
		isAdmin := make(chan bool, 1)
		done := make(chan struct{})

		go func() {
			defer func() {
				close(done) // Signal completion to prevent goroutine leak
				if r := recover(); r != nil {
					log.WithFields(log.Fields{
						"chatId": chat.Id,
						"userId": userId,
						"panic":  r,
					}).Error("Panic in admin check goroutine")
				}
			}()

			select {
			case isAdmin <- chat_status.IsUserAdmin(b, chat.Id, userId):
				// Successfully sent result
			case <-ctx_timeout.Done():
				// Context cancelled, exit goroutine early
				return
			}
		}()

		select {
		case admin := <-isAdmin:
			if admin {
				m.updateFlood(chat.Id, 0, 0) // empty message queue when admin sends a message
				return ext.ContinueGroups
			}
		case <-ctx_timeout.Done():
			// Admin check timed out, treat as non-admin for safety
			log.WithFields(log.Fields{
				"chatId": chat.Id,
				"userId": userId,
			}).Warn("Admin check timed out, treating user as non-admin")
		}

		// Wait for goroutine cleanup with timeout to prevent indefinite blocking
		select {
		case <-done:
			// Goroutine completed cleanly
		case <-time.After(1 * time.Second):
			// Log if goroutine takes too long to cleanup
			log.WithFields(log.Fields{
				"chatId": chat.Id,
				"userId": userId,
			}).Warn("Admin check goroutine cleanup timeout")
		}
	default:
		// Semaphore full, skip admin check for performance (treat as non-admin)
		log.WithFields(log.Fields{
			"chatId": chat.Id,
			"userId": userId,
		}).Debug("Admin check semaphore full, skipping admin check")
	}

	// update flood for user
	flooded, floodCrc := m.updateFlood(chat.Id, userId, msg.MessageId)
	if !flooded {
		return ext.ContinueGroups
	}

	flood := db.GetFlood(chat.Id)

	if flood.DeleteAntifloodMessage {
		for _, i := range floodCrc.messageIDs {
			_, err := b.DeleteMessage(chat.Id, i, nil)
			// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
			// 	log.WithFields(
			// 		log.Fields{
			// 			"chat":    chat.Id,
			// 			"message": i,
			// 		},
			// 	).Error("error deleting message")
			// 	return ext.EndGroups
			// } else
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		_, err := msg.Delete(b, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	switch flood.Mode {
	case "mute":
		// don't work on anonymous channels
		if user.IsAnonymousChannel() {
			return ext.ContinueGroups
		}
		fmode = "muted"
		keyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Unmute (Admins Only)",
					CallbackData: fmt.Sprintf("unrestrict.unmute.%d", user.Id()),
				},
			},
		}

		_, err := chat.RestrictMember(b, userId,
			gotgbot.ChatPermissions{
				CanSendMessages:       false,
				CanSendPhotos:         false,
				CanSendVideos:         false,
				CanSendAudios:         false,
				CanSendDocuments:      false,
				CanSendVideoNotes:     false,
				CanSendVoiceNotes:     false,
				CanAddWebPagePreviews: false,
				CanChangeInfo:         false,
				CanInviteUsers:        false,
				CanPinMessages:        false,
				CanManageTopics:       false,
				CanSendPolls:          false,
				CanSendOtherMessages:  false,
			},
			nil,
		)
		if err != nil {
			log.Errorf(" checkFlood: %d (%d) - %v", chat.Id, user.Id(), err)
			return err
		}
	case "kick":
		// don't work on anonymous channels
		if user.IsAnonymousChannel() {
			return ext.ContinueGroups
		}
		fmode = "kicked"
		keyboard = nil
		_, err := chat.BanMember(b, userId, nil)
		if err != nil {
			log.Errorf(" checkFlood: %d (%d) - %v", chat.Id, user.Id(), err)
			return err
		}
		// Use non-blocking delayed unban for kick action
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.WithField("panic", r).Error("Panic in antiflood delayed unban goroutine")
				}
			}()

			time.Sleep(3 * time.Second)
			_, unbanErr := chat.UnbanMember(b, userId, nil)
			if unbanErr != nil {
				log.WithFields(log.Fields{
					"chatId": chat.Id,
					"userId": userId,
					"error":  unbanErr,
				}).Error("Failed to unban user after antiflood kick")
			}
		}()
	case "ban":
		fmode = "banned"
		if !user.IsAnonymousChannel() {
			_, err := chat.BanMember(b, userId, nil)
			if err != nil {
				log.Errorf(" checkFlood: %d (%d) - %v", chat.Id, user.Id(), err)
				return err
			}
		} else {
			keyboard = [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "Unban (Admins Only)",
						CallbackData: fmt.Sprintf("unrestrict.unban.%d", user.Id()),
					},
				},
			}
			_, err := chat.BanSenderChat(b, userId, nil)
			if err != nil {
				log.Errorf(" checkFlood: %d (%d) - %v", chat.Id, user.Id(), err)
				return err
			}
		}
	}
	if _, err := b.SendMessage(chat.Id,
		tr.Message("antiflood_check_flood_perform_action", i18n.Params{
			"user": helpers.MentionHtml(userId, user.Name()),
			"action": fmode,
		}),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: keyboard,
			},
			MessageThreadId: msg.MessageThreadId,
		},
	); err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

// setFlood handles the /setflood command to configure flood detection limits.
// Sets the maximum number of messages allowed before triggering flood protection.
func (m *moduleStruct) setFlood(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	args := ctx.Args()[1:]

	var replyText string

	if len(args) == 0 {
		replyText = tr.Message("antiflood_errors_expected_args", nil)
	} else {
		if string_handling.FindInStringSlice([]string{"off", "no", "false", "0"}, strings.ToLower(args[0])) {
			replyText = tr.Message("antiflood_set_flood_disabled", nil)
			go db.SetFlood(chat.Id, 0)
		} else {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				replyText = tr.Message("antiflood_errors_invalid_int", nil)
			} else {
				if num < 3 || num > 100 {
					replyText = tr.Message("antiflood_errors_set_in_limit", nil)
				} else {
					go db.SetFlood(chat.Id, num)
					replyText = tr.Message("antiflood_set_flood_success", i18n.Params{
						"limit": num,
					})
				}
			}
		}
	}

	_, err := msg.Reply(b, replyText, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// flood handles the /flood command to display current flood protection settings.
// Shows the flood limit and action (mute/kick/ban) for the chat.
func (m *moduleStruct) flood(b *gotgbot.Bot, ctx *ext.Context) error {
	var text string
	msg := ctx.EffectiveMessage

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "flood") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, false, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat

	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	flood := db.GetFlood(chat.Id)
	if flood.Limit == 0 {
		text = tr.Message("antiflood_flood_disabled", nil)
	} else {
		var mode string
		switch flood.Mode {
		case "mute":
			mode = "muted"
		case "ban":
			mode = "banned"
		case "kick":
			mode = "kicked"
		}
		text = tr.Message("antiflood_flood_show_settings", i18n.Params{
			"limit": flood.Limit,
			"action": mode,
		})
	}
	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		return err
	}
	return ext.EndGroups
}

// setFloodMode handles the /setfloodmode command to configure flood protection actions.
// Allows setting the punishment type (ban/kick/mute) for flood violations.
func (m *moduleStruct) setFloodMode(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	args := ctx.Args()[1:]

	if len(args) > 0 {
		selectedMode := strings.ToLower(args[0])
		if string_handling.FindInStringSlice([]string{"ban", "kick", "mute"}, selectedMode) {
			_, err := msg.Reply(b, tr.Message("antiflood_set_flood_mode_success", i18n.Params{
				"mode": selectedMode,
			}), helpers.Shtml())
			if err != nil {
				log.Error(err)
			}
			go db.SetFloodMode(chat.Id, selectedMode)
			return ext.EndGroups
		} else {
			_, err := msg.Reply(b, tr.Message("antiflood_set_flood_mode_unknown_type", i18n.Params{
				"mode": args[0],
			}), helpers.Shtml())
			if err != nil {
				return err
			}
		}
	} else {
		text := tr.Message("antiflood_set_flood_mode_specify_action", nil)
		_, err := msg.Reply(b, text, helpers.Smarkdown())
		if err != nil {
			return err
		}
	}
	return ext.EndGroups
}

// setFloodDeleter handles the /delflood command to toggle message deletion on flood.
// Configures whether to delete all flood messages or just the triggering message.
func (m *moduleStruct) setFloodDeleter(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	args := ctx.Args()[1:]
	var text string

	if len(args) > 0 {
		selectedMode := strings.ToLower(args[0])
		switch selectedMode {
		case "on", "yes":
			go db.SetFloodMsgDel(chat.Id, true)
			text = tr.Message("antiflood_flood_deleter_enabled", nil)
		case "off", "no":
			go db.SetFloodMsgDel(chat.Id, false)
			text = tr.Message("antiflood_flood_deleter_disabled", nil)
		default:
			text = tr.Message("antiflood_flood_deleter_invalid_option", nil)
		}
	} else {
		currSet := db.GetFlood(chat.Id).DeleteAntifloodMessage
		if currSet {
			text = tr.Message("antiflood_flood_deleter_already_enabled", nil)
		} else {
			text = tr.Message("antiflood_flood_deleter_already_disabled", nil)
		}
	}
	_, err := msg.Reply(b, text, helpers.Smarkdown())
	if err != nil {
		return err
	}

	return ext.EndGroups
}

// LoadAntiflood registers all antiflood module handlers with the dispatcher.
// Sets up flood detection commands and message monitoring handlers.
func LoadAntiflood(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(antifloodModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("setflood", antifloodModule.setFlood))
	dispatcher.AddHandler(handlers.NewCommand("setfloodmode", antifloodModule.setFloodMode))
	dispatcher.AddHandler(handlers.NewCommand("delflood", antifloodModule.setFloodDeleter))
	dispatcher.AddHandler(handlers.NewCommand("flood", antifloodModule.flood))
	misc.AddCmdToDisableable("flood")
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, antifloodModule.checkFlood), antifloodModule.handlerGroup)
}
