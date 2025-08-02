package conversion

import (
	"testing"
)

func TestSafeAtoi(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"Valid positive number", "123", 123, false},
		{"Valid negative number", "-456", -456, false},
		{"Valid zero", "0", 0, false},
		{"Empty string", "", 0, true},
		{"Invalid string", "abc", 0, true},
		{"Mixed string", "123abc", 0, true},
		{"Float string", "123.45", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeAtoi(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeAtoi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeAtoi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeAtoiWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		want         int
	}{
		{"Valid number", "123", 999, 123},
		{"Invalid string", "abc", 999, 999},
		{"Empty string", "", 999, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeAtoiWithDefault(tt.input, tt.defaultValue)
			if got != tt.want {
				t.Errorf("SafeAtoiWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeParseInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"Valid positive number", "123456789", 123456789, false},
		{"Valid negative number", "-987654321", -987654321, false},
		{"Valid zero", "0", 0, false},
		{"Empty string", "", 0, true},
		{"Invalid string", "abc", 0, true},
		{"Large number", "9223372036854775807", 9223372036854775807, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeParseInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeParseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeParseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		min     int
		max     int
		wantErr bool
	}{
		{"Valid range", 5, 1, 10, false},
		{"At minimum", 1, 1, 10, false},
		{"At maximum", 10, 1, 10, false},
		{"Below minimum", 0, 1, 10, true},
		{"Above maximum", 11, 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntRange(tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSafeAtoiWithRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		min     int
		max     int
		want    int
		wantErr bool
	}{
		{"Valid in range", "5", 1, 10, 5, false},
		{"Invalid string", "abc", 1, 10, 0, true},
		{"Out of range low", "0", 1, 10, 0, true},
		{"Out of range high", "11", 1, 10, 0, true},
		{"At boundary min", "1", 1, 10, 1, false},
		{"At boundary max", "10", 1, 10, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeAtoiWithRange(tt.input, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeAtoiWithRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeAtoiWithRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSafeAtoi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SafeAtoi("12345")
	}
}

func BenchmarkSafeParseInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SafeParseInt64("1234567890")
	}
}
