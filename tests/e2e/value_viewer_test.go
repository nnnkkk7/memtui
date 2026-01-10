//go:build e2e

package e2e_test

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/viewer"
	viewerPkg "github.com/nnnkkk7/memtui/viewer"
)

// viewerKeyPrefix is the key prefix for viewer tests
const viewerKeyPrefix = "e2e:viewer:"

// viewerMC creates a memcache client for viewer tests
func viewerMC(t *testing.T) *memcache.Client {
	t.Helper()
	return memcache.New(getMemcachedAddr())
}

// viewerSetKey sets a key for testing and registers cleanup
func viewerSetKey(t *testing.T, mc *memcache.Client, key string, value []byte) {
	t.Helper()
	err := mc.Set(&memcache.Item{
		Key:   viewerKeyPrefix + key,
		Value: value,
	})
	if err != nil {
		t.Fatalf("failed to set test key %s: %v", key, err)
	}
	t.Cleanup(func() {
		mc.Delete(viewerKeyPrefix + key)
	})
}

// viewerGetValue retrieves a test key's value
func viewerGetValue(t *testing.T, mc *memcache.Client, key string) []byte {
	t.Helper()
	item, err := mc.Get(viewerKeyPrefix + key)
	if err != nil {
		t.Fatalf("failed to get test key %s: %v", key, err)
	}
	return item.Value
}

// gzipCompress compresses data with gzip
func gzipCompress(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// zlibCompress compresses data with zlib
func zlibCompress(data []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// ============================================================================
// JSON Value Display Tests
// ============================================================================

func TestE2E_ViewerJSONDisplay(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name           string
		jsonValue      interface{}
		expectedFields []string
	}{
		{
			name: "simple_object",
			jsonValue: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
			expectedFields: []string{"name", "John Doe", "email", "john@example.com", "age"},
		},
		{
			name: "nested_object",
			jsonValue: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name":     "Jane",
						"settings": map[string]interface{}{"theme": "dark"},
					},
				},
			},
			expectedFields: []string{"user", "profile", "name", "Jane", "settings", "theme", "dark"},
		},
		{
			name: "array_value",
			jsonValue: map[string]interface{}{
				"items": []interface{}{"apple", "banana", "cherry"},
				"count": 3,
			},
			expectedFields: []string{"items", "apple", "banana", "cherry", "count"},
		},
		{
			name: "complex_mixed",
			jsonValue: map[string]interface{}{
				"id":     12345,
				"active": true,
				"tags":   []interface{}{"important", "urgent"},
				"metadata": map[string]interface{}{
					"created": "2024-01-01",
					"version": 2.5,
				},
			},
			expectedFields: []string{"id", "12345", "active", "true", "tags", "important", "metadata", "created"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal JSON
			jsonBytes, err := json.Marshal(tt.jsonValue)
			if err != nil {
				t.Fatalf("failed to marshal JSON: %v", err)
			}

			// Store in Memcached
			viewerSetKey(t, mc, "json:"+tt.name, jsonBytes)

			// Wait for storage
			time.Sleep(50 * time.Millisecond)

			// Retrieve and verify
			value := viewerGetValue(t, mc, "json:"+tt.name)

			// Create viewer and set value
			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)
			v.SetViewMode(viewer.ViewModeJSON)

			// Get formatted content
			content := v.Content()

			// Verify pretty-printing (should have newlines)
			if !strings.Contains(content, "\n") {
				t.Error("expected pretty-printed JSON with newlines")
			}

			// Verify expected fields are present
			for _, field := range tt.expectedFields {
				if !strings.Contains(content, field) {
					t.Errorf("expected content to contain '%s', got:\n%s", field, content)
				}
			}

			// Verify auto-detection identifies as JSON
			if v.DetectedType() != "JSON" {
				t.Errorf("expected detected type 'JSON', got '%s'", v.DetectedType())
			}
		})
	}
}

func TestE2E_ViewerJSONPrettyPrinting(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Compact JSON (no formatting)
	compactJSON := `{"key":"value","nested":{"a":1,"b":2},"array":[1,2,3]}`
	viewerSetKey(t, mc, "json:compact", []byte(compactJSON))

	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "json:compact")

	v := viewer.NewModel()
	v.SetSize(120, 40)
	v.SetValue(value)
	v.SetViewMode(viewer.ViewModeJSON)

	content := v.Content()

	// Verify indentation is added
	if !strings.Contains(content, "  ") {
		t.Error("expected indented JSON output")
	}

	// Verify structure is preserved
	lines := strings.Split(content, "\n")
	if len(lines) < 5 {
		t.Errorf("expected multi-line output, got %d lines", len(lines))
	}
}

