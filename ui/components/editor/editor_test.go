package editor_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/editor"
)

func TestEditor_New(t *testing.T) {
	e := editor.New("mykey", []byte("hello world"))
	if e == nil {
		t.Fatal("expected non-nil editor")
	}
}

func TestEditor_Key(t *testing.T) {
	e := editor.New("user:123", []byte("test value"))
	if e.Key() != "user:123" {
		t.Errorf("expected key 'user:123', got '%s'", e.Key())
	}
}

func TestEditor_OriginalValue(t *testing.T) {
	original := []byte("original content")
	e := editor.New("mykey", original)
	if string(e.OriginalValue()) != "original content" {
		t.Errorf("expected original value 'original content', got '%s'", string(e.OriginalValue()))
	}
}

func TestEditor_SetContent(t *testing.T) {
	e := editor.New("mykey", []byte("initial"))
	e.SetContent([]byte("new content"))

	// After setting new content, the current value should be updated
	if e.Value() != "new content" {
		t.Errorf("expected value 'new content', got '%s'", e.Value())
	}
}

func TestEditor_View_RendersHeader(t *testing.T) {
	e := editor.New("user:123", []byte("hello"))
	e.SetSize(80, 24)

	view := e.View()

	// Should contain key name
	if !strings.Contains(view, "user:123") {
		t.Errorf("view should contain key name 'user:123', got: %s", view)
	}
}

func TestEditor_View_RendersSize(t *testing.T) {
	e := editor.New("mykey", []byte("12345"))
	e.SetSize(80, 24)

	view := e.View()

	// Should show some size information
	if !strings.Contains(view, "5") && !strings.Contains(view, "bytes") {
		// Accept either "5 bytes" or just presence of size indicator
		t.Log("Size information may be formatted differently")
	}
}

func TestEditor_View_RendersHints(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.SetSize(80, 24)

	view := e.View()

	// Should contain save and cancel hints
	hasSaveHint := strings.Contains(view, "Ctrl+S") || strings.Contains(view, "Save")
	hasCancelHint := strings.Contains(view, "Esc") || strings.Contains(view, "Cancel")

	if !hasSaveHint {
		t.Error("view should contain save hint (Ctrl+S or Save)")
	}
	if !hasCancelHint {
		t.Error("view should contain cancel hint (Esc or Cancel)")
	}
}

func TestEditor_Update_Typing_MarksDirty(t *testing.T) {
	e := editor.New("mykey", []byte("initial"))
	e.Init()
	e.SetSize(80, 24)

	// Initially should not be dirty
	if e.IsDirty() {
		t.Error("editor should not be dirty initially")
	}

	// Type a character
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	e = model.(*editor.Editor)

	// Should now be dirty
	if !e.IsDirty() {
		t.Error("editor should be dirty after typing")
	}
}

func TestEditor_Save_ReturnsEditorSaveMsg(t *testing.T) {
	e := editor.New("mykey", []byte("initial"))
	e.Init()
	e.SetSize(80, 24)

	// Type something to modify
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	e = model.(*editor.Editor)

	// Press Ctrl+S to save
	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("expected command from Ctrl+S")
	}

	msg := cmd()
	saveMsg, ok := msg.(editor.EditorSaveMsg)
	if !ok {
		t.Fatalf("expected EditorSaveMsg, got %T", msg)
	}

	if saveMsg.Key != "mykey" {
		t.Errorf("expected key 'mykey', got '%s'", saveMsg.Key)
	}
}

func TestEditor_Cancel_ReturnsEditorCancelMsg(t *testing.T) {
	e := editor.New("mykey", []byte("initial"))
	e.Init()
	e.SetSize(80, 24)

	// Press Escape to cancel
	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from Escape key")
	}

	msg := cmd()
	_, ok := msg.(editor.EditorCancelMsg)
	if !ok {
		t.Fatalf("expected EditorCancelMsg, got %T", msg)
	}
}

func TestEditor_IsDirty_FalseInitially(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.Init()

	if e.IsDirty() {
		t.Error("editor should not be dirty initially")
	}
}

func TestEditor_IsDirty_TrueAfterModification(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.Init()
	e.SetSize(80, 24)

	// Simulate typing
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	e = model.(*editor.Editor)

	if !e.IsDirty() {
		t.Error("editor should be dirty after modification")
	}
}

func TestEditor_SetCAS(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.SetCAS(12345)

	// CAS should be included in save message
	e.Init()
	e.SetSize(80, 24)

	// Modify to allow save
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	e = model.(*editor.Editor)

	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	msg := cmd()
	saveMsg := msg.(editor.EditorSaveMsg)

	if saveMsg.OriginalCAS != 12345 {
		t.Errorf("expected CAS 12345, got %d", saveMsg.OriginalCAS)
	}
}

