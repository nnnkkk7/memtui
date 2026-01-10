// Package viewer provides formatting and highlighting for data display.
package viewer

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Highlighter applies syntax highlighting to input text.
type Highlighter interface {
	Highlight(input string) string
}

// HighlightColors defines the color palette for syntax highlighting.
type HighlightColors struct {
	String  string // Color for string values (default: green)
	Number  string // Color for numeric values (default: cyan)
	Boolean string // Color for true/false (default: yellow)
	Null    string // Color for null (default: magenta)
	Key     string // Color for object keys (default: blue)
	Bracket string // Color for brackets/braces (default: white)
}

// DefaultHighlightColors returns the default color scheme for JSON highlighting.
func DefaultHighlightColors() HighlightColors {
	return HighlightColors{
		String:  "#98C379", // Green
		Number:  "#56B6C2", // Cyan
		Boolean: "#E5C07B", // Yellow
		Null:    "#C678DD", // Magenta
		Key:     "#61AFEF", // Blue
		Bracket: "#ABB2BF", // White/Gray
	}
}

// JSONHighlighter applies syntax highlighting to JSON text.
type JSONHighlighter struct {
	stringStyle  lipgloss.Style
	numberStyle  lipgloss.Style
	boolStyle    lipgloss.Style
	nullStyle    lipgloss.Style
	keyStyle     lipgloss.Style
	bracketStyle lipgloss.Style
}

// NewJSONHighlighter creates a new JSON highlighter with default colors.
func NewJSONHighlighter() *JSONHighlighter {
	return NewJSONHighlighterWithColors(DefaultHighlightColors())
}

// NewJSONHighlighterWithColors creates a new JSON highlighter with custom colors.
func NewJSONHighlighterWithColors(colors HighlightColors) *JSONHighlighter {
	return &JSONHighlighter{
		stringStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color(colors.String)),
		numberStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Number)),
		boolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Boolean)),
		nullStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Null)),
		keyStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Key)),
		bracketStyle: lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Bracket)),
	}
}

// Highlight applies JSON syntax highlighting to the input string.
// It handles strings, numbers, booleans, null, keys, and structural characters.
// For invalid JSON, it returns the input with best-effort highlighting.
func (h *JSONHighlighter) Highlight(input string) string {
	if input == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(input) * 2) // Pre-allocate for ANSI codes

	i := 0
	for i < len(input) {
		ch := input[i]

		switch {
		case ch == '"':
			// Handle strings (could be key or value)
			strEnd := h.findStringEnd(input, i)
			str := input[i : strEnd+1]

			// Check if this is a key (followed by colon after whitespace)
			isKey := h.isKey(input, strEnd+1)
			if isKey {
				result.WriteString(h.keyStyle.Render(str))
			} else {
				result.WriteString(h.stringStyle.Render(str))
			}
			i = strEnd + 1

		case ch == '{' || ch == '}' || ch == '[' || ch == ']':
			result.WriteString(h.bracketStyle.Render(string(ch)))
			i++

		case ch == ':' || ch == ',':
			// Structural characters - keep as-is
			result.WriteByte(ch)
			i++

		case ch == 't' && i+4 <= len(input) && input[i:i+4] == "true":
			result.WriteString(h.boolStyle.Render("true"))
			i += 4

		case ch == 'f' && i+5 <= len(input) && input[i:i+5] == "false":
			result.WriteString(h.boolStyle.Render("false"))
			i += 5

		case ch == 'n' && i+4 <= len(input) && input[i:i+4] == "null":
			result.WriteString(h.nullStyle.Render("null"))
			i += 4

		case h.isNumberStart(ch, input, i):
			numEnd := h.findNumberEnd(input, i)
			num := input[i:numEnd]
			result.WriteString(h.numberStyle.Render(num))
			i = numEnd

		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			// Preserve whitespace exactly
			result.WriteByte(ch)
			i++

		default:
			// Unknown character - preserve as-is (handles invalid JSON gracefully)
			result.WriteByte(ch)
			i++
		}
	}

	return result.String()
}

// findStringEnd finds the closing quote of a JSON string, handling escape sequences.
// Returns the index of the closing quote, or len(input)-1 if not found.
func (h *JSONHighlighter) findStringEnd(input string, start int) int {
	// Start after the opening quote
	i := start + 1
	for i < len(input) {
		if input[i] == '\\' && i+1 < len(input) {
			// Skip escaped character
			i += 2
			continue
		}
		if input[i] == '"' {
			return i
		}
		i++
	}
	// Unclosed string - return end of input
	return len(input) - 1
}

// isKey checks if the string at the given position is an object key.
// A key is followed by a colon (with optional whitespace).
func (h *JSONHighlighter) isKey(input string, afterString int) bool {
	for i := afterString; i < len(input); i++ {
		ch := input[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			continue
		}
		return ch == ':'
	}
	return false
}

// isNumberStart checks if the character at position i starts a number.
func (h *JSONHighlighter) isNumberStart(ch byte, input string, i int) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	if ch == '-' && i+1 < len(input) {
		next := input[i+1]
		return next >= '0' && next <= '9'
	}
	return false
}

// findNumberEnd finds the end of a JSON number.
// Handles integers, floats, negative numbers, and scientific notation.
func (h *JSONHighlighter) findNumberEnd(input string, start int) int {
	i := start

	// Optional negative sign
	if i < len(input) && input[i] == '-' {
		i++
	}

	// Integer part
	for i < len(input) && input[i] >= '0' && input[i] <= '9' {
		i++
	}

	// Decimal part
	if i < len(input) && input[i] == '.' {
		i++
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
	}

	// Exponent part (scientific notation)
	if i < len(input) && (input[i] == 'e' || input[i] == 'E') {
		i++
		// Optional sign after exponent
		if i < len(input) && (input[i] == '+' || input[i] == '-') {
			i++
		}
		// Exponent digits
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
	}

	return i
}
