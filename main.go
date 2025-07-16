// Package main is the entry point for the Alita Telegram bot.
//
// This package initializes the bot configuration, loads internationalization
// locales, sets up the Telegram bot polling, and starts all bot modules.
// It handles the complete bot lifecycle from startup to shutdown.
package main

// Main entry point for the Alita Telegram bot.
// Sets up configuration, loads locales, initializes the bot, and starts polling for updates.

import (
	"embed"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita"
)

//go:embed locales
var Locales embed.FS

func main() {
	// main initializes and starts the Alita Telegram bot.
	// It configures logging, loads locales, creates the bot instance,
	// performs initial checks, sets up polling, loads modules, and sends a startup message.
	// The function blocks at the end to keep the bot running.
	// All critical startup errors are logged and cause termination.

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
		panic("failed to create new bot: " + err.Error())
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
		log.Fatal(err)
	}

	// Loading Modules
	alita.LoadModules(dispatcher)

	// Start CAPTCHA scheduler
	alita.StartCaptchaScheduler(b)

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
		log.Fatal(err)
	}

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}
