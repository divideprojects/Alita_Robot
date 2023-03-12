package helpers

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	tgmd2html "github.com/PaulSonOfLars/gotg_md2html"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"

	"github.com/Divkix/Alita_Robot/alita/db"
	"github.com/Divkix/Alita_Robot/alita/utils/extraction"
	"github.com/Divkix/Alita_Robot/alita/utils/parsemode"
)

func GetNoteAndFilterType(msg *gotgbot.Message, isFilter bool) (keyWord, fileid, text string, dataType int, buttons []db.Button, pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif bool, errorMsg string) {
	dataType = -1 // not defined datatype; invalid note
	if isFilter {
		errorMsg = "You need to give the filter some content!"
	} else {
		errorMsg = "You need to give the note some content!"
	}

	var (
		rawText string
		args    = strings.Fields(msg.Text)[1:]
	)
	_buttons := make([]tgmd2html.ButtonV2, 0) // make a slice for buttons
	replyMsg := msg.ReplyToMessage

	// set rawText from helper function
	setRawText(msg, args, &rawText)

	// extract the noteword
	if len(args) >= 2 && replyMsg == nil {
		keyWord, text = extraction.ExtractQuotes(rawText, isFilter, true)
		text, _buttons = tgmd2html.MD2HTMLButtonsV2(text)
		dataType = db.TEXT
	} else if replyMsg != nil && len(args) >= 1 {
		keyWord, _ = extraction.ExtractQuotes(strings.Join(args, " "), isFilter, true)

		if replyMsg.ReplyMarkup == nil {
			text, _buttons = tgmd2html.MD2HTMLButtonsV2(rawText)
		} else {
			text, _ = tgmd2html.MD2HTMLButtonsV2(rawText)
			_buttons = InlineKeyboardMarkupToTgmd2htmlButtonV2(replyMsg.ReplyMarkup)
		}

		if replyMsg.Text != "" {
			dataType = db.TEXT
		} else if replyMsg.Sticker != nil {
			fileid = replyMsg.Sticker.FileId
			dataType = db.STICKER
		} else if replyMsg.Document != nil {
			fileid = replyMsg.Document.FileId
			dataType = db.DOCUMENT
		} else if len(replyMsg.Photo) > 0 {
			fileid = replyMsg.Photo[len(replyMsg.Photo)-1].FileId // using -1 index to get best photo quality
			dataType = db.PHOTO
		} else if replyMsg.Audio != nil {
			fileid = replyMsg.Audio.FileId
			dataType = db.AUDIO
		} else if replyMsg.Voice != nil {
			fileid = replyMsg.Voice.FileId
			dataType = db.VOICE
		} else if replyMsg.Video != nil {
			fileid = replyMsg.Video.FileId
			dataType = db.VIDEO
		} else if replyMsg.VideoNote != nil {
			fileid = replyMsg.VideoNote.FileId
			dataType = db.VideoNote
		}
	}

	// pre-fix the data before sending it back
	preFixes(_buttons, keyWord, &text, &dataType, fileid, &buttons, &errorMsg)

	// return if datatype is invalid
	if dataType != -1 && !isFilter {
		// parse options such as pvtOnly, adminOnly, webPrev and replace them
		pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif, _ = notesParser(text)
	}

	return
}

