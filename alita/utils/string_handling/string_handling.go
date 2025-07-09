package string_handling

/*
FindInStringSlice returns true if the given value exists in the string slice.

Performs a linear search for the value.
*/
func FindInStringSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

/*
FindInInt64Slice returns true if the given int64 value exists in the slice.

Performs a linear search for the value.
*/
func FindInInt64Slice(slice []int64, val int64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

/*
RemoveFromInt64Slice removes the first occurrence of the given int64 value from the slice.

Returns a new slice with the value removed, or the original slice if not found.
*/
func RemoveFromInt64Slice(s []int64, r int64) []int64 {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

/*
IsDuplicateInStringSlice checks for duplicate strings in the slice.

Returns the first duplicate found and true, or an empty string and false if no duplicates exist.
*/
func IsDuplicateInStringSlice(arr []string) (string, bool) {
	visited := make(map[string]bool)
	for i := 0; i < len(arr); i++ {
		if visited[arr[i]] {
			return arr[i], true
		} else {
			visited[arr[i]] = true
		}
	}
	return "", false
}

/*
FindIndexInt64 returns the index of the given int64 value in the slice.

Returns -1 if the value is not found.
*/
func FindIndexInt64(chatIds []int64, chatId int64) int {
	for k, v := range chatIds {
		if chatId == v {
			return k
		}
	}
	return -1
}
