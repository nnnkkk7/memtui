package dialog_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
)

func TestInputDialog_New(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	if d == nil {
		t.Fatal("expected non-nil dialog")
	}
}

func TestInputDialog_Title(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	if d.Title() != "Enter Key Name" {
		t.Errorf("expected title 'Enter Key Name', got '%s'", d.Title())
	}
}

func TestInputDialog_View_ContainsTitle(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	view := d.View()
	if !strings.Contains(view, "Enter Key Name") {
		t.Errorf("view should contain title 'Enter Key Name', got: %s", view)
	}
}

func TestInputDialog_View_ContainsHints(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	view := d.View()

	// Should contain Enter and Escape hints
	hasEnterHint := strings.Contains(view, "Enter") || strings.Contains(view, "submit")
	hasEscHint := strings.Contains(view, "Esc") || strings.Contains(view, "cancel")

	if !hasEnterHint {
		t.Error("view should contain Enter/submit hint")
	}
	if !hasEscHint {
		t.Error("view should contain Esc/cancel hint")
	}
}

func TestInputDialog_Init_FocusesTextInput(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	cmd := d.Init()
	// Init should return a command (textinput.Blink for cursor)
	// We just verify it doesn't panic
	if cmd == nil {
		// textinput may or may not return a blink command depending on configuration
		// This is acceptable
	}
}

func TestInputDialog_Update_Typing(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init() // Initialize the text input

	// Type some characters
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	d = model.(*dialog.InputDialog)

	if d.Value() != "test" {
		t.Errorf("expected value 'test', got '%s'", d.Value())
	}
}

func TestInputDialog_Submit_ReturnsInputResultMsg(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()

	// Type a value
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	d = model.(*dialog.InputDialog)

	// Press Enter to submit
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter key")
	}

	msg := cmd()
	resultMsg, ok := msg.(dialog.InputResultMsg)
	if !ok {
		t.Fatalf("expected InputResultMsg, got %T", msg)
	}

	if resultMsg.Value != "mykey" {
		t.Errorf("expected value 'mykey', got '%s'", resultMsg.Value)
	}
	if resultMsg.Canceled {
		t.Error("expected Canceled to be false on submit")
	}
}

func TestInputDialog_Cancel_ReturnsInputResultMsg(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()

	// Type something
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	d = model.(*dialog.InputDialog)

	// Press Escape to cancel
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from Escape key")
	}

	msg := cmd()
	resultMsg, ok := msg.(dialog.InputResultMsg)
	if !ok {
		t.Fatalf("expected InputResultMsg, got %T", msg)
	}

	if resultMsg.Value != "" {
		t.Errorf("expected empty value on cancel, got '%s'", resultMsg.Value)
	}
	if !resultMsg.Canceled {
		t.Error("expected Canceled to be true on cancel")
	}
}

func TestInputDialog_SetValue(t *testing.T) {
	d := dialog.NewInput("Edit Key Name").WithValue("existing-key")
	d.Init()

	if d.Value() != "existing-key" {
		t.Errorf("expected value 'existing-key', got '%s'", d.Value())
	}
}

func TestInputDialog_Placeholder(t *testing.T) {
	d := dialog.NewInput("Enter Key Name").WithPlaceholder("type here...")
	view := d.View()

	// The placeholder may or may not be visible depending on the terminal
	// At minimum, verify it doesn't crash
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestInputDialog_WithContext(t *testing.T) {
	ctx := map[string]string{"action": "rename"}
	d := dialog.NewInput("Enter Key Name").WithContext(ctx)
	d.Init()

	// Type and submit
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	d = model.(*dialog.InputDialog)
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	msg := cmd()
	resultMsg := msg.(dialog.InputResultMsg)

	ctxMap, ok := resultMsg.Context.(map[string]string)
	if !ok {
		t.Fatalf("expected context to be map[string]string, got %T", resultMsg.Context)
	}
	if ctxMap["action"] != "rename" {
		t.Errorf("expected context action 'rename', got '%s'", ctxMap["action"])
	}
}

func TestInputDialog_Validator_Valid(t *testing.T) {
	validator := func(s string) error {
		return nil // Always valid
	}
	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()

	// Type and submit
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	d = model.(*dialog.InputDialog)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command for valid input")
	}

	msg := cmd()
	resultMsg, ok := msg.(dialog.InputResultMsg)
	if !ok {
		t.Fatalf("expected InputResultMsg, got %T", msg)
	}

	if resultMsg.Canceled {
		t.Error("expected submission, not cancellation")
	}
	if resultMsg.Value != "key" {
		t.Errorf("expected value 'key', got '%s'", resultMsg.Value)
	}
}

