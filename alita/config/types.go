package config

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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
	val, err := strconv.Atoi(t.str)
	if err != nil {
		log.Warnf("Failed to parse '%s' as int: %v. Defaulting to 0.", t.str, err)
	}
	return val
}

// Int64 converts the string to an int64.
// Returns 0 if conversion fails.
func (t typeConvertor) Int64() int64 {
	val, err := strconv.ParseInt(t.str, 10, 64)
	if err != nil {
		log.Warnf("Failed to parse '%s' as int64: %v. Defaulting to 0.", t.str, err)
	}
	return val
}

// Bool converts the string to a boolean using strconv.ParseBool which
// recognises a wider range of truthy/falsey values ("1", "t", "T", "true", "TRUE", "yes", etc.).
func (t typeConvertor) Bool() bool {
	v, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(t.str)))
	if err != nil {
		return false
	}
	return v
}