func GetWelcomeType(msg *gotgbot.Message, greetingType string) (text string, dataType int, fileid string, buttons []db.Button, errorMsg string) {
	dataType = -1
	errorMsg = fmt.Sprintf("You need to give me some content to %s users!", greetingType)
	var (
		rawText string
		args    = strings.Fields(msg.Text)[1:]
	)
	_buttons := make([]tgmd2html.ButtonV2, 0)
	replyMsg := msg.ReplyToMessage

	// set rawText from helper function
	setRawText(msg, args, &rawText)

	if len(args) >= 1 && msg.ReplyToMessage == nil {
		fileid = ""
		text, _buttons = tgmd2html.MD2HTMLButtonsV2(strings.SplitN(rawText, " ", 2)[1])
		dataType = db.TEXT
	} else if msg.ReplyToMessage != nil {
		if replyMsg.ReplyMarkup == nil {
			text, _buttons = tgmd2html.MD2HTMLButtonsV2(rawText)
		} else {
			text, _ = tgmd2html.MD2HTMLButtonsV2(rawText)
			_buttons = InlineKeyboardMarkupToTgmd2htmlButtonV2(replyMsg.ReplyMarkup)
		}
		if len(args) == 0 && replyMsg.Text != "" {
			dataType = db.TEXT
		} else if replyMsg.Sticker != nil {
			fileid = replyMsg.Sticker.FileId
			dataType = db.STICKER
		} else if replyMsg.Document != nil {
			fileid = replyMsg.Document.FileId
			dataType = db.DOCUMENT
		} else if len(replyMsg.Photo) > 0 {
			fileid = replyMsg.Photo[len(replyMsg.Photo)-1].FileId
			dataType = db.PHOTO
		} else if replyMsg.Audio != nil {
			fileid = replyMsg.Audio.FileId
			dataType = db.AUDIO
		} else if replyMsg.Voice != nil {
			fileid = replyMsg.Voice.FileId
			dataType = db.VOICE
		} else if replyMsg.Video != nil {
			fileid = replyMsg.Video.FileId
			dataType = db.VIDEO
		} else if replyMsg.VideoNote != nil {
			fileid = replyMsg.VideoNote.FileId
			dataType = db.VideoNote
		}
	}

	// pre-fix the data before sending it back
	preFixes(_buttons, "Greeting", &text, &dataType, fileid, &buttons, &errorMsg)

	return
}

// SendFilter Simple function used to send a filter with help from EnumFuncMap, this just organises data for it
func SendFilter(b *gotgbot.Bot, ctx *ext.Context, filterData *db.ChatFilters, replyMsgId int64) (*gotgbot.Message, error) {
	chat := ctx.EffectiveChat

	var (
		keyb          = make([][]gotgbot.InlineKeyboardButton, 0)
		buttons       []db.Button
		sent          string
		tmpfilterData db.ChatFilters
	)
	tmpfilterData = *filterData
	buttons = tmpfilterData.Buttons

	// Random data goes there
	rand.Seed(time.Now().Unix())
	rstrings := strings.Split(tmpfilterData.FilterReply, "%%%")
	if len(rstrings) == 1 {
		sent = rstrings[0]
	} else {
		n := rand.Int() % len(rstrings)
		sent = rstrings[n]
	}

	tmpfilterData.FilterReply, buttons = FormattingReplacer(b, chat, ctx.EffectiveUser, sent, buttons)
	keyb = BuildKeyboard(buttons)
	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}

	// using false as last arg because we don't want to noformat the message
	msg, err := FiltersEnumFuncMap[tmpfilterData.MsgType](b, ctx, tmpfilterData, &keyboard, replyMsgId, false, filterData.NoNotif)

	return msg, err
}

func notesParser(sent string) (pvtOnly, grpOnly, adminOnly, webPrev, protectedContent, noNotif bool, sentBack string) {
	pvtOnly, err := regexp.MatchString(`({private})`, sent)
	if err != nil {
		log.Error(err)
		return
	}

	grpOnly, err = regexp.MatchString(`({noprivate})`, sent)
	if err != nil {
		log.Error(err)
		return
	}

	adminOnly, err = regexp.MatchString(`({admin})`, sent)
	if err != nil {
		log.Error(err)
		return
	}

	webPrev, err = regexp.MatchString(`({preview})`, sent)
	if err != nil {
		log.Error(err)
		return
	}

	protectedContent, err = regexp.MatchString(`({protect})`, sent)
	if err != nil {
		log.Error(err)
		return
	}

	noNotif, err = regexp.MatchString(`({nonotif})`, sent)
	if err != nil {
		log.Error(err)
		return
	}

	sent = strings.NewReplacer(
		"{private}", "",
		"{admin}", "",
		"{preview}", "",
		"{noprivate}", "",
		"{protect}", "",
		"{nonotif}", "",
	).Replace(sent)

	return pvtOnly, grpOnly, adminOnly, webPrev, protectedContent, noNotif, sent
}

