package security

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/validation"
)

// RateLimiter implements a simple rate limiter for user requests
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[int64][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[int64][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// IsAllowed checks if a user is allowed to make a request
func (rl *RateLimiter) IsAllowed(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get user's request history
	requests := rl.requests[userID]

	// Remove old requests
	var validRequests []time.Time
	for _, req := range requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	// Check if under limit
	if len(validRequests) >= rl.limit {
		rl.requests[userID] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[userID] = validRequests

	return true
}

// cleanup removes old entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)

		for userID, requests := range rl.requests {
			var validRequests []time.Time
			for _, req := range requests {
				if req.After(cutoff) {
					validRequests = append(validRequests, req)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, userID)
			} else {
				rl.requests[userID] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}

// SecurityMiddleware provides security checks for incoming messages
type SecurityMiddleware struct {
	rateLimiter *RateLimiter
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		rateLimiter: NewRateLimiter(30, time.Minute), // 30 requests per minute
	}
}

// ValidateMessage performs security validation on incoming messages
func (sm *SecurityMiddleware) ValidateMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	if msg == nil || user == nil {
		return ext.ContinueGroups
	}

	// Rate limiting
	if !sm.rateLimiter.IsAllowed(user.Id) {
		log.WithFields(log.Fields{
			"user_id":  user.Id,
			"username": user.Username,
		}).Warn("Rate limit exceeded")

		_, err := msg.Reply(b, "⚠️ You're sending messages too quickly. Please slow down.", nil)
		if err != nil {
			log.WithError(err).Error("Failed to send rate limit message")
		}
		return ext.EndGroups
	}

	// Validate message content
	if msg.Text != "" {
		if err := sm.validateTextContent(msg.Text); err != nil {
			log.WithFields(log.Fields{
				"user_id": user.Id,
				"error":   err.Error(),
			}).Warn("Invalid message content")

			_, err := msg.Reply(b, "⚠️ Your message contains invalid content.", nil)
			if err != nil {
				log.WithError(err).Error("Failed to send validation error message")
			}
			return ext.EndGroups
		}
	}

	return ext.ContinueGroups
}

// validateTextContent validates text message content
func (sm *SecurityMiddleware) validateTextContent(text string) error {
	// Check for extremely long messages
	if err := validation.ValidateStringLength(text, 0, 4096); err != nil {
		return fmt.Errorf("message too long: %w", err)
	}

	// Check for suspicious patterns
	if sm.containsSuspiciousPatterns(text) {
		return fmt.Errorf("message contains suspicious patterns")
	}

	// Check for potential injection attempts
	if sm.containsInjectionAttempts(text) {
		return fmt.Errorf("message contains potential injection attempts")
	}

	return nil
}

// containsSuspiciousPatterns checks for suspicious patterns in text
func (sm *SecurityMiddleware) containsSuspiciousPatterns(text string) bool {
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"data:text/html",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
	}

	lowerText := strings.ToLower(text)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}

	return false
}

// containsInjectionAttempts checks for potential injection attempts
func (sm *SecurityMiddleware) containsInjectionAttempts(text string) bool {
	// SQL injection patterns
	sqlPatterns := []string{
		"' OR '1'='1",
		"'; DROP TABLE",
		"UNION SELECT",
		"INSERT INTO",
		"DELETE FROM",
		"UPDATE SET",
	}

	// NoSQL injection patterns
	nosqlPatterns := []string{
		"$where",
		"$regex",
		"$ne",
		"$gt",
		"$lt",
	}

	upperText := strings.ToUpper(text)

	for _, pattern := range sqlPatterns {
		if strings.Contains(upperText, strings.ToUpper(pattern)) {
			return true
		}
	}

	for _, pattern := range nosqlPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}

	return false
}

// ValidateUserInput validates user input with context
func ValidateUserInput(ctx context.Context, input string, maxLength int) error {
	// Check context timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Basic validation
	if err := validation.ValidateUserInput(input, maxLength); err != nil {
		return err
	}

	// Additional security checks
	sanitized := validation.SanitizeText(input)
	if sanitized != input {
		return fmt.Errorf("input contains invalid characters")
	}

	return nil
}

// ValidateCommand validates command input
func ValidateCommand(command string) error {
	// Remove leading slash if present
	command = strings.TrimPrefix(command, "/")

	// Validate command format
	if err := validation.ValidateCommand(command); err != nil {
		return err
	}

	// Check for command injection
	if strings.ContainsAny(command, ";&|`$(){}[]") {
		return fmt.Errorf("command contains invalid characters")
	}

	return nil
}

// ValidateFileUpload validates file uploads
func ValidateFileUpload(filename string, fileSize int64) error {
	// Check file size (10MB limit)
	const maxFileSize = 10 * 1024 * 1024
	if fileSize > maxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", fileSize, maxFileSize)
	}

	// Check filename
	if err := validation.ValidateStringLength(filename, 1, 255); err != nil {
		return fmt.Errorf("invalid filename length: %w", err)
	}

	// Check for dangerous file extensions
	dangerousExts := []string{
		".exe", ".bat", ".cmd", ".com", ".pif", ".scr",
		".vbs", ".js", ".jar", ".php", ".asp", ".jsp",
	}

	lowerFilename := strings.ToLower(filename)
	for _, ext := range dangerousExts {
		if strings.HasSuffix(lowerFilename, ext) {
			return fmt.Errorf("dangerous file type: %s", ext)
		}
	}

	// Check for path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("filename contains path traversal characters")
	}

	return nil
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes
func SanitizeHTML(input string) string {
	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")

	// Remove dangerous attributes
	attrRegex := regexp.MustCompile(`(?i)\s+(on\w+|style|href|src)\s*=\s*["'][^"']*["']`)
	input = attrRegex.ReplaceAllString(input, "")

	// Remove dangerous tags (both opening and closing)
	tagRegex := regexp.MustCompile(`(?i)</?(?:script|iframe|object|embed|form|input|meta|link)[^>]*>`)
	input = tagRegex.ReplaceAllString(input, "")

	return input
}
