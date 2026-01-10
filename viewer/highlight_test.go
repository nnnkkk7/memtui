package viewer_test

import (
	"strings"
	"testing"

	"github.com/nnnkkk7/memtui/viewer"
)

func TestHighlightJSON_SimpleObject(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `{"name": "John"}`

	result := h.Highlight(input)

	// Should contain the key "name" and value "John"
	if !strings.Contains(result, "name") {
		t.Error("result should contain key 'name'")
	}
	if !strings.Contains(result, "John") {
		t.Error("result should contain value 'John'")
	}
	// Should preserve structure
	if !strings.Contains(result, "{") || !strings.Contains(result, "}") {
		t.Error("result should preserve braces")
	}
	if !strings.Contains(result, ":") {
		t.Error("result should preserve colon")
	}
}

func TestHighlightJSON_Array(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `[1, 2, 3]`

	result := h.Highlight(input)

	// Should contain the array values
	if !strings.Contains(result, "1") {
		t.Error("result should contain '1'")
	}
	if !strings.Contains(result, "2") {
		t.Error("result should contain '2'")
	}
	if !strings.Contains(result, "3") {
		t.Error("result should contain '3'")
	}
	// Should preserve brackets
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Error("result should preserve brackets")
	}
}

func TestHighlightJSON_NestedObject(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `{
  "user": {
    "name": "Alice",
    "age": 30
  }
}`

	result := h.Highlight(input)

	// Should contain nested keys and values
	if !strings.Contains(result, "user") {
		t.Error("result should contain 'user'")
	}
	if !strings.Contains(result, "name") {
		t.Error("result should contain 'name'")
	}
	if !strings.Contains(result, "Alice") {
		t.Error("result should contain 'Alice'")
	}
	if !strings.Contains(result, "age") {
		t.Error("result should contain 'age'")
	}
	if !strings.Contains(result, "30") {
		t.Error("result should contain '30'")
	}
	// Should preserve newlines (whitespace preservation)
	if !strings.Contains(result, "\n") {
		t.Error("result should preserve newlines")
	}
}

func TestHighlightJSON_BooleanValues(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `{"active": true, "deleted": false}`

	result := h.Highlight(input)

	// Should contain boolean values
	if !strings.Contains(result, "true") {
		t.Error("result should contain 'true'")
	}
	if !strings.Contains(result, "false") {
		t.Error("result should contain 'false'")
	}
	// Should contain keys
	if !strings.Contains(result, "active") {
		t.Error("result should contain 'active'")
	}
	if !strings.Contains(result, "deleted") {
		t.Error("result should contain 'deleted'")
	}
}

func TestHighlightJSON_NullValue(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `{"data": null}`

	result := h.Highlight(input)

	// Should contain null value
	if !strings.Contains(result, "null") {
		t.Error("result should contain 'null'")
	}
	if !strings.Contains(result, "data") {
		t.Error("result should contain 'data'")
	}
}

func TestHighlightJSON_StringWithEscapes(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "escaped quote",
			input:    `{"text": "Hello \"World\""}`,
			contains: `\"World\"`,
		},
		{
			name:     "escaped backslash",
			input:    `{"path": "C:\\Users\\test"}`,
			contains: `C:\\Users\\test`,
		},
		{
			name:     "escaped newline",
			input:    `{"text": "line1\nline2"}`,
			contains: `\n`,
		},
		{
			name:     "escaped tab",
			input:    `{"text": "col1\tcol2"}`,
			contains: `\t`,
		},
		{
			name:     "unicode escape",
			input:    `{"emoji": "\u0048\u0065\u006c\u006c\u006f"}`,
			contains: `\u0048`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain '%s', got: %s", tt.contains, result)
			}
		})
	}
}

func TestHighlightJSON_Numbers(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "integer",
			input:    `{"count": 42}`,
			contains: []string{"42", "count"},
		},
		{
			name:     "float",
			input:    `{"price": 19.99}`,
			contains: []string{"19.99", "price"},
		},
		{
			name:     "negative integer",
			input:    `{"balance": -100}`,
			contains: []string{"-100", "balance"},
		},
		{
			name:     "negative float",
			input:    `{"temp": -3.5}`,
			contains: []string{"-3.5", "temp"},
		},
		{
			name:     "scientific notation",
			input:    `{"large": 1.5e10}`,
			contains: []string{"1.5e10", "large"},
		},
		{
			name:     "scientific notation negative exponent",
			input:    `{"small": 2.5e-8}`,
			contains: []string{"2.5e-8", "small"},
		},
		{
			name:     "scientific notation uppercase E",
			input:    `{"value": 3E6}`,
			contains: []string{"3E6", "value"},
		},
		{
			name:     "zero",
			input:    `{"zero": 0}`,
			contains: []string{"0", "zero"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.input)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("result should contain '%s', got: %s", s, result)
				}
			}
		})
	}
}

