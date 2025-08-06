//go:build ignore
// +build ignore

// This file demonstrates how to use the new i18n system
// It's marked with build ignore so it won't be included in builds
package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
)

//go:embed locales
var localeFS embed.FS

// main demonstrates how to use the i18n system with various features including
// basic usage, backward compatibility, parameter interpolation, pluralization and error handling.
func main() {
	// Initialize cache (optional but recommended for production)
	if err := cache.InitCache(); err != nil {
		log.Printf("Warning: Cache initialization failed: %v", err)
	}

	// Get the locale manager singleton
	manager := i18n.GetManager()

	// Initialize with custom configuration
	config := i18n.DefaultManagerConfig()
	config.Cache.EnableCache = true
	config.Loader.DefaultLanguage = "en"
	config.Loader.StrictMode = false // Don't fail if some locales have errors

	if err := manager.Initialize(&localeFS, "locales", config); err != nil {
		log.Fatalf("Failed to initialize locale manager: %v", err)
	}

	// Example 1: Basic usage with new system
	demoNewSystem()

	// Example 2: Backward compatibility
	demoBackwardCompatibility()

	// Example 3: Parameter interpolation
	demoParameterInterpolation()

	// Example 4: Pluralization
	demoPluralization()

	// Example 5: Error handling
	demoErrorHandling()
}

// demoNewSystem demonstrates basic usage of the new i18n system with translator instances.
func demoNewSystem() {
	fmt.Println("=== New System Demo ===")

	// Get translator for English
	translator, err := i18n.NewTranslator("en")
	if err != nil {
		log.Printf("Error creating translator: %v", err)
		return
	}

	// Get a simple string
	greeting, err := translator.GetString("strings.CommonStrings.greetings.welcome")
	if err != nil {
		log.Printf("Error getting string: %v", err)
	} else {
		fmt.Printf("Greeting: %s\n", greeting)
	}

	// Get available languages
	languages := i18n.GetAvailableLanguages()
	fmt.Printf("Available languages: %v\n", languages)
}

// demoBackwardCompatibility shows how the old I18n system still works alongside the new system.
func demoBackwardCompatibility() {
	fmt.Println("\n=== Backward Compatibility Demo ===")

	// Old way still works
	oldi18n := i18n.I18n{LangCode: "en"}
	text := oldi18n.GetString("strings.CommonStrings.greetings.welcome")
	fmt.Printf("Old system result: %s\n", text)

	// Initialize old system (this also initializes new system)
	i18n.LoadLocaleFiles(&localeFS, "locales")
	text2 := oldi18n.GetString("strings.Admin.adminlist")
	fmt.Printf("Old system with fallback: %s\n", text2)
}

// demoParameterInterpolation demonstrates how to use parameter interpolation with translation keys.
func demoParameterInterpolation() {
	fmt.Println("\n=== Parameter Interpolation Demo ===")

	translator, _ := i18n.NewTranslator("en")

	// Using {key} style parameters
	params := i18n.TranslationParams{
		"name":  "Alice",
		"count": 5,
	}

	// This would work if the translation had {name} and {count} placeholders
	result, err := translator.GetString("strings.example.welcome_user", params)
	if err != nil {
		// Fallback to simple message
		result = "Welcome to the bot!"
	}
	fmt.Printf("Parameterized message: %s\n", result)
}

// demoPluralization shows how to handle plural forms of translations based on count values.
func demoPluralization() {
	fmt.Println("\n=== Pluralization Demo ===")

	translator, _ := i18n.NewTranslator("en")

	// Demo different plural forms
	counts := []int{0, 1, 2, 5}
	for _, count := range counts {
		params := i18n.TranslationParams{"count": count}
		message, err := translator.GetPlural("strings.example.items_count", count, params)
		if err != nil {
			// Fallback
			message = fmt.Sprintf("%d items", count)
		}
		fmt.Printf("Count %d: %s\n", count, message)
	}
}

// demoErrorHandling demonstrates error handling scenarios including invalid languages and missing keys.
func demoErrorHandling() {
	fmt.Println("\n=== Error Handling Demo ===")

	// Try unsupported language
	translator, err := i18n.NewTranslator("invalid")
	if err != nil {
		fmt.Printf("Expected error for invalid language: %v\n", err)

		// Check error type
		if i18n.IsI18nError(err) {
			fmt.Println("Error is of type I18nError")
		}
	}

	// Try missing key
	translator, _ = i18n.NewTranslator("en")
	_, err = translator.GetString("nonexistent.key")
	if err != nil {
		fmt.Printf("Expected error for missing key: %v\n", err)

		if i18n.IsKeyNotFound(err) {
			fmt.Println("Error indicates key was not found")
		}
	}

	// Get manager stats
	manager := i18n.GetManager()
	stats := manager.GetStats()
	fmt.Printf("Manager stats: %+v\n", stats)
}
