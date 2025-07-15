package db

import (
	"context"
	"fmt"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// withTransaction executes operations within a MongoDB transaction
func withTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	if mongoClient == nil {
		return fmt.Errorf("database client is not initialized")
	}
	
	session, err := mongoClient.StartSession()
	if err != nil {
		log.Errorf("[Database][withTransaction] Failed to start session: %v", err)
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	if err != nil {
		log.Errorf("[Database][withTransaction] Transaction failed: %v", err)
	}
	return err
}

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
	result, _ := GetAllNotesPaginated(chatId, PaginationOptions{Limit: 0})
	return result.Data
}

// GetAllNotesPaginated returns paginated notes for a chat.
func GetAllNotesPaginated(chatID int64, opts PaginationOptions) (PaginatedResult[*ChatNotes], error) {
	paginator := NewMongoPagination[*ChatNotes](notesColl)

	filter := bson.M{"chat_id": chatID}
	if opts.Cursor != nil {
		filter["_id"] = bson.M{"$gt": opts.Cursor}
	}

	if opts.Offset > 0 {
		return paginator.GetPageByOffset(context.Background(), filter, PaginationOptions{
			Offset:        opts.Offset,
			Limit:         opts.Limit,
			SortDirection: 1,
		})
	}

	return paginator.GetNextPage(context.Background(), filter, PaginationOptions{
		Limit:         opts.Limit,
		SortDirection: 1,
	})
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
// Returns true if a new note was added, false if it already existed.
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
	err := withTransaction(context.Background(), func(sessCtx mongo.SessionContext) error {
		// Perform the upsert operation within the transaction
		err := findOneAndUpsert(notesColl, filter, update, result)
		if err != nil {
			return err
		}

		// Update cache only if this was a new insert
		if result.ChatId == chatID && result.NoteName == noteName {
			if err := cache.Marshal.Set(cache.Context, chatID, result, store.WithExpiration(10*time.Minute)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Errorf("[Database][AddNotes]: %d - %v", chatID, err)
		return false
	}

	// Return true if this was a new insert
	return result.ChatId == chatID && result.NoteName == noteName
}

// RemoveNote deletes a note by name from the chat.
func RemoveNote(chatID int64, noteName string) {
	err := withTransaction(context.Background(), func(sessCtx mongo.SessionContext) error {
		// Perform the delete operation within the transaction
		err := deleteOne(notesColl, bson.M{"chat_id": chatID, "note_name": noteName})
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}

		// Invalidate cache after successful deletion
		if err := cache.Marshal.Delete(cache.Context, chatID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Errorf("[Database][RemoveNote]: %d - %v", chatID, err)
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
	notesMap := make(map[int64]struct{})
	paginator := NewMongoPagination[*ChatNotes](notesColl)

	var cursor interface{}
	for {
		result, err := paginator.GetNextPage(context.Background(), bson.M{}, PaginationOptions{
			Cursor:        cursor,
			Limit:         100, // Process 100 docs at a time
			SortDirection: 1,
		})
		if err != nil || len(result.Data) == 0 {
			break
		}

		for _, note := range result.Data {
			notesNum++
			notesMap[note.ChatId] = struct{}{}
		}

		cursor = result.NextCursor
		if cursor == nil {
			break
		}
	}

	notesUsingChats = int64(len(notesMap))
	return
}
