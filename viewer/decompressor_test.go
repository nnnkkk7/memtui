package viewer_test

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/nnnkkk7/memtui/viewer"
)

// Helper function to create gzip compressed data
func createGzipData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		t.Fatalf("failed to write gzip data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	return buf.Bytes()
}

// Helper function to create zlib compressed data
func createZlibData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		t.Fatalf("failed to write zlib data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zlib writer: %v", err)
	}
	return buf.Bytes()
}

// Helper function to create zstd compressed data
func createZstdData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("failed to create zstd writer: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		t.Fatalf("failed to write zstd data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zstd writer: %v", err)
	}
	return buf.Bytes()
}

func TestDecompress_Gzip(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "simple text",
			input:    createGzipData(t, []byte("hello world")),
			expected: []byte("hello world"),
		},
		{
			name:     "json data",
			input:    createGzipData(t, []byte(`{"key": "value"}`)),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "utf8 text",
			input:    createGzipData(t, []byte("Hello 🌍 World")),
			expected: []byte("Hello 🌍 World"),
		},
		{
			name:     "empty data",
			input:    createGzipData(t, []byte("")),
			expected: []byte(""),
		},
		{
			name:     "large data",
			input:    createGzipData(t, bytes.Repeat([]byte("abcdefghij"), 1000)),
			expected: bytes.Repeat([]byte("abcdefghij"), 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := viewer.Decompress(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDecompress_Zlib(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "simple text",
			input:    createZlibData(t, []byte("hello world")),
			expected: []byte("hello world"),
		},
		{
			name:     "json data",
			input:    createZlibData(t, []byte(`{"key": "value"}`)),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "utf8 text",
			input:    createZlibData(t, []byte("Héllo Wörld 🚀")),
			expected: []byte("Héllo Wörld 🚀"),
		},
		{
			name:     "binary data",
			input:    createZlibData(t, []byte{0x00, 0x01, 0x02, 0x03, 0x04}),
			expected: []byte{0x00, 0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := viewer.Decompress(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDecompress_UncompressedData(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "plain text",
			input: []byte("hello world"),
		},
		{
			name:  "json",
			input: []byte(`{"key": "value"}`),
		},
		{
			name:  "binary",
			input: []byte{0x00, 0x01, 0x02, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := viewer.Decompress(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Uncompressed data should be returned as-is
			if !bytes.Equal(result, tt.input) {
				t.Errorf("expected %q, got %q", tt.input, result)
			}
		})
	}
}

func TestDecompress_EmptyAndNil(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := viewer.Decompress(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != 0 {
				t.Errorf("expected empty result, got %q", result)
			}
		})
	}
}

func TestDecompress_InvalidCompressedData(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "invalid gzip header",
			input: []byte{0x1f, 0x8b, 0x00, 0x00, 0x00}, // Invalid gzip
		},
		{
			name:  "truncated gzip",
			input: []byte{0x1f, 0x8b, 0x08}, // Truncated
		},
		{
			name:  "invalid zlib header",
			input: []byte{0x78, 0x9c, 0x00, 0x00, 0x00}, // Invalid zlib
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := viewer.Decompress(tt.input)
			if err == nil {
				t.Error("expected error for invalid compressed data")
			}
		})
	}
}

