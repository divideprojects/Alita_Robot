package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// CaptchaScheduler handles expired CAPTCHA challenges
type CaptchaScheduler struct {
	bot      *gotgbot.Bot
	interval time.Duration
	stopChan chan struct{}
	running  bool
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewCaptchaScheduler creates a new CAPTCHA scheduler
func NewCaptchaScheduler(bot *gotgbot.Bot, interval time.Duration) *CaptchaScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &CaptchaScheduler{
		bot:      bot,
		interval: interval,
		stopChan: make(chan struct{}),
		running:  false,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the scheduler's background processing
func (cs *CaptchaScheduler) Start() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.running {
		log.Warn("CAPTCHA scheduler is already running")
		return
	}

	log.Info("Starting CAPTCHA scheduler...")
	cs.running = true

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
			cs.mu.Lock()
			cs.running = false
			cs.mu.Unlock()
			return
		case <-cs.ctx.Done():
			log.Info("CAPTCHA scheduler stopped due to context cancellation")
			cs.mu.Lock()
			cs.running = false
			cs.mu.Unlock()
			return
		}
	}
}

// Stop stops the scheduler
func (cs *CaptchaScheduler) Stop() {
	cs.mu.RLock()
	running := cs.running
	cs.mu.RUnlock()

	if !running {
		log.Debug("CAPTCHA scheduler is already stopped")
		return
	}

	log.Info("Stopping CAPTCHA scheduler...")
	
	// Use select to prevent blocking
	select {
	case cs.stopChan <- struct{}{}:
		// Signal sent successfully
	default:
		// Channel is full or closed, use context cancellation
		cs.cancel()
	}

	// Wait for scheduler to actually stop with timeout
	timeout := time.After(5 * time.Second)
	for {
		cs.mu.RLock()
		running := cs.running
		cs.mu.RUnlock()

		if !running {
			break
		}

		select {
		case <-timeout:
			log.Warn("CAPTCHA scheduler stop timeout, forcing cancellation")
			cs.cancel()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// processExpiredChallenges handles all expired CAPTCHA challenges
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

// processExpiredChallenge handles a single expired challenge
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

// kickUser kicks a user from the chat
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

// unmuteUser unmutes a user (auto-unmute)
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

// isUserStillMuted checks if a user is still muted in the chat
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

// IsRunning returns true if the scheduler is currently running
func (cs *CaptchaScheduler) IsRunning() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.running
}

// StartCaptchaScheduler starts the global CAPTCHA scheduler
func StartCaptchaScheduler(bot *gotgbot.Bot) *CaptchaScheduler {
	scheduler := NewCaptchaScheduler(bot, 30*time.Second) // Check every 30 seconds
	go scheduler.Start()
	return scheduler
}

// SchedulerLifecycleManager manages the CAPTCHA scheduler lifecycle
type SchedulerLifecycleManager struct {
	scheduler *CaptchaScheduler
	bot       *gotgbot.Bot
}

// NewSchedulerLifecycleManager creates a new scheduler lifecycle manager
func NewSchedulerLifecycleManager(bot *gotgbot.Bot) *SchedulerLifecycleManager {
	return &SchedulerLifecycleManager{
		bot: bot,
	}
}

// Name returns the component name
func (s *SchedulerLifecycleManager) Name() string {
	return "scheduler"
}

// Priority returns the shutdown priority
func (s *SchedulerLifecycleManager) Priority() int {
	return 30 // Lower priority for shutdown
}

// Initialize starts the scheduler
func (s *SchedulerLifecycleManager) Initialize(ctx context.Context) error {
	log.Info("Initializing CAPTCHA scheduler...")
	
	if s.bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	s.scheduler = NewCaptchaScheduler(s.bot, 30*time.Second)
	go s.scheduler.Start()

	// Wait a moment to ensure scheduler starts
	time.Sleep(100 * time.Millisecond)

	if !s.scheduler.IsRunning() {
		return fmt.Errorf("scheduler failed to start")
	}

	log.Info("CAPTCHA scheduler initialized successfully")
	return nil
}

// Shutdown stops the scheduler
func (s *SchedulerLifecycleManager) Shutdown(ctx context.Context) error {
	log.Info("Shutting down CAPTCHA scheduler...")
	
	if s.scheduler == nil {
		log.Debug("Scheduler is nil, nothing to shutdown")
		return nil
	}

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Stop the scheduler in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.scheduler.Stop()
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-done:
		log.Info("CAPTCHA scheduler shutdown completed")
		return nil
	case <-shutdownCtx.Done():
		log.Error("CAPTCHA scheduler shutdown timed out")
		return fmt.Errorf("scheduler shutdown timeout")
	}
}

// HealthCheck verifies the scheduler is running
func (s *SchedulerLifecycleManager) HealthCheck(ctx context.Context) error {
	if s.scheduler == nil {
		return fmt.Errorf("scheduler is not initialized")
	}

	if !s.scheduler.IsRunning() {
		return fmt.Errorf("scheduler is not running")
	}

	if s.bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	return nil
}

// GetSchedulerLifecycleManager returns a new scheduler lifecycle manager
func GetSchedulerLifecycleManager(bot *gotgbot.Bot) *SchedulerLifecycleManager {
	return NewSchedulerLifecycleManager(bot)
}
