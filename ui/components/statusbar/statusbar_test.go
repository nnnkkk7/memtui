package statusbar_test

import (
	"strings"
	"testing"

	"github.com/nnnkkk7/memtui/ui/components/statusbar"
)

func TestNewModel(t *testing.T) {
	m := statusbar.NewModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestModel_SetAddress(t *testing.T) {
	m := statusbar.NewModel()
	m.SetAddress("localhost:11211")

	view := m.View()
	if !strings.Contains(view, "localhost:11211") {
		t.Errorf("view should contain address, got: %s", view)
	}
}

func TestModel_SetVersion(t *testing.T) {
	m := statusbar.NewModel()
	m.SetVersion("1.6.22")

	view := m.View()
	if !strings.Contains(view, "1.6.22") {
		t.Errorf("view should contain version, got: %s", view)
	}
}

func TestModel_SetKeyCount(t *testing.T) {
	m := statusbar.NewModel()
	m.SetKeyCount(42)

	view := m.View()
	if !strings.Contains(view, "42") {
		t.Errorf("view should contain key count, got: %s", view)
	}
}

func TestModel_SetStatus(t *testing.T) {
	tests := []struct {
		status   statusbar.Status
		expected string
	}{
		{statusbar.StatusConnecting, "Connecting"},
		{statusbar.StatusConnected, "Connected"},
		{statusbar.StatusLoading, "Loading"},
		{statusbar.StatusReady, "Ready"},
		{statusbar.StatusError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			m := statusbar.NewModel()
			m.SetStatus(tt.status)

			view := m.View()
			if !strings.Contains(view, tt.expected) {
				t.Errorf("view should contain '%s', got: %s", tt.expected, view)
			}
		})
	}
}

func TestModel_SetError(t *testing.T) {
	m := statusbar.NewModel()
	m.SetError("connection refused")

	view := m.View()
	if !strings.Contains(view, "connection refused") {
		t.Errorf("view should contain error message, got: %s", view)
	}
}

func TestModel_SetWidth(t *testing.T) {
	m := statusbar.NewModel()
	m.SetWidth(80)

	view := m.View()
	// View should be rendered (not empty)
	if view == "" {
		t.Error("view should not be empty after setting width")
	}
}

func TestModel_FullInfo(t *testing.T) {
	m := statusbar.NewModel()
	m.SetAddress("localhost:11211")
	m.SetVersion("1.6.22")
	m.SetKeyCount(100)
	m.SetStatus(statusbar.StatusReady)
	m.SetWidth(80)

	view := m.View()

	checks := []string{"localhost:11211", "1.6.22", "100", "Ready"}
	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("view should contain '%s', got: %s", check, view)
		}
	}
}