func SendNote(b *gotgbot.Bot, chat *gotgbot.Chat, ctx *ext.Context, noteData *db.ChatNotes, replyMsgId int64) (*gotgbot.Message, error) {
	var (
		keyb    = make([][]gotgbot.InlineKeyboardButton, 0)
		buttons []db.Button
		sent    string
	)

	// copy just in case
	buttons = noteData.Buttons

	// Random data goes there
	rand.Seed(time.Now().Unix())
	rstrings := strings.Split(noteData.NoteContent, "%%%")
	if len(rstrings) == 1 {
		sent = rstrings[0]
	} else {
		n := rand.Int() % len(rstrings)
		sent = rstrings[n]
	}

	noteData.NoteContent, buttons = FormattingReplacer(b, chat, ctx.EffectiveUser, sent, buttons)
	// below is an additional step, need to remove it
	_, _, _, _, _, _, noteData.NoteContent = notesParser(noteData.NoteContent) // replaces the text
	keyb = BuildKeyboard(buttons)
	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}
	// using false as last arg to format the note
	msg, err := NotesEnumFuncMap[noteData.MsgType](b, ctx, noteData, &keyboard, replyMsgId, noteData.WebPreview, noteData.IsProtected, false, noteData.NoNotif)
	// if strings.Contains(err.Error(), "replied message not found") {
	// 	return nil, nil
	// }
	if err != nil {
		log.Error(err)
		return msg, err
	}

	return msg, nil
}

