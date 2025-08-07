package modules

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/mojocn/base64Captcha"
	log "github.com/sirupsen/logrus"
)

var captchaModule = moduleStruct{moduleName: "Captcha"}

// captchaCommand handles the /captcha command to enable/disable captcha verification.
// Admins can use this to toggle captcha protection for their group.
func (moduleStruct) captchaCommand(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// Check permissions
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(bot, ctx, nil, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		// Show current status
		settings, _ := db.GetCaptchaSettings(chat.Id)
		status := "disabled"
		if settings.Enabled {
			status = "enabled"
		}

		text := fmt.Sprintf(
			"<b>Captcha Settings:</b>\n"+
				"Status: <code>%s</code>\n"+
				"Mode: <code>%s</code>\n"+
				"Timeout: <code>%d minutes</code>\n"+
				"Failure Action: <code>%s</code>\n"+
				"Max Attempts: <code>%d</code>\n\n"+
				"Use <code>/captcha on</code> or <code>/captcha off</code> to change status.",
			status, settings.CaptchaMode, settings.Timeout, settings.FailureAction, settings.MaxAttempts,
		)

		_, err := msg.Reply(bot, text, helpers.Shtml())
		return err
	}

	switch strings.ToLower(args[0]) {
	case "on", "enable", "yes":
		err := db.SetCaptchaEnabled(chat.Id, true)
		if err != nil {
			_, _ = msg.Reply(bot, "Failed to enable captcha. Please try again.", nil)
			return err
		}
		_, err = msg.Reply(bot, "✅ Captcha verification has been <b>enabled</b>. New members will need to complete a captcha to join.", helpers.Shtml())
		return err

	case "off", "disable", "no":
		err := db.SetCaptchaEnabled(chat.Id, false)
		if err != nil {
			_, _ = msg.Reply(bot, "Failed to disable captcha. Please try again.", nil)
			return err
		}
		// Clean up any pending captcha attempts
		go func() {
			if err := db.DeleteAllCaptchaAttempts(chat.Id); err != nil {
				log.Errorf("Failed to delete captcha attempts: %v", err)
			}
		}()
		_, err = msg.Reply(bot, "❌ Captcha verification has been <b>disabled</b>.", helpers.Shtml())
		return err

	default:
		_, err := msg.Reply(bot, "Please use <code>/captcha on</code> or <code>/captcha off</code>", helpers.Shtml())
		return err
	}
}

// captchaModeCommand handles the /captchamode command to set captcha type.
// Admins can choose between math and text captcha modes.
func (moduleStruct) captchaModeCommand(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// Check permissions
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		_, err := msg.Reply(bot, "Please specify a mode: <code>math</code> or <code>text</code>", helpers.Shtml())
		return err
	}

	mode := strings.ToLower(args[0])
	if mode != "math" && mode != "text" {
		_, err := msg.Reply(bot, "Invalid mode. Use <code>math</code> or <code>text</code>", helpers.Shtml())
		return err
	}

	err := db.SetCaptchaMode(chat.Id, mode)
	if err != nil {
		_, _ = msg.Reply(bot, "Failed to set captcha mode. Please try again.", nil)
		return err
	}

	modeDesc := "mathematical problems"
	if mode == "text" {
		modeDesc = "text recognition from images"
	}

	_, err = msg.Reply(bot, fmt.Sprintf("✅ Captcha mode set to <b>%s</b> (%s)", mode, modeDesc), helpers.Shtml())
	return err
}

// captchaTimeCommand handles the /captchatime command to set verification timeout.
// Admins can set how long users have to complete the captcha (1-10 minutes).
func (moduleStruct) captchaTimeCommand(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// Check permissions
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		_, err := msg.Reply(bot, "Please specify timeout in minutes (1-10)", nil)
		return err
	}

	timeout, err := strconv.Atoi(args[0])
	if err != nil || timeout < 1 || timeout > 10 {
		_, err = msg.Reply(bot, "Invalid timeout. Please use a number between 1 and 10.", nil)
		return err
	}

	err = db.SetCaptchaTimeout(chat.Id, timeout)
	if err != nil {
		_, _ = msg.Reply(bot, "Failed to set timeout. Please try again.", nil)
		return err
	}

	_, err = msg.Reply(bot, fmt.Sprintf("✅ Captcha timeout set to <b>%d minutes</b>", timeout), helpers.Shtml())
	return err
}

