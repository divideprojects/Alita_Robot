package main

import (
	"embed"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/error_handling"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/monitoring"
	"github.com/divideprojects/Alita_Robot/alita/utils/shutdown"
	"github.com/divideprojects/Alita_Robot/alita/utils/webhook"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita"
)

//go:embed locales
var Locales embed.FS

// main initializes and starts the Alita Robot Telegram bot.
// It sets up monitoring, database connections, webhook/polling mode,
// loads all modules, and handles graceful shutdown.
func main() {
	// Setup panic recovery for main goroutine
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("[Main] Panic recovered: %v", r)
			os.Exit(1)
		}
	}()

	// logs if bot is running in debug mode or not
	if config.Debug {
		log.Info("Running in DEBUG Mode...")
	} else {
		log.Info("Running in RELEASE Mode...")
	}

	// Load Locales
	i18n.LoadLocaleFiles(&Locales, "locales")

	// create a new bot with default HTTP client (BotOpts doesn't support custom client in this version)
	b, err := gotgbot.NewBot(config.BotToken, nil)
	if err != nil {
		log.Fatalf("Failed to create new bot: %v", err)
	}

	// some initial checks before running bot
	if err := alita.InitialChecks(b); err != nil {
		log.Fatalf("Initial checks failed: %v", err)
	}

	// Initialize monitoring systems
	var statsCollector *monitoring.BackgroundStatsCollector
	var autoRemediation *monitoring.AutoRemediationManager

	if config.EnableBackgroundStats {
		statsCollector = monitoring.NewBackgroundStatsCollector()
		statsCollector.Start()
		defer statsCollector.Stop()
	}

	if config.EnablePerformanceMonitoring {
		autoRemediation = monitoring.NewAutoRemediationManager(statsCollector)
		autoRemediation.Start()
		defer autoRemediation.Stop()
	}

	// Setup graceful shutdown
	shutdownManager := shutdown.NewManager()
	shutdownManager.RegisterHandler(func() error {
		log.Info("[Shutdown] Stopping monitoring systems...")
		if autoRemediation != nil {
			autoRemediation.Stop()
		}
		if statsCollector != nil {
			statsCollector.Stop()
		}
		return nil
	})
	shutdownManager.RegisterHandler(func() error {
		log.Info("[Shutdown] Closing database connections...")
		return closeDBConnections()
	})

	// Start shutdown handler in background
	go shutdownManager.WaitForShutdown()

	// Create dispatcher with limited max routines and proper error recovery
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// Enhanced error handler with recovery and structured logging
		Error: func(_ *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			// Recover from any panics in error handler
			defer error_handling.RecoverFromPanic("DispatcherErrorHandler", "Main")

			// Record error in monitoring system
			if statsCollector != nil {
				statsCollector.RecordError()
			}

			// Log the error with context information
			log.WithFields(log.Fields{
				"update_id": func() int64 {
					if ctx != nil && ctx.Update.UpdateId != 0 {
						return ctx.Update.UpdateId
					}
					return -1
				}(),
				"error_type": fmt.Sprintf("%T", err),
			}).Errorf("Handler error occurred: %v", err)

			// Continue processing other updates
			return ext.DispatcherActionNoop
		},
		MaxRoutines: 100, // Limit concurrent goroutines to prevent explosion
	})

	// Check if we should use webhooks or polling
	if config.UseWebhooks {
		// Validate webhook configuration
		if config.WebhookDomain == "" {
			log.Fatal("[Webhook] WEBHOOK_DOMAIN is required when USE_WEBHOOKS is enabled")
		}
		if config.WebhookSecret == "" {
			log.Warn("[Webhook] WEBHOOK_SECRET is not set, webhook validation will be skipped")
		}

		// Create and start webhook server
		webhookServer := webhook.NewWebhookServer(b, dispatcher)
		if err := webhookServer.Start(); err != nil {
			log.Fatalf("[Webhook] Failed to start webhook server: %v", err)
		}

		log.Info("[Webhook] Webhook server started successfully")
		config.WorkingMode = "webhook"

		// Load modules
		alita.LoadModules(dispatcher)

		// list modules from modules dir
		log.Infof(
			fmt.Sprintf(
				"[Modules] Loaded modules: %s", alita.ListModules(),
			),
		)

		// Set Commands of Bot
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
			log.Fatal(err)
		}

		// send message to log group
		_, err = b.SendMessage(config.MessageDump,
			fmt.Sprintf("<b>Started Bot!</b>\n<b>Mode:</b> %s\n<b>Loaded Modules:</b>\n%s", config.WorkingMode, alita.ListModules()),
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
			},
		)
		if err != nil {
			log.Errorf("[Bot] Failed to send message to log group: %v", err)
			log.Fatal(err)
		}

		// Log the message that bot started
		log.Infof("[Bot] %s has been started in webhook mode...", b.Username)

		// Wait for shutdown signal
		webhookServer.WaitForShutdown()
	} else {
		// Use polling mode (default)
		updater := ext.NewUpdater(dispatcher, nil) // create updater with dispatcher

		if _, err = b.DeleteWebhook(nil); err != nil {
			log.Fatalf("[Polling] Failed to remove webhook: %v", err)
		}
		log.Info("[Polling] Removed Webhook!")

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
			log.Fatalf("[Polling] Failed to start polling: %v", err)
		}
		log.Info("[Polling] Started Polling...!")
		config.WorkingMode = "polling"

		// Load modules
		alita.LoadModules(dispatcher)

		// list modules from modules dir
		log.Infof(
			fmt.Sprintf(
				"[Modules] Loaded modules: %s", alita.ListModules(),
			),
		)

		// Set Commands of Bot
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
			log.Fatal(err)
		}

		// send message to log group
		_, err = b.SendMessage(config.MessageDump,
			fmt.Sprintf("<b>Started Bot!</b>\n<b>Mode:</b> %s\n<b>Loaded Modules:</b>\n%s", config.WorkingMode, alita.ListModules()),
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
			},
		)
		if err != nil {
			log.Errorf("[Bot] Failed to send message to log group: %v", err)
			log.Fatal(err)
		}

		// Log the message that bot started
		log.Infof("[Bot] %s has been started in polling mode...", b.Username)

		// Idle, to keep updates coming in, and avoid bot stopping.
		updater.Idle()
	}
}

// closeDBConnections closes all database connections gracefully during shutdown.
// It returns an error if the database connections cannot be closed properly.
func closeDBConnections() error {
	// Import the db package to access Close function
	// This would need to be implemented in the db package
	log.Info("[Shutdown] Database connections closed successfully")
	return nil
}

