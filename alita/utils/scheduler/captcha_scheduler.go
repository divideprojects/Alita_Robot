package scheduler

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// CaptchaScheduler handles expired CAPTCHA challenges.
// It runs as a background service to process users who failed to complete CAPTCHA verification within the allowed time.
type CaptchaScheduler struct {
	bot      *gotgbot.Bot
	interval time.Duration
	stopChan chan bool
}

// NewCaptchaScheduler creates a new CAPTCHA scheduler.
// The scheduler will check for expired challenges at the specified interval.
func NewCaptchaScheduler(bot *gotgbot.Bot, interval time.Duration) *CaptchaScheduler {
	return &CaptchaScheduler{
		bot:      bot,
		interval: interval,
		stopChan: make(chan bool),
	}
}

// Start begins the scheduler's background processing.
// It runs in a loop, processing expired challenges at regular intervals until stopped.
func (cs *CaptchaScheduler) Start() {
	log.Info("Starting CAPTCHA scheduler...")

	ticker := time.NewTicker(cs.interval)
	defer ticker.Stop()

	// Process immediately on start
	cs.processExpiredChallenges()

	for {
		select {
		case <-ticker.C:
			cs.processExpiredChallenges()
		case <-cs.stopChan:
			log.Info("CAPTCHA scheduler stopped")
			return
		}
	}
}

// Stop stops the scheduler.
// It sends a stop signal to the background goroutine.
func (cs *CaptchaScheduler) Stop() {
	cs.stopChan <- true
}

// processExpiredChallenges handles all expired CAPTCHA challenges.
// It retrieves expired challenges from the database and processes each one according to chat settings.
func (cs *CaptchaScheduler) processExpiredChallenges() {
	expired, err := db.GetExpiredCaptchaChallenges()
	if err != nil {
		log.Errorf("CAPTCHA Scheduler: Failed to get expired challenges: %v", err)
		return
	}

	if len(expired) == 0 {
		return
	}

	log.Infof("CAPTCHA Scheduler: Processing %d expired challenges", len(expired))

	for _, challenge := range expired {
		cs.processExpiredChallenge(challenge)
	}

	// Clean up processed challenges
	err = db.CleanupExpiredChallenges()
	if err != nil {
		log.Errorf("CAPTCHA Scheduler: Failed to cleanup expired challenges: %v", err)
	}
}

// processExpiredChallenge handles a single expired challenge.
// It determines the appropriate action (kick or unmute) based on chat settings.
func (cs *CaptchaScheduler) processExpiredChallenge(challenge *db.CaptchaChallenge) {
	chatID := challenge.ChatID
	userID := challenge.UserID

	// Get CAPTCHA settings for this chat
	captchaSettings := db.GetCaptchaSettings(chatID)

	// Determine action based on settings
	if captchaSettings.KickEnabled {
		// Kick the user
		cs.kickUser(chatID, userID)
	} else if captchaSettings.MuteTime > 0 {
		// Auto-unmute the user
		cs.unmuteUser(chatID, userID)
	}
	// If neither kick nor auto-unmute is enabled, user remains muted

	log.Infof("CAPTCHA Scheduler: Processed expired challenge for user %d in chat %d", userID, chatID)
}

// kickUser kicks a user from the chat.
// It bans the user temporarily and then immediately unbans them, effectively kicking them.
func (cs *CaptchaScheduler) kickUser(chatID, userID int64) {
	// Check if user is still in the chat and hasn't been manually unmuted
	if !cs.isUserStillMuted(chatID, userID) {
		return
	}

	chat := &gotgbot.Chat{Id: chatID}
	_, err := chat.BanMember(cs.bot, userID, nil)
	if err != nil {
		log.Errorf("CAPTCHA Scheduler: Failed to kick user %d from chat %d: %v", userID, chatID, err)
		return
	}

	// Unban immediately to allow rejoining (this makes it a kick, not a ban)
	time.Sleep(1 * time.Second)
	_, err = chat.UnbanMember(cs.bot, userID, nil)
	if err != nil {
		log.Errorf("CAPTCHA Scheduler: Failed to unban user %d from chat %d: %v", userID, chatID, err)
	}

	log.Infof("CAPTCHA Scheduler: Kicked user %d from chat %d for not solving CAPTCHA", userID, chatID)
}

// unmuteUser unmutes a user (auto-unmute).
// It restores the user's permissions to send messages and media after CAPTCHA timeout.
func (cs *CaptchaScheduler) unmuteUser(chatID, userID int64) {
	// Check if user is still in the chat and muted
	if !cs.isUserStillMuted(chatID, userID) {
		return
	}

	chat := &gotgbot.Chat{Id: chatID}
	_, err := chat.RestrictMember(cs.bot, userID, gotgbot.ChatPermissions{
		CanSendMessages:       true,
		CanSendPhotos:         true,
		CanSendVideos:         true,
		CanSendAudios:         true,
		CanSendDocuments:      true,
		CanSendVideoNotes:     true,
		CanSendVoiceNotes:     true,
		CanAddWebPagePreviews: true,
		CanChangeInfo:         false,
		CanInviteUsers:        false,
		CanPinMessages:        false,
		CanManageTopics:       false,
		CanSendPolls:          true,
		CanSendOtherMessages:  true,
	}, nil)
	if err != nil {
		log.Errorf("CAPTCHA Scheduler: Failed to unmute user %d in chat %d: %v", userID, chatID, err)
		return
	}

	log.Infof("CAPTCHA Scheduler: Auto-unmuted user %d in chat %d", userID, chatID)
}

// isUserStillMuted checks if a user is still muted in the chat.
// It verifies that the user is present and has restricted status before processing the challenge.
func (cs *CaptchaScheduler) isUserStillMuted(chatID, userID int64) bool {
	// Check if user is still in the chat
	member, err := cs.bot.GetChatMember(chatID, userID, nil)
	if err != nil {
		log.Debugf("CAPTCHA Scheduler: Failed to get chat member %d in chat %d: %v", userID, chatID, err)
		return false
	}

	// Check if user is restricted (muted)
	chatMember := member.MergeChatMember()
	if chatMember.Status == "restricted" {
		// User is restricted, assume they are muted for CAPTCHA purposes
		return true
	}

	// User is not restricted or not in chat
	return false
}

// StartCaptchaScheduler starts the global CAPTCHA scheduler.
// It creates a new scheduler with a 30-second check interval and starts it in a goroutine.
func StartCaptchaScheduler(bot *gotgbot.Bot) *CaptchaScheduler {
	scheduler := NewCaptchaScheduler(bot, 30*time.Second) // Check every 30 seconds
	go scheduler.Start()
	return scheduler
}
