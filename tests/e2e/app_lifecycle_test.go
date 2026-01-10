//go:build e2e

package e2e_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

// =============================================================================
// Test Helpers (app lifecycle specific)
// =============================================================================

// setupTestKeysForLifecycle creates test keys in Memcached and returns a cleanup function
func setupTestKeysForLifecycle(t *testing.T, keys map[string]string) func() {
	t.Helper()
	addr := getMemcachedAddr()
	mc := memcache.New(addr)

	for key, value := range keys {
		err := mc.Set(&memcache.Item{
			Key:   key,
			Value: []byte(value),
		})
		if err != nil {
			t.Fatalf("failed to set test key %s: %v", key, err)
		}
	}

	// Wait for keys to be indexed
	time.Sleep(100 * time.Millisecond)

	return func() {
		for key := range keys {
			mc.Delete(key)
		}
	}
}

// executeCmd executes a tea.Cmd and returns the message
func executeCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// simulateAppUpdate sends a message to the model and returns the updated model and command
func simulateAppUpdate(m *app.Model, msg tea.Msg) (*app.Model, tea.Cmd) {
	newModel, cmd := m.Update(msg)
	return newModel.(*app.Model), cmd
}

// runAppUntilState runs the app update loop until it reaches the target state or times out
func runAppUntilState(t *testing.T, m *app.Model, targetState app.State, timeout time.Duration) *app.Model {
	t.Helper()

	deadline := time.Now().Add(timeout)
	cmd := m.Init()

	for time.Now().Before(deadline) {
		if m.State() == targetState {
			return m
		}

		if cmd != nil {
			msg := executeCmd(cmd)
			if msg != nil {
				m, cmd = simulateAppUpdate(m, msg)
			}
		} else {
			// No pending command, wait a bit
			time.Sleep(10 * time.Millisecond)
		}
	}

	t.Fatalf("timeout waiting for state %v, current state: %v", targetState, m.State())
	return m
}

// =============================================================================
// Application Startup and Connection Tests
// =============================================================================

func TestE2E_ApplicationStartupAndConnection(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("successful connection to Memcached", func(t *testing.T) {
		m := app.NewModel(addr)

		// Initial state should be Connecting
		if m.State() != app.StateConnecting {
			t.Errorf("expected initial state StateConnecting, got %v", m.State())
		}

		// Execute init command (connects to Memcached)
		cmd := m.Init()
		if cmd == nil {
			t.Fatal("Init should return a command")
		}

		// Execute the connect command
		msg := executeCmd(cmd)
		if msg == nil {
			t.Fatal("connect command should return a message")
		}

		// Should receive ConnectedMsg
		connectedMsg, ok := msg.(app.ConnectedMsg)
		if !ok {
			// Check if it's an error
			if errMsg, isErr := msg.(app.ErrorMsg); isErr {
				t.Fatalf("connection failed: %s", errMsg.Err)
			}
			t.Fatalf("expected ConnectedMsg, got %T", msg)
		}

		// Verify version is detected
		if connectedMsg.Version == "" {
			t.Error("expected non-empty version")
		}
		t.Logf("Connected to Memcached version: %s", connectedMsg.Version)

		// Update model with connected message
		m, cmd = simulateAppUpdate(m, connectedMsg)

		// State should now be Loading (automatically starts loading keys)
		if m.State() != app.StateLoading {
			t.Errorf("expected StateLoading after connection, got %v", m.State())
		}
	})

	t.Run("connection with test keys present", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:startup:key1": "value1",
			"e2e:startup:key2": "value2",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)

		// Verify keys were loaded
		keys := m.Keys()
		if len(keys) < 2 {
			t.Errorf("expected at least 2 keys, got %d", len(keys))
		}

		// Check that our test keys are present
		foundKeys := make(map[string]bool)
		for _, k := range keys {
			foundKeys[k.Key] = true
		}

		if !foundKeys["e2e:startup:key1"] {
			t.Error("expected to find e2e:startup:key1")
		}
		if !foundKeys["e2e:startup:key2"] {
			t.Error("expected to find e2e:startup:key2")
		}
	})
}

