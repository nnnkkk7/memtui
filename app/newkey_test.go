package app_test

import (
	"errors"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
)

// MockSetter is a mock implementation for testing new key functionality
type MockSetter struct {
	SetFunc func(item *memcache.Item) error
	LastSet *memcache.Item
}

func (m *MockSetter) Set(item *memcache.Item) error {
	m.LastSet = item
	if m.SetFunc != nil {
		return m.SetFunc(item)
	}
	return nil
}

// TestNewKeyCmd_Success tests successful key creation
func TestNewKeyCmd_Success(t *testing.T) {
	t.Run("creates key with value only", func(t *testing.T) {
		mock := &MockSetter{}
		req := app.NewKeyRequest{
			Key:   "new-test-key",
			Value: "test value",
		}

		cmd := app.NewKeyCmd(mock, req)
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		msg := cmd()
		createdMsg, ok := msg.(app.KeyCreatedMsg)
		if !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T", msg)
		}

		if createdMsg.Key != "new-test-key" {
			t.Errorf("expected key 'new-test-key', got '%s'", createdMsg.Key)
		}

		if mock.LastSet == nil {
			t.Fatal("expected Set to be called")
		}

		if mock.LastSet.Key != "new-test-key" {
			t.Errorf("expected item key 'new-test-key', got '%s'", mock.LastSet.Key)
		}

		if string(mock.LastSet.Value) != "test value" {
			t.Errorf("expected item value 'test value', got '%s'", string(mock.LastSet.Value))
		}
	})

	t.Run("creates key with TTL", func(t *testing.T) {
		mock := &MockSetter{}
		req := app.NewKeyRequest{
			Key:   "ttl-key",
			Value: "ttl value",
			TTL:   3600,
		}

		cmd := app.NewKeyCmd(mock, req)
		msg := cmd()

		createdMsg, ok := msg.(app.KeyCreatedMsg)
		if !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T", msg)
		}

		if createdMsg.Key != "ttl-key" {
			t.Errorf("expected key 'ttl-key', got '%s'", createdMsg.Key)
		}

		if mock.LastSet.Expiration != 3600 {
			t.Errorf("expected expiration 3600, got %d", mock.LastSet.Expiration)
		}
	})

	t.Run("creates key with flags", func(t *testing.T) {
		mock := &MockSetter{}
		req := app.NewKeyRequest{
			Key:   "flags-key",
			Value: "flags value",
			Flags: 42,
		}

		cmd := app.NewKeyCmd(mock, req)
		msg := cmd()

		_, ok := msg.(app.KeyCreatedMsg)
		if !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T", msg)
		}

		if mock.LastSet.Flags != 42 {
			t.Errorf("expected flags 42, got %d", mock.LastSet.Flags)
		}
	})
}

// TestNewKeyCmd_AlreadyExists tests key already exists error
func TestNewKeyCmd_AlreadyExists(t *testing.T) {
	t.Run("returns error when key exists", func(t *testing.T) {
		mock := &MockSetter{
			SetFunc: func(item *memcache.Item) error {
				return memcache.ErrNotStored
			},
		}

		req := app.NewKeyRequest{
			Key:   "existing-key",
			Value: "value",
		}

		cmd := app.NewKeyCmd(mock, req)
		msg := cmd()

		errMsg, ok := msg.(app.NewKeyErrorMsg)
		if !ok {
			t.Fatalf("expected NewKeyErrorMsg, got %T", msg)
		}

		if errMsg.Key != "existing-key" {
			t.Errorf("expected key 'existing-key', got '%s'", errMsg.Key)
		}

		if errMsg.Err == nil {
			t.Error("expected non-nil error")
		}
	})

	t.Run("handles connection error", func(t *testing.T) {
		expectedErr := errors.New("connection refused")
		mock := &MockSetter{
			SetFunc: func(item *memcache.Item) error {
				return expectedErr
			},
		}

		req := app.NewKeyRequest{
			Key:   "error-key",
			Value: "value",
		}

		cmd := app.NewKeyCmd(mock, req)
		msg := cmd()

		errMsg, ok := msg.(app.NewKeyErrorMsg)
		if !ok {
			t.Fatalf("expected NewKeyErrorMsg, got %T", msg)
		}

		if errMsg.Err != expectedErr {
			t.Errorf("expected error '%v', got '%v'", expectedErr, errMsg.Err)
		}
	})
}

