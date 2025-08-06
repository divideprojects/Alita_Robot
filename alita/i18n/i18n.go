package i18n

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const defaultLangCode = "en"

// localeMap is the map of the embedded locale (DEPRECATED: use LocaleManager)
var localeMap = make(map[string][]byte)

// I18n is the struct for i18n (DEPRECATED: use Translator from LocaleManager)
// This struct is kept for backward compatibility. New code should use the
// LocaleManager and Translator types for better performance and thread safety.
type I18n struct {
	LangCode string
}

// LoadLocaleFiles loads locale files which are embedded in the binary.
// DEPRECATED: Use LocaleManager.Initialize() instead for better error handling
// and performance. This function is kept for backward compatibility.
func LoadLocaleFiles(fs *embed.FS, path string) {
	// Also initialize the new system if not already done
	manager := GetManager()
	if manager.localeFS == nil {
		config := DefaultManagerConfig()
		_ = manager.Initialize(fs, path, config) // Ignore error for backward compatibility
	}

	// Keep old behavior for existing code
	entries, _ := fs.ReadDir(path)
	for _, entry := range entries {
		fp := filepath.Join(path, entry.Name())
		content, _ := fs.ReadFile(fp)
		langCode := func() (langCode string) {
			langCode = strings.ReplaceAll(entry.Name(), ".yml", "")
			langCode = strings.ReplaceAll(langCode, ".yaml", "")
			return
		}
		localeMap[langCode()] = content
	}
}

// GetString gets the string from the embedded locale using the specified key.
// DEPRECATED: Use Translator.GetString() instead for better performance
// and parameter interpolation support. This method is kept for backward compatibility.
func (goloc I18n) GetString(key string) string {
	// Try new system first if available
	manager := GetManager()
	if manager.localeFS != nil {
		translator, err := manager.GetTranslator(goloc.LangCode)
		if err == nil {
			result, err := translator.GetString(key)
			if err == nil {
				return result
			}
		}
	}

	// Fallback to old behavior
	vi := viper.New()
	vi.SetConfigType("yaml")
	if err := vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode])); err != nil {
		// If config reading fails, fallback to default language
		// Prevent infinite recursion
		if goloc.LangCode == defaultLangCode {
			return fmt.Sprintf("[MISSING:%s]", key)
		}
		return I18n{LangCode: defaultLangCode}.GetString(key)
	}
	text := vi.GetString(key)

	// if the language code is not available, return the default
	if text == "<nil>" || text == "" {
		if goloc.LangCode == defaultLangCode {
			return fmt.Sprintf("[MISSING:%s]", key)
		}
		return I18n{LangCode: defaultLangCode}.GetString(key)
	}

	return text
}

// GetStringSlice gets the string slice from the embedded locale using the specified key.
// DEPRECATED: Use Translator.GetStringSlice() instead for better performance
// and caching support. This method is kept for backward compatibility.
func (goloc I18n) GetStringSlice(key string) []string {
	// Try new system first if available
	manager := GetManager()
	if manager.localeFS != nil {
		translator, err := manager.GetTranslator(goloc.LangCode)
		if err == nil {
			result, err := translator.GetStringSlice(key)
			if err == nil {
				return result
			}
		}
	}

	// Fallback to old behavior
	vi := viper.New()
	vi.SetConfigType("yaml")
	if err := vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode])); err != nil {
		// If config reading fails, fallback to default language
		if goloc.LangCode == defaultLangCode {
			return []string{fmt.Sprintf("[MISSING:%s]", key)}
		}
		return I18n{LangCode: defaultLangCode}.GetStringSlice(key)
	}
	text := vi.GetStringSlice(key)

	// if the language code is not available, return the default
	if len(text) == 0 {
		if goloc.LangCode == defaultLangCode {
			return []string{fmt.Sprintf("[MISSING:%s]", key)}
		}
		return I18n{LangCode: defaultLangCode}.GetStringSlice(key)
	}

	return text
}

// NewTranslator creates a new Translator instance using the modern LocaleManager.
// This is the recommended way to handle translations in new code.
func NewTranslator(langCode string) (*Translator, error) {
	manager := GetManager()
	return manager.GetTranslator(langCode)
}

// MustNewTranslator creates a new Translator instance and panics on error.
// Useful for initialization where errors should be fatal.
func MustNewTranslator(langCode string) *Translator {
	translator, err := NewTranslator(langCode)
	if err != nil {
		panic(fmt.Sprintf("Failed to create translator for %s: %v", langCode, err))
	}
	return translator
}

// GetAvailableLanguages returns all available language codes.
// This is a convenience function that uses the LocaleManager.
func GetAvailableLanguages() []string {
	manager := GetManager()
	return manager.GetAvailableLanguages()
}

// IsLanguageSupported checks if a language is supported.
// This is a convenience function that uses the LocaleManager.
func IsLanguageSupported(langCode string) bool {
	manager := GetManager()
	return manager.IsLanguageSupported(langCode)
}
