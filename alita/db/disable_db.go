package db

import (
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// DisableCMD Disable CMD in chat
func DisableCMD(chatID int64, cmd string) {
	// Create a new disable setting
	disableSetting := &DisableSettings{
		ChatId:   chatID,
		Command:  cmd,
		Disabled: true,
	}

	err := CreateRecord(disableSetting)
	if err != nil {
		log.Errorf("[Database][DisableCMD]: %v", err)
	}
}

// EnableCMD Enable CMD in chat
func EnableCMD(chatID int64, cmd string) {
	err := DB.Where("chat_id = ? AND command = ?", chatID, cmd).Delete(&DisableSettings{}).Error
	if err != nil {
		log.Errorf("[Database][EnableCMD]: %v", err)
	}
}

// GetChatDisabledCMDs Get disabled commands of chat
func GetChatDisabledCMDs(chatId int64) []string {
	var disableSettings []*DisableSettings
	err := GetRecords(&disableSettings, DisableSettings{ChatId: chatId, Disabled: true})
	if err != nil {
		log.Errorf("[Database] GetChatDisabledCMDs: %v - %d", err, chatId)
		return []string{}
	}

	commands := make([]string, len(disableSettings))
	for i, setting := range disableSettings {
		commands[i] = setting.Command
	}
	return commands
}

// IsCommandDisabled Check if command is disabled or not
func IsCommandDisabled(chatId int64, cmd string) bool {
	return string_handling.FindInStringSlice(GetChatDisabledCMDs(chatId), cmd)
}

// ToggleDel Toggle Command Deleting - Note: This functionality is not directly supported in the new model
func ToggleDel(chatId int64, pref bool) {
	log.Warnf("[Database] ToggleDel: Command deletion toggle not supported in new model for chat %d", chatId)
}

// ShouldDel Check if cmd del is enabled or not - Note: This functionality is not directly supported in the new model
func ShouldDel(chatId int64) bool {
	log.Warnf("[Database] ShouldDel: Command deletion check not supported in new model for chat %d", chatId)
	return false
}

func LoadDisableStats() (disabledCmds, disableEnabledChats int64) {
	// Count total disabled commands
	err := DB.Model(&DisableSettings{}).Where("disabled = ?", true).Count(&disabledCmds).Error
	if err != nil {
		log.Errorf("[Database] LoadDisableStats (commands): %v", err)
		return 0, 0
	}

	// Count distinct chats with disabled commands
	err = DB.Model(&DisableSettings{}).Where("disabled = ?", true).Distinct("chat_id").Count(&disableEnabledChats).Error
	if err != nil {
		log.Errorf("[Database] LoadDisableStats (chats): %v", err)
		return disabledCmds, 0
	}

	return
}
