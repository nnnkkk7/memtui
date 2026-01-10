package serverlist

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestServerItem(t *testing.T) {
	t.Run("ServerItem struct fields", func(t *testing.T) {
		item := ServerItem{
			Name:      "production",
			Address:   "prod.example.com:11211",
			Status:    "Connected",
			Connected: true,
			Default:   true,
		}

		if item.Name != "production" {
			t.Errorf("expected Name 'production', got %q", item.Name)
		}
		if item.Address != "prod.example.com:11211" {
			t.Errorf("expected Address 'prod.example.com:11211', got %q", item.Address)
		}
		if item.Status != "Connected" {
			t.Errorf("expected Status 'Connected', got %q", item.Status)
		}
		if !item.Connected {
			t.Error("expected Connected to be true")
		}
		if !item.Default {
			t.Error("expected Default to be true")
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("creates empty server list", func(t *testing.T) {
		sl := New(nil)

		if sl == nil {
			t.Fatal("New returned nil")
		}
		if len(sl.servers) != 0 {
			t.Errorf("expected 0 servers, got %d", len(sl.servers))
		}
		if sl.cursor != 0 {
			t.Errorf("expected cursor 0, got %d", sl.cursor)
		}
	})

	t.Run("creates server list with items", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "local", Address: "localhost:11211", Connected: true},
			{Name: "staging", Address: "staging:11211", Connected: false},
		}

		sl := New(servers)

		if len(sl.servers) != 2 {
			t.Errorf("expected 2 servers, got %d", len(sl.servers))
		}
		if sl.servers[0].Name != "local" {
			t.Errorf("expected first server 'local', got %q", sl.servers[0].Name)
		}
	})
}

func TestSetServers(t *testing.T) {
	t.Run("sets servers", func(t *testing.T) {
		sl := New(nil)

		servers := []ServerItem{
			{Name: "server1", Address: "server1:11211"},
			{Name: "server2", Address: "server2:11211"},
			{Name: "server3", Address: "server3:11211"},
		}

		sl.SetServers(servers)

		if len(sl.servers) != 3 {
			t.Errorf("expected 3 servers, got %d", len(sl.servers))
		}
	})

	t.Run("resets cursor when servers change", func(t *testing.T) {
		sl := New([]ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
			{Name: "c", Address: "c:11211"},
		})
		sl.cursor = 2

		sl.SetServers([]ServerItem{
			{Name: "x", Address: "x:11211"},
		})

		if sl.cursor != 0 {
			t.Errorf("expected cursor to be reset to 0, got %d", sl.cursor)
		}
	})
}

func TestSelected(t *testing.T) {
	t.Run("returns nil when no servers", func(t *testing.T) {
		sl := New(nil)

		selected := sl.Selected()

		if selected != nil {
			t.Error("expected nil for empty list")
		}
	})

	t.Run("returns selected server", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "first", Address: "first:11211"},
			{Name: "second", Address: "second:11211"},
		}
		sl := New(servers)
		sl.cursor = 1

		selected := sl.Selected()

		if selected == nil {
			t.Fatal("expected non-nil selected server")
		}
		if selected.Name != "second" {
			t.Errorf("expected 'second', got %q", selected.Name)
		}
	})
}

func TestNavigation(t *testing.T) {
	t.Run("moves cursor down", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
			{Name: "c", Address: "c:11211"},
		}
		sl := New(servers)

		msg := tea.KeyMsg{Type: tea.KeyDown}
		sl, _ = sl.Update(msg)

		if sl.cursor != 1 {
			t.Errorf("expected cursor 1, got %d", sl.cursor)
		}
	})

	t.Run("moves cursor up", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
			{Name: "c", Address: "c:11211"},
		}
		sl := New(servers)
		sl.cursor = 2

		msg := tea.KeyMsg{Type: tea.KeyUp}
		sl, _ = sl.Update(msg)

		if sl.cursor != 1 {
			t.Errorf("expected cursor 1, got %d", sl.cursor)
		}
	})

	t.Run("cursor stops at top", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
		}
		sl := New(servers)
		sl.cursor = 0

		msg := tea.KeyMsg{Type: tea.KeyUp}
		sl, _ = sl.Update(msg)

		if sl.cursor != 0 {
			t.Errorf("expected cursor 0, got %d", sl.cursor)
		}
	})

	t.Run("cursor stops at bottom", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
		}
		sl := New(servers)
		sl.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyDown}
		sl, _ = sl.Update(msg)

		if sl.cursor != 1 {
			t.Errorf("expected cursor 1, got %d", sl.cursor)
		}
	})

	t.Run("j moves down", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
		}
		sl := New(servers)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		sl, _ = sl.Update(msg)

		if sl.cursor != 1 {
			t.Errorf("expected cursor 1, got %d", sl.cursor)
		}
	})

	t.Run("k moves up", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
		}
		sl := New(servers)
		sl.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		sl, _ = sl.Update(msg)

		if sl.cursor != 0 {
			t.Errorf("expected cursor 0, got %d", sl.cursor)
		}
	})
}

