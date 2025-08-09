package catalog

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Params represents parameters for message interpolation.
// Uses string keys for simple {key} style replacement.
type Params map[string]any

// Message represents a translatable message with embedded English default.
type Message struct {
	// Key is the unique identifier for this message (flat structure like "admin.promote_success")
	Key string `json:"key" yaml:"key"`
	
	// Default is the English text that will be used if no translation exists
	Default string `json:"default" yaml:"default"`
	
	// Params defines the expected parameters for this message
	// This is used for validation and documentation
	Params []string `json:"params,omitempty" yaml:"params,omitempty"`
	
	// Description provides context about when/how this message is used
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// MessageCatalog defines the interface for message registration and retrieval.
type MessageCatalog interface {
	// Register adds a message to the catalog
	Register(msg Message) error
	
	// RegisterBulk adds multiple messages to the catalog
	RegisterBulk(messages []Message) error
	
	// Get retrieves a message by key, returns the message with default text
	Get(key string) (Message, bool)
	
	// Keys returns all registered message keys
	Keys() []string
	
	// Count returns the number of registered messages
	Count() int
	
	// Validate checks if all expected parameters are provided for a message
	Validate(key string, params Params) error
	
	// Clear removes all messages from the catalog (useful for tests)
	Clear()
}

// TranslationProvider defines interface for loading translations from external sources.
type TranslationProvider interface {
	// LoadTranslations loads translations for a specific language
	// Returns a map of key -> translated text
	LoadTranslations(ctx context.Context, lang string) (map[string]string, error)
	
	// SupportedLanguages returns list of supported language codes
	SupportedLanguages(ctx context.Context) ([]string, error)
}

// ParamValidator provides parameter validation functionality.
type ParamValidator interface {
	// ValidateParams checks if provided parameters match expected ones
	ValidateParams(expected []string, provided Params) error
	
	// RequiredParams returns list of required parameters from a text template
	RequiredParams(text string) []string
}

// InterpolationError represents errors during parameter interpolation.
type InterpolationError struct {
	Key          string
	Template     string
	MissingParam string
	Err          error
}

func (e *InterpolationError) Error() string {
	if e.MissingParam != "" {
		return fmt.Sprintf("interpolation error for key '%s': missing parameter '%s'", 
			e.Key, e.MissingParam)
	}
	if e.Err != nil {
		return fmt.Sprintf("interpolation error for key '%s': %v", e.Key, e.Err)
	}
	return fmt.Sprintf("interpolation error for key '%s' in template '%s'", e.Key, e.Template)
}

func (e *InterpolationError) Unwrap() error {
	return e.Err
}

// ValidationError represents parameter validation errors.
type ValidationError struct {
	Key            string
	ExpectedParams []string
	ProvidedParams []string
	MissingParams  []string
	ExtraParams    []string
}

func (e *ValidationError) Error() string {
	var parts []string
	
	if len(e.MissingParams) > 0 {
		parts = append(parts, fmt.Sprintf("missing parameters: %v", e.MissingParams))
	}
	
	if len(e.ExtraParams) > 0 {
		parts = append(parts, fmt.Sprintf("unexpected parameters: %v", e.ExtraParams))
	}
	
	return fmt.Sprintf("validation error for key '%s': %s", e.Key, strings.Join(parts, ", "))
}

// CatalogError represents errors in catalog operations.
type CatalogError struct {
	Operation string
	Key       string
	Err       error
}

func (e *CatalogError) Error() string {
	return fmt.Sprintf("catalog %s error for key '%s': %v", e.Operation, e.Key, e.Err)
}

func (e *CatalogError) Unwrap() error {
	return e.Err
}

// Standard parameter validation implementation
type StandardParamValidator struct {
	// paramRegex matches {param_name} patterns
	paramRegex *regexp.Regexp
}

// NewStandardParamValidator creates a new parameter validator.
func NewStandardParamValidator() *StandardParamValidator {
	return &StandardParamValidator{
		paramRegex: regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`),
	}
}

// ValidateParams checks if provided parameters match expected ones.
func (v *StandardParamValidator) ValidateParams(expected []string, provided Params) error {
	if len(expected) == 0 && len(provided) == 0 {
		return nil
	}
	
	// Convert expected to set for efficient lookup
	expectedSet := make(map[string]bool, len(expected))
	for _, param := range expected {
		expectedSet[param] = true
	}
	
	// Check for missing parameters
	var missing []string
	for _, param := range expected {
		if _, exists := provided[param]; !exists {
			missing = append(missing, param)
		}
	}
	
	// Check for extra parameters
	var extra []string
	providedKeys := make([]string, 0, len(provided))
	for key := range provided {
		providedKeys = append(providedKeys, key)
		if !expectedSet[key] {
			extra = append(extra, key)
		}
	}
	
	if len(missing) > 0 || len(extra) > 0 {
		return &ValidationError{
			ExpectedParams: expected,
			ProvidedParams: providedKeys,
			MissingParams:  missing,
			ExtraParams:    extra,
		}
	}
	
	return nil
}

// RequiredParams extracts parameter names from a text template.
func (v *StandardParamValidator) RequiredParams(text string) []string {
	matches := v.paramRegex.FindAllStringSubmatch(text, -1)
	
	// Use map to deduplicate
	paramSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			paramSet[match[1]] = true
		}
	}
	
	// Convert to slice
	params := make([]string, 0, len(paramSet))
	for param := range paramSet {
		params = append(params, param)
	}
	
	return params
}

// PluralRule defines pluralization forms for different languages.
type PluralRule struct {
	Zero  string `json:"zero,omitempty" yaml:"zero,omitempty"`   // 0 items
	One   string `json:"one,omitempty" yaml:"one,omitempty"`     // 1 item  
	Two   string `json:"two,omitempty" yaml:"two,omitempty"`     // 2 items
	Few   string `json:"few,omitempty" yaml:"few,omitempty"`     // 2-4 items (some languages)
	Many  string `json:"many,omitempty" yaml:"many,omitempty"`   // 5+ items (some languages)
	Other string `json:"other,omitempty" yaml:"other,omitempty"` // default fallback
}

// PluralMessage represents a message with plural forms.
type PluralMessage struct {
	Key         string     `json:"key" yaml:"key"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Default     PluralRule `json:"default" yaml:"default"` // English plural forms
	Params      []string   `json:"params,omitempty" yaml:"params,omitempty"`
}

// PluralSelector defines interface for selecting appropriate plural form.
type PluralSelector interface {
	// SelectForm chooses the correct plural form based on count and language rules
	SelectForm(rule PluralRule, count int, lang string) string
}

// StandardPluralSelector implements basic English pluralization rules.
type StandardPluralSelector struct{}

// NewStandardPluralSelector creates a new standard plural selector.
func NewStandardPluralSelector() *StandardPluralSelector {
	return &StandardPluralSelector{}
}

// SelectForm selects the appropriate plural form.
// For now, implements basic English rules. Can be extended for other languages.
func (s *StandardPluralSelector) SelectForm(rule PluralRule, count int, lang string) string {
	// English pluralization rules
	switch {
	case count == 0 && rule.Zero != "":
		return rule.Zero
	case count == 1 && rule.One != "":
		return rule.One
	case count == 2 && rule.Two != "":
		return rule.Two
	case rule.Other != "":
		return rule.Other
	case rule.Many != "":
		return rule.Many
	case rule.Few != "":
		return rule.Few
	case rule.One != "":
		return rule.One
	default:
		return ""
	}
}

// Config holds configuration for the catalog system.
type Config struct {
	// DefaultLanguage is the fallback language (usually "en")
	DefaultLanguage string
	
	// StrictValidation enables strict parameter validation
	StrictValidation bool
	
	// CacheTranslations enables caching of loaded translations
	CacheTranslations bool
	
	// AllowMissingTranslations allows messages to fall back to default text
	AllowMissingTranslations bool
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		DefaultLanguage:          "en",
		StrictValidation:         false,
		CacheTranslations:        true,
		AllowMissingTranslations: true,
	}
}