package modules

import (
	"fmt"
	"html"
	"regexp"
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

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

/*
filtersModule provides logic for managing keyword-based filters in group chats.

Implements commands to add, remove, list, and configure filters and their actions.
*/
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
/*
addFilter adds a filter to a specific keyword in the chat.

Only admins can add new filters. Handles filter limits, overwriting, and input validation.
Connection: true, true
*/
func (m moduleStruct) addFilter(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	msg := ctx.EffectiveMessage
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
		_, err := msg.Reply(b,
			fmt.Sprint("Filters limit exceeded, a group can only have maximum 150 filters!\n",
				"This limitation is due to bot running free without any donations by users."),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	if msg.ReplyToMessage != nil && len(args) <= 1 {
		_, err := msg.Reply(b, tr.GetString("Notes.errors.no_keyword_reply"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if len(args) <= 2 && msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, tr.GetString("Filters.errors.invalid_filter"), helpers.Shtml())
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
			"Filter already exists!\nDo you want to overwrite it?",
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         tr.GetString("strings.CommonStrings.buttons.yes"),
								CallbackData: "filters_overwrite." + filterWord,
							},
							{
								Text:         tr.GetString("strings.CommonStrings.buttons.no"),
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

	_, err := msg.Reply(b, fmt.Sprintf(tr.GetString("Filters.add.success"), filterWord), helpers.Shtml())
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
/*
rmFilter removes a filter for a specific keyword in the chat.

Only admins can remove filters. Handles input validation and replies with the result.
Connection: true, true
*/
func (moduleStruct) rmFilter(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
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
		_, err := msg.Reply(b, tr.GetString("Filters.remove.no_word_specified"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	} else {

		filterWord, _ := extraction.ExtractQuotes(strings.Join(args, " "), true, true)

		if !string_handling.FindInStringSlice(db.GetFiltersList(chat.Id), strings.ToLower(filterWord)) {
			_, err := msg.Reply(b, tr.GetString("Filters.errors.does_not_exist"), helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			go db.RemoveFilter(chat.Id, strings.ToLower(filterWord))
			_, err := msg.Reply(b, fmt.Sprintf(tr.GetString("Filters.remove.success"), filterWord), helpers.Shtml())
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
/*
filtersList lists all filters in the chat.

Anyone can view the filters. Replies with the current list or a message if none exist.
Connection: false, true
*/
func (moduleStruct) filtersList(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
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
	info := tr.GetString("Filters.remove_all.no_filters")
	newFilterKeys := make([]string, 0)

	for _, fkey := range filterKeys {
		newFilterKeys = append(newFilterKeys, fmt.Sprintf("<code>%s</code>", html.EscapeString(fkey)))
	}

	if len(newFilterKeys) > 0 {
		info = "These are the current filters in this Chat:"
		info += "\n - " + strings.Join(newFilterKeys, "\n - ")
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
/*
rmAllFilters removes all filters from the current chat.

Only the chat owner can use this command to clear all filters.
*/
func (moduleStruct) rmAllFilters(b *gotgbot.Bot, ctx *ext.Context) error {
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	filterKeys := db.GetFiltersList(chat.Id)

	if len(filterKeys) == 0 {
		_, err := msg.Reply(b, "There are no filters in this chat!", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	if chat_status.RequireUserOwner(b, ctx, chat, user.Id, false) {
		_, err := msg.Reply(b, tr.GetString("Filters.remove_all.confirm"),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{Text: tr.GetString("strings.CommonStrings.buttons.yes"), CallbackData: "rmAllFilters.yes"},
							{Text: tr.GetString("strings.CommonStrings.buttons.no"), CallbackData: "rmAllFilters.no"},
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
/*
filtersButtonHandler handles callback queries for removing all filters.

Processes the owner's confirmation and removes all filters if confirmed.
*/
func (moduleStruct) filtersButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	user := query.From
	chat := ctx.EffectiveChat

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
		helpText = "Removed all Filters from this Chat ✅"
	case "no":
		helpText = "Cancelled removing all Filters from this Chat ❌"
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
/*
filterOverWriteHandler handles callback queries for overwriting an existing filter.

Allows admins to confirm and overwrite an existing filter with new data.
*/
func (m moduleStruct) filterOverWriteHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	user := query.From
	chat := ctx.EffectiveChat

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
		helpText = "Filter has been overwritten successfully ✅"
	} else {
		helpText = "Cancelled overwritting of filter ❌"
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
/*
filtersWatcher monitors messages for filtered keywords and replies with the appropriate data.

Handles both formatted and unformatted responses, and enforces admin-only access for noformat requests.
*/
func (moduleStruct) filtersWatcher(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	var err error

	filterKeys := db.GetFiltersList(chat.Id)

	for _, i := range filterKeys {
		match, _ := regexp.MatchString(fmt.Sprintf(`(\b|\s)%s\b`, i), strings.ToLower(msg.Text))
		noformatMatch, _ := regexp.MatchString(fmt.Sprintf("%s noformat", i), strings.ToLower(msg.Text))

		if match {
			filtData := db.GetFilter(chat.Id, i)

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
				_, err = helpers.FiltersEnumFuncMap[filtData.MsgType](
					b,
					ctx,
					*filtData,
					&gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil},
					msg.MessageId,
					true,
					filtData.NoNotif,
				)

			} else {
				_, err = helpers.SendFilter(b, ctx, filtData, msg.MessageId)
			}

			if err != nil {
				log.Error(err)
				return err
			}

			return ext.ContinueGroups
		}
	}

	return ext.ContinueGroups
}

/*
LoadFilters registers all filter-related command handlers with the dispatcher.

Enables the filters module and adds handlers for filter management and enforcement.
*/
func LoadFilters(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(filtersModule.moduleName, true)

	HelpModule.helpableKb[filtersModule.moduleName] = [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         tr.GetString("strings.Filters.formatting"),
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
