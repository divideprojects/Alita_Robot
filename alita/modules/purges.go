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

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
)

var (
	purgesModule = moduleStruct{moduleName: "Purges"}
	delMsgs      = map[int64]int64{}
)

// purgeMsgs performs the actual message deletion operation for purge commands,
// deleting messages in the specified range with error handling for old messages.
func (moduleStruct) purgeMsgs(bot *gotgbot.Bot, chat *gotgbot.Chat, pFrom bool, msgId, deleteTo int64) bool {
	if !pFrom {
		_, err := bot.DeleteMessage(chat.Id, msgId, nil)
		if err != nil {
			if err.Error() == "unable to deleteMessage: Bad Request: message can't be deleted" {
				// Get translator for error messages
				translator := i18n.MustNewTranslator("en") // fallback for error handling
				_, err = bot.SendMessage(chat.Id,
					translator.Message("purges_old_message_limit", nil),
					&gotgbot.SendMessageOpts{
						ReplyParameters: &gotgbot.ReplyParameters{
							MessageId:                deleteTo + 1,
							AllowSendingWithoutReply: true,
						},
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

// purge handles the /purge command to delete all messages from a replied
// message up to the command message, requiring admin permissions.
func (m moduleStruct) purge(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

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
		_, err := msg.Reply(bot, translator.Message("purges_reply_to_start", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

// delCmd handles the /del command to delete a specific replied message
// along with the command message, requiring admin permissions.
func (moduleStruct) delCmd(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

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
		_, err := msg.Reply(bot, translator.Message("purges_reply_to_delete", nil), nil)
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

// deleteButtonHandler processes callback queries from delete buttons
// to remove specific messages, requiring admin permissions.
func (moduleStruct) deleteButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
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

// purgeFrom handles the /purgefrom command to mark a starting message
// for range deletion, requiring admin permissions.
func (moduleStruct) purgeFrom(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

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
			_, _ = msg.Reply(bot, translator.Message("purges_already_marked", nil), nil)
			return ext.EndGroups
		}
		_, err := bot.DeleteMessage(chat.Id, msg.MessageId, nil)
		if err != nil {
			_, _ = msg.Reply(bot, err.Error(), nil)
			return ext.EndGroups
		}
		pMsg, err := bot.SendMessage(chat.Id,
			translator.Message("purges_marked_for_deletion", nil),
			&gotgbot.SendMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                TodelId,
					AllowSendingWithoutReply: true,
				},
			},
		)
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
		_, err := msg.Reply(bot, translator.Message("purges_reply_to_start", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// purgeTo handles the /purgeto command to complete range deletion
// from a previously marked message, requiring admin permissions.
func (m moduleStruct) purgeTo(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

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
			_, err := msg.Reply(bot, translator.Message("purges_need_purgefrom_first", nil), nil)
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.EndGroups
		}
		deleteTo := msg.ReplyToMessage.MessageId
		if msgId == deleteTo {
			_, err := msg.Reply(bot, translator.Message("purges_use_del_for_single", nil), nil)
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
		_, err := msg.Reply(bot, translator.Message("purges_reply_to_end", nil), nil)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

// LoadPurges registers all purges module handlers with the dispatcher,
// including message deletion commands and callback handlers.
func LoadPurges(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(purgesModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("del", purgesModule.delCmd))
	dispatcher.AddHandler(handlers.NewCommand("purge", purgesModule.purge))
	dispatcher.AddHandler(handlers.NewCommand("purgefrom", purgesModule.purgeFrom))
	dispatcher.AddHandler(handlers.NewCommand("purgeto", purgesModule.purgeTo))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("deleteMsg."), purgesModule.deleteButtonHandler))
}
