package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// ErrInvalidInput represents a generic input validation error
	ErrInvalidInput = errors.New("invalid input")

	// ErrInputTooLong represents an input that exceeds maximum length
	ErrInputTooLong = errors.New("input too long")

	// ErrInputTooShort represents an input that is below minimum length
	ErrInputTooShort = errors.New("input too short")

	// ErrInvalidCharacters represents input containing invalid characters
	ErrInvalidCharacters = errors.New("input contains invalid characters")
)

// Common validation patterns
var (
	// UsernamePattern matches valid Telegram usernames (5-32 chars, alphanumeric + underscore)
	UsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{5,32}$`)

	// CommandPattern matches valid bot commands (1-32 chars, alphanumeric + underscore)
	CommandPattern = regexp.MustCompile(`^[a-zA-Z0-9_]{1,32}$`)

	// SafeTextPattern matches text that doesn't contain potentially dangerous characters
	SafeTextPattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?()]+$`)
)

// ValidateStringLength checks if a string is within the specified length bounds
func ValidateStringLength(input string, minLen, maxLen int) error {
	length := utf8.RuneCountInString(input)

	if length < minLen {
		return ErrInputTooShort
	}

	if length > maxLen {
		return ErrInputTooLong
	}

	return nil
}

// ValidateUsername checks if a username follows Telegram's username rules
func ValidateUsername(username string) error {
	if username == "" {
		return ErrInvalidInput
	}

	// Remove @ prefix if present
	username = strings.TrimPrefix(username, "@")

	if !UsernamePattern.MatchString(username) {
		return ErrInvalidCharacters
	}

	return nil
}

// ValidateCommand checks if a command name is valid
func ValidateCommand(command string) error {
	if command == "" {
		return ErrInvalidInput
	}

	// Remove / prefix if present
	command = strings.TrimPrefix(command, "/")

	if !CommandPattern.MatchString(command) {
		return ErrInvalidCharacters
	}

	return nil
}

// ValidateUserInput sanitizes and validates general user input
func ValidateUserInput(input string, maxLen int) error {
	if input == "" {
		return ErrInvalidInput
	}

	// Check length
	if err := ValidateStringLength(input, 1, maxLen); err != nil {
		return err
	}

	// Check for potentially dangerous characters
	if strings.ContainsAny(input, "<>\"'&;") {
		return ErrInvalidCharacters
	}

	return nil
}

// ValidateChatID checks if a chat ID is within valid range
func ValidateChatID(chatID int64) error {
	// Telegram chat IDs are typically negative for groups/channels
	// and positive for users, with specific ranges
	if chatID == 0 {
		return ErrInvalidInput
	}

	// Basic range check (Telegram uses 64-bit integers)
	if chatID < -1000000000000 || chatID > 1000000000000 {
		return ErrInvalidInput
	}

	return nil
}

// ValidateUserID checks if a user ID is within valid range
func ValidateUserID(userID int64) error {
	// User IDs are positive integers
	if userID <= 0 {
		return ErrInvalidInput
	}

	// Basic range check
	if userID > 1000000000000 {
		return ErrInvalidInput
	}

	return nil
}

// SanitizeText removes potentially dangerous characters from text
func SanitizeText(input string) string {
	// Remove null bytes and other control characters
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.ReplaceAll(input, "\r", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// ValidateFilterKeyword checks if a filter keyword is valid
func ValidateFilterKeyword(keyword string) error {
	if keyword == "" {
		return ErrInvalidInput
	}

	// Check length (reasonable limits for filter keywords)
	if err := ValidateStringLength(keyword, 1, 100); err != nil {
		return err
	}

	// Ensure it doesn't contain only whitespace
	if strings.TrimSpace(keyword) == "" {
		return ErrInvalidInput
	}

	return nil
}
