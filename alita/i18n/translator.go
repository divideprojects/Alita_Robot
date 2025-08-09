package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Translator provides translation functionality using the message catalog.
type Translator struct {
	mu            sync.RWMutex
	lang          string
	config        Config
	translations  map[string]string       // Loaded translations for this language
	pluralRules   map[string]PluralRule   // Plural forms for this language
	catalog       MessageCatalog          // Reference to message catalog
	validator     ParamValidator          // Parameter validator
	pluralSelector PluralSelector         // Plural form selector
}

// newTranslatorWithConfig creates a new translator for the specified language.
func newTranslatorWithConfig(lang string, config Config) *Translator {
	if config.DefaultLanguage == "" {
		config = DefaultConfig()
	}
	
	return &Translator{
		lang:           lang,
		config:         config,
		translations:   make(map[string]string),
		pluralRules:    make(map[string]PluralRule),
		catalog:        globalCatalog,
		validator:      NewStandardParamValidator(),
		pluralSelector: NewStandardPluralSelector(),
	}
}

// LoadFromYAML loads translations from a YAML file.
// The YAML should have flat key-value structure like:
//   admin.promote_success: "Successfully promoted {user}!"
//   admin.demote_success: "Successfully demoted {user}!"
func (t *Translator) LoadFromYAML(yamlPath string) error {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file '%s': %w", yamlPath, err)
	}
	
	return t.LoadFromYAMLBytes(data)
}

// LoadFromYAMLBytes loads translations from YAML bytes.
func (t *Translator) LoadFromYAMLBytes(yamlData []byte) error {
	var flatMap map[string]string
	if err := yaml.Unmarshal(yamlData, &flatMap); err != nil {
		// Try nested structure and flatten it
		var nestedMap map[string]any
		if err2 := yaml.Unmarshal(yamlData, &nestedMap); err2 != nil {
			return fmt.Errorf("failed to parse YAML: %w (original error: %v)", err2, err)
		}
		flatMap = flattenMap(nestedMap, "")
	}
	
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Clear existing translations
	t.translations = make(map[string]string, len(flatMap))
	
	// Load translations
	for key, value := range flatMap {
		t.translations[key] = value
	}
	
	return nil
}

// LoadFromDirectory loads all YAML files from a directory.
// Files should be named like "en.yaml", "es.yaml", etc.
func (t *Translator) LoadFromDirectory(dir string) error {
	filename := fmt.Sprintf("%s.yaml", t.lang)
	yamlPath := filepath.Join(dir, filename)
	
	// Try .yaml first, then .yml
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		filename = fmt.Sprintf("%s.yml", t.lang)
		yamlPath = filepath.Join(dir, filename)
	}
	
	return t.LoadFromYAML(yamlPath)
}

// GetString provides backward compatibility with the old i18n system.
// This method maps old dot-separated keys to new flat keys.
func (t *Translator) GetString(key string) (string, error) {
	// Map old format to new format
	// Remove "strings." prefix and convert module.action.name to module_action_name
	cleanKey := strings.TrimPrefix(key, "strings.")
	cleanKey = strings.ReplaceAll(cleanKey, ".", "_")
	
	return t.GetMessage(cleanKey, nil)
}

// GetStringSlice provides backward compatibility with the old i18n system.
// For now, returns an empty slice as it's mainly used for alt_names which
// can be handled differently in the new system.
func (t *Translator) GetStringSlice(key string) ([]string, error) {
	// This was used for alternative module names
	// For compatibility, return empty slice and let the caller handle defaults
	return []string{}, nil
}

// Message retrieves a translated message with parameter interpolation.
// If no translation exists, returns the default English text from the catalog.
func (t *Translator) Message(key string, params Params) string {
	text, _ := t.GetMessage(key, params)
	return text
}

// GetMessage retrieves a translated message with error handling.
// Returns the translated/default text and any error that occurred.
func (t *Translator) GetMessage(key string, params Params) (string, error) {
	// Get message from catalog (includes default English text)
	msg, exists := t.catalog.Get(key)
	if !exists {
		return fmt.Sprintf("{{%s}}", key), fmt.Errorf("message key '%s' not found in catalog", key)
	}
	
	// Get translation or use default
	text := t.getTranslationOrDefault(key, msg.Default)
	
	// Validate parameters if strict mode is enabled
	if t.config.StrictValidation {
		if err := t.validator.ValidateParams(msg.Params, params); err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				validationErr.Key = key
			}
			return text, err
		}
	}
	
	// Interpolate parameters
	result, err := InterpolateParams(text, params)
	if err != nil {
		if interpErr, ok := err.(*InterpolationError); ok {
			interpErr.Key = key
		}
		return text, err
	}
	
	return result, nil
}

