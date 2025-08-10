package db

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Captcha validation errors
var (
	ErrInvalidCaptchaMode   = errors.New("INVALID_CAPTCHA_MODE")
	ErrInvalidTimeout       = errors.New("INVALID_TIMEOUT_RANGE")
	ErrInvalidFailureAction = errors.New("INVALID_FAILURE_ACTION")
	ErrInvalidMaxAttempts   = errors.New("INVALID_MAX_ATTEMPTS")
	ErrNoActiveCaptcha      = errors.New("NO_ACTIVE_CAPTCHA")
)

// GetCaptchaSettings retrieves captcha settings for a chat.
// Returns default settings if the chat doesn't have custom settings.
func GetCaptchaSettings(chatID int64) (*CaptchaSettings, error) {
	settings := &CaptchaSettings{}
	err := GetRecord(settings, map[string]interface{}{"chat_id": chatID})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return default settings if not found
		return &CaptchaSettings{
			ChatID:        chatID,
			Enabled:       false,
			CaptchaMode:   "math",
			Timeout:       2,
			FailureAction: "kick",
			MaxAttempts:   3,
		}, nil
	}

	if err != nil {
		log.Errorf("[Database][GetCaptchaSettings]: %v", err)
		return nil, err
	}

	return settings, nil
}

// SetCaptchaEnabled enables or disables captcha for a chat.
// Creates settings record if it doesn't exist.
func SetCaptchaEnabled(chatID int64, enabled bool) error {
	settings := &CaptchaSettings{
		ChatID:  chatID,
		Enabled: enabled,
	}

	err := DB.Where("chat_id = ?", chatID).Assign(settings).FirstOrCreate(&CaptchaSettings{}).Error
	if err != nil {
		log.Errorf("[Database][SetCaptchaEnabled]: %v", err)
		return err
	}

	return nil
}

// SetCaptchaMode sets the captcha mode (math or text) for a chat.
// Creates settings record if it doesn't exist.
func SetCaptchaMode(chatID int64, mode string) error {
	if mode != "math" && mode != "text" {
		return ErrInvalidCaptchaMode
	}

	settings := &CaptchaSettings{
		ChatID:      chatID,
		CaptchaMode: mode,
	}

	err := DB.Where("chat_id = ?", chatID).Assign(settings).FirstOrCreate(&CaptchaSettings{}).Error
	if err != nil {
		log.Errorf("[Database][SetCaptchaMode]: %v", err)
		return err
	}

	return nil
}

// SetCaptchaTimeout sets the timeout duration (in minutes) for captcha verification.
// Creates settings record if it doesn't exist.
func SetCaptchaTimeout(chatID int64, timeout int) error {
	if timeout < 1 || timeout > 10 {
		return ErrInvalidTimeout
	}

	settings := &CaptchaSettings{
		ChatID:  chatID,
		Timeout: timeout,
	}

	err := DB.Where("chat_id = ?", chatID).Assign(settings).FirstOrCreate(&CaptchaSettings{}).Error
	if err != nil {
		log.Errorf("[Database][SetCaptchaTimeout]: %v", err)
		return err
	}

	return nil
}

// SetCaptchaFailureAction sets the action to take when captcha verification fails.
// Valid actions are: kick, ban, mute
func SetCaptchaFailureAction(chatID int64, action string) error {
	if action != "kick" && action != "ban" && action != "mute" {
		return ErrInvalidFailureAction
	}

	settings := &CaptchaSettings{
		ChatID:        chatID,
		FailureAction: action,
	}

	err := DB.Where("chat_id = ?", chatID).Assign(settings).FirstOrCreate(&CaptchaSettings{}).Error
	if err != nil {
		log.Errorf("[Database][SetCaptchaFailureAction]: %v", err)
		return err
	}

	return nil
}

// CreateCaptchaAttemptPreMessage creates a captcha attempt before sending a message,
// setting message_id to 0 temporarily and returning the created attempt with ID.
func CreateCaptchaAttemptPreMessage(userID, chatID int64, answer string, timeout int) (*CaptchaAttempts, error) {
	attempt := &CaptchaAttempts{
		UserID:       userID,
		ChatID:       chatID,
		Answer:       answer,
		Attempts:     0,
		MessageID:    0,
		RefreshCount: 0,
		ExpiresAt:    time.Now().Add(time.Duration(timeout) * time.Minute),
	}

	// Delete any existing attempt for this user in this chat
	_ = DeleteCaptchaAttempt(userID, chatID)

	err := DB.Create(attempt).Error
	if err != nil {
		log.Errorf("[Database][CreateCaptchaAttemptPreMessage]: %v", err)
		return nil, err
	}
	return attempt, nil
}

