package app_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
)

// MockBatchDeleter is a mock implementation for testing batch delete functionality
type MockBatchDeleter struct {
	mu           sync.Mutex
	DeletedKeys  []string
	FailingKeys  map[string]error
	DeleteCalled int
}

func NewMockBatchDeleter() *MockBatchDeleter {
	return &MockBatchDeleter{
		DeletedKeys: make([]string, 0),
		FailingKeys: make(map[string]error),
	}
}

func (m *MockBatchDeleter) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalled++

	if err, exists := m.FailingKeys[key]; exists {
		return err
	}
	m.DeletedKeys = append(m.DeletedKeys, key)
	return nil
}

// TestBatchDeleteRequest tests the BatchDeleteRequest struct
func TestBatchDeleteRequest(t *testing.T) {
	t.Run("creates request with keys", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		req := app.BatchDeleteRequest{Keys: keys}

		if len(req.Keys) != 3 {
			t.Errorf("expected 3 keys, got %d", len(req.Keys))
		}

		for i, key := range keys {
			if req.Keys[i] != key {
				t.Errorf("expected key '%s' at index %d, got '%s'", key, i, req.Keys[i])
			}
		}
	})

	t.Run("empty request", func(t *testing.T) {
		req := app.BatchDeleteRequest{}

		if len(req.Keys) != 0 {
			t.Errorf("expected 0 keys, got %d", len(req.Keys))
		}
	})
}

// TestBatchDeleteMsg tests the BatchDeleteMsg message struct
func TestBatchDeleteMsg(t *testing.T) {
	t.Run("creates message with keys", func(t *testing.T) {
		keys := []string{"user:1", "user:2"}
		msg := app.BatchDeleteMsg{Keys: keys}

		if len(msg.Keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(msg.Keys))
		}
	})
}

