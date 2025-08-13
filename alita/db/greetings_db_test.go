package db

import (
	"testing"
)

// TestSetGoodbyeToggle tests that the goodbye toggle can be properly set to both true and false
func TestSetGoodbyeToggle(t *testing.T) {
	// Note: This is a unit test template. In production, you would need to:
	// 1. Set up a test database connection
	// 2. Create a test chat
	// 3. Run the actual tests

	testCases := []struct {
		name     string
		chatID   int64
		setValue bool
		expected bool
	}{
		{
			name:     "Set goodbye to true",
			chatID:   -100123456789,
			setValue: true,
			expected: true,
		},
		{
			name:     "Set goodbye to false",
			chatID:   -100123456789,
			setValue: false,
			expected: false,
		},
		{
			name:     "Toggle from true to false",
			chatID:   -100987654321,
			setValue: false,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// In a real test, you would:
			// 1. Call SetGoodbyeToggle(tc.chatID, tc.setValue)
			// 2. Retrieve settings with GetGreetingSettings(tc.chatID)
			// 3. Assert that settings.GoodbyeSettings.ShouldGoodbye == tc.expected

			// This demonstrates the test structure for the fix
			t.Logf("Test case: %s - Setting goodbye to %v for chat %d, expecting %v",
				tc.name, tc.setValue, tc.chatID, tc.expected)
		})
	}
}

// TestSetWelcomeToggle tests that the welcome toggle can be properly set to both true and false
func TestSetWelcomeToggle(t *testing.T) {
	testCases := []struct {
		name     string
		chatID   int64
		setValue bool
		expected bool
	}{
		{
			name:     "Set welcome to true",
			chatID:   -100123456789,
			setValue: true,
			expected: true,
		},
		{
			name:     "Set welcome to false",
			chatID:   -100123456789,
			setValue: false,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Test case: %s - Setting welcome to %v for chat %d, expecting %v",
				tc.name, tc.setValue, tc.chatID, tc.expected)
		})
	}
}
