package cache

import (
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
)

// LoadAdminCache loads the admin cache for the chat.
func LoadAdminCache(b *gotgbot.Bot, chat *gotgbot.Chat) AdminCache {
	adminList, err := chat.GetAdministrators(b, nil)
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
			ChatId: chat.Id,
		},
		AdminCache{
			ChatId:   chat.Id,
			UserInfo: userList,
			Cached:   true,
		},
		store.WithExpiration(10*time.Minute),
	)
	if err != nil {
		log.Error(err)
		return AdminCache{}
	}

	_, newAdminList := GetAdminCacheList(chat.Id)
	return newAdminList
}

// GetAdminCacheList gets the admin cache for the chat.
func GetAdminCacheList(chatId int64) (bool, AdminCache) {
	gotAdminlist, _ := Marshal.Get(
		Context,
		AdminCache{
			ChatId: chatId,
		},
		new(AdminCache),
	)
	if gotAdminlist == nil {
		return false, AdminCache{}
	}
	return true, *gotAdminlist.(*AdminCache)
}
