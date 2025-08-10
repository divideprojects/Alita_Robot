package db

import (
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
)

// CachePrewarmer handles prewarming of frequently accessed data
type CachePrewarmer struct {
	enabled bool
}

// NewCachePrewarmer creates a new cache prewarmer instance
func NewCachePrewarmer() *CachePrewarmer {
	return &CachePrewarmer{
		enabled: config.EnableCachePrewarming,
	}
}

// PrewarmCaches loads frequently accessed data into cache during startup
// This reduces cache misses for commonly requested data
func (cp *CachePrewarmer) PrewarmCaches() error {
	if !cp.enabled || cache.Marshal == nil {
		log.Info("[CachePrewarming] Cache prewarming disabled or cache not available")
		return nil
	}

	log.Info("[CachePrewarming] Starting cache prewarming process...")
	startTime := time.Now()

	// Prewarm active chats (chats with activity in last 24 hours)
	if err := cp.prewarmActiveChats(); err != nil {
		log.WithError(err).Warn("[CachePrewarming] Failed to prewarm active chats")
	}

	// Prewarm active users (users with activity in last 7 days)
	if err := cp.prewarmActiveUsers(); err != nil {
		log.WithError(err).Warn("[CachePrewarming] Failed to prewarm active users")
	}

	// Prewarm language settings
	if err := cp.prewarmLanguageSettings(); err != nil {
		log.WithError(err).Warn("[CachePrewarming] Failed to prewarm language settings")
	}

	elapsed := time.Since(startTime)
	log.WithField("duration", elapsed).Info("[CachePrewarming] Cache prewarming completed")

	return nil
}

// prewarmSingleChat prewarns a single chat's data into cache
func (cp *CachePrewarmer) prewarmSingleChat(chat Chat) {
	// Cache chat data
	chatKey := chatCacheKey(chat.ChatId)
	if err := cache.Marshal.Set(cache.Context, chatKey, &chat, store.WithExpiration(CacheTTLChatSettings)); err != nil {
		log.WithFields(log.Fields{
			"chat_id": chat.ChatId,
			"error":   err,
		}).Debug("[CachePrewarming] Failed to cache chat")
		return
	}

	// Cache chat settings
	settingsKey := chatSettingsCacheKey(chat.ChatId)
	if err := cache.Marshal.Set(cache.Context, settingsKey, &chat, store.WithExpiration(CacheTTLChatSettings)); err != nil {
		log.WithFields(log.Fields{
			"chat_id": chat.ChatId,
			"error":   err,
		}).Debug("[CachePrewarming] Failed to cache chat settings")
	}

	// Cache language if set
	if chat.Language != "" {
		langKey := chatLanguageCacheKey(chat.ChatId)
		if err := cache.Marshal.Set(cache.Context, langKey, chat.Language, store.WithExpiration(CacheTTLLanguage)); err != nil {
			log.WithFields(log.Fields{
				"chat_id": chat.ChatId,
				"error":   err,
			}).Debug("[CachePrewarming] Failed to cache chat language")
		}
	}
}

// prewarmActiveChats loads recently active chats into cache using concurrent processing
func (cp *CachePrewarmer) prewarmActiveChats() error {
	var activeChats []Chat

	// Get chats with activity in last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	err := DB.Where("last_activity > ? OR last_activity IS NULL", cutoff).
		Limit(1000). // Limit to prevent memory issues
		Find(&activeChats).Error

	if err != nil {
		return err
	}

	log.WithField("count", len(activeChats)).Info("[CachePrewarming] Prewarming active chats")

	// For small numbers, process sequentially
	if len(activeChats) <= 10 {
		for _, chat := range activeChats {
			cp.prewarmSingleChat(chat)
		}
		return nil
	}

	// Use worker pool for concurrent processing
	numWorkers := 20
	if len(activeChats) < numWorkers {
		numWorkers = len(activeChats)
	}

	chatChan := make(chan Chat, len(activeChats))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chat := range chatChan {
				cp.prewarmSingleChat(chat)
			}
		}()
	}

	// Send chats to workers
	for _, chat := range activeChats {
		chatChan <- chat
	}
	close(chatChan)

	// Wait for all workers to complete
	wg.Wait()

	return nil
}

// prewarmSingleUser prewarns a single user's data into cache
func (cp *CachePrewarmer) prewarmSingleUser(user User) {
	// Cache user data
	userKey := userCacheKey(user.UserId)
	if err := cache.Marshal.Set(cache.Context, userKey, &user, store.WithExpiration(CacheTTLLanguage)); err != nil {
		log.WithFields(log.Fields{
			"user_id": user.UserId,
			"error":   err,
		}).Debug("[CachePrewarming] Failed to cache user")
		return
	}

	// Cache language if set
	if user.Language != "" {
		langKey := userLanguageCacheKey(user.UserId)
		if err := cache.Marshal.Set(cache.Context, langKey, user.Language, store.WithExpiration(CacheTTLLanguage)); err != nil {
			log.WithFields(log.Fields{
				"user_id": user.UserId,
				"error":   err,
			}).Debug("[CachePrewarming] Failed to cache user language")
		}
	}
}