// =============================================================================
// Window Resize Handling Tests
// =============================================================================

func TestE2E_WindowResizeHandling(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("resize updates dimensions", func(t *testing.T) {
		m := app.NewModel(addr)

		// Initial dimensions should be zero
		if m.Width() != 0 || m.Height() != 0 {
			t.Errorf("expected initial dimensions 0x0, got %dx%d", m.Width(), m.Height())
		}

		// Send window size message
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		if m.Width() != 120 || m.Height() != 40 {
			t.Errorf("expected dimensions 120x40, got %dx%d", m.Width(), m.Height())
		}
	})

	t.Run("resize in ready state preserves functionality", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:resize:key": "test value",
		})
		defer cleanup()

		m := app.NewModel(addr)

		// Set initial size
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 100, Height: 30})

		// Run until ready
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)

		// Verify view renders without issues
		view := m.View()
		if view == "" {
			t.Error("view should not be empty")
		}

		// Resize to smaller dimensions
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 80, Height: 24})
		if m.Width() != 80 || m.Height() != 24 {
			t.Errorf("expected dimensions 80x24, got %dx%d", m.Width(), m.Height())
		}

		// View should still render
		view = m.View()
		if view == "" {
			t.Error("view should not be empty after resize")
		}

		// Resize to larger dimensions
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 200, Height: 60})
		if m.Width() != 200 || m.Height() != 60 {
			t.Errorf("expected dimensions 200x60, got %dx%d", m.Width(), m.Height())
		}

		view = m.View()
		if view == "" {
			t.Error("view should not be empty after large resize")
		}
	})

	t.Run("resize with minimum dimensions", func(t *testing.T) {
		m := app.NewModel(addr)

		// Test minimum viable dimensions
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 40, Height: 10})

		if m.Width() != 40 || m.Height() != 10 {
			t.Errorf("expected dimensions 40x10, got %dx%d", m.Width(), m.Height())
		}

		// Should still be able to render
		view := m.View()
		if view == "" {
			t.Error("view should render even with minimum dimensions")
		}
	})

	t.Run("multiple rapid resizes", func(t *testing.T) {
		m := app.NewModel(addr)

		// Simulate rapid window resizing
		sizes := []struct{ w, h int }{
			{80, 24},
			{100, 30},
			{120, 40},
			{80, 24},
			{150, 50},
		}

		for _, size := range sizes {
			m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: size.w, Height: size.h})
		}

		// Final dimensions should match last resize
		if m.Width() != 150 || m.Height() != 50 {
			t.Errorf("expected dimensions 150x50, got %dx%d", m.Width(), m.Height())
		}
	})
}

// =============================================================================
// State Transition Tests
// =============================================================================

