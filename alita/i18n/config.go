package i18n

import (
	"os"
	"strings"
	"sync"
)

// Environment constants
const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
	EnvTest        = "test"
)

// FallbackMode constants
const (
	FallbackModeFriendly = "friendly" // User-friendly messages in production
	FallbackModeDebug    = "debug"    // Show @@key@@ markers for debugging
	FallbackModeMixed    = "mixed"    // Friendly in prod, debug in dev
)

// I18nConfig holds configuration for the i18n system
type I18nConfig struct {
	// Environment detection
	Environment string

	// Fallback behavior
	FallbackMode string

	// Logging configuration
	LogMissingKeys bool
	LogLevel       string

	// Fallback messages per language
	FallbackMessages map[string]string

	// Feature flags
	EnableStructuredLogging bool
	EnableMetrics           bool
}

var (
	globalConfig *I18nConfig
	configMu     sync.RWMutex
	configOnce   sync.Once
)

// DefaultFallbackMessages provides user-friendly messages when translations are missing
var DefaultFallbackMessages = map[string]string{
	"en": "Message not available",
	"es": "Mensaje no disponible",
	"fr": "Message non disponible",
	"de": "Nachricht nicht verfügbar",
	"it": "Messaggio non disponibile",
	"pt": "Mensagem não disponível",
	"ru": "Сообщение недоступно",
	"zh": "消息不可用",
	"ja": "メッセージが利用できません",
	"hi": "संदेश उपलब्ध नहीं है",
	"tr": "Mesaj mevcut değil",
	"nl": "Bericht niet beschikbaar",
	"pl": "Wiadomość niedostępna",
	"sv": "Meddelande inte tillgängligt",
	"da": "Besked ikke tilgængelig",
	"fi": "Viesti ei saatavilla",
	"ro": "Mesaj indisponibil",
	"uk": "Повідомлення недоступне",
	"el": "Μήνυμα μη διαθέσιμο",
	"cs": "Zpráva není k dispozici",
}

// GetConfig returns the global i18n configuration, initializing it if necessary
func GetConfig() *I18nConfig {
	configOnce.Do(func() {
		configMu.Lock()
		defer configMu.Unlock()
		if globalConfig == nil {
			globalConfig = loadConfig()
		}
	})

	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

// SetConfig allows setting a custom configuration (mainly for testing)
func SetConfig(config *I18nConfig) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = config
}

// ReloadConfig forces a reload of the configuration from environment
func ReloadConfig() {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = loadConfig()
}

// loadConfig loads configuration from environment variables
func loadConfig() *I18nConfig {
	config := &I18nConfig{
		Environment:             detectEnvironment(),
		FallbackMode:            getFallbackMode(),
		LogMissingKeys:          getBoolEnv("I18N_LOG_MISSING_KEYS", true),
		LogLevel:                getStringEnv("I18N_LOG_LEVEL", "info"),
		FallbackMessages:        DefaultFallbackMessages,
		EnableStructuredLogging: getBoolEnv("I18N_ENABLE_STRUCTURED_LOGGING", true),
		EnableMetrics:           getBoolEnv("I18N_ENABLE_METRICS", true),
	}

	// Allow custom fallback messages via environment
	if customFallback := getStringEnv("I18N_FALLBACK_MESSAGE", ""); customFallback != "" {
		// Set the same message for all languages if custom is provided
		for lang := range config.FallbackMessages {
			config.FallbackMessages[lang] = customFallback
		}
	}

	return config
}

// detectEnvironment determines the current environment
func detectEnvironment() string {
	env := strings.ToLower(getStringEnv("ENVIRONMENT", ""))

	switch env {
	case EnvProduction, "prod":
		return EnvProduction
	case EnvDevelopment, "dev":
		return EnvDevelopment
	case EnvTest, "testing":
		return EnvTest
	default:
		// Default to development if not specified
		return EnvDevelopment
	}
}

// getFallbackMode determines the fallback behavior
func getFallbackMode() string {
	mode := strings.ToLower(getStringEnv("I18N_FALLBACK_MODE", ""))

	switch mode {
	case FallbackModeFriendly:
		return FallbackModeFriendly
	case FallbackModeDebug:
		return FallbackModeDebug
	case FallbackModeMixed:
		return FallbackModeMixed
	default:
		// Default to mixed mode
		return FallbackModeMixed
	}
}

// ShouldUseFriendlyFallback determines if we should use friendly fallback messages
func (c *I18nConfig) ShouldUseFriendlyFallback() bool {
	switch c.FallbackMode {
	case FallbackModeFriendly:
		return true
	case FallbackModeDebug:
		return false
	case FallbackModeMixed:
		return c.Environment == EnvProduction
	default:
		return c.Environment == EnvProduction
	}
}

// GetFallbackMessage returns the appropriate fallback message for a language
func (c *I18nConfig) GetFallbackMessage(langCode string) string {
	if message, exists := c.FallbackMessages[langCode]; exists {
		return message
	}

	// Fall back to English if language not found
	if message, exists := c.FallbackMessages["en"]; exists {
		return message
	}

	// Ultimate fallback
	return "Message not available"
}

// IsProductionEnvironment checks if we're running in production
func (c *I18nConfig) IsProductionEnvironment() bool {
	return c.Environment == EnvProduction
}

// IsDevelopmentEnvironment checks if we're running in development
func (c *I18nConfig) IsDevelopmentEnvironment() bool {
	return c.Environment == EnvDevelopment
}

// IsTestEnvironment checks if we're running in test mode
func (c *I18nConfig) IsTestEnvironment() bool {
	return c.Environment == EnvTest
}

// Helper functions for environment variable parsing

func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := strings.ToLower(os.Getenv(key))
	switch value {
	case "true", "yes", "1", "on":
		return true
	case "false", "no", "0", "off":
		return false
	default:
		return defaultValue
	}
}
