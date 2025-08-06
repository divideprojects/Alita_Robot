package i18n

import (
	"embed"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/spf13/viper"
)

// TranslationParams represents parameters for translation interpolation
type TranslationParams map[string]interface{}

// PluralRule defines pluralization rules for different languages
type PluralRule struct {
	Zero  string // "zero" form (e.g., 0 items)
	One   string // "one" form (e.g., 1 item)
	Two   string // "two" form (e.g., 2 items)
	Few   string // "few" form (e.g., 2-4 items in some languages)
	Many  string // "many" form (e.g., 5+ items in some languages)
	Other string // "other" form (default fallback)
}

// LocaleManager manages all locales with thread-safe operations
type LocaleManager struct {
	mu          sync.RWMutex
	viperCache  map[string]*viper.Viper // Pre-compiled viper instances
	localeData  map[string][]byte       // Raw YAML data
	cacheClient *cache.ChainCache[any]  // External cache for translations
	defaultLang string
	localeFS    *embed.FS
	localePath  string
}

// Translator provides translation methods for a specific language
type Translator struct {
	langCode    string
	manager     *LocaleManager
	viper       *viper.Viper // Pre-compiled viper instance
	cachePrefix string       // Cache key prefix for this language
}

// CacheConfig defines cache configuration for translations
type CacheConfig struct {
	TTL               time.Duration
	EnableCache       bool
	CacheKeyPrefix    string
	MaxCacheSize      int64
	InvalidateOnError bool
}

// LoaderConfig defines configuration for locale loading
type LoaderConfig struct {
	DefaultLanguage string
	ValidateYAML    bool
	StrictMode      bool // Fail if any locale file has errors
}

// ManagerConfig combines all configuration options
type ManagerConfig struct {
	Cache  CacheConfig
	Loader LoaderConfig
}

// DefaultManagerConfig returns sensible defaults for ManagerConfig.
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		Cache: CacheConfig{
			TTL:               30 * time.Minute,
			EnableCache:       true,
			CacheKeyPrefix:    "i18n:",
			MaxCacheSize:      1000,
			InvalidateOnError: false,
		},
		Loader: LoaderConfig{
			DefaultLanguage: "en",
			ValidateYAML:    true,
			StrictMode:      false,
		},
	}
}
