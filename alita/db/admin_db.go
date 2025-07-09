package db

import (
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
AdminSettings represents admin-related settings for a chat, such as whether anonymous admin mode is enabled.

Fields:
  - ChatId: Unique identifier for the chat.
  - AnonAdmin: Indicates if anonymous admin mode is enabled for the chat.
*/
type AdminSettings struct {
	ChatId    int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AnonAdmin bool  `bson:"anon_admin" json:"anon_admin"`
}

// GetAdminSettings retrieves the admin settings for a given chat ID.
// If no settings exist, it initializes them with default values.
func GetAdminSettings(chatID int64) *AdminSettings {
	return checkAdminSetting(chatID)
}

// checkAdminSetting fetches admin settings for a chat from the database.
// If no document exists, it creates one with default values.
// Returns a pointer to the AdminSettings struct.
func checkAdminSetting(chatID int64) (adminSrc *AdminSettings) {
	defaultAdminSrc := &AdminSettings{ChatId: chatID, AnonAdmin: false}

	err := findOne(adminSettingsColl, bson.M{"_id": chatID}).Decode(&adminSrc)
	if err == mongo.ErrNoDocuments {
		adminSrc = defaultAdminSrc
		err := updateOne(adminSettingsColl, bson.M{"_id": chatID}, defaultAdminSrc)
		if err != nil {
			log.Errorf("[Database][checkAdminSetting]: %v ", err)
		}
	} else if err != nil {
		adminSrc = defaultAdminSrc
		log.Errorf("[Database][checkAdminSetting]: %v ", err)
	}
	return adminSrc
}

// SetAnonAdminMode updates the anonymous admin mode for a specific chat.
// It ensures the admin settings exist before updating the AnonAdmin field.
func SetAnonAdminMode(chatID int64, val bool) {
	adminSrc := checkAdminSetting(chatID)
	adminSrc.AnonAdmin = val

	err := updateOne(adminSettingsColl, bson.M{"_id": chatID}, adminSrc)
	if err != nil {
		log.Errorf("[Database] SetAnonAdminMode: %v - %d", err, chatID)
	}
}
