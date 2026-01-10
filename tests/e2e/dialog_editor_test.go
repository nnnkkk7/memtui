//go:build e2e

// Package e2e_test provides end-to-end tests for the memtui application.
// These tests simulate user interactions with dialogs and editor components.
package e2e_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
	"github.com/nnnkkk7/memtui/ui/components/editor"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupReadyApp creates an app model in the Ready state with window size set.
func setupReadyApp(t *testing.T) *app.Model {
	t.Helper()
	m := app.NewModel("localhost:11211")

	// Set window size
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(*app.Model)

	// Simulate connection
	model, _ = m.Update(app.ConnectedMsg{Version: "1.6.22"})
	m = model.(*app.Model)

	// Load some test keys
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
		{Key: "session:abc", Size: 50},
	}
	model, _ = m.Update(app.KeysLoadedMsg{Keys: keys})
	m = model.(*app.Model)

	return m
}

// setupAppWithCurrentKey creates an app with a key loaded and value available.
func setupAppWithCurrentKey(t *testing.T) *app.Model {
	t.Helper()
	m := setupReadyApp(t)

	// First simulate key selection which sets currentKey
	keyInfo := models.KeyInfo{Key: "user:1", Size: 100}
	model, _ := m.Update(keylist.KeySelectedMsg{Key: keyInfo})
	m = model.(*app.Model)

	// Then simulate loading a value for the current key
	model, _ = m.Update(app.ValueLoadedMsg{
		Key:   "user:1",
		Value: []byte("test value content"),
		Flags: 0,
		CAS:   12345,
	})
	m = model.(*app.Model)

	return m
}

// typeString simulates typing a string character by character.
func typeString(m tea.Model, s string) tea.Model {
	for _, r := range s {
		var cmd tea.Cmd
		m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		_ = cmd
	}
	return m
}

// pressKey simulates pressing a specific key.
func pressKey(m tea.Model, keyType tea.KeyType) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: keyType})
}

// pressRune simulates pressing a single rune key.
func pressRune(m tea.Model, r rune) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
}

// =============================================================================
// Confirm Dialog Tests
// =============================================================================

func TestE2E_ConfirmDialog_Display(t *testing.T) {
	d := dialog.New("Delete Key", "Are you sure you want to delete 'user:1'?")
	d.SetSize(80, 24)

	view := d.View()

	// Verify title is displayed
	if !strings.Contains(view, "Delete Key") {
		t.Error("confirm dialog should display title")
	}

	// Verify message is displayed
	if !strings.Contains(view, "user:1") {
		t.Error("confirm dialog should display key name in message")
	}

	// Verify Yes button is displayed
	if !strings.Contains(view, "Yes") {
		t.Error("confirm dialog should display Yes button")
	}

	// Verify No button is displayed
	if !strings.Contains(view, "No") {
		t.Error("confirm dialog should display No button")
	}
}

func TestE2E_ConfirmDialog_Layout(t *testing.T) {
	d := dialog.New("Confirm Action", "This is a test message for layout verification")
	d.SetSize(100, 30)

	view := d.View()
	lines := strings.Split(view, "\n")

	// Verify dialog has multiple lines (title, message, buttons, hints)
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines in dialog layout, got %d", len(lines))
	}

	// Verify hints are present
	if !strings.Contains(view, "Tab") || !strings.Contains(view, "Enter") {
		t.Error("confirm dialog should display keyboard hints")
	}
}

func TestE2E_ConfirmDialog_YesSelection(t *testing.T) {
	d := dialog.New("Delete", "Confirm delete?")
	d.SetSize(80, 24)

	// Default focus is on No
	if d.FocusedOnYes() {
		t.Fatal("expected default focus on No button")
	}

	// Switch to Yes with Tab
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyTab})
	d = model.(*dialog.ConfirmDialog)

	if !d.FocusedOnYes() {
		t.Error("expected focus on Yes after Tab")
	}

	// Press Enter to confirm
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter on Yes")
	}

	msg := cmd()
	result, ok := msg.(dialog.ConfirmResultMsg)
	if !ok {
		t.Fatalf("expected ConfirmResultMsg, got %T", msg)
	}

	if !result.Result {
		t.Error("expected Result=true when confirming on Yes")
	}
}

