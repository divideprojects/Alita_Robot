package modules

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
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

// Benchmark regex compilation overhead
func BenchmarkRegexCompilation_Individual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, keyword := range testFilterKeywords {
			pattern := fmt.Sprintf(`(\b|\s)%s\b`, regexp.QuoteMeta(keyword))
			if _, err := regexp.Compile(pattern); err != nil {
				b.Errorf("Failed to compile regex pattern %s: %v", pattern, err)
			}
		}
	}
}

func BenchmarkRegexCompilation_Combined(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buildFilterRegex(testFilterKeywords)
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

// Benchmark with caching (simulating real usage)
func BenchmarkWithCaching_Optimized(b *testing.B) {
	chatId := int64(12345)

	// Pre-build and cache regex
	regex := getOrBuildFilterRegex(chatId, testFilterKeywords)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, message := range testMessages {
			findMatchingKeyword(regex, message, testFilterKeywords)
		}
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocation_Original(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		checkFiltersOriginal(testFilterKeywords, "This message contains spam")
	}
}

func BenchmarkMemoryAllocation_Optimized(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		checkFiltersOptimized(testFilterKeywords, "This message contains spam")
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

// Best case scenario: early match
func BenchmarkBestCase_Original(b *testing.B) {
	keywords := generateLargeFilterKeywords(200)
	message := "keyword0 appears early in this message"

	for i := 0; i < b.N; i++ {
		checkFiltersOriginal(keywords, message)
	}
}

func BenchmarkBestCase_Optimized(b *testing.B) {
	keywords := generateLargeFilterKeywords(200)
	message := "keyword0 appears early in this message"

	for i := 0; i < b.N; i++ {
		checkFiltersOptimized(keywords, message)
	}
}

// Test correctness of optimized implementation
func TestFilterOptimization_Correctness(t *testing.T) {
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
			t.Errorf("Test case %d: found mismatch. Original: %v, Optimized: %v", i, originalFound, optimizedFound)
		}

		if originalFound && originalKeyword != optimizedKeyword {
			t.Errorf("Test case %d: keyword mismatch. Original: %s, Optimized: %s", i, originalKeyword, optimizedKeyword)
		}

		if tc.expected != optimizedFound {
			t.Errorf("Test case %d: expected %v, got %v", i, tc.expected, optimizedFound)
		}

		if tc.expected && tc.keyword != optimizedKeyword {
			t.Errorf("Test case %d: expected keyword %s, got %s", i, tc.keyword, optimizedKeyword)
		}
	}
}
