package cache

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/marshaler"
	"github.com/eko/gocache/v3/store"
	"github.com/go-redis/redis/v8"

	"github.com/Divkix/Alita_Robot/alita/config"
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

func InitCache() {
	// Initialize gocache cache and Redis client
	// gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     config.RedisUri,
			Password: config.RedisPassword,
		},
	)

	// Initialize stores
	// gocacheStore := store.NewGoCache(gocacheClient, nil)
	redisStore := store.NewRedis(redisClient)

	// Initialize chained cache
	Manager = cache.NewChain[any](cache.New[any](redisStore))

	// Initializes marshaler
	Marshal = marshaler.New(Manager)
}

func GetAdminCacheUser(chatId, userId int64) (bool, gotgbot.MergedChatMember) {
	adminList, _ := Marshal.Get(Context, AdminCache{ChatId: chatId}, new(AdminCache))
	for i := range adminList.(*AdminCache).UserInfo {
		admin := &adminList.(*AdminCache).UserInfo[i]
		if admin.User.Id == userId {
			return true, *admin
		}
	}
	return false, gotgbot.MergedChatMember{}
}
