package captcha

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/mojocn/base64Captcha"
)

// ButtonChallenge represents a simple button-click CAPTCHA challenge.
// This is the simplest form of CAPTCHA where users just need to click a button to verify.
type ButtonChallenge struct {
	Type string `json:"type"`
}

// TextChallenge represents a text selection CAPTCHA challenge.
// Users must select the correct text option that matches a generated image.
type TextChallenge struct {
	Type    string   `json:"type"`
	Image   string   `json:"image"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

// MathChallenge represents a mathematical problem CAPTCHA challenge.
// Users must solve a basic arithmetic problem and select the correct answer.
type MathChallenge struct {
	Type     string `json:"type"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Options  []int  `json:"options"`
}

// Text2Challenge represents a character input CAPTCHA challenge.
// Users must identify and input characters shown in a distorted image.
type Text2Challenge struct {
	Type       string   `json:"type"`
	Image      string   `json:"image"`
	Characters []string `json:"characters"`
	Answer     string   `json:"answer"`
}

// CaptchaResult contains the complete challenge data and associated image.
// It includes the serialized challenge data, the correct answer, and optional image bytes for visual challenges.
type CaptchaResult struct {
	ChallengeData string
	CorrectAnswer string
	ImageBytes    []byte // nil for non-image modes like button and math
}

// CaptchaGenerator handles creation of different CAPTCHA challenges.
// It provides methods to generate various types of CAPTCHA challenges including button, text, math, and text2 modes.
type CaptchaGenerator struct{}

// NewCaptchaGenerator creates a new CAPTCHA generator instance.
// Returns a configured generator ready to create different types of CAPTCHA challenges.
func NewCaptchaGenerator() *CaptchaGenerator {
	return &CaptchaGenerator{}
}

// GenerateChallenge creates a CAPTCHA challenge based on the specified mode.
// Supported modes: "button", "text", "math", "text2".
// Returns a CaptchaResult containing the challenge data and any associated image bytes.
func (cg *CaptchaGenerator) GenerateChallenge(mode string) (*CaptchaResult, error) {
	switch mode {
	case "button":
		return cg.generateButtonChallenge()
	case "text":
		return cg.generateTextChallenge()
	case "math":
		return cg.generateMathChallenge()
	case "text2":
		return cg.generateText2Challenge()
	default:
		return nil, fmt.Errorf("unsupported CAPTCHA mode: %s", mode)
	}
}

// generateButtonChallenge creates a simple button click challenge.
// This is the most basic CAPTCHA type requiring only a button click for verification.
func (*CaptchaGenerator) generateButtonChallenge() (*CaptchaResult, error) {
	challenge := ButtonChallenge{
		Type: "button",
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return nil, err
	}

	return &CaptchaResult{
		ChallengeData: string(data),
		CorrectAnswer: "button_click",
		ImageBytes:    nil, // No image for button mode
	}, nil
}

// generateTextChallenge creates a text selection challenge with an actual generated image.
// It generates an image with a word and provides multiple choice options for the user to select.
func (*CaptchaGenerator) generateTextChallenge() (*CaptchaResult, error) {
	// Text options
	words := []string{
		"HOUSE", "TREE", "CAR", "BOOK", "PHONE", "LAMP", "DOOR", "WINDOW",
		"CHAIR", "TABLE", "PLANT", "CLOCK", "GLASS", "PENCIL", "PAPER", "MOUSE",
	}

	correctWord := words[rand.Intn(len(words))]

	// Generate 3 wrong options
	options := []string{correctWord}
	used := map[string]bool{correctWord: true}

	for len(options) < 4 {
		word := words[rand.Intn(len(words))]
		if !used[word] {
			options = append(options, word)
			used[word] = true
		}
	}

	// Shuffle options
	for i := range options {
		j := rand.Intn(i + 1)
		options[i], options[j] = options[j], options[i]
	}

	correctIndex := -1
	for i, option := range options {
		if option == correctWord {
			correctIndex = i
			break
		}
	}

	// Generate the actual image using base64Captcha with DrawCaptcha
	driver := base64Captcha.NewDriverString(80, 240, 0, base64Captcha.OptionShowHollowLine, 4, correctWord, nil, base64Captcha.DefaultEmbeddedFonts, nil)

	// Use DrawCaptcha directly with our chosen word
	item, err := driver.DrawCaptcha(correctWord)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text captcha image: %v", err)
	}

	// Get the base64 string
	b64s := item.EncodeB64string()

	// Decode base64 to bytes (remove data:image/png;base64, prefix if present)
	b64Data := strings.TrimPrefix(b64s, "data:image/png;base64,")
	imgBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode captcha image: %v", err)
	}

	challenge := TextChallenge{
		Type:    "text",
		Image:   fmt.Sprintf("text_image_%s", strings.ToLower(correctWord)),
		Options: options,
		Answer:  correctIndex,
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return nil, err
	}

	return &CaptchaResult{
		ChallengeData: string(data),
		CorrectAnswer: correctWord,
		ImageBytes:    imgBytes,
	}, nil
}

