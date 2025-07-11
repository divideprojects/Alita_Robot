package cache

import (
	"context"

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

/*
AdminCache represents the cached administrator information for a chat.

Fields:
  - ChatId:   The unique identifier for the chat.
  - UserInfo: A slice of merged chat member information for each admin.
  - userMap:  Internal map for O(1) user lookups (not marshaled).
  - Cached:   Indicates if the cache is valid and populated.
*/
type AdminCache struct {
	ChatId   int64                               `json:"chat_id"`
	UserInfo []gotgbot.MergedChatMember          `json:"user_info"`
	userMap  map[int64]gotgbot.MergedChatMember  `json:"-"` // not marshaled
	Cached   bool                                `json:"cached"`
}

// buildUserMap creates the internal map for fast user lookups
func (ac *AdminCache) buildUserMap() {
	ac.userMap = make(map[int64]gotgbot.MergedChatMember, len(ac.UserInfo))
	for _, user := range ac.UserInfo {
		ac.userMap[user.User.Id] = user
	}
}

// GetUser returns the user from cache with O(1) lookup
func (ac *AdminCache) GetUser(userId int64) (gotgbot.MergedChatMember, bool) {
	if ac.userMap == nil {
		ac.buildUserMap()
	}
	user, found := ac.userMap[userId]
	return user, found
}

/*
InitCache initializes the caching system for the application.

It sets up both Redis and Ristretto as cache backends, creates a chain cache manager,
and initializes the marshaler for serializing and deserializing cached data.
Panics if Ristretto cache initialization fails.
*/
func InitCache() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword, // no password set
		DB:       config.RedisDB,       // use default DB
	})
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,        // 10x expected keys for better performance
		MaxCost:     100 * 1024,   // 100KB instead of 100B
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	// initialize cache manager
	redisStore := redis_store.NewRedis(redisClient)
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.NewChain(cache.New[any](ristrettoStore), cache.New[any](redisStore))

	// Initializes marshaler
	Marshal = marshaler.New(cacheManager)
}
