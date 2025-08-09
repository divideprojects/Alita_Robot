package i18n

import (
	"embed"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// Global translator manager instance
var globalTranslatorManager *TranslatorManager

// InitializeI18n initializes the new catalog-based i18n system.
// This loads the global message catalog and sets up the translator manager.
func InitializeI18n(locales *embed.FS, localesDir string) error {
	// Initialize global translator manager with default config
	config := DefaultConfig()
	globalTranslatorManager = NewTranslatorManager(config, localesDir)
	
	// Register all message defaults here
	// This replaces the need for individual message files
	registerAllMessages()
	
	log.Infof("I18n system initialized with %d messages in catalog", Count())
	return nil
}

// NewTranslator creates a new translator instance for the specified language.
// This is the main API for getting translators in the new system.
func NewTranslator(langCode string) (*Translator, error) {
	if globalTranslatorManager == nil {
		return nil, fmt.Errorf("i18n system not initialized - call InitializeI18n first")
	}
	return globalTranslatorManager.GetTranslator(langCode)
}

// MustNewTranslator creates a new translator and panics on error.
// Useful for initialization where errors should be fatal.
func MustNewTranslator(langCode string) *Translator {
	translator, err := NewTranslator(langCode)
	if err != nil {
		panic(fmt.Sprintf("Failed to create translator for %s: %v", langCode, err))
	}
	return translator
}

// GetAvailableLanguages returns all available language codes.
func GetAvailableLanguages() []string {
	if globalTranslatorManager == nil {
		return []string{"en"} // Default fallback
	}
	return globalTranslatorManager.LoadedLanguages()
}

// IsLanguageSupported checks if a language is supported.
func IsLanguageSupported(langCode string) bool {
	if globalTranslatorManager == nil {
		return langCode == "en" // Default fallback
	}
	_, err := globalTranslatorManager.GetTranslator(langCode)
	return err == nil
}

// GetMessage is a convenience function to get a translated message.
// This provides a simple API similar to the old system.
func GetMessage(langCode, key string, params Params) string {
	translator, err := NewTranslator(langCode)
	if err != nil {
		// Fallback to English on error
		translator, _ = NewTranslator("en")
	}
	if translator == nil {
		return fmt.Sprintf("{{%s}}", key)
	}
	return translator.Message(key, params)
}

// registerAllMessages registers all message defaults with the global catalog.
// This replaces the old separate message files with a centralized registration.
func registerAllMessages() {
	// Admin module messages
	MustRegister("admin_admin_list", "Admins in <b>%s</b>:", "chat_title")
	MustRegister("admin_promote_success", "Successfully promoted %s!", "user")
	MustRegister("admin_demote_success", "Successfully demoted %s!", "user")
	MustRegister("admin_promote_is_admin", "This person is already an admin, and how would am I supposed to promote them?")
	MustRegister("admin_promote_is_bot_itself", "If only I could do this to myself ;_;")
	MustRegister("admin_promote_is_owner", "This person created this chat, and how would am I supposed to promote them?")
	MustRegister("admin_demote_is_admin", "This person is not an admin, and how I am supposed to demote them?")
	MustRegister("admin_demote_is_bot_itself", "I am not going to demote myself.")
	MustRegister("admin_demote_is_owner", "This person created this chat, and how would I demote them?")
	MustRegister("admin_title_success_set", "Successfully set %s's admin title to <b>%s</b>", "user", "title")
	MustRegister("admin_title_is_admin", "This person is already an admin, how would I set a custom admin title for them?")
	MustRegister("admin_title_is_bot_itself", "If only I could do this to myself ;_;")
	MustRegister("admin_title_is_owner", "This person created this chat, and how would am I supposed to set a admin title to them?")
	MustRegister("admin_errors_err_cannot_promote", "Failed to promote; I might not be the admin, or they may be promoted by another admin.")
	MustRegister("admin_errors_err_cannot_demote", "Failed to demote; I might not be the admin, or they may be promoted by another admin.")
	MustRegister("admin_errors_err_set_title", "Failed to set custom admin title; The Title may not be correct or may contain emojis.")
	MustRegister("admin_errors_title_empty", "You need to give me an admin title to set it.")
	MustRegister("admin_promote_admin_title_truncated", "Admin title truncated to 16 characters from %d", "original_length")
	MustRegister("admin_anon_admin_enabled", "AnonAdmin mode is currently <b>enabled</b> for %s.\nThis allows all anonymous admin to perform admin actions without restriction.", "chat_title")
	MustRegister("admin_anon_admin_disabled", "AnonAdmin mode is currently <b>disabled</b> for %s.\nThis requires anonymous admins to press a button to confirm their permissions.", "chat_title")
	MustRegister("admin_anon_admin_enabled_now", "AnonAdmin mode is now <b>enabled</b> for %s.\nFrom now onwards, I will ask the admins to verify permissions from anonymous admins.", "chat_title")
	MustRegister("admin_anon_admin_disabled_now", "AnonAdmin mode is now <b>disabled</b> for %s.\nFrom now onwards, I won't ask the admins to verify for permissions anymore from anonymous admins.", "chat_title")
	MustRegister("admin_anon_admin_already_enabled", "AnonAdmin mode is already <b>enabled</b> for %s", "chat_title")
	MustRegister("admin_anon_admin_already_disabled", "AnonAdmin mode is already <b>disabled</b> for %s", "chat_title")
	MustRegister("admin_anon_admin_invalid_arg", "Invalid argument, I only understand <code>on</code>, <code>off</code>, <code>yes</code>, <code>no</code>")
	
	// Help module messages  
	MustRegister("admin_help_msg", "Make it easy to promote and demote users with the admin module!\n\n*User Commands:*\n× /adminlist: List the admins in the current chat.\n\n*Admin Commands:*\n× /promote `<reply/username/mention/userid>`: Promote a user.\n× /demote `<reply/username/mention/userid>`: Demote a user.\n× /title `<reply/username/mention/userid>` `<custom title>`: Set custom title for user")
	
	// Antiflood module messages
	MustRegister("antiflood_check_flood_perform_action", "Yeah, I don't like your flooding. %s has been %s!", "user", "action")
	MustRegister("antiflood_errors_expected_args", "I expected some arguments! Either off, or an integer. eg: `/setflood 5`, or `/setflood off`")
	MustRegister("antiflood_errors_invalid_int", "That's not a valid integer. Please give me a valid integer, or `off`.")
	MustRegister("antiflood_errors_set_in_limit", "The antiflood limit has to be set between 3 and 100.")
	MustRegister("antiflood_setflood_disabled", "Antiflood has been disabled.")
	MustRegister("antiflood_setflood_success", "Successfully set antiflood limit to %d messages.", "limit")
	MustRegister("antiflood_flood_disabled", "Antiflood is currently disabled in %s.", "chat_title")
	MustRegister("antiflood_flood_show_settings", "Current antiflood settings for %s:\n• Limit: %d messages\n• Action: %s\n• Delete flood messages: %s", "chat_title", "limit", "action", "delete_status")
	MustRegister("antiflood_setfloodmode_success", "Successfully updated antiflood action to %s.", "action")
	MustRegister("antiflood_setfloodmode_unknown_type", "Unknown action type: %s", "action")
	MustRegister("antiflood_setfloodmode_specify_action", "Please specify an action: ban, kick, mute, or warn.")
	MustRegister("antiflood_flood_deleter_enabled", "Flood message deletion is currently enabled.")
	MustRegister("antiflood_flood_deleter_disabled", "Flood message deletion is currently disabled.")
	MustRegister("antiflood_flood_deleter_invalid_option", "Invalid option. Use 'on' or 'off'.")
	MustRegister("antiflood_flood_deleter_already_enabled", "Flood message deletion is already enabled.")
	MustRegister("antiflood_flood_deleter_already_disabled", "Flood message deletion is already disabled.")
	
	// Ban module messages
	MustRegister("bans_ban_success", "Successfully banned %s!", "user")
	MustRegister("bans_unban_success", "Successfully unbanned %s!", "user")
	MustRegister("bans_kick_success", "Successfully kicked %s!", "user")
	MustRegister("bans_mute_success", "Successfully muted %s!", "user")
	MustRegister("bans_unmute_success", "Successfully unmuted %s!", "user")
	MustRegister("bans_user_not_found", "User not found.")
	MustRegister("bans_cannot_ban_admin", "I can't ban an admin.")
	MustRegister("bans_cannot_ban_self", "I can't ban myself.")
	MustRegister("bans_cannot_ban_owner", "I can't ban the chat owner.")
	
	// Connection module messages
	MustRegister("connection_connected", "Connected to %s!", "chat_title")
	MustRegister("connection_disconnected", "Disconnected from %s!", "chat_title")
	MustRegister("connection_not_connected", "You are not connected to any chat.")
	MustRegister("connection_already_connected", "You are already connected to %s!", "chat_title")
	
	// Blacklist module messages
	MustRegister("blacklists_added", "Added blacklist trigger: %s", "trigger")
	MustRegister("blacklists_removed", "Removed blacklist trigger: %s", "trigger")
	MustRegister("blacklists_not_found", "Blacklist trigger not found: %s", "trigger")
	MustRegister("blacklists_action_performed", "Blacklisted word detected! %s has been %s!", "user", "action")
	
	// Formatting module messages
	MustRegister("formatting_help", "Here's how you can format your messages:")
	MustRegister("formatting_bold", "**bold text**")
	MustRegister("formatting_italic", "*italic text*")
	MustRegister("formatting_code", "`code text`")
	MustRegister("formatting_pre", "```pre-formatted text```")
	
	// Generic error messages
	MustRegister("error_user_not_found", "User not found.")
	MustRegister("error_admin_required", "Admin privileges required.")
	MustRegister("error_bot_admin_required", "I need to be an admin to perform this action.")
	MustRegister("error_group_only", "This command can only be used in groups.")
	MustRegister("error_private_only", "This command can only be used in private chat.")
	MustRegister("error_reply_required", "Please reply to a user or provide username/mention/user ID.")
	MustRegister("error_command_disabled", "This command has been disabled in this chat.")
	MustRegister("error_anonymous_user_not_supported", "This command cannot be used on anonymous user.")
	MustRegister("error_specify_user", "I don't know who you're talking about, you're going to need to specify a user...!")
	
	// Generic success messages
	MustRegister("success_operation_completed", "Operation completed successfully.")
	MustRegister("success_settings_saved", "Settings saved successfully.")
	
	// Misc messages
	MustRegister("misc_loading", "Loading...")
	MustRegister("misc_processing", "Processing...")
	MustRegister("misc_cancelled", "Operation cancelled.")
	MustRegister("misc_timeout", "Operation timed out.")
	
	// Cache messages
	MustRegister("cache_reloaded", "Admin cache has been reloaded!")
	
	// Connection messages
	MustRegister("connections_is_user_connected_need_group", "This command can only be used in groups.")
	MustRegister("connections_is_user_connected_bot_not_admin", "I need to be an admin to check connection status.")
	MustRegister("connections_is_user_connected_user_not_admin", "You need to be an admin to use this command.")
	
	// Main language info
	MustRegister("main_language_name", "English")
	MustRegister("main_language_flag", "🇺🇸")
	
	log.Infof("Registered %d default messages in i18n catalog", Count())
}
