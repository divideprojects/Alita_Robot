package i18n

import (
	"bytes"
	"embed"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const defaultLangCode = "en"

// localeMap holds the embedded locale YAML files, keyed by language code.
var localeMap = make(map[string][]byte)

// I18n provides methods for retrieving localized strings and slices for a given language code.
type I18n struct {
	LangCode string
}

// LoadLocaleFiles loads locale YAML files embedded in the binary into localeMap.
// It expects files to be named with their language code (e.g., "en.yml").
func LoadLocaleFiles(fs *embed.FS, path string) {
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

// GetString retrieves a localized string for the given key from the embedded locale.
// If the key is not found or the language code is unavailable, it falls back to the default language.
func (goloc I18n) GetString(key string) string {
	vi := viper.New()
	vi.SetConfigType("yaml")
	if err := vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode])); err != nil {
		log.Printf("Failed to read config for locale %s: %v", goloc.LangCode, err)
	}
	text := vi.GetString(key)

	// if the language code is not available, return the default
	if text == "<nil>" {
		return I18n{LangCode: defaultLangCode}.GetString(key)
	}

	return text
}

// GetStringSlice retrieves a localized string slice for the given key from the embedded locale.
// If the key is not found or the language code is unavailable, it falls back to the default language.
func (goloc I18n) GetStringSlice(key string) []string {
	vi := viper.New()
	vi.SetConfigType("yaml")
	if err := vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode])); err != nil {
		log.Printf("Failed to read config for locale %s: %v", goloc.LangCode, err)
	}
	text := vi.GetStringSlice(key)

	// if the language code is not available, return the default
	if len(text) == 0 {
		return I18n{LangCode: defaultLangCode}.GetStringSlice(key)
	}

	return text
}
