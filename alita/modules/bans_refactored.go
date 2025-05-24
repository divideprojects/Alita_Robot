package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/command"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/messaging"
	"github.com/divideprojects/Alita_Robot/alita/utils/permissions"
	"github.com/divideprojects/Alita_Robot/alita/utils/validation"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var bansModuleRefactored = moduleStruct{moduleName: "BansRefactored"}

// REFACTORED: Using new frameworks - dramatically reduced code duplication

// kickHandler demonstrates the new framework usage
func (m moduleStruct) kickRefactored(b *gotgbot.Bot, ctx *ext.Context) error {
	return command.NewCommandHandler().
		WithPermissions(
			permissions.NewPermissionChecker(b, ctx).
				RequireGroup().
				RequireUserAdmin(ctx.EffectiveSender.User.Id).
				RequireBotAdmin().
				CanUserRestrict(ctx.EffectiveSender.User.Id).
				CanBotRestrict(),
		).
		WithUserValidation(validation.ValidateUser).
		WithArgValidation(validateKickArgs).
		WithHandler(executeKick).
		Execute(b, ctx)
}

// banHandler using the new framework
func (m moduleStruct) banRefactored(b *gotgbot.Bot, ctx *ext.Context) error {
	return command.NewCommandHandler().
		WithPermissions(
			permissions.NewPermissionChecker(b, ctx).
				RequireGroup().
				RequireUserAdmin(ctx.EffectiveSender.User.Id).
				RequireBotAdmin().
				CanUserRestrict(ctx.EffectiveSender.User.Id).
				CanBotRestrict(),
		).
		WithUserValidation(validation.ValidateUser).
		WithArgValidation(validateBanArgs).
		WithHandler(executeBan).
		Execute(b, ctx)
}

// dbanHandler (delete + ban) using the new framework
func (m moduleStruct) dbanRefactored(b *gotgbot.Bot, ctx *ext.Context) error {
	return command.NewCommandHandler().
		WithPermissions(
			permissions.NewPermissionChecker(b, ctx).
				RequireGroup().
				RequireUserAdmin(ctx.EffectiveSender.User.Id).
				RequireBotAdmin().
				CanUserRestrict(ctx.EffectiveSender.User.Id).
				CanBotRestrict().
				CanUserDelete(ctx.EffectiveSender.User.Id).
				CanBotDelete(),
		).
		WithUserValidation(validation.ValidateUser).
		WithArgValidation(validateDbanArgs).
		WithHandler(executeDban).
		Execute(b, ctx)
}

// unbanHandler using the new framework
func (m moduleStruct) unbanRefactored(b *gotgbot.Bot, ctx *ext.Context) error {
	return command.NewCommandHandler().
		WithPermissions(
			permissions.NewPermissionChecker(b, ctx).
				RequireGroup().
				RequireUserAdmin(ctx.EffectiveSender.User.Id).
				RequireBotAdmin().
				CanUserRestrict(ctx.EffectiveSender.User.Id).
				CanBotRestrict(),
		).
		WithUserValidation(validation.ValidateUser).
		WithArgValidation(validateUnbanArgs).
		WithHandler(executeUnban).
		Execute(b, ctx)
}

// Validation functions - centralized and reusable

func validateKickArgs(args []string, msg *gotgbot.Message) (bool, string) {
	if msg.ReplyToMessage == nil && len(args) < 2 {
		return false, "I don't know who you're talking about, you're going to need to specify a user...!"
	}
	return true, ""
}

func validateBanArgs(args []string, msg *gotgbot.Message) (bool, string) {
	if msg.ReplyToMessage == nil && len(args) < 2 {
		return false, "I don't know who you're talking about, you're going to need to specify a user...!"
	}
	return true, ""
}

func validateDbanArgs(args []string, msg *gotgbot.Message) (bool, string) {
	if msg.ReplyToMessage == nil {
		return false, "Reply to a user's message to delete and ban them!"
	}
	return true, ""
}

func validateUnbanArgs(args []string, msg *gotgbot.Message) (bool, string) {
	if msg.ReplyToMessage == nil && len(args) < 2 {
		return false, "I don't know who you're talking about, you're going to need to specify a user...!"
	}
	return true, ""
}

// Handler functions - focused on business logic only

func executeKick(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User, args []string) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Extract user and reason using existing utility
	userResult, err := validation.ValidateUser(b, ctx)
	if err != nil {
		return err
	}
	if !userResult.Valid {
		return ext.EndGroups
	}

	// Additional business logic checks
	if userResult.IsUserBot() && userResult.UserID == b.Id {
		content := &messaging.GreetingContent{
			Text:     tr.GetString("strings." + bansModuleRefactored.moduleName + ".kick.is_bot_itself"),
			DataType: db.TEXT,
		}
		_, err := messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
			ReplyToMessageID: msg.MessageId,
			ParseMode:        helpers.HTML,
		})
		return err
	}

	// Perform the kick
	_, err = chat.BanMember(b, userResult.UserID, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	time.Sleep(2 * time.Second)

	_, err = chat.UnbanMember(b, userResult.UserID, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// Send success message using new messaging framework
	baseStr := tr.GetString("strings." + bansModuleRefactored.moduleName + ".kick.kicked_user")

	content := &messaging.GreetingContent{
		Text:     fmt.Sprintf(baseStr, userResult.GetUserMention()),
		DataType: db.TEXT,
	}

	_, err = messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
	})

	return err
}

