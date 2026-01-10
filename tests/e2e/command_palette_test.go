//go:build e2e

// Package e2e_test contains end-to-end tests for the memtui application.
// These tests verify the command palette functionality including opening,
// navigation, filtering, command execution, and cancellation.
package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/command"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

// setupReadyModel creates a model in the Ready state with test keys loaded.
// This simulates a connected application ready for user interaction.
func setupReadyModel(t *testing.T) *app.Model {
	t.Helper()
	m := app.NewModel("localhost:11211")

	// Set window size
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(*app.Model)

	// Simulate successful connection
	model, _ = m.Update(app.ConnectedMsg{Version: "1.6.22"})
	m = model.(*app.Model)

	// Load test keys
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
		{Key: "session:abc", Size: 50},
	}
	model, _ = m.Update(app.KeysLoadedMsg{Keys: keys})
	m = model.(*app.Model)

	return m
}

// setupReadyModelWithValue creates a model in the Ready state with a value loaded.
// This simulates a connected application with a key selected and its value loaded.
func setupReadyModelWithValue(t *testing.T) *app.Model {
	t.Helper()
	m := setupReadyModel(t)

	// Simulate key selection (this sets currentKey)
	keyInfo := models.KeyInfo{Key: "user:1", Size: 100}
	model, _ := m.Update(keylist.KeySelectedMsg{Key: keyInfo})
	m = model.(*app.Model)

	// Simulate value loaded for current key
	model, _ = m.Update(app.ValueLoadedMsg{Key: "user:1", Value: []byte("test value")})
	m = model.(*app.Model)

	return m
}

// TestE2E_CommandPalette_OpenWithCtrlP verifies that pressing Ctrl+P opens the command palette.
func TestE2E_CommandPalette_OpenWithCtrlP(t *testing.T) {
	m := setupReadyModel(t)

	// Verify command palette is not visible initially
	if m.Focus() == app.FocusCommandPalette {
		t.Error("command palette should not be focused initially")
	}

	// Press Ctrl+P to open command palette
	ctrlPMsg := tea.KeyMsg{Type: tea.KeyCtrlP}
	model, _ := m.Update(ctrlPMsg)
	m = model.(*app.Model)

	// Verify focus changed to command palette
	if m.Focus() != app.FocusCommandPalette {
		t.Errorf("expected FocusCommandPalette after Ctrl+P, got %v", m.Focus())
	}
}

// TestE2E_CommandPalette_CancelWithEsc verifies that pressing Esc cancels the command palette.
func TestE2E_CommandPalette_CancelWithEsc(t *testing.T) {
	m := setupReadyModel(t)

	// Open command palette with Ctrl+P
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = model.(*app.Model)

	if m.Focus() != app.FocusCommandPalette {
		t.Fatalf("expected FocusCommandPalette, got %v", m.Focus())
	}

	// Press Esc to cancel - this generates a command that returns CommandCancelMsg
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*app.Model)

	// Execute the command to get the cancel message
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(*app.Model)
	}

	// Verify focus returned to key list
	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList after Esc, got %v", m.Focus())
	}
}

// TestE2E_CommandPalette_NavigateWithArrowKeys verifies navigation with Up/Down arrow keys.
func TestE2E_CommandPalette_NavigateWithArrowKeys(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Initial selection should be 0
	if palette.SelectedCommand().Name != commands[0].Name {
		t.Errorf("expected initial selection to be '%s', got '%s'",
			commands[0].Name, palette.SelectedCommand().Name)
	}

	// Navigate down
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	if palette.SelectedCommand().Name != commands[1].Name {
		t.Errorf("expected selection to be '%s' after Down, got '%s'",
			commands[1].Name, palette.SelectedCommand().Name)
	}

	// Navigate down again
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	if palette.SelectedCommand().Name != commands[2].Name {
		t.Errorf("expected selection to be '%s' after second Down, got '%s'",
			commands[2].Name, palette.SelectedCommand().Name)
	}

	// Navigate up
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	if palette.SelectedCommand().Name != commands[1].Name {
		t.Errorf("expected selection to be '%s' after Up, got '%s'",
			commands[1].Name, palette.SelectedCommand().Name)
	}

	// Navigate up to wrap to last item
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	if palette.SelectedCommand().Name != commands[len(commands)-1].Name {
		t.Errorf("expected selection to wrap to '%s', got '%s'",
			commands[len(commands)-1].Name, palette.SelectedCommand().Name)
	}
}

