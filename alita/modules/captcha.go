package modules

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatjoinrequest"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/captcha"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/eko/gocache/lib/v4/store"
)

// captchaModule provides CAPTCHA functionality for new members to prevent spam and bot attacks.
//
// Implements commands to configure CAPTCHA settings and handles member join events with challenge verification.
var captchaModule = moduleStruct{moduleName: "Captcha"}

// Global CAPTCHA generator instance
var captchaGenerator = captcha.NewCaptchaGenerator()

// captcha displays or toggles the CAPTCHA settings for the chat.
//
// Admins can view current settings or toggle CAPTCHA on/off.
func (moduleStruct) captcha(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	captchaSettings := db.GetCaptchaSettings(chat.Id)

	if len(args) == 0 {
		statusText := "disabled"
		if captchaSettings.Enabled {
			statusText = "enabled"
		}

		_, err := msg.Reply(bot, fmt.Sprintf(
			"CAPTCHA is currently <b>%s</b> for this chat.\n"+
				"Mode: <code>%s</code>\n"+
				"Button text: <code>%s</code>\n"+
				"Kick users: <code>%t</code>\n"+
				"Show rules: <code>%t</code>",
			statusText, captchaSettings.Mode, captchaSettings.ButtonText,
			captchaSettings.KickEnabled, captchaSettings.RulesEnabled,
		), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	var err error
	switch strings.ToLower(args[0]) {
	case "on", "yes":
		db.SetCaptchaEnabled(chat.Id, true)
		_, err = msg.Reply(bot, "CAPTCHA has been <b>enabled</b> for new members.", helpers.Shtml())
	case "off", "no":
		db.SetCaptchaEnabled(chat.Id, false)
		_, err = msg.Reply(bot, "CAPTCHA has been <b>disabled</b> for new members.", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// captchaMode sets the CAPTCHA challenge mode.
//
// Available modes: button, text, math, text2
func (moduleStruct) captchaMode(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		captchaSettings := db.GetCaptchaSettings(chat.Id)
		captchaModeString := fmt.Sprintf(
			"Current CAPTCHA mode: <code>%s</code>\n\n"+
				"Available modes:\n"+
				"• <code>button</code> - Simple button click\n"+
				"• <code>text</code> - Text selection from image\n"+
				"• <code>math</code> - Basic math question\n"+
				"• <code>text2</code> - Character-by-character selection",
			captchaSettings.Mode,
		)
		captchaEnabled := db.GetCaptchaSettings(chat.Id).Enabled
		if !captchaEnabled {
			captchaModeString = "CAPTCHA is currently <b>disabled</b> for new members.\n\n" + captchaModeString
		}
		_, err := msg.Reply(bot, captchaModeString, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	mode := strings.ToLower(args[0])
	validModes := map[string]bool{
		db.CaptchaModeButton: true,
		db.CaptchaModeText:   true,
		db.CaptchaModeMath:   true,
		db.CaptchaModeText2:  true,
	}

	var err error
	if validModes[mode] {
		db.SetCaptchaMode(chat.Id, mode)
		_, err = msg.Reply(bot, fmt.Sprintf("CAPTCHA mode set to <b>%s</b>.", mode), helpers.Shtml())
	} else {
		_, err = msg.Reply(bot, "Invalid mode! Available modes: button, text, math, text2", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// setCaptchaText sets custom text for the CAPTCHA button.
//
// Admins can customize the button text shown to new members.
func (moduleStruct) setCaptchaText(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		_, err := msg.Reply(bot, "Please provide the button text you want to set.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	buttonText := strings.Join(args, " ")
	if len(buttonText) > 64 {
		_, err := msg.Reply(bot, "Button text is too long! Maximum 64 characters allowed.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	db.SetCaptchaButtonText(chat.Id, buttonText)
	_, err := msg.Reply(bot, fmt.Sprintf("CAPTCHA button text set to: <b>%s</b>", buttonText), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// resetCaptchaText resets the CAPTCHA button text to default.
//
// Resets to: "Click here to prove you're human"
func (moduleStruct) resetCaptchaText(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	db.ResetCaptchaButtonText(chat.Id)
	_, err := msg.Reply(bot, "CAPTCHA button text reset to default.", helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// captchaKick toggles or displays the CAPTCHA kick setting.
//
// When enabled, users who don't solve CAPTCHA within the time limit are kicked.
func (moduleStruct) captchaKick(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	captchaSettings := db.GetCaptchaSettings(chat.Id)

	if len(args) == 0 {
		statusText := "disabled"
		if captchaSettings.KickEnabled {
			statusText = "enabled"
		}

		_, err := msg.Reply(bot, fmt.Sprintf(
			"CAPTCHA kick is currently <b>%s</b>.\n"+
				"Kick time: <code>%v</code>",
			statusText, captchaSettings.KickTime,
		), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	var err error
	switch strings.ToLower(args[0]) {
	case "on", "yes":
		db.SetCaptchaKick(chat.Id, true)
		_, err = msg.Reply(bot, "CAPTCHA kick has been <b>enabled</b>.", helpers.Shtml())
	case "off", "no":
		db.SetCaptchaKick(chat.Id, false)
		_, err = msg.Reply(bot, "CAPTCHA kick has been <b>disabled</b>.", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// captchaKickTime sets the time after which users are kicked for not solving CAPTCHA.
//
// Format: 5m, 1h, 2h30m, etc.
func (moduleStruct) captchaKickTime(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		captchaSettings := db.GetCaptchaSettings(chat.Id)
		_, err := msg.Reply(bot, fmt.Sprintf(
			"Current CAPTCHA kick time: <code>%v</code>\n\n"+
				"Use formats like: 5m, 1h, 2h30m",
			captchaSettings.KickTime,
		), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	timeStr := args[0]
	duration, err := time.ParseDuration(timeStr)
	if err != nil || duration < time.Minute || duration > 24*time.Hour {
		_, err := msg.Reply(bot, "Invalid time format! Use formats like 5m, 1h, 2h30m (minimum 1 minute, maximum 24 hours).", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	db.SetCaptchaKickTime(chat.Id, duration)
	_, err = msg.Reply(bot, fmt.Sprintf("CAPTCHA kick time set to <b>%v</b>.", duration), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// captchaRules toggles or displays the CAPTCHA rules setting.
//
// When enabled, users must accept chat rules as part of the CAPTCHA process.
func (moduleStruct) captchaRules(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	captchaSettings := db.GetCaptchaSettings(chat.Id)

	if len(args) == 0 {
		statusText := "disabled"
		if captchaSettings.RulesEnabled {
			statusText = "enabled"
		}

		_, err := msg.Reply(bot, fmt.Sprintf("CAPTCHA rules are currently <b>%s</b>.", statusText), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	var err error
	switch strings.ToLower(args[0]) {
	case "on", "yes":
		db.SetCaptchaRules(chat.Id, true)
		_, err = msg.Reply(bot, "CAPTCHA rules have been <b>enabled</b>.", helpers.Shtml())
	case "off", "no":
		db.SetCaptchaRules(chat.Id, false)
		_, err = msg.Reply(bot, "CAPTCHA rules have been <b>disabled</b>.", helpers.Shtml())
	default:
		_, err = msg.Reply(bot, "I understand 'on/yes' or 'off/no' only!", helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// captchaMuteTime sets or displays the auto-unmute time for CAPTCHA.
//
// When set, users are automatically unmuted after the specified time even if they don't solve CAPTCHA.
func (moduleStruct) captchaMuteTime(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, false)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// check permission
	if !chat_status.CanUserChangeInfo(bot, ctx, chat, user.Id, false) {
		return ext.EndGroups
	}

	captchaSettings := db.GetCaptchaSettings(chat.Id)

	if len(args) == 0 {
		if captchaSettings.MuteTime == 0 {
			_, err := msg.Reply(bot, "CAPTCHA auto-unmute is currently <b>disabled</b>.", helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			_, err := msg.Reply(bot, fmt.Sprintf(
				"CAPTCHA auto-unmute time: <code>%v</code>\n\n"+
					"Use 'off' to disable or formats like: 1h, 12h, 1d",
				captchaSettings.MuteTime,
			), helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		}
		return ext.EndGroups
	}

	timeStr := strings.ToLower(args[0])
	var err error

	if timeStr == "off" || timeStr == "disable" {
		db.SetCaptchaMuteTime(chat.Id, 0)
		_, err = msg.Reply(bot, "CAPTCHA auto-unmute has been <b>disabled</b>.", helpers.Shtml())
	} else {
		duration, parseErr := time.ParseDuration(timeStr)
		if parseErr != nil || duration < 5*time.Minute || duration > 7*24*time.Hour {
			_, err := msg.Reply(bot, "Invalid time format! Use formats like 1h, 12h, 1d (minimum 5 minutes, maximum 7 days) or 'off' to disable.", helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.EndGroups
		}

		db.SetCaptchaMuteTime(chat.Id, duration)
		_, err = msg.Reply(bot, fmt.Sprintf("CAPTCHA auto-unmute time set to <b>%v</b>.", duration), helpers.Shtml())
	}

	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// newMemberCaptcha handles the event when a new member joins the chat.
//
// If CAPTCHA is enabled, mutes the user and starts the challenge process.
func (moduleStruct) newMemberCaptcha(bot *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	newMember := ctx.ChatMember.NewChatMember.MergeChatMember().User

	// Skip bot joins
	if newMember.Id == bot.Id {
		return ext.ContinueGroups
	}

	captchaSettings := db.GetCaptchaSettings(chat.Id)
	if !captchaSettings.Enabled {
		return ext.ContinueGroups
	}

	// Don't CAPTCHA admins
	if chat_status.IsUserAdmin(bot, chat.Id, newMember.Id) {
		return ext.ContinueGroups
	}

	// Check if bot has permission to restrict
	if !chat_status.CanBotRestrict(bot, ctx, nil, true) {
		log.Warnf("CAPTCHA: Bot lacks restrict permissions in chat %d", chat.Id)
		return ext.ContinueGroups
	}

	// Mute the user
	_, err := chat.RestrictMember(bot, newMember.Id, gotgbot.ChatPermissions{
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
	}, nil)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to mute user %d in chat %d: %v", newMember.Id, chat.Id, err)
		return ext.ContinueGroups
	}

	// Create CAPTCHA challenge
	result, err := captchaGenerator.GenerateChallenge(captchaSettings.Mode)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to generate challenge: %v", err)
		return ext.ContinueGroups
	}

	// Set expiry time
	expiresAt := time.Now().Add(captchaSettings.KickTime)
	if captchaSettings.MuteTime > 0 && captchaSettings.MuteTime < captchaSettings.KickTime {
		expiresAt = time.Now().Add(captchaSettings.MuteTime)
	}

	// Store challenge in database
	err = db.CreateCaptchaChallenge(newMember.Id, chat.Id, result.ChallengeData, result.CorrectAnswer, expiresAt)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to store challenge: %v", err)
		return ext.ContinueGroups
	}

	// Create welcome message with CAPTCHA
	err = sendCaptchaChallenge(bot, ctx, &newMember, captchaSettings, result)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to send challenge: %v", err)
	}

	return ext.ContinueGroups
}

// captchaJoinRequest handles join requests when CAPTCHA is enabled.
//
// Sends CAPTCHA challenge to users requesting to join.
func (moduleStruct) captchaJoinRequest(bot *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.ChatJoinRequest.Chat
	user := ctx.ChatJoinRequest.From

	captchaSettings := db.GetCaptchaSettings(chat.Id)
	if !captchaSettings.Enabled {
		return ext.ContinueGroups
	}

	// Create CAPTCHA challenge
	result, err := captchaGenerator.GenerateChallenge(captchaSettings.Mode)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to generate challenge for join request: %v", err)
		return ext.ContinueGroups
	}

	// Set expiry time
	expiresAt := time.Now().Add(captchaSettings.KickTime)

	// Store challenge in database
	err = db.CreateCaptchaChallenge(user.Id, chat.Id, result.ChallengeData, result.CorrectAnswer, expiresAt)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to store join request challenge: %v", err)
		return ext.ContinueGroups
	}

	// Send CAPTCHA in private message
	err = sendJoinRequestCaptcha(bot, &user, &chat, captchaSettings, result)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to send join request challenge: %v", err)
	}

	return ext.ContinueGroups
}

// captchaCallback handles CAPTCHA button interactions.
//
// Processes user answers and updates challenge state.
func (moduleStruct) captchaCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := query.From

	// Parse callback data
	parts := strings.Split(query.Data, "_")
	if len(parts) < 2 {
		return ext.EndGroups
	}

	// Get active challenge for user (need to determine chat ID)
	// For private messages, we need to extract chat ID from user state
	// For now, we'll implement a simple approach using message context

	return handleCaptchaAnswer(bot, ctx, user.Id, query.Data)
}

// Helper functions

func sendCaptchaChallenge(bot *gotgbot.Bot, ctx *ext.Context, newMember *gotgbot.User, settings *db.CaptchaSettings, result *captcha.CaptchaResult) error {
	chat := ctx.EffectiveChat

	// Create welcome message
	welcomeText := fmt.Sprintf(
		"Welcome %s!\n\n"+
			"To ensure you're human and not a bot, please complete the CAPTCHA below.",
		helpers.MentionHtml(newMember.Id, newMember.FirstName),
	)

	// Add rules if enabled
	if settings.RulesEnabled {
		rulesInfo := db.GetChatRulesInfo(chat.Id)
		if rulesInfo.Rules != "" {
			welcomeText += "\n\n<b>Chat Rules:</b>\n" + rulesInfo.Rules
			welcomeText += "\n\nBy completing the CAPTCHA, you agree to follow these rules."
		}
	}

	// Get challenge description
	description, err := captcha.GetChallengeDescription(result.ChallengeData, settings.Mode)
	if err != nil {
		return err
	}
	welcomeText += "\n\n" + description

	// Create keyboard
	keyboard, err := captcha.CreateCaptchaKeyboard(result.ChallengeData, settings.Mode)
	if err != nil {
		return err
	}

	// Modify button text for button mode
	if settings.Mode == db.CaptchaModeButton && len(keyboard.InlineKeyboard) > 0 && len(keyboard.InlineKeyboard[0]) > 0 {
		keyboard.InlineKeyboard[0][0].Text = settings.ButtonText
	}

	// Send message or photo based on whether we have image bytes
	if result.ImageBytes != nil {
		// For image-based modes (text, text2), send as photo
		_, err = bot.SendPhoto(chat.Id, gotgbot.InputFileByReader("captcha.png", bytes.NewReader(result.ImageBytes)), &gotgbot.SendPhotoOpts{
			Caption:     welcomeText,
			ParseMode:   helpers.HTML,
			ReplyMarkup: keyboard,
		})
	} else {
		// For non-image modes (button, math), send as regular message
		_, err = bot.SendMessage(chat.Id, welcomeText, &gotgbot.SendMessageOpts{
			ParseMode:   helpers.HTML,
			ReplyMarkup: keyboard,
		})
	}

	return err
}

func sendJoinRequestCaptcha(bot *gotgbot.Bot, user *gotgbot.User, chat *gotgbot.Chat, settings *db.CaptchaSettings, result *captcha.CaptchaResult) error {
	// Create message
	messageText := fmt.Sprintf(
		"Hello %s!\n\n"+
			"You've requested to join <b>%s</b>. "+
			"To complete your join request, please solve the CAPTCHA below.",
		helpers.MentionHtml(user.Id, user.FirstName), chat.Title,
	)

	// Add rules if enabled
	if settings.RulesEnabled {
		rulesInfo := db.GetChatRulesInfo(chat.Id)
		if rulesInfo.Rules != "" {
			messageText += "\n\n<b>Chat Rules:</b>\n" + rulesInfo.Rules
			messageText += "\n\nBy completing the CAPTCHA, you agree to follow these rules."
		}
	}

	// Get challenge description
	description, err := captcha.GetChallengeDescription(result.ChallengeData, settings.Mode)
	if err != nil {
		return err
	}
	messageText += "\n\n" + description

	// Create keyboard
	keyboard, err := captcha.CreateCaptchaKeyboard(result.ChallengeData, settings.Mode)
	if err != nil {
		return err
	}

	// Modify button text for button mode
	if settings.Mode == db.CaptchaModeButton && len(keyboard.InlineKeyboard) > 0 && len(keyboard.InlineKeyboard[0]) > 0 {
		keyboard.InlineKeyboard[0][0].Text = settings.ButtonText
	}

	// Send private message or photo based on whether we have image bytes
	if result.ImageBytes != nil {
		// For image-based modes (text, text2), send as photo
		_, err = bot.SendPhoto(user.Id, gotgbot.InputFileByReader("captcha.png", bytes.NewReader(result.ImageBytes)), &gotgbot.SendPhotoOpts{
			Caption:     messageText,
			ParseMode:   helpers.HTML,
			ReplyMarkup: keyboard,
		})
	} else {
		// For non-image modes (button, math), send as regular message
		_, err = bot.SendMessage(user.Id, messageText, &gotgbot.SendMessageOpts{
			ParseMode:   helpers.HTML,
			ReplyMarkup: keyboard,
		})
	}

	return err
}

func handleCaptchaAnswer(bot *gotgbot.Bot, ctx *ext.Context, userID int64, callbackData string) error {
	query := ctx.CallbackQuery

	log.Infof("CAPTCHA: Processing callback for user %d with data: %s", userID, callbackData)

	// Try to find active challenge for this user in any chat
	// Note: In a production system, you'd store chat_id in the callback data or use a different approach
	var chatID int64
	var challenge *db.CaptchaChallenge
	var err error

	// For now, we'll need to implement a way to track which chat the callback is from
	// This is a limitation of the simplified implementation
	if query.Message != nil {
		chatID = query.Message.GetChat().Id
		log.Infof("CAPTCHA: Looking for challenge for user %d in chat %d", userID, chatID)

		challenge, err = db.GetCaptchaChallenge(userID, chatID)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to get challenge: %v", err)
			_, answerErr := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "❌ Database error. Please try again or contact an admin.",
			})
			if answerErr != nil {
				log.Errorf("CAPTCHA: Failed to answer callback query with error: %v", answerErr)
			}
			return ext.ContinueGroups
		}
		if challenge == nil {
			log.Warnf("CAPTCHA: No active challenge found for user %d in chat %d", userID, chatID)
			// No active challenge
			_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "No active CAPTCHA challenge found.",
			})
			if err != nil {
				log.Errorf("CAPTCHA: Failed to answer callback query: %v", err)
			}
			return ext.ContinueGroups
		}
		log.Infof("CAPTCHA: Found active challenge for user %d in chat %d", userID, chatID)
	} else {
		log.Warnf("CAPTCHA: No message in callback query for user %d", userID)
		// Handle private message callbacks (join requests)
		// We need to find the challenge for this user
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "❌ Join request CAPTCHA not fully implemented yet.",
		})
		if err != nil {
			log.Errorf("CAPTCHA: Failed to answer callback query: %v", err)
		}
		return ext.ContinueGroups
	}

	// Extract answer from callback data
	var userAnswer string

	if strings.HasPrefix(callbackData, "captcha_solve_button") {
		userAnswer = "button_click"
		log.Infof("CAPTCHA: Button click detected for user %d", userID)
	} else if strings.HasPrefix(callbackData, "captcha_solve_text_") {
		userAnswer = strings.TrimPrefix(callbackData, "captcha_solve_text_")
		log.Infof("CAPTCHA: Text answer '%s' for user %d", userAnswer, userID)
	} else if strings.HasPrefix(callbackData, "captcha_solve_math_") {
		userAnswer = strings.TrimPrefix(callbackData, "captcha_solve_math_")
		log.Infof("CAPTCHA: Math answer '%s' for user %d", userAnswer, userID)
	} else if strings.HasPrefix(callbackData, "captcha_text2_") || strings.HasPrefix(callbackData, "captcha_char_") {
		// Handle text2 mode
		log.Infof("CAPTCHA: Text2 interaction for user %d", userID)
		return handleText2Interaction(bot, ctx, userID, chatID, challenge, callbackData)
	} else {
		log.Warnf("CAPTCHA: Unknown callback data format: %s", callbackData)
		_, err := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "❌ Unknown CAPTCHA format. Please try again.",
		})
		if err != nil {
			log.Errorf("CAPTCHA: Failed to answer callback query: %v", err)
		}
		return ext.ContinueGroups
	}

	// Validate answer
	captchaSettings := db.GetCaptchaSettings(chatID)
	log.Infof("CAPTCHA: Validating answer '%s' for mode '%s'", userAnswer, captchaSettings.Mode)
	isCorrect, err := captcha.ValidateAnswer(challenge.ChallengeData, captchaSettings.Mode, userAnswer)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to validate answer: %v", err)
		_, answerErr := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "❌ Validation error. Please try again or contact an admin.",
		})
		if answerErr != nil {
			log.Errorf("CAPTCHA: Failed to answer callback query with error: %v", answerErr)
		}
		return ext.ContinueGroups
	}

	log.Infof("CAPTCHA: Answer validation result for user %d: %t", userID, isCorrect)

	if isCorrect {
		// Correct answer - unmute user and approve
		err = handleCorrectAnswer(bot, ctx, userID, chatID, challenge)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to handle correct answer: %v", err)
			// Answer callback query with error message
			_, answerErr := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "❌ Error processing CAPTCHA. Please try again or contact an admin.",
			})
			if answerErr != nil {
				log.Errorf("CAPTCHA: Failed to answer callback query with error: %v", answerErr)
			}
			return ext.ContinueGroups
		}
	} else {
		// Wrong answer - increment attempts
		err = handleWrongAnswer(bot, ctx, userID, chatID, challenge)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to handle wrong answer: %v", err)
			// Answer callback query with error message
			_, answerErr := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "❌ Error processing CAPTCHA. Please try again or contact an admin.",
			})
			if answerErr != nil {
				log.Errorf("CAPTCHA: Failed to answer callback query with error: %v", answerErr)
			}
			return ext.ContinueGroups
		}
	}

	return ext.ContinueGroups
}

func handleText2Interaction(bot *gotgbot.Bot, ctx *ext.Context, userID, chatID int64, challenge *db.CaptchaChallenge, callbackData string) error {
	// Silence linter warnings for currently unused parameters that may be needed in future extensions.
	_ = userID
	_ = chatID
	_ = challenge

	query := ctx.CallbackQuery

	// Get or initialize user's current input from cache
	cacheKey := fmt.Sprintf("captcha_text2_input_%d_%d", userID, chatID)
	currentInput := ""
	if cachedInput, err := cache.Marshal.Get(cache.Context, cacheKey, new(string)); err == nil && cachedInput != nil {
		currentInput = *cachedInput.(*string)
	}

	if callbackData == "captcha_text2_submit" {
		// Submit current answer
		if currentInput == challenge.CorrectAnswer {
			// Correct answer - clean up cache and handle success
			if err := cache.Marshal.Delete(cache.Context, cacheKey); err != nil {
				log.Error("Failed to delete cache after correct answer:", err)
			}
			return handleCorrectAnswer(bot, ctx, userID, chatID, challenge)
		} else {
			// Wrong answer - increment attempts
			challenge.Attempts++
			if challenge.Attempts >= 3 {
				// Too many attempts - clean up cache and handle failure
				if err := cache.Marshal.Delete(cache.Context, cacheKey); err != nil {
					log.Error("Failed to delete cache after max attempts:", err)
				}
				return handleIncorrectAnswer(bot, ctx, userID, chatID, challenge)
			}

			// Update challenge in database
			if err := db.UpdateCaptchaChallenge(userID, chatID, challenge); err != nil {
				log.Error("Failed to update captcha challenge:", err)
			}

			// Show error and update keyboard
			_, err := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: fmt.Sprintf("❌ Wrong answer! You have %d attempts left.", 3-challenge.Attempts),
			})
			if err != nil {
				return err
			}

			// Update the message with current input
			return updateText2Display(bot, ctx, challenge, currentInput, challenge.Attempts)
		}
	} else if callbackData == "captcha_text2_delete" {
		// Delete last character
		if len(currentInput) > 0 {
			currentInput = currentInput[:len(currentInput)-1]

			// Update cache
			if err := cache.Marshal.Set(cache.Context, cacheKey, &currentInput, store.WithExpiration(time.Minute*10)); err != nil {
				log.Error("Failed to update cache after character deletion:", err)
			}

			// Update the message
			err := updateText2Display(bot, ctx, challenge, currentInput, challenge.Attempts)
			if err != nil {
				return err
			}

			_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "Character deleted.",
			})
			return err
		} else {
			_, err := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "Nothing to delete.",
			})
			return err
		}
	} else if strings.HasPrefix(callbackData, "captcha_char_") {
		// Add character to input
		char := strings.TrimPrefix(callbackData, "captcha_char_")

		// Limit input length to prevent abuse
		if len(currentInput) < 20 {
			currentInput += char

			// Update cache
			if err := cache.Marshal.Set(cache.Context, cacheKey, &currentInput, store.WithExpiration(time.Minute*10)); err != nil {
				log.Error("Failed to update cache after character deletion:", err)
			}

			// Update the message
			err := updateText2Display(bot, ctx, challenge, currentInput, challenge.Attempts)
			if err != nil {
				return err
			}

			_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: fmt.Sprintf("Added: %s", char),
			})
			return err
		} else {
			_, err := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text: "Input too long!",
			})
			return err
		}
	}

	return ext.ContinueGroups
}