func executeBan(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User, args []string) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	userResult, err := validation.ValidateUser(b, ctx)
	if err != nil {
		return err
	}
	if !userResult.Valid {
		return ext.EndGroups
	}

	// Handle anonymous users
	if strings.HasPrefix(fmt.Sprint(userResult.UserID), "-100") {
		return handleAnonymousBan(b, ctx, chat, msg)
	}

	// Perform the ban
	_, err = chat.BanMember(b, userResult.UserID, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// Send success message with unban button
	baseStr := tr.GetString("strings." + bansModuleRefactored.moduleName + ".ban.normal_ban")

	content := &messaging.GreetingContent{
		Text:     fmt.Sprintf(baseStr, userResult.GetUserMention()),
		DataType: db.TEXT,
	}

	opts := &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
		Keyboard: &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "Unban (Admin Only)",
						CallbackData: fmt.Sprintf("unrestrict.unban.%d", userResult.UserID),
					},
				},
			},
		},
	}

	_, err = messaging.SendMessage(b, ctx, content, opts)
	return err
}

func executeDban(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User, args []string) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	userResult, err := validation.ValidateUser(b, ctx)
	if err != nil {
		return err
	}
	if !userResult.Valid {
		return ext.EndGroups
	}

	// Delete the replied message
	if msg.ReplyToMessage != nil {
		_, err = msg.ReplyToMessage.Delete(b, nil)
		if err != nil {
			log.Error(err)
		}
	}

	// Perform the ban
	_, err = chat.BanMember(b, userResult.UserID, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// Send success message
	baseStr := tr.GetString("strings." + bansModuleRefactored.moduleName + ".ban.normal_ban")

	content := &messaging.GreetingContent{
		Text:     fmt.Sprintf(baseStr, userResult.GetUserMention()),
		DataType: db.TEXT,
	}

	opts := &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
		Keyboard: &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "Unban (Admin Only)",
						CallbackData: fmt.Sprintf("unrestrict.unban.%d", userResult.UserID),
					},
				},
			},
		},
	}

	_, err = messaging.SendMessage(b, ctx, content, opts)
	return err
}

func executeUnban(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User, args []string) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	userResult, err := validation.ValidateUser(b, ctx)
	if err != nil {
		return err
	}
	if !userResult.Valid {
		return ext.EndGroups
	}

	// Handle anonymous users
	if strings.HasPrefix(fmt.Sprint(userResult.UserID), "-100") {
		return handleAnonymousUnban(b, ctx, chat, msg)
	}

	// Perform the unban
	_, err = chat.UnbanMember(b, userResult.UserID, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// Send success message
	text := fmt.Sprintf(
		tr.GetString("strings."+bansModuleRefactored.moduleName+".unban.unbanned_user"),
		userResult.GetUserMention(),
	)

	content := &messaging.GreetingContent{
		Text:     text,
		DataType: db.TEXT,
	}

	_, err = messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
	})

	return err
}

// Helper functions for special cases

func handleAnonymousBan(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, msg *gotgbot.Message) error {
	var text string

	if msg.ReplyToMessage != nil {
		userId := msg.ReplyToMessage.GetSender().Id()
		_, err := b.BanChatSenderChat(chat.Id, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}
		text = "Banned user: " + helpers.MentionHtml(userId, msg.ReplyToMessage.GetSender().Name())
	} else {
		text = "You can only ban an anonymous user by replying to their message."
	}

	content := &messaging.GreetingContent{
		Text:     text,
		DataType: db.TEXT,
	}

	_, err := messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
	})

	return err
}

func handleAnonymousUnban(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, msg *gotgbot.Message) error {
	var text string

	if msg.ReplyToMessage != nil {
		userId := msg.ReplyToMessage.GetSender().Id()
		_, err := b.UnbanChatSenderChat(chat.Id, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}
		text = "Unbanned user: " + helpers.MentionHtml(userId, msg.ReplyToMessage.GetSender().Name())
	} else {
		text = "You can only unban an anonymous user by replying to their message."
	}

	content := &messaging.GreetingContent{
		Text:     text,
		DataType: db.TEXT,
	}

	_, err := messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
	})

	return err
}