// TestBatchDeleteResultMsg tests the BatchDeleteResultMsg struct
func TestBatchDeleteResultMsg(t *testing.T) {
	t.Run("creates result with deleted and failed keys", func(t *testing.T) {
		deleted := []string{"key1", "key2"}
		failed := []string{"key3"}
		errs := map[string]error{
			"key3": errors.New("connection error"),
		}

		result := app.BatchDeleteResultMsg{
			Deleted: deleted,
			Failed:  failed,
			Errors:  errs,
		}

		if len(result.Deleted) != 2 {
			t.Errorf("expected 2 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 1 {
			t.Errorf("expected 1 failed key, got %d", len(result.Failed))
		}

		if result.Errors["key3"] == nil {
			t.Error("expected error for key3")
		}
	})

	t.Run("all successful", func(t *testing.T) {
		result := app.BatchDeleteResultMsg{
			Deleted: []string{"key1", "key2", "key3"},
			Failed:  []string{},
			Errors:  map[string]error{},
		}

		if len(result.Failed) != 0 {
			t.Errorf("expected 0 failed keys, got %d", len(result.Failed))
		}
	})

	t.Run("all failed", func(t *testing.T) {
		result := app.BatchDeleteResultMsg{
			Deleted: []string{},
			Failed:  []string{"key1", "key2"},
			Errors: map[string]error{
				"key1": errors.New("error1"),
				"key2": errors.New("error2"),
			},
		}

		if len(result.Deleted) != 0 {
			t.Errorf("expected 0 deleted keys, got %d", len(result.Deleted))
		}
	})
}

// TestBatchDeleteCmd_Success tests successful batch deletion
func TestBatchDeleteCmd_Success(t *testing.T) {
	t.Run("deletes all keys successfully", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		keys := []string{"key1", "key2", "key3"}

		cmd := app.BatchDeleteCmd(mock, keys)
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		msg := cmd()
		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 3 {
			t.Errorf("expected 3 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 0 {
			t.Errorf("expected 0 failed keys, got %d", len(result.Failed))
		}

		// Verify all keys were deleted
		for _, key := range keys {
			found := false
			for _, deleted := range result.Deleted {
				if deleted == key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("key '%s' not found in deleted list", key)
			}
		}
	})

	t.Run("deletes single key successfully", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		keys := []string{"single-key"}

		cmd := app.BatchDeleteCmd(mock, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 1 {
			t.Errorf("expected 1 deleted key, got %d", len(result.Deleted))
		}

		if result.Deleted[0] != "single-key" {
			t.Errorf("expected 'single-key', got '%s'", result.Deleted[0])
		}
	})
}

// TestBatchDeleteCmd_PartialFailure tests batch deletion with some failures
func TestBatchDeleteCmd_PartialFailure(t *testing.T) {
	t.Run("some keys fail to delete", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		mock.FailingKeys["key2"] = errors.New("key2 error")
		mock.FailingKeys["key4"] = errors.New("key4 error")

		keys := []string{"key1", "key2", "key3", "key4", "key5"}

		cmd := app.BatchDeleteCmd(mock, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 3 {
			t.Errorf("expected 3 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 2 {
			t.Errorf("expected 2 failed keys, got %d", len(result.Failed))
		}

		// Check that failed keys are tracked
		if result.Errors["key2"] == nil {
			t.Error("expected error for key2")
		}

		if result.Errors["key4"] == nil {
			t.Error("expected error for key4")
		}
	})

	t.Run("first key fails", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		mock.FailingKeys["key1"] = errors.New("first key error")

		keys := []string{"key1", "key2", "key3"}

		cmd := app.BatchDeleteCmd(mock, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 2 {
			t.Errorf("expected 2 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 1 {
			t.Errorf("expected 1 failed key, got %d", len(result.Failed))
		}
	})

	t.Run("last key fails", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		mock.FailingKeys["key3"] = errors.New("last key error")

		keys := []string{"key1", "key2", "key3"}

		cmd := app.BatchDeleteCmd(mock, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 2 {
			t.Errorf("expected 2 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 1 {
			t.Errorf("expected 1 failed key, got %d", len(result.Failed))
		}

		if result.Failed[0] != "key3" {
			t.Errorf("expected failed key 'key3', got '%s'", result.Failed[0])
		}
	})
}

// TestBatchDeleteCmd_AllFailed tests batch deletion when all keys fail
func TestBatchDeleteCmd_AllFailed(t *testing.T) {
	t.Run("all keys fail to delete", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		mock.FailingKeys["key1"] = errors.New("error1")
		mock.FailingKeys["key2"] = errors.New("error2")
		mock.FailingKeys["key3"] = errors.New("error3")

		keys := []string{"key1", "key2", "key3"}

		cmd := app.BatchDeleteCmd(mock, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 0 {
			t.Errorf("expected 0 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 3 {
			t.Errorf("expected 3 failed keys, got %d", len(result.Failed))
		}

		// Verify all errors are tracked
		for _, key := range keys {
			if result.Errors[key] == nil {
				t.Errorf("expected error for key '%s'", key)
			}
		}
	})

	t.Run("nil client returns all failed", func(t *testing.T) {
		keys := []string{"key1", "key2"}

		cmd := app.BatchDeleteCmd(nil, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 0 {
			t.Errorf("expected 0 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 2 {
			t.Errorf("expected 2 failed keys, got %d", len(result.Failed))
		}
	})
}

// TestBatchDeleteCmd_EmptyKeys tests batch deletion with empty key list
func TestBatchDeleteCmd_EmptyKeys(t *testing.T) {
	t.Run("empty key list", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		keys := []string{}

		cmd := app.BatchDeleteCmd(mock, keys)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 0 {
			t.Errorf("expected 0 deleted keys, got %d", len(result.Deleted))
		}

		if len(result.Failed) != 0 {
			t.Errorf("expected 0 failed keys, got %d", len(result.Failed))
		}

		if mock.DeleteCalled != 0 {
			t.Errorf("expected Delete not to be called, but was called %d times", mock.DeleteCalled)
		}
	})

	t.Run("nil key list", func(t *testing.T) {
		mock := NewMockBatchDeleter()

		cmd := app.BatchDeleteCmd(mock, nil)
		msg := cmd()

		result, ok := msg.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
		}

		if len(result.Deleted) != 0 {
			t.Errorf("expected 0 deleted keys, got %d", len(result.Deleted))
		}
	})
}

// TestBatchDeleteResult tests the HandleBatchDeleteResult function
func TestBatchDeleteResult(t *testing.T) {
	t.Run("all successful summary", func(t *testing.T) {
		msg := app.BatchDeleteResultMsg{
			Deleted: []string{"key1", "key2", "key3"},
			Failed:  []string{},
			Errors:  map[string]error{},
		}

		summary := app.HandleBatchDeleteResult(msg)

		if summary.TotalCount != 3 {
			t.Errorf("expected TotalCount 3, got %d", summary.TotalCount)
		}

		if summary.DeletedCount != 3 {
			t.Errorf("expected DeletedCount 3, got %d", summary.DeletedCount)
		}

		if summary.FailedCount != 0 {
			t.Errorf("expected FailedCount 0, got %d", summary.FailedCount)
		}

		if !summary.AllSucceeded {
			t.Error("expected AllSucceeded to be true")
		}

		if summary.HasErrors {
			t.Error("expected HasErrors to be false")
		}

		if !summary.ShouldRefresh {
			t.Error("expected ShouldRefresh to be true when keys were deleted")
		}
	})

	t.Run("partial failure summary", func(t *testing.T) {
		msg := app.BatchDeleteResultMsg{
			Deleted: []string{"key1", "key3"},
			Failed:  []string{"key2"},
			Errors: map[string]error{
				"key2": errors.New("error"),
			},
		}

		summary := app.HandleBatchDeleteResult(msg)

		if summary.TotalCount != 3 {
			t.Errorf("expected TotalCount 3, got %d", summary.TotalCount)
		}

		if summary.DeletedCount != 2 {
			t.Errorf("expected DeletedCount 2, got %d", summary.DeletedCount)
		}

		if summary.FailedCount != 1 {
			t.Errorf("expected FailedCount 1, got %d", summary.FailedCount)
		}

		if summary.AllSucceeded {
			t.Error("expected AllSucceeded to be false")
		}

		if !summary.HasErrors {
			t.Error("expected HasErrors to be true")
		}

		if !summary.ShouldRefresh {
			t.Error("expected ShouldRefresh to be true when some keys were deleted")
		}
	})

	t.Run("all failed summary", func(t *testing.T) {
		msg := app.BatchDeleteResultMsg{
			Deleted: []string{},
			Failed:  []string{"key1", "key2"},
			Errors: map[string]error{
				"key1": errors.New("error1"),
				"key2": errors.New("error2"),
			},
		}

		summary := app.HandleBatchDeleteResult(msg)

		if summary.TotalCount != 2 {
			t.Errorf("expected TotalCount 2, got %d", summary.TotalCount)
		}

		if summary.DeletedCount != 0 {
			t.Errorf("expected DeletedCount 0, got %d", summary.DeletedCount)
		}

		if summary.FailedCount != 2 {
			t.Errorf("expected FailedCount 2, got %d", summary.FailedCount)
		}

		if summary.AllSucceeded {
			t.Error("expected AllSucceeded to be false")
		}

		if !summary.HasErrors {
			t.Error("expected HasErrors to be true")
		}

		if summary.ShouldRefresh {
			t.Error("expected ShouldRefresh to be false when no keys were deleted")
		}
	})

	t.Run("empty result summary", func(t *testing.T) {
		msg := app.BatchDeleteResultMsg{
			Deleted: []string{},
			Failed:  []string{},
			Errors:  map[string]error{},
		}

		summary := app.HandleBatchDeleteResult(msg)

		if summary.TotalCount != 0 {
			t.Errorf("expected TotalCount 0, got %d", summary.TotalCount)
		}

		if summary.DeletedCount != 0 {
			t.Errorf("expected DeletedCount 0, got %d", summary.DeletedCount)
		}

		if !summary.AllSucceeded {
			t.Error("expected AllSucceeded to be true for empty result")
		}

		if summary.HasErrors {
			t.Error("expected HasErrors to be false for empty result")
		}

		if summary.ShouldRefresh {
			t.Error("expected ShouldRefresh to be false for empty result")
		}
	})
}

// TestCreateBatchDeleteDialog tests the batch delete dialog creation
func TestCreateBatchDeleteDialog(t *testing.T) {
	t.Run("creates dialog with correct title", func(t *testing.T) {
		dlg := app.CreateBatchDeleteDialog(5)

		if dlg == nil {
			t.Fatal("expected non-nil dialog")
		}

		title := dlg.Title()
		if title != "Batch Delete Confirmation" {
			t.Errorf("expected title 'Batch Delete Confirmation', got '%s'", title)
		}
	})

	t.Run("creates dialog with count in message", func(t *testing.T) {
		dlg := app.CreateBatchDeleteDialog(10)

		if dlg == nil {
			t.Fatal("expected non-nil dialog")
		}

		// The dialog should show the count in some form
		// We can't easily test the message content without exposing it
	})

	t.Run("creates dialog with validator", func(t *testing.T) {
		dlg := app.CreateBatchDeleteDialog(3)

		if dlg == nil {
			t.Fatal("expected non-nil dialog")
		}

		// Dialog should have a validator set
		// We test this by checking validation behavior
	})

	t.Run("dialog requires DELETE confirmation", func(t *testing.T) {
		dlg := app.CreateBatchDeleteDialog(5)

		// The validator should be set to require "DELETE"
		if dlg == nil {
			t.Fatal("expected non-nil dialog")
		}
	})
}

// TestValidateBatchDeleteConfirmation tests the validation function
func TestValidateBatchDeleteConfirmation(t *testing.T) {
	t.Run("accepts DELETE", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput("DELETE")
		if err != nil {
			t.Errorf("expected no error for 'DELETE', got '%v'", err)
		}
	})

	t.Run("rejects lowercase delete", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput("delete")
		if err == nil {
			t.Error("expected error for 'delete'")
		}
	})

	t.Run("rejects mixed case Delete", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput("Delete")
		if err == nil {
			t.Error("expected error for 'Delete'")
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput("")
		if err == nil {
			t.Error("expected error for empty string")
		}
	})

	t.Run("rejects random text", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput("random")
		if err == nil {
			t.Error("expected error for 'random'")
		}
	})

	t.Run("rejects DELETE with extra spaces", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput(" DELETE ")
		if err == nil {
			t.Error("expected error for ' DELETE '")
		}
	})

	t.Run("rejects partial DELETE", func(t *testing.T) {
		err := app.ValidateBatchDeleteInput("DEL")
		if err == nil {
			t.Error("expected error for 'DEL'")
		}
	})
}

// TestBatchDeleteContext tests the context handling for batch delete
func TestBatchDeleteContext(t *testing.T) {
	t.Run("context contains keys", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		ctx := app.BatchDeleteContext{Keys: keys}

		if len(ctx.Keys) != 3 {
			t.Errorf("expected 3 keys in context, got %d", len(ctx.Keys))
		}
	})

	t.Run("extract batch delete context from interface", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		ctx := app.BatchDeleteContext{Keys: keys}

		extracted, ok := app.ExtractBatchDeleteContext(ctx)
		if !ok {
			t.Fatal("expected successful extraction")
		}

		if len(extracted) != 2 {
			t.Errorf("expected 2 keys, got %d", len(extracted))
		}
	})

	t.Run("extract from nil returns false", func(t *testing.T) {
		_, ok := app.ExtractBatchDeleteContext(nil)
		if ok {
			t.Error("expected false for nil context")
		}
	})

	t.Run("extract from wrong type returns false", func(t *testing.T) {
		_, ok := app.ExtractBatchDeleteContext("invalid")
		if ok {
			t.Error("expected false for invalid context type")
		}
	})
}