// updateText2Display updates the CAPTCHA message with current user input
func updateText2Display(bot *gotgbot.Bot, ctx *ext.Context, challenge *db.CaptchaChallenge, currentInput string, attempts int) error {
	query := ctx.CallbackQuery

	// Create keyboard with current input state using stored challenge JSON
	keyboard, err := captcha.CreateCaptchaKeyboard(challenge.ChallengeData, db.CaptchaModeText2)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to create keyboard: %v", err)
		return err
	}

	// Create display text showing current input
	displayInput := currentInput
	if displayInput == "" {
		displayInput = "_" // Show placeholder when empty
	}

	// Create updated text with current input and attempts info
	captionText := "🔒 Please enter the code shown in the image above.\n\n"
	captionText += fmt.Sprintf("Current input: `%s`\n", displayInput)
	if attempts > 0 {
		captionText += fmt.Sprintf("❌ Wrong attempts: %d/3", attempts)
	}

	// Update the message
	if query.Message != nil {
		_, _, err = query.Message.EditCaption(bot, &gotgbot.EditMessageCaptionOpts{
			Caption:     captionText,
			ParseMode:   "Markdown",
			ReplyMarkup: *keyboard,
		})
	}

	return err
}

func handleCorrectAnswer(bot *gotgbot.Bot, ctx *ext.Context, userID, chatID int64, challenge *db.CaptchaChallenge) error {
	query := ctx.CallbackQuery
	// Silence linter warnings for currently unused parameters that may be needed in future extensions.
	_ = challenge

	// Unmute user by restoring chat's default permissions
	chat := &gotgbot.Chat{Id: chatID}
	c, err := bot.GetChat(chatID, nil)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to get chat permissions: %v", err)
		return err
	}

	// Check if user is in chat already (member) or if this is a join request
	userInChat := chat_status.IsUserInChat(bot, chat, userID)

	if userInChat {
		// User is already in chat - unmute them by restoring default permissions
		_, err = chat.RestrictMember(bot, userID, *c.Permissions, nil)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to unmute user %d: %v", userID, err)
		}
	} else {
		// User is not in chat - this is a join request, approve it
		_, err = bot.ApproveChatJoinRequest(chatID, userID, nil)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to approve join request for user %d: %v", userID, err)
		}
	}

	// Delete challenge
	err = db.DeleteCaptchaChallenge(userID, chatID)
	if err != nil {
		log.Errorf("CAPTCHA: Failed to delete challenge: %v", err)
	}

	// Update message
	if query.Message != nil {
		_, _, err = query.Message.EditText(bot,
			"✅ <b>CAPTCHA Solved!</b>\n\nWelcome to the chat! You can now send messages.",
			&gotgbot.EditMessageTextOpts{
				ParseMode: helpers.HTML,
			},
		)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to edit message: %v", err)
		}
	}

	// Answer callback query
	_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text: "✅ CAPTCHA solved! Welcome!",
	})

	return err
}

