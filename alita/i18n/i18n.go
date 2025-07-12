// Package i18n provides internationalization support for the Alita bot.
//
// The package loads localized YAML files from embedded filesystem and provides
// methods to retrieve localized strings with configurable fallback chains.
//
// All i18n keys MUST start with the "strings." prefix for consistency and clarity.
//
// Basic usage:
//
//	// Load locales once at startup
//	if err := i18n.LoadLocaleFiles(&localesFS, "locales"); err != nil {
//		log.Fatal("Failed to load locales:", err)
//	}
//
//	// Create i18n instance with language code
//	tr := i18n.New("en")
//	text := tr.GetString("strings.welcome.message")
//
//	// Or use the convenience constructor
//	text := i18n.GetString("en", "strings.welcome.message")
//
// Key Format Requirements:
//
//	// ✅ Correct - keys must start with "strings." prefix:
//	tr.GetString("strings.welcome.message")
//	tr.GetString("strings.errors.not_found")
//
//	// ❌ Incorrect - keys without prefix will not be found:
//	tr.GetString("welcome.message")     // Will fail
//	tr.GetString("errors.not_found")    // Will fail
package i18n

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const (
	// DefaultLangCode is the fallback language when translations are missing
	DefaultLangCode = "en"
	// MissingKeyMarker is returned when a key is not found in any language
	MissingKeyMarker = "@@%s@@"
)

var (
	// ErrLanguageNotFound is returned when a requested language is not available
	ErrLanguageNotFound = errors.New("language not found")
	// ErrNoLocalesLoaded is returned when no locales have been loaded
	ErrNoLocalesLoaded = errors.New("no locales loaded")
	// ErrEmptyKey is returned when an empty key is requested
	ErrEmptyKey = errors.New("empty key provided")
)

// LoadError represents an error that occurred while loading a locale file
type LoadError struct {
	File string
	Err  error
}

func (e LoadError) Error() string {
	return fmt.Sprintf("failed to load locale file %s: %v", e.File, e.Err)
}

// LoadErrors is a collection of errors that occurred during locale loading
type LoadErrors []LoadError