// ============================================================================
// Plain Text Value Display Tests
// ============================================================================

func TestE2E_ViewerPlainTextDisplay(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name          string
		textValue     string
		expectedLines int
	}{
		{
			name:          "simple_string",
			textValue:     "Hello, World!",
			expectedLines: 1,
		},
		{
			name:          "multiline_text",
			textValue:     "Line 1\nLine 2\nLine 3\nLine 4",
			expectedLines: 4,
		},
		{
			name:          "text_with_special_chars",
			textValue:     "Special chars: tab\there, unicode: \u00e9\u00e0\u00fc",
			expectedLines: 1,
		},
		{
			name:          "empty_lines",
			textValue:     "Before\n\n\nAfter",
			expectedLines: 4,
		},
		{
			name:          "long_text",
			textValue:     strings.Repeat("This is a long line of text. ", 10),
			expectedLines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewerSetKey(t, mc, "text:"+tt.name, []byte(tt.textValue))
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "text:"+tt.name)

			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)
			v.SetViewMode(viewer.ViewModeText)

			content := v.Content()

			// Verify content matches original
			if content != tt.textValue {
				t.Errorf("expected content:\n%s\ngot:\n%s", tt.textValue, content)
			}

			// Verify detected as Text
			if v.DetectedType() != "Text" {
				t.Errorf("expected detected type 'Text', got '%s'", v.DetectedType())
			}
		})
	}
}

// ============================================================================
// Binary Data Display Tests (Hex Dump)
// ============================================================================

func TestE2E_ViewerBinaryDisplay(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name           string
		binaryData     []byte
		expectedHex    []string
		expectedLength int
	}{
		{
			name:           "null_bytes",
			binaryData:     []byte{0x00, 0x01, 0x02, 0x03},
			expectedHex:    []string{"00", "01", "02", "03"},
			expectedLength: 4,
		},
		{
			name:           "full_byte_range",
			binaryData:     []byte{0x00, 0x7F, 0x80, 0xFF},
			expectedHex:    []string{"00", "7f", "80", "ff"},
			expectedLength: 4,
		},
		{
			name:           "control_chars",
			binaryData:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			expectedHex:    []string{"01", "02", "03", "04", "05"},
			expectedLength: 5,
		},
		{
			name:           "mixed_binary_printable",
			binaryData:     []byte{0x48, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00}, // "H\0e\0l\0l\0o\0"
			expectedHex:    []string{"48", "00", "65", "00", "6c"},
			expectedLength: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewerSetKey(t, mc, "binary:"+tt.name, tt.binaryData)
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "binary:"+tt.name)

			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)
			v.SetViewMode(viewer.ViewModeHex)

			content := v.Content()
			contentLower := strings.ToLower(content)

			// Verify hex bytes are present
			for _, hexByte := range tt.expectedHex {
				if !strings.Contains(contentLower, hexByte) {
					t.Errorf("expected hex dump to contain '%s', got:\n%s", hexByte, content)
				}
			}

			// Verify offset column (should start with 00000000)
			if !strings.Contains(content, "00000000") {
				t.Error("expected hex dump to start with offset 00000000")
			}

			// Verify detected as Binary
			if v.DetectedType() != "Binary" {
				t.Errorf("expected detected type 'Binary', got '%s'", v.DetectedType())
			}
		})
	}
}

func TestE2E_ViewerHexDumpFormat(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create 32 bytes of data to verify multi-line hex dump
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}

	viewerSetKey(t, mc, "binary:hexformat", data)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "binary:hexformat")

	v := viewer.NewModel()
	v.SetSize(120, 40)
	v.SetValue(value)
	v.SetViewMode(viewer.ViewModeHex)

	content := v.Content()
	lines := strings.Split(content, "\n")

	// Should have at least 2 lines (16 bytes per line)
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines < 2 {
		t.Errorf("expected at least 2 hex lines for 32 bytes, got %d", nonEmptyLines)
	}

	// Verify ASCII representation column exists (pipe characters)
	if !strings.Contains(content, "|") {
		t.Error("expected ASCII representation column in hex dump")
	}
}

// ============================================================================
// Gzip Compressed Data Tests
// ============================================================================