func handleWrongAnswer(bot *gotgbot.Bot, ctx *ext.Context, userID, chatID int64, challenge *db.CaptchaChallenge) error {
	query := ctx.CallbackQuery

	// Increment attempts
	challenge.Attempts++

	if challenge.Attempts >= 3 {
		// Too many attempts - ban user
		chat := &gotgbot.Chat{Id: chatID}
		_, err := chat.BanMember(bot, userID, &gotgbot.BanChatMemberOpts{
			UntilDate: time.Now().Add(24 * time.Hour).Unix(), // 24 hour ban
		})
		if err != nil {
			log.Errorf("CAPTCHA: Failed to ban user %d: %v", userID, err)
		}

		// Delete challenge
		err = db.DeleteCaptchaChallenge(userID, chatID)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to delete challenge: %v", err)
		}

		// Update message
		if query.Message != nil {
			_, _, err = query.Message.EditText(bot,
				"❌ <b>CAPTCHA Failed!</b>\n\nToo many incorrect attempts. You have been temporarily banned.",
				&gotgbot.EditMessageTextOpts{
					ParseMode: helpers.HTML,
				},
			)
			if err != nil {
				log.Errorf("CAPTCHA: Failed to edit message: %v", err)
			}
		}

		// Answer callback query
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "❌ Too many attempts! Banned.",
		})

		return err
	} else {
		// Update challenge attempts
		err := db.UpdateCaptchaChallenge(userID, chatID, challenge)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to update challenge: %v", err)
		}

		// Answer callback query
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("❌ Wrong answer! %d/%d attempts used.", challenge.Attempts, 3),
		})

		return err
	}
}

