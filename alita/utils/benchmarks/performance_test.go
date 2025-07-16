package benchmarks

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// Test data for benchmarks
var (
	testFilterKeywords = []string{
		"spam", "scam", "fake", "bot", "hack", "virus", "malware", "phishing",
		"promotion", "advertisement", "sell", "buy", "money", "crypto", "bitcoin",
		"badword1", "badword2", "badword3", "badword4", "badword5",
	}

	testMessages = []string{
		"This is a normal message",
		"Check out this spam offer!",
		"Don't fall for this scam",
		"This fake website is dangerous",
		"Click here to buy crypto",
		"This message contains badword3 somewhere",
		"A very long message with lots of text but no filtered words in it at all",
		"Short msg",
		"spam scam fake - multiple matches",
		"No matches here just regular conversation",
	}

	largeFilterKeywords = generateLargeFilterKeywords(100)
	largeMessages       = generateLargeMessages(50)
)

// Generate large test datasets
func generateLargeFilterKeywords(count int) []string {
	keywords := make([]string, count)
	for i := 0; i < count; i++ {
		keywords[i] = fmt.Sprintf("keyword%d", i)
	}
	return keywords
}

func generateLargeMessages(count int) []string {
	messages := make([]string, count)
	for i := 0; i < count; i++ {
		if i%10 == 0 {
			// 10% of messages contain a keyword
			messages[i] = fmt.Sprintf("This message contains keyword%d for testing", i%20)
		} else {
			// 90% are regular messages
			messages[i] = fmt.Sprintf("Regular message number %d with no filtered content", i)
		}
	}
	return messages
}

// Original implementation (for comparison)
func checkFiltersOriginal(keywords []string, text string) (string, bool) {
	lowerText := strings.ToLower(text)
	for _, keyword := range keywords {
		pattern := fmt.Sprintf(`(\b|\s)%s\b`, regexp.QuoteMeta(keyword))
		if matched, _ := regexp.MatchString(pattern, lowerText); matched {
			return keyword, true
		}
	}
	return "", false
}

// Optimized implementation
func buildFilterRegex(keywords []string) *regexp.Regexp {
	if len(keywords) == 0 {
		return nil
	}

	// Escape special regex characters in keywords and build pattern
	escapedKeywords := make([]string, len(keywords))
	for i, keyword := range keywords {
		escapedKeywords[i] = regexp.QuoteMeta(keyword)
	}

	// Create pattern that matches any keyword with word boundaries
	pattern := fmt.Sprintf(`(\b|\s)(%s)\b`, strings.Join(escapedKeywords, "|"))

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	return regex
}

func findMatchingKeyword(regex *regexp.Regexp, text string, keywords []string) (string, bool) {
	lowerText := strings.ToLower(text)

	// First check if any keyword matches
	if !regex.MatchString(lowerText) {
		return "", false
	}

	// Find which specific keyword matched (fallback to individual checks)
	for _, keyword := range keywords {
		keywordPattern := fmt.Sprintf(`(\b|\s)%s\b`, regexp.QuoteMeta(keyword))
		if matched, _ := regexp.MatchString(keywordPattern, lowerText); matched {
			return keyword, true
		}
	}

	return "", false
}

func checkFiltersOptimized(keywords []string, text string) (string, bool) {
	regex := buildFilterRegex(keywords)
	if regex == nil {
		return "", false
	}
	return findMatchingKeyword(regex, text, keywords)
}

// Benchmarks comparing original vs optimized approaches

func BenchmarkFilters_Original_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, message := range testMessages {
			checkFiltersOriginal(testFilterKeywords, message)
		}
	}
}

func BenchmarkFilters_Optimized_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, message := range testMessages {
			checkFiltersOptimized(testFilterKeywords, message)
		}
	}
}

func BenchmarkFilters_Original_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, message := range largeMessages {
			checkFiltersOriginal(largeFilterKeywords, message)
		}
	}
}

func BenchmarkFilters_Optimized_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, message := range largeMessages {
			checkFiltersOptimized(largeFilterKeywords, message)
		}
	}
}

// Benchmark realistic chat scenario
func BenchmarkRealisticChatScenario_Original(b *testing.B) {
	// Simulate 100 messages per second with 20 filter keywords
	messages := generateLargeMessages(100)
	keywords := testFilterKeywords

	for i := 0; i < b.N; i++ {
		for _, message := range messages {
			checkFiltersOriginal(keywords, message)
		}
	}
}

