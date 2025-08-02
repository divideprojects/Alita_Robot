package chat_status

import (
	"testing"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

func TestTgAdminList(t *testing.T) {
	// Test that the anonymous bot ID is in the admin list
	if !string_handling.FindInInt64Slice(tgAdminList, groupAnonymousBot) {
		t.Errorf("Expected groupAnonymousBot (%d) to be in tgAdminList", groupAnonymousBot)
	}

	// Test that a random user ID is not in the admin list
	randomUserId := int64(12345)
	if string_handling.FindInInt64Slice(tgAdminList, randomUserId) {
		t.Errorf("Expected random user ID (%d) to not be in tgAdminList", randomUserId)
	}
}

func TestConstants(t *testing.T) {
	// Verify that the constants have expected values
	expectedGroupAnonymousBot := int64(1087968824)
	expectedTgUserId := int64(777000)

	if groupAnonymousBot != expectedGroupAnonymousBot {
		t.Errorf("Expected groupAnonymousBot to be %d, got %d", expectedGroupAnonymousBot, groupAnonymousBot)
	}

	if tgUserId != expectedTgUserId {
		t.Errorf("Expected tgUserId to be %d, got %d", expectedTgUserId, tgUserId)
	}
}

// Test helper function to validate admin list behavior
func TestAdminListBehavior(t *testing.T) {
	tests := []struct {
		name     string
		userId   int64
		expected bool
	}{
		{"Anonymous bot should be admin", groupAnonymousBot, true},
		{"Regular user should not be admin", 123456789, false},
		{"Zero ID should not be admin", 0, false},
		{"Negative ID should not be admin", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string_handling.FindInInt64Slice(tgAdminList, tt.userId)
			if result != tt.expected {
				t.Errorf("FindInInt64Slice(%d) = %v, want %v", tt.userId, result, tt.expected)
			}
		})
	}
}
