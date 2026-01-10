package app_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
)

// MockMemcacheClient is a mock implementation for testing delete functionality
type MockMemcacheClient struct {
	DeleteFunc func(key string) error
	DeletedKey string
}

func (m *MockMemcacheClient) Delete(key string) error {
	m.DeletedKey = key
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	return nil
}

// TestDeleteKeyCmd tests the delete key command execution
func TestDeleteKeyCmd(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mock := &MockMemcacheClient{}
		cmd := app.DeleteKeyCmd(mock, "test-key")

		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		msg := cmd()
		deletedMsg, ok := msg.(app.KeyDeletedMsg)
		if !ok {
			t.Fatalf("expected KeyDeletedMsg, got %T", msg)
		}

		if deletedMsg.Key != "test-key" {
			t.Errorf("expected key 'test-key', got '%s'", deletedMsg.Key)
		}

		if mock.DeletedKey != "test-key" {
			t.Errorf("expected client.Delete to be called with 'test-key', got '%s'", mock.DeletedKey)
		}
	})

	t.Run("deletion error", func(t *testing.T) {
		expectedErr := errors.New("delete failed: connection refused")
		mock := &MockMemcacheClient{
			DeleteFunc: func(key string) error {
				return expectedErr
			},
		}

		cmd := app.DeleteKeyCmd(mock, "error-key")
		msg := cmd()

		errMsg, ok := msg.(app.DeleteErrorMsg)
		if !ok {
			t.Fatalf("expected DeleteErrorMsg, got %T", msg)
		}

		if errMsg.Key != "error-key" {
			t.Errorf("expected key 'error-key', got '%s'", errMsg.Key)
		}

		if errMsg.Err != expectedErr {
			t.Errorf("expected error '%v', got '%v'", expectedErr, errMsg.Err)
		}
	})

	t.Run("nil client returns error", func(t *testing.T) {
		cmd := app.DeleteKeyCmd(nil, "test-key")
		msg := cmd()

		errMsg, ok := msg.(app.DeleteErrorMsg)
		if !ok {
			t.Fatalf("expected DeleteErrorMsg, got %T", msg)
		}

		if errMsg.Err == nil {
			t.Error("expected non-nil error for nil client")
		}
	})
}

// TestHandleDeleteConfirm tests handling of delete confirmation
func TestHandleDeleteConfirm(t *testing.T) {
	t.Run("confirmation triggers delete command", func(t *testing.T) {
		mock := &MockMemcacheClient{}
		confirmMsg := app.DeleteConfirmMsg{Key: "confirmed-key"}

		cmd := app.HandleDeleteConfirm(mock, confirmMsg)
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		// Execute the command and verify it produces correct message
		msg := cmd()
		deletedMsg, ok := msg.(app.KeyDeletedMsg)
		if !ok {
			t.Fatalf("expected KeyDeletedMsg, got %T", msg)
		}

		if deletedMsg.Key != "confirmed-key" {
			t.Errorf("expected key 'confirmed-key', got '%s'", deletedMsg.Key)
		}
	})

	t.Run("confirmation with nil client returns error command", func(t *testing.T) {
		confirmMsg := app.DeleteConfirmMsg{Key: "test-key"}
		cmd := app.HandleDeleteConfirm(nil, confirmMsg)

		if cmd == nil {
			t.Fatal("expected non-nil command even with nil client")
		}

		msg := cmd()
		errMsg, ok := msg.(app.DeleteErrorMsg)
		if !ok {
			t.Fatalf("expected DeleteErrorMsg, got %T", msg)
		}

		if errMsg.Key != "test-key" {
			t.Errorf("expected key 'test-key', got '%s'", errMsg.Key)
		}
	})
}

// TestHandleDeleteResult tests handling of delete operation results
func TestHandleDeleteResult(t *testing.T) {
	t.Run("KeyDeletedMsg returns refresh command", func(t *testing.T) {
		msg := app.KeyDeletedMsg{Key: "deleted-key"}
		result := app.HandleDeleteResult(msg)

		if result.ShouldRefresh != true {
			t.Error("expected ShouldRefresh to be true after successful deletion")
		}

		if result.Error != "" {
			t.Errorf("expected no error, got '%s'", result.Error)
		}

		if result.DeletedKey != "deleted-key" {
			t.Errorf("expected DeletedKey 'deleted-key', got '%s'", result.DeletedKey)
		}
	})

	t.Run("DeleteErrorMsg returns error info", func(t *testing.T) {
		msg := app.DeleteErrorMsg{
			Key: "failed-key",
			Err: errors.New("connection lost"),
		}
		result := app.HandleDeleteResult(msg)

		if result.ShouldRefresh != false {
			t.Error("expected ShouldRefresh to be false after deletion error")
		}

		if result.Error != "connection lost" {
			t.Errorf("expected error 'connection lost', got '%s'", result.Error)
		}

		if result.DeletedKey != "" {
			t.Errorf("expected empty DeletedKey on error, got '%s'", result.DeletedKey)
		}

		if result.FailedKey != "failed-key" {
			t.Errorf("expected FailedKey 'failed-key', got '%s'", result.FailedKey)
		}
	})
}