// captchaActionCommand handles the /captchaaction command to set failure action.
// Admins can choose what happens when users fail the captcha: kick, ban, or mute.
func (moduleStruct) captchaActionCommand(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()[1:]

	// Check permissions
	if !chat_status.RequireGroup(bot, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(bot, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	if len(args) == 0 {
		_, err := msg.Reply(bot, "Please specify an action: <code>kick</code>, <code>ban</code>, or <code>mute</code>", helpers.Shtml())
		return err
	}

	action := strings.ToLower(args[0])
	if action != "kick" && action != "ban" && action != "mute" {
		_, err := msg.Reply(bot, "Invalid action. Use <code>kick</code>, <code>ban</code>, or <code>mute</code>", helpers.Shtml())
		return err
	}

	err := db.SetCaptchaFailureAction(chat.Id, action)
	if err != nil {
		_, _ = msg.Reply(bot, "Failed to set failure action. Please try again.", nil)
		return err
	}

	_, err = msg.Reply(bot, fmt.Sprintf("✅ Captcha failure action set to <b>%s</b>", action), helpers.Shtml())
	return err
}

// generateMathCaptcha generates a random math problem and returns the question and answer.
func generateMathCaptcha() (string, string, []string) {
	operations := []string{"+", "-", "*"}
	operation := operations[rand.Intn(len(operations))]

	var a, b, answer int
	var question string

	switch operation {
	case "+":
		a = rand.Intn(50) + 1
		b = rand.Intn(50) + 1
		answer = a + b
		question = fmt.Sprintf("%d + %d", a, b)
	case "-":
		a = rand.Intn(50) + 20
		b = rand.Intn(a) + 1
		answer = a - b
		question = fmt.Sprintf("%d - %d", a, b)
	case "*":
		a = rand.Intn(12) + 1
		b = rand.Intn(12) + 1
		answer = a * b
		question = fmt.Sprintf("%d × %d", a, b)
	}

	// Generate wrong answers
	options := []string{strconv.Itoa(answer)}
	for len(options) < 4 {
		// Generate a wrong answer within reasonable range
		wrongAnswer := answer + rand.Intn(20) - 10
		if wrongAnswer != answer && wrongAnswer > 0 {
			wrongStr := strconv.Itoa(wrongAnswer)
			// Check if this option already exists
			if !slices.Contains(options, wrongStr) {
				options = append(options, wrongStr)
			}
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	return question, strconv.Itoa(answer), options
}

// generateTextCaptcha generates a captcha image with random text.
func generateTextCaptcha() (string, []byte, []string, error) {
	// Create captcha store (using memory store)
	store := base64Captcha.DefaultMemStore

	// Create a string driver for text captcha
	driver := base64Captcha.NewDriverString(
		80,                                 // height
		240,                                // width
		0,                                  // noiseCount
		2,                                  // showLineOptions
		4,                                  // length
		"234567890abcdefghjkmnpqrstuvwxyz", // source characters
		nil,                                // bgColor
		nil,                                // fonts
		[]string{},
	)

	// Create captcha
	captcha := base64Captcha.NewCaptcha(driver, store)

	// Generate the captcha
	id, b64s, answer, err := captcha.Generate()
	if err != nil {
		return "", nil, nil, err
	}
	_ = id // We don't use the ID

	// Decode base64 image
	// Remove data:image/png;base64, prefix if present
	if strings.HasPrefix(b64s, "data:image/") {
		parts := strings.Split(b64s, ",")
		if len(parts) > 1 {
			b64s = parts[1]
		}
	}

	imageBytes, err := base64.StdEncoding.DecodeString(b64s)
	if err != nil {
		return "", nil, nil, err
	}

	// Generate decoy answers
	options := []string{answer}
	characters := "234567890abcdefghjkmnpqrstuvwxyz"
	for len(options) < 4 {
		// Generate a random string of same length as answer
		decoy := ""
		for i := 0; i < len(answer); i++ {
			decoy += string(characters[rand.Intn(len(characters))])
		}
		// Check if this option already exists
		if !slices.Contains(options, decoy) {
			options = append(options, decoy)
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	return answer, imageBytes, options, nil
}

// SendCaptcha sends a captcha challenge to a new member.
// Called when a new member joins a group with captcha enabled.
func SendCaptcha(bot *gotgbot.Bot, ctx *ext.Context, userID int64, userName string) error {
	chat := ctx.EffectiveChat
	settings, _ := db.GetCaptchaSettings(chat.Id)

	if !settings.Enabled {
		return nil
	}

	var question string
	var answer string
	var options []string
	var imageBytes []byte

	if settings.CaptchaMode == "math" {
		// Generate math captcha
		question, answer, options = generateMathCaptcha()
	} else {
		// Generate text captcha
		var err error
		answer, imageBytes, options, err = generateTextCaptcha()
		if err != nil {
			log.Errorf("Failed to generate text captcha: %v", err)
			// Fallback to math captcha
			question, answer, options = generateMathCaptcha()
			settings.CaptchaMode = "math"
		}
	}

	// Create inline keyboard with options
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, option := range options {
		button := gotgbot.InlineKeyboardButton{
			Text:         option,
			CallbackData: fmt.Sprintf("captcha_verify.%d.%s", userID, option),
		}
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{button})
	}

	// Add refresh button for text captcha
	if settings.CaptchaMode == "text" && imageBytes != nil {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{
				Text:         "🔄 New Image",
				CallbackData: fmt.Sprintf("captcha_refresh.%d", userID),
			},
		})
	}

	keyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	// Prepare message text
	var msgText string
	if settings.CaptchaMode == "math" {
		msgText = fmt.Sprintf(
			"👋 Welcome %s!\n\n"+
				"Please solve this math problem to verify you're human:\n\n"+
				"<b>%s = ?</b>\n\n"+
				"⏱ You have <b>%d minutes</b> to answer.",
			helpers.MentionHtml(userID, userName), question, settings.Timeout,
		)
	} else {
		msgText = fmt.Sprintf(
			"👋 Welcome %s!\n\n"+
				"Please select the text shown in the image to verify you're human:\n\n"+
				"⏱ You have <b>%d minutes</b> to answer.",
			helpers.MentionHtml(userID, userName), settings.Timeout,
		)
	}

	// Send the captcha message
	var sent *gotgbot.Message
	var err error

	if settings.CaptchaMode == "text" && imageBytes != nil {
		// Send photo with text captcha
		sent, err = bot.SendPhoto(chat.Id, gotgbot.InputFileByReader("captcha.png", bytes.NewReader(imageBytes)), &gotgbot.SendPhotoOpts{
			Caption:     msgText,
			ParseMode:   helpers.HTML,
			ReplyMarkup: keyboard,
		})
	} else {
		// Send text message for math captcha
		sent, err = bot.SendMessage(chat.Id, msgText, &gotgbot.SendMessageOpts{
			ParseMode:   helpers.HTML,
			ReplyMarkup: keyboard,
		})
	}

	if err != nil {
		log.Errorf("Failed to send captcha: %v", err)
		return err
	}

	// Ensure user and chat exist in database (required for foreign key constraints)
	if err := db.EnsureUserInDb(userID, userName, userName); err != nil {
		log.Errorf("Failed to ensure user in database: %v", err)
		// Delete the captcha message since we can't track it
		_, _ = bot.DeleteMessage(chat.Id, sent.MessageId, nil)
		return err
	}
	
	if err := db.EnsureChatInDb(chat.Id, chat.Title); err != nil {
		log.Errorf("Failed to ensure chat in database: %v", err)
		// Delete the captcha message since we can't track it
		_, _ = bot.DeleteMessage(chat.Id, sent.MessageId, nil)
		return err
	}

	// Store the captcha attempt in database
	err = db.CreateCaptchaAttempt(userID, chat.Id, answer, sent.MessageId, settings.Timeout)
	if err != nil {
		log.Errorf("Failed to create captcha attempt: %v", err)
		// Delete the message if we can't track it
		_, _ = bot.DeleteMessage(chat.Id, sent.MessageId, nil)
		return err
	}

	// Schedule cleanup after timeout
	go func() {
		time.Sleep(time.Duration(settings.Timeout) * time.Minute)

		// Check if attempt still exists (not completed)
		attempt, _ := db.GetCaptchaAttempt(userID, chat.Id)
		if attempt != nil {
			// User failed to complete captcha in time
			handleCaptchaTimeout(bot, chat.Id, userID, sent.MessageId, settings.FailureAction)
		}
	}()

	return nil
}

// handleCaptchaTimeout handles when a user fails to complete captcha in time.
func handleCaptchaTimeout(bot *gotgbot.Bot, chatID, userID int64, messageID int64, action string) {
	// Delete the captcha message
	_, _ = bot.DeleteMessage(chatID, messageID, nil)

	// Delete the attempt from database
	_ = db.DeleteCaptchaAttempt(userID, chatID)

	// Execute the failure action
	switch action {
	case "kick":
		_, err := bot.UnbanChatMember(chatID, userID, &gotgbot.UnbanChatMemberOpts{OnlyIfBanned: false})
		if err != nil {
			log.Errorf("Failed to kick user %d: %v", userID, err)
		}
	case "ban":
		_, err := bot.BanChatMember(chatID, userID, nil)
		if err != nil {
			log.Errorf("Failed to ban user %d: %v", userID, err)
		}
	case "mute":
		// User remains muted (already muted when they joined)
		// Just log it
		log.Infof("User %d remains muted due to captcha timeout", userID)
	}
}

// captchaVerifyCallback handles captcha answer button clicks.
// Verifies if the selected answer is correct and takes appropriate action.
func (moduleStruct) captchaVerifyCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	chat := ctx.EffectiveChat
	user := query.From

	// Parse callback data: captcha_verify.{user_id}.{answer}
	parts := strings.Split(query.Data, ".")
	if len(parts) != 3 {
		_, err := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Invalid captcha data"})
		return err
	}

	targetUserID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Invalid user ID"})
		return err
	}

	// Check if this is the correct user
	if user.Id != targetUserID {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "This captcha is not for you!"})
		return err
	}

	selectedAnswer := parts[2]

	// Get the captcha attempt
	attempt, err := db.GetCaptchaAttempt(targetUserID, chat.Id)
	if err != nil || attempt == nil {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Captcha expired or not found"})
		return err
	}

	settings, _ := db.GetCaptchaSettings(chat.Id)

	// Check if answer is correct
	if selectedAnswer == attempt.Answer {
		// Correct answer - unmute the user
		_, err = chat.RestrictMember(bot, targetUserID, gotgbot.ChatPermissions{
			CanSendMessages:       true,
			CanSendPhotos:         true,
			CanSendVideos:         true,
			CanSendAudios:         true,
			CanSendDocuments:      true,
			CanSendVideoNotes:     true,
			CanSendVoiceNotes:     true,
			CanAddWebPagePreviews: true,
			CanChangeInfo:         false,
			CanInviteUsers:        true,
			CanPinMessages:        false,
			CanManageTopics:       false,
			CanSendPolls:          true,
			CanSendOtherMessages:  true,
		}, nil)

		if err != nil {
			log.Errorf("Failed to unmute user %d: %v", targetUserID, err)
			_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Failed to verify. Please contact an admin."})
			return err
		}

		// Delete the captcha message
		_, _ = bot.DeleteMessage(chat.Id, attempt.MessageID, nil)

		// Delete the attempt from database
		_ = db.DeleteCaptchaAttempt(targetUserID, chat.Id)

		// Send success message
		successMsg := fmt.Sprintf("✅ %s has been verified and can now chat!", helpers.MentionHtml(targetUserID, user.FirstName))
		sent, _ := bot.SendMessage(chat.Id, successMsg, &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})

		// Delete success message after 5 seconds
		if sent != nil {
			go func() {
				time.Sleep(5 * time.Second)
				_, _ = bot.DeleteMessage(chat.Id, sent.MessageId, nil)
			}()
		}

		// Send welcome message after successful verification
		if err = SendWelcomeMessage(bot, ctx, targetUserID, user.FirstName); err != nil {
			log.Errorf("Failed to send welcome message after captcha verification: %v", err)
		}

		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "✅ Verified successfully!"})
		return err

	} else {
		// Wrong answer - increment attempts
		attempt, err = db.IncrementCaptchaAttempts(targetUserID, chat.Id)
		if err != nil {
			_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Error processing answer"})
			return err
		}

		if attempt.Attempts >= settings.MaxAttempts {
			// Max attempts reached - execute failure action
			handleCaptchaTimeout(bot, chat.Id, targetUserID, attempt.MessageID, settings.FailureAction)

			actionText := "kicked"
			switch settings.FailureAction {
			case "ban":
				actionText = "banned"
			case "mute":
				actionText = "muted permanently"
			}

			_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
				Text:      fmt.Sprintf("❌ Wrong answer! You have been %s.", actionText),
				ShowAlert: true,
			})
			return err
		}

		remainingAttempts := settings.MaxAttempts - attempt.Attempts
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text:      fmt.Sprintf("❌ Wrong answer! %d attempts remaining.", remainingAttempts),
			ShowAlert: true,
		})
		return err
	}
}

