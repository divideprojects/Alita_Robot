package config

import (
	"strconv"
	"strings"
)

// typeConvertor provides methods to convert a string to various Go types.
// Used for parsing environment variables into usable config values.
type typeConvertor struct {
	str string
}

// StringArray returns a string slice from a comma-separated string.
// Whitespace is trimmed from each element.
func (t typeConvertor) StringArray() []string {
	allUpdates := strings.Split(t.str, ",")
	for i, j := range allUpdates {
		allUpdates[i] = strings.TrimSpace(j) // this will trim the whitespace
	}
	return allUpdates
}

// Int converts the string to an int.
// Returns 0 if conversion fails.
func (t typeConvertor) Int() int {
	val, _ := strconv.Atoi(t.str)
	return val
}

// Int64 converts the string to an int64.
// Returns 0 if conversion fails.
func (t typeConvertor) Int64() int64 {
	val, _ := strconv.ParseInt(t.str, 10, 64)
	return val
}

// Bool returns true if the string is "yes" or "true" (case-sensitive).
func (t typeConvertor) Bool() bool {
	return t.str == "yes" || t.str == "true"
}
