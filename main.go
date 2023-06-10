package main

import (
	"crypto/sha512"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/parsemode"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	bot "github.com/divideprojects/Alita_Robot/alita"
)

//go:embed locales
var Locales embed.FS

func main() {
	// logs if bot is running in debug mode or not
	if config.Debug {
		log.Info("Running in DEBUG Mode...")
	} else {
		log.Info("Running in RELEASE Mode...")
	}

	// Load Locales
	i18n.LoadLocaleFiles(&Locales, "locales")

	// create a new bot
	b, err := gotgbot.NewBot(
		config.BotToken,
		&gotgbot.BotOpts{
			DefaultRequestOpts: &gotgbot.RequestOpts{
				APIURL: config.ApiServer,
			},
		},
	)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// some initial checks before running bot
	bot.InitialChecks(b)

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})

	// extract dispatcher from updater
	dispatcher := updater.Dispatcher

	// if webhooks are enabled in config
	if config.EnableWebhook {
		log.Info("[Webhook] Starting webhook...")
		config.WorkingMode = "webhook"

		webhookOpts := ext.WebhookOpts{
			ListenAddr:  fmt.Sprintf("0.0.0.0:%d", config.WebhookPort),
			SecretToken: config.SecretToken,
		}

		// Start the webhook
		// We use the token as the urlPath for the webhook, as using a secret ensures that strangers aren't crafting fake updates.
		err = updater.StartWebhook(b, config.BotToken, webhookOpts)
		if err != nil {
			log.Fatalf("[Webhook] Failed to start webhook: %v", err)
		}
		log.Info("[Webhook] Webhook started Successfully!")

		// Get the full webhook URL that we are expecting to receive updates at.

		// Set Webhook
		err = updater.SetAllBotWebhooks(config.WebhookURL, &gotgbot.SetWebhookOpts{
			MaxConnections:     100,
			DropPendingUpdates: true,
			SecretToken:        webhookOpts.SecretToken,
		})
		if err != nil {
			log.Fatalf("[Webhook] Failed to set webhook: %v", err)
		}
		log.Infof("[Webhook] Set Webhook to: %s", config.WebhookURL)

	} else {
		if _, err = b.DeleteWebhook(nil); err != nil {
			log.Fatalf("[Polling] Failed to remove webhook: %v", err)
		}
		log.Info("[Polling] Removed Webhook!")

		err = updater.StartPolling(b,
			&ext.PollingOpts{
				DropPendingUpdates: config.DropPendingUpdates,
				GetUpdatesOpts: gotgbot.GetUpdatesOpts{
					AllowedUpdates: config.AllowedUpdates,
				},
			},
		)
		if err != nil {
			log.Fatalf("[Polling] Failed to start polling: %v", err)
		}
		log.Info("[Polling] Started Polling...!")

	}

	// list modules from modules dir
	log.Infof(
		fmt.Sprintf(
			"[Modules] Loaded modules: %s", bot.ListModules(),
		),
	)

	// Log the message that bot started
	log.Infof("[Bot] %s has been started...", b.Username)

	// send message to log group
	_, err = b.SendMessage(config.MessageDump,
		fmt.Sprintf("<b>Started Bot!</b>\n<b>Mode:</b> %s\n<b>Loaded Modules:</b>\n%s", config.WorkingMode, bot.ListModules()),
		&gotgbot.SendMessageOpts{
			ParseMode: parsemode.HTML,
		},
	)
	if err != nil {
		log.Errorf("[Bot] Failed to send message to log group: %v", err)
		log.Fatal(err)
	}

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

	// Loading Modules
	bot.LoadModules(dispatcher)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

// function to handle dispatcher errors
func dispatcherErrorHandler(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
	chat := ctx.EffectiveChat
	tgErr := err.(*gotgbot.TelegramError)

	// these two just makes sure that errors are not logged and passed
	// as these are predefined by library
	if err == ext.ContinueGroups {
		return ext.DispatcherActionContinueGroups
	}

	if err == ext.EndGroups {
		return ext.DispatcherActionEndGroups
	}

	// if bot is not able to send any message to chat, it will leave the group
	if tgErr.Description == "Bad Request: have no rights to send a message" {
		_, _ = b.LeaveChat(chat.Id, nil)
		return ext.DispatcherActionEndGroups
	}

	update := ctx.Update
	uMsg := update.Message
	errorJson, _ := json.MarshalIndent(tgErr, "", "  ")
	updateJson, _ := json.Marshal(update)

	hash := func() string {
		// Generate a new Sha1 Hash
		shaHash := sha512.New()
		shaHash.Write([]byte(string(errorJson) + string(updateJson)))
		return hex.EncodeToString(shaHash.Sum(nil))
	}()

	pasted, logUrl := helpers.PasteToNekoBin("Error Report" + string(errorJson) + "\n" + string(updateJson) + "\n" + tgErr.Error())

	if pasted {
		// Send Message to Log Group
		_, _ = b.SendMessage(
			config.MessageDump,
			"⚠️ An ERROR Occurred ⚠️\n\n"+
				"An exception was raised while handling an update."+
				"\n\n"+
				fmt.Sprintf("<b>Error ID:</b> <code>%s</code>", hash)+
				"\n"+
				fmt.Sprintf("<b>Chat ID:</b> <code>%d</code>", uMsg.Chat.Id)+
				"\n"+
				fmt.Sprintf("<b>Command:</b> <code>%s</code>", uMsg.Text)+
				"\n"+
				fmt.Sprintf("<b>Error Log:</b> https://nekobin.com/%s", logUrl)+
				"\n\n"+
				"Please Check logs ASAP!",
			parsemode.Shtml(),
		)
	} else {
		_, _ = b.SendMessage(
			config.MessageDump,
			"Failed to paste error message to nekobin, please check logs!"+
				fmt.Sprintf("\n<b>Error ID:</b> <code>%s</code>", hash),
			parsemode.Shtml(),
		)
	}

	// log stuff
	log.WithFields(
		log.Fields{
			"ErrorId":       hash,
			"TelegramError": string(errorJson),
			"Update":        string(updateJson),
			"LogURL":        logUrl,
		},
	).Error(
		tgErr.Error(),
	)

	return ext.DispatcherActionNoop
}
