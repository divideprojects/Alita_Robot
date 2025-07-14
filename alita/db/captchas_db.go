package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CAPTCHA mode constants
const (
	CaptchaModeButton = "button"
	CaptchaModeText   = "text"
	CaptchaModeMath   = "math"
	CaptchaModeText2  = "text2"
)

// Default CAPTCHA settings
const (
	DefaultCaptchaButtonText = "Click here to prove you're human"
	DefaultCaptchaKickTime   = 5 * time.Minute
	DefaultCaptchaMuteTime   = 0 // disabled by default
)

// CaptchaSettings holds all CAPTCHA-related configuration for a chat.
//
// Fields:
//   - ChatID: Unique identifier for the chat.
//   - Enabled: Whether CAPTCHA is enabled for new members.
//   - Mode: CAPTCHA mode (button, text, math, text2).
//   - ButtonText: Custom text for the CAPTCHA button.
//   - KickEnabled: Whether to kick users who don't solve CAPTCHA.
//   - KickTime: Time after which to kick unsolved users.
//   - RulesEnabled: Whether to show rules as part of CAPTCHA.
//   - MuteTime: Time after which to auto-unmute (0 = disabled).
type CaptchaSettings struct {
	ChatID       int64         `bson:"_id,omitempty" json:"_id,omitempty"`
	Enabled      bool          `bson:"enabled" json:"enabled" default:"false"`
	Mode         string        `bson:"mode" json:"mode" default:"button"`
	ButtonText   string        `bson:"button_text" json:"button_text"`
	KickEnabled  bool          `bson:"kick_enabled" json:"kick_enabled" default:"false"`
	KickTime     time.Duration `bson:"kick_time" json:"kick_time"`
	RulesEnabled bool          `bson:"rules_enabled" json:"rules_enabled" default:"false"`
	MuteTime     time.Duration `bson:"mute_time" json:"mute_time" default:"0"`
}

