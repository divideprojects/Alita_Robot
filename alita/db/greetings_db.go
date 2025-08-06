package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// check Chat Welcome Settings, used to get data before performing any operation
func checkGreetingSettings(chatID int64) (greetingSrc *GreetingSettings) {
	greetingSrc = &GreetingSettings{}
	err := GetRecord(greetingSrc, map[string]interface{}{"chat_id": chatID})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Ensure chat exists before creating greeting settings
		if !ChatExists(chatID) {
			// Chat doesn't exist, return default settings without creating record
			log.Warnf("[Database][checkGreetingSettings]: Chat %d doesn't exist, returning default settings", chatID)
			return &GreetingSettings{
				ChatID:             chatID,
				ShouldCleanService: false,
				WelcomeSettings: &WelcomeSettings{
					LastMsgId:     0,
					CleanWelcome:  false,
					ShouldWelcome: true,
					WelcomeText:   DefaultWelcome,
					WelcomeType:   TEXT,
					Button:        ButtonArray{},
				},
				GoodbyeSettings: &GoodbyeSettings{
					LastMsgId:     0,
					CleanGoodbye:  false,
					ShouldGoodbye: false,
					GoodbyeText:   DefaultGoodbye,
					GoodbyeType:   TEXT,
					Button:        ButtonArray{},
				},
			}
		}

		// Create default settings only if chat exists
		greetingSrc = &GreetingSettings{
			ChatID:             chatID,
			ShouldCleanService: false,
			WelcomeSettings: &WelcomeSettings{
				LastMsgId:     0,
				CleanWelcome:  false,
				ShouldWelcome: true,
				WelcomeText:   DefaultWelcome,
				WelcomeType:   TEXT,
				Button:        ButtonArray{},
			},
			GoodbyeSettings: &GoodbyeSettings{
				LastMsgId:     0,
				CleanGoodbye:  false,
				ShouldGoodbye: false,
				GoodbyeText:   DefaultGoodbye,
				GoodbyeType:   TEXT,
				Button:        ButtonArray{},
			},
		}

		err := CreateRecord(greetingSrc)
		if err != nil {
			log.Errorf("[Database][checkGreetingSettings]: %v ", err)
		}
	} else if err != nil {
		log.Errorf("[Database][checkGreetingSettings]: %v", err)
		// Return default settings on error
		greetingSrc = &GreetingSettings{
			ChatID:             chatID,
			ShouldCleanService: false,
			WelcomeSettings: &WelcomeSettings{
				LastMsgId:     0,
				CleanWelcome:  false,
				ShouldWelcome: true,
				WelcomeText:   DefaultWelcome,
				WelcomeType:   TEXT,
				Button:        ButtonArray{},
			},
			GoodbyeSettings: &GoodbyeSettings{
				LastMsgId:     0,
				CleanGoodbye:  false,
				ShouldGoodbye: false,
				GoodbyeText:   DefaultGoodbye,
				GoodbyeType:   TEXT,
				Button:        ButtonArray{},
			},
		}
	}
	
	// Ensure WelcomeSettings and GoodbyeSettings are initialized even for existing records
	if greetingSrc.WelcomeSettings == nil {
		greetingSrc.WelcomeSettings = &WelcomeSettings{
			LastMsgId:     0,
			CleanWelcome:  false,
			ShouldWelcome: true,
			WelcomeText:   DefaultWelcome,
			WelcomeType:   TEXT,
			Button:        ButtonArray{},
		}
	}
	if greetingSrc.GoodbyeSettings == nil {
		greetingSrc.GoodbyeSettings = &GoodbyeSettings{
			LastMsgId:     0,
			CleanGoodbye:  false,
			ShouldGoodbye: false,
			GoodbyeText:   DefaultGoodbye,
			GoodbyeType:   TEXT,
			Button:        ButtonArray{},
		}
	}
	
	return greetingSrc
}

func GetGreetingSettings(chatID int64) *GreetingSettings {
	return checkGreetingSettings(chatID)
}

func GetWelcomeButtons(chatId int64) []Button {
	greetingSettings := checkGreetingSettings(chatId)
	if greetingSettings.WelcomeSettings != nil {
		return []Button(greetingSettings.WelcomeSettings.Button)
	}
	return []Button{}
}

func GetGoodbyeButtons(chatId int64) []Button {
	greetingSettings := checkGreetingSettings(chatId)
	if greetingSettings.GoodbyeSettings != nil {
		return []Button(greetingSettings.GoodbyeSettings.Button)
	}
	return []Button{}
}

func SetWelcomeText(chatID int64, welcometxt, fileId string, buttons []Button, welcType int) {
	welcomeSrc := checkGreetingSettings(chatID)
	if welcomeSrc.WelcomeSettings == nil {
		welcomeSrc.WelcomeSettings = &WelcomeSettings{}
	}
	welcomeSrc.WelcomeSettings.WelcomeText = welcometxt
	welcomeSrc.WelcomeSettings.Button = ButtonArray(buttons)
	welcomeSrc.WelcomeSettings.WelcomeType = welcType
	welcomeSrc.WelcomeSettings.FileID = fileId

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, welcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetWelcomeText]: %v", err)
		return
	}
}

func SetWelcomeToggle(chatID int64, pref bool) {
	welcomeSrc := checkGreetingSettings(chatID)
	if welcomeSrc.WelcomeSettings == nil {
		welcomeSrc.WelcomeSettings = &WelcomeSettings{}
	}
	welcomeSrc.WelcomeSettings.ShouldWelcome = pref

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, welcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetWelcomeToggle]: %v", err)
		return
	}
}