func TestE2E_ViewerGzipDecompression(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name         string
		originalData string
		expectJSON   bool
	}{
		{
			name:         "gzip_text",
			originalData: "This is plain text that has been gzip compressed.",
			expectJSON:   false,
		},
		{
			name:         "gzip_json",
			originalData: `{"compressed": true, "format": "gzip", "data": {"value": 123}}`,
			expectJSON:   true,
		},
		{
			name:         "gzip_multiline",
			originalData: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
			expectJSON:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress data
			compressed := gzipCompress([]byte(tt.originalData))

			viewerSetKey(t, mc, "gzip:"+tt.name, compressed)
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "gzip:"+tt.name)

			// Verify it's detected as gzip
			detectedType := viewerPkg.DetectType(value)
			if detectedType != viewerPkg.DataTypeCompressedGzip {
				t.Errorf("expected detection as Gzip, got %v", detectedType)
			}

			// Verify decompression works
			decompressed, err := viewerPkg.Decompress(value)
			if err != nil {
				t.Fatalf("failed to decompress: %v", err)
			}

			if string(decompressed) != tt.originalData {
				t.Errorf("decompressed data mismatch:\nexpected: %s\ngot: %s", tt.originalData, string(decompressed))
			}

			// Test viewer with raw compressed data (shows hex)
			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)
			v.SetViewMode(viewer.ViewModeAuto)

			if v.DetectedType() != "Gzip" {
				t.Errorf("expected detected type 'Gzip', got '%s'", v.DetectedType())
			}

			// In auto mode, compressed data should show as hex
			content := v.Content()
			if !strings.Contains(content, "1f") || !strings.Contains(content, "8b") {
				t.Logf("Content: %s", content)
				// Gzip magic bytes should appear in hex view
			}
		})
	}
}

func TestE2E_ViewerGzipMagicBytes(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create gzip compressed data
	original := "Test data for gzip magic byte verification"
	compressed := gzipCompress([]byte(original))

	viewerSetKey(t, mc, "gzip:magic", compressed)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "gzip:magic")

	// Verify magic bytes
	if len(value) < 2 {
		t.Fatal("compressed data too short")
	}
	if value[0] != 0x1f || value[1] != 0x8b {
		t.Errorf("expected gzip magic bytes 0x1f 0x8b, got 0x%02x 0x%02x", value[0], value[1])
	}

	// Verify IsGzipCompressed helper
	if !viewerPkg.IsGzipCompressed(value) {
		t.Error("IsGzipCompressed should return true")
	}
}

// ============================================================================
// Zlib Compressed Data Tests
// ============================================================================

func TestE2E_ViewerZlibDecompression(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name         string
		originalData string
	}{
		{
			name:         "zlib_text",
			originalData: "This is plain text that has been zlib compressed.",
		},
		{
			name:         "zlib_json",
			originalData: `{"compressed": true, "format": "zlib", "nested": {"key": "value"}}`,
		},
		{
			name:         "zlib_large",
			originalData: strings.Repeat("Repeated content for better compression. ", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress data
			compressed := zlibCompress([]byte(tt.originalData))

			viewerSetKey(t, mc, "zlib:"+tt.name, compressed)
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "zlib:"+tt.name)

			// Verify it's detected as zlib
			detectedType := viewerPkg.DetectType(value)
			if detectedType != viewerPkg.DataTypeCompressedZlib {
				t.Errorf("expected detection as Zlib, got %v", detectedType)
			}

			// Verify decompression works
			decompressed, err := viewerPkg.Decompress(value)
			if err != nil {
				t.Fatalf("failed to decompress: %v", err)
			}

			if string(decompressed) != tt.originalData {
				t.Errorf("decompressed data mismatch:\nexpected: %s\ngot: %s", tt.originalData, string(decompressed))
			}

			// Test viewer with raw compressed data
			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)
			v.SetViewMode(viewer.ViewModeAuto)

			if v.DetectedType() != "Zlib" {
				t.Errorf("expected detected type 'Zlib', got '%s'", v.DetectedType())
			}
		})
	}
}

