package main

// Main entry point for the Alita Telegram bot.
// Sets up configuration, loads locales, initializes the bot, and starts polling for updates.

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/lifecycle"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/scheduler"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita"
)

//go:embed locales
var Locales embed.FS

// registerLifecycleComponents registers all lifecycle components
func registerLifecycleComponents(manager *lifecycle.Manager, bot *gotgbot.Bot) error {
	// Register MongoDB lifecycle manager
	if err := manager.Register(db.GetMongoLifecycleManager()); err != nil {
		return fmt.Errorf("failed to register MongoDB lifecycle manager: %w", err)
	}

	// Register cache lifecycle manager
	if err := manager.Register(cache.GetCacheLifecycleManager()); err != nil {
		return fmt.Errorf("failed to register cache lifecycle manager: %w", err)
	}

	// Register scheduler lifecycle manager
	if err := manager.Register(scheduler.GetSchedulerLifecycleManager(bot)); err != nil {
		return fmt.Errorf("failed to register scheduler lifecycle manager: %w", err)
	}

	log.Info("All lifecycle components registered successfully")
	return nil
}

func main() {
	// main initializes and starts the Alita Telegram bot with proper lifecycle management.
	// It configures logging, loads locales, creates the bot instance,
	// performs initial checks, sets up polling, loads modules, and sends a startup message.
	// The function uses graceful shutdown handling and proper resource cleanup.

	// Create main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize lifecycle manager
	lifecycleManager := lifecycle.NewManager()

	// logs if bot is running in debug mode or not
	if config.Debug {
		log.Info("Running in DEBUG Mode...")
	} else {
		log.Info("Running in RELEASE Mode...")
	}

	// Load Locales
	i18n.LoadLocaleFiles(&Locales, "locales")

	// create a new bot with default HTTP client (BotOpts doesn't support custom client in this version)
	// BotToken is loaded from config.
	b, err := gotgbot.NewBot(config.BotToken, nil)
	if err != nil {
		log.Errorf("Failed to create new bot: %v", err)
		os.Exit(1)
	}

	// Register lifecycle components
	if err := registerLifecycleComponents(lifecycleManager, b); err != nil {
		log.Errorf("Failed to register lifecycle components: %v", err)
		os.Exit(1)
	}

	// Initialize all components
	initCtx, initCancel := context.WithTimeout(ctx, 60*time.Second)
	defer initCancel()

	if err := lifecycleManager.Initialize(initCtx); err != nil {
		log.Errorf("Failed to initialize components: %v", err)
		os.Exit(1)
	}

	// some initial checks before running bot
	alita.InitialChecks(b)

	// Create updater and dispatcher with limited max routines.
	// Dispatcher handles incoming updates and routes them to handlers.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(_ *gotgbot.Bot, _ *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: 100, // Limit concurrent goroutines to prevent explosion
	})
	updater := ext.NewUpdater(dispatcher, nil) // create updater with dispatcher

	if _, err = b.DeleteWebhook(nil); err != nil {
		log.Errorf("[Polling] Failed to remove webhook: %v", err)
		// Don't exit here, try to continue
	} else {
		log.Info("[Polling] Removed Webhook!")
	}

	// start the bot in polling mode
	err = updater.StartPolling(b,
		&ext.PollingOpts{
			DropPendingUpdates: config.DropPendingUpdates,
			GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
				AllowedUpdates: config.AllowedUpdates,
			},
		},
	)
	if err != nil {
		log.Errorf("[Polling] Failed to start polling: %v", err)
		// Cleanup and exit
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		lifecycleManager.Shutdown(shutdownCtx, lifecycle.DefaultShutdownOptions())
		os.Exit(1)
	}
	log.Info("[Polling] Started Polling...!")

	// Log the message that bot started
	log.Infof("[Bot] %s has been started...", b.Username)

	// Set custom commands for private messages.
	log.Info("Setting Custom Commands for PM...!")
	_, err = b.SetMyCommands(
		[]gotgbot.BotCommand{
			{Command: "start", Description: "Starts the Bot"},
			{Command: "help", Description: "Check Help section of bot"},
		},
		&gotgbot.SetMyCommandsOpts{
			Scope:        gotgbot.BotCommandScopeAllPrivateChats{},
			LanguageCode: "en",
		},
	)
	if err != nil {
		log.Errorf("Failed to set bot commands: %v", err)
		// Continue without commands
	}

	// Loading Modules
	alita.LoadModules(dispatcher)

	// CAPTCHA scheduler is now managed by lifecycle manager

	// List loaded modules from the modules directory.
	log.Infof(
		fmt.Sprintf(
			"[Modules] Loaded modules: %s", alita.ListModules(),
		),
	)

	// send message to log group
	_, err = b.SendMessage(config.MessageDump,
		fmt.Sprintf("<b>Started Bot!</b>\n<b>Mode:</b> %s\n<b>Loaded Modules:</b>\n%s", config.WorkingMode, alita.ListModules()),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Errorf("[Bot] Failed to send message to log group: %v", err)
		// Continue without sending startup message
	}

	// Setup graceful shutdown
	log.Info("Setting up graceful shutdown...")
	shutdownOptions := lifecycle.DefaultShutdownOptions()
	shutdownOptions.Timeout = 30 * time.Second

	// Start graceful shutdown handler
	go func() {
		if err := lifecycle.GracefulShutdown(ctx, lifecycleManager, shutdownOptions); err != nil {
			log.Errorf("Error during graceful shutdown: %v", err)
		}
	}()

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}
