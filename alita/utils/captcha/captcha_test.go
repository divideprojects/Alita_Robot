package captcha

import (
	"bytes"
	"encoding/json"
	"image/png"
	"strconv"
	"strings"
	"testing"
)

func TestCaptchaGenerator_GenerateChallenge(t *testing.T) {
	generator := NewCaptchaGenerator()

	tests := []struct {
		mode        string
		expectImage bool
		expectError bool
		description string
	}{
		{
			mode:        "button",
			expectImage: false,
			expectError: false,
			description: "Button mode should not generate image",
		},
		{
			mode:        "text",
			expectImage: true,
			expectError: false,
			description: "Text mode should generate image",
		},
		{
			mode:        "math",
			expectImage: false,
			expectError: false,
			description: "Math mode should not generate image",
		},
		{
			mode:        "text2",
			expectImage: true,
			expectError: false,
			description: "Text2 mode should generate image",
		},
		{
			mode:        "invalid",
			expectImage: false,
			expectError: true,
			description: "Invalid mode should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result, err := generator.GenerateChallenge(tt.mode)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for mode %s, but got none", tt.mode)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for mode %s: %v", tt.mode, err)
				return
			}

			if result == nil {
				t.Errorf("Expected result for mode %s, but got nil", tt.mode)
				return
			}

			if result.ChallengeData == "" {
				t.Errorf("Expected non-empty challenge data for mode %s", tt.mode)
			}

			if result.CorrectAnswer == "" {
				t.Errorf("Expected non-empty correct answer for mode %s", tt.mode)
			}

			if tt.expectImage {
				if result.ImageBytes == nil {
					t.Errorf("Expected image bytes for mode %s, but got nil", tt.mode)
				} else if len(result.ImageBytes) == 0 {
					t.Errorf("Expected non-zero length image bytes for mode %s", tt.mode)
				}
			} else {
				if result.ImageBytes != nil {
					t.Errorf("Expected no image bytes for mode %s, but got %d bytes", tt.mode, len(result.ImageBytes))
				}
			}
		})
	}
}

func TestCaptchaGenerator_ButtonChallenge(t *testing.T) {
	generator := NewCaptchaGenerator()

	result, err := generator.GenerateChallenge("button")
	if err != nil {
		t.Fatalf("Failed to generate button challenge: %v", err)
	}

	// Validate challenge structure
	var challenge ButtonChallenge
	err = json.Unmarshal([]byte(result.ChallengeData), &challenge)
	if err != nil {
		t.Fatalf("Failed to unmarshal button challenge: %v", err)
	}

	if challenge.Type != "button" {
		t.Errorf("Expected challenge type 'button', got '%s'", challenge.Type)
	}

	if result.CorrectAnswer != "button_click" {
		t.Errorf("Expected correct answer 'button_click', got '%s'", result.CorrectAnswer)
	}
}

func TestCaptchaGenerator_TextChallenge(t *testing.T) {
	generator := NewCaptchaGenerator()

	result, err := generator.GenerateChallenge("text")
	if err != nil {
		t.Fatalf("Failed to generate text challenge: %v", err)
	}

	// Validate challenge structure
	var challenge TextChallenge
	err = json.Unmarshal([]byte(result.ChallengeData), &challenge)
	if err != nil {
		t.Fatalf("Failed to unmarshal text challenge: %v", err)
	}

	if challenge.Type != "text" {
		t.Errorf("Expected challenge type 'text', got '%s'", challenge.Type)
	}

	if len(challenge.Options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(challenge.Options))
	}

	if challenge.Answer < 0 || challenge.Answer >= len(challenge.Options) {
		t.Errorf("Answer index %d is out of range for options length %d", challenge.Answer, len(challenge.Options))
	}

	// Verify the correct answer is in the options
	correctOption := challenge.Options[challenge.Answer]
	if correctOption != result.CorrectAnswer {
		t.Errorf("Expected correct answer '%s', got '%s'", correctOption, result.CorrectAnswer)
	}

	// Validate that image can be decoded as PNG
	if result.ImageBytes != nil {
		_, err := png.Decode(bytes.NewReader(result.ImageBytes))
		if err != nil {
			t.Errorf("Failed to decode generated image as PNG: %v", err)
		}
	}
}