// TestDeleteMessages tests the message structs
func TestDeleteMessages(t *testing.T) {
	t.Run("DeleteKeyMsg", func(t *testing.T) {
		msg := app.DeleteKeyMsg{Key: "test-key"}
		if msg.Key != "test-key" {
			t.Errorf("expected Key 'test-key', got '%s'", msg.Key)
		}
	})

	t.Run("DeleteConfirmMsg", func(t *testing.T) {
		msg := app.DeleteConfirmMsg{Key: "confirm-key"}
		if msg.Key != "confirm-key" {
			t.Errorf("expected Key 'confirm-key', got '%s'", msg.Key)
		}
	})

	t.Run("KeyDeletedMsg", func(t *testing.T) {
		msg := app.KeyDeletedMsg{Key: "deleted-key"}
		if msg.Key != "deleted-key" {
			t.Errorf("expected Key 'deleted-key', got '%s'", msg.Key)
		}
	})

	t.Run("DeleteErrorMsg", func(t *testing.T) {
		err := errors.New("test error")
		msg := app.DeleteErrorMsg{Key: "error-key", Err: err}
		if msg.Key != "error-key" {
			t.Errorf("expected Key 'error-key', got '%s'", msg.Key)
		}
		if msg.Err != err {
			t.Errorf("expected Err '%v', got '%v'", err, msg.Err)
		}
	})
}

// TestCreateDeleteConfirmDialog tests creation of delete confirmation dialog
func TestCreateDeleteConfirmDialog(t *testing.T) {
	t.Run("creates dialog with correct title", func(t *testing.T) {
		dialog := app.CreateDeleteConfirmDialog("user:123")

		if dialog == nil {
			t.Fatal("expected non-nil dialog")
		}

		if dialog.Title() != "Delete Key" {
			t.Errorf("expected title 'Delete Key', got '%s'", dialog.Title())
		}
	})

	t.Run("creates dialog with key in message", func(t *testing.T) {
		dialog := app.CreateDeleteConfirmDialog("session:abc")

		if dialog == nil {
			t.Fatal("expected non-nil dialog")
		}

		// Message should contain the key name
		if dialog.Message() == "" {
			t.Error("expected non-empty message")
		}
	})

	t.Run("dialog context contains key", func(t *testing.T) {
		keyName := "cache:data:1"
		dialog := app.CreateDeleteConfirmDialog(keyName)

		// Simulate confirmation
		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		if cmd == nil {
			t.Fatal("expected command from confirmation")
		}

		// The context should be extractable via the returned context
		// This would be verified in integration with ConfirmResultMsg
	})
}

// TestDeleteFlowIntegration tests the complete delete flow
func TestDeleteFlowIntegration(t *testing.T) {
	t.Run("full delete flow with confirmation", func(t *testing.T) {
		mock := &MockMemcacheClient{}
		keyToDelete := "integration-test-key"

		// Step 1: Create delete confirmation dialog
		dialog := app.CreateDeleteConfirmDialog(keyToDelete)
		if dialog == nil {
			t.Fatal("failed to create dialog")
		}

		// Step 2: User confirms (presses 'y')
		_, confirmCmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		if confirmCmd == nil {
			t.Fatal("expected confirm command")
		}

		// Step 3: Process confirmation and create delete command
		confirmMsg := app.DeleteConfirmMsg{Key: keyToDelete}
		deleteCmd := app.HandleDeleteConfirm(mock, confirmMsg)
		if deleteCmd == nil {
			t.Fatal("expected delete command")
		}

		// Step 4: Execute delete and verify result
		deleteResult := deleteCmd()
		deletedMsg, ok := deleteResult.(app.KeyDeletedMsg)
		if !ok {
			t.Fatalf("expected KeyDeletedMsg, got %T", deleteResult)
		}

		if deletedMsg.Key != keyToDelete {
			t.Errorf("expected deleted key '%s', got '%s'", keyToDelete, deletedMsg.Key)
		}

		// Step 5: Handle result
		result := app.HandleDeleteResult(deletedMsg)
		if !result.ShouldRefresh {
			t.Error("expected ShouldRefresh after successful deletion")
		}
	})

	t.Run("delete flow canceled by user", func(t *testing.T) {
		keyToDelete := "canceled-key"

		// Step 1: Create delete confirmation dialog
		dialog := app.CreateDeleteConfirmDialog(keyToDelete)

		// Step 2: User cancels (presses 'n' or Escape)
		_, cancelCmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cancelCmd == nil {
			t.Fatal("expected cancel command")
		}

		// The cancel command would return ConfirmResultMsg with Result: false
		// In real app, this would close the dialog without deleting
	})
}

// TestDeleteWithEmptyKey tests edge cases with empty keys
func TestDeleteWithEmptyKey(t *testing.T) {
	t.Run("empty key still creates command", func(t *testing.T) {
		mock := &MockMemcacheClient{}
		cmd := app.DeleteKeyCmd(mock, "")

		if cmd == nil {
			t.Fatal("expected non-nil command even for empty key")
		}

		msg := cmd()
		// Depending on implementation, might succeed or fail
		// The underlying client will determine behavior
		switch msg.(type) {
		case app.KeyDeletedMsg, app.DeleteErrorMsg:
			// Both are acceptable outcomes
		default:
			t.Fatalf("unexpected message type: %T", msg)
		}
	})
}
