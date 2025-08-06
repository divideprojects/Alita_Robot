package db

import (
	"fmt"
	"time"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
)

const (
	// Cache expiration times
	CacheTTLChatSettings    = 30 * time.Minute
	CacheTTLLanguage        = 1 * time.Hour
	CacheTTLFilterList      = 30 * time.Minute
	CacheTTLBlacklist       = 30 * time.Minute
	CacheTTLGreetings       = 30 * time.Minute
	CacheTTLNotesList       = 30 * time.Minute
	CacheTTLWarnSettings    = 30 * time.Minute
	CacheTTLAntiflood       = 30 * time.Minute
)

// Singleflight group for preventing cache stampede
var (
	cacheGroup singleflight.Group
)

// Cache key generators with "alita:" prefix for better organization
func chatSettingsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:chat_settings:%d", chatID)
}

func userLanguageCacheKey(userID int64) string {
	return fmt.Sprintf("alita:user_lang:%d", userID)
}

func chatLanguageCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:chat_lang:%d", chatID)
}

func filterListCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:filter_list:%d", chatID)
}

func blacklistCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:blacklist:%d", chatID)
}

func greetingsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:greetings:%d", chatID)
}

func notesListCacheKey(chatID int64, admin bool) string {
	return fmt.Sprintf("alita:notes_list:%d:%v", chatID, admin)
}

func warnSettingsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:warn_settings:%d", chatID)
}

func antifloodCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:antiflood:%d", chatID)
}

// InvalidateChatCache invalidates all cache entries for a chat
func InvalidateChatCache(chatID int64) {
	if cache.Marshal == nil {
		return
	}

	keys := []string{
		chatSettingsCacheKey(chatID),
		chatLanguageCacheKey(chatID),
		filterListCacheKey(chatID),
		blacklistCacheKey(chatID),
		greetingsCacheKey(chatID),
		notesListCacheKey(chatID, true),
		notesListCacheKey(chatID, false),
		warnSettingsCacheKey(chatID),
		antifloodCacheKey(chatID),
	}

	for _, key := range keys {
		err := cache.Marshal.Delete(cache.Context, key)
		if err != nil {
			log.Debugf("[Cache] Failed to invalidate key %s: %v", key, err)
		}
	}
}

// InvalidateUserCache invalidates all cache entries for a user
func InvalidateUserCache(userID int64) {
	if cache.Marshal == nil {
		return
	}

	err := cache.Marshal.Delete(cache.Context, userLanguageCacheKey(userID))
	if err != nil {
		log.Debugf("[Cache] Failed to invalidate user cache %d: %v", userID, err)
	}
}

// getFromCacheOrLoad is a generic helper to get from cache or load from database with stampede protection
func getFromCacheOrLoad[T any](key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var result T
	
	if cache.Marshal == nil {
		// Cache not initialized, load directly
		return loader()
	}

	// Try to get from cache
	_, err := cache.Marshal.Get(cache.Context, key, &result)
	if err == nil {
		// Cache hit
		return result, nil
	}

	// Cache miss, use singleflight to prevent stampede
	v, err, _ := cacheGroup.Do(key, func() (interface{}, error) {
		// Load from database
		data, loadErr := loader()
		if loadErr != nil {
			return data, loadErr
		}

		// Store in cache
		cacheErr := cache.Marshal.Set(cache.Context, key, data, store.WithExpiration(ttl))
		if cacheErr != nil {
			log.Debugf("[Cache] Failed to set cache for key %s: %v", key, cacheErr)
		}

		return data, nil
	})

	if err != nil {
		return result, err
	}

	// Type assert the result
	if typedResult, ok := v.(T); ok {
		return typedResult, nil
	}

	// Fallback if type assertion fails
	return result, fmt.Errorf("type assertion failed for cache key %s", key)
}

// deleteCache is a helper to delete a value from cache
func deleteCache(key string) {
	if cache.Marshal == nil {
		return
	}

	err := cache.Marshal.Delete(cache.Context, key)
	if err != nil {
		log.Debugf("[Cache] Failed to delete cache for key %s: %v", key, err)
	}
}