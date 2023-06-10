package helpers

import (
	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// ExtractJoinLeftStatusChange Takes a ChatMemberUpdated instance and extracts whether the 'old_chat_member' was a member
// of the chat and whether the 'new_chat_member' is a member of the chat. Returns false, if
// the status didn't change.
func ExtractJoinLeftStatusChange(u *gotgbot.ChatMemberUpdated) (bool, bool) {
	// return false for channels
	if u.Chat.Type == "channel" {
		return false, false
	}

	oldMemberStatus := u.OldChatMember.MergeChatMember().Status
	newMemberStatus := u.NewChatMember.MergeChatMember().Status
	oldIsMember := u.OldChatMember.MergeChatMember().IsMember
	newIsMember := u.NewChatMember.MergeChatMember().IsMember

	if oldMemberStatus == newMemberStatus {
		return false, false
	}

	wasMember := string_handling.FindInStringSlice(
		[]string{"member", "administrator", "creator"},
		oldMemberStatus,
	) || (oldMemberStatus == "restricted" && oldIsMember)

	isMember := string_handling.FindInStringSlice(
		[]string{"member", "administrator", "creator"},
		newMemberStatus,
	) || (newMemberStatus == "restricted" && newIsMember)

	return wasMember, isMember
}

// ExtractAdminUpdateStatusChange Takes a ChatMemberUpdated instance and extracts whether the 'old_chat_member' was a member or admin
// of the chat and whether the 'new_chat_member' is a admin of the chat. Returns false, if
// the status didn't change.
func ExtractAdminUpdateStatusChange(u *gotgbot.ChatMemberUpdated) bool {
	// return false for channels
	if u.Chat.Type == "channel" {
		return false
	}

	oldMemberStatus := u.OldChatMember.MergeChatMember().Status
	newMemberStatus := u.NewChatMember.MergeChatMember().Status

	// status remains same
	if oldMemberStatus == newMemberStatus {
		return false
	}

	adminStatusChanged := (string_handling.FindInStringSlice(
		[]string{"administrator", "creator"},
		oldMemberStatus,
	) && !string_handling.FindInStringSlice(
		[]string{"administrator", "creator"},
		newMemberStatus,
	)) ||
		(string_handling.FindInStringSlice(
			[]string{"administrator", "creator"},
			newMemberStatus,
		) && !string_handling.FindInStringSlice(
			[]string{"administrator", "creator"},
			oldMemberStatus,
		))

	return adminStatusChanged
}