func TestCaptchaGenerator_MathChallenge(t *testing.T) {
	generator := NewCaptchaGenerator()

	result, err := generator.GenerateChallenge("math")
	if err != nil {
		t.Fatalf("Failed to generate math challenge: %v", err)
	}

	// Validate challenge structure
	var challenge MathChallenge
	err = json.Unmarshal([]byte(result.ChallengeData), &challenge)
	if err != nil {
		t.Fatalf("Failed to unmarshal math challenge: %v", err)
	}

	if challenge.Type != "math" {
		t.Errorf("Expected challenge type 'math', got '%s'", challenge.Type)
	}

	if len(challenge.Options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(challenge.Options))
	}

	if challenge.Question == "" {
		t.Error("Expected non-empty question")
	}

	if challenge.Answer == "" {
		t.Error("Expected non-empty answer")
	}

	// Verify the answer is valid
	answerInt, err := strconv.Atoi(challenge.Answer)
	if err != nil {
		t.Errorf("Answer should be a valid integer: %v", err)
	}

	// Verify the correct answer is in the options
	found := false
	for _, option := range challenge.Options {
		if option == answerInt {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Correct answer %d not found in options %v", answerInt, challenge.Options)
	}

	// Verify math operations are valid
	validOps := []string{"+", "-", "×"}
	hasValidOp := false
	for _, op := range validOps {
		if strings.Contains(challenge.Question, op) {
			hasValidOp = true
			break
		}
	}
	if !hasValidOp {
		t.Errorf("Question '%s' doesn't contain a valid operation", challenge.Question)
	}
}

func TestCaptchaGenerator_Text2Challenge(t *testing.T) {
	generator := NewCaptchaGenerator()

	result, err := generator.GenerateChallenge("text2")
	if err != nil {
		t.Fatalf("Failed to generate text2 challenge: %v", err)
	}

	// Validate challenge structure
	var challenge Text2Challenge
	err = json.Unmarshal([]byte(result.ChallengeData), &challenge)
	if err != nil {
		t.Fatalf("Failed to unmarshal text2 challenge: %v", err)
	}

	if challenge.Type != "text2" {
		t.Errorf("Expected challenge type 'text2', got '%s'", challenge.Type)
	}

	if len(challenge.Characters) == 0 {
		t.Error("Expected non-empty characters array")
	}

	if challenge.Answer == "" {
		t.Error("Expected non-empty answer")
	}

	// Verify answer length is between 4-5 characters
	if len(challenge.Answer) < 4 || len(challenge.Answer) > 5 {
		t.Errorf("Expected answer length between 4-5, got %d", len(challenge.Answer))
	}

	// Verify answer contains only valid characters
	validChars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for _, char := range challenge.Answer {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("Answer contains invalid character: %c", char)
		}
	}

	// Validate that image can be decoded as PNG
	if result.ImageBytes != nil {
		_, err := png.Decode(bytes.NewReader(result.ImageBytes))
		if err != nil {
			t.Errorf("Failed to decode generated image as PNG: %v", err)
		}
	}
}

