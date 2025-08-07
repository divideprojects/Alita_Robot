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

// RemoveFromInt64Slice removes the first occurrence of a value from an int64 slice.
// Returns a new slice with the element removed, or the original slice if not found.
func RemoveFromInt64Slice(s []int64, r int64) []int64 {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
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

// FindIndexInt64 finds the index of an int64 value in an int64 slice.
// Returns the index if found, or -1 if the value is not in the slice.
func FindIndexInt64(chatIds []int64, chatId int64) int {
	for k, v := range chatIds {
		if chatId == v {
			return k
		}
	}
	return -1
}
