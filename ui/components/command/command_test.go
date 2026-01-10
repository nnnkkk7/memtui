// Package command provides a command palette component for the memtui application.
package command

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestCommand_Struct verifies the Command struct has all required fields.
func TestCommand_Struct(t *testing.T) {
	cmd := Command{
		Name:        "Test Command",
		Description: "A test command",
		Shortcut:    "t",
		Action: func() tea.Msg {
			return nil
		},
	}

	if cmd.Name != "Test Command" {
		t.Errorf("expected Name to be 'Test Command', got %q", cmd.Name)
	}
	if cmd.Description != "A test command" {
		t.Errorf("expected Description to be 'A test command', got %q", cmd.Description)
	}
	if cmd.Shortcut != "t" {
		t.Errorf("expected Shortcut to be 't', got %q", cmd.Shortcut)
	}
	if cmd.Action == nil {
		t.Error("expected Action to be non-nil")
	}
}

// TestCommandPalette_New verifies that New creates a properly initialized CommandPalette.
func TestCommandPalette_New(t *testing.T) {
	commands := []Command{
		{Name: "Command 1", Shortcut: "1"},
		{Name: "Command 2", Shortcut: "2"},
		{Name: "Command 3", Shortcut: "3"},
	}

	palette := New(commands)

	if palette == nil {
		t.Fatal("expected New to return non-nil palette")
	}

	if len(palette.commands) != 3 {
		t.Errorf("expected 3 commands, got %d", len(palette.commands))
	}

	if len(palette.filtered) != 3 {
		t.Errorf("expected 3 filtered commands initially, got %d", len(palette.filtered))
	}

	if palette.selected != 0 {
		t.Errorf("expected selected to be 0, got %d", palette.selected)
	}

	if palette.visible {
		t.Error("expected palette to be hidden initially (shown with Ctrl+P)")
	}
}

// TestCommandPalette_View verifies that View renders the input and command list.
func TestCommandPalette_View(t *testing.T) {
	commands := []Command{
		{Name: "Refresh keys", Shortcut: "r"},
		{Name: "Delete key", Shortcut: "d"},
	}

	palette := New(commands)
	palette.Show() // Show palette first
	view := palette.View()

	// Should contain the title
	if !strings.Contains(view, "Command Palette") {
		t.Error("expected view to contain 'Command Palette'")
	}

	// Should contain command names
	if !strings.Contains(view, "Refresh keys") {
		t.Error("expected view to contain 'Refresh keys'")
	}
	if !strings.Contains(view, "Delete key") {
		t.Error("expected view to contain 'Delete key'")
	}

	// Should contain shortcuts
	if !strings.Contains(view, "r") {
		t.Error("expected view to contain shortcut 'r'")
	}
	if !strings.Contains(view, "d") {
		t.Error("expected view to contain shortcut 'd'")
	}
}

// TestCommandPalette_Filter verifies fuzzy search filtering of commands.
func TestCommandPalette_Filter(t *testing.T) {
	commands := []Command{
		{Name: "Refresh keys", Shortcut: "r"},
		{Name: "Delete key", Shortcut: "d"},
		{Name: "New key", Shortcut: "n"},
		{Name: "Show stats", Shortcut: "s"},
	}

	palette := New(commands)

	// Initially all commands should be shown
	if len(palette.filtered) != 4 {
		t.Errorf("expected 4 filtered commands initially, got %d", len(palette.filtered))
	}

	// Filter with "key" should match commands containing "key"
	palette.filterCommands("key")
	if len(palette.filtered) != 3 {
		t.Errorf("expected 3 filtered commands for 'key', got %d", len(palette.filtered))
	}

	// Verify the filtered commands contain "key"
	for _, cmd := range palette.filtered {
		if !strings.Contains(strings.ToLower(cmd.Name), "key") {
			t.Errorf("expected filtered command to contain 'key', got %q", cmd.Name)
		}
	}

	// Filter with "stats" should match only "Show stats"
	palette.filterCommands("stats")
	if len(palette.filtered) != 1 {
		t.Errorf("expected 1 filtered command for 'stats', got %d", len(palette.filtered))
	}
	if palette.filtered[0].Name != "Show stats" {
		t.Errorf("expected 'Show stats', got %q", palette.filtered[0].Name)
	}

	// Empty filter should show all commands
	palette.filterCommands("")
	if len(palette.filtered) != 4 {
		t.Errorf("expected 4 filtered commands for empty filter, got %d", len(palette.filtered))
	}
}