// UpdateCaptchaAttemptMessageID sets the message_id for an existing attempt by ID.
func UpdateCaptchaAttemptMessageID(attemptID uint, messageID int64) error {
	err := DB.Model(&CaptchaAttempts{}).Where("id = ?", attemptID).Update("message_id", messageID).Error
	if err != nil {
		log.Errorf("[Database][UpdateCaptchaAttemptMessageID]: %v", err)
		return err
	}
	return nil
}

// GetCaptchaAttempt retrieves an active captcha attempt for a user in a chat.
// Returns nil if no active attempt exists or if it has expired.
func GetCaptchaAttempt(userID, chatID int64) (*CaptchaAttempts, error) {
	attempt := &CaptchaAttempts{}
	err := DB.Where("user_id = ? AND chat_id = ? AND expires_at > ?",
		userID, chatID, time.Now()).First(attempt).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Errorf("[Database][GetCaptchaAttempt]: %v", err)
		return nil, err
	}

	return attempt, nil
}

// IncrementCaptchaAttempts increments the attempt counter for a captcha.
// Returns the updated attempt record.
func IncrementCaptchaAttempts(userID, chatID int64) (*CaptchaAttempts, error) {
	attempt, err := GetCaptchaAttempt(userID, chatID)
	if err != nil {
		return nil, err
	}

	if attempt == nil {
		return nil, ErrNoActiveCaptcha
	}

	attempt.Attempts++
	err = DB.Save(attempt).Error
	if err != nil {
		log.Errorf("[Database][IncrementCaptchaAttempts]: %v", err)
		return nil, err
	}

	return attempt, nil
}

// DeleteCaptchaAttempt removes a captcha attempt record.
// Called when verification is successful or when user is kicked/banned.
func DeleteCaptchaAttempt(userID, chatID int64) error {
	result := DB.Where("user_id = ? AND chat_id = ?", userID, chatID).Delete(&CaptchaAttempts{})
	if result.Error != nil {
		log.Errorf("[Database][DeleteCaptchaAttempt]: %v", result.Error)
		return result.Error
	}
	return nil
}

// CleanupExpiredCaptchaAttempts removes all expired captcha attempts from the database.
// This should be called periodically to clean up old records.
func CleanupExpiredCaptchaAttempts() (int64, error) {
	result := DB.Where("expires_at < ?", time.Now()).Delete(&CaptchaAttempts{})
	if result.Error != nil {
		log.Errorf("[Database][CleanupExpiredCaptchaAttempts]: %v", result.Error)
		return 0, result.Error
	}

	if result.RowsAffected > 0 {
		log.Infof("[Database][CleanupExpiredCaptchaAttempts]: Cleaned up %d expired captcha attempts", result.RowsAffected)
	}

	return result.RowsAffected, nil
}

// DeleteAllCaptchaAttempts removes all captcha attempts for a chat.
// Used when captcha is disabled or for admin cleanup.
func DeleteAllCaptchaAttempts(chatID int64) error {
	result := DB.Where("chat_id = ?", chatID).Delete(&CaptchaAttempts{})
	if result.Error != nil {
		log.Errorf("[Database][DeleteAllCaptchaAttempts]: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Infof("[Database][DeleteAllCaptchaAttempts]: Deleted %d captcha attempts for chat %d", result.RowsAffected, chatID)
	}

	return nil
}

// UpdateCaptchaAttemptOnRefreshByID updates answer, message ID and increments refresh_count by attempt ID.
func UpdateCaptchaAttemptOnRefreshByID(attemptID uint, newAnswer string, newMessageID int64) (*CaptchaAttempts, error) {
	attempt := &CaptchaAttempts{}
	err := DB.First(attempt, attemptID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("[Database][UpdateCaptchaAttemptOnRefreshByID:Find]: %v", err)
		return nil, err
	}

	updates := map[string]interface{}{
		"answer":        newAnswer,
		"message_id":    newMessageID,
		"refresh_count": gorm.Expr("COALESCE(refresh_count, 0) + 1"),
	}
	if err := UpdateRecord(&CaptchaAttempts{}, map[string]interface{}{"id": attemptID}, updates); err != nil {
		return nil, err
	}
	// Reload
	err = DB.First(attempt, attemptID).Error
	if err != nil {
		log.Errorf("[Database][UpdateCaptchaAttemptOnRefreshByID:Reload]: %v", err)
		return nil, err
	}
	return attempt, nil
}
