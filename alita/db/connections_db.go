package db

import (
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Connections represents a user's connection state to a chat.

Fields:
  - UserId: Unique identifier for the user.
  - ChatId: ID of the chat the user is connected to.
  - Connected: Whether the user is currently connected to a chat.
*/
type Connections struct {
	UserId    int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	ChatId    int64 `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	Connected bool  `bson:"connected" json:"connected" default:"false"`
}

// ConnectionSettings represents connection permissions for a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - AllowConnect: Whether users are allowed to connect to this chat.
type ConnectionSettings struct {
	ChatId       int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AllowConnect bool  `bson:"can_connect" json:"can_connect" default:"false"`
}

// ToggleAllowConnect sets whether users are allowed to connect to the specified chat.
func ToggleAllowConnect(chatID int64, pref bool) {
	connectionSrc := GetChatConnectionSetting(chatID)
	connectionSrc.AllowConnect = pref
	err := updateOne(connectionSettingsColl, bson.M{"_id": chatID}, connectionSrc)
	if err != nil {
		log.Errorf("[Database] ToggleAllowConnect: %d - %v", chatID, err)
	}
}

// GetChatConnectionSetting retrieves the connection settings for a chat.
// If no settings exist, it initializes them with default values.
func GetChatConnectionSetting(chatID int64) (connectionSrc *ConnectionSettings) {
	defaultConnectionSrc := &ConnectionSettings{ChatId: chatID, AllowConnect: false}
	errF := findOne(connectionSettingsColl, bson.M{"_id": chatID}).Decode(&connectionSrc)
	if errF == mongo.ErrNoDocuments {
		connectionSrc = defaultConnectionSrc
		err := updateOne(connectionSettingsColl, bson.M{"_id": chatID}, connectionSrc)
		if err != nil {
			log.Errorf("[Database] GetChatConnectionSetting: %d - %v", chatID, err)
		}
	} else if errF != nil {
		connectionSrc = defaultConnectionSrc
		log.Errorf("[Database] GetChatSetting: %d - %v", chatID, errF)
	}
	return connectionSrc
}

func getUserConnectionSetting(userID int64) (connectionSrc *Connections) {
	defaultConnectionSrc := &Connections{UserId: userID, Connected: false}
	errF := findOne(connectionColl, bson.M{"_id": userID}).Decode(&connectionSrc)
	if errF == mongo.ErrNoDocuments {
		connectionSrc = defaultConnectionSrc
		err := updateOne(connectionColl, bson.M{"_id": userID}, connectionSrc)
		if err != nil {
			log.Errorf("[Database] GetChatConnectionSetting: %d - %v", userID, err)
		}
	} else if errF != nil {
		connectionSrc = defaultConnectionSrc
		log.Errorf("[Database] GetUserSetting: %d - %v", userID, errF)
	}

	return connectionSrc
}

// Connection retrieves the connection state for a user.
func Connection(UserID int64) *Connections {
	return getUserConnectionSetting(UserID)
}

// ConnectId marks a user as connected to a specific chat.
func ConnectId(UserID, chatID int64) {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = true
	connectionUpdate.ChatId = chatID
	err := updateOne(connectionColl, bson.M{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] ConnectId: %v - %d", err, chatID)
	}
}

// DisconnectId marks a user as disconnected from any chat.
func DisconnectId(UserID int64) {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = false
	err := updateOne(connectionColl, bson.M{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] DisconnectId: %v - %d", err, UserID)
	}
}

// ReconnectId marks a user as connected and returns the chat ID they are connected to.
// Returns 0 if the update fails.
func ReconnectId(UserID int64) int64 {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = true
	err := updateOne(connectionColl, bson.M{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] ReconnectId: %v - %d", err, UserID)
		return 0
	}
	return connectionUpdate.ChatId
}

// LoadConnectionStats returns the number of users currently connected and the number of chats allowing connections.
func LoadConnectionStats() (connectedUsers, connectedChats int64) {
	connectedChats, err := countDocs(
		connectionSettingsColl,
		bson.M{"can_connect": true},
	)
	if err != nil {
		log.Error(err)
		return
	}

	connectedUsers, err = countDocs(
		connectionColl,
		bson.M{"connected": true},
	)
	if err != nil {
		log.Error(err)
		return
	}

	return
}
