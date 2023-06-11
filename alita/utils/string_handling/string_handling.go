package string_handling

// FindInStringSlice Find takes a slice and looks for an element in it. If found it will
// return true, otherwise it will return a bool of false.
func FindInStringSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// FindInInt64Slice Find takes a slice and looks for an element in it. If found it will
// return true, otherwise it will return a bool of false.
func FindInInt64Slice(slice []int64, val int64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// RemoveFromInt64Slice Find takes a slice and looks for an element in it. If found it will
func RemoveFromInt64Slice(s []int64, r int64) []int64 {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// IsDuplicateInStringSlice Find takes a slice and looks for an element in it. If found it will
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

// FindIndexInt64 Find takes a slice and looks for an element in it. If found it will
// return true, otherwise it will return a bool of false.
func FindIndexInt64(chatIds []int64, chatId int64) int {
	for k, v := range chatIds {
		if chatId == v {
			return k
		}
	}
	return -1
}
