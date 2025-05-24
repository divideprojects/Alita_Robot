package messaging

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

// SendOptions contains common options for sending messages
type SendOptions struct {
	ReplyToMessageID int64
	Keyboard         *gotgbot.InlineKeyboardMarkup
	ParseMode        string
	WebPreview       bool
	IsProtected      bool
	NoFormat         bool
	NoNotif          bool
	MessageThreadID  int64
	Caption          string
}

// MessageContent represents the content to be sent
type MessageContent interface {
	GetText() string
	GetFileID() string
	GetDataType() int
}

// NoteContent implements MessageContent for notes
type NoteContent struct {
	*db.ChatNotes
}

func (nc *NoteContent) GetText() string   { return nc.NoteContent }
func (nc *NoteContent) GetFileID() string { return nc.FileID }
func (nc *NoteContent) GetDataType() int  { return nc.MsgType }

// FilterContent implements MessageContent for filters
type FilterContent struct {
	*db.ChatFilters
}

func (fc *FilterContent) GetText() string   { return fc.FilterReply }
func (fc *FilterContent) GetFileID() string { return fc.FileID }
func (fc *FilterContent) GetDataType() int  { return fc.MsgType }

// GreetingContent implements MessageContent for greetings
type GreetingContent struct {
	Text     string
	FileID   string
	DataType int
}

func (gc *GreetingContent) GetText() string   { return gc.Text }
func (gc *GreetingContent) GetFileID() string { return gc.FileID }
func (gc *GreetingContent) GetDataType() int  { return gc.DataType }

// SendMessage sends a message using the generic sender framework
// This replaces the three massive EnumFuncMap structures with a single, type-safe implementation
func SendMessage(bot *gotgbot.Bot, ctx *ext.Context, content MessageContent, opts *SendOptions) (*gotgbot.Message, error) {
	if opts == nil {
		opts = &SendOptions{}
	}

	// Set default parse mode if not specified
	if opts.ParseMode == "" {
		if opts.NoFormat {
			opts.ParseMode = helpers.None
		} else {
			opts.ParseMode = helpers.HTML
		}
	}

	chatID := ctx.EffectiveChat.Id
	if ctx.Update.Message != nil {
		chatID = ctx.Update.Message.Chat.Id
	}

	// Set message thread ID if available
	if opts.MessageThreadID == 0 && ctx.EffectiveMessage != nil {
		opts.MessageThreadID = ctx.EffectiveMessage.MessageThreadId
	}

	// Route to appropriate sender based on data type
	switch content.GetDataType() {
	case db.TEXT:
		return sendTextMessage(bot, chatID, content.GetText(), opts)
	case db.STICKER:
		return sendStickerMessage(bot, chatID, content.GetFileID(), opts)
	case db.DOCUMENT:
		return sendDocumentMessage(bot, chatID, content.GetFileID(), content.GetText(), opts)
	case db.PHOTO:
		return sendPhotoMessage(bot, chatID, content.GetFileID(), content.GetText(), opts)
	case db.AUDIO:
		return sendAudioMessage(bot, chatID, content.GetFileID(), content.GetText(), opts)
	case db.VOICE:
		return sendVoiceMessage(bot, chatID, content.GetFileID(), content.GetText(), opts)
	case db.VIDEO:
		return sendVideoMessage(bot, chatID, content.GetFileID(), content.GetText(), opts)
	case db.VideoNote:
		return sendVideoNoteMessage(bot, chatID, content.GetFileID(), opts)
	default:
		return sendTextMessage(bot, chatID, content.GetText(), opts)
	}
}