// MessageWithDefault retrieves a message with a custom default.
// Useful when you want to specify default text inline without registering it.
func (t *Translator) MessageWithDefault(key, defaultText string, params Params) string {
	text := t.getTranslationOrDefault(key, defaultText)
	
	result, err := InterpolateParams(text, params)
	if err != nil {
		return text // Return uninterpolated text on error
	}
	
	return result
}

// Plural retrieves a pluralized message based on count.
func (t *Translator) Plural(key string, count int, params Params) string {
	text, _ := t.GetPlural(key, count, params)
	return text
}

// GetPlural retrieves a pluralized message with error handling.
func (t *Translator) GetPlural(key string, count int, params Params) (string, error) {
	// Try to get plural rule from translations
	rule := t.getPluralRule(key)
	if rule == (PluralRule{}) {
		// No plural rule found, try regular message
		return t.GetMessage(key, params)
	}
	
	// Select appropriate form
	selectedText := t.pluralSelector.SelectForm(rule, count, t.lang)
	if selectedText == "" {
		return fmt.Sprintf("{{%s}}", key), fmt.Errorf("no appropriate plural form found for key '%s'", key)
	}
	
	// Add count to parameters
	if params == nil {
		params = make(Params)
	}
	params["count"] = count
	
	// Interpolate parameters
	result, err := InterpolateParams(selectedText, params)
	if err != nil {
		return selectedText, err
	}
	
	return result, nil
}

// HasTranslation checks if a translation exists for the given key.
func (t *Translator) HasTranslation(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	_, exists := t.translations[key]
	return exists
}

// Language returns the language code for this translator.
func (t *Translator) Language() string {
	return t.lang
}

// IsDefaultLanguage checks if this translator uses the default language.
func (t *Translator) IsDefaultLanguage() bool {
	return t.lang == t.config.DefaultLanguage
}

// TranslationCount returns the number of loaded translations.
func (t *Translator) TranslationCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	return len(t.translations)
}

// MissingKeys returns catalog keys that don't have translations in this language.
func (t *Translator) MissingKeys() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	catalogKeys := t.catalog.Keys()
	var missing []string
	
	for _, key := range catalogKeys {
		if _, exists := t.translations[key]; !exists {
			missing = append(missing, key)
		}
	}
	
	return missing
}

// ExtraKeys returns translation keys that don't exist in the catalog.
func (t *Translator) ExtraKeys() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	var extra []string
	
	for key := range t.translations {
		if !HasMessage(key) {
			extra = append(extra, key)
		}
	}
	
	return extra
}

// Coverage returns the percentage of catalog messages that have translations.
func (t *Translator) Coverage() float64 {
	catalogCount := t.catalog.Count()
	if catalogCount == 0 {
		return 100.0
	}
	
	translationCount := t.TranslationCount()
	missing := len(t.MissingKeys())
	
	covered := translationCount - missing
	if covered < 0 {
		covered = 0
	}
	
	return (float64(covered) / float64(catalogCount)) * 100.0
}

// Private helper methods

func (t *Translator) getTranslationOrDefault(key, defaultText string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if translation, exists := t.translations[key]; exists {
		return translation
	}
	
	// Return default text if missing translations are allowed
	if t.config.AllowMissingTranslations {
		return defaultText
	}
	
	// Return key placeholder if translations are required
	return fmt.Sprintf("{{%s}}", key)
}

func (t *Translator) getPluralRule(key string) PluralRule {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if rule, exists := t.pluralRules[key]; exists {
		return rule
	}
	
	// Try to construct plural rule from individual translations
	rule := PluralRule{
		Zero:  t.translations[key+".zero"],
		One:   t.translations[key+".one"],
		Two:   t.translations[key+".two"],
		Few:   t.translations[key+".few"],
		Many:  t.translations[key+".many"],
		Other: t.translations[key+".other"],
	}
	
	// Cache the constructed rule
	t.pluralRules[key] = rule
	return rule
}

// Utility functions

