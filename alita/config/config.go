package config

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Config holds all configuration for the bot
type Config struct {
	// Core configuration
	BotToken    string `validate:"required"`
	BotVersion  string
	ApiServer   string
	WorkingMode string
	Debug       bool

	// Bot settings
	OwnerId            int64 `validate:"required,min=1"`
	MessageDump        int64 `validate:"required,min=1"`
	DropPendingUpdates bool
	AllowedUpdates     []string
	ValidLangCodes     []string

	// Database configuration
	DatabaseURL string `validate:"required"`

	// Redis configuration
	RedisAddress  string `validate:"required"`
	RedisPassword string
	RedisDB       int

	// Webhook configuration
	UseWebhooks   bool
	WebhookDomain string
	WebhookSecret string
	WebhookPort   int `validate:"min=1,max=65535"`

	// Worker pool configuration for concurrent processing
	ChatValidationWorkers  int `validate:"min=1,max=100"`
	DatabaseWorkers        int `validate:"min=1,max=50"`
	MessagePipelineWorkers int `validate:"min=1,max=50"`
	BulkOperationWorkers   int `validate:"min=1,max=20"`
	CacheWorkers           int `validate:"min=1,max=20"`
	StatsCollectionWorkers int `validate:"min=1,max=10"`

	// Safety and performance limits
	MaxConcurrentOperations     int           `validate:"min=1,max=1000"`
	OperationTimeoutSeconds     int           `validate:"min=1,max=300"`
	OperationTimeout            time.Duration // Computed from OperationTimeoutSeconds
	EnablePerformanceMonitoring bool
	EnableBackgroundStats       bool

	// Cache configuration
	CacheNumCounters int64 `validate:"min=1000,max=1000000"`  // Ristretto NumCounters
	CacheMaxCost     int64 `validate:"min=100,max=100000000"` // Ristretto MaxCost
}

// Global configuration instance
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
	// Webhook configuration
	UseWebhooks   bool
	WebhookDomain string
	WebhookSecret string
	WebhookPort   int

	// Worker pool configuration for concurrent processing
	ChatValidationWorkers  int
	DatabaseWorkers        int
	MessagePipelineWorkers int
	BulkOperationWorkers   int
	CacheWorkers           int
	StatsCollectionWorkers int

	// Safety and performance limits
	MaxConcurrentOperations     int
	OperationTimeoutSeconds     int
	EnablePerformanceMonitoring bool
	EnableBackgroundStats       bool

	// Cache configuration
	CacheNumCounters int64
	CacheMaxCost     int64

	// Global config instance
	AppConfig *Config
)

// ValidateConfig validates the configuration struct and returns an error if any required
// fields are missing or values are outside acceptable ranges.
func ValidateConfig(cfg *Config) error {
	if cfg.BotToken == "" {
		return fmt.Errorf("BOT_TOKEN is required")
	}
	if cfg.OwnerId == 0 {
		return fmt.Errorf("OWNER_ID is required and must be greater than 0")
	}
	if cfg.MessageDump == 0 {
		return fmt.Errorf("MESSAGE_DUMP is required and must be greater than 0")
	}
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisAddress == "" {
		return fmt.Errorf("REDIS_ADDRESS is required")
	}

	// Validate webhook configuration if webhooks are enabled
	if cfg.UseWebhooks {
		if cfg.WebhookDomain == "" {
			return fmt.Errorf("WEBHOOK_DOMAIN is required when USE_WEBHOOKS is enabled")
		}
		if cfg.WebhookPort <= 0 || cfg.WebhookPort > 65535 {
			return fmt.Errorf("WEBHOOK_PORT must be between 1 and 65535")
		}
	}

	// Validate worker configurations
	if cfg.ChatValidationWorkers <= 0 || cfg.ChatValidationWorkers > 100 {
		return fmt.Errorf("CHAT_VALIDATION_WORKERS must be between 1 and 100")
	}
	if cfg.DatabaseWorkers <= 0 || cfg.DatabaseWorkers > 50 {
		return fmt.Errorf("DATABASE_WORKERS must be between 1 and 50")
	}
	if cfg.MessagePipelineWorkers <= 0 || cfg.MessagePipelineWorkers > 50 {
		return fmt.Errorf("MESSAGE_PIPELINE_WORKERS must be between 1 and 50")
	}
	if cfg.BulkOperationWorkers <= 0 || cfg.BulkOperationWorkers > 20 {
		return fmt.Errorf("BULK_OPERATION_WORKERS must be between 1 and 20")
	}
	if cfg.CacheWorkers <= 0 || cfg.CacheWorkers > 20 {
		return fmt.Errorf("CACHE_WORKERS must be between 1 and 20")
	}
	if cfg.StatsCollectionWorkers <= 0 || cfg.StatsCollectionWorkers > 10 {
		return fmt.Errorf("STATS_COLLECTION_WORKERS must be between 1 and 10")
	}

	// Validate cache configuration
	if cfg.CacheNumCounters != 0 && (cfg.CacheNumCounters < 1000 || cfg.CacheNumCounters > 1000000) {
		return fmt.Errorf("CACHE_NUM_COUNTERS must be between 1000 and 1000000")
	}
	if cfg.CacheMaxCost != 0 && (cfg.CacheMaxCost < 100 || cfg.CacheMaxCost > 100000000) {
		return fmt.Errorf("CACHE_MAX_COST must be between 100 and 100000000")
	}

	// Validate performance limits
	if cfg.MaxConcurrentOperations <= 0 || cfg.MaxConcurrentOperations > 1000 {
		return fmt.Errorf("MAX_CONCURRENT_OPERATIONS must be between 1 and 1000")
	}
	if cfg.OperationTimeoutSeconds <= 0 || cfg.OperationTimeoutSeconds > 300 {
		return fmt.Errorf("OPERATION_TIMEOUT_SECONDS must be between 1 and 300")
	}

	return nil
}