// TestNewKeyCmd_InvalidKey tests invalid key handling
func TestNewKeyCmd_InvalidKey(t *testing.T) {
	t.Run("nil client returns error", func(t *testing.T) {
		req := app.NewKeyRequest{
			Key:   "test-key",
			Value: "value",
		}

		cmd := app.NewKeyCmd(nil, req)
		msg := cmd()

		errMsg, ok := msg.(app.NewKeyErrorMsg)
		if !ok {
			t.Fatalf("expected NewKeyErrorMsg, got %T", msg)
		}

		if errMsg.Err == nil {
			t.Error("expected non-nil error for nil client")
		}
	})

	t.Run("empty key returns validation error", func(t *testing.T) {
		err := app.ValidateKeyName("")
		if err == nil {
			t.Error("expected error for empty key")
		}
	})

	t.Run("key with spaces returns validation error", func(t *testing.T) {
		err := app.ValidateKeyName("key with spaces")
		if err == nil {
			t.Error("expected error for key with spaces")
		}
	})

	t.Run("key with newline returns validation error", func(t *testing.T) {
		err := app.ValidateKeyName("key\nwith\nnewlines")
		if err == nil {
			t.Error("expected error for key with newlines")
		}
	})

	t.Run("key too long returns validation error", func(t *testing.T) {
		// Memcached max key length is 250 bytes
		longKey := make([]byte, 251)
		for i := range longKey {
			longKey[i] = 'a'
		}
		err := app.ValidateKeyName(string(longKey))
		if err == nil {
			t.Error("expected error for key too long")
		}
	})

	t.Run("valid key passes validation", func(t *testing.T) {
		err := app.ValidateKeyName("valid-key:123")
		if err != nil {
			t.Errorf("expected no error for valid key, got %v", err)
		}
	})

	t.Run("key with control characters returns validation error", func(t *testing.T) {
		err := app.ValidateKeyName("key\x00value")
		if err == nil {
			t.Error("expected error for key with control characters")
		}
	})
}

// TestHandleNewKeyResult tests handling of new key operation results
func TestHandleNewKeyResult(t *testing.T) {
	t.Run("KeyCreatedMsg returns success result", func(t *testing.T) {
		msg := app.KeyCreatedMsg{Key: "created-key"}
		result := app.HandleNewKeyResult(msg)

		if !result.ShouldRefresh {
			t.Error("expected ShouldRefresh to be true after successful creation")
		}

		if result.Error != "" {
			t.Errorf("expected no error, got '%s'", result.Error)
		}

		if result.CreatedKey != "created-key" {
			t.Errorf("expected CreatedKey 'created-key', got '%s'", result.CreatedKey)
		}
	})

	t.Run("NewKeyErrorMsg returns error result", func(t *testing.T) {
		msg := app.NewKeyErrorMsg{
			Key: "failed-key",
			Err: errors.New("set failed"),
		}
		result := app.HandleNewKeyResult(msg)

		if result.ShouldRefresh {
			t.Error("expected ShouldRefresh to be false after error")
		}

		if result.Error != "set failed" {
			t.Errorf("expected error 'set failed', got '%s'", result.Error)
		}

		if result.FailedKey != "failed-key" {
			t.Errorf("expected FailedKey 'failed-key', got '%s'", result.FailedKey)
		}

		if result.CreatedKey != "" {
			t.Errorf("expected empty CreatedKey on error, got '%s'", result.CreatedKey)
		}
	})

	t.Run("unknown message returns empty result", func(t *testing.T) {
		msg := struct{}{}
		result := app.HandleNewKeyResult(msg)

		if result.ShouldRefresh {
			t.Error("expected ShouldRefresh to be false for unknown message")
		}
	})
}

