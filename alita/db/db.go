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

// dbInstance func
func init() {
	mongoClient, err := mongo.NewClient(
		options.Client().ApplyURI(config.DatabaseURI),
	)
	if err != nil {
		log.Errorf("[Database][Client]: %v", err)
	}

	ctx, cancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer cancel()

	err = mongoClient.Connect(ctx)
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
