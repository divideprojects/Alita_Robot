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

type Button struct {
	Name     string `bson:"name,omitempty" json:"name,omitempty"`
	Url      string `bson:"url,omitempty" json:"url,omitempty"`
	SameLine bool   `bson:"btn_sameline" json:"btn_sameline" default:"false"`
}

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
)

// Generic GetOrCreateByID for MongoDB collections
func GetOrCreateByID[T any](
	coll *mongo.Collection,
	filter interface{},
	defaultFactory func() *T,
) *T {
	var result *T
	err := coll.FindOne(bgCtx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		result = defaultFactory()
		_, err := coll.UpdateOne(bgCtx, filter, bson.M{"$set": result}, options.Update().SetUpsert(true))
		if err != nil {
			log.Errorf("[Database][GetOrCreateByID]: %v", err)
		}
	} else if err != nil {
		log.Errorf("[Database][GetOrCreateByID]: %v", err)
		result = defaultFactory()
	}
	return result
}

// UpdateField updates a specific field in a document
func UpdateField[T any](
	coll *mongo.Collection,
	filter interface{},
	fieldName string,
	value T,
) error {
	_, err := coll.UpdateOne(bgCtx, filter, bson.M{"$set": bson.M{fieldName: value}}, options.Update().SetUpsert(true))
	if err != nil {
		log.Errorf("[Database][UpdateField]: %v", err)
	}
	return err
}

// ToggleField toggles a boolean field in a document
func ToggleField(
	coll *mongo.Collection,
	filter interface{},
	fieldName string,
) (bool, error) {
	// First get the current value
	var result bson.M
	err := coll.FindOne(bgCtx, filter).Decode(&result)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("[Database][ToggleField]: %v", err)
		return false, err
	}

	// Get current value or default to false
	currentValue := false
	if val, exists := result[fieldName]; exists {
		if boolVal, ok := val.(bool); ok {
			currentValue = boolVal
		}
	}

	// Toggle the value
	newValue := !currentValue
	err = UpdateField(coll, filter, fieldName, newValue)
	return newValue, err
}

// GetList retrieves a list of documents matching the filter
func GetList[T any](
	coll *mongo.Collection,
	filter interface{},
	opts ...*options.FindOptions,
) ([]*T, error) {
	cursor, err := coll.Find(bgCtx, filter, opts...)
	if err != nil {
		log.Errorf("[Database][GetList]: %v", err)
		return nil, err
	}
	defer cursor.Close(bgCtx)

	var results []*T
	err = cursor.All(bgCtx, &results)
	if err != nil {
		log.Errorf("[Database][GetList]: %v", err)
		return nil, err
	}

	return results, nil
}

// CheckExists checks if a document exists matching the filter
func CheckExists(
	coll *mongo.Collection,
	filter interface{},
) (bool, error) {
	count, err := coll.CountDocuments(bgCtx, filter)
	if err != nil {
		log.Errorf("[Database][CheckExists]: %v", err)
		return false, err
	}
	return count > 0, nil
}

// GetByID retrieves a document by ID
func GetByID[T any](
	coll *mongo.Collection,
	id interface{},
) (*T, error) {
	var result *T
	err := coll.FindOne(bgCtx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Errorf("[Database][GetByID]: %v", err)
		return nil, err
	}
	return result, nil
}

// DeleteByID deletes a document by ID
func DeleteByID(
	coll *mongo.Collection,
	id interface{},
) error {
	_, err := coll.DeleteOne(bgCtx, bson.M{"_id": id})
	if err != nil {
		log.Errorf("[Database][DeleteByID]: %v", err)
	}
	return err
}

// UpsertDocument inserts or updates a document
func UpsertDocument[T any](
	coll *mongo.Collection,
	filter interface{},
	document *T,
) error {
	_, err := coll.ReplaceOne(bgCtx, filter, document, options.Replace().SetUpsert(true))
	if err != nil {
		log.Errorf("[Database][UpsertDocument]: %v", err)
	}
	return err
}

// GetCount returns the count of documents matching the filter
func GetCount(
	coll *mongo.Collection,
	filter interface{},
) (int64, error) {
	count, err := coll.CountDocuments(bgCtx, filter)
	if err != nil {
		log.Errorf("[Database][GetCount]: %v", err)
		return 0, err
	}
	return count, nil
}

// dbInstance func
func init() {
	ctx, cancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseURI))
	if err != nil {
		log.Errorf("[Database][Connect]: %v", err)
	}

	// Open Connections to Collections
	log.Info("Opening Database Collections...")
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
}

// updateOne func to update one document
func updateOne(collecion *mongo.Collection, filter bson.M, data interface{}) (err error) {
	_, err = collecion.UpdateOne(tdCtx, filter, bson.M{"$set": data}, options.Update().SetUpsert(true))
	if err != nil {
		log.Errorf("[Database][updateOne]: %v", err)
	}
	return
}

// findOne func to find one document
func findOne(collecion *mongo.Collection, filter bson.M) (res *mongo.SingleResult) {
	res = collecion.FindOne(tdCtx, filter)
	return
}

// countDocs func to count documents
func countDocs(collecion *mongo.Collection, filter bson.M) (count int64, err error) {
	count, err = collecion.CountDocuments(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][countDocs]: %v", err)
	}
	return
}

// findAll func to find all documents
func findAll(collecion *mongo.Collection, filter bson.M) (cur *mongo.Cursor) {
	cur, err := collecion.Find(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][findAll]: %v", err)
	}
	return
}

// deleteOne func to delete one document
func deleteOne(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteOne(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteOne]: %v", err)
	}
	return
}

// deleteMany func to delete many documents
func deleteMany(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteMany(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteMany]: %v", err)
	}
	return
}
