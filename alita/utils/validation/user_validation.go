package validation

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	log "github.com/sirupsen/logrus"
)

// UserValidationResult contains the result of user validation
type UserValidationResult struct {
	UserID int64
	User   *gotgbot.User
	Reason string
	Valid  bool
	Error  error
}

// ValidateUser extracts and validates a user from the message context
// This replaces the repetitive user extraction and validation logic found throughout the codebase
func ValidateUser(bot *gotgbot.Bot, ctx *ext.Context) (*UserValidationResult, error) {
	msg := ctx.EffectiveMessage

	// Extract user ID using existing extraction utility
	userID := extraction.ExtractUser(bot, ctx)

	result := &UserValidationResult{
		UserID: userID,
		Valid:  false,
	}

	// Handle extraction failure
	if userID == -1 {
		result.Reason = "Failed to extract user from message"
		result.Error = fmt.Errorf("user extraction failed")
		return result, nil
	}

	// Handle anonymous users (channels)
	if strings.HasPrefix(fmt.Sprint(userID), "-100") {
		result.Reason = "This command cannot be used on anonymous user."
		_, err := msg.Reply(bot, result.Reason, nil)
		if err != nil {
			log.Error(err)
			result.Error = err
		}
		return result, nil
	}

	// Handle case where no user was specified
	if userID == 0 {
		result.Reason = "I don't know who you're talking about, you're going to need to specify a user...!"
		_, err := msg.Reply(bot, result.Reason, helpers.Shtml())
		if err != nil {
			log.Error(err)
			result.Error = err
		}
		return result, nil
	}

	// Get user information
	chat := ctx.EffectiveChat
	userMember, err := chat.GetMember(bot, userID, nil)
	if err != nil {
		result.Reason = "Failed to get user information"
		result.Error = err
		log.Error(err)
		return result, err
	}

	// Extract user from member
	user := userMember.MergeChatMember().User
	result.User = &user
	result.Valid = true

	return result, nil
}

// ValidateUserWithText extracts user and additional text from the message
// This is useful for commands that need both a user target and additional parameters
func ValidateUserWithText(bot *gotgbot.Bot, ctx *ext.Context) (userResult *UserValidationResult, text string, err error) {
	msg := ctx.EffectiveMessage

	// Extract user ID and text using existing extraction utility
	userID, extractedText := extraction.ExtractUserAndText(bot, ctx)

	result := &UserValidationResult{
		UserID: userID,
		Valid:  false,
	}

	// Handle extraction failure
	if userID == -1 {
		result.Reason = "Failed to extract user from message"
		result.Error = fmt.Errorf("user extraction failed")
		return result, extractedText, nil
	}

	// Handle anonymous users (channels)
	if strings.HasPrefix(fmt.Sprint(userID), "-100") {
		result.Reason = "This command cannot be used on anonymous user."
		_, err := msg.Reply(bot, result.Reason, nil)
		if err != nil {
			log.Error(err)
			result.Error = err
		}
		return result, extractedText, nil
	}

	// Handle case where no user was specified
	if userID == 0 {
		result.Reason = "I don't know who you're talking about, you're going to need to specify a user...!"
		_, err := msg.Reply(bot, result.Reason, helpers.Shtml())
		if err != nil {
			log.Error(err)
			result.Error = err
		}
		return result, extractedText, nil
	}

	// Get user information
	chat := ctx.EffectiveChat
	userMember, err := chat.GetMember(bot, userID, nil)
	if err != nil {
		result.Reason = "Failed to get user information"
		result.Error = err
		log.Error(err)
		return result, extractedText, err
	}

	// Extract user from member
	user := userMember.MergeChatMember().User
	result.User = &user
	result.Valid = true

	return result, extractedText, nil
}

// IsUserBot checks if the validated user is a bot
func (uvr *UserValidationResult) IsUserBot() bool {
	if !uvr.Valid || uvr.User == nil {
		return false
	}
	return uvr.User.IsBot
}

// GetUserMention returns an HTML mention for the validated user
func (uvr *UserValidationResult) GetUserMention() string {
	if !uvr.Valid || uvr.User == nil {
		return ""
	}
	return helpers.MentionHtml(uvr.User.Id, uvr.User.FirstName)
}

// GetUserFullName returns the full name of the validated user
func (uvr *UserValidationResult) GetUserFullName() string {
	if !uvr.Valid || uvr.User == nil {
		return ""
	}
	return helpers.GetFullName(uvr.User.FirstName, uvr.User.LastName)
}
