package i18n

import (
	"bytes"
	"embed"
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
	vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode]))

	// helper to read a key and determine if value exists
	read := func(k string) string {
		val := vi.GetString(k)
		if val == "<nil>" {
			val = ""
		}
		return val
	}

	// 1) try the key as-is
	text := read(key)

	// 2) if not found and missing global prefix, try with "strings." prefix
	if text == "" && !strings.HasPrefix(key, "strings.") {
		text = read("strings." + key)
	}

	// 3) fallback to default language, if we are not already on default
	if text == "" && goloc.LangCode != defaultLangCode {
		return I18n{LangCode: defaultLangCode}.GetString(key)
	}

	return text
}

// GetStringSlice retrieves a localized string slice for the given key from the embedded locale.
// If the key is not found or the language code is unavailable, it falls back to the default language.
func (goloc I18n) GetStringSlice(key string) []string {
	vi := viper.New()
	vi.SetConfigType("yaml")
	vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode]))

	read := func(k string) []string {
		return vi.GetStringSlice(k)
	}

	text := read(key)

	if len(text) == 0 && !strings.HasPrefix(key, "strings.") {
		text = read("strings." + key)
	}

	if len(text) == 0 && goloc.LangCode != defaultLangCode {
		return I18n{LangCode: defaultLangCode}.GetStringSlice(key)
	}

	return text
}