func TestCreateCaptchaKeyboard(t *testing.T) {
	generator := NewCaptchaGenerator()

	tests := []struct {
		mode        string
		description string
	}{
		{"button", "Button keyboard should have verify button"},
		{"text", "Text keyboard should have option buttons"},
		{"math", "Math keyboard should have number buttons"},
		{"text2", "Text2 keyboard should have character buttons"},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			// Generate a challenge first to get valid challenge data
			result, err := generator.GenerateChallenge(tt.mode)
			if err != nil {
				t.Fatalf("Failed to generate %s challenge: %v", tt.mode, err)
			}

			keyboard, err := CreateCaptchaKeyboard(result.ChallengeData, tt.mode)
			if err != nil {
				t.Errorf("Failed to create keyboard for mode %s: %v", tt.mode, err)
				return
			}

			if keyboard == nil {
				t.Errorf("Expected keyboard for mode %s, got nil", tt.mode)
				return
			}

			if len(keyboard.InlineKeyboard) == 0 {
				t.Errorf("Expected non-empty keyboard for mode %s", tt.mode)
			}

			// Validate specific keyboard structures
			switch tt.mode {
			case "button":
				if len(keyboard.InlineKeyboard) != 1 || len(keyboard.InlineKeyboard[0]) != 1 {
					t.Errorf("Button keyboard should have 1 row with 1 button")
				}
				if keyboard.InlineKeyboard[0][0].Text != "✅ Verify" {
					t.Errorf("Expected button text '✅ Verify', got '%s'", keyboard.InlineKeyboard[0][0].Text)
				}
			case "text":
				if len(keyboard.InlineKeyboard) != 4 {
					t.Errorf("Text keyboard should have 4 rows, got %d", len(keyboard.InlineKeyboard))
				}
			case "math":
				if len(keyboard.InlineKeyboard) != 4 {
					t.Errorf("Math keyboard should have 4 rows, got %d", len(keyboard.InlineKeyboard))
				}
			case "text2":
				// Should have character rows + control row
				if len(keyboard.InlineKeyboard) < 2 {
					t.Errorf("Text2 keyboard should have at least 2 rows, got %d", len(keyboard.InlineKeyboard))
				}
				// Last row should have Delete and Submit buttons
				lastRow := keyboard.InlineKeyboard[len(keyboard.InlineKeyboard)-1]
				if len(lastRow) != 2 {
					t.Errorf("Last row should have 2 buttons, got %d", len(lastRow))
				}
			}
		})
	}
}

func TestCreateCaptchaKeyboard_InvalidMode(t *testing.T) {
	_, err := CreateCaptchaKeyboard("{}", "invalid_mode")
	if err == nil {
		t.Error("Expected error for invalid mode, got nil")
	}
}

func TestGetChallengeDescription(t *testing.T) {
	generator := NewCaptchaGenerator()

	tests := []struct {
		mode     string
		expected string
	}{
		{"button", "Click the button below to verify you're human."},
		{"text", "Select the text that matches the image above."},
		{"text2", "Enter the characters you see in the image above, one by one."},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			// Generate a challenge first to get valid challenge data
			result, err := generator.GenerateChallenge(tt.mode)
			if err != nil {
				t.Fatalf("Failed to generate %s challenge: %v", tt.mode, err)
			}

			description, err := GetChallengeDescription(result.ChallengeData, tt.mode)
			if err != nil {
				t.Errorf("Failed to get description for mode %s: %v", tt.mode, err)
				return
			}

			if tt.mode == "math" {
				// Math descriptions are dynamic, just check it contains "Solve this math problem"
				if !strings.Contains(description, "Solve this math problem") {
					t.Errorf("Math description should contain 'Solve this math problem', got '%s'", description)
				}
			} else {
				if description != tt.expected {
					t.Errorf("Expected description '%s', got '%s'", tt.expected, description)
				}
			}
		})
	}
}

func TestGetChallengeDescription_InvalidMode(t *testing.T) {
	_, err := GetChallengeDescription("{}", "invalid_mode")
	if err == nil {
		t.Error("Expected error for invalid mode, got nil")
	}
}

