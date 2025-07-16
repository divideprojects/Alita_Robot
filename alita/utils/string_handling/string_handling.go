// Package string_handling provides utility functions for string and slice operations.
//
// This package contains optimized functions for common operations on string slices,
// int64 slices, and string manipulation tasks used throughout the Alita Robot codebase.
package string_handling

// FindInStringSlice returns true if the given value exists in the string slice.
//
// Performs a linear search for the value.
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

// ===== OPTIMIZED MAP-BASED FUNCTIONS =====

/*
StringSliceToMap converts a string slice to a map for O(1) lookups.

Returns a map where keys are the slice elements and values are true.
Use this when you need to perform multiple lookups on the same slice.
*/
func StringSliceToMap(slice []string) map[string]bool {
	result := make(map[string]bool, len(slice))
	for _, item := range slice {
		result[item] = true
	}
	return result
}

/*
Int64SliceToMap converts an int64 slice to a map for O(1) lookups.

Returns a map where keys are the slice elements and values are true.
Use this when you need to perform multiple lookups on the same slice.
*/
func Int64SliceToMap(slice []int64) map[int64]bool {
	result := make(map[int64]bool, len(slice))
	for _, item := range slice {
		result[item] = true
	}
	return result
}

/*
FindInStringMap returns true if the given value exists in the string map.

O(1) lookup performance. Use StringSliceToMap() to convert slices first.
*/
func FindInStringMap(m map[string]bool, val string) bool {
	return m[val]
}

/*
FindInInt64Map returns true if the given int64 value exists in the map.

O(1) lookup performance. Use Int64SliceToMap() to convert slices first.
*/
func FindInInt64Map(m map[int64]bool, val int64) bool {
	return m[val]
}

/*
OptimizedStringLookup provides O(1) string lookups for frequently accessed slices.

Create once, use many times for optimal performance.
*/
type OptimizedStringLookup struct {
	lookupMap map[string]bool
}

/*
NewOptimizedStringLookup creates a new optimized string lookup from a slice.

Use this when you need to perform many lookups on the same set of strings.
*/
func NewOptimizedStringLookup(slice []string) *OptimizedStringLookup {
	return &OptimizedStringLookup{
		lookupMap: StringSliceToMap(slice),
	}
}

/*
Contains returns true if the value exists in the lookup set.

O(1) performance.
*/
func (o *OptimizedStringLookup) Contains(val string) bool {
	return o.lookupMap[val]
}

/*
Update refreshes the lookup set with new values.

Call this when the underlying data changes.
*/
func (o *OptimizedStringLookup) Update(slice []string) {
	o.lookupMap = StringSliceToMap(slice)
}

/*
OptimizedInt64Lookup provides O(1) int64 lookups for frequently accessed slices.

Create once, use many times for optimal performance.
*/
type OptimizedInt64Lookup struct {
	lookupMap map[int64]bool
}

/*
NewOptimizedInt64Lookup creates a new optimized int64 lookup from a slice.

Use this when you need to perform many lookups on the same set of int64 values.
*/
func NewOptimizedInt64Lookup(slice []int64) *OptimizedInt64Lookup {
	return &OptimizedInt64Lookup{
		lookupMap: Int64SliceToMap(slice),
	}
}

/*
Contains returns true if the value exists in the lookup set.

O(1) performance.
*/
func (o *OptimizedInt64Lookup) Contains(val int64) bool {
	return o.lookupMap[val]
}

/*
Update refreshes the lookup set with new values.

Call this when the underlying data changes.
*/
func (o *OptimizedInt64Lookup) Update(slice []int64) {
	o.lookupMap = Int64SliceToMap(slice)
}