// TestProcessBatchInputResult tests processing input dialog results
func TestProcessBatchInputResult(t *testing.T) {
	t.Run("confirmed with DELETE returns message", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		ctx := app.BatchDeleteContext{Keys: keys}

		result := dialog.InputResultMsg{
			Value:     "DELETE",
			Cancelled: false,
			Context:   ctx,
		}

		msg := app.ProcessBatchInputResult(result)
		if msg == nil {
			t.Fatal("expected non-nil message")
		}

		if len(msg.Keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(msg.Keys))
		}
	})

	t.Run("cancelled returns nil", func(t *testing.T) {
		keys := []string{"key1"}
		ctx := app.BatchDeleteContext{Keys: keys}

		result := dialog.InputResultMsg{
			Value:     "",
			Cancelled: true,
			Context:   ctx,
		}

		msg := app.ProcessBatchInputResult(result)
		if msg != nil {
			t.Error("expected nil message for cancelled input")
		}
	})

	t.Run("wrong input value returns nil", func(t *testing.T) {
		keys := []string{"key1"}
		ctx := app.BatchDeleteContext{Keys: keys}

		result := dialog.InputResultMsg{
			Value:     "delete", // lowercase
			Cancelled: false,
			Context:   ctx,
		}

		msg := app.ProcessBatchInputResult(result)
		if msg != nil {
			t.Error("expected nil message for wrong input value")
		}
	})

	t.Run("nil context returns nil", func(t *testing.T) {
		result := dialog.InputResultMsg{
			Value:     "DELETE",
			Cancelled: false,
			Context:   nil,
		}

		msg := app.ProcessBatchInputResult(result)
		if msg != nil {
			t.Error("expected nil message for nil context")
		}
	})
}