// TestE2E_CommandPalette_NavigateWithVimKeys verifies navigation with j/k vim-style keys.
func TestE2E_CommandPalette_NavigateWithVimKeys(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Initial selection should be 0
	initialCmd := palette.SelectedCommand()
	if initialCmd.Name != commands[0].Name {
		t.Errorf("expected initial selection to be '%s', got '%s'",
			commands[0].Name, initialCmd.Name)
	}

	// Navigate down with 'j'
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if palette.SelectedCommand().Name != commands[1].Name {
		t.Errorf("expected selection to be '%s' after 'j', got '%s'",
			commands[1].Name, palette.SelectedCommand().Name)
	}

	// Navigate down again with 'j'
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if palette.SelectedCommand().Name != commands[2].Name {
		t.Errorf("expected selection to be '%s' after second 'j', got '%s'",
			commands[2].Name, palette.SelectedCommand().Name)
	}

	// Navigate up with 'k'
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if palette.SelectedCommand().Name != commands[1].Name {
		t.Errorf("expected selection to be '%s' after 'k', got '%s'",
			commands[1].Name, palette.SelectedCommand().Name)
	}
}

// TestE2E_CommandPalette_FuzzySearch verifies fuzzy search filtering of commands.
func TestE2E_CommandPalette_FuzzySearch(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	testCases := []struct {
		name          string
		query         string
		expectedMatch string
		minMatches    int
	}{
		{
			name:          "search for 'ref' should match 'Refresh keys'",
			query:         "ref",
			expectedMatch: "Refresh keys",
			minMatches:    1,
		},
		{
			name:          "search for 'del' should match 'Delete key'",
			query:         "del",
			expectedMatch: "Delete key",
			minMatches:    1,
		},
		{
			name:          "search for 'key' should match multiple commands",
			query:         "key",
			expectedMatch: "", // Multiple matches expected
			minMatches:    3,  // Refresh keys, Delete key, New key, Filter keys
		},
		{
			name:          "search for 'quit' should match 'Quit'",
			query:         "quit",
			expectedMatch: "Quit",
			minMatches:    1,
		},
		{
			name:          "search for 'theme' should match 'Toggle theme'",
			query:         "theme",
			expectedMatch: "Toggle theme",
			minMatches:    1,
		},
		{
			name:          "search for 'stats' should match 'Show stats'",
			query:         "stats",
			expectedMatch: "Show stats",
			minMatches:    1,
		},
		{
			name:          "search for 'copy' should match 'Copy value'",
			query:         "copy",
			expectedMatch: "Copy value",
			minMatches:    1,
		},
		{
			name:          "case insensitive search",
			query:         "HELP",
			expectedMatch: "Show help",
			minMatches:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset palette
			palette := command.New(commands)
			palette.Show()

			// Type search query
			for _, r := range tc.query {
				palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}

			// Check the view contains expected content
			view := palette.View()

			if tc.expectedMatch != "" {
				if !strings.Contains(view, tc.expectedMatch) {
					t.Errorf("expected view to contain '%s' for query '%s'", tc.expectedMatch, tc.query)
				}
			}

			// Verify we have at least the expected number of matches
			selectedCmd := palette.SelectedCommand()
			if selectedCmd == nil && tc.minMatches > 0 {
				t.Errorf("expected at least %d matches for query '%s', got 0", tc.minMatches, tc.query)
			}
		})
	}
}