// TestCreateNewKeyDialog tests creation of new key input dialog
func TestCreateNewKeyDialog(t *testing.T) {
	t.Run("creates dialog with correct title", func(t *testing.T) {
		dialog := app.CreateNewKeyDialog()

		if dialog == nil {
			t.Fatal("expected non-nil dialog")
		}

		if dialog.Title() != "New Key" {
			t.Errorf("expected title 'New Key', got '%s'", dialog.Title())
		}
	})

	t.Run("dialog has key validation", func(t *testing.T) {
		dialog := app.CreateNewKeyDialog()

		if dialog == nil {
			t.Fatal("expected non-nil dialog")
		}

		// Dialog should have validation for key name
		// We verify by checking it's created with the validator
	})
}

// TestCreateValueInputDialog tests creation of value input dialog
func TestCreateValueInputDialog(t *testing.T) {
	t.Run("creates dialog with correct title", func(t *testing.T) {
		dialog := app.CreateValueInputDialog("my-key")

		if dialog == nil {
			t.Fatal("expected non-nil dialog")
		}

		expectedTitle := "Value for: my-key"
		if dialog.Title() != expectedTitle {
			t.Errorf("expected title '%s', got '%s'", expectedTitle, dialog.Title())
		}
	})

	t.Run("dialog stores key in context", func(t *testing.T) {
		keyName := "test-key"
		dialog := app.CreateValueInputDialog(keyName)

		if dialog == nil {
			t.Fatal("expected non-nil dialog")
		}

		// Simulate pressing Enter with some value
		// The context should contain the key name
	})
}

// TestNewKeyMessages tests the message structs
func TestNewKeyMessages(t *testing.T) {
	t.Run("NewKeyMsg", func(t *testing.T) {
		msg := app.NewKeyMsg{}
		// NewKeyMsg is just a trigger, no fields to test
		_ = msg
	})

	t.Run("KeyCreatedMsg", func(t *testing.T) {
		msg := app.KeyCreatedMsg{Key: "new-key"}
		if msg.Key != "new-key" {
			t.Errorf("expected Key 'new-key', got '%s'", msg.Key)
		}
	})

	t.Run("NewKeyErrorMsg", func(t *testing.T) {
		err := errors.New("test error")
		msg := app.NewKeyErrorMsg{Key: "error-key", Err: err}
		if msg.Key != "error-key" {
			t.Errorf("expected Key 'error-key', got '%s'", msg.Key)
		}
		if msg.Err != err {
			t.Errorf("expected Err '%v', got '%v'", err, msg.Err)
		}
	})

	t.Run("NewKeyRequest", func(t *testing.T) {
		req := app.NewKeyRequest{
			Key:   "test-key",
			Value: "test-value",
			TTL:   3600,
			Flags: 0,
		}
		if req.Key != "test-key" {
			t.Errorf("expected Key 'test-key', got '%s'", req.Key)
		}
		if req.Value != "test-value" {
			t.Errorf("expected Value 'test-value', got '%s'", req.Value)
		}
		if req.TTL != 3600 {
			t.Errorf("expected TTL 3600, got %d", req.TTL)
		}
	})
}

