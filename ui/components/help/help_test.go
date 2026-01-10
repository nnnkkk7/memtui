package help_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/help"
)

func TestNewModel(t *testing.T) {
	m := help.NewModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestModel_InitiallyHidden(t *testing.T) {
	m := help.NewModel()
	if m.Visible() {
		t.Error("help overlay should be hidden by default")
	}
}

func TestModel_Show(t *testing.T) {
	m := help.NewModel()
	m.Show()
	if !m.Visible() {
		t.Error("expected help overlay to be visible after Show()")
	}
}

func TestModel_Hide(t *testing.T) {
	m := help.NewModel()
	m.Show()
	m.Hide()
	if m.Visible() {
		t.Error("expected help overlay to be hidden after Hide()")
	}
}

func TestModel_Toggle(t *testing.T) {
	m := help.NewModel()

	// Initially hidden
	if m.Visible() {
		t.Error("expected help overlay to be hidden initially")
	}

	// Toggle to show
	m.Toggle()
	if !m.Visible() {
		t.Error("expected help overlay to be visible after first toggle")
	}

	// Toggle to hide
	m.Toggle()
	if m.Visible() {
		t.Error("expected help overlay to be hidden after second toggle")
	}
}

func TestModel_QuestionMarkKeyToggle(t *testing.T) {
	m := help.NewModel()

	// Press '?' to show
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.Visible() {
		t.Error("expected help overlay to be visible after pressing '?'")
	}

	// Press '?' again to hide
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m.Visible() {
		t.Error("expected help overlay to be hidden after pressing '?' again")
	}
}

func TestModel_EscapeKeyHides(t *testing.T) {
	m := help.NewModel()
	m.Show()

	// Press Escape to hide
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.Visible() {
		t.Error("expected help overlay to be hidden after pressing Escape")
	}
}

func TestModel_SetSize(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 24)

	// Should not panic
	m.Show()
	view := m.View()
	if view == "" {
		t.Error("view should not be empty after setting size")
	}
}

func TestModel_ViewContainsKeyBindings(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 40)
	m.Show()

	view := m.View()

	// Global keybindings from Appendix A
	globalBindings := []string{
		"q", "Ctrl+C", "?", "Tab", "r", "s",
	}
	for _, key := range globalBindings {
		if !strings.Contains(view, key) {
			t.Errorf("view should contain global keybinding '%s', got: %s", key, view)
		}
	}
}

func TestModel_ViewContainsKeyListBindings(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 40)
	m.Show()

	view := m.View()

	// Keylist pane keybindings from Appendix A
	keylistBindings := []string{
		"j", "k", "Enter", "l", "h", "/", "d", "n", "m",
	}
	for _, key := range keylistBindings {
		if !strings.Contains(view, key) {
			t.Errorf("view should contain keylist keybinding '%s', got: %s", key, view)
		}
	}
}

func TestModel_ViewContainsViewerBindings(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 40)
	m.Show()

	view := m.View()

	// Viewer pane keybindings from Appendix A
	viewerBindings := []string{
		"e", "J", "H", "T", "A", "c",
	}
	for _, key := range viewerBindings {
		if !strings.Contains(view, key) {
			t.Errorf("view should contain viewer keybinding '%s', got: %s", key, view)
		}
	}
}

func TestModel_ViewContainsSectionHeaders(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 40)
	m.Show()

	view := m.View()

	// Should contain section headers
	sections := []string{
		"Global",
		"Key List",
		"Viewer",
	}
	for _, section := range sections {
		if !strings.Contains(view, section) {
			t.Errorf("view should contain section '%s', got: %s", section, view)
		}
	}
}

func TestModel_ViewWhenHidden(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 24)
	// Don't show

	view := m.View()
	if view != "" {
		t.Errorf("view should be empty when hidden, got: %s", view)
	}
}

func TestModel_KeyBindingCount(t *testing.T) {
	m := help.NewModel()
	bindings := m.KeyBindings()

	// Based on Appendix A, we should have:
	// Global: 5 bindings (q/Ctrl+C, ?, Tab, r, s)
	// KeyList: 8 bindings
	// Viewer: 8 bindings
	// Total: at least 20 bindings
	if len(bindings) < 20 {
		t.Errorf("expected at least 20 keybindings, got %d", len(bindings))
	}
}

func TestModel_KeyBindingStructure(t *testing.T) {
	m := help.NewModel()
	bindings := m.KeyBindings()

	for i, binding := range bindings {
		if binding.Key == "" {
			t.Errorf("binding %d has empty key", i)
		}
		if binding.Action == "" {
			t.Errorf("binding %d has empty action", i)
		}
		if binding.Category == "" {
			t.Errorf("binding %d has empty category", i)
		}
	}
}

func TestKeyBinding_GlobalCategory(t *testing.T) {
	m := help.NewModel()
	bindings := m.KeyBindings()

	globalCount := 0
	for _, b := range bindings {
		if b.Category == help.CategoryGlobal {
			globalCount++
		}
	}

	// Should have at least 5 global bindings
	if globalCount < 5 {
		t.Errorf("expected at least 5 global bindings, got %d", globalCount)
	}
}

func TestKeyBinding_KeyListCategory(t *testing.T) {
	m := help.NewModel()
	bindings := m.KeyBindings()

	keylistCount := 0
	for _, b := range bindings {
		if b.Category == help.CategoryKeyList {
			keylistCount++
		}
	}

	// Should have at least 8 keylist bindings
	if keylistCount < 8 {
		t.Errorf("expected at least 8 keylist bindings, got %d", keylistCount)
	}
}

func TestKeyBinding_ViewerCategory(t *testing.T) {
	m := help.NewModel()
	bindings := m.KeyBindings()

	viewerCount := 0
	for _, b := range bindings {
		if b.Category == help.CategoryViewer {
			viewerCount++
		}
	}

	// Should have at least 8 viewer bindings
	if viewerCount < 8 {
		t.Errorf("expected at least 8 viewer bindings, got %d", viewerCount)
	}
}

func TestModel_ViewOverlayStyle(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 24)
	m.Show()

	view := m.View()

	// View should have some content that looks like an overlay
	// It should contain the title
	if !strings.Contains(view, "Help") && !strings.Contains(view, "Keybindings") {
		t.Errorf("view should contain title 'Help' or 'Keybindings', got: %s", view)
	}
}

func TestModel_ViewContainsCloseHint(t *testing.T) {
	m := help.NewModel()
	m.SetSize(80, 24)
	m.Show()

	view := m.View()

	// Should show how to close the help
	if !strings.Contains(view, "?") && !strings.Contains(view, "Esc") {
		t.Errorf("view should contain close hint (? or Esc), got: %s", view)
	}
}

func TestModel_Update_UnrelatedKeyDoesNotToggle(t *testing.T) {
	m := help.NewModel()
	m.Show()

	// Press an unrelated key
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if !m.Visible() {
		t.Error("unrelated key should not hide help overlay")
	}
}

func TestModel_Update_ReturnsCmd(t *testing.T) {
	m := help.NewModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	// Cmd can be nil for simple toggle operations
	_ = cmd // Just verify it doesn't panic
}
