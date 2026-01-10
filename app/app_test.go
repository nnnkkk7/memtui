package app_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/models"
)

func TestNewModel(t *testing.T) {
	m := app.NewModel("localhost:11211")
	if m == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestModel_Init(t *testing.T) {
	m := app.NewModel("localhost:11211")
	cmd := m.Init()
	// Init should return a command (typically to connect)
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModel_InitialState(t *testing.T) {
	m := app.NewModel("localhost:11211")
	if m.State() != app.StateConnecting {
		t.Errorf("expected StateConnecting, got %v", m.State())
	}
}

func TestModel_UpdateWindowSize(t *testing.T) {
	m := app.NewModel("localhost:11211")
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	newModel, _ := m.Update(msg)
	updated := newModel.(*app.Model)

	if updated.Width() != 120 || updated.Height() != 40 {
		t.Errorf("expected size 120x40, got %dx%d", updated.Width(), updated.Height())
	}
}

func TestModel_UpdateQuit(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestModel_UpdateConnected(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// Simulate connected message
	// Note: After connection, app immediately transitions to Loading state
	msg := app.ConnectedMsg{Version: "1.6.22"}
	newModel, _ := m.Update(msg)
	updated := newModel.(*app.Model)

	// State should be Loading (immediately starts loading keys after connect)
	if updated.State() != app.StateLoading {
		t.Errorf("expected StateLoading, got %v", updated.State())
	}
}

func TestModel_UpdateKeysLoaded(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// First connect
	m.Update(app.ConnectedMsg{Version: "1.6.22"})

	// Then load keys
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
	}
	msg := app.KeysLoadedMsg{Keys: keys}
	newModel, _ := m.Update(msg)
	updated := newModel.(*app.Model)

	if len(updated.Keys()) != 2 {
		t.Errorf("expected 2 keys, got %d", len(updated.Keys()))
	}
}

func TestModel_UpdateError(t *testing.T) {
	m := app.NewModel("localhost:11211")

	msg := app.ErrorMsg{Err: "connection failed"}
	newModel, _ := m.Update(msg)
	updated := newModel.(*app.Model)

	if updated.State() != app.StateError {
		t.Errorf("expected StateError, got %v", updated.State())
	}
	if updated.Error() != "connection failed" {
		t.Errorf("expected error message 'connection failed', got '%s'", updated.Error())
	}
}

func TestModel_View(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// Set window size first
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestAppState_String(t *testing.T) {
	tests := []struct {
		state    app.State
		expected string
	}{
		{app.StateConnecting, "Connecting"},
		{app.StateConnected, "Connected"},
		{app.StateLoading, "Loading"},
		{app.StateReady, "Ready"},
		{app.StateError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.state.String())
			}
		})
	}
}

func TestModel_TabSwitchesFocus(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// Set up ready state - chain updates properly
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(*app.Model)
	model, _ = m.Update(app.ConnectedMsg{Version: "1.6.22"})
	m = model.(*app.Model)
	model, _ = m.Update(app.KeysLoadedMsg{Keys: []models.KeyInfo{{Key: "test", Size: 10}}})
	m = model.(*app.Model)

	// Initial focus should be KeyList
	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected initial focus FocusKeyList, got %v", m.Focus())
	}

	// Press Tab to switch to Viewer
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	model, _ = m.Update(tabMsg)
	m = model.(*app.Model)

	if m.Focus() != app.FocusViewer {
		t.Errorf("expected FocusViewer after Tab, got %v", m.Focus())
	}

	// Press Tab again to switch back to KeyList
	model, _ = m.Update(tabMsg)
	m = model.(*app.Model)

	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList after second Tab, got %v", m.Focus())
	}
}

func TestModel_EscReturnsToKeyList(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// Set up ready state - chain updates properly
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(*app.Model)
	model, _ = m.Update(app.ConnectedMsg{Version: "1.6.22"})
	m = model.(*app.Model)
	model, _ = m.Update(app.KeysLoadedMsg{Keys: []models.KeyInfo{{Key: "test", Size: 10}}})
	m = model.(*app.Model)

	// Set focus to Viewer
	m.SetFocus(app.FocusViewer)

	if m.Focus() != app.FocusViewer {
		t.Fatalf("expected FocusViewer, got %v", m.Focus())
	}

	// Press Esc to return to KeyList
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	model, _ = m.Update(escMsg)
	m = model.(*app.Model)

	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList after Esc, got %v", m.Focus())
	}
}

func TestModel_EscDoesNothingInKeyList(t *testing.T) {
	m := app.NewModel("localhost:11211")

	// Set up ready state - chain updates properly
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(*app.Model)
	model, _ = m.Update(app.ConnectedMsg{Version: "1.6.22"})
	m = model.(*app.Model)
	model, _ = m.Update(app.KeysLoadedMsg{Keys: []models.KeyInfo{{Key: "test", Size: 10}}})
	m = model.(*app.Model)

	// Focus should be KeyList initially
	if m.Focus() != app.FocusKeyList {
		t.Fatalf("expected FocusKeyList, got %v", m.Focus())
	}

	// Press Esc - should stay in KeyList
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	model, _ = m.Update(escMsg)
	m = model.(*app.Model)

	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList to remain after Esc, got %v", m.Focus())
	}
}