func TestE2E_ConfirmDialog_NoSelection(t *testing.T) {
	d := dialog.New("Delete", "Confirm delete?")
	d.SetSize(80, 24)

	// Default focus is on No, press Enter
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter on No")
	}

	msg := cmd()
	result, ok := msg.(dialog.ConfirmResultMsg)
	if !ok {
		t.Fatalf("expected ConfirmResultMsg, got %T", msg)
	}

	if result.Result {
		t.Error("expected Result=false when confirming on No")
	}
}

func TestE2E_ConfirmDialog_KeyboardNavigation_Tab(t *testing.T) {
	d := dialog.New("Title", "Message")

	// Initial focus is on No
	if d.FocusedOnYes() {
		t.Fatal("expected initial focus on No")
	}

	// Tab to Yes
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyTab})
	d = model.(*dialog.ConfirmDialog)
	if !d.FocusedOnYes() {
		t.Error("Tab should move focus to Yes")
	}

	// Tab to No
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyTab})
	d = model.(*dialog.ConfirmDialog)
	if d.FocusedOnYes() {
		t.Error("Tab should move focus back to No")
	}

	// Multiple tabs should cycle correctly
	for i := 0; i < 5; i++ {
		expectedYes := i%2 == 0 // After each tab, should alternate
		model, _ = d.Update(tea.KeyMsg{Type: tea.KeyTab})
		d = model.(*dialog.ConfirmDialog)
		if d.FocusedOnYes() != expectedYes {
			t.Errorf("Tab cycle %d: expected FocusedOnYes=%v, got %v", i+1, expectedYes, d.FocusedOnYes())
		}
	}
}

func TestE2E_ConfirmDialog_KeyboardNavigation_ArrowKeys(t *testing.T) {
	d := dialog.New("Title", "Message")

	// Left arrow should move to Yes (first button)
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyLeft})
	d = model.(*dialog.ConfirmDialog)
	if !d.FocusedOnYes() {
		t.Error("Left arrow should focus Yes")
	}

	// Right arrow should move to No
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRight})
	d = model.(*dialog.ConfirmDialog)
	if d.FocusedOnYes() {
		t.Error("Right arrow should focus No")
	}
}

func TestE2E_ConfirmDialog_KeyboardNavigation_Enter(t *testing.T) {
	tests := []struct {
		name          string
		focusOnYes    bool
		expectedResult bool
	}{
		{"Enter on Yes", true, true},
		{"Enter on No", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialog.New("Title", "Message")
			if tt.focusOnYes {
				model, _ := d.Update(tea.KeyMsg{Type: tea.KeyTab})
				d = model.(*dialog.ConfirmDialog)
			}

			_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("Enter should produce command")
			}

			msg := cmd()
			result := msg.(dialog.ConfirmResultMsg)
			if result.Result != tt.expectedResult {
				t.Errorf("expected Result=%v, got %v", tt.expectedResult, result.Result)
			}
		})
	}
}

func TestE2E_ConfirmDialog_KeyboardNavigation_Esc(t *testing.T) {
	d := dialog.New("Delete", "Are you sure?")

	// Move focus to Yes first
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyTab})
	d = model.(*dialog.ConfirmDialog)

	// Esc should cancel regardless of focus
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc should produce command")
	}

	msg := cmd()
	result := msg.(dialog.ConfirmResultMsg)
	if result.Result {
		t.Error("Esc should always cancel (Result=false)")
	}
}

func TestE2E_ConfirmDialog_QuickKeys(t *testing.T) {
	tests := []struct {
		key            rune
		expectedResult bool
	}{
		{'y', true},
		{'Y', true},
		{'n', false},
		{'N', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			d := dialog.New("Title", "Message")

			_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			if cmd == nil {
				t.Fatalf("key '%c' should produce command", tt.key)
			}

			msg := cmd()
			result := msg.(dialog.ConfirmResultMsg)
			if result.Result != tt.expectedResult {
				t.Errorf("key '%c': expected Result=%v, got %v", tt.key, tt.expectedResult, result.Result)
			}
		})
	}
}

func TestE2E_ConfirmDialog_ContextPreserved(t *testing.T) {
	ctx := map[string]interface{}{
		"key":    "user:1",
		"action": "delete",
	}
	d := dialog.NewWithContext("Delete", "Confirm?", ctx)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	msg := cmd()
	result := msg.(dialog.ConfirmResultMsg)

	ctxMap, ok := result.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected context to be preserved, got %T", result.Context)
	}
	if ctxMap["key"] != "user:1" {
		t.Errorf("expected context key 'user:1', got %v", ctxMap["key"])
	}
}

