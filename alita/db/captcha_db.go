package db

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
		return errors.New("invalid captcha mode: must be 'math' or 'text'")
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
		return errors.New("timeout must be between 1 and 10 minutes")
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
		return errors.New("invalid failure action: must be 'kick', 'ban', or 'mute'")
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

// SetCaptchaMaxAttempts sets the maximum number of attempts allowed for captcha verification.
func SetCaptchaMaxAttempts(chatID int64, maxAttempts int) error {
	if maxAttempts < 1 || maxAttempts > 10 {
		return errors.New("max attempts must be between 1 and 10")
	}
	
	settings := &CaptchaSettings{
		ChatID:      chatID,
		MaxAttempts: maxAttempts,
	}
	
	err := DB.Where("chat_id = ?", chatID).Assign(settings).FirstOrCreate(&CaptchaSettings{}).Error
	if err != nil {
		log.Errorf("[Database][SetCaptchaMaxAttempts]: %v", err)
		return err
	}
	
	return nil
}

// CreateCaptchaAttempt creates a new captcha attempt record for a user.
// This is called when a new member joins and needs to complete captcha.
func CreateCaptchaAttempt(userID, chatID int64, answer string, messageID int64, timeout int) error {
	attempt := &CaptchaAttempts{
		UserID:    userID,
		ChatID:    chatID,
		Answer:    answer,
		Attempts:  0,
		MessageID: messageID,
        RefreshCount: 0,
		ExpiresAt: time.Now().Add(time.Duration(timeout) * time.Minute),
	}
	
	// Delete any existing attempt for this user in this chat
	_ = DeleteCaptchaAttempt(userID, chatID)
	
	return CreateRecord(attempt)
}

// UpdateCaptchaAttemptOnRefresh updates answer, messageID and increments refresh_count for an existing attempt.
// It preserves Attempts and ExpiresAt and returns the updated record.
func UpdateCaptchaAttemptOnRefresh(userID, chatID int64, newAnswer string, newMessageID int64) (*CaptchaAttempts, error) {
    attempt := &CaptchaAttempts{}
    err := DB.Where("user_id = ? AND chat_id = ? AND expires_at > ?", userID, chatID, time.Now()).First(attempt).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        log.Errorf("[Database][UpdateCaptchaAttemptOnRefresh:Find]: %v", err)
        return nil, err
    }

    updates := map[string]interface{}{
        "answer":        newAnswer,
        "message_id":    newMessageID,
        "refresh_count": gorm.Expr("COALESCE(refresh_count, 0) + 1"),
    }

    if err := UpdateRecord(&CaptchaAttempts{}, map[string]interface{}{"id": attempt.ID}, updates); err != nil {
        return nil, err
    }

    // Reload updated attempt
    err = DB.First(attempt, attempt.ID).Error
    if err != nil {
        log.Errorf("[Database][UpdateCaptchaAttemptOnRefresh:Reload]: %v", err)
        return nil, err
    }
    return attempt, nil
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
		return nil, errors.New("no active captcha attempt found")
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

// GetAllActiveCaptchaAttempts retrieves all active captcha attempts for a chat.
// Useful for admin overview and bulk operations.
func GetAllActiveCaptchaAttempts(chatID int64) ([]*CaptchaAttempts, error) {
	var attempts []*CaptchaAttempts
	err := DB.Where("chat_id = ? AND expires_at > ?", chatID, time.Now()).Find(&attempts).Error
	
	if err != nil {
		log.Errorf("[Database][GetAllActiveCaptchaAttempts]: %v", err)
		return nil, err
	}
	
	return attempts, nil
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

// IsCaptchaEnabled checks if captcha is enabled for a chat.
// This is a convenience function for quick checks.
func IsCaptchaEnabled(chatID int64) bool {
	settings, err := GetCaptchaSettings(chatID)
	if err != nil {
		return false
	}
	return settings.Enabled
}