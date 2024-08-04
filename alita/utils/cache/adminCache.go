package cache

import (
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

	adminList, err := b.GetChatAdministrators(chatId, nil)
	if err != nil {
		log.Error(err)
		return AdminCache{}
	}

	var userList []gotgbot.MergedChatMember
	for _, admin := range adminList {
		userList = append(userList, admin.MergeChatMember())
	}

	err = Marshal.Set(
		Context,
		AdminCache{
			ChatId: chatId,
		},
		AdminCache{
			ChatId:   chatId,
			UserInfo: userList,
			Cached:   true,
		},
		store.WithExpiration(10*time.Minute),
	)
	if err != nil {
		log.Error(err)
		return AdminCache{}
	}

	_, newAdminList := GetAdminCacheList(chatId)
	return newAdminList
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