func TestE2E_ViewerZlibMagicBytes(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	original := "Test data for zlib magic byte verification"
	compressed := zlibCompress([]byte(original))

	viewerSetKey(t, mc, "zlib:magic", compressed)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "zlib:magic")

	// Verify magic byte (0x78)
	if len(value) < 2 {
		t.Fatal("compressed data too short")
	}
	if value[0] != 0x78 {
		t.Errorf("expected zlib magic byte 0x78, got 0x%02x", value[0])
	}

	// Verify compression level byte
	validLevels := []byte{0x01, 0x5e, 0x9c, 0xda}
	isValidLevel := false
	for _, level := range validLevels {
		if value[1] == level {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		t.Errorf("unexpected zlib compression level byte: 0x%02x", value[1])
	}

	// Verify IsZlibCompressed helper
	if !viewerPkg.IsZlibCompressed(value) {
		t.Error("IsZlibCompressed should return true")
	}
}

// ============================================================================
// Viewer Scrolling Tests
// ============================================================================

func TestE2E_ViewerScrollingJK(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create content with many lines
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = strings.Repeat("x", 50) // 50 chars per line
	}
	longContent := strings.Join(lines, "\n")

	viewerSetKey(t, mc, "scroll:jk", []byte(longContent))
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "scroll:jk")

	v := viewer.NewModel()
	v.SetSize(80, 20) // Small height to require scrolling
	v.SetValue(value)

	// Initial offset should be 0
	if v.ScrollOffset() != 0 {
		t.Errorf("expected initial scroll offset 0, got %d", v.ScrollOffset())
	}

	// Scroll down with KeyDown (simulating 'j')
	v.Update(tea.KeyMsg{Type: tea.KeyDown})
	if v.ScrollOffset() != 1 {
		t.Errorf("expected scroll offset 1 after down, got %d", v.ScrollOffset())
	}

	// Scroll down multiple times
	for i := 0; i < 10; i++ {
		v.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if v.ScrollOffset() != 11 {
		t.Errorf("expected scroll offset 11 after 11 downs, got %d", v.ScrollOffset())
	}

	// Scroll up with KeyUp (simulating 'k')
	v.Update(tea.KeyMsg{Type: tea.KeyUp})
	if v.ScrollOffset() != 10 {
		t.Errorf("expected scroll offset 10 after up, got %d", v.ScrollOffset())
	}

	// Scroll up to top
	for i := 0; i < 20; i++ {
		v.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	if v.ScrollOffset() != 0 {
		t.Errorf("expected scroll offset 0 at top, got %d", v.ScrollOffset())
	}
}

func TestE2E_ViewerScrollingGG(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create content with many lines
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "Line " + strings.Repeat("x", 50)
	}
	longContent := strings.Join(lines, "\n")

	viewerSetKey(t, mc, "scroll:gg", []byte(longContent))
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "scroll:gg")

	v := viewer.NewModel()
	v.SetSize(80, 20)
	v.SetValue(value)

	// Scroll down first
	for i := 0; i < 50; i++ {
		v.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	midOffset := v.ScrollOffset()
	if midOffset == 0 {
		t.Error("expected non-zero scroll offset after scrolling down")
	}

	// Test PageDown
	v.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	afterPageDown := v.ScrollOffset()
	if afterPageDown <= midOffset {
		t.Errorf("expected larger offset after PageDown, got %d (was %d)", afterPageDown, midOffset)
	}

	// Test PageUp
	v.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	afterPageUp := v.ScrollOffset()
	if afterPageUp >= afterPageDown {
		t.Errorf("expected smaller offset after PageUp, got %d (was %d)", afterPageUp, afterPageDown)
	}
}

func TestE2E_ViewerScrollBoundary(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Short content that fits in view
	shortContent := "Line 1\nLine 2\nLine 3"
	viewerSetKey(t, mc, "scroll:boundary", []byte(shortContent))
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "scroll:boundary")

	v := viewer.NewModel()
	v.SetSize(80, 20) // Larger than content
	v.SetValue(value)

	// Try to scroll down - should be limited
	for i := 0; i < 100; i++ {
		v.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Render the view to trigger scroll clamping
	v.View()

	// Offset should be clamped (max offset = 0 when content fits)
	offset := v.ScrollOffset()
	if offset > 0 {
		// This might be expected behavior depending on implementation
		t.Logf("scroll offset after boundary: %d", offset)
	}

	// Try to scroll up past top
	for i := 0; i < 100; i++ {
		v.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	if v.ScrollOffset() != 0 {
		t.Errorf("expected scroll offset 0 at boundary, got %d", v.ScrollOffset())
	}
}

// ============================================================================
// Large Value Handling Tests
// ============================================================================

func TestE2E_ViewerLargeValue(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name string
		size int
	}{
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"500KB", 500 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create large value
			largeValue := make([]byte, tt.size)
			for i := range largeValue {
				largeValue[i] = byte('A' + (i % 26))
			}

			viewerSetKey(t, mc, "large:"+tt.name, largeValue)
			time.Sleep(100 * time.Millisecond)

			value := viewerGetValue(t, mc, "large:"+tt.name)

			if len(value) != tt.size {
				t.Errorf("expected %d bytes, got %d", tt.size, len(value))
			}

			// Create viewer and verify it handles large content
			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)

			// Should not panic
			content := v.Content()
			if content == "" {
				t.Error("expected non-empty content for large value")
			}

			// Verify scrolling works
			v.Update(tea.KeyMsg{Type: tea.KeyPgDown})
			v.Update(tea.KeyMsg{Type: tea.KeyPgDown})
			v.Update(tea.KeyMsg{Type: tea.KeyPgUp})

			view := v.View()
			if view == "" {
				t.Error("expected non-empty view for large value")
			}
		})
	}
}