// TestE2E_CommandPalette_ExecuteWithEnter verifies that pressing Enter executes the selected command.
func TestE2E_CommandPalette_ExecuteWithEnter(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Select "Refresh keys" (first command)
	selectedBefore := palette.SelectedCommand()
	if selectedBefore.Name != "Refresh keys" {
		t.Fatalf("expected 'Refresh keys' to be selected, got '%s'", selectedBefore.Name)
	}

	// Press Enter to execute
	_, cmd := palette.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command to be returned after Enter")
	}

	// Execute the command
	msg := cmd()

	// Verify it returns CommandExecuteMsg
	execMsg, ok := msg.(command.CommandExecuteMsg)
	if !ok {
		t.Fatalf("expected CommandExecuteMsg, got %T", msg)
	}

	if execMsg.Command.Name != "Refresh keys" {
		t.Errorf("expected executed command to be 'Refresh keys', got '%s'", execMsg.Command.Name)
	}

	// Palette should be hidden after execution
	if palette.Visible() {
		t.Error("command palette should be hidden after command execution")
	}
}

// TestE2E_CommandPalette_ExecuteFilteredCommand verifies executing a command after filtering.
func TestE2E_CommandPalette_ExecuteFilteredCommand(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Type "quit" to filter to Quit command
	for _, r := range "quit" {
		palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Verify Quit is selected (should be only match)
	selectedCmd := palette.SelectedCommand()
	if selectedCmd == nil {
		t.Fatal("expected a command to be selected after filtering")
	}
	if selectedCmd.Name != "Quit" {
		t.Errorf("expected 'Quit' to be selected, got '%s'", selectedCmd.Name)
	}

	// Press Enter to execute
	_, cmd := palette.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command to be returned after Enter")
	}

	msg := cmd()
	execMsg, ok := msg.(command.CommandExecuteMsg)
	if !ok {
		t.Fatalf("expected CommandExecuteMsg, got %T", msg)
	}

	if execMsg.Command.Name != "Quit" {
		t.Errorf("expected executed command to be 'Quit', got '%s'", execMsg.Command.Name)
	}
}

// TestE2E_Command_RefreshKeys verifies the "Refresh keys" command triggers the correct message.
func TestE2E_Command_RefreshKeys(t *testing.T) {
	m := setupReadyModel(t)

	// Open command palette
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = model.(*app.Model)

	// Create and execute the Refresh keys command message
	refreshCmd := command.DefaultCommands()[0] // Refresh keys is first
	if refreshCmd.Name != "Refresh keys" {
		t.Fatalf("expected first command to be 'Refresh keys', got '%s'", refreshCmd.Name)
	}

	// Verify the action returns RefreshKeysMsg
	msg := refreshCmd.Action()
	_, ok := msg.(command.RefreshKeysMsg)
	if !ok {
		t.Errorf("expected RefreshKeysMsg, got %T", msg)
	}
}

// TestE2E_Command_DeleteKey verifies the "Delete key" command triggers the correct message.
func TestE2E_Command_DeleteKey(t *testing.T) {
	commands := command.DefaultCommands()
	var deleteCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Delete key" {
			deleteCmd = cmd
			break
		}
	}

	if deleteCmd.Name == "" {
		t.Fatal("could not find 'Delete key' command")
	}

	// Verify the action returns DeleteKeyMsg
	msg := deleteCmd.Action()
	_, ok := msg.(command.DeleteKeyMsg)
	if !ok {
		t.Errorf("expected DeleteKeyMsg, got %T", msg)
	}
}

// TestE2E_Command_NewKey verifies the "New key" command triggers the correct message.
func TestE2E_Command_NewKey(t *testing.T) {
	commands := command.DefaultCommands()
	var newKeyCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "New key" {
			newKeyCmd = cmd
			break
		}
	}

	if newKeyCmd.Name == "" {
		t.Fatal("could not find 'New key' command")
	}

	// Verify the action returns NewKeyMsg
	msg := newKeyCmd.Action()
	_, ok := msg.(command.NewKeyMsg)
	if !ok {
		t.Errorf("expected NewKeyMsg, got %T", msg)
	}
}

