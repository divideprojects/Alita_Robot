package cache

import (
	"context"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/redis/go-redis/v9"
)

var (
	Context = context.Background()
	Marshal *marshaler.Marshaler
	Manager *cache.Cache[any]
)

type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	Cached   bool
}

// InitCache initializes the Redis-only cache system.
// It establishes connection to Redis and returns an error if initialization fails.
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

	// Initialize cache manager with Redis only
	redisStore := redis_store.NewRedis(redisClient)
	cacheManager := cache.New[any](redisStore)

	// Initializes marshaler
	Marshal = marshaler.New(cacheManager)
	Manager = cacheManager

	return nil
}
