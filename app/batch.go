package app

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
)

// BatchDeleteRequest holds the request parameters for batch deletion.
type BatchDeleteRequest struct {
	Keys []string
}

// BatchDeleteMsg is the message that triggers a batch delete operation.
type BatchDeleteMsg struct {
	Keys []string
}

// BatchDeleteResultMsg holds the result of a batch delete operation.
type BatchDeleteResultMsg struct {
	Deleted []string         // Keys that were successfully deleted
	Failed  []string         // Keys that failed to delete
	Errors  map[string]error // Error messages for failed keys
}

// BatchDeleter is an interface for deleting keys from Memcached.
// This allows for easy mocking in tests.
type BatchDeleter interface {
	Delete(key string) error
}

// BatchDeleteContext holds contextual information for a batch delete operation.
// Used to pass data between dialog confirmation and batch delete execution.
type BatchDeleteContext struct {
	Keys []string
}

// BatchDeleteSummary provides a summary of the batch delete operation result.
type BatchDeleteSummary struct {
	TotalCount    int      // Total number of keys attempted
	DeletedCount  int      // Number of successfully deleted keys
	FailedCount   int      // Number of failed deletions
	AllSucceeded  bool     // True if all deletions succeeded
	HasErrors     bool     // True if any deletion failed
	ShouldRefresh bool     // Whether the key list should be refreshed
	FailedKeys    []string // List of keys that failed to delete
}

// String returns a human-readable summary of the batch delete result.
func (s BatchDeleteSummary) String() string {
	if s.TotalCount == 0 {
		return "No keys to delete"
	}

	if s.AllSucceeded {
		if s.DeletedCount == 1 {
			return "Successfully deleted 1 key"
		}
		return fmt.Sprintf("Successfully deleted %d keys", s.DeletedCount)
	}

	if s.DeletedCount == 0 {
		if s.FailedCount == 1 {
			return "Failed to delete 1 key"
		}
		return fmt.Sprintf("Failed to delete all %d keys", s.FailedCount)
	}

	return fmt.Sprintf("Deleted %d keys, %d failed", s.DeletedCount, s.FailedCount)
}

// BatchDeleteCmd creates a tea.Cmd that deletes multiple keys from Memcached.
// Returns BatchDeleteResultMsg with the results of all deletion attempts.
func BatchDeleteCmd(client BatchDeleter, keys []string) tea.Cmd {
	return func() tea.Msg {
		result := BatchDeleteResultMsg{
			Deleted: make([]string, 0),
			Failed:  make([]string, 0),
			Errors:  make(map[string]error),
		}

		// Handle nil or empty keys
		if len(keys) == 0 {
			return result
		}

		// Handle nil client
		if client == nil {
			clientErr := errors.New("client is nil")
			for _, key := range keys {
				result.Failed = append(result.Failed, key)
				result.Errors[key] = clientErr
			}
			return result
		}

		// Delete each key and track results
		for _, key := range keys {
			err := client.Delete(key)
			if err != nil {
				result.Failed = append(result.Failed, key)
				result.Errors[key] = err
			} else {
				result.Deleted = append(result.Deleted, key)
			}
		}

		return result
	}
}

// HandleBatchDeleteResult processes a BatchDeleteResultMsg and returns a summary.
func HandleBatchDeleteResult(msg BatchDeleteResultMsg) BatchDeleteSummary {
	deletedCount := len(msg.Deleted)
	failedCount := len(msg.Failed)
	totalCount := deletedCount + failedCount

	return BatchDeleteSummary{
		TotalCount:    totalCount,
		DeletedCount:  deletedCount,
		FailedCount:   failedCount,
		AllSucceeded:  failedCount == 0,
		HasErrors:     failedCount > 0,
		ShouldRefresh: deletedCount > 0,
		FailedKeys:    msg.Failed,
	}
}

// CreateBatchDeleteDialog creates an input dialog for batch delete confirmation.
// The dialog requires the user to type "DELETE" to confirm the operation.
func CreateBatchDeleteDialog(_ int) *dialog.InputDialog {
	title := "Batch Delete Confirmation"

	dlg := dialog.NewInput(title).
		WithPlaceholder("Type DELETE to confirm").
		WithValidator(ValidateBatchDeleteInput)

	return dlg
}

// CreateBatchDeleteDialogWithKeys creates an input dialog with the keys context attached.
// This allows the keys to be retrieved when the user confirms the deletion.
func CreateBatchDeleteDialogWithKeys(keys []string) *dialog.InputDialog {
	count := len(keys)
	ctx := BatchDeleteContext{Keys: keys}

	title := "Batch Delete Confirmation"

	dlg := dialog.NewInput(title).
		WithPlaceholder("Type DELETE to confirm").
		WithValidator(ValidateBatchDeleteInput).
		WithContext(ctx)

	// Set a message that includes the count
	_ = count // The count could be displayed in an extended version

	return dlg
}

// ValidateBatchDeleteInput validates that the input matches "DELETE" exactly.
// This is a safety measure to prevent accidental batch deletions.
func ValidateBatchDeleteInput(input string) error {
	if input != "DELETE" {
		return errors.New("type DELETE (all caps) to confirm")
	}
	return nil
}

// ExtractBatchDeleteContext extracts the keys from a BatchDeleteContext.
// Returns the keys slice and a boolean indicating success.
func ExtractBatchDeleteContext(ctx interface{}) ([]string, bool) {
	if ctx == nil {
		return nil, false
	}

	if bdc, ok := ctx.(BatchDeleteContext); ok {
		return bdc.Keys, true
	}

	return nil, false
}

// ProcessBatchInputResult processes an input dialog result for batch deletion.
// Returns a BatchDeleteMsg if confirmed with "DELETE", or nil if canceled or invalid.
func ProcessBatchInputResult(result dialog.InputResultMsg) *BatchDeleteMsg {
	// User canceled
	if result.Canceled {
		return nil
	}

	// Validate input
	if ValidateBatchDeleteInput(result.Value) != nil {
		return nil
	}

	// Extract keys from context
	keys, ok := ExtractBatchDeleteContext(result.Context)
	if !ok {
		return nil
	}

	return &BatchDeleteMsg{Keys: keys}
}

// HandleBatchDeleteConfirm processes a batch delete confirmation and returns the command.
// This is called after the user confirms deletion by typing "DELETE".
func HandleBatchDeleteConfirm(client BatchDeleter, msg BatchDeleteMsg) tea.Cmd {
	return BatchDeleteCmd(client, msg.Keys)
}
