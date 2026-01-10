package viewer_test

import (
	"testing"

	"github.com/nnnkkk7/memtui/viewer"
)

func TestDetectType_JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected viewer.DataType
	}{
		{"valid object", []byte(`{"key": "value"}`), viewer.DataTypeJSON},
		{"valid array", []byte(`[1, 2, 3]`), viewer.DataTypeJSON},
		{"with whitespace", []byte(`  {"key": "value"}  `), viewer.DataTypeJSON},
		{"nested object", []byte(`{"user": {"name": "test"}}`), viewer.DataTypeJSON},
		{"invalid json", []byte(`{key: value}`), viewer.DataTypeText},
		{"partial json", []byte(`{"key":`), viewer.DataTypeText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.DetectType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectType_Binary(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected viewer.DataType
	}{
		{"null bytes", []byte{0x00, 0x01, 0x02, 0x03}, viewer.DataTypeBinary},
		{"high bytes", []byte{0xFF, 0xFE, 0xFD}, viewer.DataTypeBinary},
		{"control chars", []byte{0x01, 0x02, 0x03, 0x04}, viewer.DataTypeBinary},
		{"mixed with text", []byte("hello\x00world"), viewer.DataTypeBinary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.DetectType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectType_Gzip(t *testing.T) {
	// Gzip magic bytes: 0x1f, 0x8b
	gzipData := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00}
	result := viewer.DetectType(gzipData)
	if result != viewer.DataTypeCompressedGzip {
		t.Errorf("expected DataTypeCompressedGzip, got %v", result)
	}
}

func TestDetectType_Zlib(t *testing.T) {
	// Zlib magic bytes: 0x78 (0x9c or 0xda for different compression levels)
	zlibData := []byte{0x78, 0x9c, 0x00, 0x00, 0x00, 0x00}
	result := viewer.DetectType(zlibData)
	if result != viewer.DataTypeCompressedZlib {
		t.Errorf("expected DataTypeCompressedZlib, got %v", result)
	}
}

func TestDetectType_Text(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected viewer.DataType
	}{
		{"plain text", []byte("hello world"), viewer.DataTypeText},
		{"with newlines", []byte("line1\nline2\nline3"), viewer.DataTypeText},
		{"utf8", []byte("Hello 🌍 World"), viewer.DataTypeText},
		{"numbers", []byte("12345"), viewer.DataTypeText},
		{"empty", []byte(""), viewer.DataTypeText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.DetectType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDataType_String(t *testing.T) {
	tests := []struct {
		dt       viewer.DataType
		expected string
	}{
		{viewer.DataTypeJSON, "JSON"},
		{viewer.DataTypeBinary, "Binary"},
		{viewer.DataTypeText, "Text"},
		{viewer.DataTypeCompressedGzip, "Gzip"},
		{viewer.DataTypeCompressedZlib, "Zlib"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.dt.String() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.dt.String())
			}
		})
	}
}

func TestDetectType_ShortData(t *testing.T) {
	// Test edge cases with very short data
	tests := []struct {
		name  string
		input []byte
	}{
		{"single byte", []byte{0x41}},
		{"two bytes", []byte{0x41, 0x42}},
		{"empty", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = viewer.DetectType(tt.input)
		})
	}
}
