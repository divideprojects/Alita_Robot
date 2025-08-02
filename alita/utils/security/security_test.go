package security

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	userID := int64(12345)

	// Test normal usage within limit
	for i := 0; i < 3; i++ {
		if !rl.IsAllowed(userID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Test rate limiting
	if rl.IsAllowed(userID) {
		t.Error("Request should be rate limited")
	}

	// Test different user
	otherUserID := int64(67890)
	if !rl.IsAllowed(otherUserID) {
		t.Error("Different user should be allowed")
	}
}

func TestValidateTextContent(t *testing.T) {
	sm := NewSecurityMiddleware()

	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"Normal text", "Hello world!", false},
		{"Empty text", "", false},
		{"Long text", string(make([]byte, 5000)), true}, // Too long
		{"Script tag", "<script>alert('xss')</script>", true},
		{"JavaScript URL", "javascript:alert('xss')", true},
		{"SQL injection", "'; DROP TABLE users; --", true},
		{"NoSQL injection", "{$where: 'this.password.length > 0'}", true},
		{"Normal HTML", "<b>Bold text</b>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.validateTextContent(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTextContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContainsSuspiciousPatterns(t *testing.T) {
	sm := NewSecurityMiddleware()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"Normal text", "Hello world", false},
		{"Script tag lowercase", "<script>", true},
		{"Script tag uppercase", "<SCRIPT>", true},
		{"JavaScript URL", "javascript:void(0)", true},
		{"OnLoad event", "onload=alert(1)", true},
		{"Data URL", "data:text/html,<script>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sm.containsSuspiciousPatterns(tt.text)
			if got != tt.want {
				t.Errorf("containsSuspiciousPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsInjectionAttempts(t *testing.T) {
	sm := NewSecurityMiddleware()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"Normal text", "Hello world", false},
		{"SQL injection OR", "' OR '1'='1", true},
		{"SQL injection DROP", "'; DROP TABLE users; --", true},
		{"SQL injection UNION", "UNION SELECT password FROM users", true},
		{"NoSQL injection $where", "{$where: 'this.password'}", true},
		{"NoSQL injection $regex", "{$regex: '^admin'}", true},
		{"Normal query", "SELECT * FROM products WHERE name = 'test'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sm.containsInjectionAttempts(tt.text)
			if got != tt.want {
				t.Errorf("containsInjectionAttempts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"Valid command", "start", false},
		{"Valid command with slash", "/help", false},
		{"Valid command with underscore", "admin_list", false},
		{"Valid command with numbers", "warn123", false},
		{"Empty command", "", true},
		{"Command with semicolon", "test;rm", true},
		{"Command with pipe", "test|cat", true},
		{"Command with backtick", "test`whoami`", true},
		{"Command with dollar", "test$USER", true},
		{"Command with parentheses", "test()", true},
		{"Command with braces", "test{}", true},
		{"Command with brackets", "test[]", true},
		{"Command with ampersand", "test&", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileUpload(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		fileSize int64
		wantErr  bool
	}{
		{"Valid file", "document.pdf", 1024, false},
		{"Valid image", "photo.jpg", 2048, false},
		{"Too large", "large.zip", 11 * 1024 * 1024, true}, // 11MB
		{"Executable file", "virus.exe", 1024, true},
		{"Script file", "malware.js", 1024, true},
		{"PHP file", "shell.php", 1024, true},
		{"Path traversal", "../../../etc/passwd", 1024, true},
		{"Windows path", "..\\..\\windows\\system32", 1024, true},
		{"Empty filename", "", 1024, true},
		{"Long filename", string(make([]byte, 300)), 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileUpload(tt.filename, tt.fileSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileUpload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"Remove script tags",
			"<p>Hello</p><script>alert('xss')</script>",
			"<p>Hello</p>",
		},
		{
			"Remove dangerous attributes",
			"<a href='javascript:alert(1)' onclick='alert(2)'>Link</a>",
			"<a>Link</a>",
		},
		{
			"Remove dangerous tags",
			"<iframe src='evil.com'></iframe><p>Safe</p>",
			"<p>Safe</p>",
		},
		{
			"Normal HTML",
			"<p><b>Bold</b> and <i>italic</i> text</p>",
			"<p><b>Bold</b> and <i>italic</i> text</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTML(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rl := NewRateLimiter(100, time.Minute)
	userID := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.IsAllowed(userID)
	}
}

func BenchmarkValidateTextContent(b *testing.B) {
	sm := NewSecurityMiddleware()
	text := "This is a normal message with some text content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.validateTextContent(text)
	}
}
