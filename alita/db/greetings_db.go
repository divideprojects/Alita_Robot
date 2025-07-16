package db

import (
	"context"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// default strings when no settings are set
const (
	DefaultWelcome = "Hey {first}, how are you?"
	DefaultGoodbye = "Sad to see you leaving {first}"
)

// GreetingSettings holds all greeting-related configuration for a chat.
//
// Fields:
//   - ChatID: Unique identifier for the chat.
//   - ShouldCleanService: Whether to clean service messages (e.g., join/leave).
//   - WelcomeSettings: Settings for welcome messages.
//   - GoodbyeSettings: Settings for goodbye messages.
//   - ShouldAutoApprove: Whether to auto-approve users on join.
type GreetingSettings struct {
	ChatID             int64            `bson:"_id,omitempty" json:"_id,omitempty"`
	ShouldCleanService bool             `bson:"clean_service_settings" json:"clean_service_settings" default:"false"`
	WelcomeSettings    *WelcomeSettings `bson:"welcome_settings" json:"welcome_settings" default:"false"`
	GoodbyeSettings    *GoodbyeSettings `bson:"goodbye_settings" json:"goodbye_settings" default:"false"`
	ShouldAutoApprove  bool             `bson:"auto_approve" json:"auto_approve" default:"false"`
}

// GoodbyeSettings holds configuration for goodbye messages in a chat.
//
// Fields:
//   - CleanGoodbye: Whether to clean up previous goodbye messages.
//   - LastMsgId: The last goodbye message ID sent.
//   - ShouldGoodbye: Whether goodbye messages are enabled.
//   - GoodbyeText: The goodbye message text.
//   - FileID: Optional file ID for media attachments.
//   - GoodbyeType: Type of goodbye message (e.g., text, media).
//   - Button: List of buttons to attach to the goodbye message.
type GoodbyeSettings struct {
	CleanGoodbye  bool     `bson:"clean_old" json:"clean_old" default:"false"`
	LastMsgId     int64    `bson:"last_msg_id,omitempty" json:"last_msg_id,omitempty"`
	ShouldGoodbye bool     `bson:"enabled" json:"enabled" default:"true"`
	GoodbyeText   string   `bson:"text,omitempty" json:"text,omitempty"`
	FileID        string   `bson:"file_id,omitempty" json:"file_id,omitempty"`
	GoodbyeType   int      `bson:"type,omitempty" json:"type,omitempty"`
	Button        []Button `bson:"btns,omitempty" json:"btns,omitempty"`
}

// WelcomeSettings holds configuration for welcome messages in a chat.
//
// Fields:
//   - CleanWelcome: Whether to clean up previous welcome messages.
//   - LastMsgId: The last welcome message ID sent.
//   - ShouldWelcome: Whether welcome messages are enabled.
//   - WelcomeText: The welcome message text.
//   - FileID: Optional file ID for media attachments.
//   - WelcomeType: Type of welcome message (e.g., text, media).
//   - Button: List of buttons to attach to the welcome message.
type WelcomeSettings struct {
	CleanWelcome  bool     `bson:"clean_old" json:"clean_old" default:"false"`
	LastMsgId     int64    `bson:"last_msg_id,omitempty" json:"last_msg_id,omitempty"`
	ShouldWelcome bool     `bson:"enabled" json:"welcome_enabled" default:"true"`
	WelcomeText   string   `bson:"text,omitempty" json:"welcome_text,omitempty"`
	FileID        string   `bson:"file_id,omitempty" json:"file_id,omitempty"`
	WelcomeType   int      `bson:"type,omitempty" json:"welcome_type,omitempty"`
	Button        []Button `bson:"btns,omitempty" json:"btns,omitempty"`
}

// checkGreetingSettings fetches greeting settings for a chat from the database.
// If no document exists, it creates one with default values (welcome enabled, goodbye disabled).
// Returns a pointer to the GreetingSettings struct with either existing or default values.
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

// GetGreetingSettings retrieves the greeting settings for a given chat ID.
// If no settings exist, it initializes them with default values (welcome enabled, goodbye disabled).
// This is the main function for accessing greeting settings.
func GetGreetingSettings(chatID int64) *GreetingSettings {
	return checkGreetingSettings(chatID)
}

// GetWelcomeButtons returns the list of buttons configured for welcome messages in a chat.
// Returns an empty slice if no buttons are configured for the welcome message.
func GetWelcomeButtons(chatId int64) []Button {
	btns := checkGreetingSettings(chatId).WelcomeSettings.Button
	return btns
}

// GetGoodbyeButtons returns the list of buttons configured for goodbye messages in a chat.
// Returns an empty slice if no buttons are configured for the goodbye message.
func GetGoodbyeButtons(chatId int64) []Button {
	btns := checkGreetingSettings(chatId).GoodbyeSettings.Button
	return btns
}

// SetWelcomeText updates the welcome message text, buttons, type, and file ID for a chat.
// Supports various message types including text, media, and interactive buttons.
// The welcType parameter defines the message format (text, photo, video, etc.).
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

// SetWelcomeToggle enables or disables welcome messages for a chat.
// When disabled, new members will not receive welcome messages.
func SetWelcomeToggle(chatID int64, pref bool) {
	welcomeSrc := checkGreetingSettings(chatID)
	welcomeSrc.WelcomeSettings.ShouldWelcome = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, welcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetWelcomeToggle]: %v", err)
		return
	}
}

