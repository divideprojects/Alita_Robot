package db

import (
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"

	"go.mongodb.org/mongo-driver/bson"
)

type ChatReportSettings struct {
	ChatId      int64   `bson:"_id,omitempty" json:"_id,omitempty"`
	Status      bool    `bson:"status,omitempty" json:"status,omitempty"`
	BlockedList []int64 `bson:"blocked_list,omitempty" json:"blocked_list,omitempty"`
}

type UserReportSettings struct {
	UserId int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	Status bool  `bson:"status,omitempty" json:"status,omitempty"`
}

var chatReportSettingsHandler = &SettingsHandler[ChatReportSettings]{
	Collection: reportChatColl,
	Default: func(chatID int64) *ChatReportSettings {
		return &ChatReportSettings{
			ChatId:      chatID,
			Status:      true,
			BlockedList: make([]int64, 0),
		}
	},
}

var userReportSettingsHandler = &SettingsHandler[UserReportSettings]{
	Collection: reportUserColl,
	Default: func(userID int64) *UserReportSettings {
		return &UserReportSettings{
			UserId: userID,
			Status: true,
		}
	},
}

func GetChatReportSettings(chatID int64) *ChatReportSettings {
	return chatReportSettingsHandler.CheckOrInit(chatID)
}

func SetChatReportStatus(chatID int64, pref bool) {
	reportsUpdate := GetChatReportSettings(chatID)
	reportsUpdate.Status = pref
	err := updateOne(reportChatColl, bson.M{"_id": chatID}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

func BlockReportUser(chatId, userId int64) {
	reportsUpdate := GetChatReportSettings(chatId)

	if string_handling.FindInInt64Slice(reportsUpdate.BlockedList, userId) {
		return
	}

	reportsUpdate.BlockedList = append(reportsUpdate.BlockedList, userId)
	err := updateOne(reportChatColl, bson.M{"_id": chatId}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

func UnblockReportUser(chatId, userId int64) {
	reportsUpdate := GetChatReportSettings(chatId)

	if !string_handling.FindInInt64Slice(reportsUpdate.BlockedList, userId) {
		return
	}

	reportsUpdate.BlockedList = string_handling.RemoveFromInt64Slice(reportsUpdate.BlockedList, userId)
	err := updateOne(reportChatColl, bson.M{"_id": chatId}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

/* user settings below */

func GetUserReportSettings(userId int64) *UserReportSettings {
	return userReportSettingsHandler.CheckOrInit(userId)
}

func SetUserReportSettings(chatID int64, pref bool) {
	reportsUpdate := &ChatReportSettings{
		ChatId: chatID,
		Status: pref,
	}
	err := updateOne(reportUserColl, bson.M{"_id": chatID}, reportsUpdate)
	if err != nil {
		log.Error(chatID)
	}
}

func LoadReportStats() (uRCount, gRCount int64) {
	uRCount, acErr := countDocs(
		reportUserColl,
		bson.M{"status": true},
	)
	if acErr != nil {
		log.Errorf("[Database] loadStats: %v", acErr)
	}
	gRCount, clErr := countDocs(
		reportChatColl,
		bson.M{"status": true},
	)
	if clErr != nil {
		log.Errorf("[Database] loadStats: %v", clErr)
	}
	return
}