func BenchmarkRealisticChatScenario_Optimized(b *testing.B) {
	// Simulate 100 messages per second with 20 filter keywords
	messages := generateLargeMessages(100)
	keywords := testFilterKeywords

	for i := 0; i < b.N; i++ {
		for _, message := range messages {
			checkFiltersOptimized(keywords, message)
		}
	}
}

// Worst case scenario: many keywords, no matches
func BenchmarkWorstCase_Original(b *testing.B) {
	keywords := generateLargeFilterKeywords(200)
	message := "This is a very long message with lots of words but none of them match any filter keywords at all"

	for i := 0; i < b.N; i++ {
		checkFiltersOriginal(keywords, message)
	}
}

func BenchmarkWorstCase_Optimized(b *testing.B) {
	keywords := generateLargeFilterKeywords(200)
	message := "This is a very long message with lots of words but none of them match any filter keywords at all"

	for i := 0; i < b.N; i++ {
		checkFiltersOptimized(keywords, message)
	}
}

// Test correctness of optimized implementation
func TestFilterOptimization_Correctness(b *testing.T) {
	testCases := []struct {
		keywords []string
		message  string
		expected bool
		keyword  string
	}{
		{
			keywords: []string{"spam", "scam"},
			message:  "This is spam",
			expected: true,
			keyword:  "spam",
		},
		{
			keywords: []string{"spam", "scam"},
			message:  "No matches here",
			expected: false,
			keyword:  "",
		},
		{
			keywords: []string{"test", "example"},
			message:  "This is a test message",
			expected: true,
			keyword:  "test",
		},
		{
			keywords: []string{"word"},
			message:  "password", // Should not match due to word boundaries
			expected: false,
			keyword:  "",
		},
	}

	for i, tc := range testCases {
		originalKeyword, originalFound := checkFiltersOriginal(tc.keywords, tc.message)
		optimizedKeyword, optimizedFound := checkFiltersOptimized(tc.keywords, tc.message)

		if originalFound != optimizedFound {
			b.Errorf("Test case %d: found mismatch. Original: %v, Optimized: %v", i, originalFound, optimizedFound)
		}

		if originalFound && originalKeyword != optimizedKeyword {
			b.Errorf("Test case %d: keyword mismatch. Original: %s, Optimized: %s", i, originalKeyword, optimizedKeyword)
		}

		if tc.expected != optimizedFound {
			b.Errorf("Test case %d: expected %v, got %v", i, tc.expected, optimizedFound)
		}

		if tc.expected && tc.keyword != optimizedKeyword {
			b.Errorf("Test case %d: expected keyword %s, got %s", i, tc.keyword, optimizedKeyword)
		}
	}
}

// Benchmark for user collection queries by user_id and username
func BenchmarkUserCollectionIndexes(b *testing.B) {
	// Setup: insert test users
	const numUsers = 10000
	users := make([]interface{}, numUsers)
	for i := 0; i < numUsers; i++ {
		users[i] = map[string]interface{}{
			"user_id":  int64(i),
			"username": fmt.Sprintf("user%d", i),
			"name":     fmt.Sprintf("Test User %d", i),
			"language": "en",
		}
	}
	userColl := getUserCollection()
	if _, err := userColl.DeleteMany(context.TODO(), map[string]interface{}{}); err != nil {
		b.Errorf("Failed to clean up before test: %v", err)
	}
	if _, err := userColl.InsertMany(context.TODO(), users); err != nil {
		b.Errorf("Failed to insert test users: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Query by user_id
		_ = userColl.FindOne(context.TODO(), map[string]interface{}{"user_id": int64(i % numUsers)})
		// Query by username
		_ = userColl.FindOne(context.TODO(), map[string]interface{}{"username": fmt.Sprintf("user%d", i%numUsers)})
	}

	b.StopTimer()
	if _, err := userColl.DeleteMany(context.TODO(), map[string]interface{}{}); err != nil {
		b.Errorf("Failed to clean up after test: %v", err)
	}
}

// Helper to get user collection
func getUserCollection() *mongo.Collection {
	return db.GetTestCollection() // Replace with db.userColl if accessible
}
