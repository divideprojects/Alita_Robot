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

// Challenge data structures for different CAPTCHA modes
type ButtonChallenge struct {
	Type string `json:"type"`
}

type TextChallenge struct {
	Type    string   `json:"type"`
	Image   string   `json:"image"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

type MathChallenge struct {
	Type     string `json:"type"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Options  []int  `json:"options"`
}

type Text2Challenge struct {
	Type       string   `json:"type"`
	Image      string   `json:"image"`
	Characters []string `json:"characters"`
	Answer     string   `json:"answer"`
}

// CaptchaResult contains both the challenge data and image bytes (if applicable)
type CaptchaResult struct {
	ChallengeData string
	CorrectAnswer string
	ImageBytes    []byte // nil for non-image modes like button and math
}

// CaptchaGenerator handles creation of different CAPTCHA challenges
type CaptchaGenerator struct{}

// NewCaptchaGenerator creates a new CAPTCHA generator instance
func NewCaptchaGenerator() *CaptchaGenerator {
	return &CaptchaGenerator{}
}

// GenerateChallenge creates a CAPTCHA challenge based on the specified mode
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

// generateButtonChallenge creates a simple button click challenge
func (cg *CaptchaGenerator) generateButtonChallenge() (*CaptchaResult, error) {
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

// generateTextChallenge creates a text selection challenge with an actual generated image
func (cg *CaptchaGenerator) generateTextChallenge() (*CaptchaResult, error) {
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

// generateMathChallenge creates a basic math question
func (cg *CaptchaGenerator) generateMathChallenge() (*CaptchaResult, error) {
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

// generateText2Challenge creates a character input challenge with an actual generated image
func (cg *CaptchaGenerator) generateText2Challenge() (*CaptchaResult, error) {
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

// CreateCaptchaKeyboard generates the appropriate inline keyboard for a CAPTCHA challenge
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

// GetChallengeDescription returns a human-readable description of the challenge
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

// ValidateAnswer checks if the provided answer is correct for the challenge
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
