package db

import (
	"context"
	"sync"
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

// Database represents the database connection and collections
type Database struct {
	client *mongo.Client
	db     *mongo.Database
	mu     sync.RWMutex

	// Collections
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
}

var (
	dbInstance *Database
	once       sync.Once

	// Contexts for backward compatibility
	bgCtx = context.Background()
)

// Global collection variables for backward compatibility
var (
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

// GetDatabase returns the singleton database instance
func GetDatabase() *Database {
	once.Do(func() {
		dbInstance = &Database{}
		if err := dbInstance.Connect(); err != nil {
			log.Fatalf("[Database][Connect]: %v", err)
		}
	})
	return dbInstance
}

// Connect establishes connection to MongoDB and initializes collections
func (d *Database) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configure client options with connection pooling
	clientOptions := options.Client().
		ApplyURI(config.DatabaseURI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetSocketTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.client = client
	d.db = client.Database(config.MainDbName)

	// Initialize collections
	d.initCollections()

	// Set global variables for backward compatibility
	d.setGlobalCollections()

	log.Info("Database connection established successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return d.client.Disconnect(ctx)
}

// initCollections initializes all collection references
func (d *Database) initCollections() {
	d.adminSettingsColl = d.db.Collection("admin")
	d.blacklistsColl = d.db.Collection("blacklists")
	d.pinColl = d.db.Collection("pins")
	d.userColl = d.db.Collection("users")
	d.reportChatColl = d.db.Collection("report_chat_settings")
	d.reportUserColl = d.db.Collection("report_user_settings")
	d.devsColl = d.db.Collection("devs")
	d.chatColl = d.db.Collection("chats")
	d.channelColl = d.db.Collection("channels")
	d.antifloodSettingsColl = d.db.Collection("antiflood_settings")
	d.connectionColl = d.db.Collection("connection")
	d.connectionSettingsColl = d.db.Collection("connection_settings")
	d.disableColl = d.db.Collection("disable")
	d.rulesColl = d.db.Collection("rules")
	d.warnSettingsColl = d.db.Collection("warns_settings")
	d.warnUsersColl = d.db.Collection("warns_users")
	d.greetingsColl = d.db.Collection("greetings")
	d.lockColl = d.db.Collection("locks")
	d.filterColl = d.db.Collection("filters")
	d.notesColl = d.db.Collection("notes")
	d.notesSettingsColl = d.db.Collection("notes_settings")
}

// setGlobalCollections sets global collection variables for backward compatibility
func (d *Database) setGlobalCollections() {
	adminSettingsColl = d.adminSettingsColl
	blacklistsColl = d.blacklistsColl
	pinColl = d.pinColl
	userColl = d.userColl
	reportChatColl = d.reportChatColl
	reportUserColl = d.reportUserColl
	devsColl = d.devsColl
	chatColl = d.chatColl
	channelColl = d.channelColl
	antifloodSettingsColl = d.antifloodSettingsColl
	connectionColl = d.connectionColl
	connectionSettingsColl = d.connectionSettingsColl
	disableColl = d.disableColl
	rulesColl = d.rulesColl
	warnSettingsColl = d.warnSettingsColl
	warnUsersColl = d.warnUsersColl
	greetingsColl = d.greetingsColl
	lockColl = d.lockColl
	filterColl = d.filterColl
	notesColl = d.notesColl
	notesSettingsColl = d.notesSettingsColl
}

// init initializes the database connection
func init() {
	// Initialize database connection
	_ = GetDatabase()
}

// updateOne updates a single document in the specified collection.
// If no document matches the filter, a new one is inserted (upsert).
func updateOne(collection *mongo.Collection, filter bson.M, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": data}, options.Update().SetUpsert(true))
	if err != nil {
		log.Errorf("[Database][updateOne]: %v", err)
	}
	return err
}

// findOne finds a single document in the specified collection matching the filter.
func findOne(collection *mongo.Collection, filter bson.M) *mongo.SingleResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return collection.FindOne(ctx, filter)
}

// findOneWithContext finds a single document with provided context
func findOneWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) *mongo.SingleResult {
	return collection.FindOne(ctx, filter)
}

// countDocs returns the number of documents in the collection matching the filter.
func countDocs(collection *mongo.Collection, filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Errorf("[Database][countDocs]: %v", err)
	}
	return count, err
}

// countDocsWithContext counts documents with provided context
func countDocsWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (int64, error) {
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Errorf("[Database][countDocsWithContext]: %v", err)
	}
	return count, err
}

// findAll returns a cursor for all documents in the collection matching the filter.
func findAll(collection *mongo.Collection, filter bson.M) *mongo.Cursor {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		log.Errorf("[Database][findAll]: %v", err)
		return nil
	}
	return cur
}

// findAllWithContext finds all documents with provided context
func findAllWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) (*mongo.Cursor, error) {
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		log.Errorf("[Database][findAllWithContext]: %v", err)
	}
	return cur, err
}

// deleteOne deletes a single document from the collection matching the filter.
func deleteOne(collection *mongo.Collection, filter bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Errorf("[Database][deleteOne]: %v", err)
	}
	return err
}

// deleteOneWithContext deletes a single document with provided context
func deleteOneWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) error {
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Errorf("[Database][deleteOneWithContext]: %v", err)
	}
	return err
}

// deleteMany deletes all documents from the collection matching the filter.
func deleteMany(collection *mongo.Collection, filter bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Errorf("[Database][deleteMany]: %v", err)
	}
	return err
}

// deleteManyWithContext deletes multiple documents with provided context
func deleteManyWithContext(ctx context.Context, collection *mongo.Collection, filter bson.M) error {
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Errorf("[Database][deleteManyWithContext]: %v", err)
	}
	return err
}