// flattenMap converts a nested map to a flat map with dot-separated keys.
func flattenMap(nested map[string]any, prefix string) map[string]string {
	flat := make(map[string]string)
	
	for key, value := range nested {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case string:
			flat[fullKey] = v
		case map[string]any:
			// Recursively flatten nested maps
			nestedFlat := flattenMap(v, fullKey)
			for nestedKey, nestedValue := range nestedFlat {
				flat[nestedKey] = nestedValue
			}
		case map[any]any:
			// Handle map[any]any from YAML
			stringMap := make(map[string]any)
			for k, v := range v {
				if keyStr, ok := k.(string); ok {
					stringMap[keyStr] = v
				}
			}
			nestedFlat := flattenMap(stringMap, fullKey)
			for nestedKey, nestedValue := range nestedFlat {
				flat[nestedKey] = nestedValue
			}
		default:
			// Convert other types to string
			flat[fullKey] = fmt.Sprintf("%v", v)
		}
	}
	
	return flat
}

// TranslatorManager manages multiple translators for different languages.
type TranslatorManager struct {
	mu           sync.RWMutex
	translators  map[string]*Translator
	config       Config
	localesDir   string
}

// NewTranslatorManager creates a new translator manager.
func NewTranslatorManager(config Config, localesDir string) *TranslatorManager {
	return &TranslatorManager{
		translators: make(map[string]*Translator),
		config:      config,
		localesDir:  localesDir,
	}
}

// GetTranslator returns a translator for the specified language.
// Creates and loads the translator if it doesn't exist.
func (tm *TranslatorManager) GetTranslator(lang string) (*Translator, error) {
	tm.mu.RLock()
	translator, exists := tm.translators[lang]
	tm.mu.RUnlock()
	
	if exists {
		return translator, nil
	}
	
	// Create new translator
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	// Double-check after acquiring write lock
	if translator, exists := tm.translators[lang]; exists {
		return translator, nil
	}
	
	translator = newTranslatorWithConfig(lang, tm.config)
	
	// Load translations if locales directory is specified
	if tm.localesDir != "" {
		if err := translator.LoadFromDirectory(tm.localesDir); err != nil {
			// Don't fail if translation file doesn't exist for non-default languages
			if lang != tm.config.DefaultLanguage && os.IsNotExist(err) {
				// Continue with empty translations
			} else {
				return nil, fmt.Errorf("failed to load translations for language '%s': %w", lang, err)
			}
		}
	}
	
	tm.translators[lang] = translator
	return translator, nil
}

// LoadedLanguages returns a list of loaded language codes.
func (tm *TranslatorManager) LoadedLanguages() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	languages := make([]string, 0, len(tm.translators))
	for lang := range tm.translators {
		languages = append(languages, lang)
	}
	
	return languages
}

// ReloadTranslations reloads translations for all loaded languages.
func (tm *TranslatorManager) ReloadTranslations() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	for lang, translator := range tm.translators {
		if tm.localesDir != "" {
			if err := translator.LoadFromDirectory(tm.localesDir); err != nil {
				return fmt.Errorf("failed to reload translations for language '%s': %w", lang, err)
			}
		}
	}
	
	return nil
}

// ClearTranslators removes all loaded translators.
func (tm *TranslatorManager) ClearTranslators() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	tm.translators = make(map[string]*Translator)
}

// GetStats returns statistics for all loaded translators.
func (tm *TranslatorManager) GetStats() map[string]TranslatorStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	stats := make(map[string]TranslatorStats)
	
	for lang, translator := range tm.translators {
		stats[lang] = TranslatorStats{
			Language:         lang,
			TranslationCount: translator.TranslationCount(),
			MissingKeys:      len(translator.MissingKeys()),
			ExtraKeys:        len(translator.ExtraKeys()),
			Coverage:         translator.Coverage(),
		}
	}
	
	return stats
}

// TranslatorStats holds statistics for a single translator.
type TranslatorStats struct {
	Language         string  `json:"language"`
	TranslationCount int     `json:"translation_count"`
	MissingKeys      int     `json:"missing_keys"`
	ExtraKeys        int     `json:"extra_keys"`
	Coverage         float64 `json:"coverage"`
}

// Global translator manager
var globalManager *TranslatorManager

// InitGlobalManager initializes the global translator manager.
func InitGlobalManager(config Config, localesDir string) {
	globalManager = NewTranslatorManager(config, localesDir)
}

// T returns a translator for the specified language using the global manager.
// This is a convenience function for easy access to translators.
func T(lang string) (*Translator, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("global translator manager not initialized")
	}
	return globalManager.GetTranslator(lang)
}

// MustT returns a translator for the specified language and panics on error.
// Useful when you're confident the language exists.
func MustT(lang string) *Translator {
	translator, err := T(lang)
	if err != nil {
		panic(fmt.Sprintf("failed to get translator for language '%s': %v", lang, err))
	}
	return translator
}