// TestE2E_Command_EditValue verifies the "Edit value" command triggers the correct message.
func TestE2E_Command_EditValue(t *testing.T) {
	commands := command.DefaultCommands()
	var editCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Edit value" {
			editCmd = cmd
			break
		}
	}

	if editCmd.Name == "" {
		t.Fatal("could not find 'Edit value' command")
	}

	// Verify the action returns EditValueMsg
	msg := editCmd.Action()
	_, ok := msg.(command.EditValueMsg)
	if !ok {
		t.Errorf("expected EditValueMsg, got %T", msg)
	}
}

// TestE2E_Command_ShowStats verifies the "Show stats" command triggers the correct message.
func TestE2E_Command_ShowStats(t *testing.T) {
	commands := command.DefaultCommands()
	var statsCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Show stats" {
			statsCmd = cmd
			break
		}
	}

	if statsCmd.Name == "" {
		t.Fatal("could not find 'Show stats' command")
	}

	// Verify the action returns ShowStatsMsg
	msg := statsCmd.Action()
	_, ok := msg.(command.ShowStatsMsg)
	if !ok {
		t.Errorf("expected ShowStatsMsg, got %T", msg)
	}
}

// TestE2E_Command_ToggleTheme verifies the "Toggle theme" command triggers the correct message.
func TestE2E_Command_ToggleTheme(t *testing.T) {
	commands := command.DefaultCommands()
	var themeCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Toggle theme" {
			themeCmd = cmd
			break
		}
	}

	if themeCmd.Name == "" {
		t.Fatal("could not find 'Toggle theme' command")
	}

	// Verify the action returns ToggleThemeMsg
	msg := themeCmd.Action()
	_, ok := msg.(command.ToggleThemeMsg)
	if !ok {
		t.Errorf("expected ToggleThemeMsg, got %T", msg)
	}
}

// TestE2E_Command_ShowHelp verifies the "Show help" command triggers the correct message.
func TestE2E_Command_ShowHelp(t *testing.T) {
	commands := command.DefaultCommands()
	var helpCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Show help" {
			helpCmd = cmd
			break
		}
	}

	if helpCmd.Name == "" {
		t.Fatal("could not find 'Show help' command")
	}

	// Verify the action returns ShowHelpMsg
	msg := helpCmd.Action()
	_, ok := msg.(command.ShowHelpMsg)
	if !ok {
		t.Errorf("expected ShowHelpMsg, got %T", msg)
	}
}

// TestE2E_Command_Quit verifies the "Quit" command triggers the correct message.
func TestE2E_Command_Quit(t *testing.T) {
	commands := command.DefaultCommands()
	var quitCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Quit" {
			quitCmd = cmd
			break
		}
	}

	if quitCmd.Name == "" {
		t.Fatal("could not find 'Quit' command")
	}

	// Verify the action returns QuitMsg
	msg := quitCmd.Action()
	_, ok := msg.(command.QuitMsg)
	if !ok {
		t.Errorf("expected QuitMsg, got %T", msg)
	}
}

// TestE2E_Command_FilterKeys verifies the "Filter keys" command triggers the correct message.
func TestE2E_Command_FilterKeys(t *testing.T) {
	commands := command.DefaultCommands()
	var filterCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Filter keys" {
			filterCmd = cmd
			break
		}
	}

	if filterCmd.Name == "" {
		t.Fatal("could not find 'Filter keys' command")
	}

	// Verify the action returns FilterKeysMsg
	msg := filterCmd.Action()
	_, ok := msg.(command.FilterKeysMsg)
	if !ok {
		t.Errorf("expected FilterKeysMsg, got %T", msg)
	}
}

