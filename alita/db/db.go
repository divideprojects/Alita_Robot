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
	// Package-level MongoDB client
	mongoClient *mongo.Client

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
)

// createIndexes creates database indexes for optimal performance
func createIndexes() {
	log.Info("Creating database indexes...")

	// Filter collection indexes
	_, err := filterColl.Indexes().CreateMany(bgCtx, []mongo.IndexModel{
		// Existing unique index
		{
			Keys:    bson.D{{Key: "chat_id", Value: 1}, {Key: "keyword", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// Pagination optimization: (_id) for cursor-based pagination
		{
			Keys: bson.D{{Key: "_id", Value: 1}},
		},
		// Compound index for paginated queries
		{
			Keys: bson.D{
				{Key: "chat_id", Value: 1},
				{Key: "_id", Value: 1},
			},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create filter indexes: %v", err)
	}

	// Notes collection indexes
	_, err = notesColl.Indexes().CreateMany(bgCtx, []mongo.IndexModel{
		// Existing unique index
		{
			Keys:    bson.D{{Key: "chat_id", Value: 1}, {Key: "note_name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// Pagination optimization
		{
			Keys: bson.D{{Key: "_id", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "chat_id", Value: 1},
				{Key: "_id", Value: 1},
			},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create notes indexes: %v", err)
	}

	log.Info("Done creating database indexes!")
}

// init initializes the MongoDB client and opens all required collections.
//
// This function is automatically called when the package is imported.
// It sets up global collection variables for use throughout the db package.
func init() {
	var err error
	
	ctx, cancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer cancel()

	// Use modern single-step client creation pattern (MongoDB driver v1.4+)
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseURI))
	if err != nil {
		log.Errorf("[Database][Connect]: %v", err)
		// Return early if connection fails to prevent nil pointer panic
		return
	}

	// Open Connections to Collections
	log.Info("Opening Database Collections...")
	log.Debugf("[DB] Initializing collections with client status: %t", mongoClient != nil)
	adminSettingsColl = mongoClient.Database(config.MainDbName).Collection("admin")
	blacklistsColl = mongoClient.Database(config.MainDbName).Collection("blacklists")
	pinColl = mongoClient.Database(config.MainDbName).Collection("pins")
	userColl = mongoClient.Database(config.MainDbName).Collection("users")
	reportChatColl = mongoClient.Database(config.MainDbName).Collection("report_chat_settings")
	reportUserColl = mongoClient.Database(config.MainDbName).Collection("report_user_settings")
	devsColl = mongoClient.Database(config.MainDbName).Collection("devs")
	chatColl = mongoClient.Database(config.MainDbName).Collection("chats")
	channelColl = mongoClient.Database(config.MainDbName).Collection("channels")
	antifloodSettingsColl = mongoClient.Database(config.MainDbName).Collection("antiflood_settings")
	connectionColl = mongoClient.Database(config.MainDbName).Collection("connection")
	connectionSettingsColl = mongoClient.Database(config.MainDbName).Collection("connection_settings")
	disableColl = mongoClient.Database(config.MainDbName).Collection("disable")
	rulesColl = mongoClient.Database(config.MainDbName).Collection("rules")
	warnSettingsColl = mongoClient.Database(config.MainDbName).Collection("warns_settings")
	warnUsersColl = mongoClient.Database(config.MainDbName).Collection("warns_users")
	greetingsColl = mongoClient.Database(config.MainDbName).Collection("greetings")
	lockColl = mongoClient.Database(config.MainDbName).Collection("locks")
	filterColl = mongoClient.Database(config.MainDbName).Collection("filters")
	notesColl = mongoClient.Database(config.MainDbName).Collection("notes")
	notesSettingsColl = mongoClient.Database(config.MainDbName).Collection("notes_settings")
	log.Info("Done opening all database collections!")

	// Create indexes for optimal performance
	createIndexes()
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

// deleteMany deletes all documents from the collection matching the filter.
func deleteMany(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteMany(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteMany]: %v", err)
	}
	return
}

// findOneAndUpsert performs an atomic find-and-update operation with upsert.
// Returns the document after the operation (either existing or newly created).
func findOneAndUpsert(collection *mongo.Collection, filter bson.M, update bson.M, result interface{}) error {
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	err := collection.FindOneAndUpdate(tdCtx, filter, update, opts).Decode(result)
	if err != nil {
		log.Errorf("[Database][findOneAndUpsert]: %v", err)
	}
	return err
}

// GetTestCollection returns a collection for benchmark testing
func GetTestCollection() *mongo.Collection {
	return getCollection("benchmark_test")
}

// getCollection is a helper to safely access collections
func getCollection(name string) *mongo.Collection {
	if mongoClient == nil {
		return nil
	}
	return mongoClient.Database(config.MainDbName).Collection(name)
}
