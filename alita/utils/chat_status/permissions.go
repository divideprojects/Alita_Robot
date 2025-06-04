package chat_status

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// NoPermissions represents a chat permission set with all actions disabled.
var NoPermissions = gotgbot.ChatPermissions{
	CanSendMessages:       false,
	CanSendPhotos:         false,
	CanSendVideos:         false,
	CanSendAudios:         false,
	CanSendDocuments:      false,
	CanSendVideoNotes:     false,
	CanSendVoiceNotes:     false,
	CanAddWebPagePreviews: false,
	CanChangeInfo:         false,
	CanInviteUsers:        false,
	CanPinMessages:        false,
	CanManageTopics:       false,
	CanSendPolls:          false,
	CanSendOtherMessages:  false,
}

// CheckRestrictPermissions consolidates common moderation permission checks.
func CheckRestrictPermissions(b *gotgbot.Bot, ctx *ext.Context, userId int64) bool {
	if !RequireGroup(b, ctx, nil, false) {
		return false
	}
	if !RequireUserAdmin(b, ctx, nil, userId, false) {
		return false
	}
	if !RequireBotAdmin(b, ctx, nil, false) {
		return false
	}
	if !CanUserRestrict(b, ctx, nil, userId, false) {
		return false
	}
	if !CanBotRestrict(b, ctx, nil, false) {
		return false
	}
	return true
}
