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
	MustRegister("formatting_help_msg", "Alita supports a large number of formatting options to make your messages more expressive. Take a look by clicking the buttons below!")
	MustRegister("formatting_press_button_markdown_help", "Press the button below to get Markdown Help!")
	MustRegister("formatting_markdown", "<b>Markdown Formatting</b>\n\nYou can format your message using <b>bold</b>, <i>italics</i>, <u>underline</u>, and much more. Go ahead and experiment!\n<b>Supported markdown</b>:\n- <code>`code words`</code>: Backticks are used for monospace fonts. Shows as: <code>code words</code>.\n- <code>_italic words_</code>: Underscores are used for italic fonts. Shows as: <i>italic words</i>.\n- <code>*bold words*</code>: Asterisks are used for bold fonts. Shows as: <b>bold words</b>.\n- <code>~strikethrough~</code>: Tildes are used for strikethrough. Shows as: <strike>strikethrough</strike>.\n- <code>||spoiler||</code>: Double vertical bars are used for spoilers. Shows as: <tg-spoiler>Spoiler</tg-spoiler>.\n- <code>```pre```</code>: To make the formatter ignore other formatting characters inside the text formatted with '```', will be like: <code>**bold** | *bold*</code>.\n- <code>__underline__</code>: Double underscores are used for underlines. Shows as: underline. NOTE: Some clients try to be smart and interpret it as italic. In that case, try to use your app's built-in formatting.\n- <code>[hyperlink](example.com)</code>: This is the formatting used for hyperlinks. Shows as: <a href='https://example.com/'>hyperlink</a>.\n- <code>[My Button](buttonurl://example.com)</code>: This is the formatting used for creating buttons. This example will create a button named \"My button\" which opens <code>example.com</code> when clicked.\n\nIf you would like to send buttons on the same row, use the <code>:same</code> formatting.\n<b>Example:</b>\n<code>[button 1](buttonurl:example.com)</code>\n<code>[button 2](buttonurl://example.com:same)</code>\n<code>[button 3](buttonurl://example.com)</code>\nThis will show button 1 and 2 on the same line, with 3 underneath.")
	MustRegister("formatting_fillings", "<b>Fillings</b>\n\nYou can also customise the contents of your message with contextual data. For example, you could mention a user by name in the welcome message, or mention them in a filter!\nYou can use these to mention a user in notes too!\n\n<b>Supported fillings:</b>\n- <code>{first}</code>: The user's first name.\n- <code>{last}</code>: The user's last name.\n- <code>{fullname}</code>: The user's full name.\n- <code>{username}</code>: The user's username. If they don't have one, mentions the user instead.\n- <code>{mention}</code>: Mentions the user with their firstname.\n- <code>{id}</code>: The user's ID.\n- <code>{chatname}</code>: The chat's name.\n- <code>{rules}</code>: Adds Rules Button to Message.\n- <code>{protect}</code>: Protects the content from being shared.\n- <code>{preview}</code>: Enables previews in the messages.\n- <code>{nonotif}</code>: Disables the notification for that message.")
	MustRegister("formatting_random", "<b>Random Content</b>\n\nAnother thing that can be fun, is to randomise the contents of a message. Make things a little more personal by changing welcome messages, or changing notes!\nHow to use random contents:\n- %%%: This separator can be used to add  random replies to the bot.\nFor example:\n<code>hello\n%%%\nhow are you</code>\nThis will randomly choose between sending the first message, \"hello\", or the second message, \"how are you\".\nUse this to make Alita feel a bit more customised! (only works in filters/notes)\nExample welcome message:\n- Every time a new user joins, they'll be presented with one of the three messages shown here.\n-> /filter \"hey\"\nhello there <code>{first}</code>!\n%%%\nOoooh, <code>{first}</code> how are you?\n%%%\nSup? <code>{first}</code>")
	
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

	// Misc module messages
	MustRegister("misc_info_reply_or_mention", "Reply to a user or mention them to get their info.")
	MustRegister("misc_current_chat_info", "Current chat ID: <code>%d</code>", "chat_id")
	MustRegister("misc_ping_text", "Pong!")
	MustRegister("misc_translate_specify_lang", "Please specify a language code and text to translate.")
	MustRegister("misc_stats_calculating", "Calculating stats...")
	MustRegister("misc_keyboard_remove_success", "Bot keyboard has been removed.")

	// Developer module messages
	MustRegister("devs_eval_success", "✅ <b>Eval Result:</b>\n<pre>%s</pre>", "result")
	MustRegister("devs_eval_error", "❌ <b>Eval Error:</b>\n<pre>%s</pre>", "error")
	MustRegister("devs_shell_success", "✅ <b>Shell Command Result:</b>\n<pre>%s</pre>", "result")
	MustRegister("devs_shell_error", "❌ <b>Shell Command Error:</b>\n<pre>%s</pre>", "error")
	MustRegister("devs_logs_header", "<b>Bot Logs:</b>")
	MustRegister("devs_stats_header", "<b>Bot Statistics:</b>")
	MustRegister("devs_restart_message", "🔄 Restarting bot...")
	MustRegister("devs_update_success", "✅ Bot updated successfully!")
	MustRegister("devs_update_error", "❌ Update failed: %s", "error")
	MustRegister("devs_broadcast_success", "✅ Broadcast sent to %d chats", "count")
	MustRegister("devs_broadcast_error", "❌ Broadcast failed: %s", "error")
	MustRegister("devs_only_owner", "This command can only be used by the bot owner.")

	// Disabling module messages
	MustRegister("disabling_disabled_success", "Disabled <code>%s</code> for all non-admin users in this chat.", "command")
	MustRegister("disabling_enabled_success", "Enabled <code>%s</code> for all users in this chat.", "command")
	MustRegister("disabling_already_disabled", "Command <code>%s</code> is already disabled.", "command")
	MustRegister("disabling_already_enabled", "Command <code>%s</code> is already enabled.", "command")
	MustRegister("disabling_not_disableable", "Command <code>%s</code> cannot be disabled.", "command")
	MustRegister("disabling_specify_command", "Please specify a command to disable/enable.")
	MustRegister("disabling_available_commands", "<b>Available commands to disable:</b>\n%s", "commands")
	MustRegister("disabling_disabled_commands", "<b>Disabled commands in this chat:</b>\n%s", "commands")
	MustRegister("disabling_no_disabled_commands", "No commands are currently disabled in this chat.")
	MustRegister("disabling_del_enabled", "Disabled command deletion is now <b>enabled</b>. Disabled commands used by non-admins will be deleted.")
	MustRegister("disabling_del_disabled", "Disabled command deletion is now <b>disabled</b>. Disabled commands used by non-admins will not be deleted.")
	MustRegister("disabling_del_invalid_option", "Please use <code>yes</code>, <code>no</code>, <code>on</code>, or <code>off</code>.")

	// Language module messages
	MustRegister("language_current", "Your current language is set to <b>%s</b>.", "language")
	MustRegister("language_changed", "Language changed to <b>%s</b>!", "language")
	MustRegister("language_change_failed", "Failed to change language. Please try again.")
	MustRegister("language_not_supported", "Language <code>%s</code> is not supported.", "language")
	MustRegister("language_select_prompt", "Please select your preferred language:")
	MustRegister("language_group_changed", "Group language changed to <b>%s</b>!", "language")
	MustRegister("language_user_changed", "Your language changed to <b>%s</b>!", "language")

	// Bot updates module messages
	MustRegister("updates_enabled", "Update notifications are now <b>enabled</b> for this chat.")
	MustRegister("updates_disabled", "Update notifications are now <b>disabled</b> for this chat.")
	MustRegister("updates_status_enabled", "Update notifications are currently <b>enabled</b> for this chat.")
	MustRegister("updates_status_disabled", "Update notifications are currently <b>disabled</b> for this chat.")
	MustRegister("updates_invalid_option", "Please use <code>on</code> or <code>off</code>.")
	MustRegister("updates_new_version", "🚀 <b>New version available!</b>\n\n<b>Version:</b> %s\n<b>Changes:</b>\n%s", "version", "changes")
	MustRegister("updates_check_failed", "Failed to check for updates. Please try again later.")

	// Antispam module messages
	MustRegister("antispam_enabled", "Antispam protection is now <b>enabled</b> for this chat.")
	MustRegister("antispam_disabled", "Antispam protection is now <b>disabled</b> for this chat.")
	MustRegister("antispam_status_enabled", "Antispam protection is currently <b>enabled</b> for this chat.")
	MustRegister("antispam_status_disabled", "Antispam protection is currently <b>disabled</b> for this chat.")
	MustRegister("antispam_invalid_option", "Please use <code>on</code> or <code>off</code>.")
	MustRegister("antispam_user_detected", "🚨 Spam detected from %s! Taking action: %s", "user", "action")
	MustRegister("antispam_whitelist_added", "Added %s to antispam whitelist.", "user")
	MustRegister("antispam_whitelist_removed", "Removed %s from antispam whitelist.", "user")
	MustRegister("antispam_action_set", "Antispam action set to <b>%s</b>.", "action")
	MustRegister("antispam_action_invalid", "Invalid action. Use <code>ban</code>, <code>kick</code>, <code>mute</code>, or <code>warn</code>.")
	
	// Cache messages
	MustRegister("cache_reloaded", "Admin cache has been reloaded!")

	// Locks module messages
	MustRegister("locks_lock_success", "Locked the following in this group:\n - %s", "locked_items")
	MustRegister("locks_unlock_success", "Un-Locked the following in this group:\n - %s", "unlocked_items")
	MustRegister("locks_lock_what", "What do you want to lock? Check /locktypes for available options.")
	MustRegister("locks_unlock_what", "What do you want to unlock? Check /locktypes for available options.")
	MustRegister("locks_invalid_type", "`%s` is not a correct lock type, check /locktypes.", "type")
	MustRegister("locks_current_locks", "These are the locks in this chat:")
	MustRegister("locks_available_types", "Locks: \n - %s", "types")
	MustRegister("locks_bot_cant_join", "I see a bot, and I've been told to stop them joining... but I'm not admin!")
	MustRegister("locks_bot_cant_ban", "I see a bot, and I've been told to stop them joining... but I don't have permission to ban them!")
	MustRegister("locks_bots_locked", "Only admins are allowed to add bots to this chat!")

	// Mutes module messages
	MustRegister("mutes_mute_success", "Shh...\nMuted %s.", "user")
	MustRegister("mutes_mute_success_reason", "Shh...\nMuted %s.\n<b>Reason: </b>%s", "user", "reason")
	MustRegister("mutes_tmute_success", "Shh...\nMuted %s for %s", "user", "time")
	MustRegister("mutes_tmute_success_reason", "Shh...\nMuted %s for %s\n<b>Reason: </b>%s", "user", "time", "reason")
	MustRegister("mutes_unmute_success", "Alright!\nI'll allow %s to speak again.", "user")
	MustRegister("mutes_user_not_in_chat", "This user is not in this chat, how can I restrict them?")
	MustRegister("mutes_cannot_mute_admin", "Why would I mute an admin? That sounds like a pretty dumb idea.")
	MustRegister("mutes_cannot_mute_admin_alt", "I don't think you'd want me to mute an admin.")
	MustRegister("mutes_cannot_mute_self", "Why would I restrict myself?")
	MustRegister("mutes_button_unmute_admin", "Unmute (Admin Only)")

	// Notes module messages
	MustRegister("notes_saved_success", "Saved Note <b>%s</b>!\nGet it with <code>#%s</code> or <code>/get %s</code>.", "name", "name", "name")
	MustRegister("notes_saved_success_conflicting_flags", "Saved Note <b>%s</b>!\nGet it with <code>#%s</code> or <code>/get %s</code>.\n\n<b>Note:</b> This note will be sent to default setting of group notes, because it has both <code>{private}</code> and <code>{noprivate}</code>.", "name", "name", "name")
	MustRegister("notes_need_keyword", "Please give a keyword to reply to!")
	MustRegister("notes_invalid_note", "Invalid Note!")
	MustRegister("notes_already_exists", "Note already exists!\nDo you want to overwrite it?")
	MustRegister("notes_overwrite_success", "Note has been overwritten successfully ✅")
	MustRegister("notes_overwrite_cancelled", "Cancelled overwriting of note ❌")
	MustRegister("notes_removed_success", "Removed note <b>%s</b>.", "name")
	MustRegister("notes_not_found", "Note does not exist!")
	MustRegister("notes_need_keyword_remove", "Please give a keyword to remove!")
	MustRegister("notes_private_on", "Turned on Private Notes.\nNow users will get the notes as a private message.")
	MustRegister("notes_private_off", "Turned off Private Notes.\nNow all the notes will be sent to Group Chat.")
	MustRegister("notes_private_status_on", "Private Notes are currently turned on!")
	MustRegister("notes_private_status_off", "Private Notes are currently turned off!")
	MustRegister("notes_private_invalid_option", "I only understand an option from <on/off/yes/no>")
	MustRegister("notes_no_notes", "There are no notes in this chat!")
	MustRegister("notes_list_private_button", "Check on the button below to get Notes!")
	MustRegister("notes_list_private_click", "Click Me!")
	MustRegister("notes_current_notes", "These are the current notes in this Chat:")
	MustRegister("notes_current_notes_chat", "These are the current notes in <b>%s</b>:", "chat_title")
	MustRegister("notes_get_instructions", "You can get a note by <code>#notename</code> or <code>/get notename</code>")
	MustRegister("notes_not_enough_args", "Not enough arguments.")
	MustRegister("notes_not_exists", "Note doesn't exists!")
	MustRegister("notes_error_parsing", "There's some error parsing the note, please report this in support chat.")
	MustRegister("notes_admin_only", "This note can only be accessed by a admin!")
	MustRegister("notes_click_get_private", "Click on the button below to get the note *%s*", "name")
	MustRegister("notes_only_creator_can_clear_all", "Only Chat Creator can use this command.")
	MustRegister("notes_confirm_clear_all", "Are you sure you want to remove all Notes from this chat?")
	MustRegister("notes_cleared_all", "Removed all Notes from this Chat ✅")
	MustRegister("notes_clear_all_cancelled", "Cancelled removing all notes from this Chat ❌")

	// Pins module messages
	MustRegister("pins_reply_to_pin", "Reply to a message to pin it")
	MustRegister("pins_pinned_message", "I have pinned <a href='%s'>this message</a>.", "link")
	MustRegister("pins_pinned_message_loud", "I have pinned <a href='%s'>this message</a> loudly!", "link")
	MustRegister("pins_reply_not_pinned", "Replied message is not a pinned message.")
	MustRegister("pins_unpinned_this", "Unpinned this message.")
	MustRegister("pins_unpinned_last", "Unpinned the last pinned message.")
	MustRegister("pins_confirm_unpin_all", "Are you sure you want to unpin all pinned messages?")
	MustRegister("pins_unpinned_all", "Unpinned all pinned messages in this chat!")
	MustRegister("pins_unpin_all_cancelled", "Cancelled operation to unpin messages!")
	MustRegister("pins_reply_or_text", "Please reply to a message or give some text to pin.")
	MustRegister("pins_unsupported_data", "Permapin not supported for data type!")
	MustRegister("pins_no_pinned_message", "No message has been pinned in the current chat!")
	MustRegister("pins_here_pinned", "<a href='%s'>Here</a> is the pinned message.", "link")
	MustRegister("pins_pinned_message_button", "Pinned Message")
	MustRegister("pins_antichannel_enabled", "<b>Enabled</b> anti channel pins. Automatic pins from a channel will now be replaced with the previous pin.")
	MustRegister("pins_antichannel_disabled", "<b>Disabled</b> anti channel pins. Automatic pins from a channel will not be removed.")
	MustRegister("pins_antichannel_status_enabled", "Anti-Channel pins are currently <b>enabled</b> in %s. All channel posts that get auto-pinned by telegram will be replaced with the previous pin.", "chat_title")
	MustRegister("pins_antichannel_status_disabled", "Anti-Channel pins are currently <b>disabled</b> in %s.", "chat_title")
	MustRegister("pins_cleanlinked_enabled", "<b>Enabled</b> linked channel post deletion in Logs. Messages sent from the linked channel will be deleted.")
	MustRegister("pins_cleanlinked_disabled", "<b>Disabled</b> linked channel post deletion in Logs.")
	MustRegister("pins_cleanlinked_status_enabled", "Linked channel post deletion is currently <b>enabled</b> in %s. Messages sent from the linked channel will be deleted", "chat_title")
	MustRegister("pins_cleanlinked_status_disabled", "Linked channel post deletion is currently <b>disabled</b> in %s.", "chat_title")
	MustRegister("pins_invalid_option", "Your input was not recognised as one of: yes/no/on/off")

	// Purges module messages
	MustRegister("purges_reply_to_purge", "Reply to a message to select where to start purging from.")
	MustRegister("purges_reply_to_delete", "Reply to a message to delete it!")
	MustRegister("purges_purged_messages", "Purged %d messages.", "count")
	MustRegister("purges_purged_messages_reason", "Purged %d messages.\n*Reason*:\n%s", "count", "reason")
	MustRegister("purges_old_message_limit", "You cannot delete messages over two days old. Please choose a more recent message.")
	MustRegister("purges_marked_for_deletion", "Message marked for deletion. Reply to another message with /purgeto to delete all messages in between; within 30s!")
	MustRegister("purges_already_marked", "This message is already marked for purging!")
	MustRegister("purges_need_purgefrom", "You can only use this command after having used the /purgefrom command!")
	MustRegister("purges_use_del_single", "Use /del command to delete one message!")
	MustRegister("purges_reply_to_purgeto", "Reply to a message to show me till where to purge.")

	// Reports module messages
	MustRegister("reports_reply_to_report", "You need to reply to a message to report it.")
	MustRegister("reports_cannot_report_self", "You can't report your own message!")
	MustRegister("reports_need_expose", "You need to expose yourself first!")
	MustRegister("reports_special_account", "It's a special account of telegram!")
	MustRegister("reports_admin_no_report", "You're an admin, whom will I report your issues to?")
	MustRegister("reports_why_report_self", "Why would I report myself?")
	MustRegister("reports_why_report_admin", "Why would I report an admin?")
	MustRegister("reports_report_message", "<b>⚠️ Report:</b>\n<b> • Report by:</b> %s\n<b> • Reported user:</b> %s\n<b>Status:</b> <i>Pending...</i>", "reporter", "reported")
	MustRegister("reports_button_message", "➡ Message")
	MustRegister("reports_button_kick", "⚠ Kick")
	MustRegister("reports_button_ban", "⛔️ Ban")
	MustRegister("reports_button_delete", "❎ Delete Message")
	MustRegister("reports_button_resolve", "✔️ Mark Resolved")
	MustRegister("reports_turned_on_user", "Turned on reporting! You'll be notified whenever anyone reports something in groups you are admin.")
	MustRegister("reports_turned_off_user", "Turned off reporting! You'll no longer be notified whenever anyone reports something in groups you are admin.")
	MustRegister("reports_turned_on_chat", "Users will now be able to report messages.")
	MustRegister("reports_turned_off_chat", "Users will no longer be able to report via @admin or /report.")
	MustRegister("reports_private_only", "This command can only be used in a group!")
	MustRegister("reports_blocked_user", "Blocked user %s from reporting.", "user")
	MustRegister("reports_unblocked_user", "Unblocked user %s from reporting.", "user")
	MustRegister("reports_must_reply_block", "You must reply to a user to block them.")
	MustRegister("reports_must_reply_unblock", "You must reply to a user to unblock them.")
	MustRegister("reports_no_blocked_users", "No users are currently blocked from using report commands!")
	MustRegister("reports_blocked_users_list", "Users blocked from using report commands: ")
	MustRegister("reports_status_enabled_user", "Your current preference is true, You'll be notified whenever anyone reports something in groups you are admin.")
	MustRegister("reports_status_disabled_user", "You'll have nt enabled reports, You won't be notified.")
	MustRegister("reports_status_enabled_chat", "Reports are currently enabled in this chat.\nUsers can use the /report command, or mention @admin, to tag all admins.")
	MustRegister("reports_status_disabled_chat", "Reports are currently disabled in this chat.")
	MustRegister("reports_change_setting_info", "\n\nTo change this setting, try this command again, with one of the following args: yes/no/on/off")
	MustRegister("reports_invalid_option", "Your input was not recognised as one of: <code><yes/on/no/off> or <block/unblock/showblocklist></code>")
	MustRegister("reports_kicked_success", "✅ Successfully Kicked")
	MustRegister("reports_banned_success", "✅ Successfully Banned")
	MustRegister("reports_deleted_success", "✅ Successfully Deleted")
	MustRegister("reports_resolved_success", "✅ Resolved Report Successfully!")
	MustRegister("reports_action_kicked", "User kicked!Action taken by %s", "admin")
	MustRegister("reports_action_banned", "User banned!Action taken by %s", "admin")
	MustRegister("reports_action_deleted", "Message Deleted!Action taken by %s", "admin")
	MustRegister("reports_action_resolved", "<b>Resolved by:</b> %s", "admin")

	// Rules module messages
	MustRegister("rules_cleared_success", "Successfully cleared rules!")
	MustRegister("rules_private_enabled", "Use of /rules will send the rules to the user's PM.")
	MustRegister("rules_private_disabled", "All /rules commands will send the rules to %s.", "chat_title")
	MustRegister("rules_invalid_option", "Your input was not recognised as one of: yes/no/on/off")
	MustRegister("rules_for_chat", "The rules for <b>%s</b> are:\n\n", "chat_title")
	MustRegister("rules_no_rules", "This chat doesn't seem to have had any rules set yet... I wouldn't take that as an invitation though.")
	MustRegister("rules_click_button", "Click on the button to see the chat rules!")
	MustRegister("rules_need_rules", "You need to give me rules to set!\nEither reply to a message to provide them with command.")
	MustRegister("rules_set_success", "Successfully set rules for this group!")
	MustRegister("rules_button_too_long", "The custom button name you entered is too long. Please enter a text with less than 30 characters.")
	MustRegister("rules_button_set_success", "Successfully set the rules button to: <b>%s</b>", "button_text")
	MustRegister("rules_button_not_set", "You haven't set a custom rules button yet. The default text \"%s\" will be used.", "default_text")
	MustRegister("rules_button_current", "The rules button is currently set to the following text:\n %s", "button_text")
	MustRegister("rules_button_reset_success", "Successfully cleared custom rules button text!")

	// Users module messages
	MustRegister("users_info_user", "<b>User Info:</b>\n• <b>ID:</b> <code>%d</code>\n• <b>First Name:</b> %s", "id", "first_name")
	MustRegister("users_info_user_last_name", "<b>User Info:</b>\n• <b>ID:</b> <code>%d</code>\n• <b>First Name:</b> %s\n• <b>Last Name:</b> %s", "id", "first_name", "last_name")
	MustRegister("users_info_user_username", "<b>User Info:</b>\n• <b>ID:</b> <code>%d</code>\n• <b>First Name:</b> %s\n• <b>Username:</b> @%s", "id", "first_name", "username")
	MustRegister("users_info_user_full", "<b>User Info:</b>\n• <b>ID:</b> <code>%d</code>\n• <b>First Name:</b> %s\n• <b>Last Name:</b> %s\n• <b>Username:</b> @%s", "id", "first_name", "last_name", "username")
	MustRegister("users_current_chat_id", "<b>%s</b> chat ID: <code>%d</code>", "chat_title", "chat_id")
	MustRegister("users_ping_result", "<b>Ping:</b> <code>%s</code>\n<b>Time:</b> %s", "ping", "time")
	MustRegister("users_translate_usage", "Usage: <code>/tr &lt;lang code&gt; &lt;message&gt;</code> or reply to a message with <code>/tr &lt;lang code&gt;</code>")
	MustRegister("users_translate_result", "<b>Translation (%s → %s):</b>\n%s", "from_lang", "to_lang", "translation")
	MustRegister("users_translate_error", "Translation failed. Please try again.")
	MustRegister("users_keyboard_removed", "Removed stuck bot keyboard!")
	MustRegister("users_message_count", "Total messages in this chat: <b>%d</b>", "count")
	
	// Connection module messages
	MustRegister("connections_connected", "You are currently connected to <b>%s</b>!", "chat_title")
	MustRegister("connections_allow_connect_turned_on", "Turned <b>on</b> User connections to this chat!\nUsers can now connect chat to their PMs!")
	MustRegister("connections_allow_connect_turned_off", "Turned <b>off</b> User connections to this chat!\nUsers can't connect chat to their PM's!")
	MustRegister("connections_allow_connect_currently_on", "User connections are currently turned <b>on</b>.\nUsers can connect this chat to their PMs!")
	MustRegister("connections_allow_connect_currently_off", "User connections are currently turned <b>off</b>.\nUsers can't connect this chat to their PMs!")
	MustRegister("connections_allow_connect_invalid_option", "Please give me a valid option from <yes/on/no/off>")
	MustRegister("connections_connect_connection_disabled", "User connections are currently <b>disabled</b> to this chat.\nPlease ask admins to allow if you want to connect!")
	MustRegister("connections_connect_connected", "You are now connected to <b>%s</b>!", "chat_title")
	MustRegister("connections_connect_tap_btn_connect", "Please press the button below to connect this chat to your PM.")
	MustRegister("connections_connections_btns_admin_conn_cmds", "Available Admin commands:%s", "commands")
	MustRegister("connections_connections_btns_user_conn_cmds", "Available User commands:%s", "commands")
	MustRegister("connections_disconnect_disconnected", "Successfully disconnected from the connected chat.")
	MustRegister("connections_disconnect_need_pm", "You need to send this in PM to me to disconnect from the chat!")
	MustRegister("connections_not_connected", "You aren't connected to any chats.")
	MustRegister("connections_reconnect_reconnected", "You are now reconnected to <b>%s</b>!!", "chat_title")
	MustRegister("connections_reconnect_no_last_chat", "You have no last chat to reconnect!")
	MustRegister("connections_reconnect_need_pm", "You need to be in a PM with me to reconnect to a chat!")
	MustRegister("connections_is_user_connected_need_group", "This command can only be used in groups.")
	MustRegister("connections_is_user_connected_bot_not_admin", "I need to be an admin to check connection status.")
	MustRegister("connections_is_user_connected_user_not_admin", "You need to be an admin to use this command.")
	
	// Main language info
	MustRegister("main_language_name", "English")
	MustRegister("main_language_flag", "🇺🇸")
	
	// Captcha module messages
	MustRegister("captcha_settings_status", "<b>Captcha Settings:</b>\nStatus: <code>%s</code>\nMode: <code>%s</code>\nTimeout: <code>%d minutes</code>\nFailure Action: <code>%s</code>\nMax Attempts: <code>%d</code>\n\nUse <code>/captcha on</code> or <code>/captcha off</code> to change status.", "status", "mode", "timeout", "action", "attempts")
	MustRegister("captcha_enabled_success", "✅ Captcha verification has been <b>enabled</b>. New members will need to complete a captcha to join.")
	MustRegister("captcha_disabled_success", "❌ Captcha verification has been <b>disabled</b>.")
	MustRegister("captcha_enable_failed", "Failed to enable captcha. Please try again.")
	MustRegister("captcha_disable_failed", "Failed to disable captcha. Please try again.")
	MustRegister("captcha_invalid_option", "Invalid option. Use <code>on</code>/<code>off</code>, <code>enable</code>/<code>disable</code>, or <code>yes</code>/<code>no</code>.")
	MustRegister("captcha_mode_specify", "Please specify a mode: <code>math</code> or <code>text</code>")
	MustRegister("captcha_mode_invalid", "Invalid mode. Use <code>math</code> or <code>text</code>")
	MustRegister("captcha_mode_set_failed", "Failed to set captcha mode. Please try again.")
	MustRegister("captcha_mode_set_success", "✅ Captcha mode set to <b>%s</b>", "mode")
	MustRegister("captcha_timeout_specify", "Please specify timeout in minutes (1-10)")
	MustRegister("captcha_timeout_invalid", "Invalid timeout. Please use a number between 1 and 10.")
	MustRegister("captcha_timeout_set_failed", "Failed to set timeout. Please try again.")
	MustRegister("captcha_timeout_set_success", "✅ Captcha timeout set to <b>%d minutes</b>", "timeout")
	MustRegister("captcha_action_specify", "Please specify an action: <code>kick</code>, <code>ban</code>, or <code>mute</code>")
	MustRegister("captcha_action_invalid", "Invalid action. Use <code>kick</code>, <code>ban</code>, or <code>mute</code>")
	MustRegister("captcha_action_set_failed", "Failed to set failure action. Please try again.")
	MustRegister("captcha_action_set_success", "✅ Captcha failure action set to <b>%s</b>", "action")
	MustRegister("captcha_welcome_message", "Welcome to %s! Please complete the captcha below to verify you're human.", "chat_title")
	MustRegister("captcha_solve_prompt", "Solve this problem: <b>%s</b>\n\nYou have %d minutes to complete this captcha.", "problem", "timeout")
	MustRegister("captcha_text_prompt", "Enter the text shown in the image below:\n\nYou have %d minutes to complete this captcha.", "timeout")
	MustRegister("captcha_correct_answer", "✅ Correct! Welcome to the group.")
	MustRegister("captcha_wrong_answer", "❌ Wrong answer. Try again. Attempts left: %d", "attempts")
	MustRegister("captcha_failed_attempts", "❌ You have failed the captcha verification. You have been %s.", "action")
	MustRegister("captcha_timeout_message", "⏰ Captcha verification timed out. You have been %s.", "action")
	MustRegister("captcha_refresh_button", "🔄 New Problem")
	MustRegister("captcha_refreshed", "Here's a new problem for you:")
	MustRegister("captcha_refresh_limit", "You've reached the maximum number of refreshes (%d). Please solve the current problem.", "limit")
	MustRegister("captcha_refresh_cooldown", "Please wait %d seconds before requesting a new problem.", "seconds")
	MustRegister("captcha_already_verified", "You have already been verified!")
	MustRegister("captcha_not_for_you", "This captcha is not for you.")
	MustRegister("captcha_welcome_math_image", "👋 Welcome %s!\n\nPlease solve the problem shown in the image and select the correct answer:\n\n⏱ You have <b>%d minutes</b> to answer.", "user", "timeout")
	MustRegister("captcha_welcome_text_image", "👋 Welcome %s!\n\nPlease select the text shown in the image to verify you're human:\n\n⏱ You have <b>%d minutes</b> to answer.", "user", "timeout")
	MustRegister("captcha_welcome_math_text", "👋 Welcome %s!\n\nPlease solve this math problem to verify you're human:\n\n<b>%s = ?</b>\n\n⏱ You have <b>%d minutes</b> to answer.", "user", "question", "timeout")
	MustRegister("captcha_verification_success", "✅ %s has been verified and can now chat!", "user")
	MustRegister("captcha_wrong_answer_action", "❌ Wrong answer! You have been %s.", "action")
	MustRegister("captcha_wrong_answer_attempts", "❌ Wrong answer! %d attempts remaining.", "attempts")

	// Warns module messages
	MustRegister("warns_warn_mode_updated", "Updated warn mode to: %s", "mode")
	MustRegister("warns_warn_mode_unknown", "Unknown type '%s'. Please use one of: ban/kick/mute", "type")
	MustRegister("warns_warn_mode_specify", "You need to specify an action to take upon too many warns. Current modes are: ban/kick/mute")
	MustRegister("warns_cannot_warn_admin", "I'm not going to warn an admin!")
	MustRegister("warns_user_warned", "User %s has %d/%d warnings; be careful!", "user", "current", "limit")
	MustRegister("warns_user_warned_reason", "User %s has %d/%d warnings; be careful!\n<b>Reason</b>:\n%s", "user", "current", "limit", "reason")
	MustRegister("warns_limit_exceeded_kick", "That's %d/%d warnings; So %s has been kicked!", "current", "limit", "user")
	MustRegister("warns_limit_exceeded_mute", "That's %d/%d warnings; So %s has been Muted!", "current", "limit", "user")
	MustRegister("warns_limit_exceeded_ban", "That's %d/%d warnings; So %s has been banned!", "current", "limit", "user")
	MustRegister("warns_no_warnings", "This user hasn't got any warnings!")
	MustRegister("warns_user_warnings_with_reasons", "This user has %d/%d warnings, for the following reasons:", "current", "limit")
	MustRegister("warns_user_warnings_no_reasons", "User has %d/%d warnings, but no reasons for any of them.", "current", "limit")
	MustRegister("warns_warning_removed_by", "Warn removed by %s.", "admin")
	MustRegister("warns_no_warnings_to_remove", "User already has no warns!")
	MustRegister("warns_current_settings", "The group has the following settings:\n<b>Warn Limit:</b> <code>%d</code>\n<b>Warn Mode:</b> <code>%s</code>", "limit", "mode")
	MustRegister("warns_limit_specify", "Please specify how many warns a user should be allowed to receive before being acted upon. Eg. `/setwarnlimit 5`")
	MustRegister("warns_limit_not_integer", "%s is not a valid integer.", "value")
	MustRegister("warns_limit_range_error", "The warn limit has to be set between 1 and 100.")
	MustRegister("warns_limit_updated", "Warn limit settings for this chat have been updated to %d.", "limit")
	MustRegister("warns_warnings_reset", "Warnings have been reset!")
	MustRegister("warns_no_users_warned", "No users are warned in this chat!")
	MustRegister("warns_confirm_reset_all", "Are you sure you want to remove all the warns of all the users in this chat?")

	// Filters module messages
	MustRegister("filters_limit_exceeded", "Filters limit exceeded, a group can only have maximum 150 filters!\nThis limitation is due to bot running free without any donations by users.")
	MustRegister("filters_provide_keyword", "Please give a keyword to reply to!")
	MustRegister("filters_invalid_filter", "Invalid Filter!")
	MustRegister("filters_filter_exists_overwrite", "Filter already exists!\nDo you want to overwrite it?")
	MustRegister("filters_filter_added", "Added reply for filter word <code>%s</code>", "keyword")
	MustRegister("filters_no_keyword_provided", "Please give a filter word to remove!")
	MustRegister("filters_filter_not_found", "Filter does not exist!", "keyword")
	MustRegister("filters_filter_removed", "Ok!\nI will no longer reply to <code>%s</code>", "keyword")
	MustRegister("filters_no_filters", "There are no filters in this chat!")
	MustRegister("filters_list_filters", "These are the current filters in this Chat:%s", "filters")
	MustRegister("filters_confirm_remove_all", "Are you sure you want to remove all Filters from this chat?")
	MustRegister("filters_all_removed", "Removed all Filters from this Chat ✅")
	MustRegister("filters_remove_all_cancelled", "Cancelled removing all Filters from this Chat ❌")
	MustRegister("filters_filter_overwritten", "Filter has been overwritten successfully ✅")
	MustRegister("filters_overwrite_cancelled", "Cancelled overwritting of filter ❌")

	// Greetings module messages
	MustRegister("greetings_welcome_set", "Successfully set custom welcome message!")
	MustRegister("greetings_welcome_reset", "Successfully reset custom welcome message to default!")
	MustRegister("greetings_welcome_enabled", "I'll welcome users from now on.")
	MustRegister("greetings_welcome_disabled", "I'll not welcome users from now on.")
	MustRegister("greetings_goodbye_set", "Successfully set custom goodbye message!")
	MustRegister("greetings_goodbye_reset", "Successfully reset custom goodbye message to default!")
	MustRegister("greetings_goodbye_enabled", "I'll goodbye users from now on.")
	MustRegister("greetings_goodbye_disabled", "I'll not goodbye users from now on.")
	MustRegister("greetings_cleanservice_enabled", "Service message deletion is now enabled.")
	MustRegister("greetings_cleanservice_disabled", "Service message deletion is now disabled.")
	MustRegister("greetings_cleanwelcome_enabled", "I'll try to delete old welcome messages!")
	MustRegister("greetings_cleanwelcome_disabled", "I'll not delete old welcome messages!")
	MustRegister("greetings_cleanwelcome_status_enabled", "I should be deleting welcome messages up to two days old.")
	MustRegister("greetings_cleanwelcome_status_disabled", "I'm currently not deleting old welcome messages!")
	MustRegister("greetings_autoapprove_enabled", "Auto-approval for new members is now enabled.")
	MustRegister("greetings_autoapprove_disabled", "Auto-approval for new members is now disabled.")
	MustRegister("greetings_provide_message", "Please provide a message to set.")
	MustRegister("greetings_invalid_option", "I understand 'on/yes' or 'off/no' only!")
	MustRegister("greetings_welcome_status", "I am currently welcoming users: <code>%t</code>\nI am currently deleting old welcomes: <code>%t</code>\nI am currently deleting service messages: <code>%t</code>\nThe welcome message not filling the {} is:", "should_welcome", "clean_welcome", "clean_service")
	MustRegister("greetings_goodbye_status", "I am currently goodbying users: <code>%t</code>\nI am currently deleting old goodbyes: <code>%t</code>\nI am currently deleting service messages: <code>%t</code>\nThe goodbye message not filling the {} is:", "should_goodbye", "clean_goodbye", "clean_service")
	MustRegister("greetings_cleangoodbye_enabled", "I'll try to delete old goodbye messages!")
	MustRegister("greetings_cleangoodbye_disabled", "I'll not delete old goodbye messages!")
	MustRegister("greetings_cleanservice_status_enabled", "I should be deleting `user` joined the chat messages now.")
	MustRegister("greetings_cleanservice_status_disabled", "I'm currently not deleting joined messages.")
	MustRegister("greetings_cleanservice_enabled", "I'll try to delete joined messages!")
	MustRegister("greetings_cleanservice_disabled", "I won't delete joined messages.")
	MustRegister("greetings_autoapprove_status_enabled", "I'm auto-approving new chat join requests now.")
	MustRegister("greetings_autoapprove_status_disabled", "I'm not auto-approving new chat join requests now..")
	MustRegister("greetings_autoapprove_enabled", "I'll try to auto-approve new join requests!")
	MustRegister("greetings_autoapprove_disabled", "I won't auto-approve new join requests!")

	// Help module messages
	MustRegister("help_about", "@%s  is one of the fastest and most feature-filled group managers.\n\nAlita ✨ is developed and actively maintained by @DivideProjects!\n\nAlita has been online since 2020 and has served thousands of groups with hundreds of thousands of users!\n\n<b>Why Alita:</b>\n- Simple: Easy usage and compatible with many bot commands.\n- Featured: Many features which other group management bots don't have.\n- Fast: Guess what? It's not made using Python, we use <a href='https://go.dev/'>Go</a> as our core programming language.\n\n<b>Current Version:</b> %s", "username", "version")
	MustRegister("help_donate_text", "So you want to donate? Amazing!\nWhen you donate, all the fund goes towards my development which makes on fast and responsive.\nYour donation might also get me a new feature or two, which I wasn't able to get due to server limitations.\nAll the funds would be put into my services such as database, storage, and hosting!\nYou can donate by contacting my owner here: @DivideProjectsBot")
	MustRegister("help_configuration_step_1", "Welcome to the Alita Configuration\n\nThe first thing to do is to add Alita ✨ to your group! For doing that, press the under button and select your group, then press Done to continue the tutorial.")
	MustRegister("help_configuration_step_2", "Ok, well done!\n\nNow to let me work correctly, you need to make me Admin of your Group!\nTo do that, follow these easy steps:\n▫️ Go to your group\n▫️ Press the Group's name\n▫️ Press Modify\n▫️ Press on Administrator\n▫️ Press Add Administrator\n▫️ Press the Magnifying Glass\n▫️ Search @%s\n▫️ Confirm", "username")
	MustRegister("help_configuration_step_3", "Excellent!\nNow the Bot is ready to use!\nAll commands can be used with / or !\n\nIf you're facing any difficulties in setting up me in your group, so don't hesitate to come in @DivideSupport.\nWe would love to help you.")
	MustRegister("help_click_button_info", "Click on the button below to get info about me!")
	MustRegister("help_pm_me_questions", "Hey :) PM me if you have any questions on how to use me!")
	MustRegister("help_click_here_help", "Click here for help!")
	MustRegister("help_contact_pm", "Contact me in PM for help!")
	MustRegister("help_contact_pm_module", "Contact me in PM for help regarding <code>%s</code>!", "module")
	MustRegister("help_no_rules", "This chat does not have any rules!")
	MustRegister("help_note_not_exist", "This note does not exist!")
	MustRegister("help_note_admin_only", "This note can only be accessed by a admin!")
	MustRegister("help_no_notes", "There are no notes in this chat!")
	MustRegister("help_current_notes", "These are the current notes in this Chat:")

	log.Infof("Registered %d default messages in i18n catalog", Count())
}
