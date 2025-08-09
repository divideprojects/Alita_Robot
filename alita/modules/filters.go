package modules

import (
	"fmt"
	"html"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"

	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/keyword_matcher"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

var filtersModule = moduleStruct{
	moduleName:          "Filters",
	overwriteFiltersMap: make(map[string]overwriteFilter),
	handlerGroup:        9,
}

/*
	Used to add a filter to a specific keyword in chat!

# Connection - true, true

Only admin can add new filters in the chat
*/
// addFilter creates a new filter with a keyword trigger and response content.
// Only admins can add filters. Supports text, media, and buttons with a limit of 150 filters per chat.
func (m moduleStruct) addFilter(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()

	// check permission
	if !chat_status.CanUserChangeInfo(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	filtersNum := db.CountFilters(chat.Id)
	if filtersNum >= 150 {
		_, err := msg.Reply(b, tr.Message("filters_limit_exceeded", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	if msg.ReplyToMessage != nil && len(args) <= 1 {
		_, err := msg.Reply(b, tr.Message("filters_provide_keyword", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if len(args) <= 2 && msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, tr.Message("filters_invalid_filter", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	filterWord, fileid, text, dataType, buttons, _, _, _, _, _, _, errorMsg := helpers.GetNoteAndFilterType(msg, true)
	if dataType == -1 {
		_, err := msg.Reply(b, errorMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	filterWord = strings.ToLower(filterWord) // convert string to it's lower form

	if db.DoesFilterExists(chat.Id, filterWord) {
		m.overwriteFiltersMap[fmt.Sprint(filterWord, "_", chat.Id)] = overwriteFilter{
			filterWord: filterWord,
			text:       text,
			fileid:     fileid,
			buttons:    buttons,
			dataType:   dataType,
		}
		_, err := msg.Reply(b,
			tr.Message("filters_filter_exists_overwrite", nil),
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         "Yes",
								CallbackData: "filters_overwrite." + filterWord,
							},
							{
								Text:         "No",
								CallbackData: "filters_overwrite.cancel",
							},
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

	go db.AddFilter(chat.Id, filterWord, text, fileid, buttons, dataType)

	_, err := msg.Reply(b, tr.Message("filters_filter_added", i18n.Params{"keyword": filterWord}), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Used to remove a filter to a specific keyword in chat!

# Connection - true, true

Only admin can remove filters in the chat
*/
// rmFilter removes an existing filter by its keyword trigger.
// Only admins can remove filters. Requires the exact filter keyword as argument.
func (moduleStruct) rmFilter(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		_, err := msg.Reply(b, tr.Message("filters_no_keyword_provided", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	} else {

		filterWord, _ := extraction.ExtractQuotes(strings.Join(args, " "), true, true)

		if !string_handling.FindInStringSlice(db.GetFiltersList(chat.Id), strings.ToLower(filterWord)) {
			_, err := msg.Reply(b, tr.Message("filters_filter_not_found", i18n.Params{"keyword": filterWord}), helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			go db.RemoveFilter(chat.Id, strings.ToLower(filterWord))
			_, err := msg.Reply(b, tr.Message("filters_filter_removed", i18n.Params{"keyword": filterWord}), helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return ext.EndGroups
}

/*
	Used to view all filters in the chat!

# Connection - false, true

Any user can view users in a chat
*/
// filtersList displays all active filter keywords in the current chat.
// Any user can view the list of available filters with their trigger keywords.
func (moduleStruct) filtersList(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "filters") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, false, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat

	var replyMsgId int64

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	filterKeys := db.GetFiltersList(chat.Id)
	info := tr.Message("filters_no_filters", nil)
	newFilterKeys := make([]string, 0)

	for _, fkey := range filterKeys {
		newFilterKeys = append(newFilterKeys, fmt.Sprintf("<code>%s</code>", html.EscapeString(fkey)))
	}

	if len(newFilterKeys) > 0 {
		filterList := strings.Join(newFilterKeys, "\n - ")
		info = tr.Message("filters_list_filters", i18n.Params{"filters": "\n - " + filterList})
	}

	_, err := msg.Reply(b,
		info,
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                replyMsgId,
				AllowSendingWithoutReply: true,
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Used to remove all filters from the current chat

Only owner can remove all filters from the chat
*/
// rmAllFilters removes all filters from the current chat with confirmation.
// Only chat owners can use this command. Shows confirmation buttons before deletion.
func (moduleStruct) rmAllFilters(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))
	filterKeys := db.GetFiltersList(chat.Id)

	if len(filterKeys) == 0 {
		_, err := msg.Reply(b, tr.Message("filters_no_filters", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	if chat_status.RequireUserOwner(b, ctx, chat, user.Id, false) {
		_, err := msg.Reply(b, tr.Message("filters_confirm_remove_all", nil),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{Text: "Yes", CallbackData: "rmAllFilters.yes"},
							{Text: "No", CallbackData: "rmAllFilters.no"},
						},
					},
				},
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

// CallbackQuery handler for rmAllFilters
// filtersButtonHandler handles callback queries for filter-related button interactions.
// Processes confirmation dialogs for removing all filters from a chat.
func (moduleStruct) filtersButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// permission checks
	if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	response := args[1]
	var helpText string

	switch response {
	case "yes":
		db.RemoveAllFilters(chat.Id)
		helpText = tr.Message("filters_all_removed", nil)
	case "no":
		helpText = tr.Message("filters_remove_all_cancelled", nil)
	}

	_, _, err := query.Message.EditText(b,
		helpText,
		nil,
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

// CallbackQuery handler for filters_overwite. query
// filterOverWriteHandler handles callback queries for filter overwrite confirmations.
// Processes admin decisions when attempting to overwrite existing filter keywords.
func (m moduleStruct) filterOverWriteHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From
	chat := ctx.EffectiveChat
	tr := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// permission checks
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	filterWord := args[1]
	filterWordKey := fmt.Sprint(filterWord, "_", chat.Id)
	var helpText string
	filterData := m.overwriteFiltersMap[filterWordKey]

	if db.DoesFilterExists(chat.Id, filterWord) {
		db.RemoveFilter(chat.Id, filterWord)
		db.AddFilter(chat.Id, filterData.filterWord, filterData.text, filterData.fileid, filterData.buttons, filterData.dataType)
		delete(m.overwriteFiltersMap, filterWordKey) // delete the key to make map clear
		helpText = tr.Message("filters_filter_overwritten", nil)
	} else {
		helpText = tr.Message("filters_overwrite_cancelled", nil)
	}

	_, _, err := query.Message.EditText(b,
		helpText,
		nil,
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
	Watchers for filter

Replies with appropriate data to the filter.
*/
// filtersWatcher monitors incoming messages for filter keyword matches.
// Automatically responds with filter content when keywords are detected in messages.
func (moduleStruct) filtersWatcher(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	// Use optimized cached query to fetch all filters at once (no N+1 query)
	optQueries := db.GetOptimizedQueries()
	allFilters, err := optQueries.GetChatFiltersCached(chat.Id)
	if err != nil {
		log.WithField("chatId", chat.Id).WithError(err).Error("Failed to get chat filters")
		return ext.ContinueGroups
	}

	if len(allFilters) == 0 {
		return ext.ContinueGroups
	}

	// Build keyword list for Aho-Corasick matching
	filterKeys := make([]string, len(allFilters))
	filterMap := make(map[string]*db.ChatFilters, len(allFilters))
	for i, filter := range allFilters {
		filterKeys[i] = filter.KeyWord
		filterMap[filter.KeyWord] = filter
	}

	// Use Aho-Corasick for efficient multi-pattern matching
	cache := keyword_matcher.GetGlobalCache()
	matcher := cache.GetOrCreateMatcher(chat.Id, filterKeys)

	// Check for any filter match first
	if !matcher.HasMatch(msg.Text) {
		return ext.ContinueGroups
	}

	// Get all matches to handle them individually
	matches := matcher.FindMatches(msg.Text)
	if len(matches) == 0 {
		return ext.ContinueGroups
	}

	// Process first match (same behavior as before)
	firstMatch := matches[0]
	i := firstMatch.Pattern

	// Check for noformat pattern using simpler string matching
	noformatPattern := i + " noformat"
	noformatMatch := strings.Contains(strings.ToLower(msg.Text), strings.ToLower(noformatPattern))

	// Get filter data from pre-loaded map (no additional DB query)
	filtData, exists := filterMap[i]
	if !exists {
		return ext.ContinueGroups
	}

	if noformatMatch {
		// check if user is admin or not
		if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
			return ext.EndGroups
		}

		// Reverse notedata
		filtData.FilterReply = helpers.ReverseHTML2MD(filtData.FilterReply)

		// show the buttons back as text
		filtData.FilterReply += helpers.RevertButtons(filtData.Buttons)

		// using true as last argument to prevent the message from being formatted
		var err error
		_, err = helpers.FiltersEnumFuncMap[filtData.MsgType](
			b,
			ctx,
			*filtData,
			&gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil},
			msg.MessageId,
			true,
			filtData.NoNotif,
		)
		if err != nil {
			log.Error(err)
			return err
		}

	} else {
		var err error
		_, err = helpers.SendFilter(b, ctx, filtData, msg.MessageId)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.ContinueGroups
}

// LoadFilters registers all filter-related handlers with the dispatcher.
// Sets up commands for managing filters and the message watcher for automatic responses.
func LoadFilters(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(filtersModule.moduleName, true)

	HelpModule.helpableKb[filtersModule.moduleName] = [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         "Formatting",
				CallbackData: fmt.Sprintf("helpq.%s", "Formatting"),
			},
		},
	} // Adds Formatting kb button to Filters Menu
	dispatcher.AddHandler(handlers.NewCommand("filter", filtersModule.addFilter))
	dispatcher.AddHandler(handlers.NewCommand("addfilter", filtersModule.addFilter))
	dispatcher.AddHandler(handlers.NewCommand("stop", filtersModule.rmFilter))
	dispatcher.AddHandler(handlers.NewCommand("addfilrmfilterter", filtersModule.rmFilter))
	dispatcher.AddHandler(handlers.NewCommand("filters", filtersModule.filtersList))
	misc.AddCmdToDisableable("filters")
	dispatcher.AddHandler(handlers.NewCommand("stopall", filtersModule.rmAllFilters))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("rmAllFilters"), filtersModule.filtersButtonHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("filters_overwrite."), filtersModule.filterOverWriteHandler))
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.Text, filtersModule.filtersWatcher), filtersModule.handlerGroup)
}
