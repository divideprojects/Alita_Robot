package cache

import (
	"context"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
)

/*
LoadAdminCache loads and returns the admin cache for the specified chat.

It fetches the list of chat administrators using the provided bot instance, converts them to MergedChatMember,
and stores the result in the cache with a short timeout. If the bot is nil, or if no administrators are found,
an empty AdminCache is returned. The caching operation is performed synchronously for better reliability.
*/
func LoadAdminCache(b *gotgbot.Bot, chatId int64) AdminCache {
	if b == nil {
		log.Error("LoadAdminCache: bot is nil")
		return AdminCache{}
	}

	// Create context with timeout to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adminList, err := b.GetChatAdministratorsWithContext(ctx, chatId, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"chatId": chatId,
			"error":  err,
		}).Error("LoadAdminCache: Failed to get chat administrators")
		return AdminCache{}
	}

	if len(adminList) == 0 {
		log.WithFields(log.Fields{
			"chatId": chatId,
		}).Warning("LoadAdminCache: No administrators found")
		return AdminCache{}
	}

	// Convert ChatMember to MergedChatMember
	var userList []gotgbot.MergedChatMember
	for _, admin := range adminList {
		userList = append(userList, admin.MergeChatMember())
	}

	adminCache := AdminCache{
		ChatId:   chatId,
		UserInfo: userList,
		Cached:   true,
	}

	// Build the user map for O(1) lookups
	adminCache.buildUserMap()

	// Cache the admin list synchronously with shorter timeout
	cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cacheCancel()

	// Try to cache with limited retries
	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		if err := Marshal.Set(cacheCtx, AdminCache{ChatId: chatId}, adminCache, store.WithExpiration(time.Minute*30)); err != nil {
			log.WithFields(log.Fields{
				"chatId": chatId,
				"error":  err,
				"retry":  i + 1,
			}).Warning("LoadAdminCache: Failed to cache admin list")

			if i < maxRetries-1 {
				time.Sleep(time.Millisecond * 500) // Brief wait before retry
				continue
			}
		} else {
			log.WithFields(log.Fields{
				"chatId": chatId,
			}).Debug("LoadAdminCache: Successfully cached admin list")
			break
		}
	}

	return adminCache
}

/*
GetAdminCacheList retrieves the admin cache for the specified chat ID.

Returns a boolean indicating if the cache was found, and the AdminCache object.
If the cache is not found or an error occurs, returns false and an empty AdminCache.
*/
func GetAdminCacheList(chatId int64) (bool, AdminCache) {
	gotAdminlist, err := Marshal.Get(
		Context,
		AdminCache{
			ChatId: chatId,
		},
		new(AdminCache),
	)
	if err != nil {
		log.Error(err)
		return false, AdminCache{}
	}
	if gotAdminlist == nil {
		return false, AdminCache{}
	}

	adminCache := *gotAdminlist.(*AdminCache)
	// Ensure user map is built after unmarshaling
	if adminCache.userMap == nil && len(adminCache.UserInfo) > 0 {
		adminCache.buildUserMap()
	}

	return true, adminCache
}

/*
GetAdminCacheUser retrieves the cached admin user for the given chat and user ID.

Returns true and the MergedChatMember if found, otherwise returns false and an empty MergedChatMember.
Uses O(1) map lookup for improved performance.
*/
func GetAdminCacheUser(chatId, userId int64) (bool, gotgbot.MergedChatMember) {
	found, adminCache := GetAdminCacheList(chatId)
	if !found {
		return false, gotgbot.MergedChatMember{}
	}

	user, userFound := adminCache.GetUser(userId)
	return userFound, user
}
