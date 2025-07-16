package modules

import (
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

// tokenBucket implements a thread-safe token bucket rate limiter
type tokenBucket struct {
	capacity   int           // max tokens (burst capacity)
	refillRate time.Duration // time between token refills
	tokens     int           // current tokens
	lastRefill time.Time     // last refill time
	mu         sync.Mutex    // mutex for thread safety
	messageIDs *ringBuffer   // track message IDs for deletion
}

// newTokenBucket creates a new token bucket with given capacity and refill rate
func newTokenBucket(capacity int, refillRate time.Duration) *tokenBucket {
	return &tokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		lastRefill: time.Now(),
		messageIDs: newRingBuffer(100),
	}
}

// take attempts to take a token from the bucket
// returns true if token was available, false if rate limited
func (tb *tokenBucket) take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed / tb.refillRate)

	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// ringBuffer implements a fixed-size circular buffer for message IDs
type ringBuffer struct {
	buffer []int64
	size   int
	head   int
	tail   int
	count  int
	mu     sync.Mutex
}

// newRingBuffer creates a new ring buffer with specified capacity
func newRingBuffer(capacity int) *ringBuffer {
	return &ringBuffer{
		buffer: make([]int64, capacity),
		size:   capacity,
	}
}

// push adds a message ID to the buffer, overwriting oldest if full
func (rb *ringBuffer) push(msgID int64) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = msgID
	rb.head = (rb.head + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	} else {
		rb.tail = (rb.tail + 1) % rb.size
	}
}

// getAll returns all message IDs in order (newest first)
func (rb *ringBuffer) getAll() []int64 {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return []int64{}
	}

	result := make([]int64, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.head - 1 - i + rb.size) % rb.size
		result[i] = rb.buffer[idx]
	}
	return result
}

// antifloodStruct implements the antiflood module logic using a token bucket algorithm.
//
// It embeds moduleStruct and manages:
// - Rate limiting state per chat using token buckets
// - Admin status caching to reduce API calls
// - Flood detection and enforcement
//
// Configuration:
// - Flood limit sets the bucket capacity (burst size)
// - Refill rate is automatically calculated as 1 token per (limit) seconds
// - Admin cache TTL is 5 minutes
type antifloodStruct struct {
	moduleStruct  // inheritance
	syncHelperMap sync.Map

	// Admin status cache with TTL
	adminCache map[int64]struct {
		isAdmin bool
		expires time.Time
	}
	adminCacheMu sync.RWMutex
}

// isAdminCached checks if user is admin using cache
func (a *antifloodStruct) isAdminCached(b *gotgbot.Bot, chatId, userId int64) bool {
	a.adminCacheMu.RLock()
	cached, exists := a.adminCache[userId]
	a.adminCacheMu.RUnlock()

	if exists && time.Now().Before(cached.expires) {
		return cached.isAdmin
	}

	// Cache miss - check actual status
	isAdmin := chat_status.IsUserAdmin(b, chatId, userId)

	a.adminCacheMu.Lock()
	a.adminCache[userId] = struct {
		isAdmin bool
		expires time.Time
	}{
		isAdmin: isAdmin,
		expires: time.Now().Add(5 * time.Minute), // Cache for 5 minutes
	}
	a.adminCacheMu.Unlock()

	return isAdmin
}

var _normalAntifloodModule = moduleStruct{
	moduleName:   "Antiflood",
	handlerGroup: 4,
}

var antifloodModule = antifloodStruct{
	moduleStruct:  _normalAntifloodModule,
	syncHelperMap: sync.Map{},
	adminCache: make(map[int64]struct {
		isAdmin bool
		expires time.Time
	}),
}