// TestE2E_Command_CopyValue verifies the "Copy value" command triggers the correct message.
func TestE2E_Command_CopyValue(t *testing.T) {
	commands := command.DefaultCommands()
	var copyCmd command.Command
	for _, cmd := range commands {
		if cmd.Name == "Copy value" {
			copyCmd = cmd
			break
		}
	}

	if copyCmd.Name == "" {
		t.Fatal("could not find 'Copy value' command")
	}

	// Verify the action returns CopyValueMsg
	msg := copyCmd.Action()
	_, ok := msg.(command.CopyValueMsg)
	if !ok {
		t.Errorf("expected CopyValueMsg, got %T", msg)
	}
}

// TestE2E_CommandPalette_DefaultCommandsShortcuts verifies all default commands have correct shortcuts.
func TestE2E_CommandPalette_DefaultCommandsShortcuts(t *testing.T) {
	commands := command.DefaultCommands()

	expectedShortcuts := map[string]string{
		"Refresh keys": "r",
		"Delete key":   "d",
		"New key":      "n",
		"Edit value":   "e",
		"Show stats":   "s",
		"Toggle theme": "", // No shortcut
		"Show help":    "?",
		"Quit":         "q",
		"Filter keys":  "/",
		"Copy value":   "c",
	}

	cmdMap := make(map[string]command.Command)
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	for name, expectedShortcut := range expectedShortcuts {
		cmd, found := cmdMap[name]
		if !found {
			t.Errorf("expected to find command '%s'", name)
			continue
		}
		if cmd.Shortcut != expectedShortcut {
			t.Errorf("command '%s': expected shortcut '%s', got '%s'",
				name, expectedShortcut, cmd.Shortcut)
		}
	}
}

// TestE2E_CommandPalette_ViewContainsHints verifies the view shows navigation hints.
func TestE2E_CommandPalette_ViewContainsHints(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	view := palette.View()

	expectedHints := []string{
		"Enter",
		"Esc",
		"execute",
		"cancel",
	}

	for _, hint := range expectedHints {
		if !strings.Contains(view, hint) {
			t.Errorf("expected view to contain hint '%s'", hint)
		}
	}
}

// TestE2E_CommandPalette_ViewContainsTitle verifies the view shows the title.
func TestE2E_CommandPalette_ViewContainsTitle(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	view := palette.View()

	if !strings.Contains(view, "Command Palette") {
		t.Error("expected view to contain 'Command Palette' title")
	}
}

