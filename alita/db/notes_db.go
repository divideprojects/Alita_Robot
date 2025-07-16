package db

import (
	"context"

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

// getNotesSettings fetches note settings for a chat from the database.
// If no document exists, it creates one with default values (private notes disabled).
// Returns a pointer to the NoteSettings struct with either existing or default values.
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

// getAllChatNotes retrieves all notes for a specific chat without pagination.
// Uses GetAllNotesPaginated internally with no limit to fetch all notes.
// Returns a slice of all ChatNotes for the specified chat.
func getAllChatNotes(chatId int64) (notes []*ChatNotes) {
	result, _ := GetAllNotesPaginated(chatId, PaginationOptions{Limit: 0})
	return result.Data
}

// GetAllNotesPaginated returns paginated notes for a chat using cursor or offset-based pagination.
// Supports both cursor-based and offset-based pagination depending on the options provided.
// Returns a PaginatedResult containing the notes and pagination metadata.
func GetAllNotesPaginated(chatID int64, opts PaginationOptions) (PaginatedResult[*ChatNotes], error) {
	paginator := NewMongoPagination[*ChatNotes](notesColl)

	filter := bson.M{"chat_id": chatID}
	if opts.Cursor != nil {
		filter["_id"] = bson.M{"$gt": opts.Cursor}
	}

	ctx := context.Background()
	if opts.Offset > 0 {
		return paginator.GetPageByOffset(ctx, PaginationOptions{
			Offset:        opts.Offset,
			Limit:         opts.Limit,
			SortDirection: 1,
		}, filter)
	}

	return paginator.GetNextPage(ctx, PaginationOptions{
		Limit:         opts.Limit,
		SortDirection: 1,
	}, filter)
}

// GetNotes retrieves the note settings for a given chat ID.
// If no settings exist, it initializes them with default values (private notes disabled).
// This is the main function for accessing note settings.
func GetNotes(chatID int64) *NoteSettings {
	return getNotesSettings(chatID)
}

// GetNote retrieves a specific note by keyword for a chat.
// Returns nil if the note does not exist or if database operation fails.
// Keyword matching is case-sensitive.
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
// Used to display available notes to users based on their permissions.
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
// Uses direct database query to avoid race conditions and ensure accuracy.
// Returns false if database operation fails or note doesn't exist.
func DoesNoteExists(chatID int64, noteName string) bool {
	count, err := countDocs(notesColl, bson.M{"chat_id": chatID, "note_name": noteName})
	if err != nil {
		log.Errorf("[Database][DoesNoteExists]: %d - %v", chatID, err)
		return false
	}
	return count > 0
}

// AddNote adds a new note to the chat with the specified properties.
// If a note with the same name already exists, no action is taken due to upsert behavior.
// Returns true if a new note was added, false if it already existed or operation failed.
// Supports various note types including text, media, and buttons.
func AddNote(chatID int64, noteName, replyText, fileID string, buttons []Button, filtType int, pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif bool) bool {
	filter := bson.M{"chat_id": chatID, "note_name": noteName}
	update := bson.M{
		"$setOnInsert": bson.M{
			"chat_id":      chatID,
			"note_name":    noteName,
			"note_content": replyText,
			"msgtype":      filtType,
			"fileid":       fileID,
			"note_buttons": buttons,
			"private_only": pvtOnly,
			"group_only":   grpOnly,
			"webpreview":   webPrev,
			"admin_only":   adminOnly,
			"is_protected": isProtected,
			"no_notif":     noNotif,
		},
	}

	result := &ChatNotes{}
	err := findOneAndUpsert(notesColl, filter, update, result)
	if err != nil {
		log.Errorf("[Database][AddNotes]: %d - %v", chatID, err)
		return false
	}

	// Return true if this was a new insert (the document should have our values)
	return result.ChatId == chatID && result.NoteName == noteName
}

// RemoveNote deletes a note by name from the chat.
// If the note does not exist, no action is taken. Verifies note existence before deletion.
// Operation is idempotent - safe to call multiple times.
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
// This operation cannot be undone. All note data including buttons and settings are lost.
// Use with caution as this affects all notes in the chat.
func RemoveAllNotes(chatID int64) {
	err := deleteMany(notesColl, bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][RemoveAllNotes]: %d - %v", chatID, err)
	}
}

// TooglePrivateNote enables or disables private notes for a chat.
// When enabled, notes are sent privately to users instead of in the group.
// Helps reduce chat spam and provides privacy for note content.
func TooglePrivateNote(chatID int64, pref bool) {
	noterc := getNotesSettings(chatID)
	noterc.PrivateNotesEnabled = pref
	err := updateOne(notesSettingsColl, bson.M{"_id": chatID}, noterc)
	if err != nil {
		log.Errorf("[Database][TooglePrivateNote]: %d - %v", chatID, err)
	}
}

// LoadNotesStats returns the total number of notes and the number of chats using notes.
// Counts all notes across all chats and calculates unique chats that have at least one note.
// Used for bot statistics and monitoring purposes.
func LoadNotesStats() (notesNum, notesUsingChats int64) {
	var notesArray []*ChatNotes
	notesMap := make(map[int64][]ChatNotes)

	cursor := findAll(notesColl, bson.M{})
	ctx := context.Background()
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error("Failed to close notes cursor:", err)
		}
	}()
	if err := cursor.All(ctx, &notesArray); err != nil {
		log.Error("Failed to load notes stats:", err)
		return
	}

	for _, noteC := range notesArray {
		notesNum++ // count number of filters
		notesMap[noteC.ChatId] = append(notesMap[noteC.ChatId], *noteC)
	}

	notesUsingChats = int64(len(notesMap))

	return
}