// captchaRefreshCallback handles the refresh button for text captchas.
// Generates a new captcha image when users can't read the current one.
func (moduleStruct) captchaRefreshCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	chat := ctx.EffectiveChat
	user := query.From

	// Parse callback data: captcha_refresh.{user_id}
	parts := strings.Split(query.Data, ".")
	if len(parts) != 2 {
		_, err := query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Invalid refresh data"})
		return err
	}

	targetUserID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Invalid user ID"})
		return err
	}

	// Check if this is the correct user
	if user.Id != targetUserID {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "This captcha is not for you!"})
		return err
	}

	// Get the existing attempt
	attempt, err := db.GetCaptchaAttempt(targetUserID, chat.Id)
	if err != nil || attempt == nil {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Captcha expired or not found"})
		return err
	}

	// Delete old message
	_, _ = bot.DeleteMessage(chat.Id, attempt.MessageID, nil)

	// Generate new captcha and resend
	err = SendCaptcha(bot, ctx, targetUserID, user.FirstName)
	if err != nil {
		_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Failed to generate new captcha"})
		return err
	}

	_, err = query.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "🔄 New captcha sent!"})
	return err
}

// LoadCaptcha registers all captcha module handlers with the dispatcher.
func LoadCaptcha(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(captchaModule.moduleName, true)

	// Commands
	dispatcher.AddHandler(handlers.NewCommand("captcha", captchaModule.captchaCommand))
	dispatcher.AddHandler(handlers.NewCommand("captchamode", captchaModule.captchaModeCommand))
	dispatcher.AddHandler(handlers.NewCommand("captchatime", captchaModule.captchaTimeCommand))
	dispatcher.AddHandler(handlers.NewCommand("captchaaction", captchaModule.captchaActionCommand))

	// Callbacks
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("captcha_verify."), captchaModule.captchaVerifyCallback))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("captcha_refresh."), captchaModule.captchaRefreshCallback))

	// Start periodic cleanup of expired attempts
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			count, err := db.CleanupExpiredCaptchaAttempts()
			if err != nil {
				log.Errorf("Failed to cleanup expired captcha attempts: %v", err)
			} else if count > 0 {
				log.Infof("Cleaned up %d expired captcha attempts", count)
			}
		}
	}()
}
