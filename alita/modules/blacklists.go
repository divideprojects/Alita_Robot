package modules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/cmdDecorator"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

/*
blacklistsModule provides blacklist management logic for group chats.

Implements commands to add, remove, list, and configure blacklists and their actions.
*/
var blacklistsModule = moduleStruct{
	moduleName:   "Blacklists",
	handlerGroup: 7,
	cfg:          nil, // will be set during LoadBlacklists
}

/*
	Used to add a blacklist to group!

Connection - true, true
Admin can add a blacklist to the chat
*/
/*
addBlacklist adds one or more blacklist words to the group.

Checks permissions, updates the blacklist in the database, and replies with the result.
Connection: true, true
*/
func (m moduleStruct) addBlacklist(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]
	tr := i18n.New(db.GetLanguage(ctx))

	var (
		alreadyBlacklisted, newBlacklist []string
		text                             string
	)

	// Permission Checks
	if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		return ext.EndGroups
	}
	if !chat_status.IsBotAdmin(b, ctx, chat) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, chat, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		giveBlWordMsg, giveBlWordErr := tr.GetStringWithError("strings."+m.moduleName+".blacklist.give_bl_word")
		if giveBlWordErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", giveBlWordErr)
			giveBlWordMsg = "Please give me a word to add to the blacklist!"
		}
		_, err := msg.Reply(b, giveBlWordMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if len(args) >= 1 {
		allBlWords := db.GetBlacklistSettings(chat.Id).Triggers
		for _, blWord := range args {
			if string_handling.FindInStringSlice(allBlWords, blWord) {
				alreadyBlacklisted = append(alreadyBlacklisted, blWord)
			} else {
				go db.AddBlacklist(chat.Id, blWord)
				newBlacklist = append(newBlacklist, fmt.Sprintf("<code>%s</code>", blWord))
			}
		}

		if len(alreadyBlacklisted) >= 1 {
			alreadyBlacklistedMsg, alreadyBlacklistedErr := tr.GetStringWithError("strings."+m.moduleName+".blacklist.already_blacklisted")
			if alreadyBlacklistedErr != nil {
				log.Errorf("[blacklists] missing translation for key: %v", alreadyBlacklistedErr)
				alreadyBlacklistedMsg = "These words are already blacklisted:"
			}
			text += alreadyBlacklistedMsg + fmt.Sprintf("\n - %s\n\n", strings.Join(alreadyBlacklisted, "\n - "))
		}
		if len(newBlacklist) >= 1 {
			addedBlMsg, addedBlErr := tr.GetStringWithError("strings."+m.moduleName+".blacklist.added_bl")
			if addedBlErr != nil {
				log.Errorf("[blacklists] missing translation for key: %v", addedBlErr)
				addedBlMsg = "Added these words as blacklists:"
			}
			text += addedBlMsg + fmt.Sprintf("\n - %s\n\n", strings.Join(newBlacklist, "\n - "))
		}

		_, err := msg.Reply(b, text, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

/*
	Used to remove a blacklist from group!

Connection - true, true
Admin can add a blacklist to the chat
*/
/*
removeBlacklist removes one or more blacklist words from the group.

Checks permissions, updates the blacklist in the database, and replies with the result.
Connection: true, true
*/
func (m moduleStruct) removeBlacklist(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]
	tr := i18n.New(db.GetLanguage(ctx))

	var removedBlacklists []string

	// Permission Checks
	if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		return ext.EndGroups
	}
	if !chat_status.IsBotAdmin(b, ctx, chat) {
		return ext.EndGroups
	}
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, chat, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		giveBlWordMsg, giveBlWordErr := tr.GetStringWithError("strings."+m.moduleName+".unblacklist.give_bl_word")
		if giveBlWordErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", giveBlWordErr)
			giveBlWordMsg = "Please give me a word to remove it from the blacklist!"
		}
		_, err := msg.Reply(b, giveBlWordMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else {
		allBlWords := db.GetBlacklistSettings(chat.Id).Triggers
		for _, blWord := range args {
			if string_handling.FindInStringSlice(allBlWords, blWord) {
				removedBlacklists = append(removedBlacklists, blWord)
				go db.RemoveBlacklist(chat.Id, blWord)
			}
		}
		if len(removedBlacklists) <= 0 {
			noRemovedBlMsg, noRemovedBlErr := tr.GetStringWithError("strings."+m.moduleName+".unblacklist.no_removed_bl")
			if noRemovedBlErr != nil {
				log.Errorf("[blacklists] missing translation for key: %v", noRemovedBlErr)
				noRemovedBlMsg = "None of the given words were on the blacklist which can be removed!"
			}
			_, err := msg.Reply(b, noRemovedBlMsg, nil)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			removedBlMsg, removedBlErr := tr.GetStringWithError("strings."+m.moduleName+".unblacklist.removed_bl")
			if removedBlErr != nil {
				log.Errorf("[blacklists] missing translation for key: %v", removedBlErr)
				removedBlMsg = "Successfully removed '%s' from the blacklisted words!"
			}
			_, err := msg.Reply(b,
				fmt.Sprintf(removedBlMsg, strings.Join(removedBlacklists, ", ")),
				nil,
			)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return ext.EndGroups
}

/*
	Used to list all blacklists of a group!

Connection - false, true
Anyone can view blacklists in group
*/
/*
listBlacklists lists all blacklist words in the group.

Anyone can view the blacklist. Replies with the current list or a message if none exist.
Connection: false, true
*/
func (m moduleStruct) listBlacklists(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.New(db.GetLanguage(ctx))
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

	var (
		replyMsgId     int64
		blacklistsText string
	)

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	blSrc := db.GetBlacklistSettings(chat.Id)
	sort.Strings(blSrc.Triggers)
	for _, i := range blSrc.Triggers {
		blacklistsText += fmt.Sprintf("\n - <code>%s</code>", i)
	}

	if blacklistsText != "" {
		listBlMsg, listBlErr := tr.GetStringWithError("strings."+m.moduleName+".ls_bl.list_bl")
		if listBlErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", listBlErr)
			listBlMsg = "These words are blacklisted in this chat:"
		}
		blacklistsText = listBlMsg + blacklistsText
	} else {
		noBlacklistedMsg, noBlacklistedErr := tr.GetStringWithError("strings." + m.moduleName + ".ls_bl.no_blacklisted")
		if noBlacklistedErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", noBlacklistedErr)
			noBlacklistedMsg = "There are no blacklisted words in this chat."
		}
		blacklistsText = noBlacklistedMsg
	}

	_, err := msg.Reply(b,
		blacklistsText,
		&gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                replyMsgId,
				AllowSendingWithoutReply: true,
			},
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Used to set mode for blacklists in chat

# Connection - true, true

Admin with restriction permission can set blacklist action in group out of - ick, ban, mute
*/
/*
setBlacklistAction sets the action to take when a blacklist word is triggered.

Admins can set the action to "mute", "kick", "warn", "ban", or "none".
Connection: true, true
*/
func (m moduleStruct) setBlacklistAction(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]
	tr := i18n.New(db.GetLanguage(ctx))

	var rMsg string

	// Permission Checks
	if !chat_status.CanUserRestrict(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotRestrict(b, ctx, chat, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		currAction := db.GetBlacklistSettings(chat.Id).Action
		currentModeMsg, currentModeErr := tr.GetStringWithError("strings."+m.moduleName+".set_bl_action.current_mode")
		if currentModeErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", currentModeErr)
			currentModeMsg = "The current blacklist mode for this chat is: %s"
		}
		rMsg = fmt.Sprintf(currentModeMsg, currAction)
	} else if len(args) == 1 {
		action := strings.ToLower(args[0])
		if string_handling.FindInStringSlice([]string{"mute", "kick", "warn", "ban", "none"}, action) {
			changedModeMsg, changedModeErr := tr.GetStringWithError("strings."+m.moduleName+".set_bl_action.changed_mode")
			if changedModeErr != nil {
				log.Errorf("[blacklists] missing translation for key: %v", changedModeErr)
				changedModeMsg = "Successfully Changed blacklist mode to: *%s*"
			}
			rMsg = fmt.Sprintf(changedModeMsg, action)
			go db.SetBlacklistAction(chat.Id, action)
		} else {
			chooseCorrectOptionMsg, chooseCorrectOptionErr := tr.GetStringWithError("strings." + m.moduleName + ".set_bl_action.choose_correct_option")
			if chooseCorrectOptionErr != nil {
				log.Errorf("[blacklists] missing translation for key: %v", chooseCorrectOptionErr)
				chooseCorrectOptionMsg = "Please choose an option out of <mute/kick/ban/warn/none>"
			}
			rMsg = chooseCorrectOptionMsg
		}
	} else {
		chooseCorrectOptionMsg, chooseCorrectOptionErr := tr.GetStringWithError("strings." + m.moduleName + ".set_bl_action.choose_correct_option")
		if chooseCorrectOptionErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", chooseCorrectOptionErr)
			chooseCorrectOptionMsg = "Please choose an option out of <mute/kick/ban/warn/none>"
		}
		rMsg = chooseCorrectOptionMsg
	}
	_, err := msg.Reply(b, rMsg, helpers.Smarkdown())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/*
	Used to remove all blacklists from a group

Only chat creator can use this command to remove all blacklists aat once from the current chat
*/
/*
rmAllBlacklists removes all blacklist words from the group.

Only the chat creator can use this command to clear the blacklist.
*/
func (m moduleStruct) rmAllBlacklists(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserOwner(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	askMsg, askErr := tr.GetStringWithError("strings."+m.moduleName+".rm_all_bl.ask")
	if askErr != nil {
		log.Errorf("[blacklists] missing translation for key: %v", askErr)
		askMsg = "Are you sure you want to remove all blacklisted words from this chat?"
	}

	yesMsg, yesErr := tr.GetStringWithError("strings.CommonStrings.buttons.yes")
	if yesErr != nil {
		log.Errorf("[blacklists] missing translation for key: %v", yesErr)
		yesMsg = "Yes"
	}

	noMsg, noErr := tr.GetStringWithError("strings.CommonStrings.buttons.no")
	if noErr != nil {
		log.Errorf("[blacklists] missing translation for key: %v", noErr)
		noMsg = "No"
	}

	_, err := msg.Reply(b, askMsg,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: yesMsg, CallbackData: "rmAllBlacklist.yes"},
						{Text: noMsg, CallbackData: "rmAllBlacklist.no"},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// Callback Handler for rmallblacklist
/*
buttonHandler handles callback queries for removing all blacklists.

Processes the creator's confirmation and removes all blacklist words if confirmed.
*/
func (m moduleStruct) buttonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	user := query.From
	tr := i18n.New(db.GetLanguage(ctx))

	// permission checks
	if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	creatorAction := args[1]
	var helpText string

	switch creatorAction {
	case "yes":
		go db.RemoveAllBlacklist(query.Message.GetChat().Id)
		yesButtonMsg, yesButtonErr := tr.GetStringWithError("strings." + m.moduleName + ".rm_all_bl.button_handler.true")
		if yesButtonErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", yesButtonErr)
			yesButtonMsg = "Removed all Blacklists from this Chat ✅"
		}
		helpText = yesButtonMsg
	case "no":
		noButtonMsg, noButtonErr := tr.GetStringWithError("strings." + m.moduleName + ".rm_all_bl.button_handler.false")
		if noButtonErr != nil {
			log.Errorf("[blacklists] missing translation for key: %v", noButtonErr)
			noButtonMsg = "Canceled removing all Blacklists of this Chat ❌"
		}
		helpText = noButtonMsg
	}

	_, _, err := query.Message.EditText(b,
		helpText,
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b,
		&gotgbot.AnswerCallbackQueryOpts{
			Text: helpText,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Blacklist watcher

Watcher for blacklisted words, if any of the sentence contains the word, it will remove and use the appropriate action
*/
/*
blacklistWatcher monitors messages for blacklisted words and enforces the configured action.

Deletes messages containing blacklisted words and applies the configured action (mute, ban, kick, warn) to the user.
*/
func (m moduleStruct) blacklistWatcher(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender
	if user.IsAnonymousAdmin() {
		return ext.ContinueGroups
	}

	// skip admins and creator + approved users and anonymous channel
	if !user.IsAnonymousChannel() && chat_status.IsUserAdmin(b, chat.Id, user.Id()) {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	blSettings := db.GetBlacklistSettings(chat.Id)
	tr := i18n.New(db.GetLanguage(ctx))

	for _, i := range blSettings.Triggers {
		match, _ := regexp.MatchString(fmt.Sprintf(`(\b|\s)%s\b`, i), strings.ToLower(msg.Text))
		if match {
			_, err := msg.Delete(b, nil)
			if err != nil {
				log.Error(err)
				return err
			}
			switch blSettings.Action {
			case "mute":
				// don't work on anonymous channels
				if user.IsAnonymousChannel() {
					return ext.ContinueGroups
				}

				_, err = b.RestrictChatMember(chat.Id, user.Id(), gotgbot.ChatPermissions{CanSendMessages: false}, nil)
				if err != nil {
					log.Error(err)
					return err
				}

				mutedUserMsg, mutedUserErr := tr.GetStringWithError("strings."+m.moduleName+".bl_watcher.muted_user")
				if mutedUserErr != nil {
					log.Errorf("[blacklists] missing translation for key: %v", mutedUserErr)
					mutedUserMsg = "Muted %s due to %s"
				}

				_, err = msg.Reply(b,
					fmt.Sprintf(mutedUserMsg, helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, i)),
					helpers.Shtml())
				if err != nil {
					log.Error(err)
					return err
				}
			case "ban":
				// ban anonymous channels as well
				if user.IsAnonymousChannel() {
					_, err = b.BanChatSenderChat(chat.Id, user.Id(), nil)
				} else {
					_, err = b.BanChatMember(chat.Id, user.Id(), nil)
				}
				if err != nil {
					log.Error(err)
					return err
				}

				bannedUserMsg, bannedUserErr := tr.GetStringWithError("strings."+m.moduleName+".bl_watcher.banned_user")
				if bannedUserErr != nil {
					log.Errorf("[blacklists] missing translation for key: %v", bannedUserErr)
					bannedUserMsg = "Banned %s due to %s"
				}

				_, err = msg.Reply(b,
					fmt.Sprintf(bannedUserMsg, helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, i)),
					helpers.Shtml())
				if err != nil {
					log.Error(err)
					return err
				}
			case "kick":
				// don't work on anonymous channels
				if user.IsAnonymousChannel() {
					return ext.ContinueGroups
				}

				_, err = b.BanChatMember(chat.Id, user.Id(), nil)
				if err != nil {
					log.Error(err)
					return err
				}

				kickedUserMsg, kickedUserErr := tr.GetStringWithError("strings."+m.moduleName+".bl_watcher.kicked_user")
				if kickedUserErr != nil {
					log.Errorf("[blacklists] missing translation for key: %v", kickedUserErr)
					kickedUserMsg = "Kicked %s due to %s"
				}

				_, err = msg.Reply(b,
					fmt.Sprintf(kickedUserMsg, helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, i)),
					helpers.Shtml())
				if err != nil {
					log.Error(err)
					return err
				}

				// unban the member
				time.Sleep(3 * time.Second)
				_, err = chat.UnbanMember(b, user.Id(), nil)
				if err != nil {
					log.Error(err)
					return err
				}
			case "warn":
				// don't work on anonymous channels
				if user.IsAnonymousChannel() {
					return ext.ContinueGroups
				}

				err = warnsModule.warnThisUser(b, ctx, user.Id(), fmt.Sprintf(blSettings.Reason, i), "warn")
				if err != nil {
					log.Error(err)
					return err
				}
			}
			break
		}
	}

	return ext.ContinueGroups
}

/*
LoadBlacklists registers all blacklist-related command handlers with the dispatcher.

Enables the blacklists module and adds handlers for blacklist management and enforcement.
*/
func LoadBlacklists(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	blacklistsModule.cfg = cfg

	HelpModule.AbleMap.Store(blacklistsModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("blacklists", blacklistsModule.listBlacklists))
	misc.AddCmdToDisableable("blacklists")
	dispatcher.AddHandler(handlers.NewCommand("addblacklist", blacklistsModule.addBlacklist))
	dispatcher.AddHandler(handlers.NewCommand("blacklist", blacklistsModule.addBlacklist))
	dispatcher.AddHandler(handlers.NewCommand("rmblacklist", blacklistsModule.removeBlacklist))
	dispatcher.AddHandler(handlers.NewCommand("blaction", blacklistsModule.setBlacklistAction))
	dispatcher.AddHandler(handlers.NewCommand("blacklistaction", blacklistsModule.setBlacklistAction))
	cmdDecorator.MultiCommand(dispatcher, []string{"remallbl", "rmallbl"}, blacklistsModule.rmAllBlacklists)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("rmAllBlacklist"), blacklistsModule.buttonHandler))
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.Text, blacklistsModule.blacklistWatcher), blacklistsModule.handlerGroup)
}
