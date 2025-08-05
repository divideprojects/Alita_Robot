package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm/clause"
)

// Migration functions for each collection

func (m *Migrator) migrateUsers() error {
	collection := m.mongoDB.Collection("users")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d users", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		users := make([]PgUser, 0, len(docs))
		now := time.Now()

		for _, doc := range docs {
			user := PgUser{
				UserID:    toInt64(doc["_id"]),
				Username:  toString(doc["username"]),
				Name:      toString(doc["name"]),
				Language:  toString(doc["language"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}

			if user.Language == "" {
				user.Language = "en"
			}

			users = append(users, user)
		}

		if err := m.pgDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"username", "name", "language", "updated_at"}),
		}).CreateInBatches(users, 100).Error; err != nil {
			m.stats.FailedRecords += int64(len(docs))
			return fmt.Errorf("failed to insert users: %w", err)
		}

		m.stats.SuccessRecords += int64(len(docs))
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateChats() error {
	collection := m.mongoDB.Collection("chats")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d chats", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		chats := make([]PgChat, 0, len(docs))
		chatUsers := make([]PgChatUser, 0)
		now := time.Now()

		for _, doc := range docs {
			// Convert users array to JSON
			var usersJSON []byte
			if users, ok := doc["users"].([]interface{}); ok {
				usersJSON, _ = json.Marshal(users)

				// Also create chat_users entries
				chatID := toInt64(doc["_id"])
				for _, userID := range users {
					chatUsers = append(chatUsers, PgChatUser{
						ChatID: chatID,
						UserID: toInt64(userID),
					})
				}
			} else {
				usersJSON = []byte("[]")
			}

			chat := PgChat{
				ChatID:     toInt64(doc["_id"]),
				ChatName:   toString(doc["chat_name"]),
				Language:   toString(doc["language"]),
				Users:      usersJSON,
				IsInactive: toBool(doc["is_inactive"]),
				CreatedAt:  &now,
				UpdatedAt:  &now,
			}

			chats = append(chats, chat)
		}

		// Insert chats
		if err := m.pgDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "chat_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"chat_name", "language", "users", "is_inactive", "updated_at"}),
		}).CreateInBatches(chats, 100).Error; err != nil {
			m.stats.FailedRecords += int64(len(docs))
			return fmt.Errorf("failed to insert chats: %w", err)
		}

		// Insert chat_users
		if len(chatUsers) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
				DoNothing: true,
			}).CreateInBatches(chatUsers, 100).Error; err != nil {
				log.Printf("Warning: failed to insert chat_users: %v", err)
			}
		}

		m.stats.SuccessRecords += int64(len(docs))
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateAdmin() error {
	collection := m.mongoDB.Collection("admin")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d admin settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		admins := make([]PgAdmin, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping admin for non-existent chat: %d", chatID)
				}
				continue
			}

			admin := PgAdmin{
				ChatID:    chatID,
				AnonAdmin: toBool(doc["anon_admin"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			admins = append(admins, admin)
		}

		if len(admins) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"anon_admin", "updated_at"}),
			}).CreateInBatches(admins, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert admin settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d admin records with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(admins))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateNotesSettings() error {
	collection := m.mongoDB.Collection("notes_settings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d notes settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		settings := make([]PgNotesSettings, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping notes_settings for non-existent chat: %d", chatID)
				}
				continue
			}

			setting := PgNotesSettings{
				ChatID:    chatID,
				Private:   toBool(doc["private_notes"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			settings = append(settings, setting)
		}

		if len(settings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"private", "updated_at"}),
			}).CreateInBatches(settings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert notes settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d notes_settings records with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(settings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateNotes() error {
	collection := m.mongoDB.Collection("notes")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d notes", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		notes := make([]PgNote, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["chat_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping note for non-existent chat: %d", chatID)
				}
				continue
			}

			note := PgNote{
				ChatID:      chatID,
				NoteName:    toString(doc["note_name"]),
				NoteContent: toString(doc["note_content"]),
				MsgType:     int64(toInt64(doc["msgtype"])),
				CreatedAt:   &now,
				UpdatedAt:   &now,
			}
			notes = append(notes, note)
		}

		if len(notes) > 0 {
			// Use OnConflict to handle duplicates
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}, {Name: "note_name"}},
				DoUpdates: clause.AssignmentColumns([]string{"note_content", "msg_type", "updated_at"}),
			}).CreateInBatches(notes, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert notes: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d notes with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(notes))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateFilters() error {
	collection := m.mongoDB.Collection("filters")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d filters", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		// Use a map to deduplicate filters within the batch
		filterMap := make(map[string]PgFilter)
		skipped := 0
		duplicates := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["chat_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping filter for non-existent chat: %d", chatID)
				}
				continue
			}

			keyword := toString(doc["keyword"])
			key := fmt.Sprintf("%d:%s", chatID, keyword)

			// Check if we already have this filter in the batch
			if _, exists := filterMap[key]; exists {
				duplicates++
				if m.config.Verbose {
					log.Printf("  Skipping duplicate filter for chat %d, keyword: %s", chatID, keyword)
				}
				continue
			}

			filter := PgFilter{
				ChatID:      chatID,
				Keyword:     keyword,
				FilterReply: toString(doc["filter_reply"]),
				Msgtype:     int64(toInt64(doc["msgtype"])),
				CreatedAt:   &now,
				UpdatedAt:   &now,
			}
			filterMap[key] = filter
		}

		// Convert map to slice
		filters := make([]PgFilter, 0, len(filterMap))
		for _, filter := range filterMap {
			filters = append(filters, filter)
		}

		if len(filters) > 0 {
			// Use OnConflict to handle duplicate (chat_id, keyword) combinations across batches
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}, {Name: "keyword"}},
				DoUpdates: clause.AssignmentColumns([]string{"filter_reply", "msgtype", "updated_at"}),
			}).CreateInBatches(filters, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert filters: %w", err)
			}
		}

		if skipped > 0 || duplicates > 0 {
			log.Printf("  Skipped %d filters with invalid chat references, %d duplicates", skipped, duplicates)
		}

		m.stats.SuccessRecords += int64(len(filters))
		m.stats.FailedRecords += int64(skipped + duplicates)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}
