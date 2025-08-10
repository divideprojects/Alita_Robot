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
	CacheTTLChatSettings = 30 * time.Minute
	CacheTTLLanguage     = 1 * time.Hour
	CacheTTLFilterList   = 30 * time.Minute
	CacheTTLBlacklist    = 30 * time.Minute
	CacheTTLGreetings    = 30 * time.Minute
	CacheTTLNotesList    = 30 * time.Minute
	CacheTTLWarnSettings = 30 * time.Minute
	CacheTTLAntiflood    = 30 * time.Minute
	CacheTTLDisabledCmds = 30 * time.Minute
)

// Singleflight group for preventing cache stampede
var (
	cacheGroup singleflight.Group
)

// Cache key generators with "alita:" prefix for better organization
// chatSettingsCacheKey generates a cache key for chat settings.
func chatSettingsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:chat_settings:%d", chatID)
}

// userLanguageCacheKey generates a cache key for user language settings.
func userLanguageCacheKey(userID int64) string {
	return fmt.Sprintf("alita:user_lang:%d", userID)
}

// chatLanguageCacheKey generates a cache key for chat language settings.
func chatLanguageCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:chat_lang:%d", chatID)
}

// filterListCacheKey generates a cache key for chat filter lists.
func filterListCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:filter_list:%d", chatID)
}

// blacklistCacheKey generates a cache key for chat blacklist settings.
func blacklistCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:blacklist:%d", chatID)
}

// greetingsCacheKey generates a cache key for chat greeting settings.
func greetingsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:greetings:%d", chatID)
}

// notesListCacheKey generates a cache key for chat notes lists.
// The admin parameter distinguishes between admin and regular note lists.
func notesListCacheKey(chatID int64, admin bool) string {
	return fmt.Sprintf("alita:notes_list:%d:%v", chatID, admin)
}

// warnSettingsCacheKey generates a cache key for chat warning settings.
func warnSettingsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:warn_settings:%d", chatID)
}

// antifloodCacheKey generates a cache key for chat antiflood settings.
func antifloodCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:antiflood:%d", chatID)
}

// disabledCommandsCacheKey generates a cache key for chat disabled commands.
func disabledCommandsCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:disabled_cmds:%d", chatID)
}

// InvalidateChatCache invalidates all cache entries for a chat.
// Removes all cached data related to the specified chat ID including settings, filters, and notes.
func InvalidateChatCache(chatID int64) {
	if cache.Marshal == nil {
		return
	}

	keys := []string{
		chatCacheKey(chatID),
		chatSettingsCacheKey(chatID),
		chatLanguageCacheKey(chatID),
		filterListCacheKey(chatID),
		blacklistCacheKey(chatID),
		greetingsCacheKey(chatID),
		notesListCacheKey(chatID, true),
		notesListCacheKey(chatID, false),
		warnSettingsCacheKey(chatID),
		antifloodCacheKey(chatID),
		disabledCommandsCacheKey(chatID),
	}

	// Invalidate cache keys sequentially
	if cache.Marshal != nil {
		for _, key := range keys {
			if err := cache.Marshal.Delete(cache.Context, key); err != nil {
				log.Debugf("[Cache] Failed to invalidate key %s: %v", key, err)
			}
		}
	}
}

// InvalidateUserCache invalidates all cache entries for a user.
// Currently only removes user language cache entries.
func InvalidateUserCache(userID int64) {
	if cache.Marshal == nil {
		return
	}

	err := cache.Marshal.Delete(cache.Context, userLanguageCacheKey(userID))
	if err != nil {
		log.Debugf("[Cache] Failed to invalidate user cache %d: %v", userID, err)
	}
}

// getFromCacheOrLoad is a generic helper to get from cache or load from database with stampede protection.
// Uses singleflight pattern to prevent cache stampede when multiple goroutines request the same data.
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

	// Type assertion failed - return error immediately
	// Don't return the wrong type which could cause data corruption
	var zero T
	return zero, fmt.Errorf("type assertion failed for cache key %s", key)
}

// deleteCache is a helper to delete a value from cache.
// Logs debug information if deletion fails but does not return errors.
func deleteCache(key string) {
	if cache.Marshal == nil {
		return
	}

	err := cache.Marshal.Delete(cache.Context, key)
	if err != nil {
		log.Debugf("[Cache] Failed to delete cache for key %s: %v", key, err)
	}
}