// SetGoodbyeText updates the goodbye message text, buttons, type, and file ID for a chat.
// Supports various message types including text, media, and interactive buttons.
// The goodbyeType parameter defines the message format (text, photo, video, etc.).
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

// SetGoodbyeToggle enables or disables goodbye messages for a chat.
// When disabled, leaving members will not trigger goodbye messages.
func SetGoodbyeToggle(chatID int64, pref bool) {
	goodbyeSrc := checkGreetingSettings(chatID)
	goodbyeSrc.GoodbyeSettings.ShouldGoodbye = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, goodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetGoodbyeToggle]: %v", err)
		return
	}
}

// SetShouldCleanService sets whether service messages should be cleaned for a chat.
// When enabled, automatically deletes Telegram's default join/leave service messages.
// Helps keep the chat clean by removing system notifications.
func SetShouldCleanService(chatID int64, pref bool) {
	cleanServiceSrc := checkGreetingSettings(chatID)
	cleanServiceSrc.ShouldCleanService = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, cleanServiceSrc)
	if err != nil {
		log.Errorf("[Database][SetShouldCleanService]: %v", err)
		return
	}
}

// SetShouldAutoApprove sets whether users should be auto-approved on join for a chat.
// When enabled, new members are automatically approved without admin intervention.
// Only applicable for groups with join requests enabled.
func SetShouldAutoApprove(chatID int64, pref bool) {
	autoApproveSrc := checkGreetingSettings(chatID)
	autoApproveSrc.ShouldAutoApprove = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, autoApproveSrc)
	if err != nil {
		log.Errorf("[Database][SetShouldAutoApprove]: %v", err)
		return
	}
}

// SetCleanWelcomeSetting sets whether previous welcome messages should be cleaned for a chat.
// When enabled, deletes the previous welcome message before sending a new one.
// Helps prevent welcome message spam in active groups.
func SetCleanWelcomeSetting(chatID int64, pref bool) {
	cleanWelcomeSrc := checkGreetingSettings(chatID)
	cleanWelcomeSrc.WelcomeSettings.CleanWelcome = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanWelcomeSetting]: %v", err)
		return
	}
}

// SetCleanWelcomeMsgId updates the last welcome message ID for cleaning purposes.
// Stores the message ID of the most recent welcome message to enable deletion
// when the next welcome message is sent.
func SetCleanWelcomeMsgId(chatId, msgId int64) {
	cleanWelcomeSrc := checkGreetingSettings(chatId)
	cleanWelcomeSrc.WelcomeSettings.LastMsgId = msgId
	err := updateOne(greetingsColl, bson.M{"_id": chatId}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanWelcomeMsgId]: %v", err)
		return
	}
}

// SetCleanGoodbyeSetting sets whether previous goodbye messages should be cleaned for a chat.
// When enabled, deletes the previous goodbye message before sending a new one.
// Helps prevent goodbye message spam in active groups.
func SetCleanGoodbyeSetting(chatID int64, pref bool) {
	cleanGoodbyeSrc := checkGreetingSettings(chatID)
	cleanGoodbyeSrc.GoodbyeSettings.CleanGoodbye = pref
	err := updateOne(greetingsColl, bson.M{"_id": chatID}, cleanGoodbyeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanGoodbyeSetting]: %v", err)
		return
	}
}

// SetCleanGoodbyeMsgId updates the last goodbye message ID for cleaning purposes.
// Stores the message ID of the most recent goodbye message to enable deletion
// when the next goodbye message is sent.
func SetCleanGoodbyeMsgId(chatId, msgId int64) {
	cleanWelcomeSrc := checkGreetingSettings(chatId)
	cleanWelcomeSrc.GoodbyeSettings.LastMsgId = msgId
	err := updateOne(greetingsColl, bson.M{"_id": chatId}, cleanWelcomeSrc)
	if err != nil {
		log.Errorf("[Database][SetCleanGoodbyeMsgId]: %v", err)
		return
	}
}

// LoadGreetingsStats returns counts of chats with various greeting features enabled.
// Used for bot statistics and monitoring purposes to track feature adoption.
// Iterates through all greeting settings to count enabled features.
//
// Returns:
//   - enabledWelcome: Number of chats with welcome messages enabled.
//   - enabledGoodbye: Number of chats with goodbye messages enabled.
//   - cleanServiceEnabled: Number of chats with service message cleaning enabled.
//   - cleanWelcomeEnabled: Number of chats cleaning previous welcome messages.
//   - cleanGoodbyeEnabled: Number of chats cleaning previous goodbye messages.
func LoadGreetingsStats() (enabledWelcome, enabledGoodbye, cleanServiceEnabled, cleanWelcomeEnabled, cleanGoodbyeEnabled int64) {
	var greetRcStruct []*GreetingSettings

	cursor := findAll(greetingsColl, bson.M{})
	ctx := context.Background()
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error("Failed to close greetings cursor:", err)
		}
	}()
	if err := cursor.All(ctx, &greetRcStruct); err != nil {
		log.Error("Failed to load greetings stats:", err)
		return
	}

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