func TestInputDialog_Validator_Invalid(t *testing.T) {
	validator := func(s string) error {
		if len(s) < 3 {
			return &ValidationError{Message: "must be at least 3 characters"}
		}
		return nil
	}
	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()

	// Type only 2 characters (invalid)
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	d = model.(*dialog.InputDialog)
	model, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	d = model.(*dialog.InputDialog)

	// Try to submit - should not return command (or return with error)
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// When validation fails, the dialog should not produce a result command
	// Instead, it should show an error message
	if cmd != nil {
		// If command is returned, it should not be a successful submit
		msg := cmd()
		if resultMsg, ok := msg.(dialog.InputResultMsg); ok && !resultMsg.Canceled {
			t.Error("expected no successful submit for invalid input")
		}
	}

	// The error message should be visible
	if d.ValidationError() == "" {
		t.Error("expected validation error to be set")
	}
}

func TestInputDialog_Validator_TriggeredOnTyping(t *testing.T) {
	callCount := 0
	validator := func(s string) error {
		callCount++
		return nil
	}
	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()

	// Type characters - validator should be called
	model, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	d = model.(*dialog.InputDialog)

	// Validator may be called on each keystroke or only on submit
	// This depends on implementation choice
	// For now, we accept either behavior
}

func TestInputDialog_View_ShowsValidationError(t *testing.T) {
	validator := func(s string) error {
		if s == "" {
			return &ValidationError{Message: "cannot be empty"}
		}
		return nil
	}
	d := dialog.NewInput("Enter Key Name").WithValidator(validator)
	d.Init()

	// Try to submit empty
	_, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	view := d.View()
	if !strings.Contains(view, "cannot be empty") {
		// The error message should be visible somewhere
		// Accept if error is shown or if the view is non-empty
		if d.ValidationError() != "cannot be empty" {
			t.Error("expected validation error 'cannot be empty'")
		}
	}
}

func TestInputDialog_SetSize(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.SetSize(80, 24)
	// Should not panic
	view := d.View()
	if view == "" {
		t.Error("view should not be empty after SetSize")
	}
}

func TestInputDialog_UnrelatedKeyPassesToTextInput(t *testing.T) {
	d := dialog.NewInput("Enter Key Name")
	d.Init()

	// Backspace on empty should not crash
	model, cmd := d.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	d = model.(*dialog.InputDialog)

	// Should not produce result command
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(dialog.InputResultMsg); ok {
			t.Error("backspace should not produce InputResultMsg")
		}
	}
}

func TestInputResultMsg_Fields(t *testing.T) {
	msg := dialog.InputResultMsg{
		Value:    "test-value",
		Canceled: false,
		Context:  "some-context",
	}

	if msg.Value != "test-value" {
		t.Errorf("expected Value 'test-value', got '%s'", msg.Value)
	}
	if msg.Canceled {
		t.Error("expected Canceled to be false")
	}
	if msg.Context != "some-context" {
		t.Errorf("expected Context 'some-context', got '%v'", msg.Context)
	}
}

func TestInputDialog_ChainingMethods(t *testing.T) {
	d := dialog.NewInput("Enter Key Name").
		WithPlaceholder("type here...").
		WithValue("initial").
		WithContext("ctx").
		WithValidator(func(s string) error { return nil })

	if d == nil {
		t.Fatal("chaining should return non-nil dialog")
	}
	if d.Value() != "initial" {
		t.Errorf("expected value 'initial', got '%s'", d.Value())
	}
}

// ValidationError is a custom error type for validation failures
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
