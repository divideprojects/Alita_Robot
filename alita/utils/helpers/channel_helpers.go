package helpers

import (
	"fmt"
	"strings"
)

// IsChannelID checks if a given chat ID represents a channel.
// In Telegram, channel IDs are negative numbers starting with -100.
func IsChannelID(chatID int64) bool {
	// Convert to string to check the prefix
	chatIDStr := fmt.Sprintf("%d", chatID)
	return strings.HasPrefix(chatIDStr, "-100")
}
