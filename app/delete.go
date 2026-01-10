package app

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
)

// Deleter is an interface for deleting keys from Memcached.
// This allows for easy mocking in tests.
type Deleter interface {
	Delete(key string) error
}

// DeleteResult holds the result of a delete operation for UI handling.
type DeleteResult struct {
	ShouldRefresh bool   // Whether the key list should be refreshed
	Error         string // Error message if deletion failed
	DeletedKey    string // Key that was successfully deleted
	FailedKey     string // Key that failed to delete
}

// DeleteKeyCmd creates a tea.Cmd that deletes a key from Memcached.
// Returns KeyDeletedMsg on success or DeleteErrorMsg on failure.
func DeleteKeyCmd(client Deleter, key string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return DeleteErrorMsg{
				Key: key,
				Err: errors.New("client is nil"),
			}
		}

		err := client.Delete(key)
		if err != nil {
			return DeleteErrorMsg{
				Key: key,
				Err: err,
			}
		}

		return KeyDeletedMsg{Key: key}
	}
}

// HandleDeleteConfirm processes a delete confirmation and returns the appropriate command.
// This is called after the user confirms deletion in the dialog.
func HandleDeleteConfirm(client Deleter, msg DeleteConfirmMsg) tea.Cmd {
	return DeleteKeyCmd(client, msg.Key)
}

// HandleDeleteResult processes the result of a delete operation.
// Returns a DeleteResult struct with information about what happened
// and what the UI should do next.
func HandleDeleteResult(msg tea.Msg) DeleteResult {
	switch m := msg.(type) {
	case KeyDeletedMsg:
		return DeleteResult{
			ShouldRefresh: true,
			DeletedKey:    m.Key,
		}
	case DeleteErrorMsg:
		errStr := ""
		if m.Err != nil {
			errStr = m.Err.Error()
		}
		return DeleteResult{
			ShouldRefresh: false,
			Error:         errStr,
			FailedKey:     m.Key,
		}
	default:
		return DeleteResult{}
	}
}

// CreateDeleteConfirmDialog creates a confirmation dialog for key deletion.
// The dialog includes the key name in the message and stores the key
// in the context for retrieval after confirmation.
func CreateDeleteConfirmDialog(key string) *dialog.ConfirmDialog {
	title := "Delete Key"
	message := fmt.Sprintf("Are you sure you want to delete the key?\n\n  %s\n\nThis action cannot be undone.", key)

	return dialog.NewWithContext(title, message, key)
}

// DeleteContext holds contextual information for a delete operation.
// Used to pass data between dialog confirmation and delete execution.
type DeleteContext struct {
	Key string
}

// ExtractDeleteContext extracts the key from a ConfirmResultMsg context.
// Returns the key string and a boolean indicating success.
func ExtractDeleteContext(ctx interface{}) (string, bool) {
	if ctx == nil {
		return "", false
	}

	// Check if context is a string (key directly)
	if key, ok := ctx.(string); ok {
		return key, true
	}

	// Check if context is a DeleteContext struct
	if dc, ok := ctx.(DeleteContext); ok {
		return dc.Key, true
	}

	return "", false
}

// ProcessConfirmResult processes a confirmation dialog result for deletion.
// Returns a DeleteConfirmMsg if confirmed, or nil if cancelled.
func ProcessConfirmResult(result dialog.ConfirmResultMsg) *DeleteConfirmMsg {
	if !result.Result {
		// User cancelled
		return nil
	}

	key, ok := ExtractDeleteContext(result.Context)
	if !ok {
		return nil
	}

	return &DeleteConfirmMsg{Key: key}
}