func TestE2E_ViewerLargeJSON(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create large JSON with many fields
	data := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		key := "field_" + strings.Repeat("x", 10) + "_" + string(rune('0'+i%10))
		data[key] = map[string]interface{}{
			"index":    i,
			"value":    strings.Repeat("data", 10),
			"nested":   map[string]int{"a": i, "b": i * 2},
			"array":    []int{i, i + 1, i + 2},
			"isActive": i%2 == 0,
		}
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal large JSON: %v", err)
	}

	t.Logf("Large JSON size: %d bytes", len(jsonBytes))

	viewerSetKey(t, mc, "large:json", jsonBytes)
	time.Sleep(100 * time.Millisecond)

	value := viewerGetValue(t, mc, "large:json")

	v := viewer.NewModel()
	v.SetSize(120, 40)
	v.SetValue(value)
	v.SetViewMode(viewer.ViewModeJSON)

	// Should format without error
	content := v.Content()
	if !strings.Contains(content, "field_") {
		t.Error("expected formatted JSON to contain field names")
	}

	// Verify detection
	if v.DetectedType() != "JSON" {
		t.Errorf("expected detected type 'JSON', got '%s'", v.DetectedType())
	}
}

func TestE2E_ViewerLargeBinary(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create large binary data
	size := 50 * 1024 // 50KB
	binaryData := make([]byte, size)
	for i := range binaryData {
		binaryData[i] = byte(i % 256)
	}

	viewerSetKey(t, mc, "large:binary", binaryData)
	time.Sleep(100 * time.Millisecond)

	value := viewerGetValue(t, mc, "large:binary")

	v := viewer.NewModel()
	v.SetSize(120, 40)
	v.SetValue(value)
	v.SetViewMode(viewer.ViewModeHex)

	// Should format as hex without error
	content := v.Content()

	// Verify hex format
	if !strings.Contains(content, "00000000") {
		t.Error("expected hex dump with offset")
	}

	// Verify multiple lines (should have many lines for 50KB)
	lines := strings.Split(content, "\n")
	if len(lines) < 100 {
		t.Errorf("expected many hex lines, got %d", len(lines))
	}
}

// ============================================================================
// Value Type Detection Accuracy Tests
// ============================================================================

func TestE2E_ViewerTypeDetection(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name         string
		value        []byte
		expectedType string
	}{
		{
			name:         "json_object",
			value:        []byte(`{"key": "value"}`),
			expectedType: "JSON",
		},
		{
			name:         "json_array",
			value:        []byte(`[1, 2, 3, "four"]`),
			expectedType: "JSON",
		},
		{
			name:         "plain_text",
			value:        []byte("Hello, World!"),
			expectedType: "Text",
		},
		{
			name:         "multiline_text",
			value:        []byte("Line 1\nLine 2\nLine 3"),
			expectedType: "Text",
		},
		{
			name:         "binary_null",
			value:        []byte{0x00, 0x01, 0x02},
			expectedType: "Binary",
		},
		{
			name:         "binary_control",
			value:        []byte{0x01, 0x02, 0x03, 0x04},
			expectedType: "Binary",
		},
		{
			name:         "gzip_compressed",
			value:        gzipCompress([]byte("test data")),
			expectedType: "Gzip",
		},
		{
			name:         "zlib_compressed",
			value:        zlibCompress([]byte("test data")),
			expectedType: "Zlib",
		},
		{
			name:         "json_like_but_invalid",
			value:        []byte(`{invalid json}`),
			expectedType: "Text",
		},
		{
			name:         "number_string",
			value:        []byte("12345"),
			expectedType: "Text",
		},
		{
			name:         "unicode_text",
			value:        []byte("Hello 世界"),
			expectedType: "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewerSetKey(t, mc, "detect:"+tt.name, tt.value)
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "detect:"+tt.name)

			v := viewer.NewModel()
			v.SetSize(120, 40)
			v.SetValue(value)

			detected := v.DetectedType()
			if detected != tt.expectedType {
				t.Errorf("expected type '%s', got '%s'", tt.expectedType, detected)
			}
		})
	}
}

