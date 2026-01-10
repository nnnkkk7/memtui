package viewer

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

// DataType represents the detected data type
type DataType int

const (
	DataTypeText DataType = iota
	DataTypeJSON
	DataTypeBinary
	DataTypeCompressedGzip
	DataTypeCompressedZlib
)

// String returns the string representation of the data type
func (dt DataType) String() string {
	switch dt {
	case DataTypeText:
		return "Text"
	case DataTypeJSON:
		return "JSON"
	case DataTypeBinary:
		return "Binary"
	case DataTypeCompressedGzip:
		return "Gzip"
	case DataTypeCompressedZlib:
		return "Zlib"
	default:
		return "Unknown"
	}
}

// DetectType detects the data type of the given bytes
func DetectType(data []byte) DataType {
	if len(data) == 0 {
		return DataTypeText
	}

	// Check for compression magic bytes first
	if isGzip(data) {
		return DataTypeCompressedGzip
	}
	if isZlib(data) {
		return DataTypeCompressedZlib
	}

	// Check for binary data
	if isBinary(data) {
		return DataTypeBinary
	}

	// Check for JSON
	if isJSON(data) {
		return DataTypeJSON
	}

	return DataTypeText
}

// isGzip checks for gzip magic bytes (0x1f, 0x8b)
func isGzip(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// isZlib checks for zlib magic bytes (0x78)
func isZlib(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// Zlib starts with 0x78 followed by compression level byte
	return data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x5e || data[1] == 0x9c || data[1] == 0xda)
}

// isBinary checks if the data contains binary (non-printable) characters
func isBinary(data []byte) bool {
	for i := 0; i < len(data); i++ {
		b := data[i]
		// Check for null bytes or control characters (except common ones like \t, \n, \r)
		if b == 0 {
			return true
		}
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			return true
		}
		// Check for non-UTF8 sequences
		if b >= 0x80 {
			// Try to decode as UTF8
			r, size := utf8.DecodeRune(data[i:])
			if r == utf8.RuneError && size == 1 {
				return true
			}
			i += size - 1
		}
	}
	return false
}

// isJSON checks if the data is valid JSON
func isJSON(data []byte) bool {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return false
	}

	// Quick check: JSON should start with { or [
	if data[0] != '{' && data[0] != '[' {
		return false
	}

	return json.Valid(data)
}