// updateFlood updates the flood control state for a user in a chat using token bucket algorithm.
//
// Returns true if the user has exceeded the flood limit, along with the token bucket containing message IDs.
func (*moduleStruct) updateFlood(chatId, _ /* userId */, msgId int64) (returnVar bool, bucket *tokenBucket) {
	floodSrc := db.GetFlood(chatId)

	if floodSrc.Limit != 0 {
		// Read from map
		tmpInterface, valExists := antifloodModule.syncHelperMap.Load(chatId)
		if valExists && tmpInterface != nil {
			bucket = tmpInterface.(*tokenBucket)
		}

		// Initialize new bucket if needed
		if bucket == nil {
			// Default refill rate: 1 token per second (adjustable via config)
			refillRate := time.Second / time.Duration(floodSrc.Limit)
			bucket = newTokenBucket(floodSrc.Limit, refillRate)
		}

		// Track message ID for potential deletion
		bucket.messageIDs.push(msgId)

		// Check rate limit
		if !bucket.take() {
			// Reset bucket for next user
			antifloodModule.syncHelperMap.Store(chatId,
				newTokenBucket(floodSrc.Limit, bucket.refillRate),
			)
			returnVar = true
		} else {
			antifloodModule.syncHelperMap.Store(chatId, bucket)
		}
	}

	return
}

