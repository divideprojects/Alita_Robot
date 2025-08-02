package logging

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SensitivePatterns contains regex patterns for detecting sensitive information
var SensitivePatterns = []*regexp.Regexp{
	// Bot tokens (Telegram format: digits:alphanumeric)
	regexp.MustCompile(`\b\d{8,10}:[a-zA-Z0-9_-]{35}\b`),

	// API keys (common formats)
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret)["\s]*[:=]["\s]*[a-zA-Z0-9_-]{20,}`),

	// Database URIs
	regexp.MustCompile(`mongodb://[^@\s]+:[^@\s]+@[^\s]+`),
	regexp.MustCompile(`redis://[^@\s]+:[^@\s]+@[^\s]+`),

	// Passwords in various formats
	regexp.MustCompile(`(?i)(password|pwd|pass)["\s]*[:=]["\s]*[^\s"]{6,}`),

	// Credit card numbers (basic pattern)
	regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),

	// Email addresses (when they might contain sensitive info)
	regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),

	// IP addresses (might be sensitive in some contexts)
	regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
}

// SanitizeMessage removes or masks sensitive information from log messages
func SanitizeMessage(message string) string {
	sanitized := message

	for _, pattern := range SensitivePatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			// Different masking strategies based on the type of sensitive data
			if strings.Contains(strings.ToLower(match), "token") ||
				strings.Contains(strings.ToLower(match), "key") ||
				strings.Contains(strings.ToLower(match), "secret") {
				return "[REDACTED_TOKEN]"
			}

			if strings.Contains(match, "@") && strings.Contains(match, "mongodb://") {
				return "[REDACTED_DB_URI]"
			}

			if strings.Contains(match, "@") && strings.Contains(match, "redis://") {
				return "[REDACTED_REDIS_URI]"
			}

			if strings.Contains(strings.ToLower(match), "password") ||
				strings.Contains(strings.ToLower(match), "pwd") {
				return "[REDACTED_PASSWORD]"
			}

			if strings.Contains(match, "@") {
				return "[REDACTED_EMAIL]"
			}

			// For credit cards and IPs, show partial info
			if len(match) >= 16 && regexp.MustCompile(`\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}`).MatchString(match) {
				return "****-****-****-" + match[len(match)-4:]
			}

			if regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`).MatchString(match) {
				parts := strings.Split(match, ".")
				if len(parts) == 4 {
					return parts[0] + ".***.***.***"
				}
			}

			return "[REDACTED]"
		})
	}

	return sanitized
}

// SecureLogger wraps logrus with automatic sanitization
type SecureLogger struct {
	logger *log.Logger
}

// NewSecureLogger creates a new secure logger instance
func NewSecureLogger() *SecureLogger {
	return &SecureLogger{
		logger: log.StandardLogger(),
	}
}

// Info logs an info message with sanitization
func (sl *SecureLogger) Info(args ...interface{}) {
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Info(sanitizedArgs...)
}

// Infof logs a formatted info message with sanitization
func (sl *SecureLogger) Infof(format string, args ...interface{}) {
	sanitizedFormat := SanitizeMessage(format)
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Infof(sanitizedFormat, sanitizedArgs...)
}

// Error logs an error message with sanitization
func (sl *SecureLogger) Error(args ...interface{}) {
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Error(sanitizedArgs...)
}

// Errorf logs a formatted error message with sanitization
func (sl *SecureLogger) Errorf(format string, args ...interface{}) {
	sanitizedFormat := SanitizeMessage(format)
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Errorf(sanitizedFormat, sanitizedArgs...)
}

// Warn logs a warning message with sanitization
func (sl *SecureLogger) Warn(args ...interface{}) {
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Warn(sanitizedArgs...)
}

// Warnf logs a formatted warning message with sanitization
func (sl *SecureLogger) Warnf(format string, args ...interface{}) {
	sanitizedFormat := SanitizeMessage(format)
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Warnf(sanitizedFormat, sanitizedArgs...)
}

// Debug logs a debug message with sanitization
func (sl *SecureLogger) Debug(args ...interface{}) {
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeMessage(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	sl.logger.Debug(sanitizedArgs...)
}

// WithFields creates a new entry with sanitized fields
func (sl *SecureLogger) WithFields(fields log.Fields) *log.Entry {
	sanitizedFields := make(log.Fields)
	for key, value := range fields {
		if str, ok := value.(string); ok {
			sanitizedFields[key] = SanitizeMessage(str)
		} else {
			sanitizedFields[key] = value
		}
	}
	return sl.logger.WithFields(sanitizedFields)
}

// WithError creates a new entry with an error field
func (sl *SecureLogger) WithError(err error) *log.Entry {
	if err != nil {
		sanitizedError := SanitizeMessage(err.Error())
		return sl.logger.WithField("error", sanitizedError)
	}
	return sl.logger.WithError(err)
}
