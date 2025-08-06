package i18n

import (
	"fmt"
)

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
