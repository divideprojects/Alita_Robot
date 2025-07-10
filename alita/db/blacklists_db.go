package db

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
)

type BlacklistSettings struct {
	ChatId   int64    `bson:"_id,omitempty" json:"_id,omitempty"`
	Action   string   `bson:"action,omitempty" json:"action,omitempty"`
	Triggers []string `bson:"triggers,omitempty" json:"triggers,omitempty"`
	Reason   string   `bson:"reason,omitempty" json:"reason,omitempty"`
}

var blacklistSettingsHandler = &SettingsHandler[BlacklistSettings]{
	Collection: blacklistsColl,
	Default: func(chatID int64) *BlacklistSettings {
		return &BlacklistSettings{
			ChatId:   chatID,
			Action:   "none",
			Triggers: make([]string, 0),
			Reason:   "Automated Blacklisted word %s",
		}
	},
}

func CheckBlacklistSetting(chatID int64) *BlacklistSettings {
	return blacklistSettingsHandler.CheckOrInit(chatID)
}

func AddBlacklist(chatId int64, trigger string) {
	blSrc := CheckBlacklistSetting(chatId)
	blSrc.Triggers = append(blSrc.Triggers, strings.ToLower(trigger))
	err := updateOne(blacklistsColl, map[string]interface{}{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] AddBlacklist: %v - %d", err, chatId)
	}
}

func RemoveBlacklist(chatId int64, trigger string) {
	blSrc := CheckBlacklistSetting(chatId)
	blSrc.Triggers = removeStrfromStr(blSrc.Triggers, strings.ToLower(trigger))
	err := updateOne(blacklistsColl, map[string]interface{}{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

func RemoveAllBlacklist(chatId int64) {
	blSrc := CheckBlacklistSetting(chatId)
	blSrc.Triggers = make([]string, 0)
	err := updateOne(blacklistsColl, map[string]interface{}{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

func SetBlacklistAction(chatId int64, action string) {
	blSrc := CheckBlacklistSetting(chatId)
	blSrc.Action = strings.ToLower(action)
	err := updateOne(blacklistsColl, map[string]interface{}{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] ChangeBlacklistAction: %v - %d", err, chatId)
	}
}

func GetBlacklistSettings(chatId int64) *BlacklistSettings {
	return CheckBlacklistSetting(chatId)
}

func LoadBlacklistsStats() (blacklistTriggers, blacklistChats int64) {
	var blacklistStruct []*BlacklistSettings

	cursor := findAll(blacklistsColl, map[string]interface{}{})
	defer func(cursor interface{ Close(context.Context) error }, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			log.Error(err)
		}
	}(cursor, bgCtx)

	for cursor.Next(bgCtx) {
		var blacklistSetting BlacklistSettings
		if err := cursor.Decode(&blacklistSetting); err != nil {
			log.Error("Failed to decode blacklist setting:", err)
			continue
		}
		blacklistStruct = append(blacklistStruct, &blacklistSetting)
	}

	for _, i := range blacklistStruct {
		lenBl := len(i.Triggers)
		blacklistTriggers += int64(lenBl)
		if lenBl > 0 {
			blacklistChats++
		}
	}

	return
}
