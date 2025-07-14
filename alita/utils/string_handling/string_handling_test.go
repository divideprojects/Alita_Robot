package string_handling

import (
	"strconv"
	"testing"
)

// Benchmark data
var (
	testStringSlice = []string{
		"admin", "ban", "kick", "mute", "warn", "filter", "blacklist", "note", "rule", "lock",
		"user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8", "user9", "user10",
		"command1", "command2", "command3", "command4", "command5", "command6", "command7", "command8", "command9", "command10",
	}

	testInt64Slice = []int64{
		123456789, 987654321, 111111111, 222222222, 333333333, 444444444, 555555555, 666666666, 777777777, 888888888,
		100, 200, 300, 400, 500, 600, 700, 800, 900, 1000,
		1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010,
	}

	searchStrings = []string{"admin", "user5", "command8", "nonexistent"}
	searchInt64s  = []int64{123456789, 555555555, 1008, 9999999999}
)

// Generate large test data for stress testing
func generateLargeStringSlice(size int) []string {
	slice := make([]string, size)
	for i := 0; i < size; i++ {
		slice[i] = "item" + strconv.Itoa(i)
	}
	return slice
}

func generateLargeInt64Slice(size int) []int64 {
	slice := make([]int64, size)
	for i := 0; i < size; i++ {
		slice[i] = int64(i)
	}
	return slice
}

// Benchmark original linear search functions
func BenchmarkFindInStringSlice_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, search := range searchStrings {
			FindInStringSlice(testStringSlice, search)
		}
	}
}

func BenchmarkFindInInt64Slice_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, search := range searchInt64s {
			FindInInt64Slice(testInt64Slice, search)
		}
	}
}

// Benchmark optimized map-based functions
func BenchmarkFindInStringMap_Small(b *testing.B) {
	stringMap := StringSliceToMap(testStringSlice)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, search := range searchStrings {
			FindInStringMap(stringMap, search)
		}
	}
}

func BenchmarkFindInInt64Map_Small(b *testing.B) {
	int64Map := Int64SliceToMap(testInt64Slice)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, search := range searchInt64s {
			FindInInt64Map(int64Map, search)
		}
	}
}

// Benchmark optimized lookup structs
func BenchmarkOptimizedStringLookup_Small(b *testing.B) {
	lookup := NewOptimizedStringLookup(testStringSlice)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, search := range searchStrings {
			lookup.Contains(search)
		}
	}
}

func BenchmarkOptimizedInt64Lookup_Small(b *testing.B) {
	lookup := NewOptimizedInt64Lookup(testInt64Slice)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, search := range searchInt64s {
			lookup.Contains(search)
		}
	}
}

// Large dataset benchmarks (1000 items)
func BenchmarkFindInStringSlice_Large(b *testing.B) {
	largeSlice := generateLargeStringSlice(1000)
	searchItems := []string{"item100", "item500", "item999", "nonexistent"}

	for i := 0; i < b.N; i++ {
		for _, search := range searchItems {
			FindInStringSlice(largeSlice, search)
		}
	}
}

func BenchmarkFindInStringMap_Large(b *testing.B) {
	largeSlice := generateLargeStringSlice(1000)
	stringMap := StringSliceToMap(largeSlice)
	searchItems := []string{"item100", "item500", "item999", "nonexistent"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, search := range searchItems {
			FindInStringMap(stringMap, search)
		}
	}
}

func BenchmarkFindInInt64Slice_Large(b *testing.B) {
	largeSlice := generateLargeInt64Slice(1000)
	searchItems := []int64{100, 500, 999, 9999}

	for i := 0; i < b.N; i++ {
		for _, search := range searchItems {
			FindInInt64Slice(largeSlice, search)
		}
	}
}

func BenchmarkFindInInt64Map_Large(b *testing.B) {
	largeSlice := generateLargeInt64Slice(1000)
	int64Map := Int64SliceToMap(largeSlice)
	searchItems := []int64{100, 500, 999, 9999}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, search := range searchItems {
			FindInInt64Map(int64Map, search)
		}
	}
}

// Conversion overhead benchmarks
func BenchmarkStringSliceToMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringSliceToMap(testStringSlice)
	}
}

func BenchmarkInt64SliceToMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int64SliceToMap(testInt64Slice)
	}
}

// Multiple lookup scenarios (simulating real usage)
func BenchmarkMultipleLookups_Linear(b *testing.B) {
	largeSlice := generateLargeStringSlice(100)

	for i := 0; i < b.N; i++ {
		// Simulate 10 lookups (typical for admin checks)
		for j := 0; j < 10; j++ {
			FindInStringSlice(largeSlice, "item"+strconv.Itoa(j*10))
		}
	}
}

func BenchmarkMultipleLookups_Map(b *testing.B) {
	largeSlice := generateLargeStringSlice(100)
	stringMap := StringSliceToMap(largeSlice)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate 10 lookups (typical for admin checks)
		for j := 0; j < 10; j++ {
			FindInStringMap(stringMap, "item"+strconv.Itoa(j*10))
		}
	}
}

// Real-world simulation: Admin checking scenario
func BenchmarkAdminCheck_Linear(b *testing.B) {
	// Simulate typical admin list (10-20 admins)
	adminIds := generateLargeInt64Slice(15)

	for i := 0; i < b.N; i++ {
		// Check if user is admin (worst case - user not found)
		FindInInt64Slice(adminIds, 999999)
	}
}

func BenchmarkAdminCheck_Map(b *testing.B) {
	// Simulate typical admin list (10-20 admins)
	adminIds := generateLargeInt64Slice(15)
	adminMap := Int64SliceToMap(adminIds)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Check if user is admin (worst case - user not found)
		FindInInt64Map(adminMap, 999999)
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocation_StringMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		StringSliceToMap(testStringSlice)
	}
}

func BenchmarkMemoryAllocation_Int64Map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Int64SliceToMap(testInt64Slice)
	}
}

// Test functions to verify correctness
func TestOptimizedFunctions(t *testing.T) {
	// Test string functions
	stringMap := StringSliceToMap(testStringSlice)

	for _, item := range testStringSlice {
		if !FindInStringMap(stringMap, item) {
			t.Errorf("FindInStringMap failed for item: %s", item)
		}
	}

	if FindInStringMap(stringMap, "nonexistent") {
		t.Error("FindInStringMap should return false for nonexistent item")
	}

	// Test int64 functions
	int64Map := Int64SliceToMap(testInt64Slice)

	for _, item := range testInt64Slice {
		if !FindInInt64Map(int64Map, item) {
			t.Errorf("FindInInt64Map failed for item: %d", item)
		}
	}

	if FindInInt64Map(int64Map, 999999999) {
		t.Error("FindInInt64Map should return false for nonexistent item")
	}

	// Test optimized lookup structs
	stringLookup := NewOptimizedStringLookup(testStringSlice)
	for _, item := range testStringSlice {
		if !stringLookup.Contains(item) {
			t.Errorf("OptimizedStringLookup.Contains failed for item: %s", item)
		}
	}

	int64Lookup := NewOptimizedInt64Lookup(testInt64Slice)
	for _, item := range testInt64Slice {
		if !int64Lookup.Contains(item) {
			t.Errorf("OptimizedInt64Lookup.Contains failed for item: %d", item)
		}
	}
}
