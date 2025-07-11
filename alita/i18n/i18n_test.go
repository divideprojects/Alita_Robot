package i18n

import (
	"embed"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
)

//go:embed testdata
var testFS embed.FS

func resetGlobals() {
	localeMu.Lock()
	localeMap = make(map[string]*viper.Viper)
	localeMu.Unlock()

	fallbackMu.Lock()
	fallbackChains = map[string][]string{
		"pt_BR": {"pt", DefaultLangCode},
		"es_MX": {"es", DefaultLangCode},
		"en_US": {DefaultLangCode},
		"zh_CN": {"zh", DefaultLangCode},
		"zh_TW": {"zh", DefaultLangCode},
	}
	fallbackMu.Unlock()
}

func TestLoadLocaleFiles(t *testing.T) {
	resetGlobals()

	// Test successful loading
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that locales were loaded
	languages := GetAvailableLanguages()
	if len(languages) == 0 {
		t.Fatal("No languages loaded")
	}

	// Test that we can retrieve strings
	tr := New("en")
	text := tr.GetString("test.key")
	if text == "" || strings.Contains(text, "@@") {
		t.Errorf("Expected valid text, got: %s", text)
	}
}

func TestLoadLocaleFilesErrors(t *testing.T) {
	resetGlobals()

	// Test non-existent directory
	err := LoadLocaleFiles(&testFS, "testdata/nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Test directory with invalid files
	err = LoadLocaleFiles(&testFS, "testdata/invalid")
	if err == nil {
		t.Error("Expected error for invalid files")
	}

	// Check that it's a LoadErrors type
	if loadErrs, ok := err.(LoadErrors); ok {
		if len(loadErrs) == 0 {
			t.Error("Expected non-empty LoadErrors")
		}
	}
}

func TestMustLoadLocaleFiles(t *testing.T) {
	resetGlobals()

	// Test successful load (should not panic)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustLoadLocaleFiles panicked on valid data: %v", r)
			}
		}()
		MustLoadLocaleFiles(&testFS, "testdata/valid")
	}()

	// Test failed load (should panic)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustLoadLocaleFiles should have panicked on invalid data")
			}
		}()
		MustLoadLocaleFiles(&testFS, "testdata/nonexistent")
	}()
}

func TestNewI18n(t *testing.T) {
	// Test with valid language code
	tr := New("en")
	if tr.LangCode != "en" {
		t.Errorf("Expected LangCode 'en', got: %s", tr.LangCode)
	}

	// Test with empty language code (should default)
	tr = New("")
	if tr.LangCode != DefaultLangCode {
		t.Errorf("Expected LangCode '%s', got: %s", DefaultLangCode, tr.LangCode)
	}
}