func TestValidateAnswer(t *testing.T) {
	generator := NewCaptchaGenerator()

	t.Run("button", func(t *testing.T) {
		result, err := generator.GenerateChallenge("button")
		if err != nil {
			t.Fatalf("Failed to generate button challenge: %v", err)
		}

		// Test correct answer
		valid, err := ValidateAnswer(result.ChallengeData, "button", "button_click")
		if err != nil {
			t.Errorf("Unexpected error validating button answer: %v", err)
		}
		if !valid {
			t.Error("Expected button_click to be valid answer")
		}

		// Test incorrect answer
		valid, err = ValidateAnswer(result.ChallengeData, "button", "wrong_answer")
		if err != nil {
			t.Errorf("Unexpected error validating wrong button answer: %v", err)
		}
		if valid {
			t.Error("Expected wrong_answer to be invalid")
		}
	})

	t.Run("text", func(t *testing.T) {
		result, err := generator.GenerateChallenge("text")
		if err != nil {
			t.Fatalf("Failed to generate text challenge: %v", err)
		}

		var challenge TextChallenge
		err = json.Unmarshal([]byte(result.ChallengeData), &challenge)
		if err != nil {
			t.Fatalf("Failed to unmarshal text challenge: %v", err)
		}

		// Test correct answer
		correctAnswer := strconv.Itoa(challenge.Answer)
		valid, err := ValidateAnswer(result.ChallengeData, "text", correctAnswer)
		if err != nil {
			t.Errorf("Unexpected error validating text answer: %v", err)
		}
		if !valid {
			t.Error("Expected correct answer to be valid")
		}

		// Test incorrect answer
		valid, err = ValidateAnswer(result.ChallengeData, "text", "999")
		if err != nil {
			t.Errorf("Unexpected error validating wrong text answer: %v", err)
		}
		if valid {
			t.Error("Expected wrong answer to be invalid")
		}
	})

	t.Run("math", func(t *testing.T) {
		result, err := generator.GenerateChallenge("math")
		if err != nil {
			t.Fatalf("Failed to generate math challenge: %v", err)
		}

		// Test correct answer
		valid, err := ValidateAnswer(result.ChallengeData, "math", result.CorrectAnswer)
		if err != nil {
			t.Errorf("Unexpected error validating math answer: %v", err)
		}
		if !valid {
			t.Error("Expected correct answer to be valid")
		}

		// Test incorrect answer
		valid, err = ValidateAnswer(result.ChallengeData, "math", "999999")
		if err != nil {
			t.Errorf("Unexpected error validating wrong math answer: %v", err)
		}
		if valid {
			t.Error("Expected wrong answer to be invalid")
		}
	})

	t.Run("text2", func(t *testing.T) {
		result, err := generator.GenerateChallenge("text2")
		if err != nil {
			t.Fatalf("Failed to generate text2 challenge: %v", err)
		}

		// Test correct answer
		valid, err := ValidateAnswer(result.ChallengeData, "text2", result.CorrectAnswer)
		if err != nil {
			t.Errorf("Unexpected error validating text2 answer: %v", err)
		}
		if !valid {
			t.Error("Expected correct answer to be valid")
		}

		// Test incorrect answer
		valid, err = ValidateAnswer(result.ChallengeData, "text2", "WRONG")
		if err != nil {
			t.Errorf("Unexpected error validating wrong text2 answer: %v", err)
		}
		if valid {
			t.Error("Expected wrong answer to be invalid")
		}
	})
}

func TestValidateAnswer_InvalidMode(t *testing.T) {
	_, err := ValidateAnswer("{}", "invalid_mode", "answer")
	if err == nil {
		t.Error("Expected error for invalid mode, got nil")
	}
}