// CaptchaChallenge represents an active CAPTCHA challenge for a user.
//
// Fields:
//   - UserID: Unique identifier for the user.
//   - ChatID: Unique identifier for the chat.
//   - ChallengeData: JSON string containing challenge-specific data.
//   - CorrectAnswer: The correct answer for the challenge.
//   - CreatedAt: When the challenge was created.
//   - ExpiresAt: When the challenge expires.
//   - Attempts: Number of failed attempts.
//   - Solved: Whether the challenge has been solved.
type CaptchaChallenge struct {
	UserID        int64     `bson:"user_id" json:"user_id"`
	ChatID        int64     `bson:"chat_id" json:"chat_id"`
	ChallengeData string    `bson:"challenge_data" json:"challenge_data"`
	CorrectAnswer string    `bson:"correct_answer" json:"correct_answer"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	ExpiresAt     time.Time `bson:"expires_at" json:"expires_at"`
	Attempts      int       `bson:"attempts" json:"attempts" default:"0"`
	Solved        bool      `bson:"solved" json:"solved" default:"false"`
}

// checkCaptchaSettings fetches CAPTCHA settings for a chat from the database.
// If no document exists, it creates one with default values.
func checkCaptchaSettings(chatID int64) (captchaSrc *CaptchaSettings) {
	defaultCaptchaSrc := &CaptchaSettings{
		ChatID:       chatID,
		Enabled:      false,
		Mode:         CaptchaModeButton,
		ButtonText:   DefaultCaptchaButtonText,
		KickEnabled:  false,
		KickTime:     DefaultCaptchaKickTime,
		RulesEnabled: false,
		MuteTime:     DefaultCaptchaMuteTime,
	}

	err := findOne(captchasColl, bson.M{"_id": chatID}).Decode(&captchaSrc)
	if err == mongo.ErrNoDocuments {
		captchaSrc = defaultCaptchaSrc
		err := updateOne(captchasColl, bson.M{"_id": chatID}, defaultCaptchaSrc)
		if err != nil {
			log.Errorf("[Database][checkCaptchaSettings]: %v ", err)
		}
	} else if err != nil {
		captchaSrc = defaultCaptchaSrc
		log.Errorf("[Database][checkCaptchaSettings]: %v", err)
	}
	return captchaSrc
}

// GetCaptchaSettings retrieves the CAPTCHA settings for a given chat ID.
// If no settings exist, it initializes them with default values.
func GetCaptchaSettings(chatID int64) *CaptchaSettings {
	return checkCaptchaSettings(chatID)
}

// SetCaptchaEnabled toggles CAPTCHA on/off for a chat.
func SetCaptchaEnabled(chatID int64, enabled bool) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.Enabled = enabled
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaEnabled: %v - %d", err, chatID)
	}
}

// SetCaptchaMode sets the CAPTCHA mode for a chat.
func SetCaptchaMode(chatID int64, mode string) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.Mode = mode
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaMode: %v - %d", err, chatID)
	}
}

// SetCaptchaButtonText sets custom button text for CAPTCHA.
func SetCaptchaButtonText(chatID int64, buttonText string) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.ButtonText = buttonText
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaButtonText: %v - %d", err, chatID)
	}
}

// ResetCaptchaButtonText resets button text to default.
func ResetCaptchaButtonText(chatID int64) {
	SetCaptchaButtonText(chatID, DefaultCaptchaButtonText)
}

// SetCaptchaKick toggles CAPTCHA kick functionality.
func SetCaptchaKick(chatID int64, enabled bool) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.KickEnabled = enabled
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaKick: %v - %d", err, chatID)
	}
}

// SetCaptchaKickTime sets the time after which to kick users.
func SetCaptchaKickTime(chatID int64, kickTime time.Duration) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.KickTime = kickTime
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaKickTime: %v - %d", err, chatID)
	}
}

// SetCaptchaRules toggles CAPTCHA rules functionality.
func SetCaptchaRules(chatID int64, enabled bool) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.RulesEnabled = enabled
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaRules: %v - %d", err, chatID)
	}
}

// SetCaptchaMuteTime sets the auto-unmute time (0 to disable).
func SetCaptchaMuteTime(chatID int64, muteTime time.Duration) {
	captchaSrc := checkCaptchaSettings(chatID)
	captchaSrc.MuteTime = muteTime
	err := updateOne(captchasColl, bson.M{"_id": chatID}, captchaSrc)
	if err != nil {
		log.Errorf("[Database] SetCaptchaMuteTime: %v - %d", err, chatID)
	}
}

// CreateCaptchaChallenge creates a new CAPTCHA challenge for a user.
func CreateCaptchaChallenge(userID, chatID int64, challengeData, correctAnswer string, expiresAt time.Time) error {
	challenge := &CaptchaChallenge{
		UserID:        userID,
		ChatID:        chatID,
		ChallengeData: challengeData,
		CorrectAnswer: correctAnswer,
		CreatedAt:     time.Now(),
		ExpiresAt:     expiresAt,
		Attempts:      0,
		Solved:        false,
	}

	err := updateOne(captchaChallengesColl, bson.M{"user_id": userID, "chat_id": chatID}, challenge)
	if err != nil {
		log.Errorf("[Database] CreateCaptchaChallenge: %v - %d:%d", err, chatID, userID)
	}
	return err
}

// GetCaptchaChallenge retrieves an active CAPTCHA challenge for a user.
func GetCaptchaChallenge(userID, chatID int64) (*CaptchaChallenge, error) {
	var challenge CaptchaChallenge
	err := findOne(captchaChallengesColl, bson.M{"user_id": userID, "chat_id": chatID}).Decode(&challenge)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		log.Errorf("[Database] GetCaptchaChallenge: %v - %d:%d", err, chatID, userID)
		return nil, err
	}
	return &challenge, nil
}

// UpdateCaptchaChallenge updates an existing CAPTCHA challenge.
func UpdateCaptchaChallenge(userID, chatID int64, challenge *CaptchaChallenge) error {
	err := updateOne(captchaChallengesColl, bson.M{"user_id": userID, "chat_id": chatID}, challenge)
	if err != nil {
		log.Errorf("[Database] UpdateCaptchaChallenge: %v - %d:%d", err, chatID, userID)
	}
	return err
}

// DeleteCaptchaChallenge removes a CAPTCHA challenge (after solving or expiry).
func DeleteCaptchaChallenge(userID, chatID int64) error {
	err := deleteOne(captchaChallengesColl, bson.M{"user_id": userID, "chat_id": chatID})
	if err != nil {
		log.Errorf("[Database] DeleteCaptchaChallenge: %v - %d:%d", err, chatID, userID)
	}
	return err
}

// GetExpiredCaptchaChallenges returns all challenges that have expired.
func GetExpiredCaptchaChallenges() ([]*CaptchaChallenge, error) {
	var challenges []*CaptchaChallenge
	now := time.Now()

	cursor := findAll(captchaChallengesColl, bson.M{
		"expires_at": bson.M{"$lt": now},
		"solved":     false,
	})
	defer cursor.Close(bgCtx)

	err := cursor.All(bgCtx, &challenges)
	if err != nil {
		log.Errorf("[Database] GetExpiredCaptchaChallenges: %v", err)
		return nil, err
	}

	return challenges, nil
}

// CleanupExpiredChallenges removes all expired challenges from the database.
func CleanupExpiredChallenges() error {
	now := time.Now()
	err := deleteMany(captchaChallengesColl, bson.M{
		"expires_at": bson.M{"$lt": now},
	})
	if err != nil {
		log.Errorf("[Database] CleanupExpiredChallenges: %v", err)
	}
	return err
}
