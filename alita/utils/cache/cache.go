package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var (
	Context = context.Background()
	Marshal *marshaler.Marshaler
	Manager *cache.ChainCache[any]

	// Cache state tracking
	isCacheEnabled bool
	cacheMutex     sync.RWMutex

	// Error types for cache operations
	ErrCacheNotEnabled  = errors.New("cache is not enabled")
	ErrCacheUnavailable = errors.New("cache is temporarily unavailable")
)

/*
AdminCache represents the cached administrator information for a chat.

Fields:
  - ChatId:   The unique identifier for the chat.
  - UserInfo: A slice of merged chat member information for each admin.
  - Cached:   Indicates if the cache is valid and populated.
*/
type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	UserMap  map[int64]gotgbot.MergedChatMember
	Cached   bool
	mux      sync.RWMutex
}

/*
InitCache initializes the caching system for the application.

It sets up both Redis and Ristretto as cache backends, creates a chain cache manager,
and initializes the marshaler for serializing and deserializing cached data.
Implements graceful degradation if cache initialization fails.
*/
func InitCache() {
	err := InitCacheWithFallback()
	if err != nil {
		log.WithError(err).Error("Failed to initialize cache system")
		// Cache will be disabled, but application continues
	}
}

/*
InitCacheWithFallback initializes the caching system with fallback mechanisms.

Returns error if initialization fails, but allows application to continue
without caching functionality.
*/
func InitCacheWithFallback() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Try to initialize Ristretto cache
	ristrettoCache, err := initRistrettoCache()
	if err != nil {
		log.WithError(err).Error("Failed to initialize Ristretto cache")
		return err
	}

	// Try to initialize Redis cache
	redisClient, err := initRedisCache()
	if err != nil {
		log.WithError(err).Warn("Failed to initialize Redis cache, falling back to Ristretto only")
		// Use only Ristretto cache
		ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
		cacheManager := cache.New[any](ristrettoStore)
		Marshal = marshaler.New(cacheManager)
		isCacheEnabled = true
		return nil
	}

	// Both caches available, use chain cache
	redisStore := redis_store.NewRedis(redisClient, store.WithExpiration(10*time.Minute))
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.NewChain(cache.New[any](ristrettoStore), cache.New[any](redisStore))

	// Assign global variables
	Manager = cacheManager
	Marshal = marshaler.New(cacheManager)
	isCacheEnabled = true

	log.Info("Cache system initialized successfully with Redis and Ristretto")
	return nil
}

/*
initRistrettoCache initializes the Ristretto in-memory cache.

Returns the cache instance or error if initialization fails.
*/
func initRistrettoCache() (*ristretto.Cache, error) {
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	log.Debug("Ristretto cache initialized successfully")
	return ristrettoCache, nil
}

/*
initRedisCache initializes the Redis cache client.

Returns the Redis client or error if initialization fails.
*/
func initRedisCache() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		return nil, err
	}

	log.Debug("Redis cache initialized successfully")
	return redisClient, nil
}

/*
IsCacheEnabled returns whether the cache system is currently enabled.
*/
func IsCacheEnabled() bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return isCacheEnabled
}

/*
DisableCache disables the cache system gracefully.
*/
func DisableCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	isCacheEnabled = false
	log.Warn("Cache system has been disabled")
}

/*
HealthCheckCache performs health checks on the cache system.

Returns error if cache is unhealthy or unavailable.
*/
func HealthCheckCache() error {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if !isCacheEnabled {
		return ErrCacheNotEnabled
	}

	if Marshal == nil {
		return ErrCacheUnavailable
	}

	// Test basic cache operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testKey := "health_check_test"
	testValue := "health_check_value"

	// Test set operation
	if err := Marshal.Set(ctx, testKey, testValue, store.WithExpiration(10*time.Second)); err != nil {
		return err
	}

	// Test get operation
	if _, err := Marshal.Get(ctx, testKey, new(string)); err != nil {
		return err
	}

	// Test delete operation
	if err := Marshal.Delete(ctx, testKey); err != nil {
		return err
	}

	return nil
}

/*
PurgeCache clears all cached data from both Redis and Ristretto backends.
Returns error if any operation fails.
*/
func PurgeCache() error {
	if !IsCacheEnabled() {
		return ErrCacheNotEnabled
	}

	if Marshal == nil {
		return ErrCacheUnavailable
	}

	// Clear all cached items
	err := Marshal.Clear(Context)
	if err != nil {
		log.WithError(err).Error("Failed to purge cache")
		return err
	}

	log.Info("Cache purged successfully")
	return nil
}

/*
SafeCacheOperation performs a cache operation with error handling.

Returns error if cache is not available or operation fails.
*/
func SafeCacheOperation(operation func() error) error {
	if !IsCacheEnabled() {
		return ErrCacheNotEnabled
	}

	if Marshal == nil {
		return ErrCacheUnavailable
	}

	return operation()
}
