package conversion

import (
	"errors"
	"strconv"
)

var (
	// ErrInvalidNumber represents an error when string cannot be converted to number
	ErrInvalidNumber = errors.New("invalid number format")

	// ErrNumberOutOfRange represents an error when number is out of valid range
	ErrNumberOutOfRange = errors.New("number out of range")
)

// SafeAtoi safely converts string to int with error handling
func SafeAtoi(s string) (int, error) {
	if s == "" {
		return 0, ErrInvalidNumber
	}

	result, err := strconv.Atoi(s)
	if err != nil {
		return 0, ErrInvalidNumber
	}

	return result, nil
}

// SafeAtoiWithDefault safely converts string to int with default value on error
func SafeAtoiWithDefault(s string, defaultValue int) int {
	result, err := SafeAtoi(s)
	if err != nil {
		return defaultValue
	}
	return result
}

// SafeParseInt64 safely converts string to int64 with error handling
func SafeParseInt64(s string) (int64, error) {
	if s == "" {
		return 0, ErrInvalidNumber
	}

	result, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrInvalidNumber
	}

	return result, nil
}

// SafeParseInt64WithDefault safely converts string to int64 with default value on error
func SafeParseInt64WithDefault(s string, defaultValue int64) int64 {
	result, err := SafeParseInt64(s)
	if err != nil {
		return defaultValue
	}
	return result
}

// SafeParseFloat64 safely converts string to float64 with error handling
func SafeParseFloat64(s string) (float64, error) {
	if s == "" {
		return 0, ErrInvalidNumber
	}

	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, ErrInvalidNumber
	}

	return result, nil
}

// SafeParseFloat64WithDefault safely converts string to float64 with default value on error
func SafeParseFloat64WithDefault(s string, defaultValue float64) float64 {
	result, err := SafeParseFloat64(s)
	if err != nil {
		return defaultValue
	}
	return result
}

// ValidateIntRange validates that an integer is within specified range
func ValidateIntRange(value, min, max int) error {
	if value < min || value > max {
		return ErrNumberOutOfRange
	}
	return nil
}

// ValidateInt64Range validates that an int64 is within specified range
func ValidateInt64Range(value, min, max int64) error {
	if value < min || value > max {
		return ErrNumberOutOfRange
	}
	return nil
}

// SafeAtoiWithRange safely converts string to int and validates range
func SafeAtoiWithRange(s string, min, max int) (int, error) {
	result, err := SafeAtoi(s)
	if err != nil {
		return 0, err
	}

	if err := ValidateIntRange(result, min, max); err != nil {
		return 0, err
	}

	return result, nil
}

// SafeParseInt64WithRange safely converts string to int64 and validates range
func SafeParseInt64WithRange(s string, min, max int64) (int64, error) {
	result, err := SafeParseInt64(s)
	if err != nil {
		return 0, err
	}

	if err := ValidateInt64Range(result, min, max); err != nil {
		return 0, err
	}

	return result, nil
}