func TestHighlightJSON_EmptyInput(t *testing.T) {
	h := viewer.NewJSONHighlighter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "empty object",
			input:    "{}",
			expected: "{}",
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.input)
			// For empty input, should return empty
			if tt.input == "" && result != "" {
				t.Errorf("empty input should return empty result, got: %s", result)
			}
			// For empty containers, should contain the brackets
			if tt.input != "" {
				if !strings.Contains(result, tt.input[:1]) || !strings.Contains(result, tt.input[len(tt.input)-1:]) {
					t.Errorf("should preserve brackets for input '%s', got: %s", tt.input, result)
				}
			}
		})
	}
}

func TestHighlightJSON_InvalidJSON(t *testing.T) {
	h := viewer.NewJSONHighlighter()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "malformed object",
			input: `{"key": }`,
		},
		{
			name:  "unquoted key",
			input: `{key: "value"}`,
		},
		{
			name:  "single quotes",
			input: `{'key': 'value'}`,
		},
		{
			name:  "trailing comma",
			input: `{"key": "value",}`,
		},
		{
			name:  "random text",
			input: `hello world`,
		},
		{
			name:  "unclosed string",
			input: `{"key": "unclosed`,
		},
		{
			name:  "unclosed brace",
			input: `{"key": "value"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Highlight panicked on invalid JSON: %v", r)
				}
			}()

			result := h.Highlight(tt.input)

			// Should return the input (possibly with some styling applied)
			// At minimum, the original text content should be preserved
			// (minus ANSI codes, which we can't easily strip in tests)
			if result == "" && tt.input != "" {
				t.Error("result should not be empty for non-empty input")
			}
		})
	}
}

func TestHighlightJSON_PreservesWhitespace(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `{
    "name": "John",
    "items": [
        1,
        2,
        3
    ]
}`

	result := h.Highlight(input)

	// Count newlines in input and result should be the same
	inputNewlines := strings.Count(input, "\n")
	resultNewlines := strings.Count(result, "\n")

	if inputNewlines != resultNewlines {
		t.Errorf("newline count mismatch: input has %d, result has %d", inputNewlines, resultNewlines)
	}

	// Indentation should be preserved (check for spaces at line starts)
	if !strings.Contains(result, "    ") {
		t.Error("indentation (4 spaces) should be preserved")
	}
}

func TestHighlightJSON_ComplexStructure(t *testing.T) {
	h := viewer.NewJSONHighlighter()
	input := `{
  "string": "hello",
  "number": 42,
  "float": 3.14,
  "boolean": true,
  "null": null,
  "array": [1, "two", false],
  "nested": {
    "deep": {
      "value": 100
    }
  }
}`

	result := h.Highlight(input)

	// All values should be present
	expectedContents := []string{
		"string", "hello",
		"number", "42",
		"float", "3.14",
		"boolean", "true",
		"null",
		"array", "1", "two", "false",
		"nested", "deep", "value", "100",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("result should contain '%s'", expected)
		}
	}
}

func TestJSONHighlighter_ImplementsHighlighter(t *testing.T) {
	var _ viewer.Highlighter = (*viewer.JSONHighlighter)(nil)
}

func TestNewJSONHighlighterWithColors(t *testing.T) {
	// Test that custom colors can be provided
	colors := viewer.HighlightColors{
		String:  "#00FF00",
		Number:  "#00FFFF",
		Boolean: "#FFFF00",
		Null:    "#FF00FF",
		Key:     "#0000FF",
		Bracket: "#FFFFFF",
	}

	h := viewer.NewJSONHighlighterWithColors(colors)
	result := h.Highlight(`{"key": "value"}`)

	// Should still work with custom colors
	if !strings.Contains(result, "key") {
		t.Error("result should contain 'key'")
	}
	if !strings.Contains(result, "value") {
		t.Error("result should contain 'value'")
	}
}
