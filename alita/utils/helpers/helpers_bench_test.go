package helpers

import (
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/divideprojects/Alita_Robot/alita/db"
)

// BenchmarkSplitMessage benchmarks the SplitMessage function
func BenchmarkSplitMessage(b *testing.B) {
	longMessage := strings.Repeat("This is a test message that needs to be split. ", 200)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SplitMessage(longMessage)
	}
}

// BenchmarkSplitMessageShort benchmarks SplitMessage with short messages
func BenchmarkSplitMessageShort(b *testing.B) {
	shortMessage := "Short message"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SplitMessage(shortMessage)
	}
}

// BenchmarkFormattingReplacer benchmarks the FormattingReplacer function
func BenchmarkFormattingReplacer(b *testing.B) {
	// Create mock objects
	bot := &gotgbot.Bot{}
	bot.Username = "testbot" // Set username after creation
	chat := &gotgbot.Chat{
		Id:    -1001234567890,
		Title: "Test Chat",
	}
	user := &gotgbot.User{
		Id:        123456789,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
	}
	
	message := "Welcome {first} {last} ({username}) to {chatname}! You are user #{id} and we have {count} members total. {rules:same}"
	buttons := []db.Button{{Name: "Test", Url: "https://example.com", SameLine: false}}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FormattingReplacer(bot, chat, user, message, buttons)
	}
}

// BenchmarkNotesParser benchmarks the notesParser function
func BenchmarkNotesParser(b *testing.B) {
	message := "This is a test message with {private} and {admin} and {preview} flags {protect} {nonotif}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _, _, _, _ = notesParser(message)
	}
}

// BenchmarkMentionHtml benchmarks the MentionHtml function
func BenchmarkMentionHtml(b *testing.B) {
	userId := int64(123456789)
	name := "John Doe"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MentionHtml(userId, name)
	}
}

// BenchmarkBuildKeyboard benchmarks the BuildKeyboard function
func BenchmarkBuildKeyboard(b *testing.B) {
	buttons := []db.Button{
		{Name: "Button 1", Url: "https://example1.com", SameLine: false},
		{Name: "Button 2", Url: "https://example2.com", SameLine: true},
		{Name: "Button 3", Url: "https://example3.com", SameLine: false},
		{Name: "Button 4", Url: "https://example4.com", SameLine: true},
		{Name: "Button 5", Url: "https://example5.com", SameLine: false},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildKeyboard(buttons)
	}
} 