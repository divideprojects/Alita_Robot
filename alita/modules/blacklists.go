package modules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
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

// blacklistsModule provides blacklist management logic for group chats.
//
// Implements commands to add, remove, list, and configure blacklists and their actions.
// Includes optimized regex caching for high-performance blacklist matching.
var blacklistsModule = moduleStruct{
	moduleName:   "Blacklists",
	handlerGroup: 7,
	// Initialize regex cache maps for blacklists
	filterRegexCache:    make(map[int64]*regexp.Regexp),
	filterKeywordsCache: make(map[int64][]string),
}

// Mutex to protect blacklist regex cache from concurrent access
var blacklistCacheMutex sync.RWMutex

// buildBlacklistRegex creates an optimized regex pattern that matches any of the blacklist triggers.
// Returns nil if no triggers or if regex compilation fails.
func buildBlacklistRegex(triggers []string) *regexp.Regexp {
	if len(triggers) == 0 {
		return nil
	}

	// Escape special regex characters in triggers and build pattern
	escapedTriggers := make([]string, len(triggers))
	for i, trigger := range triggers {
		escapedTriggers[i] = regexp.QuoteMeta(trigger)
	}

	// Create pattern that matches any trigger with word boundaries
	pattern := fmt.Sprintf(`(\b|\s)(%s)\b`, strings.Join(escapedTriggers, "|"))

	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Errorf("[Blacklists] Failed to compile regex pattern: %v", err)
		return nil
	}

	return regex
}

// getOrBuildBlacklistRegex retrieves cached regex or builds new one if cache is invalid.
// Thread-safe with read-write locking for optimal performance.
func getOrBuildBlacklistRegex(chatId int64, currentTriggers []string) *regexp.Regexp {
	blacklistCacheMutex.RLock()
	cachedRegex, regexExists := blacklistsModule.filterRegexCache[chatId]
	cachedTriggers, triggersExist := blacklistsModule.filterKeywordsCache[chatId]
	blacklistCacheMutex.RUnlock()

	// Check if cache is valid (regex exists and triggers haven't changed)
	if regexExists && triggersExist && slicesEqual(cachedTriggers, currentTriggers) {
		return cachedRegex
	}

	// Cache is invalid, rebuild regex
	newRegex := buildBlacklistRegex(currentTriggers)

	blacklistCacheMutex.Lock()
	blacklistsModule.filterRegexCache[chatId] = newRegex
	blacklistsModule.filterKeywordsCache[chatId] = make([]string, len(currentTriggers))
	copy(blacklistsModule.filterKeywordsCache[chatId], currentTriggers)
	blacklistCacheMutex.Unlock()

	return newRegex
}

// invalidateBlacklistCache removes cached regex for a chat when blacklists are modified.
// Should be called whenever blacklists are added, removed, or modified.
func invalidateBlacklistCache(chatId int64) {
	blacklistCacheMutex.Lock()
	delete(blacklistsModule.filterRegexCache, chatId)
	delete(blacklistsModule.filterKeywordsCache, chatId)
	blacklistCacheMutex.Unlock()
}

// findMatchingBlacklistTrigger uses the optimized regex to find which trigger matched.
// Returns the matched trigger if found.
func findMatchingBlacklistTrigger(regex *regexp.Regexp, text string, triggers []string) string {
	lowerText := strings.ToLower(text)

	// First check if any trigger matches
	if !regex.MatchString(lowerText) {
		return ""
	}

	// Find which specific trigger matched (fallback to individual checks)
	for _, trigger := range triggers {
		triggerPattern := fmt.Sprintf(`(\b|\s)%s\b`, regexp.QuoteMeta(trigger))
		if matched, _ := regexp.MatchString(triggerPattern, lowerText); matched {
			return trigger
		}
	}

	return ""
}

