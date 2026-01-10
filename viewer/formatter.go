package viewer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Formatter formats data for display
type Formatter interface {
	Format(data []byte) (string, error)
}

// JSONFormatter formats JSON data with indentation
type JSONFormatter struct {
	indent string
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		indent: "  ",
	}
}

// Format formats JSON data with indentation
func (f *JSONFormatter) Format(data []byte) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", f.indent)
	if err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return out.String(), nil
}

// HexFormatter formats binary data as hex dump
type HexFormatter struct {
	bytesPerLine int
}

// NewHexFormatter creates a new hex formatter
func NewHexFormatter() *HexFormatter {
	return &HexFormatter{
		bytesPerLine: 16,
	}
}

// Format formats data as a hex dump
func (f *HexFormatter) Format(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var out strings.Builder
	for i := 0; i < len(data); i += f.bytesPerLine {
		end := i + f.bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		// Offset
		fmt.Fprintf(&out, "%08x  ", i)

		// Hex bytes
		hexStr := hex.EncodeToString(chunk)
		for j := 0; j < len(hexStr); j += 2 {
			out.WriteString(hexStr[j : j+2])
			out.WriteByte(' ')
		}

		// Padding for incomplete lines
		padding := f.bytesPerLine - len(chunk)
		for j := 0; j < padding; j++ {
			out.WriteString("   ")
		}

		out.WriteString(" |")

		// ASCII representation
		for _, b := range chunk {
			if b >= 32 && b < 127 {
				out.WriteByte(b)
			} else {
				out.WriteByte('.')
			}
		}

		out.WriteString("|\n")
	}

	return out.String(), nil
}

// TextFormatter formats text data
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format returns the text as-is
func (f *TextFormatter) Format(data []byte) (string, error) {
	return string(data), nil
}

// AutoFormatter automatically detects and formats data
type AutoFormatter struct {
	jsonFormatter *JSONFormatter
	hexFormatter  *HexFormatter
	textFormatter *TextFormatter
}

// NewAutoFormatter creates a new auto formatter
func NewAutoFormatter() *AutoFormatter {
	return &AutoFormatter{
		jsonFormatter: NewJSONFormatter(),
		hexFormatter:  NewHexFormatter(),
		textFormatter: NewTextFormatter(),
	}
}

// Format auto-detects and formats data
func (f *AutoFormatter) Format(data []byte) (string, error) {
	dt := DetectType(data)
	switch dt {
	case DataTypeJSON:
		return f.jsonFormatter.Format(data)
	case DataTypeBinary, DataTypeCompressedGzip, DataTypeCompressedZlib:
		return f.hexFormatter.Format(data)
	default:
		return f.textFormatter.Format(data)
	}
}
