package db

import (
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
	log "github.com/sirupsen/logrus"
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
	var disbaledStruct []*DisableCommand

	cursor := findAll(disableColl, map[string]interface{}{})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &disbaledStruct)

	for _, disrc := range disbaledStruct {
		disLn := int64(len(disrc.Commands))
		disabledCmds += disLn
		if disLn > 0 {
			disableEnabledChats++
		}
	}

	return
}