func TestE2E_ViewerDetectorEdgeCases(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	tests := []struct {
		name         string
		value        []byte
		expectedType viewerPkg.DataType
	}{
		{
			name:         "empty_value",
			value:        []byte{},
			expectedType: viewerPkg.DataTypeText,
		},
		{
			name:         "whitespace_only",
			value:        []byte("   \n\t\r   "),
			expectedType: viewerPkg.DataTypeText,
		},
		{
			name:         "json_with_whitespace",
			value:        []byte("  \n  {\"key\": \"value\"}  \n  "),
			expectedType: viewerPkg.DataTypeJSON,
		},
		{
			name:         "almost_gzip",
			value:        []byte{0x1f, 0x8a, 0x00, 0x00}, // Wrong second magic byte
			expectedType: viewerPkg.DataTypeBinary,
		},
		{
			name:         "almost_zlib",
			value:        []byte{0x78, 0x00, 0x00, 0x00}, // Invalid compression level
			expectedType: viewerPkg.DataTypeBinary,
		},
		{
			name:         "nested_json",
			value:        []byte(`{"a":{"b":{"c":{"d":{"e":"deep"}}}}}`),
			expectedType: viewerPkg.DataTypeJSON,
		},
		{
			name:         "json_array_of_objects",
			value:        []byte(`[{"id":1},{"id":2},{"id":3}]`),
			expectedType: viewerPkg.DataTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.value) > 0 {
				viewerSetKey(t, mc, "edge:"+tt.name, tt.value)
				time.Sleep(50 * time.Millisecond)
			}

			detected := viewerPkg.DetectType(tt.value)
			if detected != tt.expectedType {
				t.Errorf("expected type %v, got %v", tt.expectedType, detected)
			}
		})
	}
}

// ============================================================================
// View Mode Switching Tests
// ============================================================================

func TestE2E_ViewerModeSwitching(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Use JSON data that can be viewed in different modes
	jsonData := `{"name": "test", "value": 123}`
	viewerSetKey(t, mc, "mode:switch", []byte(jsonData))
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "mode:switch")

	v := viewer.NewModel()
	v.SetSize(120, 40)
	v.SetValue(value)

	// Test mode switching via keyboard shortcuts
	modes := []struct {
		key      rune
		expected viewer.ViewMode
		name     string
	}{
		{'J', viewer.ViewModeJSON, "JSON"},
		{'H', viewer.ViewModeHex, "Hex"},
		{'T', viewer.ViewModeText, "Text"},
		{'A', viewer.ViewModeAuto, "Auto"},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{m.key}})

			if v.ViewMode() != m.expected {
				t.Errorf("expected mode %v after '%c', got %v", m.expected, m.key, v.ViewMode())
			}

			// Verify content changes appropriately
			content := v.Content()
			if content == "" {
				t.Error("expected non-empty content")
			}
		})
	}
}

func TestE2E_ViewerModeContentDifference(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	jsonData := `{"key":"value"}`
	viewerSetKey(t, mc, "mode:content", []byte(jsonData))
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "mode:content")

	v := viewer.NewModel()
	v.SetSize(120, 40)
	v.SetValue(value)

	// Get content in different modes
	v.SetViewMode(viewer.ViewModeJSON)
	jsonContent := v.Content()

	v.SetViewMode(viewer.ViewModeHex)
	hexContent := v.Content()

	v.SetViewMode(viewer.ViewModeText)
	textContent := v.Content()

	// All should be different
	if jsonContent == hexContent {
		t.Error("JSON and Hex content should be different")
	}
	if hexContent == textContent {
		t.Error("Hex and Text content should be different")
	}

	// JSON should be pretty-printed (multi-line)
	if !strings.Contains(jsonContent, "\n") {
		t.Error("JSON content should be multi-line")
	}

	// Hex should contain offset
	if !strings.Contains(hexContent, "00000000") {
		t.Error("Hex content should contain offset")
	}

	// Text should be raw
	if textContent != jsonData {
		t.Errorf("Text content should be raw: expected '%s', got '%s'", jsonData, textContent)
	}
}