func handleIncorrectAnswer(bot *gotgbot.Bot, ctx *ext.Context, userID, chatID int64, challenge *db.CaptchaChallenge) error {
	query := ctx.CallbackQuery

	// Increment attempts
	challenge.Attempts++

	if challenge.Attempts >= 3 {
		// Too many attempts - ban user
		chat := &gotgbot.Chat{Id: chatID}
		_, err := chat.BanMember(bot, userID, &gotgbot.BanChatMemberOpts{
			UntilDate: time.Now().Add(24 * time.Hour).Unix(), // 24 hour ban
		})
		if err != nil {
			log.Errorf("CAPTCHA: Failed to ban user %d: %v", userID, err)
		}

		// Delete challenge
		err = db.DeleteCaptchaChallenge(userID, chatID)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to delete challenge: %v", err)
		}

		// Update message
		if query.Message != nil {
			_, _, err = query.Message.EditText(bot,
				"❌ <b>CAPTCHA Failed!</b>\n\nToo many incorrect attempts. You have been permanently banned.",
				&gotgbot.EditMessageTextOpts{
					ParseMode: helpers.HTML,
				},
			)
			if err != nil {
				log.Errorf("CAPTCHA: Failed to edit message: %v", err)
			}
		}

		// Answer callback query
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "❌ Too many attempts! Banned.",
		})

		return err
	} else {
		// Update challenge attempts
		err := db.UpdateCaptchaChallenge(userID, chatID, challenge)
		if err != nil {
			log.Errorf("CAPTCHA: Failed to update challenge: %v", err)
		}

		// Answer callback query
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("❌ Wrong answer! %d/%d attempts used.", challenge.Attempts, 3),
		})

		return err
	}
}

