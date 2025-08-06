package modules

import (
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

var (
	usersModule = moduleStruct{
		moduleName:   "Users",
		handlerGroup: -1,
	}

	// Rate limiting for database updates
	// Maps user/chat ID to last update timestamp
	userUpdateCache    = &sync.Map{}
	chatUpdateCache    = &sync.Map{}
	channelUpdateCache = &sync.Map{}

	// Update intervals
	userUpdateInterval    = 5 * time.Minute
	chatUpdateInterval    = 5 * time.Minute
	channelUpdateInterval = 5 * time.Minute
)

// shouldUpdate checks if enough time has passed since last update
func shouldUpdate(cache *sync.Map, id int64, interval time.Duration) bool {
	if lastUpdate, ok := cache.Load(id); ok {
		if time.Since(lastUpdate.(time.Time)) < interval {
			return false
		}
	}
	cache.Store(id, time.Now())
	return true
}

func (moduleStruct) logUsers(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender
	repliedMsg := msg.ReplyToMessage

	if user.IsAnonymousChannel() {
		// Only update if enough time has passed
		if shouldUpdate(channelUpdateCache, user.Id(), channelUpdateInterval) {
			log.Debugf("Updating channel %d in db", user.Id())
			// update when users send a message
			go db.UpdateChannel(
				user.Id(),
				user.Name(),
				user.Username(),
			)
		}
	} else {
		// Don't add user to chat entry
		if chat_status.RequireGroup(bot, ctx, chat, true) {
			// Update user in chat collection with rate limiting
			if shouldUpdate(chatUpdateCache, chat.Id, chatUpdateInterval) {
				go db.UpdateChat(
					chat.Id,
					chat.Title,
					user.Id(),
				)
			}
		}

		// Only update user if enough time has passed
		if shouldUpdate(userUpdateCache, user.Id(), userUpdateInterval) {
			log.Debugf("Updating user %d in db", user.Id())
			// update when users send a message
			go db.UpdateUser(
				user.Id(),
				user.Username(),
				user.Name(),
			)
		}
	}

	// update if message is replied
	if repliedMsg != nil {
		if repliedMsg.GetSender().IsAnonymousChannel() {
			if shouldUpdate(channelUpdateCache, repliedMsg.GetSender().Id(), channelUpdateInterval) {
				log.Debugf("Updating channel %d in db", repliedMsg.GetSender().Id())
				go db.UpdateChannel(
					repliedMsg.GetSender().Id(),
					repliedMsg.GetSender().Name(),
					repliedMsg.GetSender().Username(),
				)
			}
		} else {
			if shouldUpdate(userUpdateCache, repliedMsg.GetSender().Id(), userUpdateInterval) {
				log.Debugf("Updating user %d in db", repliedMsg.GetSender().Id())
				go db.UpdateUser(
					repliedMsg.GetSender().Id(),
					repliedMsg.GetSender().Username(),
					repliedMsg.GetSender().Name(),
				)
			}
		}
	}

	// update if message is forwarded
	if msg.ForwardOrigin != nil {
		forwarded := msg.ForwardOrigin.MergeMessageOrigin()
		if forwarded.Chat != nil && forwarded.Chat.Type != "group" {
			if shouldUpdate(channelUpdateCache, forwarded.Chat.Id, channelUpdateInterval) {
				go db.UpdateChannel(
					forwarded.Chat.Id,
					forwarded.Chat.Title,
					forwarded.Chat.Username,
				)
			}
		} else if forwarded.SenderUser != nil {
			// if chat type is not group
			if shouldUpdate(userUpdateCache, forwarded.SenderUser.Id, userUpdateInterval) {
				go db.UpdateUser(
					forwarded.SenderUser.Id,
					forwarded.SenderUser.Username,
					helpers.GetFullName(
						forwarded.SenderUser.FirstName,
						forwarded.SenderUser.LastName,
					),
				)
			}
		}
	}

	return ext.ContinueGroups
}

func LoadUsers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, usersModule.logUsers), usersModule.handlerGroup)
}
