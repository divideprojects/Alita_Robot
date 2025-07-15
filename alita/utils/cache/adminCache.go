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
and stores the result in the cache with retries on failure. If the bot is nil, or if no administrators are found,
an empty AdminCache is returned. The caching operation is performed asynchronously.
*/
func LoadAdminCache(b *gotgbot.Bot, chatId int64) *AdminCache {
	if b == nil {
		log.Error("LoadAdminCache: bot is nil")
		return &AdminCache{}
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
		return &AdminCache{}
	}

	if len(adminList) == 0 {
		log.WithFields(log.Fields{
			"chatId": chatId,
		}).Warning("LoadAdminCache: No administrators found")
		return &AdminCache{}
	}

	// Convert ChatMember to MergedChatMember and build map
	var userList []gotgbot.MergedChatMember
	userMap := make(map[int64]gotgbot.MergedChatMember, len(adminList))
	for _, admin := range adminList {
		member := admin.MergeChatMember()
		userList = append(userList, member)
		userMap[member.User.Id] = member
	}

	adminCache := AdminCache{
		ChatId:   chatId,
		UserInfo: userList,
		UserMap:  userMap,
		Cached:   true,
	}

	// Cache the admin list with retry on failure in background
	go func() {
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			err := SafeCacheOperation(func() error {
				return Marshal.Set(Context, AdminCache{ChatId: chatId}, &adminCache, store.WithExpiration(time.Minute*10))
			})

			if err != nil {
				log.WithFields(log.Fields{
					"chatId": chatId,
					"error":  err,
					"retry":  i + 1,
				}).Error("LoadAdminCache: Failed to cache admin list")

				if i < maxRetries-1 {
					time.Sleep(time.Second * 2) // Wait before retry
					continue
				}
			} else {
				log.WithFields(log.Fields{
					"chatId": chatId,
				}).Debug("LoadAdminCache: Successfully cached admin list")
				break
			}
		}
	}()

	return &adminCache
}

/*
GetAdminCacheList retrieves the admin cache for the specified chat ID.

Returns a boolean indicating if the cache was found, and the AdminCache object.
If the cache is not found or an error occurs, returns false and an empty AdminCache.
*/
func GetAdminCacheList(chatId int64) (bool, *AdminCache) {
	if !IsCacheEnabled() {
		return false, &AdminCache{}
	}

	var gotAdminlist interface{}
	err := SafeCacheOperation(func() error {
		var err error
		gotAdminlist, err = Marshal.Get(
			Context,
			AdminCache{
				ChatId: chatId,
			},
			new(AdminCache),
		)
		return err
	})
	if err != nil {
		log.WithFields(log.Fields{
			"chatId": chatId,
			"error":  err,
		}).Error("GetAdminCacheList: Failed to get admin cache")
		return false, &AdminCache{}
	}
	if gotAdminlist == nil {
		return false, &AdminCache{}
	}
	return true, gotAdminlist.(*AdminCache)
}

/*
GetAdminCacheUser retrieves the cached admin user for the given chat and user ID.

Returns true and the MergedChatMember if found, otherwise returns false and an empty MergedChatMember.
Uses optimized map-based lookup for O(1) performance.
*/
func GetAdminCacheUser(chatId, userId int64) (bool, gotgbot.MergedChatMember) {
	if !IsCacheEnabled() {
		return false, gotgbot.MergedChatMember{}
	}

	var adminList interface{}
	err := SafeCacheOperation(func() error {
		var err error
		adminList, err = Marshal.Get(Context, AdminCache{ChatId: chatId}, new(AdminCache))
		return err
	})

	if err != nil || adminList == nil {
		return false, gotgbot.MergedChatMember{}
	}

	adminCache := adminList.(*AdminCache)
	adminCache.mux.RLock()
	defer adminCache.mux.RUnlock()

	if member, ok := adminCache.UserMap[userId]; ok {
		return true, member
	}
	return false, gotgbot.MergedChatMember{}
}

/*
GetAdmins retrieves the admin list for the specified chat, using cache if available.

This function abstracts the caching logic by checking the cache first and loading fresh data if needed.
Returns the list of admin users and a boolean indicating whether the data came from cache.
*/
func GetAdmins(b *gotgbot.Bot, chatId int64) ([]gotgbot.MergedChatMember, bool) {
	if adminsAvail, admins := GetAdminCacheList(chatId); adminsAvail {
		return admins.UserInfo, true
	}

	admins := LoadAdminCache(b, chatId)
	return admins.UserInfo, false
}

/*
InvalidateAdminCache removes the admin cache for the specified chat ID.

This function should be called when the admin list is known to have changed,
ensuring that subsequent calls will fetch fresh data from Telegram.
*/
func InvalidateAdminCache(chatId int64) error {
	err := SafeCacheOperation(func() error {
		return Marshal.Delete(Context, AdminCache{ChatId: chatId})
	})
	if err != nil {
		log.WithFields(log.Fields{
			"chatId": chatId,
			"error":  err,
		}).Error("InvalidateAdminCache: Failed to delete admin cache")
		return err
	}

	log.WithFields(log.Fields{
		"chatId": chatId,
	}).Debug("InvalidateAdminCache: Successfully invalidated admin cache")
	return nil
}

/*
GetAdminIds retrieves a list of admin user IDs for the specified chat.

Returns a slice of int64 user IDs for all administrators in the chat.
Uses cache if available, otherwise loads fresh data.
*/
func GetAdminIds(b *gotgbot.Bot, chatId int64) []int64 {
	admins, _ := GetAdmins(b, chatId)
	adminIds := make([]int64, len(admins))
	for i, admin := range admins {
		adminIds[i] = admin.User.Id
	}
	return adminIds
}

/*
IsUserAdminCached checks if a user is an admin using cached data with optimized lookup.

Returns true if the user is an admin, false otherwise.
Uses map-based lookup for O(1) performance when multiple checks are needed.
*/
func IsUserAdminCached(_ *gotgbot.Bot, chatId, userId int64) bool {
	if !IsCacheEnabled() {
		return false
	}

	var adminList interface{}
	err := SafeCacheOperation(func() error {
		var err error
		adminList, err = Marshal.Get(Context, AdminCache{ChatId: chatId}, new(AdminCache))
		return err
	})

	if err != nil || adminList == nil {
		return false
	}

	adminCache := adminList.(*AdminCache)
	adminCache.mux.RLock()
	defer adminCache.mux.RUnlock()

	_, ok := adminCache.UserMap[userId]
	return ok
}