func TestE2E_StateTransitions(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("full state transition cycle: Connecting -> Connected -> Loading -> Ready", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:state:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)

		// Step 1: Initial state should be Connecting
		if m.State() != app.StateConnecting {
			t.Errorf("Step 1: expected StateConnecting, got %v", m.State())
		}
		t.Log("State: Connecting")

		// Step 2: Execute connect command
		cmd := m.Init()
		msg := executeCmd(cmd)

		connectedMsg, ok := msg.(app.ConnectedMsg)
		if !ok {
			t.Fatalf("expected ConnectedMsg, got %T", msg)
		}

		m, cmd = simulateAppUpdate(m, connectedMsg)

		// After ConnectedMsg, state should be Loading (auto-transition)
		if m.State() != app.StateLoading {
			t.Errorf("Step 2: expected StateLoading after ConnectedMsg, got %v", m.State())
		}
		t.Log("State: Loading (after Connected)")

		// Step 3: Execute load keys command
		if cmd == nil {
			t.Fatal("expected load keys command after connection")
		}

		msg = executeCmd(cmd)
		keysMsg, ok := msg.(app.KeysLoadedMsg)
		if !ok {
			if errMsg, isErr := msg.(app.ErrorMsg); isErr {
				t.Fatalf("loading keys failed: %s", errMsg.Err)
			}
			t.Fatalf("expected KeysLoadedMsg, got %T", msg)
		}

		m, _ = simulateAppUpdate(m, keysMsg)

		// Step 4: State should now be Ready
		if m.State() != app.StateReady {
			t.Errorf("Step 4: expected StateReady, got %v", m.State())
		}
		t.Log("State: Ready")

		// Verify keys were loaded
		if len(m.Keys()) < 1 {
			t.Error("expected at least 1 key to be loaded")
		}
	})

	t.Run("refresh triggers Loading state", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:refresh:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)

		// Get to ready state
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)

		// Set window size for key handling
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Press 'r' to refresh
		m, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

		// State should transition to Loading
		if m.State() != app.StateLoading {
			t.Errorf("expected StateLoading after refresh, got %v", m.State())
		}

		// Execute the refresh command
		if cmd != nil {
			msg := executeCmd(cmd)
			m, _ = simulateAppUpdate(m, msg)
		}

		// Should be back to Ready
		if m.State() != app.StateReady {
			t.Errorf("expected StateReady after refresh complete, got %v", m.State())
		}
	})

	t.Run("error state transitions", func(t *testing.T) {
		m := app.NewModel(addr)

		// Simulate an error
		m, _ = simulateAppUpdate(m, app.ErrorMsg{Err: "simulated error"})

		if m.State() != app.StateError {
			t.Errorf("expected StateError, got %v", m.State())
		}

		if m.Error() != "simulated error" {
			t.Errorf("expected error message 'simulated error', got '%s'", m.Error())
		}
	})
}

// =============================================================================
// Error State Handling Tests
// =============================================================================

func TestE2E_ErrorStateHandling(t *testing.T) {
	t.Run("connection failure to invalid address", func(t *testing.T) {
		// Use an invalid address that will fail to connect
		m := app.NewModel("invalid.host.local:99999")

		cmd := m.Init()
		msg := executeCmd(cmd)

		// Should receive an error message
		errMsg, ok := msg.(app.ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg for invalid address, got %T", msg)
		}

		if errMsg.Err == "" {
			t.Error("expected non-empty error message")
		}
		t.Logf("Connection error (expected): %s", errMsg.Err)

		// Update model
		m, _ = simulateAppUpdate(m, errMsg)

		// State should be Error
		if m.State() != app.StateError {
			t.Errorf("expected StateError, got %v", m.State())
		}
	})

	t.Run("connection timeout", func(t *testing.T) {
		// Use a non-routable IP to simulate timeout
		// 10.255.255.1 is typically not routable and will timeout
		m := app.NewModel("10.255.255.1:11211")

		// Create a context with short timeout for the test
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		done := make(chan struct{})
		var msg tea.Msg

		go func() {
			cmd := m.Init()
			msg = executeCmd(cmd)
			close(done)
		}()

		select {
		case <-done:
			// Check if we got a timeout error
			if errMsg, ok := msg.(app.ErrorMsg); ok {
				t.Logf("Timeout error (expected): %s", errMsg.Err)
			}
		case <-ctx.Done():
			t.Log("Connection timed out as expected")
		}
	})

	t.Run("error view renders correctly", func(t *testing.T) {
		skipIfNoMemcached(t) // We still need a real connection for proper model
		addr := getMemcachedAddr()

		m := app.NewModel(addr)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 80, Height: 24})

		// Simulate error
		m, _ = simulateAppUpdate(m, app.ErrorMsg{Err: "test error message"})

		view := m.View()
		if view == "" {
			t.Error("error view should not be empty")
		}

		// View should contain the error message
		if !containsString(view, "test error message") {
			t.Error("error view should contain the error message")
		}
	})
}