func TestValidateAnswer_InvalidJSON(t *testing.T) {
	_, err := ValidateAnswer("invalid json", "text", "0")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCaptchaGenerator_TextModeUniqueness(t *testing.T) {
	generator := NewCaptchaGenerator()

	// Generate multiple text captchas and ensure they're different
	results := make([]*CaptchaResult, 5)
	for i := 0; i < 5; i++ {
		result, err := generator.GenerateChallenge("text")
		if err != nil {
			t.Fatalf("Failed to generate text captcha %d: %v", i, err)
		}
		results[i] = result
	}

	// Check that images are different (simple byte length check)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if len(results[i].ImageBytes) == len(results[j].ImageBytes) {
				// This might be a coincidence, but let's check a few bytes
				equal := true
				checkLen := min(len(results[i].ImageBytes), 100) // Check first 100 bytes
				for k := 0; k < checkLen; k++ {
					if results[i].ImageBytes[k] != results[j].ImageBytes[k] {
						equal = false
						break
					}
				}
				if equal {
					t.Errorf("Generated captchas %d and %d appear to be identical", i, j)
				}
			}
		}
	}
}

func TestCaptchaGenerator_Text2ModeUniqueness(t *testing.T) {
	generator := NewCaptchaGenerator()

	// Generate multiple text2 captchas and ensure they're different
	results := make([]*CaptchaResult, 5)
	for i := 0; i < 5; i++ {
		result, err := generator.GenerateChallenge("text2")
		if err != nil {
			t.Fatalf("Failed to generate text2 captcha %d: %v", i, err)
		}
		results[i] = result
	}

	// Check that answers are different
	answers := make(map[string]bool)
	for i, result := range results {
		if answers[result.CorrectAnswer] {
			t.Errorf("Duplicate answer found in text2 captcha %d: %s", i, result.CorrectAnswer)
		}
		answers[result.CorrectAnswer] = true
	}
}

func TestImageValidation(t *testing.T) {
	generator := NewCaptchaGenerator()

	imageModes := []string{"text", "text2"}
	for _, mode := range imageModes {
		t.Run(mode, func(t *testing.T) {
			result, err := generator.GenerateChallenge(mode)
			if err != nil {
				t.Fatalf("Failed to generate %s challenge: %v", mode, err)
			}

			if result.ImageBytes == nil {
				t.Fatal("Expected image bytes, got nil")
			}

			// Decode and validate image
			img, err := png.Decode(bytes.NewReader(result.ImageBytes))
			if err != nil {
				t.Fatalf("Failed to decode image: %v", err)
			}

			bounds := img.Bounds()
			if bounds.Dx() == 0 || bounds.Dy() == 0 {
				t.Error("Image has zero dimensions")
			}

			// Basic sanity check - image should be reasonable size
			if bounds.Dx() < 50 || bounds.Dx() > 500 {
				t.Errorf("Image width %d seems unreasonable", bounds.Dx())
			}
			if bounds.Dy() < 30 || bounds.Dy() > 200 {
				t.Errorf("Image height %d seems unreasonable", bounds.Dy())
			}
		})
	}
}

// Benchmark tests
func BenchmarkCaptchaGenerator_GenerateButton(b *testing.B) {
	generator := NewCaptchaGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateChallenge("button")
		if err != nil {
			b.Fatalf("Failed to generate button challenge: %v", err)
		}
	}
}

func BenchmarkCaptchaGenerator_GenerateText(b *testing.B) {
	generator := NewCaptchaGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateChallenge("text")
		if err != nil {
			b.Fatalf("Failed to generate text challenge: %v", err)
		}
	}
}

func BenchmarkCaptchaGenerator_GenerateMath(b *testing.B) {
	generator := NewCaptchaGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateChallenge("math")
		if err != nil {
			b.Fatalf("Failed to generate math challenge: %v", err)
		}
	}
}

func BenchmarkCaptchaGenerator_GenerateText2(b *testing.B) {
	generator := NewCaptchaGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateChallenge("text2")
		if err != nil {
			b.Fatalf("Failed to generate text2 challenge: %v", err)
		}
	}
}

func BenchmarkValidateAnswer(b *testing.B) {
	generator := NewCaptchaGenerator()
	result, err := generator.GenerateChallenge("button")
	if err != nil {
		b.Fatalf("Failed to generate challenge: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateAnswer(result.ChallengeData, "button", "button_click")
		if err != nil {
			b.Fatalf("Failed to validate answer: %v", err)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