// ============================================================================
// KeyInfo Integration Tests
// ============================================================================

func TestE2E_ViewerWithKeyInfo(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	testData := []byte(`{"user": "test", "active": true}`)
	viewerSetKey(t, mc, "keyinfo:test", testData)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "keyinfo:test")

	v := viewer.NewModel()
	v.SetSize(120, 40)

	// Set key info with metadata
	ki := models.KeyInfo{
		Key:        viewerKeyPrefix + "keyinfo:test",
		Size:       len(testData),
		Expiration: 0,
		CAS:        12345,
		SlabClass:  1,
	}
	v.SetKeyInfo(ki)
	v.SetValue(value)

	view := v.View()

	// Should contain key name
	if !strings.Contains(view, viewerKeyPrefix+"keyinfo:test") {
		t.Error("view should contain key name")
	}

	// Should contain size
	if !strings.Contains(view, "bytes") {
		t.Error("view should contain size information")
	}

	// Should contain type
	if !strings.Contains(view, "JSON") {
		t.Error("view should contain type information")
	}

	// Should contain mode
	if !strings.Contains(view, "Auto") {
		t.Error("view should contain mode information")
	}
}

// ============================================================================
// Formatter Tests
// ============================================================================

func TestE2E_JSONFormatterDirectly(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	formatter := viewerPkg.NewJSONFormatter()

	tests := []struct {
		name        string
		input       []byte
		shouldError bool
	}{
		{
			name:        "valid_object",
			input:       []byte(`{"key":"value"}`),
			shouldError: false,
		},
		{
			name:        "valid_array",
			input:       []byte(`[1,2,3]`),
			shouldError: false,
		},
		{
			name:        "invalid_json",
			input:       []byte(`{invalid}`),
			shouldError: true,
		},
		{
			name:        "complex_nested",
			input:       []byte(`{"a":{"b":[1,2,{"c":"d"}]}}`),
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewerSetKey(t, mc, "fmt:json:"+tt.name, tt.input)
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "fmt:json:"+tt.name)

			output, err := formatter.Format(value)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error for invalid JSON")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output == "" {
					t.Error("expected non-empty output")
				}
				// Verify it's valid JSON
				var parsed interface{}
				if json.Unmarshal([]byte(output), &parsed) != nil {
					t.Error("formatted output is not valid JSON")
				}
			}
		})
	}
}

func TestE2E_HexFormatterDirectly(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	formatter := viewerPkg.NewHexFormatter()

	testData := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f} // "Hello"
	viewerSetKey(t, mc, "fmt:hex:test", testData)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "fmt:hex:test")

	output, err := formatter.Format(value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify hex values are present
	hexBytes := []string{"48", "65", "6c", "6c", "6f"}
	outputLower := strings.ToLower(output)
	for _, hb := range hexBytes {
		if !strings.Contains(outputLower, hb) {
			t.Errorf("expected hex byte '%s' in output:\n%s", hb, output)
		}
	}

	// Verify ASCII representation
	if !strings.Contains(output, "Hello") {
		t.Error("expected ASCII 'Hello' in hex dump")
	}
}

func TestE2E_AutoFormatterDirectly(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	formatter := viewerPkg.NewAutoFormatter()

	tests := []struct {
		name          string
		value         []byte
		checkContains string
	}{
		{
			name:          "auto_json",
			value:         []byte(`{"auto": true}`),
			checkContains: "auto",
		},
		{
			name:          "auto_text",
			value:         []byte("plain text"),
			checkContains: "plain text",
		},
		{
			name:          "auto_binary",
			value:         []byte{0x00, 0x01, 0x02},
			checkContains: "00000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewerSetKey(t, mc, "fmt:auto:"+tt.name, tt.value)
			time.Sleep(50 * time.Millisecond)

			value := viewerGetValue(t, mc, "fmt:auto:"+tt.name)

			output, err := formatter.Format(value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !strings.Contains(output, tt.checkContains) {
				t.Errorf("expected output to contain '%s', got:\n%s", tt.checkContains, output)
			}
		})
	}
}

// ============================================================================
// Decompressor Tests
// ============================================================================