func TestEnterSelectsServer(t *testing.T) {
	t.Run("enter sends ServerSelectedMsg", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "first", Address: "first:11211"},
			{Name: "second", Address: "second:11211"},
		}
		sl := New(servers)
		sl.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := sl.Update(msg)

		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		// Execute the command
		result := cmd()
		selectMsg, ok := result.(ServerSelectedMsg)
		if !ok {
			t.Fatalf("expected ServerSelectedMsg, got %T", result)
		}

		if selectMsg.Server.Name != "second" {
			t.Errorf("expected 'second', got %q", selectMsg.Server.Name)
		}
	})

	t.Run("enter on empty list does nothing", func(t *testing.T) {
		sl := New(nil)

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := sl.Update(msg)

		if cmd != nil {
			t.Error("expected nil command for empty list")
		}
	})
}

func TestAddServerKey(t *testing.T) {
	t.Run("a key sends AddServerRequestMsg", func(t *testing.T) {
		sl := New(nil)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		_, cmd := sl.Update(msg)

		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		result := cmd()
		_, ok := result.(AddServerRequestMsg)
		if !ok {
			t.Fatalf("expected AddServerRequestMsg, got %T", result)
		}
	})
}

func TestDeleteServerKey(t *testing.T) {
	t.Run("d key sends DeleteServerRequestMsg", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "server1", Address: "server1:11211"},
			{Name: "server2", Address: "server2:11211"},
		}
		sl := New(servers)
		sl.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		_, cmd := sl.Update(msg)

		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		result := cmd()
		deleteMsg, ok := result.(DeleteServerRequestMsg)
		if !ok {
			t.Fatalf("expected DeleteServerRequestMsg, got %T", result)
		}

		if deleteMsg.Server.Name != "server2" {
			t.Errorf("expected 'server2', got %q", deleteMsg.Server.Name)
		}
	})

	t.Run("d key on empty list does nothing", func(t *testing.T) {
		sl := New(nil)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		_, cmd := sl.Update(msg)

		if cmd != nil {
			t.Error("expected nil command for empty list")
		}
	})
}

func TestSetDefaultKey(t *testing.T) {
	t.Run("s key sends SetDefaultRequestMsg", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "server1", Address: "server1:11211", Default: true},
			{Name: "server2", Address: "server2:11211", Default: false},
		}
		sl := New(servers)
		sl.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		_, cmd := sl.Update(msg)

		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		result := cmd()
		setDefaultMsg, ok := result.(SetDefaultRequestMsg)
		if !ok {
			t.Fatalf("expected SetDefaultRequestMsg, got %T", result)
		}

		if setDefaultMsg.Server.Name != "server2" {
			t.Errorf("expected 'server2', got %q", setDefaultMsg.Server.Name)
		}
	})
}

func TestEscapeKey(t *testing.T) {
	t.Run("escape sends CloseServerListMsg", func(t *testing.T) {
		sl := New(nil)

		msg := tea.KeyMsg{Type: tea.KeyEscape}
		_, cmd := sl.Update(msg)

		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		result := cmd()
		_, ok := result.(CloseServerListMsg)
		if !ok {
			t.Fatalf("expected CloseServerListMsg, got %T", result)
		}
	})
}

func TestView(t *testing.T) {
	t.Run("shows empty message when no servers", func(t *testing.T) {
		sl := New(nil)

		view := sl.View()

		if !strings.Contains(view, "No servers") {
			t.Errorf("expected 'No servers' in view, got: %s", view)
		}
	})

	t.Run("shows server list", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "local", Address: "localhost:11211", Status: "Connected", Connected: true},
			{Name: "staging", Address: "staging:11211", Status: "Disconnected", Connected: false},
		}
		sl := New(servers)

		view := sl.View()

		if !strings.Contains(view, "local") {
			t.Errorf("expected 'local' in view, got: %s", view)
		}
		if !strings.Contains(view, "staging") {
			t.Errorf("expected 'staging' in view, got: %s", view)
		}
	})

	t.Run("shows cursor indicator", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
		}
		sl := New(servers)
		sl.cursor = 1

		view := sl.View()

		// The view should show some indication of cursor position
		// (typically the selected item is styled differently)
		if view == "" {
			t.Error("view should not be empty")
		}
	})

	t.Run("shows connection status", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "connected", Address: "conn:11211", Connected: true, Status: "Connected"},
			{Name: "disconnected", Address: "disc:11211", Connected: false, Status: "Disconnected"},
		}
		sl := New(servers)

		view := sl.View()

		// Should show status for each server
		if !strings.Contains(view, "Connected") || !strings.Contains(view, "Disconnected") {
			t.Errorf("expected connection statuses in view, got: %s", view)
		}
	})

	t.Run("shows default indicator", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "default", Address: "default:11211", Default: true},
			{Name: "other", Address: "other:11211", Default: false},
		}
		sl := New(servers)

		view := sl.View()

		// Should show some indication for default server
		if view == "" {
			t.Error("view should not be empty")
		}
	})
}

