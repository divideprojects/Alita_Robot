package cache

import (
	"context"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
)

// LoadAdminCache loads the admin cache for the chat.
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

	// Cache the admin list with retry on failure in background
	go func() {
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			if err := Marshal.Set(Context, AdminCache{ChatId: chatId}, adminCache, store.WithExpiration(time.Minute*30)); err != nil {
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

	return adminCache
}

// GetAdminCacheList gets the admin cache for the chat.
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
	return true, *gotAdminlist.(*AdminCache)
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