// LoadConfig loads configuration from environment variables, applies defaults,
// validates the configuration, and returns a populated Config instance.
func LoadConfig() (*Config, error) {
	// load goenv config
	_ = godotenv.Load() // Ignore error as .env file is optional

	cfg := &Config{
		// Core configuration
		BotToken:    os.Getenv("BOT_TOKEN"),
		BotVersion:  "2.1.3",
		ApiServer:   os.Getenv("API_SERVER"),
		WorkingMode: "worker",
		Debug:       typeConvertor{str: os.Getenv("DEBUG")}.Bool(),

		// Bot settings
		OwnerId:            typeConvertor{str: os.Getenv("OWNER_ID")}.Int64(),
		MessageDump:        typeConvertor{str: os.Getenv("MESSAGE_DUMP")}.Int64(),
		DropPendingUpdates: typeConvertor{str: os.Getenv("DROP_PENDING_UPDATES")}.Bool(),

		// Database configuration
		DatabaseURL: os.Getenv("DATABASE_URL"),

		// Redis configuration
		RedisAddress:  os.Getenv("REDIS_ADDRESS"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       typeConvertor{str: os.Getenv("REDIS_DB")}.Int(),

		// Webhook configuration
		UseWebhooks:   typeConvertor{str: os.Getenv("USE_WEBHOOKS")}.Bool(),
		WebhookDomain: os.Getenv("WEBHOOK_DOMAIN"),
		WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
		WebhookPort:   typeConvertor{str: os.Getenv("WEBHOOK_PORT")}.Int(),

		// Worker pool configuration
		ChatValidationWorkers:  typeConvertor{str: os.Getenv("CHAT_VALIDATION_WORKERS")}.Int(),
		DatabaseWorkers:        typeConvertor{str: os.Getenv("DATABASE_WORKERS")}.Int(),
		MessagePipelineWorkers: typeConvertor{str: os.Getenv("MESSAGE_PIPELINE_WORKERS")}.Int(),
		BulkOperationWorkers:   typeConvertor{str: os.Getenv("BULK_OPERATION_WORKERS")}.Int(),
		CacheWorkers:           typeConvertor{str: os.Getenv("CACHE_WORKERS")}.Int(),
		StatsCollectionWorkers: typeConvertor{str: os.Getenv("STATS_COLLECTION_WORKERS")}.Int(),

		// Safety and performance limits
		MaxConcurrentOperations:     typeConvertor{str: os.Getenv("MAX_CONCURRENT_OPERATIONS")}.Int(),
		OperationTimeoutSeconds:     typeConvertor{str: os.Getenv("OPERATION_TIMEOUT_SECONDS")}.Int(),
		EnablePerformanceMonitoring: typeConvertor{str: os.Getenv("ENABLE_PERFORMANCE_MONITORING")}.Bool(),
		EnableBackgroundStats:       typeConvertor{str: os.Getenv("ENABLE_BACKGROUND_STATS")}.Bool(),

		// Cache configuration
		CacheNumCounters: typeConvertor{str: os.Getenv("CACHE_NUM_COUNTERS")}.Int64(),
		CacheMaxCost:     typeConvertor{str: os.Getenv("CACHE_MAX_COST")}.Int64(),
	}

	// Set defaults
	cfg.setDefaults()

	// Validate configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set computed values
	cfg.OperationTimeout = time.Duration(cfg.OperationTimeoutSeconds) * time.Second

	// Set allowed updates and valid language codes
	cfg.AllowedUpdates = []string{
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

	cfg.ValidLangCodes = typeConvertor{str: os.Getenv("ENABLED_LOCALES")}.StringArray()
	if (len(cfg.ValidLangCodes) == 1 && cfg.ValidLangCodes[0] == "") || (len(cfg.ValidLangCodes) == 0) {
		cfg.ValidLangCodes = []string{"en"}
	}

	return cfg, nil
}

// setDefaults sets default values for configuration fields that are not provided
// via environment variables. It calculates appropriate defaults based on system
// resources and production best practices.
func (cfg *Config) setDefaults() {
	if cfg.ApiServer == "" {
		cfg.ApiServer = "https://api.telegram.org"
	}
	if cfg.WorkingMode == "" {
		cfg.WorkingMode = "worker"
	}
	if cfg.RedisAddress == "" {
		cfg.RedisAddress = "localhost:6379"
	}
	if cfg.RedisDB == 0 {
		cfg.RedisDB = 1
	}
	if cfg.WebhookPort == 0 {
		cfg.WebhookPort = 8080
	}

	// Set default values for worker pool configurations
	cpuCount := runtime.NumCPU()

	if cfg.ChatValidationWorkers == 0 {
		cfg.ChatValidationWorkers = 10
	}
	if cfg.DatabaseWorkers == 0 {
		cfg.DatabaseWorkers = 5
	}
	if cfg.MessagePipelineWorkers == 0 {
		cfg.MessagePipelineWorkers = cpuCount
		if cfg.MessagePipelineWorkers > 8 {
			cfg.MessagePipelineWorkers = 8
		}
	}
	if cfg.BulkOperationWorkers == 0 {
		cfg.BulkOperationWorkers = 4
	}
	if cfg.CacheWorkers == 0 {
		cfg.CacheWorkers = 3
	}
	if cfg.StatsCollectionWorkers == 0 {
		cfg.StatsCollectionWorkers = 2
	}

	// Set cache defaults (more reasonable values than hardcoded 1000/100)
	if cfg.CacheNumCounters == 0 {
		cfg.CacheNumCounters = 10000 // 10x more counters for better performance
	}
	if cfg.CacheMaxCost == 0 {
		cfg.CacheMaxCost = 10000 // 100x larger cache for better hit rates
	}

	// Set default safety limits
	if cfg.MaxConcurrentOperations == 0 {
		cfg.MaxConcurrentOperations = 50
	}
	if cfg.OperationTimeoutSeconds == 0 {
		cfg.OperationTimeoutSeconds = 30
	}

	// Enable monitoring by default in production
	if !cfg.Debug {
		if !cfg.EnablePerformanceMonitoring {
			cfg.EnablePerformanceMonitoring = true
		}
		if !cfg.EnableBackgroundStats {
			cfg.EnableBackgroundStats = true
		}
	}

	// Default DATABASE_URL if not set
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://postgres:password@localhost:5432/alita_robot?sslmode=disable"
		log.Warn("[Config] DATABASE_URL not set, using default: ", cfg.DatabaseURL)
	}
}

// init initializes the logging configuration, loads the global configuration
// from environment variables, validates it, and sets up global variables for
// backward compatibility. This function is called automatically at package import.
func init() {
	// set logger config
	log.SetLevel(log.DebugLevel)
	// SetReportCaller will be configured after debug mode is determined
	log.SetFormatter(
		&log.JSONFormatter{
			DisableHTMLEscape: true,
			PrettyPrint:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return f.Function, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		},
	)

	// Load the structured configuration
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("[Config] Failed to load configuration: %v", err)
	}

	// Set global configuration instance
	AppConfig = cfg

	// Set global variables for backward compatibility
	BotToken = cfg.BotToken
	DatabaseURL = cfg.DatabaseURL
	BotVersion = cfg.BotVersion
	ApiServer = cfg.ApiServer
	WorkingMode = cfg.WorkingMode
	Debug = cfg.Debug
	DropPendingUpdates = cfg.DropPendingUpdates
	OwnerId = cfg.OwnerId
	MessageDump = cfg.MessageDump
	RedisAddress = cfg.RedisAddress
	RedisPassword = cfg.RedisPassword
	RedisDB = cfg.RedisDB
	UseWebhooks = cfg.UseWebhooks
	WebhookDomain = cfg.WebhookDomain
	WebhookSecret = cfg.WebhookSecret
	WebhookPort = cfg.WebhookPort
	ChatValidationWorkers = cfg.ChatValidationWorkers
	DatabaseWorkers = cfg.DatabaseWorkers
	MessagePipelineWorkers = cfg.MessagePipelineWorkers
	BulkOperationWorkers = cfg.BulkOperationWorkers
	CacheWorkers = cfg.CacheWorkers
	StatsCollectionWorkers = cfg.StatsCollectionWorkers
	MaxConcurrentOperations = cfg.MaxConcurrentOperations
	OperationTimeoutSeconds = cfg.OperationTimeoutSeconds
	EnablePerformanceMonitoring = cfg.EnablePerformanceMonitoring
	EnableBackgroundStats = cfg.EnableBackgroundStats
	CacheNumCounters = cfg.CacheNumCounters
	CacheMaxCost = cfg.CacheMaxCost
	AllowedUpdates = cfg.AllowedUpdates
	ValidLangCodes = cfg.ValidLangCodes

	// Configure logger based on debug mode
	log.SetReportCaller(cfg.Debug) // Only enable stack traces in debug mode

	log.Info("[Config] Configuration loaded and validated successfully")
}
