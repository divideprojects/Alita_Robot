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

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/cmdDecorator"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

var blacklistsModule = moduleStruct{
	moduleName:   "Blacklists",
	handlerGroup: 7,
}

/*
	Used to add a blacklist to group!

Connection - true, true
Admin can add a blacklist to the chat
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
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

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
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".blacklist.give_bl_word"), helpers.Shtml())
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
			text += tr.GetString("strings."+m.moduleName+".blacklist.already_blacklisted") + fmt.Sprintf("\n - %s\n\n", strings.Join(alreadyBlacklisted, "\n - "))
		}
		if len(newBlacklist) >= 1 {
			text += tr.GetString("strings."+m.moduleName+".blacklist.added_bl") + fmt.Sprintf("\n - %s\n\n", strings.Join(newBlacklist, "\n - "))
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
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

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
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".unblacklist.give_bl_word"), helpers.Shtml())
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
			_, err := msg.Reply(b, fmt.Sprint("strings."+m.moduleName+".unblacklist.no_removed_bl"),
				nil)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			_, err := msg.Reply(b,
				fmt.Sprintf(tr.GetString("strings."+m.moduleName+".unblacklist.removed_bl"), strings.Join(removedBlacklists, ", ")),
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
func (m moduleStruct) listBlacklists(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
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
		blacklistsText = tr.GetString("strings."+m.moduleName+".ls_bl.list_bl") + blacklistsText
	} else {
		blacklistsText = tr.GetString("strings." + m.moduleName + ".ls_bl.no_blacklisted")
	}

	_, err := msg.Reply(b,
		blacklistsText,
		&gotgbot.SendMessageOpts{
			ReplyToMessageId:         replyMsgId,
			AllowSendingWithoutReply: true,
			ParseMode:                helpers.HTML,
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
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

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
		rMsg = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".set_bl_action.current_mode"), currAction)
	} else if len(args) == 1 {
		action := strings.ToLower(args[0])
		if string_handling.FindInStringSlice([]string{"mute", "kick", "warn", "ban", "none"}, action) {
			rMsg = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".set_bl_action.changed_mode"), action)
			go db.SetBlacklistAction(chat.Id, action)
		} else {
			rMsg = tr.GetString("strings." + m.moduleName + ".set_bl_action.choose_correct_option")
		}
	} else {
		rMsg = tr.GetString("strings." + m.moduleName + ".set_bl_action.choose_correct_option")
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
func (m moduleStruct) rmAllBlacklists(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserOwner(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".rm_all_bl.ask"),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Yes", CallbackData: "rmAllBlacklist.yes"},
						{Text: "No", CallbackData: "rmAllBlacklist.no"},
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
func (m moduleStruct) buttonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	user := query.From
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// permission checks
	if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	creatorAction := args[1]
	var helpText string

	switch creatorAction {
	case "yes":
		go db.RemoveAllBlacklist(query.Message.Chat.Id)
		helpText = tr.GetString("strings." + m.moduleName + ".rm_all_bl.button_handler.yes")
	case "no":
		helpText = tr.GetString("strings." + m.moduleName + ".rm_all_bl.button_handler.yes")
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
func (m moduleStruct) blacklistWatcher(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender
	if user.IsAnonymousAdmin() {
		return ext.ContinueGroups
	}

	// skip admins and creator + approved users and anonymous channel
	if !user.IsAnonymousChannel() && (chat_status.IsUserAdmin(b, chat.Id, user.Id())) {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	blSettings := db.GetBlacklistSettings(chat.Id)
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

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

				_, err = msg.Reply(b,
					fmt.Sprintf(tr.GetString("strings."+m.moduleName+".bl_watcher.muted_user"), helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, i)),
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

				_, err = msg.Reply(b,
					fmt.Sprintf(tr.GetString("strings."+m.moduleName+".bl_watcher.banned_user"), helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, i)),
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

				_, err = msg.Reply(b,
					fmt.Sprintf(tr.GetString("strings."+m.moduleName+".bl_watcher.kicked_user"), helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, i)),
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

func LoadBlacklists(dispatcher *ext.Dispatcher) {
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