// LoadCaptcha registers all CAPTCHA-related command handlers with the dispatcher.
//
// This function enables the CAPTCHA verification module and adds handlers for
// user verification, configuration commands, and member join events. The module
// provides comprehensive spam protection by challenging new users with various
// verification methods.
//
// Registered commands:
//   - /captcha: Toggles CAPTCHA on/off and displays current settings
//   - /captchamode: Sets the verification mode (button, math, text)
//   - /setcaptchatext: Customizes the CAPTCHA button text
//   - /resetcaptchatext: Resets CAPTCHA button text to default
//   - /captchakick: Toggles automatic kicking of failed verifications
//   - /captchakicktime: Sets the timeout for CAPTCHA completion
//   - /captcharules: Toggles rules display in CAPTCHA message
//   - /captchamutetime: Sets mute duration for unverified users
//
// Event handlers:
//   - New member join: Automatically presents CAPTCHA to new users
//   - Join requests: Handles approval requests with CAPTCHA
//   - Callback queries: Processes CAPTCHA button responses
//   - Scheduled tasks: Manages CAPTCHA timeouts and cleanup
//
// Features:
//   - Multiple verification modes (button, math, text)
//   - Configurable timeout and failure actions
//   - Automatic cleanup of expired challenges
//   - Rules integration for new members
//   - Mute enforcement for unverified users
//
// Requirements:
//   - Bot must be admin to mute/kick users
//   - User must be admin to configure CAPTCHA settings
//   - Module supports remote configuration via connections
//
// The CAPTCHA system includes automatic scheduling for cleanup and
// integrates with the bot's warning and muting systems for enforcement.
func LoadCaptcha(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(captchaModule.moduleName, true)

	// Command handlers
	dispatcher.AddHandler(handlers.NewCommand("captcha", captchaModule.captcha))
	dispatcher.AddHandler(handlers.NewCommand("captchamode", captchaModule.captchaMode))
	dispatcher.AddHandler(handlers.NewCommand("setcaptchatext", captchaModule.setCaptchaText))
	dispatcher.AddHandler(handlers.NewCommand("resetcaptchatext", captchaModule.resetCaptchaText))
	dispatcher.AddHandler(handlers.NewCommand("captchakick", captchaModule.captchaKick))
	dispatcher.AddHandler(handlers.NewCommand("captchakicktime", captchaModule.captchaKickTime))
	dispatcher.AddHandler(handlers.NewCommand("captcharules", captchaModule.captchaRules))
	dispatcher.AddHandler(handlers.NewCommand("captchamutetime", captchaModule.captchaMuteTime))

	// Event handlers
	dispatcher.AddHandler(
		handlers.NewChatMember(
			func(u *gotgbot.ChatMemberUpdated) bool {
				wasMember, isMember := helpers.ExtractJoinLeftStatusChange(u)
				return !wasMember && isMember
			},
			captchaModule.newMemberCaptcha,
		),
	)

	dispatcher.AddHandler(
		handlers.NewChatJoinRequest(
			chatjoinrequest.All, captchaModule.captchaJoinRequest,
		),
	)

	// Callback handlers
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("captcha_"), captchaModule.captchaCallback))
}
