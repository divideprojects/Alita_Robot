package cache

import (
	"context"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/marshaler"
	"github.com/eko/gocache/v3/store"
	gocache "github.com/patrickmn/go-cache"
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

// InitCache initializes the cache.
func InitCache() {
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := store.NewGoCache(gocacheClient)

	// Initialize chained cache
	Manager = cache.NewChain[any](cache.New[any](gocacheStore))

	// Initializes marshaler
	Marshal = marshaler.New(Manager)
}

// GetAdminCacheUser gets the admin cache for the chat.
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