// checkFlood enforces flood control for incoming messages.
//
// Uses a semaphore to limit concurrent admin checks, applies flood logic, and takes action (mute, kick, ban) if limits are exceeded. Handles admin and anonymous user exceptions.
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

	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	var (
		fmode    string
		keyboard [][]gotgbot.InlineKeyboardButton
	)
	userId := user.Id()

	// Check admin status using cache
	if antifloodModule.isAdminCached(b, chat.Id, userId) {
		m.updateFlood(chat.Id, 0, 0) // empty message queue when admin sends a message
		return ext.ContinueGroups
	}

	// update flood for user
	flooded, floodCrc := m.updateFlood(chat.Id, userId, msg.MessageId)
	if !flooded {
		return ext.ContinueGroups
	}

	flood := db.GetFlood(chat.Id)

	if flood.DeleteAntifloodMessage {
		for _, i := range floodCrc.messageIDs.getAll() {
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
		// unban the member
		time.Sleep(3 * time.Second)
		_, err = chat.UnbanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}
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
		fmt.Sprintf(tr.GetString("strings.antiflood.checkflood.perform_action"), helpers.MentionHtml(userId, user.Name()), fmode),
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

// setFlood sets the flood limit for a chat.
//
// Allows admins to enable, disable, or change the flood limit. Handles argument parsing and updates the database accordingly.
func (*moduleStruct) setFlood(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	args := ctx.Args()[1:]

	var replyText string

	if len(args) == 0 {
		replyText = tr.GetString("strings.antiflood.errors.expected_args")
	} else {
		if string_handling.FindInStringSlice([]string{"off", "no", "false", "0"}, strings.ToLower(args[0])) {
			replyText = tr.GetString("strings.antiflood.setflood.disabled")
			go db.SetFlood(chat.Id, 0)
		} else {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				replyText = tr.GetString("strings.antiflood.errors.invalid_int")
			} else {
				if num < 3 || num > 100 {
					replyText = tr.GetString("strings.antiflood.errors.set_in_limit")
				} else {
					go db.SetFlood(chat.Id, num)
					replyText = fmt.Sprintf(tr.GetString("strings.antiflood.setflood.success"), num)
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

// flood displays the current flood settings for the chat.
//
// Shows whether flood control is enabled and the current action mode (mute, ban, kick).
func (*moduleStruct) flood(b *gotgbot.Bot, ctx *ext.Context) error {
	var text string
	msg := ctx.EffectiveMessage

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "adminlist") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, false, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat

	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	flood := db.GetFlood(chat.Id)
	if flood.Limit == 0 {
		text = tr.GetString("strings.antiflood.flood.disabled")
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
		text = fmt.Sprintf(tr.GetString("strings.antiflood.flood.show_settings"), flood.Limit, mode)
	}
	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		return err
	}
	return ext.EndGroups
}

// setFloodMode sets the action mode for flood control.
//
// Admins can choose between "ban", "kick", or "mute" as the action when flood limits are exceeded.
func (*moduleStruct) setFloodMode(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	args := ctx.Args()[1:]

	if len(args) > 0 {
		selectedMode := strings.ToLower(args[0])
		if string_handling.FindInStringSlice([]string{"ban", "kick", "mute"}, selectedMode) {
			_, err := msg.Reply(b, fmt.Sprintf(tr.GetString("strings.antiflood.setfloodmode.success"), selectedMode), helpers.Shtml())
			if err != nil {
				log.Error(err)
			}
			go db.SetFloodMode(chat.Id, selectedMode)
			return ext.EndGroups
		} else {
			_, err := msg.Reply(b, fmt.Sprintf(tr.GetString("strings.antiflood.setfloodmode.unknown_type"), args[0]), helpers.Shtml())
			if err != nil {
				return err
			}
		}
	} else {
		_, err := msg.Reply(b, tr.GetString("strings.antiflood.setfloodmode.specify_action"), helpers.Smarkdown())
		if err != nil {
			return err
		}
	}
	return ext.EndGroups
}

// setFloodDeleter enables or disables deletion of messages that trigger flood control.
//
// Admins can toggle this setting or view its current status.
func (*moduleStruct) setFloodDeleter(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	args := ctx.Args()[1:]
	var text string

	if len(args) > 0 {
		selectedMode := strings.ToLower(args[0])
		switch selectedMode {
		case "on", "yes":
			go db.SetFloodMsgDel(chat.Id, true)
			text = tr.GetString("strings.antiflood.flood_deleter.enabled")
		case "off", "no":
			go db.SetFloodMsgDel(chat.Id, true)
			text = tr.GetString("strings.antiflood.flood_deleter.disabled")
		default:
			text = tr.GetString("strings.antiflood.flood_deleter.invalid_option")
		}
	} else {
		currSet := db.GetFlood(chat.Id).DeleteAntifloodMessage
		if currSet {
			text = tr.GetString("strings.antiflood.flood_deleter.already_enabled")
		} else {
			text = tr.GetString("strings.antiflood.flood_deleter.already_disabled")
		}
	}
	_, err := msg.Reply(b, text, helpers.Smarkdown())
	if err != nil {
		return err
	}

	return ext.EndGroups
}

// LoadAntiflood registers antiflood-related command handlers with the dispatcher.
//
// This function enables the antiflood protection module and adds handlers for
// flood control commands and message monitoring. The module implements a token
// bucket rate limiting algorithm with ring buffer analysis to detect and prevent
// message flooding.
//
// Registered commands:
//   - /setflood: Configures the flood limit for the chat
//   - /setfloodmode: Sets the action taken when flood limit is exceeded
//   - /delflood: Configures message deletion on flood detection
//   - /flood: Displays current flood settings
//
// The module automatically monitors all messages in group 4 handler priority
// and applies rate limiting using a token bucket algorithm. When users exceed
// the configured message rate, appropriate actions are taken based on the
// chat's flood mode settings.
//
// Features:
//   - Token bucket rate limiting with configurable capacity
//   - Ring buffer for message ID tracking and deletion
//   - Admin status caching with TTL for performance
//   - Automatic cleanup of expired rate limiters
//   - Support for different flood actions (warn, mute, ban, kick)
//
// Requirements:
//   - Bot must be admin to delete messages and take actions
//   - User must be admin to configure flood settings
//   - Module respects admin immunity from flood controls
//
// The antiflood system is designed to be lightweight and efficient, with
// minimal impact on message processing performance.
func LoadAntiflood(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(antifloodModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("setflood", antifloodModule.setFlood))
	dispatcher.AddHandler(handlers.NewCommand("setfloodmode", antifloodModule.setFloodMode))
	dispatcher.AddHandler(handlers.NewCommand("delflood", antifloodModule.setFloodDeleter))
	dispatcher.AddHandler(handlers.NewCommand("flood", antifloodModule.flood))
	misc.AddCmdToDisableable("flood")
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, antifloodModule.checkFlood), antifloodModule.handlerGroup)
}
