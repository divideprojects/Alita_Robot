package db

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// Batch size for bulk operations
	BulkBatchSize = 100
)

// BulkAddFilters adds multiple filters in a single operation
func BulkAddFilters(chatID int64, filters map[string]string) error {
	if len(filters) == 0 {
		return nil
	}

	filterRecords := make([]ChatFilters, 0, len(filters))
	for keyword, reply := range filters {
		filterRecords = append(filterRecords, ChatFilters{
			ChatId:      chatID,
			KeyWord:     keyword,
			FilterReply: reply,
			MsgType:     TEXT, // default to text
		})
	}

	// Use CreateInBatches for efficient bulk insert
	err := DB.CreateInBatches(filterRecords, BulkBatchSize).Error
	if err != nil {
		log.Errorf("[Database][BulkAddFilters]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk add
	deleteCache(filterListCacheKey(chatID))
	return nil
}

// BulkRemoveFilters removes multiple filters in a single operation
func BulkRemoveFilters(chatID int64, keywords []string) error {
	if len(keywords) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND keyword IN ?", chatID, keywords).Delete(&ChatFilters{}).Error
	if err != nil {
		log.Errorf("[Database][BulkRemoveFilters]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk remove
	deleteCache(filterListCacheKey(chatID))
	return nil
}

// BulkAddNotes adds multiple notes in a single operation
func BulkAddNotes(chatID int64, notes map[string]string) error {
	if len(notes) == 0 {
		return nil
	}

	noteRecords := make([]Notes, 0, len(notes))
	for name, content := range notes {
		noteRecords = append(noteRecords, Notes{
			ChatId:      chatID,
			NoteName:    name,
			NoteContent: content,
			MsgType:     TEXT, // default to text
			WebPreview:  true, // default web preview enabled
		})
	}

	// Use CreateInBatches for efficient bulk insert
	err := DB.CreateInBatches(noteRecords, BulkBatchSize).Error
	if err != nil {
		log.Errorf("[Database][BulkAddNotes]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk add
	deleteCache(notesListCacheKey(chatID, true))
	deleteCache(notesListCacheKey(chatID, false))
	return nil
}

// BulkRemoveNotes removes multiple notes in a single operation
func BulkRemoveNotes(chatID int64, noteNames []string) error {
	if len(noteNames) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND note_name IN ?", chatID, noteNames).Delete(&Notes{}).Error
	if err != nil {
		log.Errorf("[Database][BulkRemoveNotes]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk remove
	deleteCache(notesListCacheKey(chatID, true))
	deleteCache(notesListCacheKey(chatID, false))
	return nil
}

// BulkAddBlacklist adds multiple blacklist words in a single operation
func BulkAddBlacklist(chatID int64, words []string, action string) error {
	if len(words) == 0 {
		return nil
	}

	blacklistRecords := make([]BlacklistSettings, 0, len(words))
	for _, word := range words {
		blacklistRecords = append(blacklistRecords, BlacklistSettings{
			ChatId: chatID,
			Word:   word,
			Action: action,
		})
	}

	// Use CreateInBatches for efficient bulk insert
	err := DB.CreateInBatches(blacklistRecords, BulkBatchSize).Error
	if err != nil {
		log.Errorf("[Database][BulkAddBlacklist]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk add
	deleteCache(blacklistCacheKey(chatID))
	return nil
}

// BulkRemoveBlacklist removes multiple blacklist words in a single operation
func BulkRemoveBlacklist(chatID int64, words []string) error {
	if len(words) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND word IN ?", chatID, words).Delete(&BlacklistSettings{}).Error
	if err != nil {
		log.Errorf("[Database][BulkRemoveBlacklist]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk remove
	deleteCache(blacklistCacheKey(chatID))
	return nil
}

// BulkWarnUsers warns multiple users in a single transaction
func BulkWarnUsers(chatID int64, userIDs []int64, reason string) error {
	if len(userIDs) == 0 {
		return nil
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		// Get warn settings
		warnSettings := &WarnSettings{}
		err := tx.Where("chat_id = ?", chatID).First(warnSettings).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create default warn settings if not found
				warnSettings = &WarnSettings{
					ChatId:    chatID,
					WarnLimit: 3,
				}
				if err := tx.Create(warnSettings).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Process each user
		for _, userID := range userIDs {
			warns := &Warns{}
			err := tx.Where("user_id = ? AND chat_id = ?", userID, chatID).First(warns).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// Create new warn record
					warns = &Warns{
						UserId:   userID,
						ChatId:   chatID,
						NumWarns: 1,
						Reasons:  StringArray{reason},
					}
					if err := tx.Create(warns).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				// Update existing warn record
				warns.NumWarns++
				warns.Reasons = append(warns.Reasons, reason)
				if err := tx.Save(warns).Error; err != nil {
					return err
				}
			}
		}

		// Invalidate cache after bulk warn
		deleteCache(warnSettingsCacheKey(chatID))
		return nil
	})
}

// BulkResetWarns resets warnings for multiple users in a single operation
func BulkResetWarns(chatID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	err := DB.Where("chat_id = ? AND user_id IN ?", chatID, userIDs).Delete(&Warns{}).Error
	if err != nil {
		log.Errorf("[Database][BulkResetWarns]: %d - %v", chatID, err)
		return err
	}

	// Invalidate cache after bulk reset
	deleteCache(warnSettingsCacheKey(chatID))
	return nil
}