// prewarmActiveUsers loads recently active users into cache using concurrent processing
func (cp *CachePrewarmer) prewarmActiveUsers() error {
	var activeUsers []User

	// Get users with activity in last 7 days
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	err := DB.Where("last_activity > ? OR last_activity IS NULL", cutoff).
		Limit(5000). // Limit to prevent memory issues
		Find(&activeUsers).Error

	if err != nil {
		return err
	}

	log.WithField("count", len(activeUsers)).Info("[CachePrewarming] Prewarming active users")

	// For small numbers, process sequentially
	if len(activeUsers) <= 20 {
		for _, user := range activeUsers {
			cp.prewarmSingleUser(user)
		}
		return nil
	}

	// Use worker pool for concurrent processing
	numWorkers := 30
	if len(activeUsers) < numWorkers*2 {
		numWorkers = len(activeUsers) / 2
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	userChan := make(chan User, len(activeUsers))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for user := range userChan {
				cp.prewarmSingleUser(user)
			}
		}()
	}

	// Send users to workers
	for _, user := range activeUsers {
		userChan <- user
	}
	close(userChan)

	// Wait for all workers to complete
	wg.Wait()

	return nil
}

// prewarmLanguageSettings preloads language settings for better i18n performance
func (cp *CachePrewarmer) prewarmLanguageSettings() error {
	// This could be expanded to preload common translations
	// For now, we just ensure the cache keys are properly set up
	log.Info("[CachePrewarming] Language settings prewarming completed")
	return nil
}

// PrewarmSpecificChat loads a specific chat's data into cache
// Useful for high-traffic chats that need immediate cache availability
func (cp *CachePrewarmer) PrewarmSpecificChat(chatID int64) error {
	if !cp.enabled || cache.Marshal == nil {
		return nil
	}

	var chat Chat
	err := DB.Where("chat_id = ?", chatID).First(&chat).Error
	if err != nil {
		return err
	}

	// Cache all related data for this chat
	chatKey := chatCacheKey(chat.ChatId)
	if err := cache.Marshal.Set(cache.Context, chatKey, &chat, store.WithExpiration(CacheTTLChatSettings)); err != nil {
		return err
	}

	settingsKey := chatSettingsCacheKey(chat.ChatId)
	if err := cache.Marshal.Set(cache.Context, settingsKey, &chat, store.WithExpiration(CacheTTLChatSettings)); err != nil {
		return err
	}

	if chat.Language != "" {
		langKey := chatLanguageCacheKey(chat.ChatId)
		if err := cache.Marshal.Set(cache.Context, langKey, chat.Language, store.WithExpiration(CacheTTLLanguage)); err != nil {
			return err
		}
	}

	log.WithField("chat_id", chatID).Debug("[CachePrewarming] Prewarmed specific chat")
	return nil
}

// PrewarmSpecificUser loads a specific user's data into cache
func (cp *CachePrewarmer) PrewarmSpecificUser(userID int64) error {
	if !cp.enabled || cache.Marshal == nil {
		return nil
	}

	var user User
	err := DB.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return err
	}

	// Cache user data
	userKey := userCacheKey(user.UserId)
	if err := cache.Marshal.Set(cache.Context, userKey, &user, store.WithExpiration(CacheTTLLanguage)); err != nil {
		return err
	}

	if user.Language != "" {
		langKey := userLanguageCacheKey(user.UserId)
		if err := cache.Marshal.Set(cache.Context, langKey, user.Language, store.WithExpiration(CacheTTLLanguage)); err != nil {
			return err
		}
	}

	log.WithField("user_id", userID).Debug("[CachePrewarming] Prewarmed specific user")
	return nil
}

// Global cache prewarmer instance
var CachePrewarmerInstance = NewCachePrewarmer()

// PrewarmCachesOnStartup is called during bot initialization
func PrewarmCachesOnStartup() error {
	return CachePrewarmerInstance.PrewarmCaches()
}

// PrewarmChat preloads a specific chat's data
func PrewarmChat(chatID int64) error {
	return CachePrewarmerInstance.PrewarmSpecificChat(chatID)
}

// PrewarmUser preloads a specific user's data
func PrewarmUser(userID int64) error {
	return CachePrewarmerInstance.PrewarmSpecificUser(userID)
}