// Simplified kickme command using new framework
func (m moduleStruct) kickmeRefactored(b *gotgbot.Bot, ctx *ext.Context) error {
	return command.NewCommandHandler().
		WithPermissions(
			permissions.NewPermissionChecker(b, ctx).
				RequireGroup().
				CanBotRestrict(),
		).
		WithHandler(executeKickme).
		Execute(b, ctx)
}

func executeKickme(b *gotgbot.Bot, ctx *ext.Context, chat *gotgbot.Chat, user *gotgbot.User, args []string) error {
	msg := ctx.EffectiveMessage
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}

	// Don't allow admins to use the command
	if permissions.NewPermissionChecker(b, ctx).RequireUserAdmin(user.Id).Check() {
		content := &messaging.GreetingContent{
			Text:     tr.GetString("strings." + bansModuleRefactored.moduleName + ".kickme.is_admin"),
			DataType: db.TEXT,
		}
		_, err := messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
			ReplyToMessageID: msg.MessageId,
			ParseMode:        helpers.HTML,
		})
		return err
	}

	// Kick the member
	_, err := chat.UnbanMember(b, user.Id, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	content := &messaging.GreetingContent{
		Text:     tr.GetString("strings." + bansModuleRefactored.moduleName + ".kickme.ok_out"),
		DataType: db.TEXT,
	}

	_, err = messaging.SendMessage(b, ctx, content, &messaging.SendOptions{
		ReplyToMessageID: msg.MessageId,
		ParseMode:        helpers.HTML,
	})

	return err
}

// Demonstration of how much simpler the callback handlers become
func (moduleStruct) restrictButtonHandlerRefactored(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// Single permission check using new framework
	if !permissions.NewPermissionChecker(b, ctx).CanUserRestrict(user.Id).Check() {
		return ext.EndGroups
	}

	args := strings.Split(query.Data, ".")
	action := args[0]
	userId, _ := strconv.Atoi(args[1])

	actionUser, err := b.GetChat(int64(userId), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	var helpText string

	switch action {
	case "kick":
		helpText, err = performKickAction(b, chat, user, int64(userId), actionUser)
	case "mute":
		helpText, err = performMuteAction(b, chat, user, int64(userId), actionUser)
	case "ban":
		helpText, err = performBanAction(b, chat, user, int64(userId), actionUser)
	}

	if err != nil {
		return err
	}

	_, _, err = query.Message.EditText(b, helpText, &gotgbot.EditMessageTextOpts{
		ParseMode: helpers.HTML,
	})
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b, nil)
	return err
}

// Helper functions for callback actions
func performKickAction(b *gotgbot.Bot, chat *gotgbot.Chat, user *gotgbot.User, userId int64, actionUser *gotgbot.ChatFullInfo) (string, error) {
	_, err := chat.BanMember(b, userId, nil)
	if err != nil {
		return "", err
	}

	time.Sleep(3 * time.Second)
	_, err = chat.UnbanMember(b, userId, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Admin %s kicked %s from this chat!",
		helpers.MentionHtml(user.Id, user.FirstName),
		helpers.MentionHtml(userId, actionUser.FirstName),
	), nil
}

func performMuteAction(b *gotgbot.Bot, chat *gotgbot.Chat, user *gotgbot.User, userId int64, actionUser *gotgbot.ChatFullInfo) (string, error) {
	_, err := chat.RestrictMember(b, userId,
		gotgbot.ChatPermissions{
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
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Admin %s muted %s in chat!",
		helpers.MentionHtml(user.Id, user.FirstName),
		helpers.MentionHtml(userId, actionUser.FirstName),
	), nil
}

func performBanAction(b *gotgbot.Bot, chat *gotgbot.Chat, user *gotgbot.User, userId int64, actionUser *gotgbot.ChatFullInfo) (string, error) {
	_, err := chat.BanMember(b, userId, &gotgbot.BanChatMemberOpts{})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Admin %s banned %s from this chat!",
		helpers.MentionHtml(user.Id, user.FirstName),
		helpers.MentionHtml(userId, actionUser.FirstName),
	), nil
}

// Load function demonstrating the new handlers
func LoadBansRefactored(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(bansModuleRefactored.moduleName, true)

	// Refactored ban commands using new framework
	dispatcher.AddHandler(handlers.NewCommand("banr", bansModuleRefactored.banRefactored))
	dispatcher.AddHandler(handlers.NewCommand("dbanr", bansModuleRefactored.dbanRefactored))
	dispatcher.AddHandler(handlers.NewCommand("unbanr", bansModuleRefactored.unbanRefactored))

	// Refactored kick commands
	dispatcher.AddHandler(handlers.NewCommand("kickr", bansModuleRefactored.kickRefactored))
	dispatcher.AddHandler(handlers.NewCommand("kickmer", bansModuleRefactored.kickmeRefactored))

	// Refactored callback handlers
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("restrictr."), bansModuleRefactored.restrictButtonHandlerRefactored))
}
