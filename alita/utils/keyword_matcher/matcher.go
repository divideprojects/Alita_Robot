package keyword_matcher

import (
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/ahocorasick"
	log "github.com/sirupsen/logrus"
)

// KeywordMatcher provides efficient multi-pattern matching using Aho-Corasick algorithm
type KeywordMatcher struct {
	matcher  *ahocorasick.Matcher
	patterns []string
	mu       sync.RWMutex
	lastBuild time.Time
}

// MatchResult contains information about a matched pattern
type MatchResult struct {
	Pattern string // The original pattern that matched
	Start   int    // Start position of match in text
	End     int    // End position of match in text
}

// NewKeywordMatcher creates a new keyword matcher with the given patterns
func NewKeywordMatcher(patterns []string) *KeywordMatcher {
	km := &KeywordMatcher{
		patterns: make([]string, len(patterns)),
	}
	copy(km.patterns, patterns)
	km.build()
	return km
}

// build compiles the patterns into an Aho-Corasick matcher
func (km *KeywordMatcher) build() {
	if len(km.patterns) == 0 {
		km.matcher = nil
		return
	}

	// Convert patterns to lowercase for case-insensitive matching
	lowerPatterns := make([]string, len(km.patterns))
	for i, pattern := range km.patterns {
		lowerPatterns[i] = strings.ToLower(pattern)
	}

	km.matcher = ahocorasick.NewStringMatcher(lowerPatterns)
	km.lastBuild = time.Now()

	log.WithFields(log.Fields{
		"patterns_count": len(km.patterns),
		"build_time": time.Since(km.lastBuild),
	}).Debug("Built Aho-Corasick matcher")
}

// UpdatePatterns updates the matcher with new patterns
func (km *KeywordMatcher) UpdatePatterns(patterns []string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.patterns = make([]string, len(patterns))
	copy(km.patterns, patterns)
	km.build()
}

// FindMatches returns all matches in the given text
func (km *KeywordMatcher) FindMatches(text string) []MatchResult {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.matcher == nil {
		return nil
	}

	lowerText := strings.ToLower(text)
	hits := km.matcher.Match([]byte(lowerText))

	if len(hits) == 0 {
		return nil
	}

	results := make([]MatchResult, 0, len(hits))
	for _, patternIndex := range hits {
		if patternIndex < len(km.patterns) {
			pattern := km.patterns[patternIndex]
			// Find the actual position in the text
			pos := strings.Index(lowerText, strings.ToLower(pattern))
			if pos != -1 {
				results = append(results, MatchResult{
					Pattern: pattern,
					Start:   pos,
					End:     pos + len(pattern),
				})
			}
		}
	}

	return results
}

// HasMatch returns true if any pattern matches the text
func (km *KeywordMatcher) HasMatch(text string) bool {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.matcher == nil {
		return false
	}

	lowerText := strings.ToLower(text)
	hits := km.matcher.Match([]byte(lowerText))
	return len(hits) > 0
}

// FindFirstMatch returns the first match found in the text, or nil if no match
func (km *KeywordMatcher) FindFirstMatch(text string) *MatchResult {
	matches := km.FindMatches(text)
	if len(matches) == 0 {
		return nil
	}
	return &matches[0]
}

// GetPatterns returns a copy of the current patterns
func (km *KeywordMatcher) GetPatterns() []string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	patterns := make([]string, len(km.patterns))
	copy(patterns, km.patterns)
	return patterns
}

// IsEmpty returns true if no patterns are configured
func (km *KeywordMatcher) IsEmpty() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return len(km.patterns) == 0
}