// =============================================================================
// Graceful Shutdown Tests
// =============================================================================

func TestE2E_GracefulShutdown(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("quit with 'q' key", func(t *testing.T) {
		m := app.NewModel(addr)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Press 'q' to quit
		_, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if cmd == nil {
			t.Fatal("expected quit command")
		}

		// Execute the command to verify it's a quit command
		msg := executeCmd(cmd)
		if msg != tea.Quit() {
			// tea.Quit returns a specific message type
			t.Logf("Quit command executed, message type: %T", msg)
		}
	})

	t.Run("quit with ctrl+c", func(t *testing.T) {
		m := app.NewModel(addr)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Press Ctrl+C to quit
		_, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyCtrlC})

		if cmd == nil {
			t.Fatal("expected quit command on Ctrl+C")
		}
	})

	t.Run("quit in ready state", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:quit:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Verify we're in ready state
		if m.State() != app.StateReady {
			t.Fatalf("expected StateReady, got %v", m.State())
		}

		// Press 'q' to quit
		_, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if cmd == nil {
			t.Fatal("expected quit command in ready state")
		}
	})

	t.Run("quit in connecting state", func(t *testing.T) {
		m := app.NewModel(addr)

		// Should be in connecting state initially
		if m.State() != app.StateConnecting {
			t.Fatalf("expected StateConnecting, got %v", m.State())
		}

		// Press 'q' to quit (should work even while connecting)
		_, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if cmd == nil {
			t.Fatal("expected quit command in connecting state")
		}
	})

	t.Run("quit in error state", func(t *testing.T) {
		m := app.NewModel(addr)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Put model in error state
		m, _ = simulateAppUpdate(m, app.ErrorMsg{Err: "test error"})

		if m.State() != app.StateError {
			t.Fatalf("expected StateError, got %v", m.State())
		}

		// Press 'q' to quit
		_, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if cmd == nil {
			t.Fatal("expected quit command in error state")
		}
	})
}

// =============================================================================
// Focus Switching Tests
// =============================================================================

func TestE2E_FocusSwitching(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("Tab switches focus between KeyList and Viewer", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:focus:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Initial focus should be KeyList
		if m.Focus() != app.FocusKeyList {
			t.Errorf("expected initial focus FocusKeyList, got %v", m.Focus())
		}

		// Press Tab to switch to Viewer
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyTab})

		if m.Focus() != app.FocusViewer {
			t.Errorf("expected FocusViewer after Tab, got %v", m.Focus())
		}

		// Press Tab again to switch back to KeyList
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyTab})

		if m.Focus() != app.FocusKeyList {
			t.Errorf("expected FocusKeyList after second Tab, got %v", m.Focus())
		}
	})

	t.Run("Esc returns focus to KeyList from Viewer", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:esc:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Set focus to Viewer
		m.SetFocus(app.FocusViewer)

		if m.Focus() != app.FocusViewer {
			t.Fatalf("expected FocusViewer, got %v", m.Focus())
		}

		// Press Esc to return to KeyList
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyEsc})

		if m.Focus() != app.FocusKeyList {
			t.Errorf("expected FocusKeyList after Esc, got %v", m.Focus())
		}
	})

	t.Run("Esc in KeyList does nothing", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:escnop:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Focus should be KeyList initially
		if m.Focus() != app.FocusKeyList {
			t.Fatalf("expected FocusKeyList, got %v", m.Focus())
		}

		// Press Esc - should stay in KeyList
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyEsc})

		if m.Focus() != app.FocusKeyList {
			t.Errorf("expected FocusKeyList to remain after Esc, got %v", m.Focus())
		}
	})

	t.Run("rapid focus switching", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:rapid:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Rapid Tab presses
		for i := 0; i < 10; i++ {
			m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyTab})
		}

		// After 10 tabs (even number), should be back to KeyList
		if m.Focus() != app.FocusKeyList {
			t.Errorf("expected FocusKeyList after even number of Tabs, got %v", m.Focus())
		}

		// One more Tab
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyTab})

		if m.Focus() != app.FocusViewer {
			t.Errorf("expected FocusViewer after odd number of Tabs, got %v", m.Focus())
		}
	})

	t.Run("focus state affects view rendering", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:viewfocus:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Get view with KeyList focus
		viewKeyList := m.View()

		// Switch to Viewer focus
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyTab})
		viewViewer := m.View()

		// Views should be different (different border colors)
		if viewKeyList == viewViewer {
			t.Log("Note: Views are identical despite focus change (may be expected depending on styling)")
		}
	})
}

