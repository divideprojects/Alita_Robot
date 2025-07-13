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

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

/*
antifloodStruct implements the antiflood module logic.

It embeds moduleStruct and manages flood control state per chat, including a semaphore to limit concurrent admin checks for performance and safety.
*/
type antifloodStruct struct {
	moduleStruct  // inheritance
	syncHelperMap sync.Map
	// Add semaphore to limit concurrent admin checks
	adminCheckSemaphore chan struct{}
}

/*
floodControl tracks message counts and IDs for a user in a chat to enforce flood limits.
*/
type floodControl struct {
	userId       int64
	messageCount int
	messageIDs   []int64
}

var _normalAntifloodModule = moduleStruct{
	moduleName:   autoModuleName(),
	handlerGroup: 4,
	cfg:          nil, // will be set during LoadAntiflood
}

var antifloodModule = antifloodStruct{
	moduleStruct:        _normalAntifloodModule,
	syncHelperMap:       sync.Map{},
	adminCheckSemaphore: make(chan struct{}, 50), // Limit to 50 concurrent admin checks
}

/*
updateFlood updates the flood control state for a user in a chat.

Returns true if the user has exceeded the flood limit, along with the updated floodControl struct.
*/
func (*moduleStruct) updateFlood(chatId, userId, msgId int64) (returnVar bool, floodCrc floodControl) {
	floodSrc := db.GetFlood(chatId)

	if floodSrc.Limit != 0 {

		// Read from map
		tmpInterface, valExists := antifloodModule.syncHelperMap.Load(chatId)
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
			antifloodModule.syncHelperMap.Store(chatId,
				floodControl{
					userId:       0,
					messageCount: 0,
					messageIDs:   make([]int64, 0),
				},
			)
			returnVar = true
		} else {
			antifloodModule.syncHelperMap.Store(chatId, floodCrc)
		}
	}

	return
}