func (m *Migrator) migrateGreetings() error {
	collection := m.mongoDB.Collection("greetings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d greetings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		greetings := make([]PgGreeting, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping greeting for non-existent chat: %d", chatID)
				}
				continue
			}

			greeting := PgGreeting{
				ChatID:               chatID,
				AutoApprove:          toBool(doc["auto_approve"]),
				CleanServiceSettings: toBool(doc["clean_service_settings"]),
				CreatedAt:            &now,
				UpdatedAt:            &now,
			}

			// Process welcome settings
			if welcome, ok := doc["welcome_settings"].(map[string]interface{}); ok {
				greeting.WelcomeEnabled = toBool(welcome["enabled"])
				greeting.WelcomeText = toString(welcome["text"])
				greeting.WelcomeType = toInt64(welcome["type"])
				greeting.WelcomeCleanOld = toBool(welcome["clean_old"])
			}

			// Process goodbye settings
			if goodbye, ok := doc["goodbye_settings"].(map[string]interface{}); ok {
				greeting.GoodbyeEnabled = toBool(goodbye["enabled"])
				greeting.GoodbyeText = toString(goodbye["text"])
				greeting.GoodbyeType = toInt64(goodbye["type"])
				greeting.GoodbyeCleanOld = toBool(goodbye["clean_old"])
			}

			greetings = append(greetings, greeting)
		}

		if len(greetings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"auto_approve", "clean_service_settings",
					"welcome_enabled", "welcome_text", "welcome_type", "welcome_clean_old",
					"goodbye_enabled", "goodbye_text", "goodbye_type", "goodbye_clean_old",
					"updated_at",
				}),
			}).CreateInBatches(greetings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert greetings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d greetings with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(greetings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateLocks() error {
	collection := m.mongoDB.Collection("locks")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d lock settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		locks := make([]PgLock, 0)
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping locks for non-existent chat: %d", chatID)
				}
				continue
			}

			// Process permissions
			if permissions, ok := doc["permissions"].(map[string]interface{}); ok {
				for lockType, locked := range permissions {
					lock := PgLock{
						ChatID:    chatID,
						LockType:  lockType,
						Locked:    toBool(locked),
						CreatedAt: &now,
						UpdatedAt: &now,
					}
					locks = append(locks, lock)
				}
			}

			// Process restrictions
			if restrictions, ok := doc["restrictions"].(map[string]interface{}); ok {
				for lockType, locked := range restrictions {
					lock := PgLock{
						ChatID:    chatID,
						LockType:  lockType,
						Locked:    toBool(locked),
						CreatedAt: &now,
						UpdatedAt: &now,
					}
					locks = append(locks, lock)
				}
			}
		}

		if len(locks) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}, {Name: "lock_type"}},
				DoUpdates: clause.AssignmentColumns([]string{"locked", "updated_at"}),
			}).CreateInBatches(locks, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert locks: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d lock records with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(locks))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migratePins() error {
	collection := m.mongoDB.Collection("pins")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d pin settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		pins := make([]PgPin, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping pins for non-existent chat: %d", chatID)
				}
				continue
			}

			pin := PgPin{
				ChatID:         chatID,
				AntiChannelPin: toBool(doc["antichannelpin"]),
				CleanLinked:    toBool(doc["cleanlinked"]),
				CreatedAt:      &now,
				UpdatedAt:      &now,
			}
			pins = append(pins, pin)
		}

		if len(pins) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"anti_channel_pin", "clean_linked", "updated_at"}),
			}).CreateInBatches(pins, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert pins: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d pins with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(pins))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateRules() error {
	collection := m.mongoDB.Collection("rules")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d rules", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		rules := make([]PgRule, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping rules for non-existent chat: %d", chatID)
				}
				continue
			}

			rule := PgRule{
				ChatID:    chatID,
				Rules:     toString(doc["rules"]),
				Private:   toBool(doc["privrules"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			rules = append(rules, rule)
		}

		if len(rules) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"rules", "private", "updated_at"}),
			}).CreateInBatches(rules, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert rules: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d rules with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(rules))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateWarnsSettings() error {
	collection := m.mongoDB.Collection("warns_settings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d warns settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		settings := make([]PgWarnsSetting, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping warns_settings for non-existent chat: %d", chatID)
				}
				continue
			}

			setting := PgWarnsSetting{
				ChatID:    chatID,
				WarnLimit: toInt64(doc["warn_limit"]),
				WarnMode:  toString(doc["warn_mode"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}

			if setting.WarnLimit == 0 {
				setting.WarnLimit = 3
			}

			settings = append(settings, setting)
		}

		if len(settings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"warn_limit", "warn_mode", "updated_at"}),
			}).CreateInBatches(settings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert warns settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d warns_settings with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(settings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateWarnsUsers() error {
	collection := m.mongoDB.Collection("warns_users")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d warns users", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		// Use a map to deduplicate within the batch
		warnsMap := make(map[string]PgWarnsUser)
		skipped := 0
		duplicates := 0
		now := time.Now()

		for _, doc := range docs {
			userID := toInt64(doc["user_id"])
			chatID := toInt64(doc["chat_id"])

			// Skip if user or chat doesn't exist
			if !m.validUserIDs[userID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping warns_users for non-existent user: %d", userID)
				}
				continue
			}
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping warns_users for non-existent chat: %d", chatID)
				}
				continue
			}

			key := fmt.Sprintf("%d:%d", userID, chatID)

			// Check if we already have this combination in the batch
			if _, exists := warnsMap[key]; exists {
				duplicates++
				if m.config.Verbose {
					log.Printf("  Skipping duplicate warns_users for user %d, chat %d", userID, chatID)
				}
				continue
			}

			// Convert warns array to JSON
			var warnsJSON []byte
			if warns, ok := doc["warns"].([]interface{}); ok {
				warnsJSON, _ = json.Marshal(warns)
			} else {
				warnsJSON = []byte("[]")
			}

			warnsUser := PgWarnsUser{
				UserID:    userID,
				ChatID:    chatID,
				NumWarns:  toInt64(doc["num_warns"]),
				Warns:     warnsJSON,
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			warnsMap[key] = warnsUser
		}

		// Convert map to slice
		warnsUsers := make([]PgWarnsUser, 0, len(warnsMap))
		for _, wu := range warnsMap {
			warnsUsers = append(warnsUsers, wu)
		}

		if len(warnsUsers) > 0 {
			// Use OnConflict to handle duplicates across batches
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"num_warns", "warns", "updated_at"}),
			}).CreateInBatches(warnsUsers, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert warns users: %w", err)
			}
		}

		if skipped > 0 || duplicates > 0 {
			log.Printf("  Skipped %d warns_users with invalid references, %d duplicates", skipped, duplicates)
		}

		m.stats.SuccessRecords += int64(len(warnsUsers))
		m.stats.FailedRecords += int64(skipped + duplicates)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}
