package db

import (
	"context"
	"fmt"
	"math/rand"
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
	captchasColl           *mongo.Collection
	captchaChallengesColl  *mongo.Collection
)

// createIndexes creates database indexes for optimal performance
func createIndexes(ctx context.Context) {
	log.Info("Creating database indexes...")

	// Filter collection indexes
	_, err := filterColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
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
	_, err = notesColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
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

	// Warn users collection indexes
	// Compound index on (user_id, chat_id) to speed up warning lookups and enforce uniqueness per user per chat
	_, err = warnUsersColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "chat_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create warnUsers indexes: %v", err)
	}

	// User collection indexes
	// Unique indexes on username and user_id for fast lookups and to prevent duplicates
	_, err = userColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create user indexes: %v", err)
	}

	// Chat collection indexes
	// Unique index on chat_id for fast chat lookups, index on chat_type for filtering
	_, err = chatColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "chat_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "chat_type", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create chat indexes: %v", err)
	}

	// Captcha collection indexes
	// Compound index on (user_id, chat_id) for challenge lookups, index on message_id for message-based queries
	_, err = captchasColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "chat_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "message_id", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create captcha indexes: %v", err)
	}

	// Admin collection indexes
	// Compound unique index on (user_id, chat_id) to ensure one admin record per user per chat
	_, err = adminSettingsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "chat_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create admin indexes: %v", err)
	}

	// Antiflood collection indexes
	// Compound index on (user_id, chat_id) for fast flood checks per user per chat
	_, err = antifloodSettingsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "chat_id", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create antiflood indexes: %v", err)
	}

	// Blacklists collection indexes
	// Compound unique index on (chat_id, trigger) to prevent duplicate triggers per chat
	_, err = blacklistsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "chat_id", Value: 1}, {Key: "trigger", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create blacklist indexes: %v", err)
	}

	// Greetings collection indexes
	// Index on chat_id for fast greeting settings lookup per chat
	_, err = greetingsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "chat_id", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create greetings indexes: %v", err)
	}

	// Locks collection indexes
	// Index on chat_id for fast lock settings lookup per chat
	_, err = lockColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "chat_id", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create locks indexes: %v", err)
	}

	// Reports collection indexes
	// Index on chat_id for fast report settings lookup per chat
	_, err = reportChatColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "chat_id", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create reports indexes: %v", err)
	}

	// Rules collection indexes
	// Index on chat_id for fast rules lookup per chat
	_, err = rulesColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "chat_id", Value: 1}},
		},
	})
	if err != nil {
		log.Warnf("[Database][Index] Failed to create rules indexes: %v", err)
	}

	log.Info("Done creating database indexes!")
}

// InitializeDatabase initializes the MongoDB client and opens all required collections.
// This function should be called by the lifecycle manager during startup.
// It sets up global collection variables for use throughout the db package.
func InitializeDatabase(ctx context.Context) error {
	var err error
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(config.DatabaseURI).
		SetMaxPoolSize(config.MongoMaxPoolSize).
		SetMinPoolSize(config.MongoMinPoolSize).
		SetMaxConnIdleTime(config.MongoMaxConnIdleTime)

	mongoClient, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Errorf("[Database][Connect]: %v", err)
		return err
	}

	// Test connection
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Errorf("[Database][Ping]: %v", err)
		return err
	}

	// Initialize collections
	initializeCollections()

	// Create indexes for optimal performance
	createIndexes(ctx)

	log.Info("Database initialized successfully")
	return nil
}

// initializeCollections initializes all database collections
func initializeCollections() {
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
	captchasColl = mongoClient.Database(config.MainDbName).Collection("captchas")
	captchaChallengesColl = mongoClient.Database(config.MainDbName).Collection("captcha_challenges")

	log.Info("Done opening all database collections!")
}

// CloseDatabase closes the MongoDB client connection
func CloseDatabase(ctx context.Context) error {
	if mongoClient != nil {
		log.Info("Closing database connection...")
		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Errorf("[Database][Disconnect]: %v", err)
			return err
		}
		mongoClient = nil
		log.Info("Database connection closed")
	}
	return nil
}

// HealthCheck verifies the database connection is healthy
func HealthCheck(ctx context.Context) error {
	if mongoClient == nil {
		return fmt.Errorf("database client is not initialized")
	}
	return mongoClient.Ping(ctx, nil)
}

// init function is kept minimal to prevent automatic initialization
func init() {
	// Initialize only basic variables, actual database connection
	// will be handled by InitializeDatabase function
	log.Debug("Database package initialized (connection deferred)")
}

