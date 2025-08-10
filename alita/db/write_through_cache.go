package db

import (
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
)

// WriteThroughCache provides write-through caching functionality
// This ensures cache consistency by updating cache immediately when database is modified
type WriteThroughCache struct {
	enabled bool
}

// NewWriteThroughCache creates a new write-through cache instance
func NewWriteThroughCache() *WriteThroughCache {
	return &WriteThroughCache{
		enabled: config.EnableWriteThroughCache,
	}
}

// UpdateUser updates user in database and cache simultaneously
func (wtc *WriteThroughCache) UpdateUser(user *User) error {
	// Update database first
	err := UpdateRecord(&User{}, User{UserId: user.UserId}, user)
	if err != nil {
		return err
	}

	// Update cache if write-through is enabled
	if wtc.enabled && cache.Marshal != nil {
		cacheKey := userCacheKey(user.UserId)
		cacheErr := cache.Marshal.Set(cache.Context, cacheKey, user, store.WithExpiration(CacheTTLLanguage))
		if cacheErr != nil {
			log.WithFields(log.Fields{
				"user_id": user.UserId,
				"error":   cacheErr,
			}).Debug("[WriteThrough] Failed to update user cache")
		}

		// Also update language cache
		if user.Language != "" {
			langCacheKey := userLanguageCacheKey(user.UserId)
			langCacheErr := cache.Marshal.Set(cache.Context, langCacheKey, user.Language, store.WithExpiration(CacheTTLLanguage))
			if langCacheErr != nil {
				log.WithFields(log.Fields{
					"user_id": user.UserId,
					"error":   langCacheErr,
				}).Debug("[WriteThrough] Failed to update user language cache")
			}
		}
	}

	return nil
}

// UpdateChat updates chat in database and cache simultaneously
func (wtc *WriteThroughCache) UpdateChat(chat *Chat) error {
	// Update database first
	err := UpdateRecord(&Chat{}, Chat{ChatId: chat.ChatId}, chat)
	if err != nil {
		return err
	}

	// Update cache if write-through is enabled
	if wtc.enabled && cache.Marshal != nil {
		cacheKey := chatCacheKey(chat.ChatId)
		cacheErr := cache.Marshal.Set(cache.Context, cacheKey, chat, store.WithExpiration(CacheTTLChatSettings))
		if cacheErr != nil {
			log.WithFields(log.Fields{
				"chat_id": chat.ChatId,
				"error":   cacheErr,
			}).Debug("[WriteThrough] Failed to update chat cache")
		}

		// Also update language cache
		if chat.Language != "" {
			langCacheKey := chatLanguageCacheKey(chat.ChatId)
			langCacheErr := cache.Marshal.Set(cache.Context, langCacheKey, chat.Language, store.WithExpiration(CacheTTLLanguage))
			if langCacheErr != nil {
				log.WithFields(log.Fields{
					"chat_id": chat.ChatId,
					"error":   langCacheErr,
				}).Debug("[WriteThrough] Failed to update chat language cache")
			}
		}

		// Update chat settings cache
		settingsCacheKey := chatSettingsCacheKey(chat.ChatId)
		settingsCacheErr := cache.Marshal.Set(cache.Context, settingsCacheKey, chat, store.WithExpiration(CacheTTLChatSettings))
		if settingsCacheErr != nil {
			log.WithFields(log.Fields{
				"chat_id": chat.ChatId,
				"error":   settingsCacheErr,
			}).Debug("[WriteThrough] Failed to update chat settings cache")
		}
	}

	return nil
}

