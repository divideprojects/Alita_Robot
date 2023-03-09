package db

import (
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// default strings when no settings are set
const (
	DefaultWelcome = "Hey {first}, how are you?"
	DefaultGoodbye = "Sad to see you leaving {first}"
)

type GreetingSettings struct {
	ChatID             int64            `bson:"_id,omitempty" json:"_id,omitempty"`
	ShouldCleanService bool             `bson:"clean_service_settings" json:"clean_service_settings" default:"false"`
	WelcomeSettings    *WelcomeSettings `bson:"welcome_settings" json:"welcome_settings" default:"false"`
	GoodbyeSettings    *GoodbyeSettings `bson:"goodbye_settings" json:"goodbye_settings" default:"false"`
	ShouldAutoApprove  bool             `bson:"auto_approve" json:"auto_approve" default:"false"`
}

type GoodbyeSettings struct {
	CleanGoodbye  bool     `bson:"clean_old" json:"clean_old" default:"false"`
	LastMsgId     int64    `bson:"last_msg_id,omitempty" json:"last_msg_id,omitempty"`
	ShouldGoodbye bool     `bson:"enabled" json:"enabled" default:"true"`
	GoodbyeText   string   `bson:"text,omitempty" json:"text,omitempty"`
	FileID        string   `bson:"file_id,omitempty" json:"file_id,omitempty"`
	GoodbyeType   int      `bson:"type,omitempty" json:"type,omitempty"`
	Button        []Button `bson:"btns,omitempty" json:"btns,omitempty"`
}

type WelcomeSettings struct {
	CleanWelcome  bool     `bson:"clean_old" json:"clean_old" default:"false"`
	LastMsgId     int64    `bson:"last_msg_id,omitempty" json:"last_msg_id,omitempty"`
	ShouldWelcome bool     `bson:"enabled" json:"welcome_enabled" default:"true"`
	WelcomeText   string   `bson:"text,omitempty" json:"welcome_text,omitempty"`
	FileID        string   `bson:"file_id,omitempty" json:"file_id,omitempty"`
	WelcomeType   int      `bson:"type,omitempty" json:"welcome_type,omitempty"`
	Button        []Button `bson:"btns,omitempty" json:"btns,omitempty"`
}

// check Chat Welcome Settings, used to get data before performing any operation
func checkGreetingSettings(chatID int64) (greetingSrc *GreetingSettings) {
	defaultGreetSrc := &GreetingSettings{
		ChatID:             chatID,
		ShouldCleanService: false,
		WelcomeSettings: &WelcomeSettings{
			LastMsgId:     0,
			CleanWelcome:  false,
			ShouldWelcome: true,
			WelcomeText:   DefaultWelcome,
			WelcomeType:   TEXT,
		},
		GoodbyeSettings: &GoodbyeSettings{
			LastMsgId:     0,
			CleanGoodbye:  false,
			ShouldGoodbye: false,
			GoodbyeText:   DefaultGoodbye,
			GoodbyeType:   TEXT,
		},
	}
	errS := findOne(greetingsColl, bson.M{"_id": chatID}).Decode(&greetingSrc)
	if errS == mongo.ErrNoDocuments {
		greetingSrc = defaultGreetSrc
		err := updateOne(greetingsColl, bson.M{"_id": chatID}, defaultGreetSrc)
		if err != nil {
			log.Errorf("[Database][checkGreetingSettings]: %v ", err)
		}
	} else if errS != nil {
		log.Errorf("[Database][checkGreetingSettings]: %v", errS)
		greetingSrc = defaultGreetSrc
	}
	return greetingSrc
}

func GetGreetingSettings(chatID int64) *GreetingSettings {
	return checkGreetingSettings(chatID)
}

func GetWelcomeButtons(chatId int64) []Button {
	btns := checkGreetingSettings(chatId).WelcomeSettings.Button
	return btns
}

func GetGoodbyeButtons(chatId int64) []Button {
	btns := checkGreetingSettings(chatId).GoodbyeSettings.Button
	return btns
}

func SetWelcomeText(chatID int64, welcometxt, fileId string, buttons []Button, welcType int) {
	welcomeSrc := checkGreetingSettings(chatID)
	welcomeSrc.WelcomeSettings.WelcomeText = welcometxt
	welcomeSrc.WelcomeSettings.Button = buttons
	welcomeSrc.WelcomeSettings.WelcomeType = welcType
	welcomeSrc.WelcomeSettings.FileID = fileId
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, welcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetWelcomeText]: %v", err)
		return
	}
}

