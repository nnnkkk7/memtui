package viewer_test

import (
	"strings"
	"testing"

	"github.com/nnnkkk7/memtui/viewer"
)

func TestJSONFormatter_Format(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantErr  bool
		contains []string // Substrings that should be in output
	}{
		{
			name:     "simple object",
			input:    []byte(`{"key":"value"}`),
			wantErr:  false,
			contains: []string{"key", "value"},
		},
		{
			name:     "nested object",
			input:    []byte(`{"user":{"name":"test","age":30}}`),
			wantErr:  false,
			contains: []string{"user", "name", "test", "age", "30"},
		},
		{
			name:     "array",
			input:    []byte(`[1,2,3]`),
			wantErr:  false,
			contains: []string{"1", "2", "3"},
		},
		{
			name:    "invalid json",
			input:   []byte(`{invalid}`),
			wantErr: true,
		},
	}

	f := viewer.NewJSONFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Format(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("result should contain '%s', got: %s", s, result)
				}
			}
		})
	}
}

func TestHexFormatter_Format(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		contains []string
	}{
		{
			name:     "simple bytes",
			input:    []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, // "Hello"
			contains: []string{"48", "65", "6c", "6c", "6f", "Hello"},
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0xFF, 0x0A, 0x0D},
			contains: []string{"00", "ff", "0a", "0d"},
		},
		{
			name:     "16+ bytes",
			input:    []byte("0123456789abcdef0123"),
			contains: []string{"30", "31", "32"}, // ASCII for "012"
		},
		{
			name:     "empty",
			input:    []byte{},
			contains: []string{},
		},
	}

	f := viewer.NewHexFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Format(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			for _, s := range tt.contains {
				if !strings.Contains(strings.ToLower(result), strings.ToLower(s)) {
					t.Errorf("result should contain '%s', got: %s", s, result)
				}
			}
		})
	}
}

func TestTextFormatter_Format(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple text",
			input:    []byte("hello world"),
			expected: "hello world",
		},
		{
			name:     "with newlines",
			input:    []byte("line1\nline2"),
			expected: "line1\nline2",
		},
		{
			name:     "utf8",
			input:    []byte("Café résumé naïve"),
			expected: "Café résumé naïve",
		},
	}

	f := viewer.NewTextFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Format(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAutoFormatter(t *testing.T) {
	f := viewer.NewAutoFormatter()

	// JSON should be pretty-printed
	jsonInput := []byte(`{"key":"value"}`)
	result, err := f.Format(jsonInput)
	if err != nil {
		t.Errorf("unexpected error for JSON: %v", err)
	}
	if !strings.Contains(result, "key") {
		t.Error("JSON should be formatted")
	}

	// Binary should be hex formatted
	binaryInput := []byte{0x00, 0xFF, 0x0A}
	result, err = f.Format(binaryInput)
	if err != nil {
		t.Errorf("unexpected error for binary: %v", err)
	}
	if !strings.Contains(strings.ToLower(result), "00") {
		t.Error("Binary should be hex formatted")
	}

	// Text should be plain
	textInput := []byte("plain text")
	result, err = f.Format(textInput)
	if err != nil {
		t.Errorf("unexpected error for text: %v", err)
	}
	if result != "plain text" {
		t.Errorf("expected 'plain text', got '%s'", result)
	}
}
