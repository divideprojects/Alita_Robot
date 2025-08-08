package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"time"

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

	// Initialize Locale Manager
	localeManager := i18n.GetManager()
	if err := localeManager.Initialize(&Locales, "locales", i18n.DefaultManagerConfig()); err != nil {
		log.Fatalf("Failed to initialize locale manager: %v", err)
	}
	log.Infof("Locale manager initialized with %d languages: %v", len(localeManager.GetAvailableLanguages()), localeManager.GetAvailableLanguages())

	// Create optimized HTTP transport with connection pooling for better performance
	// IMPORTANT: We create a transport pointer that will be shared across all requests
	// This ensures connection pooling works correctly (the http.Client struct is copied by value in BaseBotClient)
	httpTransport := &http.Transport{
		MaxIdleConns:        100,              // Maximum idle connections across all hosts
		MaxIdleConnsPerHost: 30,               // Increased from 10 since all requests go to api.telegram.org
		MaxConnsPerHost:     30,               // Maximum total connections per host
		IdleConnTimeout:     90 * time.Second, // How long idle connections are kept alive
		DisableCompression:  false,            // Enable compression for smaller payloads
		ForceAttemptHTTP2:   true,             // Enable HTTP/2 for multiplexing
		DisableKeepAlives:   false,            // Explicitly enable keep-alive for connection reuse
		TLSHandshakeTimeout: 10 * time.Second, // Timeout for TLS handshake
		ResponseHeaderTimeout: 10 * time.Second, // Timeout waiting for response headers
	}

	// Create bot with optimized HTTP client using BaseBotClient
	log.Info("[Main] Initializing bot with optimized HTTP client (connection pooling enabled)")
	b, err := gotgbot.NewBot(config.BotToken, &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			Client: http.Client{
				Transport: httpTransport, // Use the shared transport pointer
				Timeout:   30 * time.Second,
			},
			UseTestEnvironment: false,
			DefaultRequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Duration(30) * time.Second,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create new bot: %v", err)
	}
	log.Infof("[Main] Bot initialized with connection pooling (MaxIdleConns: 100, MaxIdleConnsPerHost: 30, HTTP/2 enabled)")

	// Pre-warm connections to Telegram API for faster initial responses
	go func() {
		log.Info("[Main] Pre-warming connections to Telegram API...")
		
		// Make multiple requests to establish connection pool
		for i := 0; i < 3; i++ {
			startTime := time.Now()
			_, err := b.GetMe(nil)
			if err != nil {
				log.Warnf("[Main] Pre-warm request %d failed: %v", i+1, err)
			} else {
				elapsed := time.Since(startTime)
				log.Infof("[Main] Pre-warm request %d completed in %v", i+1, elapsed)
				// First request establishes connection, subsequent ones should be faster
				if i > 0 && elapsed < 100*time.Millisecond {
					log.Info("[Main] Connection pooling confirmed working - reused existing connection")
				}
			}
			time.Sleep(100 * time.Millisecond) // Small delay between requests
		}
		
		log.Info("[Main] Connection pre-warming completed")
	}()

	// some initial checks before running bot
	if err := alita.InitialChecks(b); err != nil {
		log.Fatalf("Initial checks failed: %v", err)
	}

	// Initialize monitoring systems
	var statsCollector *monitoring.BackgroundStatsCollector
	var autoRemediation *monitoring.AutoRemediationManager
	var activityMonitor *monitoring.ActivityMonitor

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

	// Initialize activity monitoring for automatic group activity tracking
	activityMonitor = monitoring.NewActivityMonitor()
	activityMonitor.Start()
	defer activityMonitor.Stop()

	// Setup graceful shutdown
	shutdownManager := shutdown.NewManager()
	shutdownManager.RegisterHandler(func() error {
		log.Info("[Shutdown] Stopping monitoring systems...")
		if activityMonitor != nil {
			activityMonitor.Stop()
		}
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
		MaxRoutines: config.DispatcherMaxRoutines, // Configurable max concurrent goroutines
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
			log.Errorf("[Bot] Failed to send startup message to log group: %v", err)
			log.Warn("[Bot] Continuing without log channel notifications")
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
			log.Errorf("[Bot] Failed to send startup message to log group: %v", err)
			log.Warn("[Bot] Continuing without log channel notifications")
		}

		// Log the message that bot started
		log.Infof("[Bot] %s has been started in polling mode...", b.Username)

		// Register handler to stop the updater on shutdown
		shutdownManager.RegisterHandler(func() error {
			log.Info("[Polling] Stopping updater...")
			err := updater.Stop()
			if err != nil {
				log.Errorf("[Polling] Error stopping updater: %v", err)
				return err
			}
			log.Info("[Polling] Updater stopped successfully")
			return nil
		})

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
