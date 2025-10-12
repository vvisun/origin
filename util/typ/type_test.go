package typ

import (
	"testing"
)

func TestConvertToNumber_IntTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"int64", int64(42), 42},
		{"int", int(42), 42},
		{"int32", int32(42), 42},
		{"int16", int16(42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[int](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[int]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[int]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_UintTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected uint
	}{
		{"uint64", uint64(42), 42},
		{"uint", uint(42), 42},
		{"uint32", uint32(42), 42},
		{"uint16", uint16(42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[uint](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[uint]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[uint]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_FloatTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"float64", float64(42.5), 42.5},
		{"float32", float32(42.5), 42.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[float64](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[float64]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[float64]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_CrossTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"int64 to float64", int64(42), 42.0},
		{"int to float64", int(42), 42.0},
		{"uint64 to float64", uint64(42), 42.0},
		{"float32 to float64", float32(42.5), 42.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[float64](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[float64]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[float64]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_UnsupportedTypes(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"string", "hello"},
		{"bool", true},
		{"complex64", complex64(1 + 2i)},
		{"complex128", complex128(1 + 2i)},
		{"byte", byte('a')},
		{"int8", int8(42)},
		{"uint8", uint8(42)},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[int](tt.input)
			if err == nil {
				t.Errorf("ConvertToNumber[int]() expected error for %T, got result %v", tt.input, result)
			}
			if result != 0 {
				t.Errorf("ConvertToNumber[int]() expected 0 for unsupported type, got %v", result)
			}
		})
	}
}

func TestConvertToNumber_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"zero int64", int64(0), 0},
		{"negative int64", int64(-42), -42},
		{"max int64", int64(9223372036854775807), 9223372036854775807},
		{"min int64", int64(-9223372036854775808), -9223372036854775808},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[int64](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[int64]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[int64]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_FloatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"zero float64", float64(0), 0.0},
		{"negative float64", float64(-42.5), -42.5},
		{"large float64", float64(1.7976931348623157e+308), 1.7976931348623157e+308},
		{"small float64", float64(2.2250738585072014e-308), 2.2250738585072014e-308},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[float64](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[float64]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[float64]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_UintEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected uint64
	}{
		{"zero uint64", uint64(0), 0},
		{"max uint64", uint64(18446744073709551615), 18446744073709551615},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToNumber[uint64](tt.input)
			if err != nil {
				t.Errorf("ConvertToNumber[uint64]() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertToNumber[uint64]() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToNumber_PrecisionLoss(t *testing.T) {
	// Test conversion from float to int (precision loss)
	input := float64(42.7)
	result, err := ConvertToNumber[int](input)
	if err != nil {
		t.Errorf("ConvertToNumber[int]() error = %v", err)
		return
	}
	expected := 42 // Should truncate, not round
	if result != expected {
		t.Errorf("ConvertToNumber[int]() = %v, want %v (precision loss expected)", result, expected)
	}
}

func TestConvertToNumber_Rune(t *testing.T) {
	// Test rune which is an alias of int32
	result, err := ConvertToNumber[int32](rune('a'))
	if err != nil {
		t.Errorf("ConvertToNumber[int32]() error = %v", err)
		return
	}
	expected := int32(97)
	if result != expected {
		t.Errorf("ConvertToNumber[int32]() = %v, want %v", result, expected)
	}
}

func TestConvertToNumber_AllSupportedTypes(t *testing.T) {
	// Test all supported types to ensure none are missing
	supportedTypes := []interface{}{
		int64(42),
		int(42),
		uint(42),
		uint64(42),
		float32(42.5),
		float64(42.5),
		int32(42),
		uint32(42),
		int16(42),
		uint16(42),
	}

	for _, input := range supportedTypes {
		t.Run("supported_"+getTypeName(input), func(t *testing.T) {
			result, err := ConvertToNumber[float64](input)
			if err != nil {
				t.Errorf("ConvertToNumber[float64]() error = %v for type %T", err, input)
			}
			if result == 0 && input != 0 {
				t.Errorf("ConvertToNumber[float64]() returned 0 for non-zero input %v of type %T", input, input)
			}
		})
	}
}

// Helper function to get type name for test naming
func getTypeName(v interface{}) string {
	switch v.(type) {
	case int64:
		return "int64"
	case int:
		return "int"
	case uint:
		return "uint"
	case uint64:
		return "uint64"
	case float32:
		return "float32"
	case float64:
		return "float64"
	case int32:
		return "int32"
	case uint32:
		return "uint32"
	case int16:
		return "int16"
	case uint16:
		return "uint16"
	default:
		return "unknown"
	}
}
