package db

import (
	"context"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type DisableCommand struct {
	ChatID       int64    `bson:"_id,omitempty" json:"_id,omitempty"`
	Commands     []string `bson:"commands" json:"commands" default:"none"`
	ShouldDelete bool     `bson:"should_delete" json:"should_delete" default:"false"`
}

var disableSettingsHandler = &SettingsHandler[DisableCommand]{
	Collection: disableColl,
	Default: func(chatID int64) *DisableCommand {
		return &DisableCommand{ChatID: chatID, Commands: make([]string, 0), ShouldDelete: false}
	},
}

func CheckDisableSettings(chatID int64) *DisableCommand {
	return disableSettingsHandler.CheckOrInit(chatID)
}

func DisableCMD(chatID int64, cmd string) {
	disableCmd := CheckDisableSettings(chatID)
	disableCmd.Commands = append(disableCmd.Commands, cmd)
	err := updateOne(disableColl, map[string]interface{}{"_id": chatID}, disableCmd)
	if err != nil {
		log.Errorf("[Database][DisableCMD]: %v", err)
	}
}

func EnableCMD(chatID int64, cmd string) {
	disableCmd := CheckDisableSettings(chatID)
	disableCmd.Commands = removeStrfromStr(disableCmd.Commands, cmd)
	err := updateOne(disableColl, map[string]interface{}{"_id": chatID}, disableCmd)
	if err != nil {
		log.Errorf("[Database][EnableCMD]: %v", err)
	}
}

func GetChatDisabledCMDs(chatId int64) []string {
	return CheckDisableSettings(chatId).Commands
}

func IsCommandDisabled(chatId int64, cmd string) bool {
	return string_handling.FindInStringSlice(GetChatDisabledCMDs(chatId), cmd)
}

func ToggleDel(chatId int64, pref bool) {
	disableCmd := CheckDisableSettings(chatId)
	disableCmd.ShouldDelete = pref
	err := updateOne(disableColl, map[string]interface{}{"_id": chatId}, disableCmd)
	if err != nil {
		log.Error(err)
	}
}

func ShouldDel(chatId int64) bool {
	disableCmd := CheckDisableSettings(chatId)
	return disableCmd.ShouldDelete
}

func removeStrfromStr(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func LoadDisableStats() (disabledCmds, disableEnabledChats int64) {
	// Count chats with non-empty commands array (active disables)
	_, disableEnabledChats = CountByChat(disableColl, bson.M{"commands": bson.M{"$exists": true, "$ne": []string{}}}, "_id")
	
	// For disabled commands count, we need manual aggregation since we're counting array elements
	cursor := findAll(disableColl, bson.M{"commands": bson.M{"$exists": true, "$ne": []string{}}})
	defer func(cursor interface{ Close(context.Context) error }, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			log.Error(err)
		}
	}(cursor, bgCtx)

	for cursor.Next(bgCtx) {
		var disableCommand DisableCommand
		if err := cursor.Decode(&disableCommand); err != nil {
			log.Error("Failed to decode disable command:", err)
			continue
		}
		disabledCmds += int64(len(disableCommand.Commands))
	}

	return
}
