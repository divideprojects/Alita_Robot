package i18n

import (
	"fmt"
)

// I18nError represents all i18n related errors
type I18nError struct {
	Op      string // Operation that failed
	Lang    string // Language code involved
	Key     string // Translation key involved
	Message string // Error message
	Err     error  // Underlying error
}

func (e *I18nError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("i18n %s failed for lang=%s key=%s: %s: %v", e.Op, e.Lang, e.Key, e.Message, e.Err)
	}
	return fmt.Sprintf("i18n %s failed for lang=%s key=%s: %s", e.Op, e.Lang, e.Key, e.Message)
}

func (e *I18nError) Unwrap() error {
	return e.Err
}

// NewI18nError creates a new i18n error
func NewI18nError(op, lang, key, message string, err error) *I18nError {
	return &I18nError{
		Op:      op,
		Lang:    lang,
		Key:     key,
		Message: message,
		Err:     err,
	}
}

// Predefined error types
var (
	ErrLocaleNotFound    = fmt.Errorf("locale not found")
	ErrKeyNotFound       = fmt.Errorf("translation key not found")
	ErrInvalidYAML       = fmt.Errorf("invalid YAML format")
	ErrManagerNotInit    = fmt.Errorf("locale manager not initialized")
	ErrRecursiveFallback = fmt.Errorf("recursive fallback detected")
	ErrCacheFailure      = fmt.Errorf("cache operation failed")
	ErrInvalidParams     = fmt.Errorf("invalid translation parameters")
)

// IsI18nError checks if an error is an I18nError
func IsI18nError(err error) bool {
	_, ok := err.(*I18nError)
	return ok
}

// IsLocaleNotFound checks if error is due to locale not being found
func IsLocaleNotFound(err error) bool {
	if i18nErr, ok := err.(*I18nError); ok {
		return i18nErr.Err == ErrLocaleNotFound
	}
	return err == ErrLocaleNotFound
}

// IsKeyNotFound checks if error is due to key not being found
func IsKeyNotFound(err error) bool {
	if i18nErr, ok := err.(*I18nError); ok {
		return i18nErr.Err == ErrKeyNotFound
	}
	return err == ErrKeyNotFound
}