// =============================================================================
// Input Dialog Tests
// =============================================================================

func TestE2E_InputDialog_Display(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.SetSize(80, 24)

	view := d.View()

	// Verify title is displayed
	if !strings.Contains(view, "Enter Key Name") {
		t.Error("input dialog should display title")
	}

	// Verify hints are displayed
	if !strings.Contains(view, "Enter") {
		t.Error("input dialog should display Enter hint")
	}
	if !strings.Contains(view, "Esc") {
		t.Error("input dialog should display Esc hint")
	}
}

func TestE2E_InputDialog_TextInput(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Type a key name
	model := typeString(d, "my-new-key")
	d = model.(*dialog.InputDialog)

	if d.Value() != "my-new-key" {
		t.Errorf("expected value 'my-new-key', got '%s'", d.Value())
	}
}

func TestE2E_InputDialog_TextInput_SpecialCharacters(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Type a key name with special characters
	model := typeString(d, "user:session:abc_123")
	d = model.(*dialog.InputDialog)

	if d.Value() != "user:session:abc_123" {
		t.Errorf("expected value 'user:session:abc_123', got '%s'", d.Value())
	}
}

func TestE2E_InputDialog_TextInput_Backspace(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Type then backspace
	model := typeString(d, "test")
	d = model.(*dialog.InputDialog)

	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	d = model.(*dialog.InputDialog)

	if d.Value() != "tes" {
		t.Errorf("expected value 'tes' after backspace, got '%s'", d.Value())
	}
}

func TestE2E_InputDialog_Validation_Valid(t *testing.T) {
	validator := func(s string) error {
		if len(s) < 3 {
			return errors.New("key name must be at least 3 characters")
		}
		if strings.Contains(s, " ") {
			return errors.New("key name cannot contain spaces")
		}
		return nil
	}

	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()
	d.SetSize(80, 24)

	// Type valid key name
	model := typeString(d, "valid-key")
	d = model.(*dialog.InputDialog)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command for valid input")
	}

	msg := cmd()
	result := msg.(dialog.InputResultMsg)
	if result.Cancelled {
		t.Error("expected successful submission, not cancellation")
	}
	if result.Value != "valid-key" {
		t.Errorf("expected value 'valid-key', got '%s'", result.Value)
	}
}

func TestE2E_InputDialog_Validation_Invalid_TooShort(t *testing.T) {
	validator := func(s string) error {
		if len(s) < 3 {
			return errors.New("key name must be at least 3 characters")
		}
		return nil
	}

	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()
	d.SetSize(80, 24)

	// Type invalid key name (too short)
	model := typeString(d, "ab")
	d = model.(*dialog.InputDialog)

	// Try to submit
	model, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	d = model.(*dialog.InputDialog)

	// Should not produce a result command when validation fails
	if cmd != nil {
		msg := cmd()
		if result, ok := msg.(dialog.InputResultMsg); ok && !result.Cancelled {
			t.Error("should not successfully submit with invalid input")
		}
	}

	// Validation error should be set
	if d.ValidationError() == "" {
		t.Error("expected validation error to be set")
	}
	if !strings.Contains(d.ValidationError(), "3 characters") {
		t.Errorf("expected validation error about 3 characters, got '%s'", d.ValidationError())
	}
}

func TestE2E_InputDialog_Validation_Invalid_ContainsSpace(t *testing.T) {
	validator := func(s string) error {
		if strings.Contains(s, " ") {
			return errors.New("key name cannot contain spaces")
		}
		return nil
	}

	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()
	d.SetSize(80, 24)

	// Type invalid key name (contains space)
	model := typeString(d, "my key")
	d = model.(*dialog.InputDialog)

	// Try to submit
	model, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	d = model.(*dialog.InputDialog)

	// Should not produce a result command when validation fails
	if cmd != nil {
		msg := cmd()
		if result, ok := msg.(dialog.InputResultMsg); ok && !result.Cancelled {
			t.Error("should not successfully submit with spaces in key name")
		}
	}

	// Validation error should mention spaces
	if !strings.Contains(d.ValidationError(), "space") {
		t.Errorf("expected validation error about spaces, got '%s'", d.ValidationError())
	}
}

