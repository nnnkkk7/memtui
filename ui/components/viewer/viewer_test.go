package viewer_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/viewer"
)

func TestNewModel(t *testing.T) {
	m := viewer.NewModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestModel_SetValue(t *testing.T) {
	m := viewer.NewModel()
	m.SetValue([]byte("hello world"))

	content := m.Content()
	if content != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", content)
	}
}

func TestModel_SetKeyInfo(t *testing.T) {
	m := viewer.NewModel()
	m.SetSize(80, 24)
	ki := models.KeyInfo{
		Key:        "user:1",
		Size:       100,
		Expiration: 1704067200,
	}
	m.SetKeyInfo(ki)
	m.SetValue([]byte("test value"))

	view := m.View()
	if !strings.Contains(view, "user:1") {
		t.Errorf("view should contain key name, got: %s", view)
	}
}

func TestModel_ViewModes(t *testing.T) {
	tests := []struct {
		mode     viewer.ViewMode
		expected string
	}{
		{viewer.ViewModeAuto, "Auto"},
		{viewer.ViewModeJSON, "JSON"},
		{viewer.ViewModeHex, "Hex"},
		{viewer.ViewModeText, "Text"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.mode.String() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.mode.String())
			}
		})
	}
}

func TestModel_SetViewMode(t *testing.T) {
	m := viewer.NewModel()
	m.SetValue([]byte(`{"key": "value"}`))

	m.SetViewMode(viewer.ViewModeJSON)
	if m.ViewMode() != viewer.ViewModeJSON {
		t.Errorf("expected ViewModeJSON, got %v", m.ViewMode())
	}

	m.SetViewMode(viewer.ViewModeHex)
	if m.ViewMode() != viewer.ViewModeHex {
		t.Errorf("expected ViewModeHex, got %v", m.ViewMode())
	}
}

func TestModel_JSONFormatting(t *testing.T) {
	m := viewer.NewModel()
	m.SetViewMode(viewer.ViewModeJSON)
	m.SetValue([]byte(`{"key":"value","nested":{"a":1}}`))

	content := m.Content()
	// Should be pretty-printed
	if !strings.Contains(content, "key") {
		t.Errorf("content should contain 'key', got: %s", content)
	}
}

func TestModel_HexFormatting(t *testing.T) {
	m := viewer.NewModel()
	m.SetViewMode(viewer.ViewModeHex)
	m.SetValue([]byte("Hello"))

	content := m.Content()
	// Should contain hex representation
	if !strings.Contains(strings.ToLower(content), "48") { // 'H' = 0x48
		t.Errorf("content should contain hex '48', got: %s", content)
	}
}

func TestModel_AutoDetection(t *testing.T) {
	m := viewer.NewModel()
	m.SetViewMode(viewer.ViewModeAuto)

	// JSON data should be detected as JSON
	m.SetValue([]byte(`{"key": "value"}`))
	detected := m.DetectedType()
	if detected != "JSON" {
		t.Errorf("expected 'JSON', got '%s'", detected)
	}

	// Binary data should be detected as binary
	m.SetValue([]byte{0x00, 0xFF, 0x0A})
	detected = m.DetectedType()
	if detected != "Binary" {
		t.Errorf("expected 'Binary', got '%s'", detected)
	}

	// Plain text
	m.SetValue([]byte("hello world"))
	detected = m.DetectedType()
	if detected != "Text" {
		t.Errorf("expected 'Text', got '%s'", detected)
	}
}

func TestModel_Scrolling(t *testing.T) {
	m := viewer.NewModel()
	m.SetSize(40, 10)

	// Create long content
	longContent := strings.Repeat("line\n", 100)
	m.SetValue([]byte(longContent))

	// Initial scroll offset should be 0
	if m.ScrollOffset() != 0 {
		t.Errorf("expected scroll offset 0, got %d", m.ScrollOffset())
	}

	// Scroll down
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.ScrollOffset() != 1 {
		t.Errorf("expected scroll offset 1, got %d", m.ScrollOffset())
	}

	// Scroll up
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.ScrollOffset() != 0 {
		t.Errorf("expected scroll offset 0, got %d", m.ScrollOffset())
	}
}

func TestModel_PageNavigation(t *testing.T) {
	m := viewer.NewModel()
	m.SetSize(40, 10)

	longContent := strings.Repeat("line\n", 100)
	m.SetValue([]byte(longContent))

	// Page down
	m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if m.ScrollOffset() < 5 {
		t.Errorf("expected scroll offset >= 5 after page down, got %d", m.ScrollOffset())
	}
}

func TestModel_SetSize(t *testing.T) {
	m := viewer.NewModel()
	m.SetSize(80, 24)

	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("view should not be empty after setting size")
	}
}

func TestModel_EmptyValue(t *testing.T) {
	m := viewer.NewModel()
	m.SetSize(40, 20)

	view := m.View()
	// Should show some placeholder
	if view == "" {
		t.Error("view should show empty state")
	}
}

func TestModel_KeyboardShortcuts(t *testing.T) {
	m := viewer.NewModel()
	m.SetValue([]byte(`{"key": "value"}`))

	// 'J' for JSON mode
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	if m.ViewMode() != viewer.ViewModeJSON {
		t.Errorf("expected ViewModeJSON after 'J', got %v", m.ViewMode())
	}

	// 'H' for Hex mode
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	if m.ViewMode() != viewer.ViewModeHex {
		t.Errorf("expected ViewModeHex after 'H', got %v", m.ViewMode())
	}

	// 'T' for Text mode
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
	if m.ViewMode() != viewer.ViewModeText {
		t.Errorf("expected ViewModeText after 'T', got %v", m.ViewMode())
	}

	// 'A' for Auto mode
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if m.ViewMode() != viewer.ViewModeAuto {
		t.Errorf("expected ViewModeAuto after 'A', got %v", m.ViewMode())
	}
}
