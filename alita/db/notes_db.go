package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

func getNotesSettings(chatID int64) (noteSrc *NotesSettings) {
	noteSrc = &NotesSettings{}
	err := GetRecord(noteSrc, NotesSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		noteSrc = &NotesSettings{ChatId: chatID, Private: false}
		err := CreateRecord(noteSrc)
		if err != nil {
			log.Errorf("[Database][getNotesSettings]: %d - %v", chatID, err)
		}
	} else if err != nil {
		// Return default on error
		noteSrc = &NotesSettings{ChatId: chatID, Private: false}
		log.Errorf("[Database] getNotesSettings: %v - %d", err, chatID)
	}
	return
}

func getAllChatNotes(chatId int64) (notes []*Notes) {
	err := GetRecords(&notes, Notes{ChatId: chatId})
	if err != nil {
		log.Errorf("[Database] getAllChatNotes: %v - %d", err, chatId)
		return []*Notes{}
	}
	return
}

func GetNotes(chatID int64) *NotesSettings {
	return getNotesSettings(chatID)
}

func GetNote(chatID int64, keyword string) (noteSrc *Notes) {
	noteSrc = &Notes{}
	err := GetRecord(noteSrc, Notes{ChatId: chatID, NoteName: keyword})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		log.Errorf("[Database] GetNote: %v - %d", err, chatID)
		return nil
	}

	return
}
func GetNotesList(chatID int64, admin bool) (allNotes []string) {
	noteSrc := getAllChatNotes(chatID)
	for _, note := range noteSrc {
		if admin {
			allNotes = append(allNotes, note.NoteName)
		} else {
			// Note: The new Notes model doesn't have AdminOnly field
			// All notes are accessible to all users
			allNotes = append(allNotes, note.NoteName)
		}
	}

	return
}

func DoesNoteExists(chatID int64, noteName string) bool {
	return string_handling.FindInStringSlice(GetNotesList(chatID, true), noteName)
}

func AddNote(chatID int64, noteName, replyText, fileID string, buttons ButtonArray, filtType int, pvtOnly, grpOnly, adminOnly, webPrev, isProtected, noNotif bool) {
	if string_handling.FindInStringSlice(GetNotesList(chatID, true), noteName) {
		return
	}

	noterc := Notes{
		ChatId:      chatID,
		NoteName:    noteName,
		NoteContent: replyText,
		MsgType:     filtType,
		FileID:      fileID,
		Buttons:     buttons,
		AdminOnly:   adminOnly,
		PrivateOnly: pvtOnly,
		GroupOnly:   grpOnly,
		WebPreview:  webPrev,
		IsProtected: isProtected,
		NoNotif:     noNotif,
	}

	err := CreateRecord(&noterc)
	if err != nil {
		log.Errorf("[Database][AddNotes]: %d - %v", chatID, err)
		return
	}
}

func RemoveNote(chatID int64, noteName string) {
	if !string_handling.FindInStringSlice(GetNotesList(chatID, true), noteName) {
		return
	}

	err := DB.Where("chat_id = ? AND note_name = ?", chatID, noteName).Delete(&Notes{}).Error
	if err != nil {
		log.Errorf("[Database][RemoveNote]: %d - %v", chatID, err)
		return
	}
}

func RemoveAllNotes(chatID int64) {
	err := DB.Where("chat_id = ?", chatID).Delete(&Notes{}).Error
	if err != nil {
		log.Errorf("[Database][RemoveAllNotes]: %d - %v", chatID, err)
	}
}

func TooglePrivateNote(chatID int64, pref bool) {
	err := UpdateRecord(&NotesSettings{}, NotesSettings{ChatId: chatID}, NotesSettings{Private: pref})
	if err != nil {
		log.Errorf("[Database][TooglePrivateNote]: %d - %v", chatID, err)
	}
}

func LoadNotesStats() (notesNum, notesUsingChats int64) {
	// Count total notes
	err := DB.Model(&Notes{}).Count(&notesNum).Error
	if err != nil {
		log.Errorf("[Database] LoadNotesStats (notes): %v", err)
		return 0, 0
	}

	// Count distinct chats with notes
	err = DB.Model(&Notes{}).Distinct("chat_id").Count(&notesUsingChats).Error
	if err != nil {
		log.Errorf("[Database] LoadNotesStats (chats): %v", err)
		return notesNum, 0
	}

	return
}
