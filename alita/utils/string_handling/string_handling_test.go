package string_handling

import (
	"reflect"
	"testing"
)

func TestFindInStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		val      string
		expected bool
	}{
		{"Found in slice", []string{"apple", "banana", "cherry"}, "banana", true},
		{"Not found in slice", []string{"apple", "banana", "cherry"}, "orange", false},
		{"Empty slice", []string{}, "apple", false},
		{"Single item found", []string{"apple"}, "apple", true},
		{"Single item not found", []string{"apple"}, "banana", false},
		{"Case sensitive", []string{"Apple"}, "apple", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindInStringSlice(tt.slice, tt.val)
			if result != tt.expected {
				t.Errorf("FindInStringSlice(%v, %q) = %v, want %v", tt.slice, tt.val, result, tt.expected)
			}
		})
	}
}

func TestFindInInt64Slice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int64
		val      int64
		expected bool
	}{
		{"Found in slice", []int64{1, 2, 3, 4, 5}, 3, true},
		{"Not found in slice", []int64{1, 2, 3, 4, 5}, 6, false},
		{"Empty slice", []int64{}, 1, false},
		{"Single item found", []int64{42}, 42, true},
		{"Single item not found", []int64{42}, 43, false},
		{"Negative numbers", []int64{-1, -2, -3}, -2, true},
		{"Zero value", []int64{0, 1, 2}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindInInt64Slice(tt.slice, tt.val)
			if result != tt.expected {
				t.Errorf("FindInInt64Slice(%v, %d) = %v, want %v", tt.slice, tt.val, result, tt.expected)
			}
		})
	}
}

func TestRemoveFromInt64Slice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int64
		remove   int64
		expected []int64
	}{
		{"Remove from middle", []int64{1, 2, 3, 4, 5}, 3, []int64{1, 2, 4, 5}},
		{"Remove first element", []int64{1, 2, 3}, 1, []int64{2, 3}},
		{"Remove last element", []int64{1, 2, 3}, 3, []int64{1, 2}},
		{"Remove non-existent", []int64{1, 2, 3}, 4, []int64{1, 2, 3}},
		{"Remove from single element", []int64{42}, 42, []int64{}},
		{"Remove from empty slice", []int64{}, 1, []int64{}},
		{"Remove duplicate (first occurrence)", []int64{1, 2, 2, 3}, 2, []int64{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveFromInt64Slice(tt.slice, tt.remove)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveFromInt64Slice(%v, %d) = %v, want %v", tt.slice, tt.remove, result, tt.expected)
			}
		})
	}
}

func TestIsDuplicateInStringSlice(t *testing.T) {
	tests := []struct {
		name           string
		slice          []string
		expectedDup    string
		expectedExists bool
	}{
		{"No duplicates", []string{"apple", "banana", "cherry"}, "", false},
		{"Has duplicates", []string{"apple", "banana", "apple"}, "apple", true},
		{"Multiple duplicates", []string{"a", "b", "a", "c", "b"}, "a", true},
		{"Empty slice", []string{}, "", false},
		{"Single element", []string{"apple"}, "", false},
		{"All same elements", []string{"same", "same", "same"}, "same", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup, exists := IsDuplicateInStringSlice(tt.slice)
			if exists != tt.expectedExists {
				t.Errorf("IsDuplicateInStringSlice(%v) exists = %v, want %v", tt.slice, exists, tt.expectedExists)
			}
			if exists && dup != tt.expectedDup {
				t.Errorf("IsDuplicateInStringSlice(%v) duplicate = %q, want %q", tt.slice, dup, tt.expectedDup)
			}
		})
	}
}