func (e LoadErrors) Error() string {
	if len(e) == 0 {
		return "no errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("multiple locale loading errors: %s", strings.Join(msgs, "; "))
}

// localeMap holds the parsed locale configurations, keyed by language code.
var (
	localeMap = make(map[string]*viper.Viper)
	localeMu  sync.RWMutex

	// fallbackChains defines the fallback sequence for each language
	fallbackChains = map[string][]string{
		"pt_BR": {"pt", DefaultLangCode},
		"es_MX": {"es", DefaultLangCode},
		"en_US": {DefaultLangCode},
		"zh_CN": {"zh", DefaultLangCode},
		"zh_TW": {"zh", DefaultLangCode},
	}
	fallbackMu sync.RWMutex
)

// I18n provides methods for retrieving localized strings and slices for a given language code.
type I18n struct {
	LangCode string
}

// New creates a new I18n instance with the specified language code.
// If the language code is empty, it defaults to DefaultLangCode.
func New(langCode string) *I18n {
	if langCode == "" {
		langCode = DefaultLangCode
	}
	return &I18n{LangCode: langCode}
}

// IsLanguageAvailable checks if a language code is available.
func IsLanguageAvailable(langCode string) bool {
	localeMu.RLock()
	defer localeMu.RUnlock()
	_, exists := localeMap[langCode]
	return exists
}

// GetAvailableLanguages returns a slice of all available language codes.
func GetAvailableLanguages() []string {
	localeMu.RLock()
	defer localeMu.RUnlock()

	languages := make([]string, 0, len(localeMap))
	for lang := range localeMap {
		languages = append(languages, lang)
	}
	return languages
}

// SetFallbackChain sets a custom fallback chain for a language.
// The chain should not include the language itself - it's automatically tried first.
func SetFallbackChain(langCode string, chain []string) {
	fallbackMu.Lock()
	defer fallbackMu.Unlock()
	fallbackChains[langCode] = append([]string{}, chain...) // copy slice
}

// GetFallbackChain returns the fallback chain for a language.
func GetFallbackChain(langCode string) []string {
	fallbackMu.RLock()
	defer fallbackMu.RUnlock()

	if chain, exists := fallbackChains[langCode]; exists {
		return append([]string{}, chain...) // return copy
	}

	// Default fallback to DefaultLangCode if not the same
	if langCode != DefaultLangCode {
		return []string{DefaultLangCode}
	}
	return nil
}

// LoadLocaleFiles loads locale YAML files embedded in the binary into the locale map.
// It expects files to be named with their language code (e.g., "en.yml", "fr.yaml").
// Returns any errors encountered during loading.
func LoadLocaleFiles(fs *embed.FS, path string) error {
	entries, err := fs.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no files found in directory %s", path)
	}

	var loadErrors LoadErrors
	newLocaleMap := make(map[string]*viper.Viper)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".yml") && !strings.HasSuffix(filename, ".yaml") {
			continue
		}

		fp := filepath.Join(path, filename)
		content, err := fs.ReadFile(fp)
		if err != nil {
			loadErrors = append(loadErrors, LoadError{File: fp, Err: err})
			continue
		}

		langCode := extractLangCode(filename)
		if langCode == "" {
			loadErrors = append(loadErrors, LoadError{
				File: fp,
				Err:  fmt.Errorf("could not extract language code from filename"),
			})
			continue
		}

		vi := viper.New()
		vi.SetConfigType("yaml")
		if err := vi.ReadConfig(bytes.NewReader(content)); err != nil {
			loadErrors = append(loadErrors, LoadError{
				File: fp,
				Err:  fmt.Errorf("failed to parse YAML: %w", err),
			})
			continue
		}

		newLocaleMap[langCode] = vi
	}

	if len(newLocaleMap) == 0 {
		if len(loadErrors) > 0 {
			return loadErrors
		}
		return errors.New("no valid locale files found")
	}

	// Atomically update the locale map
	localeMu.Lock()
	localeMap = newLocaleMap
	localeMu.Unlock()

	// Return errors if any occurred, but still loaded some locales
	if len(loadErrors) > 0 {
		return loadErrors
	}

	return nil
}

// MustLoadLocaleFiles loads locale files and panics on error.
// Useful for main package initialization where locale loading failure should terminate the program.
func MustLoadLocaleFiles(fs *embed.FS, path string) {
	if err := LoadLocaleFiles(fs, path); err != nil {
		panic(fmt.Sprintf("failed to load locale files: %v", err))
	}
}

// Reload reloads locale files from the embedded filesystem.
// This is useful for runtime locale updates (though files are embedded, this is more for testing).
func Reload(fs *embed.FS, path string) error {
	return LoadLocaleFiles(fs, path)
}

// extractLangCode extracts the language code from a filename.
func extractLangCode(filename string) string {
	langCode := strings.TrimSuffix(filename, ".yml")
	langCode = strings.TrimSuffix(langCode, ".yaml")
	if langCode == filename {
		return "" // no valid extension found
	}
	return langCode
}

// getLocale safely retrieves a locale from the map.
func getLocale(langCode string) (*viper.Viper, bool) {
	localeMu.RLock()
	defer localeMu.RUnlock()
	locale, exists := localeMap[langCode]
	return locale, exists
}

// GetString retrieves a localized string for the given key.
// It tries the current language first, then falls back through the fallback chain.
// If no translation is found, it returns either a user-friendly message or a marked missing key
// depending on the configuration.
func (i I18n) GetString(key string) string {
	if key == "" {
		config := GetConfig()
		if config != nil && config.ShouldUseFriendlyFallback() {
			return config.GetFallbackMessage(i.LangCode)
		}
		return fmt.Sprintf(MissingKeyMarker, "empty-key")
	}

	// Try current language first
	if text := i.getStringFromLang(i.LangCode, key); text != "" {
		return text
	}

	// Try fallback chain
	for _, fallbackLang := range GetFallbackChain(i.LangCode) {
		if text := i.getStringFromLang(fallbackLang, key); text != "" {
			// Log fallback usage
			LogFallbackUsed(key, i.LangCode, fallbackLang)
			return text
		}
	}

	// Log missing key
	LogKeyNotFound(key, i.LangCode)

	// Return appropriate fallback based on configuration
	config := GetConfig()
	if config != nil && config.ShouldUseFriendlyFallback() {
		return config.GetFallbackMessage(i.LangCode)
	}

	// Return marked missing key for debugging
	return fmt.Sprintf(MissingKeyMarker, key)
}

