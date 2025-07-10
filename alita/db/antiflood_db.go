package db

import (
	log "github.com/sirupsen/logrus"
)

const defaultFloodsettingsMode string = "mute"

type FloodSettings struct {
	ChatId                 int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	Limit                  int    `bson:"limit" json:"limit"`
	Mode                   string `bson:"mode,omitempty" json:"mode,omitempty"`
	DeleteAntifloodMessage bool   `bson:"del_msg" json:"del_msg"`
}

var floodSettingsHandler = &SettingsHandler[FloodSettings]{
	Collection: antifloodSettingsColl,
	Default: func(chatID int64) *FloodSettings {
		return &FloodSettings{ChatId: chatID, Limit: 0, Mode: defaultFloodsettingsMode}
	},
}

func GetFlood(chatID int64) *FloodSettings {
	return CheckFloodSetting(chatID)
}

func CheckFloodSetting(chatID int64) *FloodSettings {
	return floodSettingsHandler.CheckOrInit(chatID)
}

func SetFlood(chatID int64, limit int) {
	floodSrc := CheckFloodSetting(chatID)

	if floodSrc.Mode == "" {
		floodSrc = &FloodSettings{ChatId: chatID, Limit: limit, Mode: defaultFloodsettingsMode}
	} else {
		floodSrc.Limit = limit
	}

	err := updateOne(antifloodSettingsColl, map[string]interface{}{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFlood: %v - %d", err, chatID)
	}
}

func SetFloodMode(chatID int64, mode string) {
	floodSrc := CheckFloodSetting(chatID)
	floodSrc.Mode = mode

	err := updateOne(antifloodSettingsColl, map[string]interface{}{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFloodMode: %v - %d", err, chatID)
	}
}

func SetFloodMsgDel(chatID int64, val bool) {
	floodSrc := CheckFloodSetting(chatID)
	floodSrc.DeleteAntifloodMessage = val

	err := updateOne(antifloodSettingsColl, map[string]interface{}{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFloodMsgDel: %v - %d", err, chatID)
	}
}

func LoadAntifloodStats() (antiCount int64) {
	totalCount, err := countDocs(antifloodSettingsColl, nil)
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
	}
	noAntiCount, err := countDocs(antifloodSettingsColl, map[string]interface{}{"limit": 0})
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
	}
	antiCount = totalCount - noAntiCount
	return
}