func TestDecompress_DetectionPriority(t *testing.T) {
	// Test that gzip detection takes priority over zlib
	original := []byte("test data for compression")

	t.Run("gzip is detected correctly", func(t *testing.T) {
		gzipData := createGzipData(t, original)
		// Verify it starts with gzip magic bytes
		if len(gzipData) < 2 || gzipData[0] != 0x1f || gzipData[1] != 0x8b {
			t.Fatal("gzip data doesn't have correct magic bytes")
		}

		result, err := viewer.Decompress(gzipData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(result, original) {
			t.Errorf("expected %q, got %q", original, result)
		}
	})

	t.Run("zlib is detected correctly", func(t *testing.T) {
		zlibData := createZlibData(t, original)
		// Verify it starts with zlib magic byte
		if len(zlibData) < 2 || zlibData[0] != 0x78 {
			t.Fatal("zlib data doesn't have correct magic byte")
		}

		result, err := viewer.Decompress(zlibData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(result, original) {
			t.Errorf("expected %q, got %q", original, result)
		}
	})
}

func TestIsGzipCompressed(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "valid gzip magic bytes",
			input:    []byte{0x1f, 0x8b, 0x08, 0x00},
			expected: true,
		},
		{
			name:     "actual gzip data",
			input:    createGzipData(t, []byte("test")),
			expected: true,
		},
		{
			name:     "not gzip - plain text",
			input:    []byte("hello world"),
			expected: false,
		},
		{
			name:     "not gzip - zlib",
			input:    []byte{0x78, 0x9c, 0x00},
			expected: false,
		},
		{
			name:     "empty",
			input:    []byte{},
			expected: false,
		},
		{
			name:     "single byte",
			input:    []byte{0x1f},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.IsGzipCompressed(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsZlibCompressed(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "zlib no compression (0x78 0x01)",
			input:    []byte{0x78, 0x01, 0x00},
			expected: true,
		},
		{
			name:     "zlib default compression (0x78 0x9c)",
			input:    []byte{0x78, 0x9c, 0x00},
			expected: true,
		},
		{
			name:     "zlib best compression (0x78 0xda)",
			input:    []byte{0x78, 0xda, 0x00},
			expected: true,
		},
		{
			name:     "zlib best speed (0x78 0x5e)",
			input:    []byte{0x78, 0x5e, 0x00},
			expected: true,
		},
		{
			name:     "actual zlib data",
			input:    createZlibData(t, []byte("test")),
			expected: true,
		},
		{
			name:     "not zlib - plain text",
			input:    []byte("hello world"),
			expected: false,
		},
		{
			name:     "not zlib - gzip",
			input:    []byte{0x1f, 0x8b, 0x08},
			expected: false,
		},
		{
			name:     "invalid zlib second byte",
			input:    []byte{0x78, 0x00, 0x00},
			expected: false,
		},
		{
			name:     "empty",
			input:    []byte{},
			expected: false,
		},
		{
			name:     "single byte",
			input:    []byte{0x78},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.IsZlibCompressed(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsZstdCompressed(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "valid zstd magic bytes",
			input:    []byte{0x28, 0xb5, 0x2f, 0xfd, 0x00},
			expected: true,
		},
		{
			name:     "actual zstd data",
			input:    createZstdData(t, []byte("test")),
			expected: true,
		},
		{
			name:     "not zstd - plain text",
			input:    []byte("hello world"),
			expected: false,
		},
		{
			name:     "not zstd - gzip",
			input:    []byte{0x1f, 0x8b, 0x08},
			expected: false,
		},
		{
			name:     "not zstd - zlib",
			input:    []byte{0x78, 0x9c, 0x00},
			expected: false,
		},
		{
			name:     "partial zstd magic (3 bytes)",
			input:    []byte{0x28, 0xb5, 0x2f},
			expected: false,
		},
		{
			name:     "empty",
			input:    []byte{},
			expected: false,
		},
		{
			name:     "single byte",
			input:    []byte{0x28},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := viewer.IsZstdCompressed(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDecompress_Zstd(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "simple text",
			input:    createZstdData(t, []byte("hello world")),
			expected: []byte("hello world"),
		},
		{
			name:     "json data",
			input:    createZstdData(t, []byte(`{"key": "value"}`)),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "utf8 text",
			input:    createZstdData(t, []byte("Hello 🌍 World")),
			expected: []byte("Hello 🌍 World"),
		},
		{
			name:     "empty data",
			input:    createZstdData(t, []byte("")),
			expected: []byte(""),
		},
		{
			name:     "large data",
			input:    createZstdData(t, bytes.Repeat([]byte("abcdefghij"), 1000)),
			expected: bytes.Repeat([]byte("abcdefghij"), 1000),
		},
		{
			name:     "binary data",
			input:    createZstdData(t, []byte{0x00, 0x01, 0x02, 0x03, 0x04}),
			expected: []byte{0x00, 0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := viewer.Decompress(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDecompress_InvalidZstd(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "invalid zstd - corrupted data after header",
			input: []byte{0x28, 0xb5, 0x2f, 0xfd, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "invalid zstd - truncated",
			input: []byte{0x28, 0xb5, 0x2f, 0xfd, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := viewer.Decompress(tt.input)
			if err == nil {
				t.Error("expected error for invalid zstd data")
			}
		})
	}
}

func TestDecompress_ZstdDetectionPriority(t *testing.T) {
	original := []byte("test data for zstd compression")

	t.Run("zstd is detected correctly", func(t *testing.T) {
		zstdData := createZstdData(t, original)
		// Verify it starts with zstd magic bytes
		if len(zstdData) < 4 || zstdData[0] != 0x28 || zstdData[1] != 0xb5 || zstdData[2] != 0x2f || zstdData[3] != 0xfd {
			t.Fatal("zstd data doesn't have correct magic bytes")
		}

		result, err := viewer.Decompress(zstdData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(result, original) {
			t.Errorf("expected %q, got %q", original, result)
		}
	})
}