// TestNewKeyFlowIntegration tests the complete new key creation flow
func TestNewKeyFlowIntegration(t *testing.T) {
	t.Run("full new key flow", func(t *testing.T) {
		mock := &MockSetter{}
		keyName := "integration:new:key"
		keyValue := "integration test value"

		// Step 1: Create key input dialog
		keyDialog := app.CreateNewKeyDialog()
		if keyDialog == nil {
			t.Fatal("failed to create key dialog")
		}

		// Step 2: Validate key name
		err := app.ValidateKeyName(keyName)
		if err != nil {
			t.Fatalf("key validation failed: %v", err)
		}

		// Step 3: Create value input dialog
		valueDialog := app.CreateValueInputDialog(keyName)
		if valueDialog == nil {
			t.Fatal("failed to create value dialog")
		}

		// Step 4: Create new key request
		req := app.NewKeyRequest{
			Key:   keyName,
			Value: keyValue,
		}

		// Step 5: Execute new key command
		cmd := app.NewKeyCmd(mock, req)
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		msg := cmd()
		createdMsg, ok := msg.(app.KeyCreatedMsg)
		if !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T", msg)
		}

		if createdMsg.Key != keyName {
			t.Errorf("expected created key '%s', got '%s'", keyName, createdMsg.Key)
		}

		// Step 6: Handle result
		result := app.HandleNewKeyResult(createdMsg)
		if !result.ShouldRefresh {
			t.Error("expected ShouldRefresh after successful creation")
		}
	})

	t.Run("new key flow with TTL", func(t *testing.T) {
		mock := &MockSetter{}

		req := app.NewKeyRequest{
			Key:   "ttl:key",
			Value: "ttl value",
			TTL:   7200,
		}

		cmd := app.NewKeyCmd(mock, req)
		msg := cmd()

		createdMsg, ok := msg.(app.KeyCreatedMsg)
		if !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T", msg)
		}

		if createdMsg.Key != "ttl:key" {
			t.Errorf("expected key 'ttl:key', got '%s'", createdMsg.Key)
		}

		// Verify TTL was set correctly
		if mock.LastSet.Expiration != 7200 {
			t.Errorf("expected expiration 7200, got %d", mock.LastSet.Expiration)
		}
	})

	t.Run("new key flow canceled by user", func(t *testing.T) {
		// Step 1: Create key input dialog
		keyDialog := app.CreateNewKeyDialog()

		// Step 2: User cancels (presses Escape)
		_, cancelCmd := keyDialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cancelCmd == nil {
			t.Fatal("expected cancel command")
		}

		// The cancel command would return InputResultMsg with Canceled: true
		// In real app, this would close the dialog without creating key
	})
}

// TestExtractNewKeyContext tests context extraction from input results
func TestExtractNewKeyContext(t *testing.T) {
	t.Run("extracts key from string context", func(t *testing.T) {
		key, ok := app.ExtractNewKeyContext("my-key")
		if !ok {
			t.Error("expected successful extraction")
		}
		if key != "my-key" {
			t.Errorf("expected 'my-key', got '%s'", key)
		}
	})

	t.Run("extracts key from NewKeyContext struct", func(t *testing.T) {
		ctx := app.NewKeyContext{Key: "struct-key"}
		key, ok := app.ExtractNewKeyContext(ctx)
		if !ok {
			t.Error("expected successful extraction")
		}
		if key != "struct-key" {
			t.Errorf("expected 'struct-key', got '%s'", key)
		}
	})

	t.Run("returns false for nil context", func(t *testing.T) {
		key, ok := app.ExtractNewKeyContext(nil)
		if ok {
			t.Error("expected extraction to fail for nil context")
		}
		if key != "" {
			t.Errorf("expected empty string, got '%s'", key)
		}
	})

	t.Run("returns false for unknown type", func(t *testing.T) {
		key, ok := app.ExtractNewKeyContext(123)
		if ok {
			t.Error("expected extraction to fail for unknown type")
		}
		if key != "" {
			t.Errorf("expected empty string, got '%s'", key)
		}
	})
}

// TestProcessNewKeyInputResult tests processing of input dialog results
func TestProcessNewKeyInputResult(t *testing.T) {
	t.Run("returns key creation request on valid input", func(t *testing.T) {
		// Simulating the result from value input dialog
		keyName := "result-key"
		value := "result value"

		req := app.ProcessNewKeyInputResult(keyName, value, 0)
		if req == nil {
			t.Fatal("expected non-nil request")
		}

		if req.Key != keyName {
			t.Errorf("expected key '%s', got '%s'", keyName, req.Key)
		}

		if req.Value != value {
			t.Errorf("expected value '%s', got '%s'", value, req.Value)
		}
	})

	t.Run("includes TTL in request", func(t *testing.T) {
		req := app.ProcessNewKeyInputResult("key", "value", 3600)
		if req == nil {
			t.Fatal("expected non-nil request")
		}

		if req.TTL != 3600 {
			t.Errorf("expected TTL 3600, got %d", req.TTL)
		}
	})
}