// generateMathChallenge creates a basic math question.
// It generates simple arithmetic problems (addition, subtraction, multiplication) with multiple choice answers.
func (*CaptchaGenerator) generateMathChallenge() (*CaptchaResult, error) {
	operations := []string{"+", "-", "*"}
	operation := operations[rand.Intn(len(operations))]

	var a, b, result int
	var question string

	switch operation {
	case "+":
		a = rand.Intn(50) + 1
		b = rand.Intn(50) + 1
		result = a + b
		question = fmt.Sprintf("%d + %d", a, b)
	case "-":
		a = rand.Intn(50) + 10
		b = rand.Intn(a)
		result = a - b
		question = fmt.Sprintf("%d - %d", a, b)
	case "*":
		a = rand.Intn(12) + 1
		b = rand.Intn(12) + 1
		result = a * b
		question = fmt.Sprintf("%d × %d", a, b)
	}

	// Generate wrong options
	options := []int{result}
	used := map[int]bool{result: true}

	for len(options) < 4 {
		// Generate plausible wrong answers
		var wrong int
		switch rand.Intn(3) {
		case 0:
			wrong = result + rand.Intn(10) + 1
		case 1:
			wrong = result - rand.Intn(10) - 1
			if wrong < 0 {
				wrong = result + rand.Intn(5) + 1
			}
		case 2:
			wrong = result + rand.Intn(20) - 10
			if wrong < 0 {
				wrong = result + rand.Intn(10) + 1
			}
		}

		if !used[wrong] && wrong >= 0 {
			options = append(options, wrong)
			used[wrong] = true
		}
	}

	// Shuffle options
	for i := range options {
		j := rand.Intn(i + 1)
		options[i], options[j] = options[j], options[i]
	}

	challenge := MathChallenge{
		Type:     "math",
		Question: question,
		Answer:   strconv.Itoa(result),
		Options:  options,
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return nil, err
	}

	return &CaptchaResult{
		ChallengeData: string(data),
		CorrectAnswer: strconv.Itoa(result),
		ImageBytes:    nil, // No image for math mode
	}, nil
}

// generateText2Challenge creates a character input challenge with an actual generated image.
// Users must identify and input the exact characters shown in a distorted image using an on-screen keyboard.
func (*CaptchaGenerator) generateText2Challenge() (*CaptchaResult, error) {
	// Character set for the code
	characters := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excluding confusing characters like I, O, 0, 1
	codeLength := 4 + rand.Intn(2)                   // Random length between 4-5

	// Generate random code
	correctCode := ""
	for i := 0; i < codeLength; i++ {
		correctCode += string(characters[rand.Intn(len(characters))])
	}

	// Generate the actual image using base64Captcha with DrawCaptcha
	driver := base64Captcha.NewDriverString(80, 240, 0, base64Captcha.OptionShowHollowLine|base64Captcha.OptionShowSlimeLine, codeLength, correctCode, nil, base64Captcha.DefaultEmbeddedFonts, nil)

	// Use DrawCaptcha directly with our chosen code
	item, err := driver.DrawCaptcha(correctCode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text2 captcha image: %v", err)
	}

	// Get the base64 string
	b64s := item.EncodeB64string()

	// Decode base64 to bytes (remove data:image/png;base64, prefix if present)
	b64Data := strings.TrimPrefix(b64s, "data:image/png;base64,")
	imgBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode captcha image: %v", err)
	}

	// Generate character options (all possible characters)
	allChars := make([]string, len(characters))
	for i, char := range characters {
		allChars[i] = string(char)
	}

	challenge := Text2Challenge{
		Type:       "text2",
		Image:      fmt.Sprintf("text2_image_%s", strings.ToLower(correctCode)),
		Characters: allChars,
		Answer:     correctCode,
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return nil, err
	}

	return &CaptchaResult{
		ChallengeData: string(data),
		CorrectAnswer: correctCode,
		ImageBytes:    imgBytes,
	}, nil
}

