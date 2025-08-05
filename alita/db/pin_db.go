package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetPinData(chatID int64) (pinrc *PinSettings) {
	pinrc = &PinSettings{}
	err := GetRecord(pinrc, PinSettings{ChatId: chatID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings
		pinrc = &PinSettings{ChatId: chatID, MsgId: 0}
		err := CreateRecord(pinrc)
		if err != nil {
			log.Errorf("[Database] GetPinData: %v - %d", err, chatID)
		}
	} else if err != nil {
		// Return default on error
		pinrc = &PinSettings{ChatId: chatID, MsgId: 0}
		log.Errorf("[Database] GetPinData: %v - %d", err, chatID)
	}
	log.Infof("[Database] GetPinData: %d", chatID)
	return
}

func SetCleanLinked(chatID int64, pref bool) {
	err := UpdateRecord(&PinSettings{}, PinSettings{ChatId: chatID}, PinSettings{CleanLinked: pref})
	if err != nil {
		log.Errorf("[Database] SetCleanLinked: %v", err)
	}
}

func SetAntiChannelPin(chatID int64, pref bool) {
	err := UpdateRecord(&PinSettings{}, PinSettings{ChatId: chatID}, PinSettings{AntiChannelPin: pref})
	if err != nil {
		log.Errorf("[Database] SetAntiChannelPin: %v", err)
	}
}

func LoadPinStats() (acCount, clCount int64) {
	// Note: The new PinSettings model doesn't support the old pin statistics
	// Return 0 for both counts as these features are not supported
	log.Warnf("[Database] LoadPinStats: Pin statistics not supported in new model")
	return 0, 0
}