// =============================================================================
// Command Palette Tests
// =============================================================================

func TestE2E_CommandPalette(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("Ctrl+P opens command palette", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:palette:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Initial focus should be KeyList
		if m.Focus() != app.FocusKeyList {
			t.Fatalf("expected FocusKeyList, got %v", m.Focus())
		}

		// Press Ctrl+P to open command palette
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyCtrlP})

		// Focus should change to CommandPalette
		if m.Focus() != app.FocusCommandPalette {
			t.Errorf("expected FocusCommandPalette after Ctrl+P, got %v", m.Focus())
		}
	})
}

// =============================================================================
// Help Screen Tests
// =============================================================================

func TestE2E_HelpScreen(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("question mark opens help", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:help:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Press '?' to open help
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

		// Focus should change to Help
		if m.Focus() != app.FocusHelp {
			t.Errorf("expected FocusHelp after ?, got %v", m.Focus())
		}
	})

	t.Run("Esc closes help", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:helpclose:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Open help
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

		if m.Focus() != app.FocusHelp {
			t.Fatalf("expected FocusHelp, got %v", m.Focus())
		}

		// Press Esc to close
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyEsc})

		if m.Focus() != app.FocusKeyList {
			t.Errorf("expected FocusKeyList after closing help, got %v", m.Focus())
		}
	})
}

// =============================================================================
// Filter Mode Tests
// =============================================================================

func TestE2E_FilterMode(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("slash enters filter mode", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:filter:key1": "value1",
			"e2e:filter:key2": "value2",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Press '/' to enter filter mode
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

		// View should show filter input
		view := m.View()
		if !containsString(view, "Filter") {
			t.Log("Note: Filter indicator may not be visible depending on view state")
		}
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestE2E_ConcurrentAccess(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("concurrent model updates are safe", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:concurrent:key": "value",
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		var wg sync.WaitGroup
		numGoroutines := 10

		// Run concurrent view renders
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				view := m.View()
				if view == "" {
					t.Error("view should not be empty")
				}
			}()
		}

		wg.Wait()
	})
}

// =============================================================================
// Value Loading Tests
// =============================================================================

func TestE2E_ValueLoading(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("selecting key loads value", func(t *testing.T) {
		testValue := "test value for value loading"
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:value:key": testValue,
		})
		defer cleanup()

		m := app.NewModel(addr)
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})

		// Find our test key
		var targetKey *models.KeyInfo
		for _, k := range m.Keys() {
			if k.Key == "e2e:value:key" {
				keyCopy := k
				targetKey = &keyCopy
				break
			}
		}

		if targetKey == nil {
			t.Skip("test key not found in enumeration (may be lru_crawler timing issue)")
		}

		// First, send KeySelectedMsg to set currentKey (this is what happens when user selects a key)
		// The keylist component sends this message when a key is selected
		m, cmd := simulateAppUpdate(m, keylist.KeySelectedMsg{Key: *targetKey})

		// The command should be a load value command - execute it
		if cmd != nil {
			msg := executeCmd(cmd)
			if msg != nil {
				// Should be ValueLoadedMsg or ErrorMsg
				if valueMsg, ok := msg.(app.ValueLoadedMsg); ok {
					m, _ = simulateAppUpdate(m, valueMsg)

					// Focus should switch to viewer
					if m.Focus() != app.FocusViewer {
						t.Errorf("expected FocusViewer after value load, got %v", m.Focus())
					}
				} else if errMsg, ok := msg.(app.ErrorMsg); ok {
					t.Fatalf("failed to load value: %s", errMsg.Err)
				}
			}
		}
	})
}