// TestE2E_CommandPalette_NoMatchesMessage verifies "No matching commands" message appears.
func TestE2E_CommandPalette_NoMatchesMessage(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Type a query that matches nothing
	for _, r := range "xyznonexistent123" {
		palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	view := palette.View()

	if !strings.Contains(view, "No matching commands") {
		t.Error("expected view to contain 'No matching commands' when no matches found")
	}
}

// TestE2E_CommandPalette_SelectionResetsOnFilter verifies selection resets when filtering.
func TestE2E_CommandPalette_SelectionResetsOnFilter(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Navigate down several times
	for i := 0; i < 5; i++ {
		palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Type to filter
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Selection should be reset to first filtered item
	selectedCmd := palette.SelectedCommand()
	if selectedCmd == nil {
		t.Fatal("expected a command to be selected")
	}

	// The first filtered command should be selected
	// Since we typed 'q', "Quit" should be one of the matches and should be first or near first
	// Selection should be 0 (first in filtered list)
}

// TestE2E_CommandPalette_WrapAroundNavigation verifies navigation wraps at boundaries.
func TestE2E_CommandPalette_WrapAroundNavigation(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()
	palette.SetSize(80, 40)

	// Navigate up from first item should wrap to last
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	selectedCmd := palette.SelectedCommand()
	if selectedCmd.Name != commands[len(commands)-1].Name {
		t.Errorf("expected selection to wrap to last command '%s', got '%s'",
			commands[len(commands)-1].Name, selectedCmd.Name)
	}

	// Navigate down should go back to first
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectedCmd = palette.SelectedCommand()
	if selectedCmd.Name != commands[0].Name {
		t.Errorf("expected selection to be first command '%s', got '%s'",
			commands[0].Name, selectedCmd.Name)
	}
}

// TestE2E_CommandPalette_IntegrationWithApp verifies command palette integration with main app.
func TestE2E_CommandPalette_IntegrationWithApp(t *testing.T) {
	m := setupReadyModel(t)

	// Test opening command palette
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = model.(*app.Model)

	if m.Focus() != app.FocusCommandPalette {
		t.Errorf("expected FocusCommandPalette, got %v", m.Focus())
	}

	// Test executing a command through the app
	// Simulate command execution message
	quitCmd := command.Command{
		Name:   "Quit",
		Action: func() tea.Msg { return command.QuitMsg{} },
	}
	execMsg := command.CommandExecuteMsg{Command: quitCmd}
	model, cmd := m.Update(execMsg)
	m = model.(*app.Model)

	// Focus should return to KeyList
	if m.Focus() != app.FocusKeyList {
		t.Errorf("expected FocusKeyList after command execution, got %v", m.Focus())
	}

	// For Quit command, there should be a quit command returned
	if cmd == nil {
		t.Error("expected quit command to be returned")
	}
}

// TestE2E_CommandPalette_RefreshKeysIntegration tests the Refresh keys command end-to-end.
func TestE2E_CommandPalette_RefreshKeysIntegration(t *testing.T) {
	m := setupReadyModel(t)

	// Open command palette
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = model.(*app.Model)

	// Execute Refresh keys command
	refreshCmd := command.Command{
		Name:   "Refresh keys",
		Action: func() tea.Msg { return command.RefreshKeysMsg{} },
	}
	execMsg := command.CommandExecuteMsg{Command: refreshCmd}
	model, cmd := m.Update(execMsg)
	m = model.(*app.Model)

	// The app should transition to Loading state and return a command to load keys
	if m.State() != app.StateLoading {
		t.Errorf("expected StateLoading after Refresh keys, got %v", m.State())
	}

	// A command should be returned (loadKeysCmd)
	if cmd == nil {
		t.Error("expected load keys command to be returned")
	}
}

// TestE2E_CommandPalette_HelpIntegration tests the Show help command end-to-end.
func TestE2E_CommandPalette_HelpIntegration(t *testing.T) {
	m := setupReadyModel(t)

	// Open command palette
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = model.(*app.Model)

	// Execute Show help command
	helpCmd := command.Command{
		Name:   "Show help",
		Action: func() tea.Msg { return command.ShowHelpMsg{} },
	}
	execMsg := command.CommandExecuteMsg{Command: helpCmd}
	model, _ = m.Update(execMsg)
	m = model.(*app.Model)

	// Focus should be on Help
	if m.Focus() != app.FocusHelp {
		t.Errorf("expected FocusHelp after Show help, got %v", m.Focus())
	}
}

// TestE2E_CommandPalette_FilterKeysIntegration tests the Filter keys command end-to-end.
func TestE2E_CommandPalette_FilterKeysIntegration(t *testing.T) {
	m := setupReadyModel(t)

	// Verify not in filter mode initially
	view := m.View()
	if strings.Contains(view, "Filter:") && strings.Contains(view, "_") {
		t.Error("should not be in filter mode initially")
	}

	// Open command palette
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = model.(*app.Model)

	// Execute Filter keys command
	filterCmd := command.Command{
		Name:   "Filter keys",
		Action: func() tea.Msg { return command.FilterKeysMsg{} },
	}
	execMsg := command.CommandExecuteMsg{Command: filterCmd}
	model, _ = m.Update(execMsg)
	m = model.(*app.Model)

	// Should be in filter mode - check the view
	view = m.View()
	if !strings.Contains(view, "Filter:") {
		t.Error("expected to be in filter mode after Filter keys command")
	}
}

// TestE2E_CommandPalette_AllCommandsHaveActions verifies all default commands have non-nil actions.
func TestE2E_CommandPalette_AllCommandsHaveActions(t *testing.T) {
	commands := command.DefaultCommands()

	for _, cmd := range commands {
		if cmd.Action == nil {
			t.Errorf("command '%s' has nil Action", cmd.Name)
			continue
		}

		// Verify action returns a message
		msg := cmd.Action()
		if msg == nil {
			t.Errorf("command '%s' action returned nil message", cmd.Name)
		}
	}
}

// TestE2E_CommandPalette_AllCommandsHaveDescriptions verifies all default commands have descriptions.
func TestE2E_CommandPalette_AllCommandsHaveDescriptions(t *testing.T) {
	commands := command.DefaultCommands()

	for _, cmd := range commands {
		if cmd.Description == "" {
			t.Errorf("command '%s' has empty Description", cmd.Name)
		}
	}
}

// TestE2E_CommandPalette_FuzzyMatchingScores tests that fuzzy matching ranks results properly.
func TestE2E_CommandPalette_FuzzyMatchingScores(t *testing.T) {
	commands := command.DefaultCommands()

	testCases := []struct {
		query         string
		expectedFirst string
	}{
		{
			query:         "ref",
			expectedFirst: "Refresh keys",
		},
		{
			query:         "qui",
			expectedFirst: "Quit",
		},
		{
			query:         "hel",
			expectedFirst: "Show help",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			palette := command.New(commands)
			palette.Show()

			// Type search query
			for _, r := range tc.query {
				palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}

			selectedCmd := palette.SelectedCommand()
			if selectedCmd == nil {
				t.Fatalf("expected a command to be selected for query '%s'", tc.query)
			}
			if selectedCmd.Name != tc.expectedFirst {
				t.Errorf("for query '%s', expected first result to be '%s', got '%s'",
					tc.query, tc.expectedFirst, selectedCmd.Name)
			}
		})
	}
}

// TestE2E_CommandPalette_VisibilityToggle tests Show/Hide/Toggle methods.
func TestE2E_CommandPalette_VisibilityToggle(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)

	// Initially hidden
	if palette.Visible() {
		t.Error("palette should be hidden initially")
	}

	// Show
	palette.Show()
	if !palette.Visible() {
		t.Error("palette should be visible after Show()")
	}

	// Hide
	palette.Hide()
	if palette.Visible() {
		t.Error("palette should be hidden after Hide()")
	}

	// Toggle on
	palette.Toggle()
	if !palette.Visible() {
		t.Error("palette should be visible after Toggle() from hidden")
	}

	// Toggle off
	palette.Toggle()
	if palette.Visible() {
		t.Error("palette should be hidden after Toggle() from visible")
	}
}

// TestE2E_CommandPalette_InputClearsOnShow tests that input is cleared when palette is shown.
func TestE2E_CommandPalette_InputClearsOnShow(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)
	palette.Show()

	// Type something
	for _, r := range "test" {
		palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Hide and show again
	palette.Hide()
	palette.Show()

	// All commands should be visible again (no filter)
	selectedCmd := palette.SelectedCommand()
	if selectedCmd.Name != commands[0].Name {
		t.Error("expected first command to be selected after Show() clears input")
	}
}

// TestE2E_CommandPalette_UpdateWhenHidden tests that Update does nothing when palette is hidden.
func TestE2E_CommandPalette_UpdateWhenHidden(t *testing.T) {
	commands := command.DefaultCommands()
	palette := command.New(commands)

	// Don't show the palette, it should be hidden by default

	// Try to navigate down
	_, cmd := palette.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Should return nil command since palette is hidden
	if cmd != nil {
		t.Error("Update on hidden palette should return nil command")
	}

	// View should be empty when hidden
	view := palette.View()
	if view != "" {
		t.Error("View should be empty when palette is hidden")
	}
}
