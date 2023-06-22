package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
)

var (
	purgesModule = moduleStruct{moduleName: "Purges"}
	delMsgs      = map[int64]int64{}
)

func (moduleStruct) purgeMsgs(bot *gotgbot.Bot, chat *gotgbot.Chat, pFrom bool, msgId, deleteTo int64) bool {
	if !pFrom {
		_, err := bot.DeleteMessage(chat.Id, msgId, nil)
		if err != nil {
			if err.Error() == "unable to deleteMessage: Bad Request: message can't be deleted" {
				_, err = bot.SendMessage(chat.Id,
					"You cannot delete messages over two days old. Please choose a more recent message.",
					&gotgbot.SendMessageOpts{
						ReplyToMessageId:         deleteTo + 1,
						AllowSendingWithoutReply: true,
					},
				)
				if err != nil {
					log.Error(err)
					return false
				}
			} else {
				log.Error(err)
				return false
			}
			// else if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
			// 	log.WithFields(
			// 		log.Fields{
			// 			"chat": chat.Id,
			// 		},
			// 	).Error("error deleting message")
			// 	return false
		}
	}
	for mId := deleteTo + 1; mId > msgId-1; mId-- {
		_, _ = bot.DeleteMessage(chat.Id, mId, nil)
		// if err != nil {
		// if err.Error() != "unable to deleteMessage: Bad Request: message to delete not found" {
		// 	log.Error(err)
		// }
		// if err != nil {
		// 	log.Error(err)
		// }
		//	log.Error(err)
	}
	return true
}

func (m moduleStruct) purge(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Permission checks
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, ctx.EffectiveUser.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserDelete(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]

	if msg.ReplyToMessage != nil {
		msgId := msg.ReplyToMessage.MessageId
		deleteTo := msg.MessageId - 1
		totalMsgs := deleteTo - msgId + 1 // adding 1 because we want to delete the message we are replying to
		purge := m.purgeMsgs(bot, chat, false, msgId, deleteTo)
		_, err := bot.DeleteMessage(chat.Id, msg.MessageId, nil)
		// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
		// 	log.WithFields(
		// 		log.Fields{
		// 			"chat": chat.Id,
		// 		},
		// 	).Error("error deleting message")
		// } else
		if err != nil {
			log.Error(err)
		}

		if purge {
			Text := fmt.Sprintf("Purged %d messages.", totalMsgs)
			if len(args) >= 1 {
				Text += fmt.Sprintf("\n*Reason*:\n%s", args[0:])
			}
			pMsg, err := bot.SendMessage(chat.Id, Text, helpers.Smarkdown())
			if err != nil {
				log.Error(err)
			}

			time.Sleep(3 * time.Second)
			_, err = pMsg.Delete(bot, nil)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		_, err := msg.Reply(bot, "Reply to a message to select where to start purging from.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

func (moduleStruct) delCmd(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Permission checks
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserDelete(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	if msg.ReplyToMessage == nil {
		_, err := msg.Reply(bot, "Reply to a message to delete it!", nil)
		if err != nil {
			log.Error(err)
			return err
		}

	} else {
		msgId := msg.ReplyToMessage.MessageId
		_, _ = bot.DeleteMessage(chat.Id, msgId, nil)
		/* if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
		// 	log.WithFields(
		// 		log.Fields{
		// 			"chat": chat.Id,
		// 		},
		// 	).Error("error deleting message")
		} else {
		if err != nil {
			log.Error(err)
		}
		_, err = msg.Delete(bot, nil)
		// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
		// 	log.WithFields(
		// 		log.Fields{
		// 			"chat": chat.Id,
		// 		},
		// 	).Error("error deleting message")
		// } else
		if err != nil {
			log.Error(err)
		} */
		_, _ = msg.Delete(bot, nil)
	}

	return ext.EndGroups
}

func (moduleStruct) deleteButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// permissions check
	if !chat_status.CanUserDelete(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(b, ctx, nil, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	msgId, _ := strconv.Atoi(args[1])

	_, err := b.DeleteMessage(chat.Id, int64(msgId), nil)
	// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
	// 	log.WithFields(
	// 		log.Fields{
	// 			"chat": chat.Id,
	// 		},
	// 	).Error("error deleting message")
	// 	return ext.EndGroups
	// } else
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

func (moduleStruct) purgeFrom(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Permission checks
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, ctx.EffectiveUser.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserDelete(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	if msg.ReplyToMessage != nil {
		TodelId := msg.ReplyToMessage.MessageId
		if delMsgs[chat.Id] == TodelId {
			_, _ = msg.Reply(bot, "This message is already marked for purging!", nil)
			return ext.EndGroups
		}
		_, err := bot.DeleteMessage(chat.Id, msg.MessageId, nil)
		if err != nil {
			_, _ = msg.Reply(bot, err.Error(), nil)
			return ext.EndGroups
		}
		pMsg, err := bot.SendMessage(chat.Id, "Message marked for deletion. Reply to another message with /purgeto to delete all messages in between; within 30s!", &gotgbot.SendMessageOpts{ReplyToMessageId: TodelId})
		if err != nil {
			log.Error(err)
			return err
		}
		delMsgs[chat.Id] = TodelId
		time.Sleep(30 * time.Second)
		if delMsgs[chat.Id] == TodelId {
			delete(delMsgs, chat.Id)
		}
		_, err = pMsg.Delete(bot, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		_, err := msg.Reply(bot, "Reply to a message to select where to start purging from.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

func (m moduleStruct) purgeTo(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Permission checks
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotDelete(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, ctx.EffectiveUser.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserDelete(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]

	if msg.ReplyToMessage != nil {
		msgId := delMsgs[chat.Id]
		if msgId == 0 {
			_, err := msg.Reply(bot, "You can only use this command after having used the /purgefrom command!", nil)
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.EndGroups
		}
		deleteTo := msg.ReplyToMessage.MessageId
		if msgId == deleteTo {
			_, err := msg.Reply(bot, "Use /del command to delete one message!", nil)
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.EndGroups
		}
		totalMsgs := int64(0)
		if deleteTo > msgId {
			totalMsgs = deleteTo - msgId + 1
		} else {
			totalMsgs = msgId - deleteTo + 1
		}
		purge := m.purgeMsgs(bot, chat, true, msgId, deleteTo)
		_, err := bot.DeleteMessage(chat.Id, msg.MessageId, nil)
		if err != nil {
			log.Error(err)
		}
		if purge {
			Text := fmt.Sprintf("Purged %d messages.", totalMsgs)
			if len(args) >= 1 {
				Text += fmt.Sprintf("\n*Reason*:\n%s", args[0:])
			}
			pMsg, err := bot.SendMessage(chat.Id, Text, helpers.Smarkdown())
			if err != nil {
				log.Error(err)
			}
			time.Sleep(3 * time.Second)
			_, err = pMsg.Delete(bot, nil)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		_, err := msg.Reply(bot, "Reply to a message to show me till where to purge.", nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

func LoadPurges(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(purgesModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("del", purgesModule.delCmd))
	dispatcher.AddHandler(handlers.NewCommand("purge", purgesModule.purge))
	dispatcher.AddHandler(handlers.NewCommand("purgefrom", purgesModule.purgeFrom))
	dispatcher.AddHandler(handlers.NewCommand("purgeto", purgesModule.purgeTo))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("deleteMsg."), purgesModule.deleteButtonHandler))
}
