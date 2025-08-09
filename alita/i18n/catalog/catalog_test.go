package catalog

import (
	"testing"
)

func TestBasicCatalogOperations(t *testing.T) {
	// Clear catalog for clean test
	Clear()
	
	// Test registration
	err := Register("test.message", "Hello {name}!", "name")
	if err != nil {
		t.Fatalf("Failed to register message: %v", err)
	}
	
	// Test retrieval
	msg, exists := Get("test.message")
	if !exists {
		t.Fatal("Message should exist after registration")
	}
	
	if msg.Key != "test.message" {
		t.Errorf("Expected key 'test.message', got '%s'", msg.Key)
	}
	
	if msg.Default != "Hello {name}!" {
		t.Errorf("Expected default 'Hello {name}!', got '%s'", msg.Default)
	}
	
	if len(msg.Params) != 1 || msg.Params[0] != "name" {
		t.Errorf("Expected params ['name'], got %v", msg.Params)
	}
}

func TestParameterInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		template string
		params   Params
		expected string
		wantErr  bool
	}{
		{
			name:     "simple interpolation",
			template: "Hello {name}!",
			params:   Params{"name": "World"},
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:     "multiple parameters",
			template: "Hello {name}, welcome to {place}!",
			params:   Params{"name": "John", "place": "Telegram"},
			expected: "Hello John, welcome to Telegram!",
			wantErr:  false,
		},
		{
			name:     "no parameters needed",
			template: "Hello World!",
			params:   Params{},
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:     "missing parameter",
			template: "Hello {name}!",
			params:   Params{"other": "value"},
			expected: "Hello {name}!",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InterpolateParams(tt.template, tt.params)
			
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTranslator(t *testing.T) {
	// Clear and register test messages
	Clear()
	MustRegister("greeting.hello", "Hello {name}!", "name")
	MustRegister("greeting.goodbye", "Goodbye {name}!", "name")
	
	// Create translator
	config := DefaultConfig()
	translator := NewTranslator("en", config)
	
	// Test message retrieval
	params := Params{"name": "Alice"}
	msg := translator.Message("greeting.hello", params)
	
	expected := "Hello Alice!"
	if msg != expected {
		t.Errorf("Expected '%s', got '%s'", expected, msg)
	}
	
	// Test missing message
	msg = translator.Message("nonexistent.key", nil)
	if msg != "{{nonexistent.key}}" {
		t.Errorf("Expected placeholder for missing key, got '%s'", msg)
	}
}

func TestValidation(t *testing.T) {
	Clear()
	MustRegister("test.validation", "Hello {name} from {place}!", "name", "place")
	
	validator := NewStandardParamValidator()
	
	// Test valid parameters
	params := Params{"name": "John", "place": "Earth"}
	err := validator.ValidateParams([]string{"name", "place"}, params)
	if err != nil {
		t.Errorf("Expected no error for valid params, got: %v", err)
	}
	
	// Test missing parameter
	params = Params{"name": "John"}
	err = validator.ValidateParams([]string{"name", "place"}, params)
	if err == nil {
		t.Error("Expected error for missing parameter")
	}
	
	// Test extra parameter
	params = Params{"name": "John", "place": "Earth", "extra": "value"}
	err = validator.ValidateParams([]string{"name", "place"}, params)
	if err == nil {
		t.Error("Expected error for extra parameter")
	}
}

func TestCatalogStats(t *testing.T) {
	Clear()
	
	// Register test messages
	MustRegister("admin.promote", "Promoted {user}!", "user")
	MustRegister("admin.demote", "Demoted {user}!", "user")
	MustRegister("greetings.hello", "Hello!")
	MustRegister("greetings.bye", "Goodbye!")
	
	stats := GetStats()
	
	if stats.TotalMessages != 4 {
		t.Errorf("Expected 4 total messages, got %d", stats.TotalMessages)
	}
	
	if stats.MessagesByPrefix["admin"] != 2 {
		t.Errorf("Expected 2 admin messages, got %d", stats.MessagesByPrefix["admin"])
	}
	
	if stats.MessagesByPrefix["greetings"] != 2 {
		t.Errorf("Expected 2 greeting messages, got %d", stats.MessagesByPrefix["greetings"])
	}
	
	if stats.MessagesWithParams != 2 {
		t.Errorf("Expected 2 messages with params, got %d", stats.MessagesWithParams)
	}
}

func BenchmarkMessageLookup(b *testing.B) {
	Clear()
	MustRegister("benchmark.test", "Hello {name}!", "name")
	
	translator := NewTranslator("en", DefaultConfig())
	params := Params{"name": "World"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = translator.Message("benchmark.test", params)
	}
}

func BenchmarkParameterInterpolation(b *testing.B) {
	template := "Hello {name}, welcome to {place} on {date}!"
	params := Params{
		"name":  "John Doe",
		"place": "Telegram Bot",
		"date":  "2024-01-01",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InterpolateParams(template, params)
	}
}