/*
checkFlood enforces flood control for incoming messages.

Uses a semaphore to limit concurrent admin checks, applies flood logic, and takes action (mute, kick, ban) if limits are exceeded. Handles admin and anonymous user exceptions.
*/
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

	tr := i18n.New(db.GetLanguage(ctx))

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

		// Check if user is admin with timeout
		isAdmin := make(chan bool, 1)
		go func() {
			isAdmin <- chat_status.IsUserAdmin(b, chat.Id, userId)
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
		unmuteAdminsOnlyMsg, unmuteAdminsOnlyErr := tr.GetStringWithError("strings.antiflood.unmute_admins_only")
		if unmuteAdminsOnlyErr != nil {
			log.Errorf("[antiflood] missing translation for unmute_admins_only: %v", unmuteAdminsOnlyErr)
			unmuteAdminsOnlyMsg = "Unmute (Admins Only)"
		}
		keyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         unmuteAdminsOnlyMsg,
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
			unbanAdminsOnlyMsg, unbanAdminsOnlyErr := tr.GetStringWithError("strings.antiflood.unban_admins_only")
			if unbanAdminsOnlyErr != nil {
				log.Errorf("[antiflood] missing translation for unban_admins_only: %v", unbanAdminsOnlyErr)
				unbanAdminsOnlyMsg = "Unban (Admins Only)"
			}
			keyboard = [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         unbanAdminsOnlyMsg,
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
	performActionMsg, performActionErr := tr.GetStringWithError("strings.antiflood.checkflood.perform_action")
	if performActionErr != nil {
		log.Errorf("[antiflood] missing translation for checkflood.perform_action: %v", performActionErr)
		performActionMsg = "Yeah, I don't like your flooding. %s has been %s!"
	}
	if _, err := b.SendMessage(chat.Id,
		fmt.Sprintf(performActionMsg, helpers.MentionHtml(userId, user.Name()), fmode),
		&gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
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

/*
setFlood sets the flood limit for a chat.

Allows admins to enable, disable, or change the flood limit. Handles argument parsing and updates the database accordingly.
*/
func (m *moduleStruct) setFlood(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.New(db.GetLanguage(ctx))
	args := ctx.Args()[1:]

	var replyText string

	if len(args) == 0 {
		expectedArgsMsg, expectedArgsErr := tr.GetStringWithError("strings.antiflood.errors.expected_args")
		if expectedArgsErr != nil {
			log.Errorf("[antiflood] missing translation for errors.expected_args: %v", expectedArgsErr)
			expectedArgsMsg = "I expected some arguments! Either off, or an integer. eg: `/setflood 5`, or `/setflood off`"
		}
		replyText = expectedArgsMsg
	} else {
		if string_handling.FindInStringSlice([]string{"off", "no", "false", "0"}, strings.ToLower(args[0])) {
			setfloodDisabledMsg, setfloodDisabledErr := tr.GetStringWithError("strings.antiflood.setflood.disabled")
			if setfloodDisabledErr != nil {
				log.Errorf("[antiflood] missing translation for setflood.disabled: %v", setfloodDisabledErr)
				setfloodDisabledMsg = "Okay.\nI won't warn users for flooding."
			}
			replyText = setfloodDisabledMsg
			go db.SetFlood(chat.Id, 0)
		} else {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				invalidIntMsg, invalidIntErr := tr.GetStringWithError("strings.antiflood.errors.invalid_int")
				if invalidIntErr != nil {
					log.Errorf("[antiflood] missing translation for errors.invalid_int: %v", invalidIntErr)
					invalidIntMsg = "That's not a valid integer. Please give me a valid integer, or `off`."
				}
				replyText = invalidIntMsg
			} else {
				if num < 3 || num > 100 {
					setInLimitMsg, setInLimitErr := tr.GetStringWithError("strings.antiflood.errors.set_in_limit")
					if setInLimitErr != nil {
						log.Errorf("[antiflood] missing translation for errors.set_in_limit: %v", setInLimitErr)
						setInLimitMsg = "The antiflood limit has to be set between 3 and 100."
					}
					replyText = setInLimitMsg
				} else {
					go db.SetFlood(chat.Id, num)
					setfloodSuccessMsg, setfloodSuccessErr := tr.GetStringWithError("strings.antiflood.setflood.success")
					if setfloodSuccessErr != nil {
						log.Errorf("[antiflood] missing translation for setflood.success: %v", setfloodSuccessErr)
						setfloodSuccessMsg = "Flood Limit has been set to <b>%d</b> messages."
					}
					replyText = fmt.Sprintf(setfloodSuccessMsg, num)
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

/*
flood displays the current flood settings for the chat.

Shows whether flood control is enabled and the current action mode (mute, ban, kick).
*/
func (m *moduleStruct) flood(b *gotgbot.Bot, ctx *ext.Context) error {
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

	tr := i18n.New(db.GetLanguage(ctx))

	flood := db.GetFlood(chat.Id)
	if flood.Limit == 0 {
		floodDisabledMsg, floodDisabledErr := tr.GetStringWithError("strings.antiflood.flood.disabled")
		if floodDisabledErr != nil {
			log.Errorf("[antiflood] missing translation for flood.disabled: %v", floodDisabledErr)
			floodDisabledMsg = "This chat is not currently enforcing flood control."
		}
		text = floodDisabledMsg
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
		showSettingsMsg, showSettingsErr := tr.GetStringWithError("strings.antiflood.flood.show_settings")
		if showSettingsErr != nil {
			log.Errorf("[antiflood] missing translation for flood.show_settings: %v", showSettingsErr)
			showSettingsMsg = "This chat is currently enforcing flood control after %d messages. Any users sending more than that amount of messages will be %s."
		}
		text = fmt.Sprintf(showSettingsMsg, flood.Limit, mode)
	}
	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		return err
	}
	return ext.EndGroups
}

/*
setFloodMode sets the action mode for flood control.

Admins can choose between "ban", "kick", or "mute" as the action when flood limits are exceeded.
*/
func (m *moduleStruct) setFloodMode(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.New(db.GetLanguage(ctx))
	args := ctx.Args()[1:]

	if len(args) > 0 {
		selectedMode := strings.ToLower(args[0])
		if string_handling.FindInStringSlice([]string{"ban", "kick", "mute"}, selectedMode) {
			setfloodmodeSuccessMsg, setfloodmodeSuccessErr := tr.GetStringWithError("strings.antiflood.setfloodmode.success")
			if setfloodmodeSuccessErr != nil {
				log.Errorf("[antiflood] missing translation for setfloodmode.success: %v", setfloodmodeSuccessErr)
				setfloodmodeSuccessMsg = "Flood mode has been set to %s."
			}
			_, err := msg.Reply(b, fmt.Sprintf(setfloodmodeSuccessMsg, selectedMode), helpers.Shtml())
			if err != nil {
				log.Error(err)
			}
			go db.SetFloodMode(chat.Id, selectedMode)
			return ext.EndGroups
		} else {
			unknownTypeMsg, unknownTypeErr := tr.GetStringWithError("strings.antiflood.setfloodmode.unknown_type")
			if unknownTypeErr != nil {
				log.Errorf("[antiflood] missing translation for setfloodmode.unknown_type: %v", unknownTypeErr)
				unknownTypeMsg = "Unknown type '%s'. Please use one of: ban/kick/mute"
			}
			_, err := msg.Reply(b, fmt.Sprintf(unknownTypeMsg, args[0]), helpers.Shtml())
			if err != nil {
				return err
			}
		}
	} else {
		specifyActionMsg, specifyActionErr := tr.GetStringWithError("strings.antiflood.setfloodmode.specify_action")
		if specifyActionErr != nil {
			log.Errorf("[antiflood] missing translation for setfloodmode.specify_action: %v", specifyActionErr)
			specifyActionMsg = "You need to specify an action to take upon flooding. Current modes are: `ban`/`kick`/`mute`"
		}
		_, err := msg.Reply(b, specifyActionMsg, helpers.Smarkdown())
		if err != nil {
			return err
		}
	}
	return ext.EndGroups
}

/*
setFloodDeleter enables or disables deletion of messages that trigger flood control.

Admins can toggle this setting or view its current status.
*/
func (m *moduleStruct) setFloodDeleter(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	tr := i18n.New(db.GetLanguage(ctx))
	args := ctx.Args()[1:]
	var text string

	if len(args) > 0 {
		selectedMode := strings.ToLower(args[0])
		switch selectedMode {
		case "on", "yes":
			go db.SetFloodMsgDel(chat.Id, true)
			floodDeleterEnabledMsg, floodDeleterEnabledErr := tr.GetStringWithError("strings.antiflood.flood_deleter.enabled")
			if floodDeleterEnabledErr != nil {
				log.Errorf("[antiflood] missing translation for flood_deleter.enabled: %v", floodDeleterEnabledErr)
				floodDeleterEnabledMsg = "Turned on Antiflood Message Deleting. Messages causing antiflood will be deleted!"
			}
			text = floodDeleterEnabledMsg
		case "off", "no":
			go db.SetFloodMsgDel(chat.Id, true)
			floodDeleterDisabledMsg, floodDeleterDisabledErr := tr.GetStringWithError("strings.antiflood.flood_deleter.disabled")
			if floodDeleterDisabledErr != nil {
				log.Errorf("[antiflood] missing translation for flood_deleter.disabled: %v", floodDeleterDisabledErr)
				floodDeleterDisabledMsg = "Turned off Antiflood Message Deleting. Messages causing antiflood will be ignored!"
			}
			text = floodDeleterDisabledMsg
		default:
			floodDeleterInvalidMsg, floodDeleterInvalidErr := tr.GetStringWithError("strings.antiflood.flood_deleter.invalid_option")
			if floodDeleterInvalidErr != nil {
				log.Errorf("[antiflood] missing translation for flood_deleter.invalid_option: %v", floodDeleterInvalidErr)
				floodDeleterInvalidMsg = "I only understand an option from: `yes`/`no`/`on`/`off`"
			}
			text = floodDeleterInvalidMsg
		}
	} else {
		currSet := db.GetFlood(chat.Id).DeleteAntifloodMessage
		if currSet {
			floodDeleterAlreadyEnabledMsg, floodDeleterAlreadyEnabledErr := tr.GetStringWithError("strings.antiflood.flood_deleter.already_enabled")
			if floodDeleterAlreadyEnabledErr != nil {
				log.Errorf("[antiflood] missing translation for flood_deleter.already_enabled: %v", floodDeleterAlreadyEnabledErr)
				floodDeleterAlreadyEnabledMsg = "Antiflood Message deleting is currently *enabled* in this chat, Messages causing antiflood will be deleted!"
			}
			text = floodDeleterAlreadyEnabledMsg
		} else {
			floodDeleterAlreadyDisabledMsg, floodDeleterAlreadyDisabledErr := tr.GetStringWithError("strings.antiflood.flood_deleter.already_disabled")
			if floodDeleterAlreadyDisabledErr != nil {
				log.Errorf("[antiflood] missing translation for flood_deleter.already_disabled: %v", floodDeleterAlreadyDisabledErr)
				floodDeleterAlreadyDisabledMsg = "Antiflood Message deleting is currently *disabled* in this chat, Messages causing antiflood will be ignored!"
			}
			text = floodDeleterAlreadyDisabledMsg
		}
	}
	_, err := msg.Reply(b, text, helpers.Smarkdown())
	if err != nil {
		return err
	}

	return ext.EndGroups
}

/*
LoadAntiflood registers antiflood-related command handlers with the dispatcher.

Enables the antiflood module and adds handlers for flood control commands and message group checks.
*/
func LoadAntiflood(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	_normalAntifloodModule.cfg = cfg

	HelpModule.AbleMap.Store(antifloodModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("setflood", antifloodModule.setFlood))
	dispatcher.AddHandler(handlers.NewCommand("setfloodmode", antifloodModule.setFloodMode))
	dispatcher.AddHandler(handlers.NewCommand("delflood", antifloodModule.setFloodDeleter))
	dispatcher.AddHandler(handlers.NewCommand("flood", antifloodModule.flood))
	misc.AddCmdToDisableable("flood")
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, antifloodModule.checkFlood), antifloodModule.handlerGroup)
}