func (m *Migrator) migrateAntifloodSettings() error {
	collection := m.mongoDB.Collection("antiflood_settings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d antiflood settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		settings := make([]PgAntifloodSetting, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping antiflood_settings for non-existent chat: %d", chatID)
				}
				continue
			}

			setting := PgAntifloodSetting{
				ChatID:                 chatID,
				Limit:                  toInt64(doc["limit"]),
				Mode:                   toString(doc["mode"]),
				DeleteAntifloodMessage: toBool(doc["del_msg"]),
				CreatedAt:              &now,
				UpdatedAt:              &now,
			}

			if setting.Limit == 0 {
				setting.Limit = 5
			}
			if setting.Mode == "" {
				setting.Mode = "mute"
			}

			settings = append(settings, setting)
		}

		if len(settings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"flood_limit", "mode", "delete_antiflood_message", "action", "updated_at"}),
			}).CreateInBatches(settings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert antiflood settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d antiflood_settings with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(settings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateBlacklists() error {
	collection := m.mongoDB.Collection("blacklists")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d blacklist settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		blacklists := make([]PgBlacklist, 0)
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping blacklists for non-existent chat: %d", chatID)
				}
				continue
			}

			action := toString(doc["action"])
			reason := toString(doc["reason"])

			// Set defaults
			if action == "" || action == "none" {
				action = "warn"
			}
			if reason == "" {
				reason = "Automated Blacklisted word %s"
			}

			// Process triggers array - each trigger becomes a separate row
			if triggers, ok := doc["triggers"].([]interface{}); ok {
				for _, trigger := range triggers {
					word := toString(trigger)
					if word != "" {
						blacklist := PgBlacklist{
							ChatID:    chatID,
							Word:      strings.ToLower(word),
							Action:    action,
							Reason:    reason,
							CreatedAt: &now,
							UpdatedAt: &now,
						}
						blacklists = append(blacklists, blacklist)
					}
				}
			}
		}

		if len(blacklists) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}, {Name: "word"}},
				DoUpdates: clause.AssignmentColumns([]string{"action", "reason", "updated_at"}),
			}).CreateInBatches(blacklists, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert blacklists: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d blacklist records with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(blacklists))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}
