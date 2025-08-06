package db

import (
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// DisableCMD disables a command in a specific chat.
// Creates a new disable setting record with disabled status set to true.
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

// EnableCMD enables a command in a specific chat.
// Removes the disable setting record for the command.
func EnableCMD(chatID int64, cmd string) {
	err := DB.Where("chat_id = ? AND command = ?", chatID, cmd).Delete(&DisableSettings{}).Error
	if err != nil {
		log.Errorf("[Database][EnableCMD]: %v", err)
	}
}

// GetChatDisabledCMDs retrieves all disabled commands for a chat.
// Returns an empty slice if no disabled commands are found or on error.
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

// IsCommandDisabled checks if a specific command is disabled in a chat.
// Returns true if the command is in the chat's disabled commands list.
func IsCommandDisabled(chatId int64, cmd string) bool {
	return string_handling.FindInStringSlice(GetChatDisabledCMDs(chatId), cmd)
}

// ToggleDel toggles the automatic deletion of disabled commands in a chat.
// Updates the DeleteCommands setting for the chat.
func ToggleDel(chatId int64, pref bool) {
	err := UpdateRecord(&DisableChatSettings{}, DisableChatSettings{ChatId: chatId}, DisableChatSettings{DeleteCommands: pref})
	if err != nil {
		log.Errorf("[Database] ToggleDel: %v", err)
	}
}

// ShouldDel checks if automatic command deletion is enabled for a chat.
// Returns false if the setting is not found or on error.
func ShouldDel(chatId int64) bool {
	var settings DisableChatSettings
	err := GetRecord(&settings, DisableChatSettings{ChatId: chatId})
	if err != nil {
		log.Errorf("[Database] ShouldDel: %v", err)
		return false
	}
	return settings.DeleteCommands
}

// LoadDisableStats returns statistics about disabled commands.
// Returns the total number of disabled commands and distinct chats using command disabling.
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
