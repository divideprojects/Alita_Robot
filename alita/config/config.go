package config

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	// AllowedUpdates specifies which Telegram update types the bot should receive.
	AllowedUpdates []string
	// ValidLangCodes lists enabled locale codes for i18n.
	ValidLangCodes []string
	// BotToken is the Telegram bot token.
	BotToken string
	// DatabaseURI is the URI for the main database connection.
	DatabaseURI string
	// MainDbName is the name of the main database.
	MainDbName string
	// BotVersion is the current version of the bot.
	BotVersion string = "2.1.3"
	// ApiServer is the Telegram API server endpoint.
	ApiServer string
	// WorkingMode indicates the current working mode (e.g., "worker").
	WorkingMode = "worker"
	// Debug enables debug logging if true.
	Debug = false
	// DropPendingUpdates determines if pending updates should be dropped on startup.
	DropPendingUpdates = true
	// OwnerId is the Telegram user ID of the bot owner.
	OwnerId int64
	// MessageDump is the chat ID where startup/log messages are sent.
	MessageDump int64
	// RedisAddress is the address of the Redis server.
	RedisAddress string
	// RedisPassword is the password for the Redis server.
	RedisPassword string
	// RedisDB is the Redis database number to use.
	RedisDB int
)

// ValidateConfig validates all configuration values
func ValidateConfig() error {
	var errors []string

	// Validate required fields
	if BotToken == "" {
		errors = append(errors, "BOT_TOKEN is required")
	}
	if DatabaseURI == "" {
		errors = append(errors, "DB_URI is required")
	}
	if OwnerId == 0 {
		errors = append(errors, "OWNER_ID is required")
	}
	if MessageDump == 0 {
		errors = append(errors, "MESSAGE_DUMP is required")
	}

	// Validate bot token format (basic check)
	if BotToken != "" && len(BotToken) < 40 {
		errors = append(errors, "BOT_TOKEN appears to be invalid (too short)")
	}

	// Validate database URI format
	if DatabaseURI != "" && !strings.HasPrefix(DatabaseURI, "mongodb://") && !strings.HasPrefix(DatabaseURI, "mongodb+srv://") {
		errors = append(errors, "DB_URI must be a valid MongoDB connection string")
	}

	// Validate API server URL
	if ApiServer != "" && !strings.HasPrefix(ApiServer, "http://") && !strings.HasPrefix(ApiServer, "https://") {
		errors = append(errors, "API_SERVER must be a valid HTTP/HTTPS URL")
	}

	// Validate Redis configuration
	if RedisAddress != "" && !strings.Contains(RedisAddress, ":") {
		errors = append(errors, "REDIS_ADDRESS must include port (e.g., localhost:6379)")
	}

	if RedisDB < 0 || RedisDB > 15 {
		errors = append(errors, "REDIS_DB must be between 0 and 15")
	}

	// Validate language codes
	validLangPattern := regexp.MustCompile(`^[a-z]{2}$`)
	for _, lang := range ValidLangCodes {
		if !validLangPattern.MatchString(lang) {
			errors = append(errors, fmt.Sprintf("Invalid language code: %s", lang))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// init initializes the config variables from environment variables and sets up logging.
// It loads .env files, parses environment variables, and applies defaults for unset values.
func init() {
	// set logger config
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(
		&log.JSONFormatter{
			DisableHTMLEscape: true,
			PrettyPrint:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return f.Function, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		},
	)

	// load goenv config
	if err := godotenv.Load(); err != nil {
		log.Debug("No .env file found, using environment variables")
	}

	// set necessary variables
	Debug = typeConvertor{str: os.Getenv("DEBUG")}.Bool()
	DropPendingUpdates = typeConvertor{str: os.Getenv("DROP_PENDING_UPDATES")}.Bool()
	DatabaseURI = os.Getenv("DB_URI")
	MainDbName = os.Getenv("DB_NAME")
	OwnerId = typeConvertor{str: os.Getenv("OWNER_ID")}.Int64()
	MessageDump = typeConvertor{str: os.Getenv("MESSAGE_DUMP")}.Int64()
	BotToken = os.Getenv("BOT_TOKEN")

	AllowedUpdates = typeConvertor{str: os.Getenv("ALLOWED_UPDATES")}.StringArray()
	// if allowed updates is not set, set it to receive all updates
	if (len(AllowedUpdates) == 1 && AllowedUpdates[0] == "") || (len(AllowedUpdates) == 0) {
		AllowedUpdates = []string{
			"message",
			"edited_message",
			"channel_post",
			"edited_channel_post",
			"inline_query",
			"chosen_inline_result",
			"callback_query",
			"shipping_query",
			"pre_checkout_query",
			"poll",
			"poll_answer",
			"my_chat_member",
			"chat_member",
			"chat_join_request",
		}
	}

	ValidLangCodes = typeConvertor{str: os.Getenv("ENABLED_LOCALES")}.StringArray()
	// if valid lang codes is not set, set it to 'en' only
	if (len(ValidLangCodes) == 1 && ValidLangCodes[0] == "") || (len(ValidLangCodes) == 0) {
		ValidLangCodes = []string{"en"}
	}

	ApiServer = os.Getenv("API_SERVER")
	// set as default api server if not set
	if ApiServer == "" {
		ApiServer = "https://api.telegram.org"
	}
	// set default db_name
	if MainDbName == "" {
		MainDbName = "Alita_Robot"
	}

	// redis config
	RedisAddress = os.Getenv("REDIS_ADDRESS")
	if os.Getenv("REDIS_ADDRESS") == "" {
		RedisAddress = "localhost:6379"
	}
	RedisPassword = os.Getenv("REDIS_PASSWORD")
	if os.Getenv("REDIS_PASSWORD") == "" {
		RedisPassword = ""
	}
	RedisDB = typeConvertor{str: os.Getenv("REDIS_DB")}.Int()
	if os.Getenv("REDIS_DB") == "" {
		RedisDB = 0
	}

	// Validate configuration
	if err := ValidateConfig(); err != nil {
		log.Fatal("Configuration validation failed: ", err)
	}
}
