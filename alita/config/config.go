package config

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	AllowedUpdates     []string
	ValidLangCodes     []string
	BotToken           string
	DatabaseURL        string // Single PostgreSQL connection string
	BotVersion         string = "2.1.3"
	ApiServer          string
	WorkingMode        = "worker"
	Debug              = false
	DropPendingUpdates = true
	OwnerId            int64
	MessageDump        int64
	RedisAddress       string
	RedisPassword      string
	RedisDB            int
)

// init initializes the config variables.
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
	_ = godotenv.Load() // Ignore error as .env file is optional

	// set necessary variables
	Debug = typeConvertor{str: os.Getenv("DEBUG")}.Bool()
	BotToken = os.Getenv("BOT_TOKEN")
	DatabaseURL = os.Getenv("DATABASE_URL")
	ApiServer = os.Getenv("API_SERVER")
	OwnerId = typeConvertor{str: os.Getenv("OWNER_ID")}.Int64()
	MessageDump = typeConvertor{str: os.Getenv("MESSAGE_DUMP")}.Int64()
	DropPendingUpdates = typeConvertor{str: os.Getenv("DROP_PENDING_UPDATES")}.Bool()
	RedisAddress = os.Getenv("REDIS_ADDRESS")
	RedisPassword = os.Getenv("REDIS_PASSWORD")
	RedisDB = typeConvertor{str: os.Getenv("REDIS_DB")}.Int()

	// set default values
	if ApiServer == "" {
		ApiServer = "https://api.telegram.org"
	}
	if WorkingMode == "" {
		WorkingMode = "worker"
	}
	if !DropPendingUpdates {
		DropPendingUpdates = true
	}
	if RedisAddress == "" {
		RedisAddress = "localhost:6379"
	}
	if RedisDB == 0 {
		RedisDB = 1
	}

	// Default DATABASE_URL if not set
	if DatabaseURL == "" {
		DatabaseURL = "postgres://postgres:password@localhost:5432/alita_robot?sslmode=disable"
		log.Warn("[Config] DATABASE_URL not set, using default: ", DatabaseURL)
	}

	// check if all necessary variables are set
	if BotToken == "" {
		log.Fatal("[Config][BotToken] BOT_TOKEN is not set")
	}
	if OwnerId == 0 {
		log.Fatal("[Config][OwnerId] OWNER_ID is not set")
	}
	if MessageDump == 0 {
		log.Fatal("[Config][MessageDump] MESSAGE_DUMP is not set")
	}

	// set allowed updates
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

	// set valid language codes
	ValidLangCodes = typeConvertor{str: os.Getenv("ENABLED_LOCALES")}.StringArray()
	// if valid lang codes is not set, set it to 'en' only
	if (len(ValidLangCodes) == 1 && ValidLangCodes[0] == "") || (len(ValidLangCodes) == 0) {
		ValidLangCodes = []string{"en"}
	}
}
