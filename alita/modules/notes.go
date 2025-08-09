package modules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"

	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/cmdDecorator"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

var notesModule = moduleStruct{
	moduleName:        "Notes",
	overwriteNotesMap: make(map[string]overwriteNote),
}

// addNote handles the /save command to create new notes
// with support for various media types and formatting options.
func (m moduleStruct) addNote(b *gotgbot.Bot, ctx *ext.Context) error {
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	args := ctx.Args()

	// check permission
	if !chat_status.CanUserChangeInfo(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	noteString := "Saved Note <b>%s</b>!\nGet it with <code>#%s</code> or <code>/get %s</code>."

	if msg.ReplyToMessage != nil && len(args) <= 1 {
		_, err := msg.Reply(b, "Please give a keyword to reply to!", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if len(args) <= 2 && msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, "Invalid Note!", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	noteWord, fileid, text, dataType, buttons, pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif, errorMsg := helpers.GetNoteAndFilterType(msg, false)
	if dataType == -1 && errorMsg != "" {
		_, err := msg.Reply(b, errorMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// if user specifies both noprivate and private, the note will be sent to default.
	// If privatenotes is enabled, the private else group
	if grpOnly && pvtOnly {
		grpOnly, pvtOnly = false, false
		noteString += "\n\n<b>Note:</b> This note will be sent to default setting of group notes, " +
			"because it has both <code>{private}</code> and <code>{noprivate}</code>."
	}

	noteWord = strings.ToLower(noteWord)

	// check if note already exists or not
	if db.DoesNoteExists(chat.Id, noteWord) {
		noteWordMapKey := fmt.Sprintf("%d_%s", chat.Id, noteWord)
		m.overwriteNotesMap[noteWordMapKey] = overwriteNote{
			noteWord:    noteWord,
			text:        text,
			fileId:      fileid,
			buttons:     buttons,
			dataType:    dataType,
			pvtOnly:     pvtOnly,
			grpOnly:     grpOnly,
			adminOnly:   adminOnly,
			webPrev:     webPrev,
			isProtected: isProtected,
			noNotif:     noNotif,
		}
		_, err := msg.Reply(b,
			"Note already exists!\nDo you want to overwrite it?",
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         "Yes",
								CallbackData: fmt.Sprintf("notes.overwrite.%s", noteWordMapKey),
							},
							{
								Text:         "No",
								CallbackData: "notes.overwrite.cancel",
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

	go db.AddNote(chat.Id, noteWord, text, fileid, buttons, dataType, pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif)

	_, err := msg.Reply(b, fmt.Sprintf(noteString, noteWord, noteWord, noteWord), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// rmNote handles the /clear command to remove existing notes
// from the chat, requiring admin permissions.
func (moduleStruct) rmNote(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	if len(args) == 1 {
		_, err := msg.Reply(b, translator.Message("notes_need_keyword_remove", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	noteWord := strings.SplitN(msg.Text, " ", 2)[1] // don't include '/clear' command
	noteWord = strings.TrimLeft(noteWord, "#")

	// check permission
	if !chat_status.CanUserChangeInfo(b, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	// check if note exists in admin notes as well
	if !string_handling.FindInStringSlice(db.GetNotesList(chat.Id, true), strings.ToLower(noteWord)) {
		_, err := msg.Reply(b, translator.Message("notes_not_found", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	noteWord, _ = extraction.ExtractQuotes(noteWord, false, true)

	db.RemoveNote(chat.Id, strings.ToLower(noteWord))

	_, err := msg.Reply(b, translator.Message("notes_removed_success", i18n.Params{
		"name": noteWord,
	}), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// privNote handles the /privnote command to toggle private notes
// setting, controlling whether notes are sent privately or in group.
func (moduleStruct) privNote(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]
	var txt string

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	if len(args) == 1 {
		option := args[0]
		switch option {
		case "on", "yes", "true":
			txt = translator.Message("notes_private_on", nil)
			go db.TooglePrivateNote(chat.Id, true)
		case "off", "no", "false":
			txt = translator.Message("notes_private_off", nil)
			go db.TooglePrivateNote(chat.Id, false)
		default:
			txt = translator.Message("notes_private_invalid_option", nil)
		}
	} else {
		tmp := db.GetNotes(chat.Id).PrivateNotesEnabled()
		if tmp {
			txt = translator.Message("notes_private_status_on", nil)
		} else {
			txt = translator.Message("notes_private_status_off", nil)
		}
	}
	_, err := msg.Reply(b, txt, helpers.Smarkdown())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// notesList handles the /notes command to display all available
// notes in the chat with appropriate access controls.
func (moduleStruct) notesList(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "notes") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, false, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	noteKeys := db.GetNotesList(chat.Id, chat_status.RequireUserAdmin(b, ctx, nil, user.Id, true))
	info := translator.Message("notes_no_notes", nil)

	if len(noteKeys) == 0 {
		_, err := msg.Reply(b, info, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// if user uses the /note command in private chat
	// No matter if privRules are set or not
	if ctx.Message.Chat.Type == "private" {
		// check if want admin notes or not
		admin := chat_status.IsUserAdmin(b, chat.Id, user.Id)
		noteKeys := db.GetNotesList(chat.Id, admin)
		info = translator.Message("notes_current_notes_chat", i18n.Params{
			"chat_title": chat.Title,
		})
		for _, note := range noteKeys {
			info += fmt.Sprintf("\n - <a href='https://t.me/%s?start=note_%d_%s'>%s</a>",
				b.Username, chat.Id, note, note)
		}
		_, err := msg.Reply(b, info, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	privNote := db.GetNotes(chat.Id).PrivateNotesEnabled()
	if privNote {
		_, err := msg.Reply(b, translator.Message("notes_list_private_button", nil),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text: translator.Message("notes_list_private_click", nil),
								Url:  fmt.Sprintf("https://t.me/%s?start=notes_%d", b.Username, chat.Id),
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
	} else {
		info = translator.Message("notes_current_notes", nil) + "\n"
		for _, note := range noteKeys {
			info += fmt.Sprintf(" - <code>#%s</code>\n", note)
		}
		info += "\n" + translator.Message("notes_get_instructions", nil)
		_, err := msg.Reply(b, info, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

// rmAllNotes handles the /clearall command to remove all notes
// from the chat, restricted to chat owners only.
func (moduleStruct) rmAllNotes(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// check notes in adminkeys as well
	noteKeys := db.GetNotesList(chat.Id, true)
	if len(noteKeys) == 0 {
		_, err := msg.Reply(b, translator.Message("notes_no_notes", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	mem, err := chat.GetMember(b, user.Id, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	if mem.MergeChatMember().Status == "creator" {
		_, err := msg.Reply(b, translator.Message("notes_confirm_clear_all", nil),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{Text: "Yes", CallbackData: "rmAllNotes.yes"},
							{Text: "No", CallbackData: "rmAllNotes.no"},
						},
					},
				},
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		_, err := msg.Reply(b, translator.Message("notes_only_creator_can_clear_all", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

// noteOverWriteHandler processes callback queries for note overwrite
// confirmations when adding notes that already exist.
func (m moduleStruct) noteOverWriteHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From

	// permission checks
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	// Get translator for the chat
	translator := i18n.MustNewTranslator("en") // fallback for callback queries

	var helpText string
	args := strings.Split(query.Data, ".")
	noteWordMapKey := args[2]

	switch noteWordMapKey {
	case "cancel":
		helpText = translator.Message("notes_overwrite_cancelled", nil)
	default:
		dataSplit := strings.Split(noteWordMapKey, "_")
		strChatId, noteWord := dataSplit[0], dataSplit[1]
		chatId, _ := strconv.ParseInt(strChatId, 10, 64)
		noteData := m.overwriteNotesMap[noteWordMapKey]
		fmt.Println(strChatId, noteWord, chatId, noteData)
		if db.DoesNoteExists(chatId, noteWord) {
			db.RemoveNote(chatId, noteWord)
			db.AddNote(chatId, noteData.noteWord, noteData.text, noteData.fileId, noteData.buttons, noteData.dataType, noteData.pvtOnly, noteData.grpOnly, noteData.adminOnly, noteData.webPrev, noteData.isProtected, noteData.noNotif)
			delete(m.overwriteNotesMap, noteWordMapKey) // delete the key to make map clear
			helpText = translator.Message("notes_overwrite_success", nil)
		}
	}

	_, _, err := query.Message.EditText(
		b,
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

// notesButtonHandler processes callback queries for the remove all notes
// confirmation dialog, restricted to chat owners.
func (moduleStruct) notesButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From

	// permission checks
	if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	// Get translator for the chat
	translator := i18n.MustNewTranslator("en") // fallback for callback queries

	args := strings.Split(query.Data, ".")
	response := args[1]
	var helpText string

	switch response {
	case "yes":
		db.RemoveAllNotes(query.Message.GetChat().Id)
		helpText = translator.Message("notes_cleared_all", nil)
	case "no":
		helpText = translator.Message("notes_clear_all_cancelled", nil)
	}

	_, _, err := query.Message.EditText(
		b,
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

// notesWatcher monitors messages starting with '#' and automatically
// sends the corresponding note if it exists in the chat.
func (m moduleStruct) notesWatcher(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	var replyMsgId int64
	var err error

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	parseText := strings.ToLower(msg.Text)[1:] // remove '#' from note name
	noteNameArgs := strings.Split(parseText, " ")
	noteName := noteNameArgs[0]
	noformatNote := len(noteNameArgs) == 2 && noteNameArgs[1] == "noformat"

	// if note does not exist, continue groups
	if !string_handling.FindInStringSlice(db.GetNotesList(chat.Id, true), strings.ToLower(noteName)) {
		return ext.EndGroups
	}

	noteData := db.GetNote(chat.Id, noteName)

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// check if notedata is correct or not
	if noteData.NoteContent == "" && noteData.FileID == "" {
		_, err := msg.Reply(b, translator.Message("notes_error_parsing", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// check for admin only notes
	// admin notes follow the group note policy
	if noteData.AdminOnly {
		if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			_, err := msg.Reply(b, translator.Message("notes_admin_only", nil), helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.EndGroups
		}
	}

	if noformatNote {
		err = m.sendNoFormatNote(b, ctx, replyMsgId, noteData)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {

		// chat has private notes enabled or note is private and not group note
		privateNoteOnly := (db.GetNotes(chat.Id).PrivateNotesEnabled() || noteData.PrivateOnly) && !noteData.GroupOnly

		// send private note if private notes is enabled or note is private, and it is not group note
		if privateNoteOnly {
			if ctx.Message.Chat.Type == "private" {
				_, err = helpers.SendNote(b, chat, ctx, noteData, replyMsgId)
			} else {
				_, err = msg.Reply(b,
					translator.Message("notes_click_get_private", i18n.Params{
						"name": noteName,
					}),
					&gotgbot.SendMessageOpts{
						ReplyParameters: &gotgbot.ReplyParameters{
							MessageId:                replyMsgId,
							AllowSendingWithoutReply: true,
						},
						ReplyMarkup: gotgbot.InlineKeyboardMarkup{
							InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
								{
									{
										Text: translator.Message("notes_list_private_click", nil),
										Url:  fmt.Sprintf("https://t.me/%s?start=note_%d_%s", b.Username, chat.Id, noteName),
									},
								},
							},
						},
						ParseMode: helpers.Markdown,
					},
				)
			}
		} else {
			_, err = helpers.SendNote(b, chat, ctx, noteData, replyMsgId)
		}
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// getNotes handles the /get command to retrieve and send
// specific notes by name with format options.
func (m moduleStruct) getNotes(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "get") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, false, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]
	var err error

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	if len(args) == 0 {
		_, err := msg.Reply(b, translator.Message("notes_not_enough_args", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	var replyMsgId int64

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	user := ctx.EffectiveSender.User
	noteName := args[0]

	// check if note exists or not
	if !string_handling.FindInStringSlice(db.GetNotesList(chat.Id, true), strings.ToLower(noteName)) {
		_, err := msg.Reply(b, translator.Message("notes_not_exists", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	noteData := db.GetNote(chat.Id, noteName)

	// check if notedata is correct or not
	if noteData.NoteContent == "" && noteData.FileID == "" {
		_, err := msg.Reply(b, translator.Message("notes_error_parsing", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	// check for admin only notes
	// admin notes follow the group note policy
	if noteData.AdminOnly {
		if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			_, err = msg.Reply(b, translator.Message("notes_admin_only", nil), helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.ContinueGroups
		}
	}

	if len(args) == 2 && strings.ToLower(args[1]) == "noformat" {
		err = m.sendNoFormatNote(b, ctx, replyMsgId, noteData)
	} else {
		// send private note if private notes is enabled or note is private, and it is not group note
		if (db.GetNotes(chat.Id).PrivateNotesEnabled() || noteData.PrivateOnly) && !noteData.GroupOnly {
			_, err = msg.Reply(b,
				translator.Message("notes_click_get_private", i18n.Params{
					"name": noteName,
				}),
				&gotgbot.SendMessageOpts{
					ReplyMarkup: gotgbot.InlineKeyboardMarkup{
						InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
							{
								{
									Text: translator.Message("notes_list_private_click", nil),
									Url:  fmt.Sprintf("https://t.me/%s?start=note_%d_%s", b.Username, chat.Id, noteName),
								},
							},
						},
					},
					ParseMode: helpers.Markdown,
					ReplyParameters: &gotgbot.ReplyParameters{
						MessageId:                replyMsgId,
						AllowSendingWithoutReply: true,
					},
				},
			)
		} else {
			_, err = helpers.SendNote(b, chat, ctx, noteData, replyMsgId)
		}
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// sendNoFormatNote sends a note in raw format without markdown processing,
// showing the original formatting codes, restricted to admins.
func (moduleStruct) sendNoFormatNote(b *gotgbot.Bot, ctx *ext.Context, replyMsgId int64, noteData *db.Notes) error {
	user := ctx.EffectiveSender.User

	// check if user is admin or not
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	// Reverse notedata
	noteData.NoteContent = helpers.ReverseHTML2MD(noteData.NoteContent)

	// show the buttons back as text
	noteData.NoteContent += helpers.RevertButtons(noteData.Buttons)

	// raw note does not need webpreview
	_, err := helpers.NotesEnumFuncMap[noteData.MsgType](
		b,
		ctx,
		noteData,
		&gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil},
		replyMsgId,
		false,
		noteData.IsProtected,
		true,
		noteData.NoNotif,
	)
	// if strings.Contains(err.Error(), "replied message not found") {
	// 	return ext.EndGroups
	// }
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// LoadNotes registers all notes module handlers with the dispatcher,
// including note management commands and the notes watcher.
func LoadNotes(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(notesModule.moduleName, true)

	HelpModule.helpableKb[notesModule.moduleName] = [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         "Formatting",
				CallbackData: fmt.Sprintf("helpq.%s", "Formatting"),
			},
		},
	} // Adds Formatting kb button to Notes Menu
	dispatcher.AddHandler(handlers.NewCommand("save", notesModule.addNote))
	dispatcher.AddHandler(handlers.NewCommand("addnote", notesModule.addNote))
	dispatcher.AddHandler(handlers.NewCommand("clear", notesModule.rmNote))
	dispatcher.AddHandler(handlers.NewCommand("rmnote", notesModule.rmNote))
	dispatcher.AddHandler(handlers.NewCommand("notes", notesModule.notesList))
	misc.AddCmdToDisableable("notes")
	dispatcher.AddHandler(handlers.NewCommand("clearall", notesModule.rmAllNotes))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("rmAllNotes"), notesModule.notesButtonHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("notes.overwrite."), notesModule.noteOverWriteHandler))
	dispatcher.AddHandler(
		handlers.NewMessage(
			func(msg *gotgbot.Message) bool {
				return strings.HasPrefix(msg.Text, "#")
			},
			notesModule.notesWatcher,
		),
	)
	cmdDecorator.MultiCommand(dispatcher, []string{"privnote", "privatenotes"}, notesModule.privNote)
	dispatcher.AddHandler(handlers.NewCommand("get", notesModule.getNotes))
	misc.AddCmdToDisableable("get")
}