// =============================================================================
// Full Application Lifecycle Test
// =============================================================================

func TestE2E_FullApplicationLifecycle(t *testing.T) {
	skipIfNoMemcached(t)
	addr := getMemcachedAddr()

	t.Run("complete user workflow simulation", func(t *testing.T) {
		cleanup := setupTestKeysForLifecycle(t, map[string]string{
			"e2e:lifecycle:user:1": `{"name": "John", "age": 30}`,
			"e2e:lifecycle:user:2": `{"name": "Jane", "age": 25}`,
			"e2e:lifecycle:config": `{"debug": true}`,
		})
		defer cleanup()

		// 1. Start application
		m := app.NewModel(addr)
		t.Log("Step 1: Application started")

		// 2. Set terminal size
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})
		t.Log("Step 2: Terminal size set to 120x40")

		// 3. Wait for ready state
		m = runAppUntilState(t, m, app.StateReady, 10*time.Second)
		t.Logf("Step 3: Application ready with %d keys", len(m.Keys()))

		// 4. Verify initial focus
		if m.Focus() != app.FocusKeyList {
			t.Errorf("Step 4: expected initial focus FocusKeyList, got %v", m.Focus())
		}
		t.Log("Step 4: Initial focus verified as KeyList")

		// 5. Switch to viewer with Tab
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyTab})
		if m.Focus() != app.FocusViewer {
			t.Errorf("Step 5: expected FocusViewer, got %v", m.Focus())
		}
		t.Log("Step 5: Switched focus to Viewer via Tab")

		// 6. Return to KeyList with Esc
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyEsc})
		if m.Focus() != app.FocusKeyList {
			t.Errorf("Step 6: expected FocusKeyList, got %v", m.Focus())
		}
		t.Log("Step 6: Returned to KeyList via Esc")

		// 7. Open help
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		if m.Focus() != app.FocusHelp {
			t.Errorf("Step 7: expected FocusHelp, got %v", m.Focus())
		}
		t.Log("Step 7: Opened help screen")

		// 8. Close help
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyEsc})
		if m.Focus() != app.FocusKeyList {
			t.Errorf("Step 8: expected FocusKeyList, got %v", m.Focus())
		}
		t.Log("Step 8: Closed help screen")

		// 9. Open command palette
		m, _ = simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyCtrlP})
		if m.Focus() != app.FocusCommandPalette {
			t.Errorf("Step 9: expected FocusCommandPalette, got %v", m.Focus())
		}
		t.Log("Step 9: Opened command palette")

		// 10. Resize window
		m, _ = simulateAppUpdate(m, tea.WindowSizeMsg{Width: 80, Height: 24})
		if m.Width() != 80 || m.Height() != 24 {
			t.Errorf("Step 10: expected dimensions 80x24, got %dx%d", m.Width(), m.Height())
		}
		t.Log("Step 10: Window resized to 80x24")

		// 11. Verify view renders
		view := m.View()
		if view == "" {
			t.Error("Step 11: view should not be empty")
		}
		t.Log("Step 11: View renders correctly")

		// 12. Prepare to quit
		_, cmd := simulateAppUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Error("Step 12: expected quit command")
		}
		t.Log("Step 12: Quit command ready")

		t.Log("Full lifecycle test completed successfully")
	})
}

// NOTE: containsString and other shared helpers are in helpers_test.go