func (m *Migrator) migrateChannels() error {
	collection := m.mongoDB.Collection("channels")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d channels", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		channels := make([]PgChannel, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])
			channelID := toInt64(doc["_id"]) // Using same ID as channel ID

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping channel for non-existent chat: %d", chatID)
				}
				continue
			}

			// Also check if channel_id references a valid chat (if it's different)
			if channelID != 0 && channelID != chatID && !m.validChatIDs[channelID] {
				// Set to NULL if channel doesn't exist
				channelID = 0
			}

			channel := PgChannel{
				ChatID:    chatID,
				ChannelID: channelID,
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			channels = append(channels, channel)
		}

		if len(channels) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"channel_id", "updated_at"}),
			}).CreateInBatches(channels, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert channels: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d channels with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(channels))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateConnections() error {
	collection := m.mongoDB.Collection("connection")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d connections", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		// Use a map to deduplicate within the batch
		connMap := make(map[string]PgConnection)
		skipped := 0
		duplicates := 0
		now := time.Now()

		for _, doc := range docs {
			userID := toInt64(doc["_id"])
			chatID := toInt64(doc["chat_id"])

			// Skip if user or chat doesn't exist
			if !m.validUserIDs[userID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping connection for non-existent user: %d", userID)
				}
				continue
			}
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping connection for non-existent chat: %d", chatID)
				}
				continue
			}

			key := fmt.Sprintf("%d:%d", userID, chatID)

			// Check if we already have this combination in the batch
			if _, exists := connMap[key]; exists {
				duplicates++
				if m.config.Verbose {
					log.Printf("  Skipping duplicate connection for user %d, chat %d", userID, chatID)
				}
				continue
			}

			connection := PgConnection{
				UserID:    userID,
				ChatID:    chatID,
				Connected: toBool(doc["connected"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			connMap[key] = connection
		}

		// Convert map to slice
		connections := make([]PgConnection, 0, len(connMap))
		for _, conn := range connMap {
			connections = append(connections, conn)
		}

		if len(connections) > 0 {
			// Use OnConflict to handle duplicates across batches
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"connected", "updated_at"}),
			}).CreateInBatches(connections, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert connections: %w", err)
			}
		}

		if skipped > 0 || duplicates > 0 {
			log.Printf("  Skipped %d connections with invalid references, %d duplicates", skipped, duplicates)
		}

		m.stats.SuccessRecords += int64(len(connections))
		m.stats.FailedRecords += int64(skipped + duplicates)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}