// NotesEnumFuncMap TODO: make a new function to merge all EnumFuncMap functions
// NotesEnumFuncMap
// A rather very complicated NotesEnumFuncMap Variable made by me to send filters in an appropriate way
var NotesEnumFuncMap = map[int]func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, webPreview, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error){
	db.TEXT: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, webPreview, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendMessage(ctx.Update.Message.Chat.Id,
			noteData.NoteContent,
			&gotgbot.SendMessageOpts{
				ParseMode:                formatMode,
				DisableWebPagePreview:    !webPreview,
				ReplyMarkup:              keyb,
				ReplyToMessageId:         replyMsgId,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.STICKER: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		return b.SendSticker(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendStickerOpts{
				ReplyToMessageId:         replyMsgId,
				ReplyMarkup:              keyb,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.DOCUMENT: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendDocument(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendDocumentOpts{
				ReplyToMessageId:         replyMsgId,
				ParseMode:                formatMode,
				ReplyMarkup:              keyb,
				Caption:                  noteData.NoteContent,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.PHOTO: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendPhoto(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendPhotoOpts{
				ReplyToMessageId:         replyMsgId,
				ParseMode:                formatMode,
				ReplyMarkup:              keyb,
				Caption:                  noteData.NoteContent,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.AUDIO: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendAudio(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendAudioOpts{
				ReplyToMessageId:         replyMsgId,
				ParseMode:                formatMode,
				ReplyMarkup:              keyb,
				Caption:                  noteData.NoteContent,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.VOICE: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendVoice(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendVoiceOpts{
				ReplyToMessageId:         replyMsgId,
				ParseMode:                formatMode,
				ReplyMarkup:              keyb,
				Caption:                  noteData.NoteContent,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.VIDEO: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendVideo(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendVideoOpts{
				ReplyToMessageId:         replyMsgId,
				ParseMode:                formatMode,
				ReplyMarkup:              keyb,
				Caption:                  noteData.NoteContent,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.VideoNote: func(b *gotgbot.Bot, ctx *ext.Context, noteData *db.ChatNotes, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, _, isProtected bool, noFormat, noNotif bool) (*gotgbot.Message, error) {
		return b.SendVideoNote(ctx.Update.Message.Chat.Id,
			noteData.FileID,
			&gotgbot.SendVideoNoteOpts{
				ReplyToMessageId:         replyMsgId,
				ReplyMarkup:              keyb,
				AllowSendingWithoutReply: true,
				ProtectContent:           isProtected,
				DisableNotification:      noNotif,
				MessageThreadId:          ctx.Update.Message.MessageThreadId,
			},
		)
	},
}

// GreetingsEnumFuncMap FIXME: when using /welcome command in private with connection, the string of welcome is sent to connected chat instead of pm
// GreetingsEnumFuncMap
// A rather very complicated GreetingsEnumFuncMap Variable made by me to send filters in an appropriate way
var GreetingsEnumFuncMap = map[int]func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error){
	db.TEXT: func(b *gotgbot.Bot, ctx *ext.Context, msg, _ string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendMessage(
			ctx.EffectiveChat.Id,
			msg,
			&gotgbot.SendMessageOpts{
				ParseMode:             parsemode.HTML,
				DisableWebPagePreview: true,
				ReplyMarkup:           keyb,
			},
		)
	},
	db.STICKER: func(b *gotgbot.Bot, ctx *ext.Context, _, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendSticker(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendStickerOpts{
				ReplyMarkup:     keyb,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.DOCUMENT: func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendDocument(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendDocumentOpts{
				ParseMode:       parsemode.HTML,
				ReplyMarkup:     keyb,
				Caption:         msg,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.PHOTO: func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendPhoto(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendPhotoOpts{
				ParseMode:       parsemode.HTML,
				ReplyMarkup:     keyb,
				Caption:         msg,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.AUDIO: func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendAudio(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendAudioOpts{
				ParseMode:       parsemode.HTML,
				ReplyMarkup:     keyb,
				Caption:         msg,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.VOICE: func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendVoice(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendVoiceOpts{
				ParseMode:       parsemode.HTML,
				ReplyMarkup:     keyb,
				Caption:         msg,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.VIDEO: func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendVideo(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendVideoOpts{
				ParseMode:       parsemode.HTML,
				ReplyMarkup:     keyb,
				Caption:         msg,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.VideoNote: func(b *gotgbot.Bot, ctx *ext.Context, msg, fileID string, keyb *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
		return b.SendVideoNote(
			ctx.EffectiveChat.Id,
			fileID,
			&gotgbot.SendVideoNoteOpts{
				ReplyMarkup:     keyb,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
}

// FiltersEnumFuncMap
// A rather very complicated FiltersEnumFuncMap Variable made by me to send filters in an appropriate way
var FiltersEnumFuncMap = map[int]func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error){
	db.TEXT: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendMessage(
			ctx.Update.Message.Chat.Id,
			filterData.FilterReply,
			&gotgbot.SendMessageOpts{
				ParseMode:             formatMode,
				DisableWebPagePreview: true,
				ReplyToMessageId:      replyMsgId,
				ReplyMarkup:           keyb,
				DisableNotification:   noNotif,
				MessageThreadId:       ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.STICKER: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		return b.SendSticker(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendStickerOpts{
				ReplyToMessageId:    replyMsgId,
				ReplyMarkup:         keyb,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.DOCUMENT: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendDocument(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendDocumentOpts{
				ReplyToMessageId:    replyMsgId,
				ParseMode:           formatMode,
				ReplyMarkup:         keyb,
				Caption:             filterData.FilterReply,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.PHOTO: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendPhoto(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendPhotoOpts{
				ReplyToMessageId:    replyMsgId,
				ParseMode:           formatMode,
				ReplyMarkup:         keyb,
				Caption:             filterData.FilterReply,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.AUDIO: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendAudio(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendAudioOpts{
				ReplyToMessageId:    replyMsgId,
				ParseMode:           formatMode,
				ReplyMarkup:         keyb,
				Caption:             filterData.FilterReply,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.VOICE: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendVoice(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendVoiceOpts{
				ReplyToMessageId:    replyMsgId,
				ParseMode:           formatMode,
				ReplyMarkup:         keyb,
				Caption:             filterData.FilterReply,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.VIDEO: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		formatMode := parsemode.HTML
		if noFormat {
			formatMode = parsemode.None
		}
		return b.SendVideo(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendVideoOpts{
				ReplyToMessageId:    replyMsgId,
				ParseMode:           formatMode,
				ReplyMarkup:         keyb,
				Caption:             filterData.FilterReply,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
	db.VideoNote: func(b *gotgbot.Bot, ctx *ext.Context, filterData db.ChatFilters, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64, noFormat, noNotif bool) (*gotgbot.Message, error) {
		return b.SendVideoNote(
			ctx.Update.Message.Chat.Id,
			filterData.FileID,
			&gotgbot.SendVideoNoteOpts{
				ReplyToMessageId:    replyMsgId,
				ReplyMarkup:         keyb,
				DisableNotification: noNotif,
				MessageThreadId:     ctx.Update.Message.MessageThreadId,
			},
		)
	},
}

// this function should fix the empty names of buttons in the buttons slice
//
// it first checks if text length is appropriate or not
//
// if it is appropriate, then perform other minor fixes
func preFixes(buttons []tgmd2html.ButtonV2, defaultNameButton string, text *string, dataType *int, fileid string, dbButtons *[]db.Button, errorMsg *string) {
	if *dataType == db.TEXT && len(*text) > 4096 {
		*dataType = -1
		*errorMsg = fmt.Sprintf("Your message text is %d characters long. The maximum length for text is 4096; please trim it to a smaller size. Note that markdown characters may take more space than expected.", len(*text))
	} else if *dataType != db.TEXT && len(*text) > 1024 {
		*dataType = -1
		*errorMsg = fmt.Sprintf("Your message caption is %d characters long. The maximum caption length is 1024; please trim it to a smaller size. Note that markdown characters may take more space than expected.", len(*text))
	} else {
		for i, button := range buttons {
			if button.Name == "" {
				buttons[i].Name = defaultNameButton
			}
		}

		// temporary variable function until we don't support notes in inline keyboard
		// will remove non url buttons from keyboard
		buttonUrlFixer := func(_buttons *[]tgmd2html.ButtonV2) {
			// regex taken from https://regexr.com/39nr7
			buttonUrlPattern, _ := regexp.Compile(`[(htps)?:/w.a-zA-Z\d@%_+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z\d@:%_+.~#?&/=]*)`)
			buttons = *_buttons
			for i, btn := range *_buttons {
				if !buttonUrlPattern.MatchString(btn.Text) {
					buttons = append(buttons[:i], buttons[i+1:]...)
				}
			}
			*_buttons = buttons
		}

		buttonUrlFixer(&buttons)
		*dbButtons = ConvertButtonV2ToDbButton(buttons)

		// trim the characters \n, \t, \r and space from the text
		// also, set the dataType to -1 to make note invalid
		*text = strings.Trim(*text, "\n\t\r ")
		if *text == "" && fileid == "" {
			*dataType = -1
		}
	}
}

// function used to get rawtext from gotgbot.Message
func setRawText(msg *gotgbot.Message, args []string, rawText *string) {
	replyMsg := msg.ReplyToMessage
	if replyMsg == nil {
		if msg.Text == "" && msg.Caption != "" {
			*rawText = strings.SplitN(msg.OriginalCaptionMDV2(), " ", 2)[1] // using [1] to remove the command
		} else if msg.Text != "" && msg.Caption == "" {
			*rawText = strings.SplitN(msg.OriginalMDV2(), " ", 2)[1] // using [1] to remove the command
		}
	} else {
		if replyMsg.Text == "" && replyMsg.Caption != "" {
			*rawText = replyMsg.OriginalCaptionMDV2()
		} else if replyMsg.Caption == "" && len(args) >= 2 {
			*rawText = strings.SplitN(msg.OriginalMDV2(), " ", 3)[2] // using [1] to remove the command
		} else {
			*rawText = replyMsg.OriginalMDV2()
		}
	}
}
