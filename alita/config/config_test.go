package config

import (
	"os"
	"testing"
)

func TestLoadConfigWithValidData(t *testing.T) {
	// Set up valid environment variables
	os.Setenv("BOT_TOKEN", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	os.Setenv("DB_URI", "mongodb://localhost:27017/test")
	os.Setenv("OWNER_ID", "123456789")
	os.Setenv("MESSAGE_DUMP", "-987654321")
	os.Setenv("DEBUG", "true")

	defer func() {
		// Clean up
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("DB_URI")
		os.Unsetenv("OWNER_ID")
		os.Unsetenv("MESSAGE_DUMP")
		os.Unsetenv("DEBUG")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.BotToken != "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11" {
		t.Errorf("Expected BotToken to be set correctly")
	}

	if cfg.DatabaseURI != "mongodb://localhost:27017/test" {
		t.Errorf("Expected DatabaseURI to be set correctly")
	}

	if cfg.OwnerId != 123456789 {
		t.Errorf("Expected OwnerId to be 123456789, got %d", cfg.OwnerId)
	}

	if cfg.MessageDump != -987654321 {
		t.Errorf("Expected MessageDump to be -987654321, got %d", cfg.MessageDump)
	}

	if !cfg.Debug {
		t.Errorf("Expected Debug to be true")
	}

	// Check defaults
	if cfg.BotVersion != DefaultBotVersion {
		t.Errorf("Expected BotVersion to be default value")
	}

	if cfg.ApiServer != DefaultApiServer {
		t.Errorf("Expected ApiServer to be default value")
	}
}

func TestLoadConfigMissingRequired(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("BOT_TOKEN")
	os.Unsetenv("DB_URI")
	os.Unsetenv("OWNER_ID")
	os.Unsetenv("MESSAGE_DUMP")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing required fields")
	}

	configErr, ok := err.(*ConfigValidationError)
	if !ok {
		t.Fatalf("Expected ConfigValidationError, got %T", err)
	}

	if len(configErr.Errors) != 4 {
		t.Errorf("Expected 4 validation errors, got %d", len(configErr.Errors))
	}
}

func TestLoadConfigInvalidNumbers(t *testing.T) {
	os.Setenv("BOT_TOKEN", "valid_token")
	os.Setenv("DB_URI", "mongodb://localhost:27017/test")
	os.Setenv("OWNER_ID", "not_a_number")
	os.Setenv("MESSAGE_DUMP", "also_not_a_number")

	defer func() {
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("DB_URI")
		os.Unsetenv("OWNER_ID")
		os.Unsetenv("MESSAGE_DUMP")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for invalid number fields")
	}

	configErr, ok := err.(*ConfigValidationError)
	if !ok {
		t.Fatalf("Expected ConfigValidationError, got %T", err)
	}

	if len(configErr.Errors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(configErr.Errors))
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"YES", true},
		{"1", true},
		{"false", false},
		{"FALSE", false},
		{"no", false},
		{"NO", false},
		{"0", false},
		{"", false},       // default for empty
		{"random", false}, // default for unknown
	}

	for _, test := range tests {
		os.Setenv("TEST_BOOL", test.value)
		result := getBool("TEST_BOOL", false)
		if result != test.expected {
			t.Errorf("getBool(%q) = %v, expected %v", test.value, result, test.expected)
		}
	}

	os.Unsetenv("TEST_BOOL")
}

func TestGetStringSlice(t *testing.T) {
	tests := []struct {
		value    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a, b, c", []string{"a", "b", "c"}}, // with spaces
		{"a,,c", []string{"a", "c"}},         // empty values removed
		{"", []string{"default"}},            // empty string returns default
	}

	defaultVal := []string{"default"}

	for _, test := range tests {
		os.Setenv("TEST_SLICE", test.value)
		result := getStringSlice("TEST_SLICE", defaultVal)

		if len(result) != len(test.expected) {
			t.Errorf("getStringSlice(%q) length = %d, expected %d", test.value, len(result), len(test.expected))
			continue
		}

		for i, v := range result {
			if v != test.expected[i] {
				t.Errorf("getStringSlice(%q)[%d] = %q, expected %q", test.value, i, v, test.expected[i])
			}
		}
	}

	os.Unsetenv("TEST_SLICE")
}
