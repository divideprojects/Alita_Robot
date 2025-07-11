package db

import (
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
)

type Rules struct {
	ChatId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	Rules    string `bson:"rules" json:"rules" default:""`
	Private  bool   `bson:"privrules" json:"privrules"`
	RulesBtn string `bson:"rules_button,omitempty" json:"rules_button,omitempty"`
}

var rulesSettingsHandler = &SettingsHandler[Rules]{
	Collection: rulesColl,
	Default: func(chatID int64) *Rules {
		return &Rules{ChatId: chatID, Rules: "", Private: false}
	},
}

// checkRulesSetting uses the generic handler to get or initialize rules settings
func checkRulesSetting(chatID int64) *Rules {
	return rulesSettingsHandler.CheckOrInit(chatID)
}

func GetChatRulesInfo(chatId int64) *Rules {
	return checkRulesSetting(chatId)
}

func SetChatRules(chatId int64, rules string) {
	rulesUpdate := checkRulesSetting(chatId)
	rulesUpdate.Rules = rules
	err := updateOne(rulesColl, bson.M{"_id": chatId}, rulesUpdate)
	if err != nil {
		log.Errorf("[Database] SetChatRules: %v - %d", err, chatId)
	}
}

func SetChatRulesButton(chatId int64, rulesButton string) {
	rulesUpdate := checkRulesSetting(chatId)
	rulesUpdate.RulesBtn = rulesButton
	err := updateOne(rulesColl, bson.M{"_id": chatId}, rulesUpdate)
	if err != nil {
		log.Errorf("[Database] SetChatRulesButton: %v - %d", err, chatId)
	}
}

func SetPrivateRules(chatId int64, pref bool) {
	rulesUpdate := checkRulesSetting(chatId)
	rulesUpdate.Private = pref
	err := updateOne(rulesColl, bson.M{"_id": chatId}, rulesUpdate)
	if err != nil {
		log.Errorf("[Database] SetPrivateRules: %v - %d", err, chatId)
	}
}

func LoadRulesStats() (setRules, pvtRules int64) {
	setRules, clErr := countDocs(
		rulesColl,
		bson.M{
			"rules": bson.M{
				"$ne": "",
			},
		},
	)
	if clErr != nil {
		log.Errorf("[Database] LoadRulesStats: %v", clErr)
	}
	pvtRules, alErr := countDocs(
		rulesColl,
		bson.M{
			"privrules": true,
		},
	)
	if alErr != nil {
		log.Errorf("[Database] LoadRulesStats: %v", clErr)
	}
	return
}