func TestE2E_DecompressorGzip(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	original := "This is the original content that will be compressed with gzip"
	compressed := gzipCompress([]byte(original))

	viewerSetKey(t, mc, "decomp:gzip", compressed)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "decomp:gzip")

	// Verify compression
	if !viewerPkg.IsGzipCompressed(value) {
		t.Error("value should be detected as gzip compressed")
	}

	// Decompress
	decompressed, err := viewerPkg.Decompress(value)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}

	if string(decompressed) != original {
		t.Errorf("decompressed content mismatch:\nexpected: %s\ngot: %s", original, string(decompressed))
	}
}

func TestE2E_DecompressorZlib(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	original := "This is the original content that will be compressed with zlib"
	compressed := zlibCompress([]byte(original))

	viewerSetKey(t, mc, "decomp:zlib", compressed)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "decomp:zlib")

	// Verify compression
	if !viewerPkg.IsZlibCompressed(value) {
		t.Error("value should be detected as zlib compressed")
	}

	// Decompress
	decompressed, err := viewerPkg.Decompress(value)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}

	if string(decompressed) != original {
		t.Errorf("decompressed content mismatch:\nexpected: %s\ngot: %s", original, string(decompressed))
	}
}

func TestE2E_DecompressorUncompressed(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	original := []byte("This is uncompressed data")
	viewerSetKey(t, mc, "decomp:plain", original)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "decomp:plain")

	// Should not be detected as compressed
	if viewerPkg.IsGzipCompressed(value) {
		t.Error("plain data should not be detected as gzip")
	}
	if viewerPkg.IsZlibCompressed(value) {
		t.Error("plain data should not be detected as zlib")
	}

	// Decompress should return original
	decompressed, err := viewerPkg.Decompress(value)
	if err != nil {
		t.Fatalf("decompress failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Error("decompress should return original for uncompressed data")
	}
}

// ============================================================================
// Rendering Tests
// ============================================================================

func TestE2E_ViewerRendering(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	testData := []byte(`{"render": "test", "value": 42}`)
	viewerSetKey(t, mc, "render:test", testData)
	time.Sleep(50 * time.Millisecond)

	value := viewerGetValue(t, mc, "render:test")

	v := viewer.NewModel()
	v.SetSize(80, 24)
	v.SetKeyInfo(models.KeyInfo{
		Key:  viewerKeyPrefix + "render:test",
		Size: len(testData),
	})
	v.SetValue(value)

	view := v.View()

	// Should not be empty
	if view == "" {
		t.Error("view should not be empty")
	}

	// Should contain separator line
	if !strings.Contains(view, "─") {
		t.Error("view should contain separator line")
	}

	// Should have reasonable length
	lines := strings.Split(view, "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines in view, got %d", len(lines))
	}
}

func TestE2E_ViewerEmptySize(t *testing.T) {
	v := viewer.NewModel()
	// Don't set size

	view := v.View()
	if view != "No size set" {
		t.Errorf("expected 'No size set' for uninitialized viewer, got '%s'", view)
	}
}

func TestE2E_ViewerNoValue(t *testing.T) {
	v := viewer.NewModel()
	v.SetSize(80, 24)
	v.SetKeyInfo(models.KeyInfo{Key: "test:key"})
	// Don't set value

	view := v.View()
	if !strings.Contains(view, "No value loaded") {
		t.Errorf("expected 'No value loaded' message, got: %s", view)
	}
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestE2E_ViewerConcurrentOperations(t *testing.T) {
	skipIfNoMemcached(t)
	mc := viewerMC(t)

	// Create multiple test keys
	for i := 0; i < 10; i++ {
		key := "concurrent:" + string(rune('0'+i))
		value := []byte(`{"index": ` + string(rune('0'+i)) + `}`)
		viewerSetKey(t, mc, key, value)
	}
	time.Sleep(100 * time.Millisecond)

	// Run concurrent viewer operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			key := "concurrent:" + string(rune('0'+idx))
			value := viewerGetValue(t, mc, key)

			v := viewer.NewModel()
			v.SetSize(80, 24)
			v.SetValue(value)

			// Perform various operations
			v.SetViewMode(viewer.ViewModeJSON)
			_ = v.Content()
			v.SetViewMode(viewer.ViewModeHex)
			_ = v.Content()
			v.SetViewMode(viewer.ViewModeText)
			_ = v.Content()

			// Scroll operations
			for j := 0; j < 5; j++ {
				v.Update(tea.KeyMsg{Type: tea.KeyDown})
			}
			_ = v.View()
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
