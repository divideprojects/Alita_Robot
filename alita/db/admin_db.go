package db

import (
	log "github.com/sirupsen/logrus"
)

type AdminSettings struct {
	ChatId    int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AnonAdmin bool  `bson:"anon_admin" json:"anon_admin"`
}

var adminSettingsHandler = &SettingsHandler[AdminSettings]{
	Collection: adminSettingsColl,
	Default: func(chatID int64) *AdminSettings {
		return &AdminSettings{ChatId: chatID, AnonAdmin: false}
	},
}

// GetAdminSettings retrieves the admin settings for a given chat ID.
func GetAdminSettings(chatID int64) *AdminSettings {
	return CheckAdminSetting(chatID)
}

// CheckAdminSetting uses the generic handler to get or initialize admin settings.
func CheckAdminSetting(chatID int64) *AdminSettings {
	return adminSettingsHandler.CheckOrInit(chatID)
}

// SetAnonAdminMode updates the anonymous admin mode for a specific chat.
func SetAnonAdminMode(chatID int64, val bool) {
	adminSrc := CheckAdminSetting(chatID)
	adminSrc.AnonAdmin = val

	err := updateOne(adminSettingsColl, map[string]interface{}{"_id": chatID}, adminSrc)
	if err != nil {
		log.Errorf("[Database] SetAnonAdminMode: %v - %d", err, chatID)
	}
}
