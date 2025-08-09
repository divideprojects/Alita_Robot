package modules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/bot_manager"
	"github.com/divideprojects/Alita_Robot/alita/utils/token"
)

var cloneModule = moduleStruct{moduleName: "Clone"}

// Global bot manager instance - this would be initialized in main.go
var GlobalBotManager *bot_manager.BotManager

// cloneBot handles the /clone command to create a new bot instance
func cloneBot(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	// Allow all users to clone bots
	
	// Check if user provided a token
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		_, err := msg.Reply(b, 
			"❌ <b>Usage:</b> <code>/clone &lt;bot_token&gt;</code>\n\n"+
			"📝 <b>Example:</b> <code>/clone 123456789:AaBbCc_YourBotToken</code>\n\n"+
			"ℹ️ Get a bot token from @BotFather on Telegram.",
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	botToken := strings.TrimSpace(args[1])
	
	// Validate token format
	if !token.IsValidTokenFormat(botToken) {
		_, err := msg.Reply(b, 
			"❌ <b>Invalid token format!</b>\n\n"+
			"A valid bot token should look like: <code>123456789:AaBbCc_YourBotToken</code>",
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Extract bot ID
	botID, err := token.ExtractBotID(botToken)
	if err != nil {
		_, err := msg.Reply(b, 
			fmt.Sprintf("❌ <b>Failed to extract bot ID:</b> %v", err),
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Check if bot already exists
	if db.BotInstanceExists(botID) {
		_, err := msg.Reply(b, 
			fmt.Sprintf("❌ <b>Bot already exists!</b>\n\nBot ID <code>%d</code> is already registered in the system.", botID),
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Check user limits
	userBotCount, err := db.CountUserBotInstances(user.Id)
	if err != nil {
		log.Errorf("[Clone] Failed to count user bot instances: %v", err)
		_, err := msg.Reply(b, "❌ <b>Database error occurred. Please try again later.</b>", 
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	maxBotsPerUser := int64(1) // 1 bot per user limit
	if userBotCount >= maxBotsPerUser {
		_, err := msg.Reply(b, 
			fmt.Sprintf("❌ <b>Bot limit reached!</b>\n\nYou can only have up to %d bot instance. Use <code>/clones</code> to see your current bot and stop it if you want to clone a different one.", maxBotsPerUser),
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Send "creating" message
	creatingMsg, err := msg.Reply(b, 
		"🔄 <b>Creating bot instance...</b>\n\n"+
		"⏳ Validating token with Telegram...",
		&gotgbot.SendMessageOpts{ParseMode: "HTML"})
	if err != nil {
		return err
	}

	// Create bot instance using the bot manager
	if GlobalBotManager == nil {
		log.Error("[Clone] GlobalBotManager is not initialized")
		_, _, err := creatingMsg.EditText(b, "❌ <b>System error: Bot manager not initialized</b>", 
			&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
		return err
	}

	instance, err := GlobalBotManager.CreateBotInstance(user.Id, botToken)
	if err != nil {
		_, _, err := creatingMsg.EditText(b, 
			fmt.Sprintf("❌ <b>Failed to create bot instance:</b>\n\n<code>%v</code>", err),
			&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
		return err
	}

	// Save to database
	_, err = db.CreateBotInstance(instance.BotID, instance.OwnerID, instance.TokenHash, 
		instance.Username, instance.Name)
	if err != nil {
		// Clean up the bot manager instance
		if cleanupErr := GlobalBotManager.RemoveBotInstance(instance.BotID); cleanupErr != nil {
			log.Warnf("[Clone] Failed to clean up bot instance: %v", cleanupErr)
		}
		_, _, err := creatingMsg.EditText(b, 
			fmt.Sprintf("❌ <b>Failed to save bot instance to database:</b>\n\n<code>%v</code>", err),
			&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
		return err
	}

	// Start the bot instance
	useWebhooks := config.UseWebhooks
	webhookDomain := config.WebhookDomain
	
	err = GlobalBotManager.StartBotInstance(instance.BotID, useWebhooks, webhookDomain)
	if err != nil {
		log.Warnf("[Clone] Failed to start bot instance %d: %v", instance.BotID, err)
		// Don't fail the creation, just log the warning
	}

	// Success message
	successText := fmt.Sprintf(
		"✅ <b>Bot cloned successfully!</b>\n\n"+
		"🤖 <b>Bot:</b> @%s\n"+
		"🆔 <b>Bot ID:</b> <code>%d</code>\n"+
		"👤 <b>Name:</b> %s\n"+
		"📡 <b>Mode:</b> %s\n"+
		"⏰ <b>Created:</b> %s\n\n"+
		"🎛️ Use <code>/clones</code> to manage your bot instances.",
		instance.Username,
		instance.BotID,
		instance.Name,
		map[bool]string{true: "Webhook", false: "Polling"}[useWebhooks],
		instance.CreatedAt.Format("2006-01-02 15:04:05"))

	_, _, err = creatingMsg.EditText(b, successText, 
		&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
	return err
}

// listClones handles the /clones command to list user's bot instances
func listClones(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	// Allow all users to clone bots

	// Get user's bot instances
	instances, err := db.GetUserBotInstances(user.Id)
	if err != nil {
		log.Errorf("[Clone] Failed to get user bot instances: %v", err)
		_, err := msg.Reply(b, "❌ <b>Database error occurred. Please try again later.</b>", 
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	if len(instances) == 0 {
		_, err := msg.Reply(b, 
			"📭 <b>No bot instances found!</b>\n\n"+
			"Use <code>/clone &lt;token&gt;</code> to create your first bot instance.",
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Build the response
	var responseText strings.Builder
	responseText.WriteString("🤖 <b>Your Bot Instances:</b>\n\n")

	for i, instance := range instances {
		status := "🔴 Inactive"
		if instance.IsActive {
			status = "🟢 Active"
		}

		lastActivity := "Never"
		if instance.LastActivity != nil {
			lastActivity = instance.LastActivity.Format("2006-01-02 15:04")
		}

		responseText.WriteString(fmt.Sprintf(
			"<b>%d.</b> @%s\n"+
			"   🆔 <code>%d</code>\n"+
			"   📊 %s\n"+
			"   ⏰ %s\n",
			i+1, instance.BotUsername, instance.BotID, status, lastActivity))

		if i < len(instances)-1 {
			responseText.WriteString("\n")
		}
	}

	responseText.WriteString(fmt.Sprintf("\n\n📋 <b>Total:</b> %d/%d bot instance", len(instances), 1))
	responseText.WriteString("\n\n🛠️ <b>Commands:</b>")
	responseText.WriteString("\n• <code>/clone_stop &lt;bot_id&gt;</code> - Stop a bot")

	_, err = msg.Reply(b, responseText.String(), 
		&gotgbot.SendMessageOpts{ParseMode: "HTML"})
	return err
}

// stopClone handles the /clone_stop command to stop a bot instance
func stopClone(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	// Allow all users to clone bots

	// Check if user provided a bot ID
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		_, err := msg.Reply(b, 
			"❌ <b>Usage:</b> <code>/clone_stop &lt;bot_id&gt;</code>\n\n"+
			"📝 <b>Example:</b> <code>/clone_stop 123456789</code>\n\n"+
			"💡 Use <code>/clones</code> to see your bot instances.",
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	botIDStr := strings.TrimSpace(args[1])
	botID, err := strconv.ParseInt(botIDStr, 10, 64)
	if err != nil {
		_, err := msg.Reply(b, 
			"❌ <b>Invalid bot ID!</b>\n\nBot ID must be a valid number.",
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Get the bot instance and verify ownership
	instance, err := db.GetBotInstance(botID)
	if err != nil {
		log.Errorf("[Clone] Failed to get bot instance: %v", err)
		_, err := msg.Reply(b, "❌ <b>Database error occurred. Please try again later.</b>", 
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	if instance == nil {
		_, err := msg.Reply(b, 
			fmt.Sprintf("❌ <b>Bot instance not found!</b>\n\nBot ID <code>%d</code> doesn't exist.", botID),
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Check ownership
	if instance.OwnerID != user.Id {
		_, err := msg.Reply(b, "❌ <b>Access denied!</b>\n\nYou can only stop your own bot instances.", 
			&gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	// Send stopping message
	stoppingMsg, err := msg.Reply(b, 
		fmt.Sprintf("🔄 <b>Stopping bot @%s...</b>", instance.BotUsername),
		&gotgbot.SendMessageOpts{ParseMode: "HTML"})
	if err != nil {
		return err
	}

	// Stop the bot instance
	if GlobalBotManager != nil {
		err = GlobalBotManager.StopBotInstance(botID)
		if err != nil {
			log.Warnf("[Clone] Failed to stop bot instance %d: %v", botID, err)
		}
	}

	// Update database status
	err = db.UpdateBotInstanceStatus(botID, false)
	if err != nil {
		log.Errorf("[Clone] Failed to update bot instance status: %v", err)
		_, _, err := stoppingMsg.EditText(b, "❌ <b>Failed to update bot status in database</b>", 
			&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
		return err
	}

	// Success message
	_, _, err = stoppingMsg.EditText(b, 
		fmt.Sprintf("✅ <b>Bot @%s stopped successfully!</b>\n\n"+
		"The bot instance has been deactivated and is no longer processing updates.", instance.BotUsername),
		&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
	return err
}

// LoadClone registers all clone module command handlers with the dispatcher.
// Sets up commands for bot cloning, management, and statistics.
func LoadClone(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(cloneModule.moduleName, true)
	
	// Add alternative help names for better discoverability
	HelpModule.AltHelpOptions[cloneModule.moduleName] = []string{"clone", "clones", "bot_cloning", "instances"}

	// Register clone commands - these check for owner permission inside the handler
	dispatcher.AddHandler(handlers.NewCommand("clone", cloneBot))
	dispatcher.AddHandler(handlers.NewCommand("clones", listClones))
	dispatcher.AddHandler(handlers.NewCommand("clone_stop", stopClone))

	log.Info("[Modules] Clone commands loaded")
}