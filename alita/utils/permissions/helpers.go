package permissions

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

// PermissionConfig defines what permissions are required for a command
type PermissionConfig struct {
	RequireGroup        bool
	RequireUserAdmin    bool
	RequireBotAdmin     bool
	RequireUserRestrict bool
	RequireBotRestrict  bool
	RequireUserDelete   bool
	RequireBotDelete    bool
	RequireUserPromote  bool
	RequireBotPromote   bool
	RequireUserOwner    bool
	RequireUserPin      bool
	RequireBotPin       bool
}

var (
	// Common permission configurations for different types of operations
	CommonRestrictionPerms = PermissionConfig{
		RequireGroup:        true,
		RequireUserAdmin:    true,
		RequireBotAdmin:     true,
		RequireUserRestrict: true,
		RequireBotRestrict:  true,
	}

	CommonAdminPerms = PermissionConfig{
		RequireGroup:     true,
		RequireUserAdmin: true,
		RequireBotAdmin:  true,
	}

	CommonPromotionPerms = PermissionConfig{
		RequireGroup:       true,
		RequireUserAdmin:   true,
		RequireBotAdmin:    true,
		RequireUserPromote: true,
		RequireBotPromote:  true,
	}

	CommonPinPerms = PermissionConfig{
		RequireGroup:     true,
		RequireUserAdmin: true,
		RequireBotAdmin:  true,
		RequireUserPin:   true,
		RequireBotPin:    true,
	}
)

// CheckPermissions validates all required permissions based on config
// Returns true if all permissions pass, false otherwise
func CheckPermissions(b *gotgbot.Bot, ctx *ext.Context, config PermissionConfig) bool {
	user := ctx.EffectiveSender.User

	if config.RequireGroup && !chat_status.RequireGroup(b, ctx, nil, false) {
		return false
	}
	if config.RequireUserAdmin && !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return false
	}
	if config.RequireBotAdmin && !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return false
	}
	if config.RequireUserRestrict && !chat_status.CanUserRestrict(b, ctx, nil, user.Id, false) {
		return false
	}
	if config.RequireBotRestrict && !chat_status.CanBotRestrict(b, ctx, nil, false) {
		return false
	}
	if config.RequireUserDelete && !chat_status.CanUserDelete(b, ctx, nil, user.Id, false) {
		return false
	}
	if config.RequireBotDelete && !chat_status.CanBotDelete(b, ctx, nil, false) {
		return false
	}
	if config.RequireUserPromote && !chat_status.CanUserPromote(b, ctx, nil, user.Id, false) {
		return false
	}
	if config.RequireBotPromote && !chat_status.CanBotPromote(b, ctx, nil, false) {
		return false
	}
	if config.RequireUserOwner && !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
		return false
	}
	if config.RequireUserPin && !chat_status.CanUserPin(b, ctx, nil, user.Id, false) {
		return false
	}
	if config.RequireBotPin && !chat_status.CanBotPin(b, ctx, nil, false) {
		return false
	}

	return true
}

// UserExtractionResult contains the result of user extraction and validation
type UserExtractionResult struct {
	UserID       int64
	Reason       string
	IsValid      bool
	ErrorMessage string
}

// ExtractAndValidateUser extracts user ID and handles common validation scenarios
// Returns UserExtractionResult with validation status and error message if any
func ExtractAndValidateUser(b *gotgbot.Bot, ctx *ext.Context, allowSelf bool) UserExtractionResult {
	userId, reason := extraction.ExtractUserAndText(b, ctx)

	// Handle extraction errors
	if userId == -1 {
		return UserExtractionResult{IsValid: false}
	}

	// Handle anonymous users
	if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		return UserExtractionResult{
			UserID:       userId,
			IsValid:      false,
			ErrorMessage: "This command cannot be used on anonymous user, these user can only be banned/unbanned.",
		}
	}

	// Handle missing user specification
	if userId == 0 {
		return UserExtractionResult{
			UserID:       userId,
			IsValid:      false,
			ErrorMessage: "I don't know who you're talking about, you're going to need to specify a user...!",
		}
	}

	// Handle bot self-targeting (if not allowed)
	if !allowSelf && userId == b.Id {
		return UserExtractionResult{
			UserID:       userId,
			IsValid:      false,
			ErrorMessage: "I can't perform this action on myself.",
		}
	}

	return UserExtractionResult{
		UserID:  userId,
		Reason:  reason,
		IsValid: true,
	}
}

// HandleUserExtractionError sends appropriate error message for user extraction failures
func HandleUserExtractionError(b *gotgbot.Bot, ctx *ext.Context, result UserExtractionResult) error {
	if result.IsValid {
		return nil // No error to handle
	}

	if result.ErrorMessage == "" {
		return nil // No message to send
	}

	msg := ctx.EffectiveMessage
	_, err := msg.Reply(b, result.ErrorMessage, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// CheckUserRestrictionProtections validates common user restrictions (admin protection, in chat, etc.)
func CheckUserRestrictionProtections(b *gotgbot.Bot, ctx *ext.Context, userID int64, requireInChat bool) (bool, string) {
	chat := ctx.EffectiveChat

	// Check if user is in chat (if required)
	if requireInChat && !chat_status.IsUserInChat(b, chat, userID) {
		return false, "This user is not in this chat, how can I restrict them?"
	}

	// Check if user is protected from restrictions
	if chat_status.IsUserBanProtected(b, ctx, nil, userID) {
		return false, "I can't restrict an admin."
	}

	return true, ""
}

// HandleRestrictionProtectionError sends appropriate error message for restriction protection failures
func HandleRestrictionProtectionError(b *gotgbot.Bot, ctx *ext.Context, errorMessage string) error {
	if errorMessage == "" {
		return nil // No error to handle
	}

	msg := ctx.EffectiveMessage
	_, err := msg.Reply(b, errorMessage, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// PerformCommonRestrictionChecks combines permission checking, user extraction, and protection validation
// This is the main helper that most restriction commands (ban, mute, warn) can use
func PerformCommonRestrictionChecks(b *gotgbot.Bot, ctx *ext.Context, config PermissionConfig, requireInChat bool) (int64, string, bool) {
	// Check permissions
	if !CheckPermissions(b, ctx, config) {
		return 0, "", false
	}

	// Extract and validate user
	userResult := ExtractAndValidateUser(b, ctx, false)
	if !userResult.IsValid {
		HandleUserExtractionError(b, ctx, userResult)
		return 0, "", false
	}

	// Check restriction protections
	canRestrict, errorMessage := CheckUserRestrictionProtections(b, ctx, userResult.UserID, requireInChat)
	if !canRestrict {
		HandleRestrictionProtectionError(b, ctx, errorMessage)
		return 0, "", false
	}

	return userResult.UserID, userResult.Reason, true
}

// PerformCommonPromotionChecks handles common promotion/demotion permission checks and user extraction
func PerformCommonPromotionChecks(b *gotgbot.Bot, ctx *ext.Context, config PermissionConfig) (int64, string, bool) {
	// Check all required permissions first
	if !CheckPermissions(b, ctx, config) {
		return 0, "", false
	}

	// Extract and validate user (don't allow self for promotion/demotion)
	result := ExtractAndValidateUser(b, ctx, false)

	if !result.IsValid {
		return 0, "", false
	}

	// For promotion/demotion, we need to extract the text as well (for custom title)
	_, text := extraction.ExtractUserAndText(b, ctx)

	return result.UserID, text, true
}
