package cache

import (
	"context"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	redis_store "github.com/eko/gocache/store/redis/v4"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
)

var (
	Context = context.Background()
	Marshal *marshaler.Marshaler
	Manager *cache.ChainCache[any]
)

type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	Cached   bool
}

// InitCache initializes the dual-layer cache system with Ristretto (L1) and Redis (L2).
// It establishes connections to both cache stores and returns an error if initialization fails.
func InitCache() error {
	// Test Redis connection first
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword, // no password set
		DB:       config.RedisDB,       // use default DB
	})

	// Test Redis connection
	if err := redisClient.Ping(Context).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Initialize Ristretto cache with proper error handling
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Ristretto cache: %w", err)
	}

	// initialize cache manager
	redisStore := redis_store.NewRedis(redisClient)
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.NewChain(cache.New[any](ristrettoStore), cache.New[any](redisStore))

	// Initializes marshaler
	Marshal = marshaler.New(cacheManager)
	Manager = cacheManager

	return nil
}