// sendTextMessage handles text message sending
func sendTextMessage(bot *gotgbot.Bot, chatID int64, text string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendMessage(chatID, text, &gotgbot.SendMessageOpts{
		ParseMode: opts.ParseMode,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: !opts.WebPreview,
		},
		ReplyMarkup: opts.Keyboard,
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendStickerMessage handles sticker message sending
func sendStickerMessage(bot *gotgbot.Bot, chatID int64, fileID string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendSticker(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendStickerOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup:         opts.Keyboard,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendDocumentMessage handles document message sending
func sendDocumentMessage(bot *gotgbot.Bot, chatID int64, fileID, caption string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendDocument(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendDocumentOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ParseMode:           opts.ParseMode,
		ReplyMarkup:         opts.Keyboard,
		Caption:             caption,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendPhotoMessage handles photo message sending
func sendPhotoMessage(bot *gotgbot.Bot, chatID int64, fileID, caption string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendPhoto(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendPhotoOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ParseMode:           opts.ParseMode,
		ReplyMarkup:         opts.Keyboard,
		Caption:             caption,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendAudioMessage handles audio message sending
func sendAudioMessage(bot *gotgbot.Bot, chatID int64, fileID, caption string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendAudio(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendAudioOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ParseMode:           opts.ParseMode,
		ReplyMarkup:         opts.Keyboard,
		Caption:             caption,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendVoiceMessage handles voice message sending
func sendVoiceMessage(bot *gotgbot.Bot, chatID int64, fileID, caption string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendVoice(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendVoiceOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ParseMode:           opts.ParseMode,
		ReplyMarkup:         opts.Keyboard,
		Caption:             caption,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendVideoMessage handles video message sending
func sendVideoMessage(bot *gotgbot.Bot, chatID int64, fileID, caption string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendVideo(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendVideoOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ParseMode:           opts.ParseMode,
		ReplyMarkup:         opts.Keyboard,
		Caption:             caption,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// sendVideoNoteMessage handles video note message sending
func sendVideoNoteMessage(bot *gotgbot.Bot, chatID int64, fileID string, opts *SendOptions) (*gotgbot.Message, error) {
	return bot.SendVideoNote(chatID, gotgbot.InputFileByID(fileID), &gotgbot.SendVideoNoteOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                opts.ReplyToMessageID,
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup:         opts.Keyboard,
		ProtectContent:      opts.IsProtected,
		DisableNotification: opts.NoNotif,
		MessageThreadId:     opts.MessageThreadID,
	})
}

// Legacy wrapper functions for backward compatibility

// SendNote sends a note using the new generic framework
func SendNote(bot *gotgbot.Bot, chat *gotgbot.Chat, ctx *ext.Context, noteData *db.ChatNotes, replyMsgId int64) (*gotgbot.Message, error) {
	content := &NoteContent{noteData}
	opts := &SendOptions{
		ReplyToMessageID: replyMsgId,
		WebPreview:       noteData.WebPreview,
		IsProtected:      noteData.IsProtected,
		NoNotif:          noteData.NoNotif,
	}

	// Build keyboard
	keyb := helpers.BuildKeyboard(noteData.Buttons)
	if len(keyb) > 0 {
		opts.Keyboard = &gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}
	}

	return SendMessage(bot, ctx, content, opts)
}

// SendFilter sends a filter using the new generic framework
func SendFilter(bot *gotgbot.Bot, ctx *ext.Context, filterData *db.ChatFilters, replyMsgId int64) (*gotgbot.Message, error) {
	content := &FilterContent{filterData}
	opts := &SendOptions{
		ReplyToMessageID: replyMsgId,
		NoNotif:          filterData.NoNotif,
	}

	// Build keyboard
	keyb := helpers.BuildKeyboard(filterData.Buttons)
	if len(keyb) > 0 {
		opts.Keyboard = &gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}
	}

	return SendMessage(bot, ctx, content, opts)
}

// SendGreeting sends a greeting using the new generic framework
func SendGreeting(bot *gotgbot.Bot, ctx *ext.Context, text, fileID string, dataType int) (*gotgbot.Message, error) {
	content := &GreetingContent{
		Text:     text,
		FileID:   fileID,
		DataType: dataType,
	}
	opts := &SendOptions{}

	return SendMessage(bot, ctx, content, opts)
}