// CreateUser creates user in database and cache simultaneously
func (wtc *WriteThroughCache) CreateUser(user *User) error {
	// Create in database first
	err := CreateRecord(user)
	if err != nil {
		return err
	}

	// Add to cache if write-through is enabled
	if wtc.enabled && cache.Marshal != nil {
		cacheKey := userCacheKey(user.UserId)
		cacheErr := cache.Marshal.Set(cache.Context, cacheKey, user, store.WithExpiration(CacheTTLLanguage))
		if cacheErr != nil {
			log.WithFields(log.Fields{
				"user_id": user.UserId,
				"error":   cacheErr,
			}).Debug("[WriteThrough] Failed to cache new user")
		}

		// Cache language if set
		if user.Language != "" {
			langCacheKey := userLanguageCacheKey(user.UserId)
			langCacheErr := cache.Marshal.Set(cache.Context, langCacheKey, user.Language, store.WithExpiration(CacheTTLLanguage))
			if langCacheErr != nil {
				log.WithFields(log.Fields{
					"user_id": user.UserId,
					"error":   langCacheErr,
				}).Debug("[WriteThrough] Failed to cache new user language")
			}
		}
	}

	return nil
}

// CreateChat creates chat in database and cache simultaneously
func (wtc *WriteThroughCache) CreateChat(chat *Chat) error {
	// Create in database first
	err := CreateRecord(chat)
	if err != nil {
		return err
	}

	// Add to cache if write-through is enabled
	if wtc.enabled && cache.Marshal != nil {
		cacheKey := chatCacheKey(chat.ChatId)
		cacheErr := cache.Marshal.Set(cache.Context, cacheKey, chat, store.WithExpiration(CacheTTLChatSettings))
		if cacheErr != nil {
			log.WithFields(log.Fields{
				"chat_id": chat.ChatId,
				"error":   cacheErr,
			}).Debug("[WriteThrough] Failed to cache new chat")
		}

		// Cache language if set
		if chat.Language != "" {
			langCacheKey := chatLanguageCacheKey(chat.ChatId)
			langCacheErr := cache.Marshal.Set(cache.Context, langCacheKey, chat.Language, store.WithExpiration(CacheTTLLanguage))
			if langCacheErr != nil {
				log.WithFields(log.Fields{
					"chat_id": chat.ChatId,
					"error":   langCacheErr,
				}).Debug("[WriteThrough] Failed to cache new chat language")
			}
		}

		// Cache settings
		settingsCacheKey := chatSettingsCacheKey(chat.ChatId)
		settingsCacheErr := cache.Marshal.Set(cache.Context, settingsCacheKey, chat, store.WithExpiration(CacheTTLChatSettings))
		if settingsCacheErr != nil {
			log.WithFields(log.Fields{
				"chat_id": chat.ChatId,
				"error":   settingsCacheErr,
			}).Debug("[WriteThrough] Failed to cache new chat settings")
		}
	}

	return nil
}

// DeleteFromCache removes entries from cache when data is deleted
func (wtc *WriteThroughCache) DeleteFromCache(keys ...string) {
	if !wtc.enabled || cache.Marshal == nil {
		return
	}

	for _, key := range keys {
		err := cache.Marshal.Delete(cache.Context, key)
		if err != nil {
			log.WithFields(log.Fields{
				"key":   key,
				"error": err,
			}).Debug("[WriteThrough] Failed to delete from cache")
		}
	}
}

// Global write-through cache instance
var WriteThroughCacheInstance = NewWriteThroughCache()

// Helper functions that use write-through cache

// UpdateUserWithCache updates user using write-through cache
func UpdateUserWithCache(user *User) error {
	return WriteThroughCacheInstance.UpdateUser(user)
}

// UpdateChatWithCache updates chat using write-through cache
func UpdateChatWithCache(chat *Chat) error {
	return WriteThroughCacheInstance.UpdateChat(chat)
}

// CreateUserWithCache creates user using write-through cache
func CreateUserWithCache(user *User) error {
	return WriteThroughCacheInstance.CreateUser(user)
}

// CreateChatWithCache creates chat using write-through cache
func CreateChatWithCache(chat *Chat) error {
	return WriteThroughCacheInstance.CreateChat(chat)
}

// Note: userCacheKey and chatCacheKey functions are defined in optimized_queries.go
