// Package catalog provides example usage of the new i18n catalog system.
// This file demonstrates how to use the catalog system and can be removed in production.
package catalog

import (
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the new catalog system.
func ExampleUsage() {
	// 1. Register messages with embedded English defaults
	registerExampleMessages()
	
	// 2. Initialize global translator manager
	config := DefaultConfig()
	InitGlobalManager(config, "/Users/divkix/GitHub/Alita_Robot/locales")
	
	// 3. Get translators for different languages
	enTranslator, err := T("en")
	if err != nil {
		log.Printf("Failed to get English translator: %v", err)
		return
	}
	
	// Try to get Spanish translator (will use English defaults if no translation)
	esTranslator, err := T("es")
	if err != nil {
		log.Printf("Failed to get Spanish translator: %v", err)
		return
	}
	
	// 4. Use messages with parameters
	params := Params{
		"user": "John Doe",
		"chat": "My Group",
	}
	
	// English messages
	fmt.Println("English:")
	fmt.Println("  ", enTranslator.Message("admin.promote_success", params))
	fmt.Println("  ", enTranslator.Message("admin.demote_success", params))
	fmt.Println("  ", enTranslator.Message("greetings.welcome", params))
	
	// Spanish messages (will fall back to English if no translation)
	fmt.Println("\nSpanish:")
	fmt.Println("  ", esTranslator.Message("admin.promote_success", params))
	fmt.Println("  ", esTranslator.Message("admin.demote_success", params))
	fmt.Println("  ", esTranslator.Message("greetings.welcome", params))
	
	// 5. Plural messages
	for _, count := range []int{0, 1, 2, 5} {
		pluralParams := Params{"count": count}
		msg := enTranslator.Plural("user.count", count, pluralParams)
		fmt.Printf("Users: %s\n", msg)
	}
	
	// 6. Show statistics
	showStatistics()
}

func registerExampleMessages() {
	// Register admin messages
	MustRegister("admin.promote_success", "Successfully promoted {user}!", "user")
	MustRegister("admin.demote_success", "Successfully demoted {user}!", "user")
	MustRegister("admin.cannot_promote_self", "I cannot promote myself!")
	MustRegister("admin.cannot_demote_owner", "I cannot demote the chat owner!")
	MustRegister("admin.user_already_admin", "{user} is already an admin!")
	
	// Register greeting messages
	MustRegister("greetings.welcome", "Welcome {user} to {chat}!", "user", "chat")
	MustRegister("greetings.goodbye", "Goodbye {user}, hope to see you again!", "user")
	
	// Register error messages
	MustRegister("errors.user_not_found", "User not found!")
	MustRegister("errors.bot_not_admin", "I need admin rights to do that!")
	MustRegister("errors.insufficient_permissions", "You don't have permission to do that!")
	
	// Register plural message (note: plurals need to be handled differently)
	// For now, we can register the individual forms
	MustRegister("user.count.zero", "No users")
	MustRegister("user.count.one", "One user")
	MustRegister("user.count.other", "{count} users", "count")
	
	// Register filter messages
	MustRegister("filters.added", "Filter '{name}' added successfully!", "name")
	MustRegister("filters.removed", "Filter '{name}' removed successfully!", "name")
	MustRegister("filters.not_found", "Filter '{name}' not found!", "name")
	MustRegister("filters.list_empty", "No filters are set for this chat.")
	
	// Register captcha messages
	MustRegister("captcha.welcome", "Welcome {user}! Please solve this captcha to verify you're human:", "user")
	MustRegister("captcha.solved", "Captcha solved! Welcome to the chat {user}!", "user")
	MustRegister("captcha.failed", "Captcha failed! Please try again.")
	MustRegister("captcha.timeout", "Captcha timeout! You have been removed from the chat.")
}

func showStatistics() {
	fmt.Println("\nCatalog Statistics:")
	stats := GetStats()
	
	fmt.Printf("Total messages: %d\n", stats.TotalMessages)
	fmt.Printf("Messages with parameters: %d\n", stats.MessagesWithParams)
	fmt.Printf("Average parameters per message: %.1f\n", stats.AverageParamCount)
	
	fmt.Println("\nMessages by prefix:")
	for _, prefix := range stats.TopPrefixes {
		count := stats.MessagesByPrefix[prefix]
		fmt.Printf("  %s: %d messages\n", prefix, count)
	}
	
	fmt.Println("\nAll registered keys:")
	keys := Keys()
	for i, key := range keys {
		if i >= 10 {
			fmt.Printf("  ... and %d more\n", len(keys)-10)
			break
		}
		fmt.Printf("  %s\n", key)
	}
}

// ExampleMigration shows how to migrate from the old nested YAML structure.
func ExampleMigration() {
	fmt.Println("Migration Example:")
	fmt.Println("Old nested structure:")
	fmt.Println("  strings.Admin.promote.success: 'Successfully promoted [user]!'")
	fmt.Println("New flat structure:")
	fmt.Println("  admin.promote_success: 'Successfully promoted {user}!'")
	fmt.Println("")
	
	// Show how to handle the migration
	oldKey := "strings.Admin.promote.success"
	newKey := "admin.promote_success"
	
	fmt.Printf("Old key: %s -> New key: %s\n", oldKey, newKey)
	fmt.Println("Old params: [user] style -> New params: {param} style")
}

// ExampleYAMLStructure shows the recommended YAML structure for translations.
func ExampleYAMLStructure() {
	yamlExample := `# locales/en.yaml - New flat structure
admin.promote_success: "Successfully promoted {user}!"
admin.demote_success: "Successfully demoted {user}!"
admin.cannot_promote_self: "I cannot promote myself!"
admin.user_already_admin: "{user} is already an admin!"

greetings.welcome: "Welcome {user} to {chat}!"
greetings.goodbye: "Goodbye {user}, hope to see you again!"

errors.user_not_found: "User not found!"
errors.bot_not_admin: "I need admin rights to do that!"

# Plural forms
user.count.zero: "No users"
user.count.one: "One user"  
user.count.other: "{count} users"

filters.added: "Filter '{name}' added successfully!"
filters.removed: "Filter '{name}' removed successfully!"
`
	
	fmt.Println("Example YAML structure:")
	fmt.Println(yamlExample)
}