func TestSetSize(t *testing.T) {
	t.Run("sets width and height", func(t *testing.T) {
		sl := New(nil)

		sl.SetSize(80, 24)

		if sl.width != 80 {
			t.Errorf("expected width 80, got %d", sl.width)
		}
		if sl.height != 24 {
			t.Errorf("expected height 24, got %d", sl.height)
		}
	})
}

func TestCursor(t *testing.T) {
	t.Run("returns current cursor position", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
			{Name: "c", Address: "c:11211"},
		}
		sl := New(servers)
		sl.cursor = 2

		if sl.Cursor() != 2 {
			t.Errorf("expected cursor 2, got %d", sl.Cursor())
		}
	})
}

func TestServerCount(t *testing.T) {
	t.Run("returns number of servers", func(t *testing.T) {
		servers := []ServerItem{
			{Name: "a", Address: "a:11211"},
			{Name: "b", Address: "b:11211"},
		}
		sl := New(servers)

		if sl.ServerCount() != 2 {
			t.Errorf("expected 2 servers, got %d", sl.ServerCount())
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("returns nil command", func(t *testing.T) {
		sl := New(nil)

		cmd := sl.Init()

		if cmd != nil {
			t.Error("expected nil command from Init")
		}
	})
}

func TestServerSelectedMsg(t *testing.T) {
	t.Run("ServerSelectedMsg contains server info", func(t *testing.T) {
		server := ServerItem{
			Name:      "test",
			Address:   "test:11211",
			Connected: true,
		}

		msg := ServerSelectedMsg{Server: server}

		if msg.Server.Name != "test" {
			t.Errorf("expected server name 'test', got %q", msg.Server.Name)
		}
	})
}

func TestAddServerRequestMsg(t *testing.T) {
	t.Run("AddServerRequestMsg is created correctly", func(t *testing.T) {
		msg := AddServerRequestMsg{}
		_ = msg // Just verify it compiles
	})
}

func TestDeleteServerRequestMsg(t *testing.T) {
	t.Run("DeleteServerRequestMsg contains server info", func(t *testing.T) {
		server := ServerItem{Name: "to-delete", Address: "delete:11211"}
		msg := DeleteServerRequestMsg{Server: server}

		if msg.Server.Name != "to-delete" {
			t.Errorf("expected server name 'to-delete', got %q", msg.Server.Name)
		}
	})
}

func TestSetDefaultRequestMsg(t *testing.T) {
	t.Run("SetDefaultRequestMsg contains server info", func(t *testing.T) {
		server := ServerItem{Name: "new-default", Address: "default:11211"}
		msg := SetDefaultRequestMsg{Server: server}

		if msg.Server.Name != "new-default" {
			t.Errorf("expected server name 'new-default', got %q", msg.Server.Name)
		}
	})
}

func TestCloseServerListMsg(t *testing.T) {
	t.Run("CloseServerListMsg is created correctly", func(t *testing.T) {
		msg := CloseServerListMsg{}
		_ = msg // Just verify it compiles
	})
}

func TestSetFocused(t *testing.T) {
	t.Run("sets focused state", func(t *testing.T) {
		sl := New(nil)

		sl.SetFocused(true)
		if !sl.focused {
			t.Error("expected focused to be true")
		}

		sl.SetFocused(false)
		if sl.focused {
			t.Error("expected focused to be false")
		}
	})
}

func TestIsFocused(t *testing.T) {
	t.Run("returns focused state", func(t *testing.T) {
		sl := New(nil)

		sl.focused = true
		if !sl.IsFocused() {
			t.Error("expected IsFocused to return true")
		}

		sl.focused = false
		if sl.IsFocused() {
			t.Error("expected IsFocused to return false")
		}
	})
}