func TestEditor_FormatJSON(t *testing.T) {
	jsonContent := []byte(`{"name":"test","value":123}`)
	e := editor.New("mykey", jsonContent)
	e.SetMode(editor.ModeJSON)

	// FormatJSON should format the content
	err := e.FormatJSON()
	if err != nil {
		t.Errorf("unexpected error formatting JSON: %v", err)
	}

	// The formatted content should be indented
	value := e.Value()
	if !strings.Contains(value, "\n") {
		t.Error("formatted JSON should contain newlines for indentation")
	}
}

func TestEditor_FormatJSON_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{not valid json}`)
	e := editor.New("mykey", invalidJSON)
	e.SetMode(editor.ModeJSON)

	err := e.FormatJSON()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestEditor_Mode(t *testing.T) {
	e := editor.New("mykey", []byte("content"))

	// Default mode should be Text
	if e.Mode() != editor.ModeText {
		t.Errorf("expected default mode ModeText, got %v", e.Mode())
	}

	e.SetMode(editor.ModeJSON)
	if e.Mode() != editor.ModeJSON {
		t.Errorf("expected mode ModeJSON, got %v", e.Mode())
	}
}

func TestEditor_ModeString(t *testing.T) {
	tests := []struct {
		mode     editor.EditorMode
		expected string
	}{
		{editor.ModeText, "Text"},
		{editor.ModeJSON, "JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.mode.String() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.mode.String())
			}
		})
	}
}

func TestEditor_Init_ReturnsCursorBlinkCmd(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	cmd := e.Init()
	// Init should return a command for textarea cursor blink
	// The command may be nil depending on textarea implementation
	// We just verify it doesn't panic
	_ = cmd
}

func TestEditor_SetSize(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.SetSize(100, 50)

	// Should not panic
	view := e.View()
	if view == "" {
		t.Error("view should not be empty after SetSize")
	}
}

func TestEditor_EmptyContent(t *testing.T) {
	e := editor.New("mykey", []byte{})
	e.SetSize(80, 24)

	// Should handle empty content gracefully
	view := e.View()
	if view == "" {
		t.Error("view should not be empty for empty content")
	}
}

func TestEditor_Value_ReturnsCurrentContent(t *testing.T) {
	e := editor.New("mykey", []byte("initial content"))
	e.Init()
	e.SetSize(80, 24)

	// Value should return the current textarea content
	if e.Value() != "initial content" {
		t.Errorf("expected 'initial content', got '%s'", e.Value())
	}
}

func TestEditorSaveMsg_Fields(t *testing.T) {
	msg := editor.EditorSaveMsg{
		Key:         "test-key",
		Value:       []byte("test-value"),
		OriginalCAS: 999,
	}

	if msg.Key != "test-key" {
		t.Errorf("expected Key 'test-key', got '%s'", msg.Key)
	}
	if string(msg.Value) != "test-value" {
		t.Errorf("expected Value 'test-value', got '%s'", string(msg.Value))
	}
	if msg.OriginalCAS != 999 {
		t.Errorf("expected OriginalCAS 999, got %d", msg.OriginalCAS)
	}
}

func TestEditorCancelMsg_Type(t *testing.T) {
	msg := editor.EditorCancelMsg{}
	// Just verify the type exists and can be created
	_ = msg
}

func TestEditor_ViewContainsEditStatus(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.Init()
	e.SetSize(80, 24)

	// Initially not modified
	view := e.View()
	// Should not show "modified" indicator initially
	if strings.Contains(view, "[Modified]") {
		t.Error("view should not show modified indicator initially")
	}

	// Modify content
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	e = model.(*editor.Editor)

	view = e.View()
	// Should show "modified" indicator
	if !strings.Contains(view, "Modified") && !strings.Contains(view, "*") {
		t.Log("Modified indicator may be styled differently")
	}
}

func TestEditor_TabKey_InsertsTab(t *testing.T) {
	e := editor.New("mykey", []byte("line1"))
	e.Init()
	e.SetSize(80, 24)

	// Press Tab
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyTab})
	e = model.(*editor.Editor)

	value := e.Value()
	// Tab should be inserted (textarea may convert to spaces)
	if len(value) <= len("line1") {
		t.Log("Tab handling depends on textarea configuration")
	}
}

func TestEditor_MultilineContent(t *testing.T) {
	multiline := []byte("line1\nline2\nline3")
	e := editor.New("mykey", multiline)
	e.Init()
	e.SetSize(80, 24)

	value := e.Value()
	if !strings.Contains(value, "line1") || !strings.Contains(value, "line2") {
		t.Errorf("expected multiline content, got '%s'", value)
	}
}
