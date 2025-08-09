package token

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// BotInfo represents the response from Telegram's getMe API
type BotInfo struct {
	ID                      int64  `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	Username                string `json:"username"`
	CanJoinGroups           bool   `json:"can_join_groups"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries"`
}

// TelegramAPIResponse represents the standard Telegram API response structure
type TelegramAPIResponse struct {
	OK     bool    `json:"ok"`
	Result BotInfo `json:"result,omitempty"`
	ErrorCode int  `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

// ExtractBotID extracts the bot ID from a Telegram bot token
// Token format: {BOT_ID}:{SECRET}
// Example: "123456789:AaZz0_secret_part" -> 123456789
func ExtractBotID(token string) (int64, error) {
	if token == "" {
		return 0, fmt.Errorf("token cannot be empty")
	}

	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid token format: expected 'BOT_ID:SECRET', got %q", token)
	}

	botIDStr := parts[0]
	if botIDStr == "" {
		return 0, fmt.Errorf("bot ID cannot be empty")
	}

	botID, err := strconv.ParseInt(botIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid bot ID %q: must be a valid integer", botIDStr)
	}

	if botID <= 0 {
		return 0, fmt.Errorf("invalid bot ID %d: must be positive", botID)
	}

	return botID, nil
}

// HashToken creates a SHA256 hash of the token for secure storage
// This allows us to validate tokens without storing them in plain text
func HashToken(token string) string {
	if token == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateTokenWithTelegram validates a token by calling Telegram's getMe API
// Returns bot information if the token is valid, or an error if invalid
func ValidateTokenWithTelegram(token string, timeout time.Duration) (*BotInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	if timeout == 0 {
		timeout = 10 * time.Second // Default timeout
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	log.Debugf("[Token] Validating token with Telegram API (bot_id: %s)", strings.Split(token, ":")[0])

	// Make the API request
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Telegram API: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warnf("[Token] Failed to close response body: %v", closeErr)
		}
	}()

	// Parse the response
	var apiResp TelegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Telegram API response: %w", err)
	}

	// Check if the API call was successful
	if !apiResp.OK {
		return nil, fmt.Errorf("telegram API returned error %d: %s", apiResp.ErrorCode, apiResp.Description)
	}

	// Verify that this is actually a bot
	if !apiResp.Result.IsBot {
		return nil, fmt.Errorf("token belongs to a user account, not a bot")
	}

	// Verify that the bot ID from the token matches the API response
	tokenBotID, err := ExtractBotID(token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bot ID from token: %w", err)
	}

	if apiResp.Result.ID != tokenBotID {
		return nil, fmt.Errorf("bot ID mismatch: token claims %d but API returned %d", tokenBotID, apiResp.Result.ID)
	}

	log.Debugf("[Token] Token validation successful for bot @%s (ID: %d)", apiResp.Result.Username, apiResp.Result.ID)

	return &apiResp.Result, nil
}

// IsValidTokenFormat checks if a token has the correct format without making API calls
// Returns true if the token format is valid (BOT_ID:SECRET format)
func IsValidTokenFormat(token string) bool {
	_, err := ExtractBotID(token)
	return err == nil
}

// SanitizeToken returns a safe representation of the token for logging
// Shows only the bot ID part: "123456789:***"
func SanitizeToken(token string) string {
	if token == "" {
		return "***"
	}

	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return "***"
	}

	return fmt.Sprintf("%s:***", parts[0])
}