// TestCommandPalette_Filter_CaseInsensitive verifies that filtering is case-insensitive.
func TestCommandPalette_Filter_CaseInsensitive(t *testing.T) {
	commands := []Command{
		{Name: "Refresh Keys", Shortcut: "r"},
		{Name: "Delete Key", Shortcut: "d"},
	}

	palette := New(commands)

	// Lowercase query should match
	palette.filterCommands("refresh")
	if len(palette.filtered) != 1 {
		t.Errorf("expected 1 filtered command for 'refresh', got %d", len(palette.filtered))
	}

	// Uppercase query should match
	palette.filterCommands("REFRESH")
	if len(palette.filtered) != 1 {
		t.Errorf("expected 1 filtered command for 'REFRESH', got %d", len(palette.filtered))
	}

	// Mixed case query should match
	palette.filterCommands("ReFrEsH")
	if len(palette.filtered) != 1 {
		t.Errorf("expected 1 filtered command for 'ReFrEsH', got %d", len(palette.filtered))
	}
}

// TestCommandPalette_Navigate verifies up/down selection navigation.
func TestCommandPalette_Navigate(t *testing.T) {
	commands := []Command{
		{Name: "Command 1", Shortcut: "1"},
		{Name: "Command 2", Shortcut: "2"},
		{Name: "Command 3", Shortcut: "3"},
	}

	palette := New(commands)
	palette.Show() // Show palette first

	// Initial selection should be 0
	if palette.selected != 0 {
		t.Errorf("expected initial selection to be 0, got %d", palette.selected)
	}

	// Move down
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	if palette.selected != 1 {
		t.Errorf("expected selection to be 1 after down, got %d", palette.selected)
	}

	// Move down again
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	if palette.selected != 2 {
		t.Errorf("expected selection to be 2 after second down, got %d", palette.selected)
	}

	// Move down at bottom should wrap to top
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	if palette.selected != 0 {
		t.Errorf("expected selection to wrap to 0, got %d", palette.selected)
	}

	// Move up should wrap to bottom
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	if palette.selected != 2 {
		t.Errorf("expected selection to wrap to 2, got %d", palette.selected)
	}

	// Move up
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	if palette.selected != 1 {
		t.Errorf("expected selection to be 1 after up, got %d", palette.selected)
	}
}

// TestCommandPalette_Navigate_Vim verifies vim-style navigation (j/k).
func TestCommandPalette_Navigate_Vim(t *testing.T) {
	commands := []Command{
		{Name: "Command 1", Shortcut: "1"},
		{Name: "Command 2", Shortcut: "2"},
		{Name: "Command 3", Shortcut: "3"},
	}

	palette := New(commands)
	palette.Show() // Show palette first

	// Move down with j
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if palette.selected != 1 {
		t.Errorf("expected selection to be 1 after 'j', got %d", palette.selected)
	}

	// Move up with k
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if palette.selected != 0 {
		t.Errorf("expected selection to be 0 after 'k', got %d", palette.selected)
	}
}

// TestCommandPalette_Execute verifies that Enter returns CommandExecuteMsg.
func TestCommandPalette_Execute(t *testing.T) {
	commands := []Command{
		{
			Name:     "Test Command",
			Shortcut: "t",
			Action: func() tea.Msg {
				return "executed"
			},
		},
	}

	palette := New(commands)
	palette.Show() // Show palette first

	// Press Enter to execute
	_, cmd := palette.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	// Execute the command
	msg := cmd()

	// Check if it's a CommandExecuteMsg
	execMsg, ok := msg.(CommandExecuteMsg)
	if !ok {
		t.Fatalf("expected CommandExecuteMsg, got %T", msg)
	}

	if execMsg.Command.Name != "Test Command" {
		t.Errorf("expected command name 'Test Command', got %q", execMsg.Command.Name)
	}
}

// TestCommandPalette_Cancel verifies that Escape returns CommandCancelMsg.
func TestCommandPalette_Cancel(t *testing.T) {
	commands := []Command{
		{Name: "Test Command", Shortcut: "t"},
	}

	palette := New(commands)
	palette.Show() // Show palette first

	// Press Escape to cancel
	_, cmd := palette.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	// Execute the command
	msg := cmd()

	// Check if it's a CommandCancelMsg
	_, ok := msg.(CommandCancelMsg)
	if !ok {
		t.Fatalf("expected CommandCancelMsg, got %T", msg)
	}

	// Palette should be hidden after cancel
	if palette.visible {
		t.Error("expected palette to be hidden after cancel")
	}
}

// TestCommandPalette_Visible verifies visibility control.
func TestCommandPalette_Visible(t *testing.T) {
	palette := New(nil)

	// Initially hidden (must call Show() or use Ctrl+P to display)
	if palette.Visible() {
		t.Error("expected palette to be hidden initially")
	}

	// Show
	palette.Show()
	if !palette.Visible() {
		t.Error("expected palette to be visible after Show()")
	}

	// Hide
	palette.Hide()
	if palette.Visible() {
		t.Error("expected palette to be hidden after Hide()")
	}

	// Show again
	palette.Show()
	if !palette.Visible() {
		t.Error("expected palette to be visible after Show()")
	}

	// Toggle
	palette.Toggle()
	if palette.Visible() {
		t.Error("expected palette to be hidden after Toggle()")
	}

	palette.Toggle()
	if !palette.Visible() {
		t.Error("expected palette to be visible after second Toggle()")
	}
}