func SetWelcomeToggle(chatID int64, pref bool) {
	welcomeSrc := checkGreetingSettings(chatID)
	welcomeSrc.WelcomeSettings.ShouldWelcome = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, welcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetWelcomeToggle]: %v", err)
		return
	}
}

func SetGoodbyeText(chatID int64, goodbyetext, fileId string, buttons []Button, goodbyeType int) {
	goodbyeSrc := checkGreetingSettings(chatID)
	goodbyeSrc.GoodbyeSettings.GoodbyeText = goodbyetext
	goodbyeSrc.GoodbyeSettings.Button = buttons
	goodbyeSrc.GoodbyeSettings.GoodbyeType = goodbyeType
	goodbyeSrc.GoodbyeSettings.FileID = fileId
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, goodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetGoodbyeText]: %v", err)
		return
	}
}

func SetGoodbyeToggle(chatID int64, pref bool) {
	goodbyeSrc := checkGreetingSettings(chatID)
	goodbyeSrc.GoodbyeSettings.ShouldGoodbye = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, goodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetGoodbyeToggle]: %v", err)
		return
	}
}

func SetShouldCleanService(chatID int64, pref bool) {
	cleanServiceSrc := checkGreetingSettings(chatID)
	cleanServiceSrc.ShouldCleanService = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, cleanServiceSrc)
	if err != nil {
		log.Errorf("[Database][SetShouldCleanService]: %v", err)
		return
	}
}

func SetShouldAutoApprove(chatID int64, pref bool) {
	autoApproveSrc := checkGreetingSettings(chatID)
	autoApproveSrc.ShouldAutoApprove = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, autoApproveSrc)
	if err != nil {
		log.Errorf("[Database][SetShouldAutoApprove]: %v", err)
		return
	}
}

func SetCleanWelcomeSetting(chatID int64, pref bool) {
	cleanWelcomeSrc := checkGreetingSettings(chatID)
	cleanWelcomeSrc.WelcomeSettings.CleanWelcome = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanWelcomeSetting]: %v", err)
		return
	}
}

func SetCleanWelcomeMsgId(chatId, msgId int64) {
	cleanWelcomeSrc := checkGreetingSettings(chatId)
	cleanWelcomeSrc.WelcomeSettings.LastMsgId = msgId
	err := updateOne(greetingsColl, bson.M{"_id": chatId}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanWelcomeMsgId]: %v", err)
		return
	}
}

func SetCleanGoodbyeSetting(chatID int64, pref bool) {
	cleanGoodbyeSrc := checkGreetingSettings(chatID)
	cleanGoodbyeSrc.GoodbyeSettings.CleanGoodbye = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, cleanGoodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanGoodbyeSetting]: %v", err)
		return
	}
}

func SetCleanGoodbyeMsgId(chatId, msgId int64) {
	cleanWelcomeSrc := checkGreetingSettings(chatId)
	cleanWelcomeSrc.GoodbyeSettings.LastMsgId = msgId
	err := updateOne(greetingsColl, bson.M{"_id": chatId}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanGoodbyeMsgId]: %v", err)
		return
	}
}

func LoadGreetingsStats() (enabledWelcome, enabledGoodbye, cleanServiceEnabled, cleanWelcomeEnabled, cleanGoodbyeEnabled int64) {
	var greetRcStruct []*GreetingSettings

	cursor := findAll(greetingsColl, bson.M{})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &greetRcStruct)

	for _, greetRc := range greetRcStruct {
		// count things
		if greetRc.WelcomeSettings.ShouldWelcome {
			enabledWelcome++
		}
		if greetRc.GoodbyeSettings.ShouldGoodbye {
			enabledGoodbye++
		}
		if greetRc.ShouldCleanService {
			cleanServiceEnabled++
		}
		if greetRc.WelcomeSettings.CleanWelcome {
			cleanWelcomeEnabled++
		}
		if greetRc.GoodbyeSettings.CleanGoodbye {
			cleanGoodbyeEnabled++
		}
	}

	return
}