func (m *Migrator) migrateConnectionSettings() error {
	collection := m.mongoDB.Collection("connection_settings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d connection settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		settings := make([]PgConnectionSetting, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping connection_settings for non-existent chat: %d", chatID)
				}
				continue
			}

			setting := PgConnectionSetting{
				ChatID:       chatID,
				// Inverted logic: MongoDB's "can_connect" means "disallow connect" when false, 
				// but PostgreSQL's "allow_connect" means "allow connect" when true.
				// See migration guide section "Connection Settings Boolean Inversion" for details.
				AllowConnect: !toBool(doc["can_connect"]),
				Enabled:      true,
				CreatedAt:    &now,
				UpdatedAt:    &now,
			}
			settings = append(settings, setting)
		}

		if len(settings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"allow_connect", "enabled", "updated_at"}),
			}).CreateInBatches(settings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert connection settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d connection_settings with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(settings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateDisable() error {
	collection := m.mongoDB.Collection("disable")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d disable settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		disables := make([]PgDisable, 0)
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping disable settings for non-existent chat: %d", chatID)
				}
				continue
			}

			// Process commands array
			if commands, ok := doc["commands"].([]interface{}); ok {
				for _, cmd := range commands {
					disable := PgDisable{
						ChatID:    chatID,
						Command:   toString(cmd),
						Disabled:  true,
						CreatedAt: &now,
						UpdatedAt: &now,
					}
					disables = append(disables, disable)
				}
			}
		}

		if len(disables) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}, {Name: "command"}},
				DoUpdates: clause.AssignmentColumns([]string{"disabled", "updated_at"}),
			}).CreateInBatches(disables, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert disable settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d disable records with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(disables))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateReportUserSettings() error {
	collection := m.mongoDB.Collection("report_user_settings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d report user settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		settings := make([]PgReportUserSetting, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			userID := toInt64(doc["_id"])

			// Skip if user doesn't exist
			if !m.validUserIDs[userID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping report_user_settings for non-existent user: %d", userID)
				}
				continue
			}

			setting := PgReportUserSetting{
				UserID:    userID,
				Status:    toBool(doc["status"]),
				Enabled:   toBool(doc["status"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			settings = append(settings, setting)
		}

		if len(settings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"status", "enabled", "updated_at"}),
			}).CreateInBatches(settings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert report user settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d report_user_settings with invalid user references", skipped)
		}

		m.stats.SuccessRecords += int64(len(settings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}

func (m *Migrator) migrateReportChatSettings() error {
	collection := m.mongoDB.Collection("report_chat_settings")

	return m.processBatch(collection, func(docs []bson.M) error {
		if m.config.DryRun {
			log.Printf("  [DRY RUN] Would migrate %d report chat settings", len(docs))
			m.stats.SuccessRecords += int64(len(docs))
			return nil
		}

		settings := make([]PgReportChatSetting, 0, len(docs))
		skipped := 0
		now := time.Now()

		for _, doc := range docs {
			chatID := toInt64(doc["_id"])

			// Skip if chat doesn't exist
			if !m.validChatIDs[chatID] {
				skipped++
				if m.config.Verbose {
					log.Printf("  Skipping report_chat_settings for non-existent chat: %d", chatID)
				}
				continue
			}

			setting := PgReportChatSetting{
				ChatID:    chatID,
				Status:    toBool(doc["status"]),
				Enabled:   toBool(doc["status"]),
				CreatedAt: &now,
				UpdatedAt: &now,
			}
			settings = append(settings, setting)
		}

		if len(settings) > 0 {
			if err := m.pgDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chat_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"status", "enabled", "updated_at"}),
			}).CreateInBatches(settings, 100).Error; err != nil {
				m.stats.FailedRecords += int64(len(docs))
				return fmt.Errorf("failed to insert report chat settings: %w", err)
			}
		}

		if skipped > 0 {
			log.Printf("  Skipped %d report_chat_settings with invalid chat references", skipped)
		}

		m.stats.SuccessRecords += int64(len(settings))
		m.stats.FailedRecords += int64(skipped)
		m.stats.TotalRecords += int64(len(docs))
		return nil
	})
}
