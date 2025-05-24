package permissions

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
)

// PermissionCheck represents a single permission validation
type PermissionCheck func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User) bool

// PermissionChecker provides a chainable interface for permission checking
type PermissionChecker struct {
	bot    *gotgbot.Bot
	ctx    *ext.Context
	chat   *gotgbot.Chat
	user   *gotgbot.User
	checks []PermissionCheck
}

// NewPermissionChecker creates a new permission checker instance
func NewPermissionChecker(bot *gotgbot.Bot, ctx *ext.Context) *PermissionChecker {
	return &PermissionChecker{
		bot:    bot,
		ctx:    ctx,
		chat:   ctx.EffectiveChat,
		user:   ctx.EffectiveSender.User,
		checks: make([]PermissionCheck, 0),
	}
}

// RequireGroup ensures the command is used in a group chat
func (pc *PermissionChecker) RequireGroup() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.RequireGroup(b, ctx, chat, false)
	})
	return pc
}

// RequirePrivate ensures the command is used in a private chat
func (pc *PermissionChecker) RequirePrivate() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.RequirePrivate(b, ctx, chat, false)
	})
	return pc
}

// RequireUserAdmin ensures the user is an admin
func (pc *PermissionChecker) RequireUserAdmin(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.RequireUserAdmin(b, ctx, chat, userID, false)
	})
	return pc
}

// RequireBotAdmin ensures the bot is an admin
func (pc *PermissionChecker) RequireBotAdmin() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.RequireBotAdmin(b, ctx, chat, false)
	})
	return pc
}

// CanUserRestrict ensures the user can restrict members
func (pc *PermissionChecker) CanUserRestrict(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanUserRestrict(b, ctx, chat, userID, false)
	})
	return pc
}

// CanBotRestrict ensures the bot can restrict members
func (pc *PermissionChecker) CanBotRestrict() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanBotRestrict(b, ctx, chat, false)
	})
	return pc
}

// CanUserPromote ensures the user can promote members
func (pc *PermissionChecker) CanUserPromote(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanUserPromote(b, ctx, chat, userID, false)
	})
	return pc
}

// CanBotPromote ensures the bot can promote members
func (pc *PermissionChecker) CanBotPromote() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanBotPromote(b, ctx, chat, false)
	})
	return pc
}

// CanUserChangeInfo ensures the user can change chat info
func (pc *PermissionChecker) CanUserChangeInfo(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanUserChangeInfo(b, ctx, chat, userID, false)
	})
	return pc
}

// CanUserPin ensures the user can pin messages
func (pc *PermissionChecker) CanUserPin(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanUserPin(b, ctx, chat, userID, false)
	})
	return pc
}

// CanBotPin ensures the bot can pin messages
func (pc *PermissionChecker) CanBotPin() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanBotPin(b, ctx, chat, false)
	})
	return pc
}

// CanUserDelete ensures the user can delete messages
func (pc *PermissionChecker) CanUserDelete(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanUserDelete(b, ctx, chat, userID, false)
	})
	return pc
}

// CanBotDelete ensures the bot can delete messages
func (pc *PermissionChecker) CanBotDelete() *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.CanBotDelete(b, ctx, chat, false)
	})
	return pc
}

// RequireUserOwner ensures the user is the chat owner
func (pc *PermissionChecker) RequireUserOwner(userID int64) *PermissionChecker {
	pc.checks = append(pc.checks, func(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User) bool {
		return chat_status.RequireUserOwner(b, ctx, chat, userID, false)
	})
	return pc
}

// Check executes all permission checks and returns true if all pass
func (pc *PermissionChecker) Check() bool {
	for _, check := range pc.checks {
		if !check(pc.bot, pc.ctx, pc.chat, pc.user) {
			return false
		}
	}
	return true
}

// CheckWithConnection executes all permission checks with connection support
// Returns the connected chat if successful, nil if any check fails
func (pc *PermissionChecker) CheckWithConnection(connectCheck func(*gotgbot.Bot, *ext.Context) *gotgbot.Chat) *gotgbot.Chat {
	// First check connection
	connectedChat := connectCheck(pc.bot, pc.ctx)
	if connectedChat == nil {
		return nil
	}

	// Update context with connected chat
	pc.ctx.EffectiveChat = connectedChat
	pc.chat = connectedChat

	// Run all permission checks
	for _, check := range pc.checks {
		if !check(pc.bot, pc.ctx, pc.chat, pc.user) {
			return nil
		}
	}

	return connectedChat
}