func TestIsLanguageAvailable(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Test available language
	if !IsLanguageAvailable("en") {
		t.Error("Expected 'en' to be available")
	}

	// Test unavailable language
	if IsLanguageAvailable("nonexistent") {
		t.Error("Expected 'nonexistent' to be unavailable")
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	resetGlobals()

	// Test with no loaded languages
	languages := GetAvailableLanguages()
	if len(languages) != 0 {
		t.Errorf("Expected 0 languages, got: %d", len(languages))
	}

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Test with loaded languages
	languages = GetAvailableLanguages()
	if len(languages) == 0 {
		t.Error("Expected some languages to be available")
	}
}

func TestFallbackChains(t *testing.T) {
	resetGlobals()

	// Test default fallback
	chain := GetFallbackChain("unknown")
	expectedDefault := []string{DefaultLangCode}
	if len(chain) != len(expectedDefault) || chain[0] != expectedDefault[0] {
		t.Errorf("Expected default fallback %v, got: %v", expectedDefault, chain)
	}

	// Test configured fallback
	chain = GetFallbackChain("pt_BR")
	expected := []string{"pt", DefaultLangCode}
	if len(chain) != len(expected) {
		t.Errorf("Expected fallback length %d, got: %d", len(expected), len(chain))
	}
	for i, lang := range expected {
		if i >= len(chain) || chain[i] != lang {
			t.Errorf("Expected fallback %v, got: %v", expected, chain)
			break
		}
	}

	// Test setting custom fallback
	SetFallbackChain("custom", []string{"fallback1", "fallback2"})
	chain = GetFallbackChain("custom")
	expected = []string{"fallback1", "fallback2"}
	if len(chain) != len(expected) {
		t.Errorf("Expected custom fallback length %d, got: %d", len(expected), len(chain))
	}
	for i, lang := range expected {
		if i >= len(chain) || chain[i] != lang {
			t.Errorf("Expected custom fallback %v, got: %v", expected, chain)
			break
		}
	}
}

func TestGetString(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tr := New("en")

	// Test existing key
	text := tr.GetString("test.key")
	if strings.Contains(text, "@@") {
		t.Errorf("Expected valid text, got missing key marker: %s", text)
	}

	// Test missing key (should return marked missing key)
	text = tr.GetString("nonexistent.key")
	expectedMarker := fmt.Sprintf(MissingKeyMarker, "nonexistent.key")
	if text != expectedMarker {
		t.Errorf("Expected missing key marker '%s', got: %s", expectedMarker, text)
	}

	// Test empty key
	text = tr.GetString("")
	expectedMarker = fmt.Sprintf(MissingKeyMarker, "empty-key")
	if text != expectedMarker {
		t.Errorf("Expected empty key marker '%s', got: %s", expectedMarker, text)
	}
}

func TestGetStringSlice(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tr := New("en")

	// Test existing slice key
	slice := tr.GetStringSlice("test.list")
	if len(slice) == 0 {
		t.Error("Expected non-empty slice")
	}

	// Test missing key (should return nil)
	slice = tr.GetStringSlice("nonexistent.list")
	if slice != nil {
		t.Errorf("Expected nil for missing key, got: %v", slice)
	}

	// Test empty key
	slice = tr.GetStringSlice("")
	if slice != nil {
		t.Errorf("Expected nil for empty key, got: %v", slice)
	}
}

func TestGetStringWithError(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tr := New("en")

	// Test existing key
	text, err := tr.GetStringWithError("test.key")
	if err != nil {
		t.Errorf("Expected no error for existing key, got: %v", err)
	}
	if text == "" {
		t.Error("Expected non-empty text")
	}

	// Test missing key
	text, err = tr.GetStringWithError("nonexistent.key")
	if err == nil {
		t.Error("Expected error for missing key")
	}
	if text != "" {
		t.Errorf("Expected empty text for missing key, got: %s", text)
	}

	// Test empty key
	_, err = tr.GetStringWithError("")
	if err != ErrEmptyKey {
		t.Errorf("Expected ErrEmptyKey, got: %v", err)
	}
}

func TestHasKey(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tr := New("en")

	// Test existing key
	if !tr.HasKey("test.key") {
		t.Error("Expected HasKey to return true for existing key")
	}

	// Test missing key
	if tr.HasKey("nonexistent.key") {
		t.Error("Expected HasKey to return false for missing key")
	}

	// Test empty key
	if tr.HasKey("") {
		t.Error("Expected HasKey to return false for empty key")
	}
}

func TestStringsPrefixFallback(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tr := New("en")

	// Test key without strings prefix (should try with prefix)
	text1 := tr.GetString("test.key")
	text2 := tr.GetString("strings.test.key")

	// Both should return the same result
	if text1 != text2 {
		t.Errorf("Expected same result for prefixed and non-prefixed keys, got: '%s' vs '%s'", text1, text2)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Test GetString convenience function
	text1 := GetString("en", "test.key")
	text2 := New("en").GetString("test.key")
	if text1 != text2 {
		t.Errorf("Convenience GetString should match instance method, got: '%s' vs '%s'", text1, text2)
	}

	// Test GetStringSlice convenience function
	slice1 := GetStringSlice("en", "test.list")
	slice2 := New("en").GetStringSlice("test.list")
	if len(slice1) != len(slice2) {
		t.Errorf("Convenience GetStringSlice should match instance method")
	}

	// Test HasKey convenience function
	has1 := HasKey("en", "test.key")
	has2 := New("en").HasKey("test.key")
	if has1 != has2 {
		t.Errorf("Convenience HasKey should match instance method")
	}
}

func TestThreadSafety(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Test concurrent reads
	const numGoroutines = 100
	const numIterations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			tr := New("en")
			for j := 0; j < numIterations; j++ {
				// Test various operations
				text := tr.GetString("test.key")
				if strings.Contains(text, "@@") {
					errors <- fmt.Errorf("goroutine %d: got missing key marker", id)
					return
				}

				slice := tr.GetStringSlice("test.list")
				if len(slice) == 0 {
					errors <- fmt.Errorf("goroutine %d: got empty slice", id)
					return
				}

				has := tr.HasKey("test.key")
				if !has {
					errors <- fmt.Errorf("goroutine %d: HasKey returned false", id)
					return
				}

				langs := GetAvailableLanguages()
				if len(langs) == 0 {
					errors <- fmt.Errorf("goroutine %d: no available languages", id)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	// Check for errors or timeout
	select {
	case err := <-errors:
		t.Errorf("Thread safety test failed: %v", err)
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Error("Thread safety test timed out")
	}
}

func TestConcurrentLoadAndRead(t *testing.T) {
	resetGlobals()

	// Test concurrent loading and reading
	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// Goroutine 1: Load locales
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := LoadLocaleFiles(&testFS, "testdata/valid")
		if err != nil {
			errors <- fmt.Errorf("load error: %v", err)
		}
	}()

	// Goroutine 2: Try to read (might get empty results initially)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Small delay to increase chance of race
		tr := New("en")
		text := tr.GetString("test.key")
		// This might be empty or valid, both are acceptable during concurrent load
		_ = text
	}()

	// Wait for completion
	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errors:
		t.Errorf("Concurrent load/read test failed: %v", err)
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("Concurrent load/read test timed out")
	}
}

func TestFallbackLogic(t *testing.T) {
	resetGlobals()

	// Load test data
	err := LoadLocaleFiles(&testFS, "testdata/valid")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Test with a language that should fall back
	SetFallbackChain("test_lang", []string{"en"})
	tr := New("test_lang")

	// This should fall back to English
	text := tr.GetString("test.key")
	if strings.Contains(text, "@@") {
		t.Errorf("Expected fallback to work, got missing key marker: %s", text)
	}
}

// Benchmark tests to verify performance improvements

func BenchmarkGetString(b *testing.B) {
	resetGlobals()
	LoadLocaleFiles(&testFS, "testdata/valid")

	tr := New("en")
	key := "test.key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tr.GetString(key)
		}
	})
}

func BenchmarkGetStringSlice(b *testing.B) {
	resetGlobals()
	LoadLocaleFiles(&testFS, "testdata/valid")

	tr := New("en")
	key := "test.list"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tr.GetStringSlice(key)
		}
	})
}

func BenchmarkHasKey(b *testing.B) {
	resetGlobals()
	LoadLocaleFiles(&testFS, "testdata/valid")

	tr := New("en")
	key := "test.key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tr.HasKey(key)
		}
	})
}

func BenchmarkConcurrentGetString(b *testing.B) {
	resetGlobals()
	LoadLocaleFiles(&testFS, "testdata/valid")

	tr := New("en")
	key := "test.key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tr.GetString(key)
		}
	})
}
