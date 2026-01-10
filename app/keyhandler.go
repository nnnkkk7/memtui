package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/command"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
	"github.com/nnnkkk7/memtui/ui/components/editor"
)

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle dialog input first
	if m.confirmDialog != nil {
		_, cmd := m.confirmDialog.Update(msg)
		return m, cmd
	}

	if m.inputDialog != nil {
		_, cmd := m.inputDialog.Update(msg)
		return m, cmd
	}

	if m.editor != nil {
		_, cmd := m.editor.Update(msg)
		return m, cmd
	}

	if m.commandPalette.Visible() {
		_, cmd := m.commandPalette.Update(msg)
		return m, cmd
	}

	if m.help.Visible() {
		switch msg.String() {
		case "q", "esc", "?":
			m.help.Hide()
			m.focus = FocusKeyList
		}
		return m, nil
	}

	// Handle filter mode
	if m.filtering {
		return m.handleFilterInput(msg)
	}

	// Global keys
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		// Return to key list from viewer
		if m.focus == FocusViewer {
			m.focus = FocusKeyList
			return m, nil
		}

	case "ctrl+p":
		m.commandPalette.Show()
		m.commandPalette.SetSize(m.width, m.height)
		m.focus = FocusCommandPalette
		return m, nil

	case "?":
		m.help.Show()
		m.help.SetSize(m.width, m.height)
		m.focus = FocusHelp
		return m, nil

	case "r":
		if m.state == StateConnected || m.state == StateReady {
			m.state = StateLoading
			return m, m.loadKeysCmd()
		}

	case "tab":
		if m.focus == FocusKeyList {
			m.focus = FocusViewer
		} else {
			m.focus = FocusKeyList
		}
		return m, nil

	case "/":
		m.filtering = true
		m.filterInput = ""
		return m, nil

	case "d":
		if m.focus == FocusKeyList || m.focus == FocusViewer {
			// Check if we have multi-selected keys
			if m.keyList.HasSelection() {
				selectedKeys := m.keyList.SelectedKeys()
				m.inputDialog = CreateBatchDeleteDialogWithKeys(selectedKeys)
				m.inputDialog.SetSize(m.width, m.height)
				m.focus = FocusDialog
				return m, m.inputDialog.Init()
			}
			// Single key delete
			if m.currentKey != nil {
				m.confirmDialog = CreateDeleteConfirmDialog(m.currentKey.Key)
				m.confirmDialog.SetSize(m.width, m.height)
				m.focus = FocusDialog
				return m, nil
			}
		}

	case "e":
		if m.currentKey != nil && m.currentValue != nil {
			m.editor = editor.New(m.currentKey.Key, m.currentValue)
			// Set CAS for optimistic locking if available
			if m.currentCASItem != nil {
				m.editor.SetCAS(m.currentCASItem.CAS)
			}
			m.editor.SetSize(m.width, m.height)
			m.focus = FocusEditor
			return m, m.editor.Init()
		}

	case "n":
		m.inputDialog = dialog.NewInput("New Key").
			WithPlaceholder("Enter key name...").
			WithValidator(ValidateKeyName)
		m.inputDialog.SetSize(m.width, m.height)
		m.focus = FocusDialog
		return m, m.inputDialog.Init()
	}

	// Focus-specific handling
	switch m.focus {
	case FocusKeyList:
		var cmd tea.Cmd
		m.keyList, cmd = m.keyList.Update(msg)
		return m, cmd

	case FocusViewer:
		var cmd tea.Cmd
		m.viewer, cmd = m.viewer.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filtering = false
		m.filterInput = ""
		m.keyList.SetFilter("")
		return m, nil

	case tea.KeyEnter:
		m.filtering = false
		return m, nil

	case tea.KeyBackspace:
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
			m.keyList.SetFilter(m.filterInput)
		}
		return m, nil

	case tea.KeyRunes:
		m.filterInput += string(msg.Runes)
		m.keyList.SetFilter(m.filterInput)
		return m, nil
	}

	return m, nil
}

func (m *Model) handleCommandExecute(cmd command.Command) (tea.Model, tea.Cmd) {
	m.commandPalette.Hide()
	m.focus = FocusKeyList

	switch cmd.Name {
	case "Refresh keys":
		m.state = StateLoading
		return m, m.loadKeysCmd()

	case "Delete key":
		if m.currentKey != nil {
			m.confirmDialog = CreateDeleteConfirmDialog(m.currentKey.Key)
			m.confirmDialog.SetSize(m.width, m.height)
			m.focus = FocusDialog
		}

	case "New key":
		m.inputDialog = dialog.NewInput("New Key").
			WithPlaceholder("Enter key name...").
			WithValidator(ValidateKeyName)
		m.inputDialog.SetSize(m.width, m.height)
		m.focus = FocusDialog
		return m, m.inputDialog.Init()

	case "Edit value":
		if m.currentKey != nil && m.currentValue != nil {
			m.editor = editor.New(m.currentKey.Key, m.currentValue)
			// Set CAS for optimistic locking if available
			if m.currentCASItem != nil {
				m.editor.SetCAS(m.currentCASItem.CAS)
			}
			m.editor.SetSize(m.width, m.height)
			m.focus = FocusEditor
			return m, m.editor.Init()
		}

	case "Show help":
		m.help.Show()
		m.help.SetSize(m.width, m.height)
		m.focus = FocusHelp

	case "Quit":
		return m, tea.Quit

	case "Filter keys":
		m.filtering = true
		m.filterInput = ""

	case "Show stats":
		// TODO: Implement stats view - for now show message
		m.err = "Stats view not yet implemented"

	case "Toggle theme":
		// TODO: Implement theme toggle - for now show message
		m.err = "Theme toggle not yet implemented"

	case "Copy value":
		if m.currentValue != nil {
			return m, m.copyToClipboardCmd(m.currentValue)
		}
	}

	return m, nil
}
