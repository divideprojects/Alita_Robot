package validation

import (
	"testing"
)

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		minLen  int
		maxLen  int
		wantErr bool
	}{
		{"Valid length", "hello", 1, 10, false},
		{"Too short", "hi", 5, 10, true},
		{"Too long", "this is a very long string", 1, 10, true},
		{"Exact min length", "hello", 5, 10, false},
		{"Exact max length", "hello", 1, 5, false},
		{"Empty string with min 0", "", 0, 10, false},
		{"Empty string with min 1", "", 1, 10, true},
		{"Unicode characters", "héllo", 1, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLength(tt.input, tt.minLen, tt.maxLen)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "testuser", false},
		{"Valid username with @", "@testuser", false},
		{"Valid username with numbers", "user123", false},
		{"Valid username with underscore", "test_user", false},
		{"Too short", "test", true},
		{"Too long", "this_is_a_very_long_username_that_exceeds_limit", true},
		{"Invalid characters", "test-user", true},
		{"Invalid characters with space", "test user", true},
		{"Empty string", "", true},
		{"Only @", "@", true},
		{"Special characters", "test@user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"Valid command", "start", false},
		{"Valid command with /", "/start", false},
		{"Valid command with numbers", "help123", false},
		{"Valid command with underscore", "test_command", false},
		{"Too long", "this_is_a_very_long_command_name_that_exceeds_the_limit", true},
		{"Invalid characters", "test-command", true},
		{"Invalid characters with space", "test command", true},
		{"Empty string", "", true},
		{"Only /", "/", true},
		{"Special characters", "test@command", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		maxLen  int
		wantErr bool
	}{
		{"Valid input", "Hello world", 50, false},
		{"Valid input with numbers", "Test 123", 50, false},
		{"Too long", "This is a very long string that exceeds the maximum length", 20, true},
		{"Empty string", "", 50, true},
		{"Dangerous characters", "Hello <script>", 50, true},
		{"SQL injection attempt", "'; DROP TABLE users; --", 50, true},
		{"XSS attempt", "<script>alert('xss')</script>", 50, true},
		{"Valid punctuation", "Hello, world!", 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserInput(tt.input, tt.maxLen)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateChatID(t *testing.T) {
	tests := []struct {
		name    string
		chatID  int64
		wantErr bool
	}{
		{"Valid positive chat ID", 123456789, false},
		{"Valid negative chat ID", -123456789, false},
		{"Zero chat ID", 0, true},
		{"Very large positive", 999999999999, false},
		{"Very large negative", -999999999999, false},
		{"Too large positive", 1000000000001, true},
		{"Too large negative", -1000000000001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChatID(tt.chatID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChatID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		wantErr bool
	}{
		{"Valid user ID", 123456789, false},
		{"Zero user ID", 0, true},
		{"Negative user ID", -123456789, true},
		{"Very large user ID", 999999999999, false},
		{"Too large user ID", 1000000000001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal text", "Hello world", "Hello world"},
		{"Text with null bytes", "Hello\x00world", "Helloworld"},
		{"Text with carriage returns", "Hello\rworld", "Helloworld"},
		{"Text with leading/trailing spaces", "  Hello world  ", "Hello world"},
		{"Text with multiple issues", "  Hello\x00\rworld  ", "Helloworld"},
		{"Empty string", "", ""},
		{"Only whitespace", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeText(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateFilterKeyword(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		wantErr bool
	}{
		{"Valid keyword", "badword", false},
		{"Valid keyword with spaces", "bad word", false},
		{"Empty string", "", true},
		{"Only whitespace", "   ", true},
		{"Too long keyword", "this_is_a_very_long_keyword_that_exceeds_the_reasonable_limit_for_filter_keywords_in_the_system_and_should_fail_validation", true},
		{"Valid keyword with special chars", "test@keyword", false},
		{"Single character", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilterKeyword(tt.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilterKeyword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
