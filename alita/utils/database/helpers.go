package database

import (
	"context"
	"time"

	"github.com/divideprojects/Alita_Robot/alita/db"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var bgCtx = context.Background()

// SettingsManager provides high-level database operations for settings management
type SettingsManager struct {
	collection *mongo.Collection
}

// NewSettingsManager creates a new settings manager for a collection
func NewSettingsManager(collection *mongo.Collection) *SettingsManager {
	return &SettingsManager{
		collection: collection,
	}
}

// GetChatSetting retrieves a setting for a specific chat
func (sm *SettingsManager) GetChatSetting(chatID int64, settingName string, defaultValue interface{}) interface{} {
	filter := bson.M{"chat_id": chatID}
	var result bson.M
	err := sm.collection.FindOne(bgCtx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return defaultValue
		}
		log.Errorf("[SettingsManager][GetChatSetting]: %v", err)
		return defaultValue
	}

	if value, exists := result[settingName]; exists {
		return value
	}
	return defaultValue
}

// SetChatSetting sets a setting for a specific chat
func (sm *SettingsManager) SetChatSetting(chatID int64, settingName string, value interface{}) error {
	filter := bson.M{"chat_id": chatID}
	return db.UpdateField(sm.collection, filter, settingName, value)
}

// ToggleChatSetting toggles a boolean setting for a specific chat
func (sm *SettingsManager) ToggleChatSetting(chatID int64, settingName string) (bool, error) {
	filter := bson.M{"chat_id": chatID}
	return db.ToggleField(sm.collection, filter, settingName)
}

// GetUserSetting retrieves a setting for a specific user
func (sm *SettingsManager) GetUserSetting(userID int64, settingName string, defaultValue interface{}) interface{} {
	filter := bson.M{"user_id": userID}
	var result bson.M
	err := sm.collection.FindOne(bgCtx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return defaultValue
		}
		log.Errorf("[SettingsManager][GetUserSetting]: %v", err)
		return defaultValue
	}

	if value, exists := result[settingName]; exists {
		return value
	}
	return defaultValue
}

// SetUserSetting sets a setting for a specific user
func (sm *SettingsManager) SetUserSetting(userID int64, settingName string, value interface{}) error {
	filter := bson.M{"user_id": userID}
	return db.UpdateField(sm.collection, filter, settingName, value)
}

// Common query patterns

// GetChatsByFeature returns all chats that have a specific feature enabled
func GetChatsByFeature(collection *mongo.Collection, featureName string) ([]int64, error) {
	filter := bson.M{featureName: true}
	cursor, err := collection.Find(bgCtx, filter)
	if err != nil {
		log.Errorf("[GetChatsByFeature]: %v", err)
		return nil, err
	}
	defer cursor.Close(bgCtx)

	var chatIDs []int64
	for cursor.Next(bgCtx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Errorf("[GetChatsByFeature][Decode]: %v", err)
			continue
		}
		if chatID, exists := result["chat_id"]; exists {
			if id, ok := chatID.(int64); ok {
				chatIDs = append(chatIDs, id)
			}
		}
	}

	return chatIDs, nil
}

// GetUsersByFeature returns all users that have a specific feature enabled
func GetUsersByFeature(collection *mongo.Collection, featureName string) ([]int64, error) {
	filter := bson.M{featureName: true}
	cursor, err := collection.Find(bgCtx, filter)
	if err != nil {
		log.Errorf("[GetUsersByFeature]: %v", err)
		return nil, err
	}
	defer cursor.Close(bgCtx)

	var userIDs []int64
	for cursor.Next(bgCtx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Errorf("[GetUsersByFeature][Decode]: %v", err)
			continue
		}
		if userID, exists := result["user_id"]; exists {
			if id, ok := userID.(int64); ok {
				userIDs = append(userIDs, id)
			}
		}
	}

	return userIDs, nil
}

// BulkUpdateSettings updates multiple settings for multiple chats/users
func BulkUpdateSettings(collection *mongo.Collection, updates []BulkUpdate) error {
	var operations []mongo.WriteModel

	for _, update := range updates {
		operation := mongo.NewUpdateOneModel().
			SetFilter(update.Filter).
			SetUpdate(bson.M{"$set": update.Update}).
			SetUpsert(true)
		operations = append(operations, operation)
	}

	if len(operations) == 0 {
		return nil
	}

	_, err := collection.BulkWrite(bgCtx, operations)
	if err != nil {
		log.Errorf("[BulkUpdateSettings]: %v", err)
	}
	return err
}

// BulkUpdate represents a single update operation in a bulk update
type BulkUpdate struct {
	Filter bson.M
	Update bson.M
}

// CleanupOldRecords removes records older than the specified number of days
func CleanupOldRecords(collection *mongo.Collection, dateField string, daysOld int) (int64, error) {
	// Calculate the cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)

	filter := bson.M{dateField: bson.M{"$lt": cutoffDate}}
	result, err := collection.DeleteMany(bgCtx, filter)
	if err != nil {
		log.Errorf("[CleanupOldRecords]: %v", err)
		return 0, err
	}

	return result.DeletedCount, nil
}

// GetStatistics returns basic statistics for a collection
func GetStatistics(collection *mongo.Collection) (*CollectionStats, error) {
	totalCount, err := db.GetCount(collection, bson.M{})
	if err != nil {
		return nil, err
	}

	// Get count of active records (assuming there's an 'active' field)
	activeCount, _ := db.GetCount(collection, bson.M{"active": true})

	return &CollectionStats{
		TotalRecords:  totalCount,
		ActiveRecords: activeCount,
	}, nil
}

// CollectionStats represents basic statistics for a collection
type CollectionStats struct {
	TotalRecords  int64
	ActiveRecords int64
}

// Migration helpers

// MigrateField migrates data from one field to another
func MigrateField(collection *mongo.Collection, oldField, newField string) error {
	// Find all documents that have the old field but not the new field
	filter := bson.M{
		oldField: bson.M{"$exists": true},
		newField: bson.M{"$exists": false},
	}

	cursor, err := collection.Find(bgCtx, filter)
	if err != nil {
		log.Errorf("[MigrateField]: %v", err)
		return err
	}
	defer cursor.Close(bgCtx)

	var updates []BulkUpdate
	for cursor.Next(bgCtx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Errorf("[MigrateField][Decode]: %v", err)
			continue
		}

		if oldValue, exists := doc[oldField]; exists {
			updates = append(updates, BulkUpdate{
				Filter: bson.M{"_id": doc["_id"]},
				Update: bson.M{
					newField: oldValue,
					"$unset": bson.M{oldField: ""},
				},
			})
		}
	}

	if len(updates) > 0 {
		return BulkUpdateSettings(collection, updates)
	}

	return nil
}

// AddMissingDefaults adds default values for missing fields
func AddMissingDefaults(collection *mongo.Collection, defaults map[string]interface{}) error {
	var updates []BulkUpdate

	for field, defaultValue := range defaults {
		filter := bson.M{field: bson.M{"$exists": false}}
		update := bson.M{field: defaultValue}

		updates = append(updates, BulkUpdate{
			Filter: filter,
			Update: update,
		})
	}

	return BulkUpdateSettings(collection, updates)
}
