package db

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// NoteSettings holds note-related configuration for a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - PrivateNotesEnabled: Whether private notes are enabled for the chat.
type NoteSettings struct {
	ChatId              int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	PrivateNotesEnabled bool  `bson:"private_notes" json:"private_notes" default:"false"`
}

// ChatNotes represents a single note stored in a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - NoteName: Name/keyword for the note.
//   - NoteContent: The note's content.
//   - MsgType: Type of message (e.g., text, media).
//   - FileID: Optional file ID for media attachments.
//   - PrivateOnly: Whether the note is only available in private chats.
//   - GroupOnly: Whether the note is only available in group chats.
//   - AdminOnly: Whether the note is restricted to admins.
//   - WebPreview: Whether web previews are enabled for the note.
//   - IsProtected: Whether the note is protected from deletion.
//   - NoNotif: Whether sending the note should suppress notifications.
//   - Buttons: List of buttons to attach to the note.
type ChatNotes struct {
	ChatId      int64    `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	NoteName    string   `bson:"note_name,omitempty" json:"note_name,omitempty"`
	NoteContent string   `bson:"note_content,omitempty" json:"note_content,omitempty"`
	MsgType     int      `bson:"msgtype" json:"msgtype" default:"1"`
	FileID      string   `bson:"fileid,omitempty" json:"fileid,omitempty"`
	PrivateOnly bool     `bson:"private_only,omitempty" json:"private_only,omitempty"`
	GroupOnly   bool     `bson:"group_only,omitempty" json:"group_only,omitempty"`
	AdminOnly   bool     `bson:"admin_only,omitempty" json:"admin_only,omitempty"`
	WebPreview  bool     `bson:"webpreview,omitempty" json:"webpreview,omitempty"`
	IsProtected bool     `bson:"is_protected,omitempty" json:"is_protected,omitempty"`
	NoNotif     bool     `bson:"no_notif,omitempty" json:"no_notif,omitempty"`
	Buttons     []Button `bson:"note_buttons,omitempty" json:"note_buttons,omitempty"`
}

func getNotesSettings(chatID int64) (noteSrc *NoteSettings) {
	defaultNotesSrc := &NoteSettings{ChatId: chatID, PrivateNotesEnabled: false}
	err := findOne(notesSettingsColl, bson.M{"_id": chatID}).Decode(&noteSrc)
	if err == mongo.ErrNoDocuments {
		noteSrc = defaultNotesSrc
		err := updateOne(notesSettingsColl, bson.M{"_id": chatID}, noteSrc)
		if err != nil {
			log.Errorf("[Database][getNotesSettings]: %d - %v", chatID, err)
		}
	} else if err != nil {
		log.Errorf("[Database] getNotesSettings: %v - %d", err, chatID)
	}
	return
}

func getAllChatNotes(chatId int64) (notes []*ChatNotes) {
	cursor := findAll(notesColl, bson.M{"chat_id": chatId})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &notes)
	return
}

// GetNotes retrieves the note settings for a given chat ID.
// If no settings exist, it initializes them with default values.
func GetNotes(chatID int64) *NoteSettings {
	return getNotesSettings(chatID)
}

// GetNote retrieves a specific note by keyword for a chat.
// Returns nil if the note does not exist.
func GetNote(chatID int64, keyword string) (noteSrc *ChatNotes) {
	err := findOne(notesColl, bson.M{"chat_id": chatID, "note_name": keyword}).Decode(&noteSrc)
	if err == mongo.ErrNoDocuments {
		noteSrc = nil
	} else if err != nil {
		log.Errorf("[Database] getNotesSettings: %v - %d", err, chatID)
	}
	return
}

// GetNotesList returns a list of note names for a chat.
// If admin is true, returns all notes; otherwise, excludes admin-only notes.
func GetNotesList(chatID int64, admin bool) (allNotes []string) {
	noteSrc := getAllChatNotes(chatID)
	for _, note := range noteSrc {
		if admin {
			allNotes = append(allNotes, note.NoteName)
		} else {
			if note.AdminOnly {
				continue
			}
			allNotes = append(allNotes, note.NoteName)
		}
	}

	return
}

// DoesNoteExists returns true if a note with the given name exists in the chat.
func DoesNoteExists(chatID int64, noteName string) bool {
	return string_handling.FindInStringSlice(GetNotesList(chatID, true), noteName)
}

// AddNote adds a new note to the chat with the specified properties.
// If a note with the same name already exists, no action is taken.
func AddNote(chatID int64, noteName, replyText, fileID string, buttons []Button, filtType int, pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif bool) {
	if string_handling.FindInStringSlice(GetNotesList(chatID, true), noteName) {
		return
	}

	noterc := ChatNotes{
		ChatId:      chatID,
		NoteName:    noteName,
		NoteContent: replyText,
		MsgType:     filtType,
		FileID:      fileID,
		Buttons:     buttons,
		PrivateOnly: pvtOnly,
		GroupOnly:   grpOnly,
		WebPreview:  webPrev,
		AdminOnly:   adminOnly,
		IsProtected: isProtected,
		NoNotif:     noNotif,
	}

	err := updateOne(notesColl, bson.M{"chat_id": chatID, "note_name": noteName}, noterc)
	if err != nil {
		log.Errorf("[Database][AddNotes]: %d - %v", chatID, err)
		return
	}
}

// RemoveNote deletes a note by name from the chat.
// If the note does not exist, no action is taken.
func RemoveNote(chatID int64, noteName string) {
	if !string_handling.FindInStringSlice(GetNotesList(chatID, true), noteName) {
		return
	}

	err := deleteOne(notesColl, bson.M{"chat_id": chatID, "note_name": noteName})
	if err != nil {
		log.Errorf("[Database][RemoveNote]: %d - %v", chatID, err)
		return
	}
}

// RemoveAllNotes deletes all notes from the specified chat.
func RemoveAllNotes(chatID int64) {
	err := deleteMany(notesColl, bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][RemoveAllNotes]: %d - %v", chatID, err)
	}
}

// TooglePrivateNote enables or disables private notes for a chat.
func TooglePrivateNote(chatID int64, pref bool) {
	noterc := getNotesSettings(chatID)
	noterc.PrivateNotesEnabled = pref
	err := updateOne(notesSettingsColl, bson.M{"_id": chatID}, noterc)
	if err != nil {
		log.Errorf("[Database][TooglePrivateNote]: %d - %v", chatID, err)
	}
}

// LoadNotesStats returns the total number of notes and the number of chats using notes.
func LoadNotesStats() (notesNum, notesUsingChats int64) {
	var notesArray []*ChatNotes
	notesMap := make(map[int64][]ChatNotes)

	cursor := findAll(notesColl, bson.M{})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &notesArray)

	for _, noteC := range notesArray {
		notesNum++ // count number of filters
		notesMap[noteC.ChatId] = append(notesMap[noteC.ChatId], *noteC)
	}

	notesUsingChats = int64(len(notesMap))

	return
}
