package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// MissingKeyEvent represents a missing translation key event
type MissingKeyEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	Key          string    `json:"key"`
	Language     string    `json:"language"`
	FallbackUsed bool      `json:"fallback_used"`
	Environment  string    `json:"environment"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
}

// Logger handles structured logging for i18n events
type Logger struct {
	config      *I18nConfig
	rateLimiter *RateLimiter
	mu          sync.RWMutex
}

// RateLimiter prevents log spam by limiting the frequency of identical events
type RateLimiter struct {
	events    map[string]time.Time
	mu        sync.RWMutex
	threshold time.Duration
}

var (
	globalLogger *Logger
	loggerOnce   sync.Once
)

// NewRateLimiter creates a new rate limiter with the specified threshold
func NewRateLimiter(threshold time.Duration) *RateLimiter {
	return &RateLimiter{
		events:    make(map[string]time.Time),
		threshold: threshold,
	}
}

// ShouldLog checks if an event should be logged based on rate limiting
func (rl *RateLimiter) ShouldLog(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	lastLogged, exists := rl.events[key]

	if !exists || now.Sub(lastLogged) > rl.threshold {
		rl.events[key] = now
		return true
	}

	return false
}

// Cleanup removes old entries from the rate limiter
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, timestamp := range rl.events {
		if now.Sub(timestamp) > rl.threshold*2 {
			delete(rl.events, key)
		}
	}
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		config := GetConfig()
		if config != nil {
			globalLogger = NewLogger(config)
		}
	})
	return globalLogger
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(config *I18nConfig) *Logger {
	logger := &Logger{
		config:      config,
		rateLimiter: NewRateLimiter(5 * time.Minute), // Rate limit to once per 5 minutes per key
	}

	// Start cleanup goroutine
	go logger.startCleanup()

	return logger
}

// startCleanup runs periodic cleanup of the rate limiter
func (l *Logger) startCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.rateLimiter.Cleanup()
	}
}

// LogMissingKey logs a missing translation key event
func (l *Logger) LogMissingKey(key, language string, fallbackUsed bool) {
	if !l.config.LogMissingKeys {
		return
	}

	// Create a unique identifier for rate limiting
	rateKey := fmt.Sprintf("%s:%s", key, language)

	if !l.rateLimiter.ShouldLog(rateKey) {
		return
	}

	event := MissingKeyEvent{
		Timestamp:    time.Now(),
		Key:          key,
		Language:     language,
		FallbackUsed: fallbackUsed,
		Environment:  l.config.Environment,
		Level:        "warning",
		Message:      fmt.Sprintf("Missing translation key '%s' for language '%s'", key, language),
	}

	l.logEvent(event)
}

// LogFallbackUsed logs when a fallback chain is used
func (l *Logger) LogFallbackUsed(key, originalLang, fallbackLang string) {
	if !l.config.LogMissingKeys {
		return
	}

	// Create a unique identifier for rate limiting
	rateKey := fmt.Sprintf("fallback:%s:%s->%s", key, originalLang, fallbackLang)

	if !l.rateLimiter.ShouldLog(rateKey) {
		return
	}

	event := MissingKeyEvent{
		Timestamp:    time.Now(),
		Key:          key,
		Language:     originalLang,
		FallbackUsed: true,
		Environment:  l.config.Environment,
		Level:        "info",
		Message:      fmt.Sprintf("Used fallback language '%s' for key '%s' (original: '%s')", fallbackLang, key, originalLang),
	}

	l.logEvent(event)
}

// LogKeyNotFound logs when a key is not found in any language
func (l *Logger) LogKeyNotFound(key, language string) {
	if !l.config.LogMissingKeys {
		return
	}

	// Create a unique identifier for rate limiting
	rateKey := fmt.Sprintf("notfound:%s:%s", key, language)

	if !l.rateLimiter.ShouldLog(rateKey) {
		return
	}

	event := MissingKeyEvent{
		Timestamp:    time.Now(),
		Key:          key,
		Language:     language,
		FallbackUsed: false,
		Environment:  l.config.Environment,
		Level:        "error",
		Message:      fmt.Sprintf("Translation key '%s' not found in language '%s' or any fallback", key, language),
	}

	l.logEvent(event)
}

// logEvent outputs the event based on configuration
func (l *Logger) logEvent(event MissingKeyEvent) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.config.EnableStructuredLogging {
		l.logStructured(event)
	} else {
		l.logSimple(event)
	}
}

// logStructured outputs structured JSON logs
func (l *Logger) logStructured(event MissingKeyEvent) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("[i18n] Failed to marshal log event: %v", err)
		return
	}

	// Use different log levels based on event severity
	switch event.Level {
	case "error":
		log.Printf("[i18n:ERROR] %s", string(jsonData))
	case "warning":
		log.Printf("[i18n:WARN] %s", string(jsonData))
	case "info":
		log.Printf("[i18n:INFO] %s", string(jsonData))
	default:
		log.Printf("[i18n] %s", string(jsonData))
	}
}

// logSimple outputs simple text logs
func (l *Logger) logSimple(event MissingKeyEvent) {
	switch event.Level {
	case "error":
		log.Printf("[i18n:ERROR] %s", event.Message)
	case "warning":
		log.Printf("[i18n:WARN] %s", event.Message)
	case "info":
		log.Printf("[i18n:INFO] %s", event.Message)
	default:
		log.Printf("[i18n] %s", event.Message)
	}
}

// GetStats returns statistics about missing keys
func (l *Logger) GetStats() map[string]interface{} {
	l.rateLimiter.mu.RLock()
	defer l.rateLimiter.mu.RUnlock()

	stats := map[string]interface{}{
		"total_tracked_keys":   len(l.rateLimiter.events),
		"rate_limit_threshold": l.rateLimiter.threshold.String(),
		"logging_enabled":      l.config.LogMissingKeys,
		"structured_logging":   l.config.EnableStructuredLogging,
	}

	return stats
}

// ResetStats clears all tracked events (useful for testing)
func (l *Logger) ResetStats() {
	l.rateLimiter.mu.Lock()
	defer l.rateLimiter.mu.Unlock()

	l.rateLimiter.events = make(map[string]time.Time)
}

// Convenience functions for global logger

// LogMissingKey logs a missing key using the global logger
func LogMissingKey(key, language string, fallbackUsed bool) {
	if logger := GetLogger(); logger != nil {
		logger.LogMissingKey(key, language, fallbackUsed)
	}
}

// LogFallbackUsed logs fallback usage using the global logger
func LogFallbackUsed(key, originalLang, fallbackLang string) {
	if logger := GetLogger(); logger != nil {
		logger.LogFallbackUsed(key, originalLang, fallbackLang)
	}
}

// LogKeyNotFound logs when a key is not found using the global logger
func LogKeyNotFound(key, language string) {
	if logger := GetLogger(); logger != nil {
		logger.LogKeyNotFound(key, language)
	}
}