// TestCommandPalette_SetSize verifies size setting.
func TestCommandPalette_SetSize(t *testing.T) {
	palette := New(nil)

	palette.SetSize(100, 50)

	if palette.width != 100 {
		t.Errorf("expected width to be 100, got %d", palette.width)
	}
	if palette.height != 50 {
		t.Errorf("expected height to be 50, got %d", palette.height)
	}
}

// TestDefaultCommands verifies that DefaultCommands returns the expected commands.
func TestDefaultCommands(t *testing.T) {
	commands := DefaultCommands()

	// Should have at least 10 commands as specified
	if len(commands) < 10 {
		t.Errorf("expected at least 10 default commands, got %d", len(commands))
	}

	// Create a map for easy lookup
	cmdMap := make(map[string]Command)
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	// Check for required commands
	expectedCommands := []struct {
		name     string
		shortcut string
	}{
		{"Refresh keys", "r"},
		{"Delete key", "d"},
		{"New key", "n"},
		{"Edit value", "e"},
		{"Show stats", "s"},
		{"Toggle theme", ""},
		{"Show help", "?"},
		{"Quit", "q"},
		{"Filter keys", "/"},
		{"Copy value", "c"},
	}

	for _, expected := range expectedCommands {
		cmd, found := cmdMap[expected.name]
		if !found {
			t.Errorf("expected to find command %q", expected.name)
			continue
		}
		if expected.shortcut != "" && cmd.Shortcut != expected.shortcut {
			t.Errorf("expected command %q to have shortcut %q, got %q",
				expected.name, expected.shortcut, cmd.Shortcut)
		}
	}
}

// TestCommandPalette_SelectionReset verifies that selection resets when filtering.
func TestCommandPalette_SelectionReset(t *testing.T) {
	commands := []Command{
		{Name: "Alpha", Shortcut: "a"},
		{Name: "Beta", Shortcut: "b"},
		{Name: "Gamma", Shortcut: "g"},
	}

	palette := New(commands)

	// Move selection down
	palette.selected = 2

	// Filter commands
	palette.filterCommands("Alpha")

	// Selection should reset to 0
	if palette.selected != 0 {
		t.Errorf("expected selection to reset to 0 after filter, got %d", palette.selected)
	}
}

// TestCommandPalette_NoCommands verifies behavior with no matching commands.
func TestCommandPalette_NoCommands(t *testing.T) {
	commands := []Command{
		{Name: "Refresh keys", Shortcut: "r"},
	}

	palette := New(commands)
	palette.Show() // Show palette first

	// Filter with non-matching query
	palette.filterCommands("xyz")

	if len(palette.filtered) != 0 {
		t.Errorf("expected 0 filtered commands, got %d", len(palette.filtered))
	}

	// View should still render without panic
	view := palette.View()
	if view == "" {
		t.Error("expected non-empty view even with no matching commands")
	}
}

// TestCommandPalette_InputUpdate verifies that typing updates the filter.
func TestCommandPalette_InputUpdate(t *testing.T) {
	commands := []Command{
		{Name: "Refresh keys", Shortcut: "r"},
		{Name: "Delete key", Shortcut: "d"},
		{Name: "Show stats", Shortcut: "s"},
	}

	palette := New(commands)
	palette.Show() // Show palette first to enable input

	// Type "stat" to filter
	for _, r := range "stat" {
		palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Should filter to only "Show stats"
	if len(palette.filtered) != 1 {
		t.Errorf("expected 1 filtered command after typing 'stat', got %d", len(palette.filtered))
	}
}

// TestCommandPalette_SelectedCommand verifies SelectedCommand returns correct command.
func TestCommandPalette_SelectedCommand(t *testing.T) {
	commands := []Command{
		{Name: "Command 1", Shortcut: "1"},
		{Name: "Command 2", Shortcut: "2"},
		{Name: "Command 3", Shortcut: "3"},
	}

	palette := New(commands)

	// Get first command
	cmd := palette.SelectedCommand()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	if cmd.Name != "Command 1" {
		t.Errorf("expected 'Command 1', got %q", cmd.Name)
	}

	// Move selection and check again
	palette.selected = 2
	cmd = palette.SelectedCommand()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	if cmd.Name != "Command 3" {
		t.Errorf("expected 'Command 3', got %q", cmd.Name)
	}

	// With empty filtered list
	palette.filtered = []Command{}
	cmd = palette.SelectedCommand()
	if cmd != nil {
		t.Error("expected nil command when no filtered commands")
	}
}