// CreateCaptchaKeyboard generates the appropriate inline keyboard for a CAPTCHA challenge.
// It creates mode-specific keyboards: simple button for "button" mode, multiple choice for "text"/"math", character grid for "text2".
func CreateCaptchaKeyboard(challengeData, mode string) (*gotgbot.InlineKeyboardMarkup, error) {
	switch mode {
	case "button":
		return &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "✅ Verify",
						CallbackData: "captcha_solve_button",
					},
				},
			},
		}, nil

	case "text":
		var challenge TextChallenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return nil, err
		}

		var keyboard [][]gotgbot.InlineKeyboardButton
		for i, option := range challenge.Options {
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				{
					Text:         option,
					CallbackData: fmt.Sprintf("captcha_solve_text_%d", i),
				},
			})
		}

		return &gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard}, nil

	case "math":
		var challenge MathChallenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return nil, err
		}

		var keyboard [][]gotgbot.InlineKeyboardButton
		for _, option := range challenge.Options {
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				{
					Text:         strconv.Itoa(option),
					CallbackData: fmt.Sprintf("captcha_solve_math_%d", option),
				},
			})
		}

		return &gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard}, nil

	case "text2":
		var challenge Text2Challenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return nil, err
		}

		// Create character selection keyboard (6 columns)
		var keyboard [][]gotgbot.InlineKeyboardButton
		var row []gotgbot.InlineKeyboardButton

		for i, char := range challenge.Characters {
			row = append(row, gotgbot.InlineKeyboardButton{
				Text:         char,
				CallbackData: fmt.Sprintf("captcha_char_%s", char),
			})

			if len(row) == 6 || i == len(challenge.Characters)-1 {
				keyboard = append(keyboard, row)
				row = []gotgbot.InlineKeyboardButton{}
			}
		}

		// Add control buttons
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			{
				Text:         "🔙 Delete",
				CallbackData: "captcha_text2_delete",
			},
			{
				Text:         "✅ Submit",
				CallbackData: "captcha_text2_submit",
			},
		})

		return &gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard}, nil

	default:
		return nil, fmt.Errorf("unsupported CAPTCHA mode for keyboard: %s", mode)
	}
}

// GetChallengeDescription returns a human-readable description of the challenge.
// It provides instructions to users on how to complete the specific CAPTCHA challenge type.
func GetChallengeDescription(challengeData, mode string) (string, error) {
	switch mode {
	case "button":
		return "Click the button below to verify you're human.", nil

	case "text":
		var challenge TextChallenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return "", err
		}
		return "Select the text that matches the image above.", nil

	case "math":
		var challenge MathChallenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return "", err
		}
		return fmt.Sprintf("Solve this math problem:\n<b>%s = ?</b>", challenge.Question), nil

	case "text2":
		var challenge Text2Challenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return "", err
		}
		return "Enter the characters you see in the image above, one by one.", nil

	default:
		return "", fmt.Errorf("unsupported CAPTCHA mode for description: %s", mode)
	}
}

// ValidateAnswer checks if the provided answer is correct for the challenge.
// It compares the user's answer against the correct answer stored in the challenge data.
func ValidateAnswer(challengeData, mode, userAnswer string) (bool, error) {
	switch mode {
	case "button":
		return userAnswer == "button_click", nil

	case "text", "math":
		// For text and math modes, compare the answer directly
		var challenge map[string]interface{}
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return false, err
		}

		if mode == "text" {
			answer, ok := challenge["answer"].(float64) // JSON numbers are float64
			if !ok {
				return false, fmt.Errorf("invalid answer format in challenge data")
			}
			return userAnswer == strconv.Itoa(int(answer)), nil
		} else { // math
			answerStr, ok := challenge["answer"].(string)
			if !ok {
				return false, fmt.Errorf("invalid answer format in challenge data")
			}
			return userAnswer == answerStr, nil
		}

	case "text2":
		var challenge Text2Challenge
		if err := json.Unmarshal([]byte(challengeData), &challenge); err != nil {
			return false, err
		}
		return userAnswer == challenge.Answer, nil

	default:
		return false, fmt.Errorf("unsupported CAPTCHA mode for validation: %s", mode)
	}
}