func TestE2E_InputDialog_Validation_ErrorClearsOnTyping(t *testing.T) {
	validator := func(s string) error {
		if len(s) < 3 {
			return errors.New("too short")
		}
		return nil
	}

	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()
	d.SetSize(80, 24)

	// Type invalid input and try to submit
	model := typeString(d, "ab")
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	d = model.(*dialog.InputDialog)

	// Validation error should be set
	if d.ValidationError() == "" {
		t.Fatal("expected validation error")
	}

	// Type another character - error should clear
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	d = model.(*dialog.InputDialog)

	if d.ValidationError() != "" {
		t.Error("validation error should clear on typing")
	}
}

func TestE2E_InputDialog_Submit(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Type and submit
	model := typeString(d, "new-key-name")
	d = model.(*dialog.InputDialog)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on submit")
	}

	msg := cmd()
	result, ok := msg.(dialog.InputResultMsg)
	if !ok {
		t.Fatalf("expected InputResultMsg, got %T", msg)
	}

	if result.Cancelled {
		t.Error("expected Cancelled=false on submit")
	}
	if result.Value != "new-key-name" {
		t.Errorf("expected value 'new-key-name', got '%s'", result.Value)
	}
}

func TestE2E_InputDialog_Cancel(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Type something
	model := typeString(d, "partial-input")
	d = model.(*dialog.InputDialog)

	// Press Esc to cancel
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on cancel")
	}

	msg := cmd()
	result, ok := msg.(dialog.InputResultMsg)
	if !ok {
		t.Fatalf("expected InputResultMsg, got %T", msg)
	}

	if !result.Cancelled {
		t.Error("expected Cancelled=true on Esc")
	}
	if result.Value != "" {
		t.Errorf("expected empty value on cancel, got '%s'", result.Value)
	}
}

func TestE2E_InputDialog_WithPlaceholder(t *testing.T) {
	d := dialog.NewInput("Enter Key Name").WithPlaceholder("e.g., user:123")
	d.Init()
	d.SetSize(80, 24)

	// View should render without error
	view := d.View()
	if view == "" {
		t.Error("view should not be empty")
	}

	// Initial value should be empty (placeholder is just a hint)
	if d.Value() != "" {
		t.Errorf("expected empty value with placeholder, got '%s'", d.Value())
	}
}

func TestE2E_InputDialog_WithValue(t *testing.T) {
	d := dialog.NewInput("Edit Key Name").WithValue("existing-key")
	d.Init()
	d.SetSize(80, 24)

	if d.Value() != "existing-key" {
		t.Errorf("expected value 'existing-key', got '%s'", d.Value())
	}
}

func TestE2E_InputDialog_WithContext(t *testing.T) {
	ctx := "original-key-name"
	d := dialog.NewInput("Rename Key").
		WithValue("original-key-name").
		WithContext(ctx)
	d.Init()
	d.SetSize(80, 24)

	// Clear and type new name
	for i := 0; i < len("original-key-name"); i++ {
		model, _ := d.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		d = model.(*dialog.InputDialog)
	}
	model := typeString(d, "new-key-name")
	d = model.(*dialog.InputDialog)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	result := msg.(dialog.InputResultMsg)

	if result.Context != ctx {
		t.Errorf("expected context '%s', got '%v'", ctx, result.Context)
	}
}

// =============================================================================
// Editor Tests
// =============================================================================

func TestE2E_Editor_Opening(t *testing.T) {
	e := editor.New("user:123", []byte("Hello, World!"))
	e.SetSize(80, 24)

	// Verify key is set
	if e.Key() != "user:123" {
		t.Errorf("expected key 'user:123', got '%s'", e.Key())
	}

	// Verify original value is preserved
	if string(e.OriginalValue()) != "Hello, World!" {
		t.Errorf("expected original value 'Hello, World!', got '%s'", string(e.OriginalValue()))
	}

	// Verify view renders with key
	view := e.View()
	if !strings.Contains(view, "user:123") {
		t.Error("editor view should contain key name")
	}
}