// GetStringSlice retrieves a localized string slice for the given key.
// It follows the same fallback logic as GetString.
func (i I18n) GetStringSlice(key string) []string {
	if key == "" {
		return nil
	}

	// Try current language first
	if slice := i.getStringSliceFromLang(i.LangCode, key); len(slice) > 0 {
		return slice
	}

	// Try fallback chain
	for _, fallbackLang := range GetFallbackChain(i.LangCode) {
		if slice := i.getStringSliceFromLang(fallbackLang, key); len(slice) > 0 {
			// Log fallback usage
			LogFallbackUsed(key, i.LangCode, fallbackLang)
			return slice
		}
	}

	// Log missing key (only if logging is enabled to avoid spam for optional slices)
	config := GetConfig()
	if config != nil && config.LogMissingKeys {
		LogKeyNotFound(key, i.LangCode)
	}

	return nil
}

// GetStringWithError retrieves a localized string and returns an error if not found.
// This is useful when you need to distinguish between empty values and missing keys.
func (i I18n) GetStringWithError(key string) (string, error) {
	if key == "" {
		return "", ErrEmptyKey
	}

	// Try current language first
	if text := i.getStringFromLang(i.LangCode, key); text != "" {
		return text, nil
	}

	// Try fallback chain
	for _, fallbackLang := range GetFallbackChain(i.LangCode) {
		if text := i.getStringFromLang(fallbackLang, key); text != "" {
			// Log fallback usage
			LogFallbackUsed(key, i.LangCode, fallbackLang)
			return text, nil
		}
	}

	// Log missing key
	LogKeyNotFound(key, i.LangCode)

	return "", fmt.Errorf("key %q not found in language %q or its fallbacks", key, i.LangCode)
}

// getStringFromLang retrieves a string from a specific language.
func (i I18n) getStringFromLang(langCode, key string) string {
	locale, exists := getLocale(langCode)
	if !exists {
		return ""
	}

	// Only try the key as-is - no automatic prefix fallback
	if val := locale.GetString(key); val != "" && val != "<nil>" {
		return val
	}

	return ""
}

// getStringSliceFromLang retrieves a string slice from a specific language.
func (i I18n) getStringSliceFromLang(langCode, key string) []string {
	locale, exists := getLocale(langCode)
	if !exists {
		return nil
	}

	// Only try the key as-is - no automatic prefix fallback
	if slice := locale.GetStringSlice(key); len(slice) > 0 {
		return slice
	}

	return nil
}

// HasKey checks if a key exists in the current language or its fallbacks.
func (i I18n) HasKey(key string) bool {
	if key == "" {
		return false
	}

	// Check current language
	if i.hasKeyInLang(i.LangCode, key) {
		return true
	}

	// Check fallback chain
	for _, fallbackLang := range GetFallbackChain(i.LangCode) {
		if i.hasKeyInLang(fallbackLang, key) {
			return true
		}
	}

	return false
}

// hasKeyInLang checks if a key exists in a specific language.
func (i I18n) hasKeyInLang(langCode, key string) bool {
	locale, exists := getLocale(langCode)
	if !exists {
		return false
	}

	// Only check key as-is - no automatic prefix fallback
	return locale.IsSet(key)
}

// Convenience functions for common use cases

// GetString is a convenience function that creates an I18n instance and retrieves a string.
func GetString(langCode, key string) string {
	return New(langCode).GetString(key)
}

// GetStringSlice is a convenience function that creates an I18n instance and retrieves a string slice.
func GetStringSlice(langCode, key string) []string {
	return New(langCode).GetStringSlice(key)
}

// HasKey is a convenience function that checks if a key exists.
func HasKey(langCode, key string) bool {
	return New(langCode).HasKey(key)
}
