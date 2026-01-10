package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/ui/components/command"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
	"github.com/nnnkkk7/memtui/ui/components/editor"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateComponentSizes()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case ConnectedMsg:
		m.state = StateConnected
		m.version = msg.Version
		m.supportsMetadump = msg.SupportsMetadump
		// Initialize unified client (supports both basic and CAS operations)
		mcClient, err := client.New(m.addr)
		if err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Sprintf("failed to create client: %v", err)}
			}
		}
		m.mcClient = mcClient
		m.state = StateLoading
		// Show warning if metadump not supported
		if !m.supportsMetadump {
			m.err = "Warning: Server does not support key enumeration (requires memcached >= 1.4.31). You can only work with known keys."
		}
		return m, m.loadKeysCmd()

	case KeysLoadedMsg:
		m.keys = msg.Keys
		m.keyList.SetKeys(m.keys)
		m.state = StateReady
		return m, nil

	case keylist.KeySelectedMsg:
		m.currentKey = &msg.Key
		return m, m.loadValueCmd(msg.Key.Key)

	case ValueLoadedMsg:
		m.currentValue = msg.Value
		// Store CAS item for later use in editor save
		if msg.CAS > 0 {
			m.currentCASItem = &client.CASItem{
				Key:        msg.Key,
				Value:      msg.Value,
				Flags:      msg.Flags,
				Expiration: msg.Expiration,
				CAS:        msg.CAS,
			}
		} else {
			m.currentCASItem = nil
		}
		m.viewer.SetKeyInfo(*m.currentKey)
		m.viewer.SetValue(msg.Value)
		m.focus = FocusViewer
		return m, nil

	case command.CommandExecuteMsg:
		return m.handleCommandExecute(msg.Command)

	case command.CommandCancelMsg:
		m.commandPalette.Hide()
		m.focus = FocusKeyList
		return m, nil

	case dialog.ConfirmResultMsg:
		return m.handleConfirmResult(msg)

	case dialog.InputResultMsg:
		return m.handleInputResult(msg)

	case editor.EditorSaveMsg:
		return m.handleEditorSave(msg)

	case editor.EditorCancelMsg:
		m.editor = nil
		m.focus = FocusKeyList
		return m, nil

	case KeyDeletedMsg:
		m.currentKey = nil
		m.currentValue = nil
		m.confirmDialog = nil
		m.focus = FocusKeyList
		return m, m.loadKeysCmd()

	case BatchDeleteResultMsg:
		summary := HandleBatchDeleteResult(msg)
		m.currentKey = nil
		m.currentValue = nil
		m.focus = FocusKeyList
		if summary.HasErrors {
			m.err = summary.String()
		}
		if summary.ShouldRefresh {
			return m, m.loadKeysCmd()
		}
		return m, nil

	case DeleteErrorMsg:
		m.err = msg.Err.Error()
		m.confirmDialog = nil
		m.focus = FocusKeyList
		return m, nil

	case KeyCreatedMsg:
		m.inputDialog = nil
		m.focus = FocusKeyList
		return m, m.loadKeysCmd()

	case NewKeyErrorMsg:
		m.err = msg.Err.Error()
		m.inputDialog = nil
		m.focus = FocusKeyList
		return m, nil

	case ClipboardCopyMsg:
		// Successfully copied to clipboard - clear any previous error
		m.err = ""
		return m, nil

	case ClipboardErrorMsg:
		m.err = fmt.Sprintf("Failed to copy to clipboard: %v", msg.Err)
		return m, nil

	case ErrorMsg:
		m.state = StateError
		m.err = msg.Err
		return m, nil
	}

	return m, nil
}

func (m *Model) handleConfirmResult(msg dialog.ConfirmResultMsg) (tea.Model, tea.Cmd) {
	m.confirmDialog = nil
	m.focus = FocusKeyList

	if !msg.Result {
		return m, nil
	}

	// Extract key from context
	key, ok := ExtractDeleteContext(msg.Context)
	if !ok {
		return m, nil
	}

	return m, DeleteKeyCmd(m.mcClient, key)
}

func (m *Model) handleInputResult(msg dialog.InputResultMsg) (tea.Model, tea.Cmd) {
	m.inputDialog = nil
	m.focus = FocusKeyList

	if msg.Cancelled {
		return m, nil
	}

	// Check if this is a batch delete confirmation
	if batchMsg := ProcessBatchInputResult(msg); batchMsg != nil {
		// Clear selection after initiating batch delete
		m.keyList.ClearSelection()
		return m, BatchDeleteCmd(m.mcClient, batchMsg.Keys)
	}

	// Check if this is the second step (value input) of new key creation
	if key, ok := ExtractNewKeyContext(msg.Context); ok {
		// Create key with the entered value
		return m, NewKeyCmd(m.mcClient, NewKeyRequest{
			Key:   key,
			Value: msg.Value,
			TTL:   0,
		})
	}

	// First step: key name entered, now ask for value
	m.inputDialog = CreateValueInputDialog(msg.Value)
	m.inputDialog.SetSize(m.width, m.height)
	m.focus = FocusDialog
	return m, m.inputDialog.Init()
}

func (m *Model) handleEditorSave(msg editor.EditorSaveMsg) (tea.Model, tea.Cmd) {
	m.editor = nil
	m.focus = FocusKeyList

	// Save the edited value with CAS if available (unified client supports CAS)
	if m.currentCASItem != nil && m.mcClient != nil {
		return m, m.saveValueWithCASCmd(msg.Key, msg.Value, m.currentCASItem)
	}
	return m, m.saveValueCmd(msg.Key, msg.Value)
}