func SetGoodbyeText(chatID int64, goodbyetext, fileId string, buttons []Button, goodbyeType int) {
	goodbyeSrc := checkGreetingSettings(chatID)
	if goodbyeSrc.GoodbyeSettings == nil {
		goodbyeSrc.GoodbyeSettings = &GoodbyeSettings{}
	}
	goodbyeSrc.GoodbyeSettings.GoodbyeText = goodbyetext
	goodbyeSrc.GoodbyeSettings.Button = ButtonArray(buttons)
	goodbyeSrc.GoodbyeSettings.GoodbyeType = goodbyeType
	goodbyeSrc.GoodbyeSettings.FileID = fileId

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, goodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetGoodbyeText]: %v", err)
		return
	}
}

func SetGoodbyeToggle(chatID int64, pref bool) {
	goodbyeSrc := checkGreetingSettings(chatID)
	if goodbyeSrc.GoodbyeSettings == nil {
		goodbyeSrc.GoodbyeSettings = &GoodbyeSettings{}
	}
	goodbyeSrc.GoodbyeSettings.ShouldGoodbye = pref

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, goodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetGoodbyeToggle]: %v", err)
		return
	}
}

func SetShouldCleanService(chatID int64, pref bool) {
	cleanServiceSrc := checkGreetingSettings(chatID)
	cleanServiceSrc.ShouldCleanService = pref

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, cleanServiceSrc)
	if err != nil {
		log.Errorf("[Database][SetShouldCleanService]: %v", err)
		return
	}
}

func SetShouldAutoApprove(chatID int64, pref bool) {
	autoApproveSrc := checkGreetingSettings(chatID)
	autoApproveSrc.ShouldAutoApprove = pref

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, autoApproveSrc)
	if err != nil {
		log.Errorf("[Database][SetShouldAutoApprove]: %v", err)
		return
	}
}

func SetCleanWelcomeSetting(chatID int64, pref bool) {
	cleanWelcomeSrc := checkGreetingSettings(chatID)
	if cleanWelcomeSrc.WelcomeSettings == nil {
		cleanWelcomeSrc.WelcomeSettings = &WelcomeSettings{}
	}
	cleanWelcomeSrc.WelcomeSettings.CleanWelcome = pref

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanWelcomeSetting]: %v", err)
		return
	}
}

func SetCleanWelcomeMsgId(chatId, msgId int64) {
	cleanWelcomeSrc := checkGreetingSettings(chatId)
	if cleanWelcomeSrc.WelcomeSettings == nil {
		cleanWelcomeSrc.WelcomeSettings = &WelcomeSettings{}
	}
	cleanWelcomeSrc.WelcomeSettings.LastMsgId = msgId

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatId}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanWelcomeMsgId]: %v", err)
		return
	}
}

func SetCleanGoodbyeSetting(chatID int64, pref bool) {
	cleanGoodbyeSrc := checkGreetingSettings(chatID)
	if cleanGoodbyeSrc.GoodbyeSettings == nil {
		cleanGoodbyeSrc.GoodbyeSettings = &GoodbyeSettings{}
	}
	cleanGoodbyeSrc.GoodbyeSettings.CleanGoodbye = pref

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatID}, cleanGoodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanGoodbyeSetting]: %v", err)
		return
	}
}

func SetCleanGoodbyeMsgId(chatId, msgId int64) {
	cleanGoodbyeSrc := checkGreetingSettings(chatId)
	if cleanGoodbyeSrc.GoodbyeSettings == nil {
		cleanGoodbyeSrc.GoodbyeSettings = &GoodbyeSettings{}
	}
	cleanGoodbyeSrc.GoodbyeSettings.LastMsgId = msgId

	err := UpdateRecord(&GreetingSettings{}, map[string]interface{}{"chat_id": chatId}, cleanGoodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanGoodbyeMsgId]: %v", err)
		return
	}
}

func LoadGreetingsStats() (enabledWelcome, enabledGoodbye, cleanServiceEnabled, cleanWelcomeEnabled, cleanGoodbyeEnabled int64) {
	// Use a single query with COUNT and CASE WHEN for better performance
	type greetingStats struct {
		EnabledWelcome      int64
		EnabledGoodbye      int64
		CleanServiceEnabled int64
		CleanWelcomeEnabled int64
		CleanGoodbyeEnabled int64
	}

	var stats greetingStats
	query := `
		SELECT 
			COUNT(CASE WHEN welcome_enabled = true THEN 1 END) as enabled_welcome,
			COUNT(CASE WHEN goodbye_enabled = true THEN 1 END) as enabled_goodbye,
			COUNT(CASE WHEN clean_service_settings = true THEN 1 END) as clean_service_enabled,
			COUNT(CASE WHEN welcome_clean_old = true THEN 1 END) as clean_welcome_enabled,
			COUNT(CASE WHEN goodbye_clean_old = true THEN 1 END) as clean_goodbye_enabled
		FROM greetings
	`

	err := DB.Raw(query).Scan(&stats).Error
	if err != nil {
		log.Errorf("[Database][LoadGreetingsStats] querying stats: %v", err)
		return 0, 0, 0, 0, 0
	}

	return stats.EnabledWelcome, stats.EnabledGoodbye, stats.CleanServiceEnabled, stats.CleanWelcomeEnabled, stats.CleanGoodbyeEnabled
}