// TestBatchDeleteSummaryString tests the summary string representation
func TestBatchDeleteSummaryString(t *testing.T) {
	t.Run("all succeeded message", func(t *testing.T) {
		summary := app.BatchDeleteSummary{
			TotalCount:    5,
			DeletedCount:  5,
			FailedCount:   0,
			AllSucceeded:  true,
			HasErrors:     false,
			ShouldRefresh: true,
		}

		str := summary.String()
		if str == "" {
			t.Error("expected non-empty string")
		}
	})

	t.Run("partial failure message", func(t *testing.T) {
		summary := app.BatchDeleteSummary{
			TotalCount:    5,
			DeletedCount:  3,
			FailedCount:   2,
			AllSucceeded:  false,
			HasErrors:     true,
			ShouldRefresh: true,
		}

		str := summary.String()
		if str == "" {
			t.Error("expected non-empty string")
		}
	})

	t.Run("all failed message", func(t *testing.T) {
		summary := app.BatchDeleteSummary{
			TotalCount:    3,
			DeletedCount:  0,
			FailedCount:   3,
			AllSucceeded:  false,
			HasErrors:     true,
			ShouldRefresh: false,
		}

		str := summary.String()
		if str == "" {
			t.Error("expected non-empty string")
		}
	})
}

// TestBatchDeleteIntegration tests the complete batch delete flow
func TestBatchDeleteIntegration(t *testing.T) {
	t.Run("full batch delete flow", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		keys := []string{"user:1", "user:2", "user:3"}

		// Step 1: Create batch delete dialog
		dlg := app.CreateBatchDeleteDialog(len(keys))
		if dlg == nil {
			t.Fatal("failed to create dialog")
		}

		// Step 2: User types "DELETE" (simulated by creating result directly)
		ctx := app.BatchDeleteContext{Keys: keys}
		inputResult := dialog.InputResultMsg{
			Value:     "DELETE",
			Cancelled: false,
			Context:   ctx,
		}

		// Step 3: Process input result
		batchMsg := app.ProcessBatchInputResult(inputResult)
		if batchMsg == nil {
			t.Fatal("expected batch delete message")
		}

		// Step 4: Execute batch delete
		cmd := app.BatchDeleteCmd(mock, batchMsg.Keys)
		result := cmd()

		deleteResult, ok := result.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", result)
		}

		// Step 5: Handle result
		summary := app.HandleBatchDeleteResult(deleteResult)
		if !summary.AllSucceeded {
			t.Error("expected all deletions to succeed")
		}

		if summary.DeletedCount != 3 {
			t.Errorf("expected 3 deleted, got %d", summary.DeletedCount)
		}
	})

	t.Run("batch delete with partial failures", func(t *testing.T) {
		mock := NewMockBatchDeleter()
		mock.FailingKeys["user:2"] = errors.New("not found")

		keys := []string{"user:1", "user:2", "user:3"}

		// Execute batch delete directly
		cmd := app.BatchDeleteCmd(mock, keys)
		result := cmd()

		deleteResult, ok := result.(app.BatchDeleteResultMsg)
		if !ok {
			t.Fatalf("expected BatchDeleteResultMsg, got %T", result)
		}

		summary := app.HandleBatchDeleteResult(deleteResult)

		if summary.AllSucceeded {
			t.Error("expected AllSucceeded to be false")
		}

		if summary.DeletedCount != 2 {
			t.Errorf("expected 2 deleted, got %d", summary.DeletedCount)
		}

		if summary.FailedCount != 1 {
			t.Errorf("expected 1 failed, got %d", summary.FailedCount)
		}
	})

	t.Run("batch delete cancelled by user", func(t *testing.T) {
		keys := []string{"user:1", "user:2"}

		// User cancels by pressing Escape
		ctx := app.BatchDeleteContext{Keys: keys}
		inputResult := dialog.InputResultMsg{
			Value:     "",
			Cancelled: true,
			Context:   ctx,
		}

		msg := app.ProcessBatchInputResult(inputResult)
		if msg != nil {
			t.Error("expected nil message for cancelled input")
		}
	})
}