// Helper for retrying DB ops
func retryDB(fn func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < 2 {
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
		}
	}
	return err
}

// updateOne with timing, retry, and slow query log
func updateOne(collection *mongo.Collection, filter bson.M, data interface{}) (err error) {
	return updateOneWithContext(context.Background(), collection, filter, data)
}

// updateOneWithContext with timing, retry, and slow query log
func updateOneWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M, data interface{}) (err error) {
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	err = retryDB(func() error {
		_, e := collection.UpdateOne(ctx, filter, bson.M{"$set": data}, options.Update().SetUpsert(true))
		return e
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][updateOne] %v %v took %v", collection.Name(), filter, dur)
	}
	if err != nil {
		log.Errorf("[Database][updateOne]: %v", err)
	}
	return
}

// findOne with timing, retry, and slow query log
func findOne(collection *mongo.Collection, filter bson.M) (res *mongo.SingleResult) {
	return findOneWithContext(context.Background(), collection, filter)
}

// findOneWithContext with timing, retry, and slow query log
func findOneWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (res *mongo.SingleResult) {
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	var result *mongo.SingleResult
	retryDB(func() error {
		result = collection.FindOne(ctx, filter)
		return result.Err()
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][findOne] %v %v took %v", collection.Name(), filter, dur)
	}
	return result
}

// countDocs with timing, retry, and slow query log
func countDocs(collection *mongo.Collection, filter bson.M) (count int64, err error) {
	return countDocsWithContext(context.Background(), collection, filter)
}

// countDocsWithContext with timing, retry, and slow query log
func countDocsWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (count int64, err error) {
	start := time.Now()
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = retryDB(func() error {
		c, e := collection.CountDocuments(ctx, filter)
		count = c
		return e
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][countDocs] %v %v took %v", collection.Name(), filter, dur)
	}
	if err != nil {
		log.Errorf("[Database][countDocs]: %v", err)
	}
	return
}

// findAll with timing, retry, and slow query log
func findAll(collection *mongo.Collection, filter bson.M) (cur *mongo.Cursor) {
	return findAllWithContext(context.Background(), collection, filter)
}

// findAllWithContext with timing, retry, and slow query log
func findAllWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (cur *mongo.Cursor) {
	start := time.Now()
	var cursor *mongo.Cursor
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	retryDB(func() error {
		c, e := collection.Find(ctx, filter)
		cursor = c
		return e
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][findAll] %v %v took %v", collection.Name(), filter, dur)
	}
	return cursor
}

// deleteOne with timing, retry, and slow query log
func deleteOne(collection *mongo.Collection, filter bson.M) (err error) {
	return deleteOneWithContext(context.Background(), collection, filter)
}

// deleteOneWithContext with timing, retry, and slow query log
func deleteOneWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (err error) {
	start := time.Now()
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = retryDB(func() error {
		_, e := collection.DeleteOne(ctx, filter)
		return e
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][deleteOne] %v %v took %v", collection.Name(), filter, dur)
	}
	if err != nil {
		log.Errorf("[Database][deleteOne]: %v", err)
	}
	return
}

// deleteMany with timing, retry, and slow query log
func deleteMany(collection *mongo.Collection, filter bson.M) (err error) {
	return deleteManyWithContext(context.Background(), collection, filter)
}

// deleteManyWithContext with timing, retry, and slow query log
func deleteManyWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (err error) {
	start := time.Now()
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = retryDB(func() error {
		_, e := collection.DeleteMany(ctx, filter)
		return e
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][deleteMany] %v %v took %v", collection.Name(), filter, dur)
	}
	if err != nil {
		log.Errorf("[Database][deleteMany]: %v", err)
	}
	return
}

// findOneAndUpsert with timing, retry, and slow query log
func findOneAndUpsert(collection *mongo.Collection, filter bson.M, update bson.M, result interface{}) error {
	return findOneAndUpsertWithContext(context.Background(), collection, filter, update, result)
}

// findOneAndUpsertWithContext with timing, retry, and slow query log
func findOneAndUpsertWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M, update bson.M, result interface{}) error {
	start := time.Now()
	var err error
	// Create a context with timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = retryDB(func() error {
		opts := options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After)
		err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(result)
		return err
	})
	dur := time.Since(start)
	if dur > 100*time.Millisecond {
		log.Warnf("[Database][SLOW][findOneAndUpsert] %v %v took %v", collection.Name(), filter, dur)
	}
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
