package i18n

import (
	"bytes"
	"embed"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const defaultLangCode = "en"

// localeMap is the map of the embedded locale
var localeMap = make(map[string][]byte)

// I18n is the struct for i18n
type I18n struct {
	LangCode string
}

// LoadLocaleFiles Load Locales files which are embedded in the binary
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

// GetString get the string from the embedded locale
func (goloc I18n) GetString(key string) string {
	vi := viper.New()
	vi.SetConfigType("yaml")
	vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode]))
	text := vi.GetString(key)

	// if the language code is not available, return the default
	if text == "<nil>" {
		return I18n{LangCode: defaultLangCode}.GetString(key)
	}

	return text
}

// GetStringSlice get the string slice from the embedded locale
func (goloc I18n) GetStringSlice(key string) []string {
	vi := viper.New()
	vi.SetConfigType("yaml")
	vi.ReadConfig(bytes.NewBuffer(localeMap[goloc.LangCode]))
	text := vi.GetStringSlice(key)

	// if the language code is not available, return the default
	if len(text) == 0 {
		return I18n{LangCode: defaultLangCode}.GetStringSlice(key)
	}

	return text
}
