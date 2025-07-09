package db

import (
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Pins holds pin-related settings for a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - AntiChannelPin: Whether anti-channel pin is enabled.
//   - CleanLinked: Whether linked messages should be cleaned.
type Pins struct {
	ChatId         int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AntiChannelPin bool  `bson:"antichannelpin" json:"antichannelpin" default:"false"`
	CleanLinked    bool  `bson:"cleanlinked" json:"cleanlinked" default:"false"`
}

// GetPinData retrieves the pin settings for a given chat ID.
// If no settings exist, it initializes them with default values.
func GetPinData(chatID int64) (pinrc *Pins) {
	defaultPinrc := &Pins{ChatId: chatID, AntiChannelPin: false, CleanLinked: false}
	err := findOne(pinColl, bson.M{"_id": chatID}).Decode(&pinrc)
	if err == mongo.ErrNoDocuments {
		pinrc = defaultPinrc
		err = updateOne(pinColl, bson.M{"_id": chatID}, pinrc)
		if err != nil {
			log.Errorf("[Database] GetPinData: %v - %d", err, chatID)
		}
	} else if err != nil {
		log.Errorf("[Database] GetPinData: %v - %d", err, chatID)
		pinrc = defaultPinrc
	}
	log.Infof("[Database] GetPinData: %d", chatID)
	return
}

// SetCleanLinked sets the CleanLinked flag for a chat and disables AntiChannelPin if enabled.
func SetCleanLinked(chatID int64, pref bool) {
	antichannelpin := false
	if pref {
		antichannelpin = false
	}
	pinsUpdate := &Pins{ChatId: chatID, AntiChannelPin: antichannelpin, CleanLinked: pref}
	err := updateOne(pinColl, bson.M{"_id": chatID}, pinsUpdate)
	if err != nil {
		log.Errorf("[Database] SetCleanLinked: %v - %d", err, chatID)
	}
}

// SetAntiChannelPin sets the AntiChannelPin flag for a chat and disables CleanLinked if enabled.
func SetAntiChannelPin(chatID int64, pref bool) {
	cleanlinked := false
	if pref {
		cleanlinked = false
	}
	pinsUpdate := &Pins{ChatId: chatID, AntiChannelPin: pref, CleanLinked: cleanlinked}
	err := updateOne(pinColl, bson.M{"_id": chatID}, pinsUpdate)
	if err != nil {
		log.Errorf("[Database] SetAntiChannelPin: %v - %d", err, chatID)
		return
	}
	log.Infof("[Database] SetAntiChannelPin: %v - %d", pref, chatID)
}

// LoadPinStats returns the number of chats with AntiChannelPin and CleanLinked enabled, respectively.
func LoadPinStats() (acCount, clCount int64) {
	acCount, err := countDocs(
		pinColl,
		bson.M{
			"cleanlinked":    false,
			"antichannelpin": true,
		},
	)
	if err != nil {
		log.Errorf("[Database] loadPinStats: %v", err)
	}
	clCount, err = countDocs(
		pinColl,
		bson.M{
			"cleanlinked":    true,
			"antichannelpin": false,
		},
	)
	if err != nil {
		log.Errorf("[Database] loadPinStats: %v", err)
	}
	return
}