// addBlacklist adds one or more blacklist words to the group.
//
// Used to add a blacklist to group!
// Connection - true, true
// Admin can add a blacklist to the chat
//
// Checks permissions, updates the blacklist in the database, and replies with the result.
// Connection: true, true
func (moduleStruct) addBlacklist(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, err := msg.Reply(b, tr.GetString("strings.blacklists.blacklist.give_bl_word"), helpers.Shtml())
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

		// Invalidate cache after adding blacklists
		if len(newBlacklist) > 0 {
			invalidateBlacklistCache(chat.Id)
		}

		if len(alreadyBlacklisted) >= 1 {
			text += tr.GetString("strings.blacklists.blacklist.already_blacklisted") + fmt.Sprintf("\n - %s\n\n", strings.Join(alreadyBlacklisted, "\n - "))
		}
		if len(newBlacklist) >= 1 {
			text += tr.GetString("strings.blacklists.blacklist.added_bl") + fmt.Sprintf("\n - %s\n\n", strings.Join(newBlacklist, "\n - "))
		}

		_, err := msg.Reply(b, text, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

// removeBlacklist removes one or more blacklist words from the group.
//
// Used to remove a blacklist from group!
// Connection - true, true
// Admin can add a blacklist to the chat
//
// Checks permissions, updates the blacklist in the database, and replies with the result.
// Connection: true, true
func (moduleStruct) removeBlacklist(b *gotgbot.Bot, ctx *ext.Context) error {
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
		_, err := msg.Reply(b, tr.GetString("strings.blacklists.unblacklist.give_bl_word"), helpers.Shtml())
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
			_, err := msg.Reply(b, "strings.blacklists.unblacklist.no_removed_bl",
				nil)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			// Invalidate cache after removing blacklists
			invalidateBlacklistCache(chat.Id)

			_, err := msg.Reply(b,
				fmt.Sprintf(tr.GetString("strings.blacklists.unblacklist.removed_bl"), strings.Join(removedBlacklists, ", ")),
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

// listBlacklists lists all blacklist words in the group.
//
// Used to list all blacklists of a group!
// Connection - false, true
// Anyone can view blacklists in group
//
// Anyone can view the blacklist. Replies with the current list or a message if none exist.
// Connection: false, true
func (moduleStruct) listBlacklists(b *gotgbot.Bot, ctx *ext.Context) error {
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
		blacklistsText = tr.GetString("strings.blacklists.ls_bl.list_bl") + blacklistsText
	} else {
		blacklistsText = tr.GetString("strings.blacklists.ls_bl.no_blacklisted")
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

// setBlacklistAction sets the action to take when a blacklist word is triggered.
//
// Used to set mode for blacklists in chat
// Connection - true, true
// Admin with restriction permission can set blacklist action in group out of - kick, ban, mute
//
// Admins can set the action to "mute", "kick", "warn", "ban", or "none".
// Connection: true, true
func (moduleStruct) setBlacklistAction(b *gotgbot.Bot, ctx *ext.Context) error {
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
		rMsg = fmt.Sprintf(tr.GetString("strings.blacklists.set_bl_action.current_mode"), currAction)
	} else if len(args) == 1 {
		action := strings.ToLower(args[0])
		if string_handling.FindInStringSlice([]string{"mute", "kick", "warn", "ban", "none"}, action) {
			rMsg = fmt.Sprintf(tr.GetString("strings.blacklists.set_bl_action.changed_mode"), action)
			go db.SetBlacklistAction(chat.Id, action)
		} else {
			rMsg = tr.GetString("strings.blacklists.set_bl_action.choose_correct_option")
		}
	} else {
		rMsg = tr.GetString("strings.blacklists.set_bl_action.choose_correct_option")
	}
	_, err := msg.Reply(b, rMsg, helpers.Smarkdown())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// rmAllBlacklists removes all blacklist words from the group.
//
// Used to remove all blacklists from a group
// Only chat creator can use this command to remove all blacklists at once from the current chat
//
// Only the chat creator can use this command to clear the blacklist.
func (moduleStruct) rmAllBlacklists(b *gotgbot.Bot, ctx *ext.Context) error {
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

	_, err := msg.Reply(b, tr.GetString("strings.blacklists.rm_all_bl.ask"),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Yes", CallbackData: "rmAllBlacklist.true"},
						{Text: "No", CallbackData: "rmAllBlacklist.false"},
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

// buttonHandler handles callback queries for removing all blacklists.
//
// Callback Handler for rmallblacklist
// Processes the creator's confirmation and removes all blacklist words if confirmed.
func (moduleStruct) buttonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
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
	case "true":
		go db.RemoveAllBlacklist(query.Message.GetChat().Id)
		invalidateBlacklistCache(query.Message.GetChat().Id) // Invalidate cache after removing all blacklists
		helpText = tr.GetString("strings.blacklists.rm_all_bl.button_handler.true")
	case "false":
		helpText = tr.GetString("strings.blacklists.rm_all_bl.button_handler.false")
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

// blacklistWatcher monitors messages for blacklisted words and enforces the configured action.
//
// Blacklist watcher
// Watcher for blacklisted words, if any of the sentence contains the word, it will remove and use the appropriate action
//
// Deletes messages containing blacklisted words and applies the configured action (mute, ban, kick, warn) to the user.
// Uses optimized regex caching for high-performance pattern matching.
func (moduleStruct) blacklistWatcher(b *gotgbot.Bot, ctx *ext.Context) error {
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
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Use optimized regex matching
	regex := getOrBuildBlacklistRegex(chat.Id, blSettings.Triggers)

	if regex == nil {
		return ext.ContinueGroups
	}

	matchedTrigger := findMatchingBlacklistTrigger(regex, msg.Text, blSettings.Triggers)

	if matchedTrigger != "" {
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
				fmt.Sprintf(tr.GetString("strings.blacklists.bl_watcher.muted_user"), helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, matchedTrigger)),
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
				fmt.Sprintf(tr.GetString("strings.blacklists.bl_watcher.banned_user"), helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, matchedTrigger)),
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
				fmt.Sprintf(tr.GetString("strings.blacklists.bl_watcher.kicked_user"), helpers.MentionHtml(user.Id(), user.Name()), fmt.Sprintf(blSettings.Reason, matchedTrigger)),
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

			err = warnsModule.warnThisUser(b, ctx, user.Id(), fmt.Sprintf(blSettings.Reason, matchedTrigger), "warn")
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return ext.ContinueGroups
}

// LoadBlacklists registers all blacklist-related command handlers with the dispatcher.
//
// This function enables the blacklists module and adds handlers for blacklist
// management and enforcement. The module provides automatic word/phrase filtering
// with configurable actions and optimized regex matching for performance.
//
// Registered commands:
//   - /blacklists: Lists all blacklisted words and phrases
//   - /addblacklist, /blacklist: Adds words/phrases to the blacklist
//   - /rmblacklist: Removes words/phrases from the blacklist
//   - /blaction, /blacklistaction: Sets action for blacklist violations
//   - /remallbl, /rmallbl: Removes all blacklisted words from chat
//
// The module automatically monitors all text messages in group 7 handler priority
// and applies blacklist filters based on configured triggers. When blacklisted
// content is detected, appropriate actions are taken based on chat settings.
//
// Features:
//   - Optimized regex compilation with caching for performance
//   - Case-insensitive word/phrase matching with word boundaries
//   - Configurable actions (delete, warn, mute, ban, kick)
//   - Bulk blacklist management operations
//   - Thread-safe regex cache management
//   - Support for complex patterns and phrases
//
// Requirements:
//   - Bot must be admin to delete blacklisted messages
//   - User must be admin to manage blacklists
//   - Module supports remote configuration via connections
//   - Integrates with warning and enforcement systems
//
// The blacklists system provides efficient content filtering with minimal
// performance impact through regex caching and optimized pattern matching.
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
