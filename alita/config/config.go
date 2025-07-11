package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Default values for configuration
const (
	DefaultBotVersion        = "2.1.3"
	DefaultApiServer         = "https://api.telegram.org"
	DefaultMainDbName        = "Alita_Robot"
	DefaultRedisAddress      = "localhost:6379"
	DefaultRedisDB           = 0
	DefaultWorkingMode       = "worker"
	DefaultDropPendingUpdate = true
)

// Default allowed updates for Telegram bot
var DefaultAllowedUpdates = []string{
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

// Default valid language codes
var DefaultValidLangCodes = []string{"en"}

// Config holds all configuration values for the bot
type Config struct {
	// Bot configuration
	BotToken           string
	BotVersion         string
	ApiServer          string
	AllowedUpdates     []string
	DropPendingUpdates bool
	WorkingMode        string
	Debug              bool

	// User configuration
	OwnerId     int64
	MessageDump int64

	// Database configuration
	DatabaseURI string
	MainDbName  string

	// Redis configuration
	RedisAddress  string
	RedisPassword string
	RedisDB       int

	// Internationalization
	ValidLangCodes []string
}

// Global config instance for backward compatibility
var globalConfig *Config

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("config validation failed for %s: %s", e.Field, e.Message)
}

// Load reads configuration from environment variables and returns a validated Config
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	cfg := &Config{
		// Set defaults first
		BotVersion:         DefaultBotVersion,
		ApiServer:          DefaultApiServer,
		MainDbName:         DefaultMainDbName,
		RedisAddress:       DefaultRedisAddress,
		RedisDB:            DefaultRedisDB,
		WorkingMode:        DefaultWorkingMode,
		DropPendingUpdates: DefaultDropPendingUpdate,
		AllowedUpdates:     DefaultAllowedUpdates,
		ValidLangCodes:     DefaultValidLangCodes,
	}

	var errors []ValidationError

	// Required fields
	if cfg.BotToken = os.Getenv("BOT_TOKEN"); cfg.BotToken == "" {
		errors = append(errors, ValidationError{"BOT_TOKEN", "is required"})
	}

	if cfg.DatabaseURI = os.Getenv("DB_URI"); cfg.DatabaseURI == "" {
		errors = append(errors, ValidationError{"DB_URI", "is required"})
	}

	// Parse and validate OwnerId
	if ownerIdStr := os.Getenv("OWNER_ID"); ownerIdStr == "" {
		errors = append(errors, ValidationError{"OWNER_ID", "is required"})
	} else {
		var err error
		if cfg.OwnerId, err = strconv.ParseInt(ownerIdStr, 10, 64); err != nil {
			errors = append(errors, ValidationError{"OWNER_ID", "must be a valid integer"})
		}
	}

	// Parse and validate MessageDump
	if messageDumpStr := os.Getenv("MESSAGE_DUMP"); messageDumpStr == "" {
		errors = append(errors, ValidationError{"MESSAGE_DUMP", "is required"})
	} else {
		var err error
		if cfg.MessageDump, err = strconv.ParseInt(messageDumpStr, 10, 64); err != nil {
			errors = append(errors, ValidationError{"MESSAGE_DUMP", "must be a valid integer"})
		}
	}

	// Optional fields with validation
	cfg.Debug = getBool("DEBUG", false)

	if dropUpdatesStr := os.Getenv("DROP_PENDING_UPDATES"); dropUpdatesStr != "" {
		cfg.DropPendingUpdates = getBool("DROP_PENDING_UPDATES", DefaultDropPendingUpdate)
	}

	if apiServer := os.Getenv("API_SERVER"); apiServer != "" {
		cfg.ApiServer = apiServer
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.MainDbName = dbName
	}

	if redisAddr := os.Getenv("REDIS_ADDRESS"); redisAddr != "" {
		cfg.RedisAddress = redisAddr
	}

	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		cfg.RedisPassword = redisPass
	}

	if redisDBStr := os.Getenv("REDIS_DB"); redisDBStr != "" {
		if redisDB, err := strconv.Atoi(redisDBStr); err != nil {
			errors = append(errors, ValidationError{"REDIS_DB", "must be a valid integer"})
		} else {
			cfg.RedisDB = redisDB
		}
	}

	// Parse string arrays
	if allowedUpdatesStr := os.Getenv("ALLOWED_UPDATES"); allowedUpdatesStr != "" {
		cfg.AllowedUpdates = getStringSlice("ALLOWED_UPDATES", DefaultAllowedUpdates)
	}

	if enabledLocalesStr := os.Getenv("ENABLED_LOCALES"); enabledLocalesStr != "" {
		cfg.ValidLangCodes = getStringSlice("ENABLED_LOCALES", DefaultValidLangCodes)
	}

	// Return validation errors if any
	if len(errors) > 0 {
		return nil, &ConfigValidationError{Errors: errors}
	}

	return cfg, nil
}

// Initialize loads and sets the global configuration
func Initialize() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	globalConfig = cfg
	return nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		panic("config not initialized - call config.Initialize() first")
	}
	return globalConfig
}

// Backward compatibility variables - these maintain the old API
var (
	BotToken           string
	BotVersion         string
	ApiServer          string
	AllowedUpdates     []string
	DropPendingUpdates bool
	WorkingMode        string
	Debug              bool
	OwnerId            int64
	MessageDump        int64
	DatabaseURI        string
	MainDbName         string
	RedisAddress       string
	RedisPassword      string
	RedisDB            int
	ValidLangCodes     []string
)

// updateGlobalVars updates the backward compatibility global variables
func updateGlobalVars(cfg *Config) {
	BotToken = cfg.BotToken
	BotVersion = cfg.BotVersion
	ApiServer = cfg.ApiServer
	AllowedUpdates = cfg.AllowedUpdates
	DropPendingUpdates = cfg.DropPendingUpdates
	WorkingMode = cfg.WorkingMode
	Debug = cfg.Debug
	OwnerId = cfg.OwnerId
	MessageDump = cfg.MessageDump
	DatabaseURI = cfg.DatabaseURI
	MainDbName = cfg.MainDbName
	RedisAddress = cfg.RedisAddress
	RedisPassword = cfg.RedisPassword
	RedisDB = cfg.RedisDB
	ValidLangCodes = cfg.ValidLangCodes
}

// LoadAndSetGlobals loads configuration and sets global variables for backward compatibility
func LoadAndSetGlobals() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	globalConfig = cfg
	updateGlobalVars(cfg)
	return nil
}

// ConfigValidationError contains multiple validation errors
type ConfigValidationError struct {
	Errors []ValidationError
}

func (e *ConfigValidationError) Error() string {
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("configuration validation failed:\n%s", strings.Join(messages, "\n"))
}

// Helper functions for parsing environment variables

// getBool parses a boolean from environment variable
func getBool(key string, defaultValue bool) bool {
	value := strings.ToLower(os.Getenv(key))
	switch value {
	case "true", "yes", "1":
		return true
	case "false", "no", "0":
		return false
	default:
		return defaultValue
	}
}



// getStringSlice parses a comma-separated string into a slice
func getStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	if len(result) == 0 {
		return defaultValue
	}
	return result
}
