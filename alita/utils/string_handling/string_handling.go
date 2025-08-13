package string_handling

import "slices"

// FindInStringSlice searches for a string value in a string slice.
// Returns true if the value is found, false otherwise.
func FindInStringSlice(slice []string, val string) bool {
	return slices.Contains(slice, val)
}

// FindInInt64Slice searches for an int64 value in an int64 slice.
// Returns true if the value is found, false otherwise.
func FindInInt64Slice(slice []int64, val int64) bool {
	return slices.Contains(slice, val)
}

// IsDuplicateInStringSlice checks for duplicate strings in a string slice.
// Returns the first duplicate found and true, or empty string and false if no duplicates.
func IsDuplicateInStringSlice(arr []string) (string, bool) {
	visited := make(map[string]bool)
	for i := range arr {
		if visited[arr[i]] {
			return arr[i], true
		} else {
			visited[arr[i]] = true
		}
	}
	return "", false
}
