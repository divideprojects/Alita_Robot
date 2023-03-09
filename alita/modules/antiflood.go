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

	"github.com/Divkix/Alita_Robot/alita/i18n"
	"github.com/Divkix/Alita_Robot/alita/utils/decorators/misc"
	"github.com/Divkix/Alita_Robot/alita/utils/helpers"

	"github.com/Divkix/Alita_Robot/alita/db"
	"github.com/Divkix/Alita_Robot/alita/utils/chat_status"
	"github.com/Divkix/Alita_Robot/alita/utils/parsemode"
	"github.com/Divkix/Alita_Robot/alita/utils/string_handling"
)

type antifloodModuleStruct struct {
	modname      string
	handlerGroup int
	floodMap     sync.Map
}

type floodControl struct {
	userId       int64
	messageCount int
	messageIDs   []int64
}

var antifloodModule = antifloodModuleStruct{
	modname:      "Antiflood",
	handlerGroup: 4,
	floodMap:     sync.Map{},
}

func (m *antifloodModuleStruct) updateFlood(chatId, userId, msgId int64) (returnVar bool, floodCrc floodControl) {
	floodSrc := db.GetFlood(chatId)

	if floodSrc.Limit != 0 {

		// Read from map
		tmpInterface, valExists := antifloodModule.floodMap.Load(chatId)
		if valExists && tmpInterface != nil {
			floodCrc = tmpInterface.(floodControl)
		}

		if floodCrc.userId != userId || floodCrc.userId == 0 {
			floodCrc.userId = userId
			floodCrc.messageCount = 0
			floodCrc.messageIDs = make([]int64, 0)
		}

		floodCrc.messageCount++
		floodCrc.messageIDs = append([]int64{msgId}, floodCrc.messageIDs...) // prepend at first

		if floodCrc.messageCount > floodSrc.Limit {
			antifloodModule.floodMap.Store(chatId,
				floodControl{
					userId:       0,
					messageCount: 0,
					messageIDs:   make([]int64, 0),
				},
			)
			returnVar = true
		} else {
			antifloodModule.floodMap.Store(chatId, floodCrc)
		}
	}

	return
}

func (m *antifloodModuleStruct) checkFlood(b *gotgbot.Bot, ctx *ext.Context) error {
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

	// if user is admin, flood count will be 0
	if chat_status.IsUserAdmin(b, chat.Id, userId) {
		m.updateFlood(chat.Id, 0, 0) // empty message queue when admin or approved user sends a message
		return ext.ContinueGroups
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
		fmt.Sprintf(tr.GetString("strings."+m.modname+".checkflood.perform_action"), helpers.MentionHtml(userId, user.Name()), fmode),
		&gotgbot.SendMessageOpts{
			ParseMode: parsemode.HTML,
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

func (m *antifloodModuleStruct) setFlood(b *gotgbot.Bot, ctx *ext.Context) error {
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
		replyText = tr.GetString("strings." + m.modname + ".errors.expected_args")
	} else {
		if string_handling.FindInStringSlice([]string{"off", "no", "false", "0"}, strings.ToLower(args[0])) {
			replyText = tr.GetString("strings." + m.modname + ".setflood.disabled")
			go db.SetFlood(chat.Id, 0)
		} else {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				replyText = tr.GetString("strings." + m.modname + ".errors.invalid_int")
			} else {
				if num < 3 || num > 100 {
					replyText = tr.GetString("strings." + m.modname + ".errors.set_in_limit")
				} else {
					go db.SetFlood(chat.Id, num)
					replyText = fmt.Sprintf(tr.GetString("strings."+m.modname+".setflood.success"), num)
				}
			}
		}
	}

	_, err := msg.Reply(b, replyText, parsemode.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func (m *antifloodModuleStruct) flood(b *gotgbot.Bot, ctx *ext.Context) error {
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
		text = tr.GetString("strings." + m.modname + ".flood.disabled")
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
		text = fmt.Sprintf(tr.GetString("strings."+m.modname+".flood.show_settings"), flood.Limit, mode)
	}
	_, err := msg.Reply(b, text, parsemode.Shtml())
	if err != nil {
		return err
	}
	return ext.EndGroups
}

func (m *antifloodModuleStruct) setFloodMode(b *gotgbot.Bot, ctx *ext.Context) error {
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
			_, err := msg.Reply(b, fmt.Sprintf(tr.GetString("strings."+m.modname+".setfloodmode.success"), selectedMode), parsemode.Shtml())
			if err != nil {
				log.Error(err)
			}
			go db.SetFloodMode(chat.Id, selectedMode)
			return ext.EndGroups
		} else {
			_, err := msg.Reply(b, fmt.Sprintf(tr.GetString("strings."+m.modname+".setfloodmode.unknown_type"), args[0]), parsemode.Shtml())
			if err != nil {
				return err
			}
		}
	} else {
		_, err := msg.Reply(b, tr.GetString("strings."+m.modname+".setfloodmode.specify_action"), parsemode.Smarkdown())
		if err != nil {
			return err
		}
	}
	return ext.EndGroups
}

func (m *antifloodModuleStruct) setFloodDeleter(b *gotgbot.Bot, ctx *ext.Context) error {
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
			text = tr.GetString("strings." + m.modname + ".flood_deleter.enabled")
		case "off", "no":
			go db.SetFloodMsgDel(chat.Id, true)
			text = tr.GetString("strings." + m.modname + ".flood_deleter.disabled")
		default:
			text = tr.GetString("strings." + m.modname + ".flood_deleter.invalid_option")
		}
	} else {
		currSet := db.GetFlood(chat.Id).DeleteAntifloodMessage
		if currSet {
			text = tr.GetString("strings." + m.modname + ".flood_deleter.already_enabled")
		} else {
			text = tr.GetString("strings." + m.modname + ".flood_deleter.already_disabled")
		}
	}
	_, err := msg.Reply(b, text, parsemode.Smarkdown())
	if err != nil {
		return err
	}

	return ext.EndGroups
}

func LoadAntiflood(dispatcher *ext.Dispatcher) {

	HelpModule.AbleMap.Store(antifloodModule.modname, true)

	dispatcher.AddHandler(handlers.NewCommand("setflood", antifloodModule.setFlood))
	dispatcher.AddHandler(handlers.NewCommand("setfloodmode", antifloodModule.setFloodMode))
	dispatcher.AddHandler(handlers.NewCommand("delflood", antifloodModule.setFloodDeleter))
	dispatcher.AddHandler(handlers.NewCommand("flood", antifloodModule.flood))
	misc.AddCmdToDisableable("flood")
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, antifloodModule.checkFlood), antifloodModule.handlerGroup)
}
