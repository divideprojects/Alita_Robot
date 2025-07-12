package db

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Button represents a UI button that can be attached to messages.
//
// Fields:
//   - Name: The button's display text.
//   - Url: The URL the button links to.
//   - SameLine: Whether the button should appear on the same line as the previous button.
type Button struct {
	Name     string `bson:"name,omitempty" json:"name,omitempty"`
	Url      string `bson:"url,omitempty" json:"url,omitempty"`
	SameLine bool   `bson:"btn_sameline" json:"btn_sameline" default:"false"`
}

// Message type constants for different sender content types.
const (
	// TEXT types of senders
	TEXT      int = 1
	STICKER   int = 2
	DOCUMENT  int = 3
	PHOTO     int = 4
	AUDIO     int = 5
	VOICE     int = 6
	VIDEO     int = 7
	VideoNote int = 8
)

var (
	// Contexts
	tdCtx = context.TODO()
	bgCtx = context.Background()

	// define collections
	adminSettingsColl      *mongo.Collection
	blacklistsColl         *mongo.Collection
	pinColl                *mongo.Collection
	userColl               *mongo.Collection
	reportChatColl         *mongo.Collection
	reportUserColl         *mongo.Collection
	devsColl               *mongo.Collection
	chatColl               *mongo.Collection
	channelColl            *mongo.Collection
	antifloodSettingsColl  *mongo.Collection
	connectionColl         *mongo.Collection
	connectionSettingsColl *mongo.Collection
	disableColl            *mongo.Collection
	rulesColl              *mongo.Collection
	warnSettingsColl       *mongo.Collection
	warnUsersColl          *mongo.Collection
	greetingsColl          *mongo.Collection
	lockColl               *mongo.Collection
	filterColl             *mongo.Collection
	notesColl              *mongo.Collection
	notesSettingsColl      *mongo.Collection

	// Flag to track if database has been initialized
	isInitialized bool
)

// Initialize initializes the MongoDB client and opens all required collections.
// This function should be called explicitly with a config before using the database.
func Initialize(cfg *config.Config) error {
	if isInitialized {
		log.Warn("Database already initialized, skipping re-initialization")
		return nil
	}

	ctx, cancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURI))
	if err != nil {
		log.Errorf("[Database][Connect]: %v", err)
		return err
	}

	// Open Connections to Collections
	log.Info("Opening Database Collections...")
	adminSettingsColl = mongoClient.Database(cfg.MainDbName).Collection("admin")
	blacklistsColl = mongoClient.Database(cfg.MainDbName).Collection("blacklists")
	pinColl = mongoClient.Database(cfg.MainDbName).Collection("pins")
	userColl = mongoClient.Database(cfg.MainDbName).Collection("users")
	reportChatColl = mongoClient.Database(cfg.MainDbName).Collection("report_chat_settings")
	reportUserColl = mongoClient.Database(cfg.MainDbName).Collection("report_user_settings")
	devsColl = mongoClient.Database(cfg.MainDbName).Collection("devs")
	chatColl = mongoClient.Database(cfg.MainDbName).Collection("chats")
	channelColl = mongoClient.Database(cfg.MainDbName).Collection("channels")
	antifloodSettingsColl = mongoClient.Database(cfg.MainDbName).Collection("antiflood_settings")
	connectionColl = mongoClient.Database(cfg.MainDbName).Collection("connection")
	connectionSettingsColl = mongoClient.Database(cfg.MainDbName).Collection("connection_settings")
	disableColl = mongoClient.Database(cfg.MainDbName).Collection("disable")
	rulesColl = mongoClient.Database(cfg.MainDbName).Collection("rules")
	warnSettingsColl = mongoClient.Database(cfg.MainDbName).Collection("warns_settings")
	warnUsersColl = mongoClient.Database(cfg.MainDbName).Collection("warns_users")
	greetingsColl = mongoClient.Database(cfg.MainDbName).Collection("greetings")
	lockColl = mongoClient.Database(cfg.MainDbName).Collection("locks")
	filterColl = mongoClient.Database(cfg.MainDbName).Collection("filters")
	notesColl = mongoClient.Database(cfg.MainDbName).Collection("notes")
	notesSettingsColl = mongoClient.Database(cfg.MainDbName).Collection("notes_settings")

	isInitialized = true
	log.Info("Done opening all database collections!")
	return nil
}

// updateOne updates a single document in the specified collection.
// If no document matches the filter, a new one is inserted (upsert).
func updateOne(collecion *mongo.Collection, filter bson.M, data interface{}) (err error) {
	_, err = collecion.UpdateOne(tdCtx, filter, bson.M{"$set": data}, options.Update().SetUpsert(true))
	if err != nil {
		log.Errorf("[Database][updateOne]: %v", err)
	}
	return
}

// findOne finds a single document in the specified collection matching the filter.
func findOne(collecion *mongo.Collection, filter bson.M) (res *mongo.SingleResult) {
	res = collecion.FindOne(tdCtx, filter)
	return
}

// countDocs returns the number of documents in the collection matching the filter.
func countDocs(collecion *mongo.Collection, filter bson.M) (count int64, err error) {
	count, err = collecion.CountDocuments(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][countDocs]: %v", err)
	}
	return
}

// findAll returns a cursor for all documents in the collection matching the filter.
func findAll(collecion *mongo.Collection, filter bson.M) (cur *mongo.Cursor) {
	cur, err := collecion.Find(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][findAll]: %v", err)
	}
	return
}

// deleteOne deletes a single document from the collection matching the filter.
func deleteOne(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteOne(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteOne]: %v", err)
	}
	return
}

// deleteMany deletes multiple documents from the collection matching the filter.
func deleteMany(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteMany(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteMany]: %v", err)
	}
	return
}
