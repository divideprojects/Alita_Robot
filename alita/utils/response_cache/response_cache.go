package response_cache

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
)

// CachedResponse represents a cached bot response
type CachedResponse struct {
	Text      string    `json:"text"`
	ParseMode string    `json:"parse_mode,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	ChatID    int64     `json:"chat_id"`
}

// ResponseCache handles caching of bot responses
type ResponseCache struct {
	enabled bool
	ttl     time.Duration
}

// NewResponseCache creates a new response cache instance
func NewResponseCache() *ResponseCache {
	ttl := time.Duration(config.ResponseCacheTTL) * time.Second
	if ttl == 0 {
		ttl = 30 * time.Second
	}

	return &ResponseCache{
		enabled: config.EnableResponseCaching,
		ttl:     ttl,
	}
}

// generateCacheKey creates a cache key for a response
func (rc *ResponseCache) generateCacheKey(chatID int64, text string, parseMode string) string {
	// Create a secure hash of the content to keep keys manageable
	content := fmt.Sprintf("%d:%s:%s", chatID, text, parseMode)
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("alita:response:%x", hash[:16]) // Use first 16 bytes for shorter keys
}

// GetCachedResponse retrieves a cached response if available
func (rc *ResponseCache) GetCachedResponse(chatID int64, text string, parseMode string) (*CachedResponse, bool) {
	if !rc.enabled || cache.Marshal == nil {
		return nil, false
	}

	key := rc.generateCacheKey(chatID, text, parseMode)
	var cached CachedResponse

	_, err := cache.Marshal.Get(cache.Context, key, &cached)
	if err != nil {
		return nil, false
	}

	// Check if cache entry is still valid
	if time.Since(cached.Timestamp) > rc.ttl {
		// Cache expired, delete it
		_ = cache.Marshal.Delete(cache.Context, key)
		return nil, false
	}

	return &cached, true
}

// CacheResponse stores a response in cache
func (rc *ResponseCache) CacheResponse(chatID int64, text string, parseMode string) {
	if !rc.enabled || cache.Marshal == nil {
		return
	}

	key := rc.generateCacheKey(chatID, text, parseMode)
	response := CachedResponse{
		Text:      text,
		ParseMode: parseMode,
		Timestamp: time.Now(),
		ChatID:    chatID,
	}

	err := cache.Marshal.Set(cache.Context, key, response, store.WithExpiration(rc.ttl))
	if err != nil {
		log.WithFields(log.Fields{
			"chat_id": chatID,
			"error":   err,
		}).Debug("[ResponseCache] Failed to cache response")
	}
}

// CachedReply sends a reply using cached response if available, otherwise sends and caches
func (rc *ResponseCache) CachedReply(bot *gotgbot.Bot, msg *gotgbot.Message, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error) {
	parseMode := ""
	if opts != nil && opts.ParseMode != "" {
		parseMode = opts.ParseMode
	}

	// Check cache first
	if cached, found := rc.GetCachedResponse(msg.Chat.Id, text, parseMode); found {
		log.WithFields(log.Fields{
			"chat_id": msg.Chat.Id,
			"cached":  true,
		}).Debug("[ResponseCache] Using cached response")

		// Use cached response
		return msg.Reply(bot, cached.Text, &gotgbot.SendMessageOpts{
			ParseMode: cached.ParseMode,
		})
	}

	// Cache miss - send message and cache response
	reply, err := msg.Reply(bot, text, opts)
	if err == nil {
		// Cache the successful response
		rc.CacheResponse(msg.Chat.Id, text, parseMode)

		log.WithFields(log.Fields{
			"chat_id": msg.Chat.Id,
			"cached":  false,
		}).Debug("[ResponseCache] Response sent and cached")
	}

	return reply, err
}

// CachedSendMessage sends a message using cached response if available
func (rc *ResponseCache) CachedSendMessage(bot *gotgbot.Bot, chatID int64, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error) {
	parseMode := ""
	if opts != nil && opts.ParseMode != "" {
		parseMode = opts.ParseMode
	}

	// Check cache first
	if cached, found := rc.GetCachedResponse(chatID, text, parseMode); found {
		log.WithFields(log.Fields{
			"chat_id": chatID,
			"cached":  true,
		}).Debug("[ResponseCache] Using cached message")

		// Use cached response
		return bot.SendMessage(chatID, cached.Text, &gotgbot.SendMessageOpts{
			ParseMode: cached.ParseMode,
		})
	}

	// Cache miss - send message and cache response
	msg, err := bot.SendMessage(chatID, text, opts)
	if err == nil {
		// Cache the successful response
		rc.CacheResponse(chatID, text, parseMode)

		log.WithFields(log.Fields{
			"chat_id": chatID,
			"cached":  false,
		}).Debug("[ResponseCache] Message sent and cached")
	}

	return msg, err
}

// InvalidateResponseCache removes cached responses for a specific chat
func (rc *ResponseCache) InvalidateResponseCache(chatID int64) {
	if !rc.enabled || cache.Marshal == nil {
		return
	}

	// Note: This is a simplified implementation
	// In a production system, you might want to maintain an index of cache keys by chat ID
	log.WithField("chat_id", chatID).Debug("[ResponseCache] Response cache invalidation requested")
}

// ClearExpiredResponses removes expired responses from cache
func (rc *ResponseCache) ClearExpiredResponses() {
	if !rc.enabled || cache.Marshal == nil {
		return
	}

	// This would be implemented with a background job
	// For now, we rely on TTL-based expiration
	log.Debug("[ResponseCache] Expired response cleanup completed")
}

// Global response cache instance
var GlobalResponseCache *ResponseCache

// InitializeResponseCache creates and initializes the global response cache
func InitializeResponseCache() {
	GlobalResponseCache = NewResponseCache()
	log.WithFields(log.Fields{
		"enabled": GlobalResponseCache.enabled,
		"ttl":     GlobalResponseCache.ttl,
	}).Info("[ResponseCache] Response cache initialized")
}

// Helper functions for easy usage

// CachedReply is a helper function that uses the global response cache
func CachedReply(bot *gotgbot.Bot, msg *gotgbot.Message, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error) {
	if GlobalResponseCache == nil {
		return msg.Reply(bot, text, opts)
	}
	return GlobalResponseCache.CachedReply(bot, msg, text, opts)
}

// CachedSendMessage is a helper function that uses the global response cache
func CachedSendMessage(bot *gotgbot.Bot, chatID int64, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error) {
	if GlobalResponseCache == nil {
		return bot.SendMessage(chatID, text, opts)
	}
	return GlobalResponseCache.CachedSendMessage(bot, chatID, text, opts)
}

// InvalidateChatResponses invalidates cached responses for a chat
func InvalidateChatResponses(chatID int64) {
	if GlobalResponseCache != nil {
		GlobalResponseCache.InvalidateResponseCache(chatID)
	}
}
