package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// check chat Rules Settings, used to get data before performing any operation
func checkRulesSetting(chatID int64) (rulesrc *RulesSettings) {
	rulesrc = &RulesSettings{}
	err := GetRecord(rulesrc, RulesSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		rulesrc = &RulesSettings{ChatId: chatID, Rules: ""}
		err := CreateRecord(rulesrc)
		if err != nil {
			log.Errorf("[Database] checkRulesSetting: %v - %d", err, chatID)
		}
	} else if err != nil {
		// Return default on error
		rulesrc = &RulesSettings{ChatId: chatID, Rules: ""}
		log.Errorf("[Database] checkRulesSetting: %v - %d", err, chatID)
	}
	return rulesrc
}

func GetChatRulesInfo(chatId int64) *RulesSettings {
	return checkRulesSetting(chatId)
}

func SetChatRules(chatId int64, rules string) {
	err := UpdateRecord(&RulesSettings{}, RulesSettings{ChatId: chatId}, RulesSettings{Rules: rules})
	if err != nil {
		log.Errorf("[Database] SetChatRules: %v - %d", err, chatId)
	}
}

func SetChatRulesButton(chatId int64, rulesButton string) {
	// Note: The new RulesSettings model doesn't have a RulesBtn field
	// This functionality is not directly supported
	log.Warnf("[Database] SetChatRulesButton: Rules button functionality not supported in new model for chat %d", chatId)
}

func SetPrivateRules(chatId int64, pref bool) {
	// Note: The new RulesSettings model doesn't have a Private field
	// This functionality is not directly supported
	log.Warnf("[Database] SetPrivateRules: Private rules functionality not supported in new model for chat %d", chatId)
}

func LoadRulesStats() (setRules, pvtRules int64) {
	// Count chats with rules set (non-empty rules)
	err := DB.Model(&RulesSettings{}).Where("rules != ?", "").Count(&setRules).Error
	if err != nil {
		log.Errorf("[Database] LoadRulesStats (set rules): %v", err)
	}

	// Note: Private rules functionality is not supported in the new model
	pvtRules = 0
	log.Warnf("[Database] LoadRulesStats: Private rules count not supported in new model")

	return
}
