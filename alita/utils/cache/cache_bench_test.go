package cache

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// BenchmarkGetAdminCacheUser benchmarks the O(1) user lookup in AdminCache
func BenchmarkGetAdminCacheUser(b *testing.B) {
	// Create a mock admin cache with multiple users
	adminCache := AdminCache{
		ChatId: -1001234567890,
		UserInfo: []gotgbot.MergedChatMember{
			{User: gotgbot.User{Id: 123456789, FirstName: "Admin1"}},
			{User: gotgbot.User{Id: 123456790, FirstName: "Admin2"}},
			{User: gotgbot.User{Id: 123456791, FirstName: "Admin3"}},
			{User: gotgbot.User{Id: 123456792, FirstName: "Admin4"}},
			{User: gotgbot.User{Id: 123456793, FirstName: "Admin5"}},
			{User: gotgbot.User{Id: 123456794, FirstName: "Admin6"}},
			{User: gotgbot.User{Id: 123456795, FirstName: "Admin7"}},
			{User: gotgbot.User{Id: 123456796, FirstName: "Admin8"}},
			{User: gotgbot.User{Id: 123456797, FirstName: "Admin9"}},
			{User: gotgbot.User{Id: 123456798, FirstName: "Admin10"}},
		},
		Cached: true,
	}

	// Build the user map for O(1) lookups
	adminCache.buildUserMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adminCache.GetUser(123456795) // Look for Admin7
	}
}

// BenchmarkBuildUserMap benchmarks the user map building process
func BenchmarkBuildUserMap(b *testing.B) {
	adminCache := AdminCache{
		ChatId: -1001234567890,
		UserInfo: []gotgbot.MergedChatMember{
			{User: gotgbot.User{Id: 123456789, FirstName: "Admin1"}},
			{User: gotgbot.User{Id: 123456790, FirstName: "Admin2"}},
			{User: gotgbot.User{Id: 123456791, FirstName: "Admin3"}},
			{User: gotgbot.User{Id: 123456792, FirstName: "Admin4"}},
			{User: gotgbot.User{Id: 123456793, FirstName: "Admin5"}},
			{User: gotgbot.User{Id: 123456794, FirstName: "Admin6"}},
			{User: gotgbot.User{Id: 123456795, FirstName: "Admin7"}},
			{User: gotgbot.User{Id: 123456796, FirstName: "Admin8"}},
			{User: gotgbot.User{Id: 123456797, FirstName: "Admin9"}},
			{User: gotgbot.User{Id: 123456798, FirstName: "Admin10"}},
		},
		Cached: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adminCache.buildUserMap()
	}
}

// BenchmarkLinearSearchComparison benchmarks the old linear search approach for comparison
func BenchmarkLinearSearchComparison(b *testing.B) {
	userList := []gotgbot.MergedChatMember{
		{User: gotgbot.User{Id: 123456789, FirstName: "Admin1"}},
		{User: gotgbot.User{Id: 123456790, FirstName: "Admin2"}},
		{User: gotgbot.User{Id: 123456791, FirstName: "Admin3"}},
		{User: gotgbot.User{Id: 123456792, FirstName: "Admin4"}},
		{User: gotgbot.User{Id: 123456793, FirstName: "Admin5"}},
		{User: gotgbot.User{Id: 123456794, FirstName: "Admin6"}},
		{User: gotgbot.User{Id: 123456795, FirstName: "Admin7"}},
		{User: gotgbot.User{Id: 123456796, FirstName: "Admin8"}},
		{User: gotgbot.User{Id: 123456797, FirstName: "Admin9"}},
		{User: gotgbot.User{Id: 123456798, FirstName: "Admin10"}},
	}

	searchUserId := int64(123456795)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate old linear search
		found := false
		for _, admin := range userList {
			if admin.User.Id == searchUserId {
				found = true
				break
			}
		}
		_ = found
	}
}