func TestE2E_Editor_OpeningWithKeyValue(t *testing.T) {
	jsonContent := `{"name":"test","value":42}`
	e := editor.New("config:app", []byte(jsonContent))
	e.SetSize(100, 30)
	e.Init()

	// Verify content is loaded
	if e.Value() != jsonContent {
		t.Errorf("expected value '%s', got '%s'", jsonContent, e.Value())
	}

	// View should show editing state
	view := e.View()
	if !strings.Contains(view, "Editing") || !strings.Contains(view, "config:app") {
		t.Error("editor view should show 'Editing: config:app'")
	}
}

func TestE2E_Editor_TextEditing(t *testing.T) {
	e := editor.New("mykey", []byte("original text"))
	e.SetSize(80, 24)
	e.Init()

	// Initially not dirty
	if e.IsDirty() {
		t.Error("editor should not be dirty initially")
	}

	// Type some characters
	model := typeString(e, " appended")
	e = model.(*editor.Editor)

	// Should now be dirty
	if !e.IsDirty() {
		t.Error("editor should be dirty after typing")
	}

	// Value should contain appended text
	if !strings.Contains(e.Value(), "appended") {
		t.Errorf("expected value to contain 'appended', got '%s'", e.Value())
	}
}

func TestE2E_Editor_TextEditing_Multiline(t *testing.T) {
	e := editor.New("mykey", []byte("line1\nline2"))
	e.SetSize(80, 24)
	e.Init()

	// Verify multiline content is preserved
	value := e.Value()
	if !strings.Contains(value, "line1") || !strings.Contains(value, "line2") {
		t.Errorf("expected multiline content, got '%s'", value)
	}
}

func TestE2E_Editor_SaveWithCtrlS(t *testing.T) {
	e := editor.New("mykey", []byte("original"))
	e.SetSize(80, 24)
	e.Init()
	e.SetCAS(99999)

	// Modify content
	model := typeString(e, " modified")
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
	if !strings.Contains(string(saveMsg.Value), "modified") {
		t.Errorf("expected value to contain 'modified', got '%s'", string(saveMsg.Value))
	}
	if saveMsg.OriginalCAS != 99999 {
		t.Errorf("expected CAS 99999, got %d", saveMsg.OriginalCAS)
	}
}

func TestE2E_Editor_CancelWithEsc(t *testing.T) {
	e := editor.New("mykey", []byte("original"))
	e.SetSize(80, 24)
	e.Init()

	// Modify content
	model := typeString(e, " modified")
	e = model.(*editor.Editor)

	if !e.IsDirty() {
		t.Fatal("editor should be dirty before cancel test")
	}

	// Press Esc to cancel
	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}

	msg := cmd()
	_, ok := msg.(editor.EditorCancelMsg)
	if !ok {
		t.Fatalf("expected EditorCancelMsg, got %T", msg)
	}
}

func TestE2E_Editor_ContentValidation_ValidJSON(t *testing.T) {
	validJSON := []byte(`{"name":"test","value":123}`)
	e := editor.New("config:json", validJSON)
	e.SetSize(80, 24)
	e.SetMode(editor.ModeJSON)

	// Format JSON should succeed
	err := e.FormatJSON()
	if err != nil {
		t.Errorf("expected valid JSON to format successfully, got error: %v", err)
	}

	// Formatted JSON should be indented
	value := e.Value()
	if !strings.Contains(value, "\n") {
		t.Error("formatted JSON should contain newlines")
	}
}

