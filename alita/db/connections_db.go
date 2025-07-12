package db

import (
	log "github.com/sirupsen/logrus"
)

type Connections struct {
	UserId    int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	ChatId    int64 `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	Connected bool  `bson:"connected" json:"connected" default:"false"`
}

type ConnectionSettings struct {
	ChatId       int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AllowConnect bool  `bson:"can_connect" json:"can_connect" default:"false"`
}

var connectionSettingsHandler = &SettingsHandler[ConnectionSettings]{
	Collection: connectionSettingsColl,
	Default: func(chatID int64) *ConnectionSettings {
		return &ConnectionSettings{ChatId: chatID, AllowConnect: false}
	},
}

func CheckChatConnectionSetting(chatID int64) *ConnectionSettings {
	return connectionSettingsHandler.CheckOrInit(chatID)
}

func ToggleAllowConnect(chatID int64, pref bool) {
	connectionSrc := CheckChatConnectionSetting(chatID)
	connectionSrc.AllowConnect = pref
	err := updateOne(connectionSettingsColl, map[string]interface{}{"_id": chatID}, connectionSrc)
	if err != nil {
		log.Errorf("[Database] ToggleAllowConnect: %d - %v", chatID, err)
	}
}

func GetChatConnectionSetting(chatID int64) *ConnectionSettings {
	return CheckChatConnectionSetting(chatID)
}

// The rest of the user connection logic remains unchanged
func getUserConnectionSetting(userID int64) (connectionSrc *Connections) {
	defaultConnectionSrc := &Connections{UserId: userID, Connected: false}
	errF := findOne(connectionColl, map[string]interface{}{"_id": userID}).Decode(&connectionSrc)
	if errF != nil {
		connectionSrc = defaultConnectionSrc
		log.Errorf("[Database] GetUserSetting: %d - %v", userID, errF)
	}
	return connectionSrc
}

func Connection(UserID int64) *Connections {
	return getUserConnectionSetting(UserID)
}

func ConnectId(UserID, chatID int64) {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = true
	connectionUpdate.ChatId = chatID
	err := updateOne(connectionColl, map[string]interface{}{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] ConnectId: %v - %d", err, chatID)
	}
}

func DisconnectId(UserID int64) {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = false
	err := updateOne(connectionColl, map[string]interface{}{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] DisconnectId: %v - %d", err, UserID)
	}
}

func ReconnectId(UserID int64) int64 {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = true
	err := updateOne(connectionColl, map[string]interface{}{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] ReconnectId: %v - %d", err, UserID)
		return 0
	}
	return connectionUpdate.ChatId
}

func LoadConnectionStats() (connectedUsers, connectedChats int64) {
	connectedChats, err := countDocs(
		connectionSettingsColl,
		map[string]interface{}{"can_connect": true},
	)
	if err != nil {
		log.Error(err)
		return
	}

	connectedUsers, err = countDocs(
		connectionColl,
		map[string]interface{}{"connected": true},
	)
	if err != nil {
		log.Error(err)
		return
	}

	return
}
