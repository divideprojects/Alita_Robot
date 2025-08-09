package bot_manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/token"
)

// BotInstance represents a single bot instance with its associated resources
type BotInstance struct {
	BotID       int64                `json:"bot_id"`
	OwnerID     int64                `json:"owner_id"`
	Bot         *gotgbot.Bot         `json:"-"` // Don't serialize the bot instance
	Updater     *ext.Updater         `json:"-"` // Don't serialize the updater
	Token       string               `json:"-"` // Don't serialize the token for security
	TokenHash   string               `json:"token_hash"`
	Username    string               `json:"username"`
	Name        string               `json:"name"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	LastActivity *time.Time          `json:"last_activity"`
	WebhookURL  string               `json:"webhook_url"`
	
	// Runtime state (not persisted)
	cancelFunc  context.CancelFunc   `json:"-"`
	isRunning   bool                 `json:"-"`
	mu          sync.RWMutex         `json:"-"`
}

// BotManager manages multiple bot instances
type BotManager struct {
	instances    map[int64]*BotInstance // bot_id -> BotInstance
	mainBotID    int64                  // ID of the main bot (usually 1)
	mu           sync.RWMutex
	
	// Configuration
	maxInstances int  // Maximum number of cloned bots per user
}

// NewBotManager creates a new bot manager instance
func NewBotManager(mainBotID int64) *BotManager {
	return &BotManager{
		instances:    make(map[int64]*BotInstance),
		mainBotID:    mainBotID,
		maxInstances: 1, // Max 1 cloned bot per user
	}
}

// CreateBotInstance creates and starts a new bot instance
func (bm *BotManager) CreateBotInstance(ownerID int64, botToken string) (*BotInstance, error) {
	// Validate token format first
	if !token.IsValidTokenFormat(botToken) {
		return nil, fmt.Errorf("invalid token format")
	}

	// Extract bot ID from token
	botID, err := token.ExtractBotID(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bot ID: %w", err)
	}

	// Check if bot already exists
	bm.mu.RLock()
	if _, exists := bm.instances[botID]; exists {
		bm.mu.RUnlock()
		return nil, fmt.Errorf("bot instance with ID %d already exists", botID)
	}

	// Check user limits
	userBotCount := bm.getUserBotCount(ownerID)
	if userBotCount >= bm.maxInstances {
		bm.mu.RUnlock()
		return nil, fmt.Errorf("user has reached maximum bot limit (%d)", bm.maxInstances)
	}

	// No total system limit needed
	bm.mu.RUnlock()

	// Validate token with Telegram API
	log.Infof("[BotManager] Validating token for bot ID %d", botID)
	botInfo, err := token.ValidateTokenWithTelegram(botToken, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Create gotgbot instance
	bot, err := gotgbot.NewBot(botToken, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot instance: %w", err)
	}

	// Create bot instance
	instance := &BotInstance{
		BotID:       botID,
		OwnerID:     ownerID,
		Bot:         bot,
		Token:       botToken,
		TokenHash:   token.HashToken(botToken),
		Username:    botInfo.Username,
		Name:        botInfo.FirstName,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		WebhookURL:  fmt.Sprintf("/webhook/clone/%d", botID),
		isRunning:   false,
	}

	// Add to manager
	bm.mu.Lock()
	bm.instances[botID] = instance
	bm.mu.Unlock()

	log.Infof("[BotManager] Created bot instance @%s (ID: %d) for user %d", 
		instance.Username, instance.BotID, instance.OwnerID)

	return instance, nil
}

// StartBotInstance starts polling or webhook for a bot instance
func (bm *BotManager) StartBotInstance(botID int64, useWebhooks bool, webhookDomain string) error {
	bm.mu.RLock()
	instance, exists := bm.instances[botID]
	bm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("bot instance %d not found", botID)
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.isRunning {
		return fmt.Errorf("bot instance %d is already running", botID)
	}

	// Create context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	instance.cancelFunc = cancel

	// Create updater with dispatcher
	// TODO: For now, we'll create a simple dispatcher for each bot instance
	// In a full implementation, we'd need to register the same handlers as the main bot
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{})
	updater := ext.NewUpdater(dispatcher, nil)
	instance.Updater = updater

	// TODO: Register handlers for this bot instance
	// This would involve setting up the same handlers as the main bot
	// but with bot-specific context

	// Start the bot
	if useWebhooks {
		// Webhook mode - will be handled by the main webhook router
		webhookURL := fmt.Sprintf("%s%s", webhookDomain, instance.WebhookURL)
		log.Infof("[BotManager] Setting webhook for bot @%s (ID: %d) to %s", 
			instance.Username, instance.BotID, webhookURL)
		
		_, err := instance.Bot.SetWebhook(webhookURL, nil)
		if err != nil {
			return fmt.Errorf("failed to set webhook: %w", err)
		}
	} else {
		// Polling mode
		log.Infof("[BotManager] Starting polling for bot @%s (ID: %d)", 
			instance.Username, instance.BotID)

		go func() {
			err := updater.StartPolling(instance.Bot, &ext.PollingOpts{
				DropPendingUpdates: true,
				GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
					Timeout: 9,
					RequestOpts: &gotgbot.RequestOpts{
						Timeout: time.Second * 10,
					},
				},
			})
			if err != nil {
				log.Errorf("[BotManager] Polling error for bot %d: %v", botID, err)
				instance.mu.Lock()
				instance.isRunning = false
				instance.mu.Unlock()
			}
		}()
	}

	instance.isRunning = true
	instance.LastActivity = &[]time.Time{time.Now()}[0]

	log.Infof("[BotManager] Bot instance @%s (ID: %d) started successfully", 
		instance.Username, instance.BotID)

	return nil
}

// StopBotInstance stops a running bot instance
func (bm *BotManager) StopBotInstance(botID int64) error {
	bm.mu.RLock()
	instance, exists := bm.instances[botID]
	bm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("bot instance %d not found", botID)
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	if !instance.isRunning {
		return fmt.Errorf("bot instance %d is not running", botID)
	}

	// Cancel the context to stop all operations
	if instance.cancelFunc != nil {
		instance.cancelFunc()
	}

	// Stop the updater
	if instance.Updater != nil {
		if stopErr := instance.Updater.Stop(); stopErr != nil {
			log.Warnf("[BotManager] Failed to stop updater for bot %d: %v", botID, stopErr)
		}
	}

	// Delete webhook if it was set
	_, err := instance.Bot.DeleteWebhook(&gotgbot.DeleteWebhookOpts{})
	if err != nil {
		log.Warnf("[BotManager] Failed to delete webhook for bot %d: %v", botID, err)
	}

	instance.isRunning = false
	instance.IsActive = false
	instance.UpdatedAt = time.Now()

	log.Infof("[BotManager] Bot instance @%s (ID: %d) stopped", 
		instance.Username, instance.BotID)

	return nil
}

// RemoveBotInstance completely removes a bot instance
func (bm *BotManager) RemoveBotInstance(botID int64) error {
	// Stop the instance first
	if err := bm.StopBotInstance(botID); err != nil {
		// Log the error but continue with removal
		log.Warnf("[BotManager] Failed to stop bot %d during removal: %v", botID, err)
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	instance, exists := bm.instances[botID]
	if !exists {
		return fmt.Errorf("bot instance %d not found", botID)
	}

	delete(bm.instances, botID)

	log.Infof("[BotManager] Bot instance @%s (ID: %d) removed", 
		instance.Username, instance.BotID)

	return nil
}

// GetBotInstance returns a bot instance by ID
func (bm *BotManager) GetBotInstance(botID int64) (*BotInstance, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	instance, exists := bm.instances[botID]
	return instance, exists
}

// GetUserBots returns all bot instances owned by a specific user
func (bm *BotManager) GetUserBots(ownerID int64) []*BotInstance {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var userBots []*BotInstance
	for _, instance := range bm.instances {
		if instance.OwnerID == ownerID {
			userBots = append(userBots, instance)
		}
	}

	return userBots
}

// GetAllBots returns all bot instances
func (bm *BotManager) GetAllBots() []*BotInstance {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var allBots []*BotInstance
	for _, instance := range bm.instances {
		allBots = append(allBots, instance)
	}

	return allBots
}

// UpdateBotActivity updates the last activity timestamp for a bot
func (bm *BotManager) UpdateBotActivity(botID int64) {
	bm.mu.RLock()
	instance, exists := bm.instances[botID]
	bm.mu.RUnlock()

	if !exists {
		return
	}

	instance.mu.Lock()
	now := time.Now()
	instance.LastActivity = &now
	instance.UpdatedAt = now
	instance.mu.Unlock()
}

// GetStats returns statistics about the bot manager
func (bm *BotManager) GetStats() map[string]interface{} {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	totalBots := len(bm.instances)
	activeBots := 0
	runningBots := 0
	userCounts := make(map[int64]int)

	for _, instance := range bm.instances {
		if instance.IsActive {
			activeBots++
		}
		
		instance.mu.RLock()
		if instance.isRunning {
			runningBots++
		}
		instance.mu.RUnlock()

		userCounts[instance.OwnerID]++
	}

	return map[string]interface{}{
		"total_bots":      totalBots,
		"active_bots":     activeBots,
		"running_bots":    runningBots,
		"max_per_user":    bm.maxInstances,
		"users_with_bots": len(userCounts),
		"main_bot_id":     bm.mainBotID,
	}
}

// getUserBotCount returns the number of bots owned by a user (internal helper)
func (bm *BotManager) getUserBotCount(ownerID int64) int {
	count := 0
	for _, instance := range bm.instances {
		if instance.OwnerID == ownerID {
			count++
		}
	}
	return count
}

// IsMainBot checks if the given bot ID is the main bot
func (bm *BotManager) IsMainBot(botID int64) bool {
	return botID == bm.mainBotID
}

// Shutdown gracefully shuts down all bot instances
func (bm *BotManager) Shutdown() error {
	log.Info("[BotManager] Shutting down all bot instances...")

	bm.mu.RLock()
	botIDs := make([]int64, 0, len(bm.instances))
	for botID := range bm.instances {
		botIDs = append(botIDs, botID)
	}
	bm.mu.RUnlock()

	// Stop all instances
	for _, botID := range botIDs {
		if err := bm.StopBotInstance(botID); err != nil {
			log.Warnf("[BotManager] Failed to stop bot %d during shutdown: %v", botID, err)
		}
	}

	log.Info("[BotManager] All bot instances shut down")
	return nil
}