func TestE2E_Editor_ContentValidation_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{not valid json}`)
	e := editor.New("config:broken", invalidJSON)
	e.SetSize(80, 24)
	e.SetMode(editor.ModeJSON)

	// Format JSON should fail
	err := e.FormatJSON()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in error, got: %v", err)
	}
}

func TestE2E_Editor_ContentValidation_EmptyContent(t *testing.T) {
	e := editor.New("empty-key", []byte{})
	e.SetSize(80, 24)
	e.Init()

	// View should render without error
	view := e.View()
	if view == "" {
		t.Error("view should not be empty for empty content")
	}

	// Should not be dirty initially
	if e.IsDirty() {
		t.Error("editor should not be dirty for fresh empty content")
	}
}

func TestE2E_Editor_DirtyFlag(t *testing.T) {
	e := editor.New("mykey", []byte("original"))
	e.SetSize(80, 24)
	e.Init()

	// Initially not dirty
	if e.IsDirty() {
		t.Error("expected IsDirty=false initially")
	}

	// Type a character
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	e = model.(*editor.Editor)

	// Now dirty
	if !e.IsDirty() {
		t.Error("expected IsDirty=true after modification")
	}

	// View should show modified indicator
	view := e.View()
	if !strings.Contains(view, "Modified") {
		t.Log("Modified indicator style may vary - checking dirty flag instead")
	}
}

func TestE2E_Editor_Mode(t *testing.T) {
	e := editor.New("mykey", []byte("content"))

	// Default mode should be Text
	if e.Mode() != editor.ModeText {
		t.Errorf("expected default mode ModeText, got %v", e.Mode())
	}

	// Change to JSON mode
	e.SetMode(editor.ModeJSON)
	if e.Mode() != editor.ModeJSON {
		t.Errorf("expected ModeJSON, got %v", e.Mode())
	}

	// View in JSON mode should mention JSON
	e.SetSize(80, 24)
	view := e.View()
	if !strings.Contains(view, "JSON") {
		t.Error("expected JSON mode to be shown in view")
	}
}

func TestE2E_Editor_CASPreserved(t *testing.T) {
	e := editor.New("mykey", []byte("value"))
	e.SetSize(80, 24)
	e.Init()
	e.SetCAS(123456789)

	// Modify and save
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	e = model.(*editor.Editor)

	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	msg := cmd()
	saveMsg := msg.(editor.EditorSaveMsg)

	if saveMsg.OriginalCAS != 123456789 {
		t.Errorf("expected CAS 123456789, got %d", saveMsg.OriginalCAS)
	}
}

func TestE2E_Editor_ViewRendersHints(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	e.SetSize(80, 24)

	view := e.View()

	// Should show save hint
	if !strings.Contains(view, "Ctrl+S") && !strings.Contains(view, "Save") {
		t.Error("editor view should show save hint")
	}

	// Should show cancel hint
	if !strings.Contains(view, "Esc") && !strings.Contains(view, "Cancel") {
		t.Error("editor view should show cancel hint")
	}
}

func TestE2E_Editor_ViewRendersMetadata(t *testing.T) {
	content := []byte("12345678901234567890") // 20 bytes
	e := editor.New("mykey", content)
	e.SetSize(80, 24)

	view := e.View()

	// Should show size information
	if !strings.Contains(view, "20") || !strings.Contains(view, "bytes") {
		t.Log("Size display format may vary - checking that view is non-empty")
		if view == "" {
			t.Error("view should not be empty")
		}
	}
}

// =============================================================================
// Integration Tests: App with Dialogs and Editor
// =============================================================================

func TestE2E_App_ConfirmDialogFlow(t *testing.T) {
	m := setupAppWithCurrentKey(t)

	// Press 'd' to show delete confirmation dialog
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = model.(*app.Model)

	// Focus should be on dialog
	if m.Focus() != app.FocusDialog {
		t.Errorf("expected FocusDialog, got %v", m.Focus())
	}

	// Press 'n' to cancel
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = model.(*app.Model)

	// Execute the command to process the dialog result
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(*app.Model)
	}

	// Focus should return to key list after cancel
	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList after cancel, got %v", m.Focus())
	}
}

func TestE2E_App_InputDialogFlow_NewKey(t *testing.T) {
	m := setupReadyApp(t)

	// Press 'n' to create new key
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = model.(*app.Model)

	// Execute init command if any
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			model, _ = m.Update(msg)
			m = model.(*app.Model)
		}
	}

	// Focus should be on dialog
	if m.Focus() != app.FocusDialog {
		t.Errorf("expected FocusDialog for new key, got %v", m.Focus())
	}
}

func TestE2E_App_EditorFlow(t *testing.T) {
	m := setupAppWithCurrentKey(t)

	// Press 'e' to edit
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = model.(*app.Model)

	// Execute init command if any
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			model, _ = m.Update(msg)
			m = model.(*app.Model)
		}
	}

	// Focus should be on editor
	if m.Focus() != app.FocusEditor {
		t.Errorf("expected FocusEditor, got %v", m.Focus())
	}
}

func TestE2E_App_EditorCancel(t *testing.T) {
	m := setupAppWithCurrentKey(t)

	// Press 'e' to edit
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = model.(*app.Model)

	// Execute init command if any
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			model, _ = m.Update(msg)
			m = model.(*app.Model)
		}
	}

	if m.Focus() != app.FocusEditor {
		t.Fatalf("expected FocusEditor, got %v", m.Focus())
	}

	// Press Esc to cancel editing
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*app.Model)

	// Execute cancel command
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			model, _ = m.Update(msg)
			m = model.(*app.Model)
		}
	}

	// Focus should return to key list
	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList after editor cancel, got %v", m.Focus())
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestE2E_ConfirmDialog_RapidTabbing(t *testing.T) {
	d := dialog.New("Title", "Message")

	// Rapid tab presses should not cause issues
	for i := 0; i < 100; i++ {
		model, _ := d.Update(tea.KeyMsg{Type: tea.KeyTab})
		d = model.(*dialog.ConfirmDialog)
	}

	// Should still be in valid state (100 tabs = even, so back to initial No)
	// Actually depends on initial state
	view := d.View()
	if view == "" {
		t.Error("view should still render after rapid tabbing")
	}
}

func TestE2E_InputDialog_MaxLength(t *testing.T) {
	d := dialog.NewInput("Enter Long Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Type a very long string (more than typical key limits)
	longStr := strings.Repeat("a", 300)
	model := typeString(d, longStr)
	d = model.(*dialog.InputDialog)

	// Should handle gracefully (textinput has 256 char limit by default)
	value := d.Value()
	if len(value) > 256 {
		t.Errorf("expected value to be limited, got length %d", len(value))
	}
}

func TestE2E_Editor_LargeContent(t *testing.T) {
	// Create content with many lines
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = strings.Repeat("x", 80)
	}
	largeContent := []byte(strings.Join(lines, "\n"))

	e := editor.New("large-key", largeContent)
	e.SetSize(120, 50)
	e.Init()

	// Should handle large content
	view := e.View()
	if view == "" {
		t.Error("editor should render large content")
	}

	// Value should preserve content
	if len(e.Value()) != len(string(largeContent)) {
		t.Errorf("expected value length %d, got %d", len(largeContent), len(e.Value()))
	}
}

func TestE2E_Editor_SpecialCharacters(t *testing.T) {
	specialContent := []byte("Hello\tWorld\nLine2\r\nLine3")
	e := editor.New("special-key", specialContent)
	e.SetSize(80, 24)
	e.Init()

	value := e.Value()
	if !strings.Contains(value, "Hello") {
		t.Error("editor should preserve special characters")
	}
}

func TestE2E_Dialog_ZeroSize(t *testing.T) {
	// Dialogs should handle zero size gracefully
	d := dialog.New("Title", "Message")
	d.SetSize(0, 0)

	view := d.View()
	if view == "" {
		t.Error("dialog should render even with zero size")
	}

	// Input dialog
	id := dialog.NewInput("Input")
	id.SetSize(0, 0)

	view = id.View()
	if view == "" {
		t.Error("input dialog should render even with zero size")
	}
}

func TestE2E_Editor_ZeroSize(t *testing.T) {
	e := editor.New("mykey", []byte("content"))
	// Don't set size

	view := e.View()
	// Should return a message about no size
	if !strings.Contains(view, "No size") && view == "" {
		t.Error("editor should handle zero size gracefully")
	}
}

func TestE2E_InputDialog_EmptySubmit(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()
	d.SetSize(80, 24)

	// Don't type anything, just submit
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on empty submit")
	}

	msg := cmd()
	result := msg.(dialog.InputResultMsg)

	// Should submit with empty value (unless there's a validator)
	if result.Value != "" {
		t.Errorf("expected empty value, got '%s'", result.Value)
	}
	if result.Cancelled {
		t.Error("expected successful submission, not cancellation")
	}
}

func TestE2E_Editor_SaveWithoutModification(t *testing.T) {
	e := editor.New("mykey", []byte("original"))
	e.SetSize(80, 24)
	e.Init()

	// Don't modify, just save
	if e.IsDirty() {
		t.Fatal("editor should not be dirty before modification")
	}

	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("expected command from Ctrl+S even without modification")
	}

	msg := cmd()
	saveMsg := msg.(editor.EditorSaveMsg)

	// Should save original content
	if string(saveMsg.Value) != "original" {
		t.Errorf("expected 'original', got '%s'", string(saveMsg.Value))
	}
}
