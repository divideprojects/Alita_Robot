package modules

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	tb := newTokenBucket(5, time.Second/5) // 5 tokens, refill 1 per 200ms

	// Test initial burst
	for i := 0; i < 5; i++ {
		if !tb.take() {
			t.Error("should have tokens available initially")
		}
	}

	// Should be rate limited after burst
	if tb.take() {
		t.Error("should be rate limited after burst")
	}

	// Test refill
	time.Sleep(300 * time.Millisecond) // Should have refilled 1 token
	if !tb.take() {
		t.Error("should have refilled a token")
	}
	if tb.take() {
		t.Error("should be rate limited again")
	}
}

func TestUpdateFlood(t *testing.T) {
	// Setup test data
	chatId := int64(123)
	userId := int64(456)
	msgId := int64(789)

	// Initialize test flood settings
	antifloodModule.syncHelperMap = sync.Map{}

	// First call - should not be flooded
	flooded, bucket := antifloodModule.updateFlood(chatId, userId, msgId)
	if flooded {
		t.Error("should not be flooded after first message")
	}
	if bucket == nil {
		t.Error("should return a bucket")
	}

	// Second call - still not flooded
	flooded, _ = antifloodModule.updateFlood(chatId, userId, msgId+1)
	if flooded {
		t.Error("should not be flooded after second message")
	}

	// Third call - should be flooded
	flooded, bucket = antifloodModule.updateFlood(chatId, userId, msgId+2)
	if !flooded {
		t.Error("should be flooded after third message")
	}

	// Verify bucket was reset
	flooded, newBucket := antifloodModule.updateFlood(chatId, userId, msgId+3)
	if flooded {
		t.Error("should not be flooded for new user")
	}
	if bucket == newBucket {
		t.Error("should have new bucket after flood")
	}
}
