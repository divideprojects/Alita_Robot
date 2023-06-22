package modules

import (
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

var usersModule = moduleStruct{
	moduleName:   "Users",
	handlerGroup: -1,
}

func (moduleStruct) logUsers(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender
	repliedMsg := msg.ReplyToMessage

	if user.IsAnonymousChannel() {
		log.Infof("Updatting channel %d in db", user.Id())
		// update when users send a message
		go db.UpdateChannel(
			user.Id(),
			user.Name(),
			user.Username(),
		)
	} else {
		// Don't add user to chat entry
		if chat_status.RequireGroup(bot, ctx, chat, true) {
			// Update user in chat collection
			go db.UpdateChat(
				chat.Id,
				chat.Title,
				user.Id(),
			)
		}

		log.Infof("Updatting user %d in db", user.Id())
		// update when users send a message
		go db.UpdateUser(
			user.Id(),
			user.Username(),
			user.Name(),
		)
	}

	// update is message is replied
	if repliedMsg != nil {
		log.Infof("Updatting %d in db", repliedMsg.GetSender().Id())
		if repliedMsg.GetSender().IsAnonymousChannel() {
			go db.UpdateChannel(
				repliedMsg.GetSender().Id(),
				repliedMsg.GetSender().Name(),
				repliedMsg.GetSender().Username(),
			)
		} else {
			go db.UpdateUser(
				repliedMsg.GetSender().Id(),
				repliedMsg.GetSender().Username(),
				repliedMsg.GetSender().Name(),
			)
		}
	}

	// update if message is forwarded
	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
		if msg.ForwardFromChat != nil && msg.ForwardFromChat.Type != "group" {
			go db.UpdateChannel(
				msg.ForwardFromChat.Id,
				msg.ForwardFromChat.Title,
				msg.ForwardFromChat.Username,
			)
		} else {
			// if chat type is not group
			go db.UpdateUser(
				msg.ForwardFrom.Id,
				msg.ForwardFrom.Username,
				helpers.GetFullName(
					msg.ForwardFrom.FirstName,
					msg.ForwardFrom.LastName,
				),
			)
		}
	}

	return ext.ContinueGroups
}

func LoadUsers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, usersModule.logUsers), usersModule.handlerGroup)
}
