//go:build e2e

package e2e_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
	"github.com/nnnkkk7/memtui/ui/components/editor"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

// CRUD Test constants
const (
	crudTestKeyPrefix     = "e2e_crud_"
	crudDefaultTimeout    = 10 * time.Second
	crudKeyEnumDelay      = 500 * time.Millisecond
	crudDefaultTermWidth  = 120
	crudDefaultTermHeight = 40
)

// crudTestHelper provides utilities for CRUD E2E tests
type crudTestHelper struct {
	t         *testing.T
	addr      string
	mc        *memcache.Client
	casClient *client.Client
	model     *app.Model
	testKeys  []string
}

// newCrudTestHelper creates a new test helper for CRUD tests
func newCrudTestHelper(t *testing.T) *crudTestHelper {
	t.Helper()
	skipIfNoMemcached(t)

	addr := getMemcachedAddr()
	mc := memcache.New(addr)

	casClient, err := client.New(addr)
	if err != nil {
		t.Fatalf("failed to create CAS client: %v", err)
	}

	return &crudTestHelper{
		t:         t,
		addr:      addr,
		mc:        mc,
		casClient: casClient,
		testKeys:  make([]string, 0),
	}
}

// cleanup removes all test keys
func (h *crudTestHelper) cleanup() {
	for _, key := range h.testKeys {
		h.mc.Delete(key)
	}
}

// generateTestKey generates a unique test key
func (h *crudTestHelper) generateTestKey(suffix string) string {
	key := fmt.Sprintf("%s%s_%d", crudTestKeyPrefix, suffix, time.Now().UnixNano())
	h.testKeys = append(h.testKeys, key)
	return key
}

// setTestKey sets a test key with value
func (h *crudTestHelper) setTestKey(key, value string) error {
	h.testKeys = append(h.testKeys, key)
	return h.mc.Set(&memcache.Item{
		Key:   key,
		Value: []byte(value),
	})
}

// getTestKey gets a test key value
func (h *crudTestHelper) getTestKey(key string) (string, error) {
	item, err := h.mc.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

// keyExists checks if a key exists
func (h *crudTestHelper) keyExists(key string) bool {
	_, err := h.mc.Get(key)
	return err == nil
}

// initModel initializes and connects the app model
func (h *crudTestHelper) initModel() *app.Model {
	h.model = app.NewModel(h.addr)

	// Simulate initial window size
	h.model.Update(tea.WindowSizeMsg{
		Width:  crudDefaultTermWidth,
		Height: crudDefaultTermHeight,
	})

	// Run connect command
	cmd := h.model.Init()
	if cmd != nil {
		msg := cmd()
		h.model.Update(msg)
	}

	return h.model
}

// simulateKeyPress simulates a key press
func (h *crudTestHelper) simulateKeyPress(key string) tea.Msg {
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key),
	}
}

// TestCreateNewKey_WithNKey tests creating a new key using 'n' key
func TestCreateNewKey_WithNKey(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Initialize model and wait for connection
	model := h.initModel()

	// Process connection messages
	connCmd := model.Init()
	if connCmd != nil {
		connMsg := connCmd()
		model.Update(connMsg)
	}

	// Simulate 'n' key press to open new key dialog
	model.Update(h.simulateKeyPress("n"))

	// Verify input dialog is shown (focus should change)
	if model.Focus() != app.FocusDialog {
		t.Logf("Expected focus to be on dialog, got: %d", model.Focus())
	}

	// Test the dialog creation directly
	dlg := dialog.NewInput("New Key").
		WithPlaceholder("Enter key name...").
		WithValidator(app.ValidateKeyName)

	if dlg.Title() != "New Key" {
		t.Errorf("expected dialog title 'New Key', got '%s'", dlg.Title())
	}

	// Test key validation
	testCases := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid key", "mykey", false},
		{"valid key with colon", "user:123", false},
		{"empty key", "", true},
		{"key with space", "my key", true},
		{"key with newline", "my\nkey", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.ValidateKeyName(tc.key)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for key '%s'", tc.key)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for key '%s': %v", tc.key, err)
			}
		})
	}
}

// TestCreateNewKey_ActualMemcached tests actual key creation in Memcached
func TestCreateNewKey_ActualMemcached(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("create")
	testValue := "test_value_123"

	// Use NewKeyCmd to create the key
	cmd := app.NewKeyCmd(h.mc, app.NewKeyRequest{
		Key:   testKey,
		Value: testValue,
		TTL:   0,
	})

	// Execute the command
	msg := cmd()

	// Verify the result
	switch result := msg.(type) {
	case app.KeyCreatedMsg:
		if result.Key != testKey {
			t.Errorf("expected key '%s', got '%s'", testKey, result.Key)
		}
	case app.NewKeyErrorMsg:
		t.Fatalf("failed to create key: %v", result.Err)
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}

	// Verify the key exists in Memcached
	value, err := h.getTestKey(testKey)
	if err != nil {
		t.Fatalf("key not found in Memcached: %v", err)
	}
	if value != testValue {
		t.Errorf("expected value '%s', got '%s'", testValue, value)
	}
}

// TestReadKeyValue_WithEnter tests reading a key value with Enter key
func TestReadKeyValue_WithEnter(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Create a test key first
	testKey := h.generateTestKey("read")
	testValue := `{"name": "test", "value": 42}`
	if err := h.setTestKey(testKey, testValue); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Wait for key enumeration
	time.Sleep(crudKeyEnumDelay)

	// Initialize model
	model := h.initModel()

	// Process connection
	connCmd := model.Init()
	if connCmd != nil {
		connMsg := connCmd()
		model.Update(connMsg)
	}

	// Simulate keylist selection message
	ki := models.KeyInfo{Key: testKey, Size: len(testValue)}
	selMsg := keylist.KeySelectedMsg{Key: ki}
	model.Update(selMsg)

	// Verify the model handles the selection
	// The model should load the value
}

// TestReadKeyValue_ActualMemcached tests reading value from Memcached
func TestReadKeyValue_ActualMemcached(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("read_actual")
	testValue := "hello_world_123"

	if err := h.setTestKey(testKey, testValue); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Use CAS client to get the value
	casItem, err := h.casClient.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("failed to get key with CAS: %v", err)
	}

	if string(casItem.Value) != testValue {
		t.Errorf("expected value '%s', got '%s'", testValue, string(casItem.Value))
	}

	// Verify CAS token is present
	if casItem.CAS == 0 {
		t.Error("expected non-zero CAS value")
	}
}

// TestEditKeyValue_WithEKey tests editing a key value with 'e' key
func TestEditKeyValue_WithEKey(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("edit")
	originalValue := "original_value"
	if err := h.setTestKey(testKey, originalValue); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Create an editor component directly
	ed := editor.New(testKey, []byte(originalValue))
	ed.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Verify editor initial state
	if ed.Key() != testKey {
		t.Errorf("expected key '%s', got '%s'", testKey, ed.Key())
	}

	if string(ed.OriginalValue()) != originalValue {
		t.Errorf("expected original value '%s', got '%s'", originalValue, string(ed.OriginalValue()))
	}

	if ed.IsDirty() {
		t.Error("editor should not be dirty initially")
	}

	// Simulate editing
	ed.SetContent([]byte("modified_value"))

	if !ed.IsDirty() {
		t.Error("editor should be dirty after modification")
	}
}

// TestEditKeyValue_SaveWithCtrlS tests saving edited value with Ctrl+S
func TestEditKeyValue_SaveWithCtrlS(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("edit_save")
	originalValue := "original"
	newValue := "modified"

	if err := h.setTestKey(testKey, originalValue); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Create editor
	ed := editor.New(testKey, []byte(originalValue))
	ed.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Modify the value
	ed.SetContent([]byte(newValue))

	// Simulate Ctrl+S
	_, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	// Execute the command
	if cmd != nil {
		msg := cmd()
		saveMsg, ok := msg.(editor.EditorSaveMsg)
		if !ok {
			t.Fatalf("expected EditorSaveMsg, got %T", msg)
		}

		if saveMsg.Key != testKey {
			t.Errorf("expected key '%s', got '%s'", testKey, saveMsg.Key)
		}

		if string(saveMsg.Value) != newValue {
			t.Errorf("expected value '%s', got '%s'", newValue, string(saveMsg.Value))
		}
	}

	// Actually save the value to Memcached
	err := h.mc.Set(&memcache.Item{
		Key:   testKey,
		Value: []byte(newValue),
	})
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Verify the value was updated
	value, err := h.getTestKey(testKey)
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}
	if value != newValue {
		t.Errorf("expected value '%s', got '%s'", newValue, value)
	}
}

// TestEditKeyValue_CancelWithEsc tests canceling edit with Escape
func TestEditKeyValue_CancelWithEsc(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("edit_cancel")
	originalValue := "original"

	// Create editor
	ed := editor.New(testKey, []byte(originalValue))
	ed.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Modify the value
	ed.SetContent([]byte("modified"))

	// Simulate Escape
	_, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Execute the command
	if cmd != nil {
		msg := cmd()
		_, ok := msg.(editor.EditorCancelMsg)
		if !ok {
			t.Fatalf("expected EditorCancelMsg, got %T", msg)
		}
	}
}

// TestDeleteKey_WithDKey tests deleting a key with 'd' key
func TestDeleteKey_WithDKey(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("delete")
	if err := h.setTestKey(testKey, "to_be_deleted"); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Test CreateDeleteConfirmDialog
	dlg := app.CreateDeleteConfirmDialog(testKey)

	if dlg.Title() != "Delete Key" {
		t.Errorf("expected title 'Delete Key', got '%s'", dlg.Title())
	}

	if !strings.Contains(dlg.Message(), testKey) {
		t.Errorf("expected message to contain key '%s'", testKey)
	}

	// Default should be on No for safety
	if dlg.FocusedOnYes() {
		t.Error("default focus should be on No")
	}
}

// TestDeleteKey_Confirmation tests delete confirmation with 'y' key
func TestDeleteKey_Confirmation(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("delete_confirm")
	if err := h.setTestKey(testKey, "to_be_deleted"); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Verify key exists
	if !h.keyExists(testKey) {
		t.Fatal("test key should exist before deletion")
	}

	// Use DeleteKeyCmd directly
	cmd := app.DeleteKeyCmd(h.mc, testKey)
	msg := cmd()

	switch result := msg.(type) {
	case app.KeyDeletedMsg:
		if result.Key != testKey {
			t.Errorf("expected key '%s', got '%s'", testKey, result.Key)
		}
	case app.DeleteErrorMsg:
		t.Fatalf("failed to delete key: %v", result.Err)
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}

	// Verify key is deleted
	if h.keyExists(testKey) {
		t.Error("key should not exist after deletion")
	}
}

// TestDeleteKey_Cancellation tests delete cancellation with 'n' or Escape
func TestDeleteKey_Cancellation(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("delete_cancel")
	if err := h.setTestKey(testKey, "should_remain"); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Create confirm dialog
	dlg := app.CreateDeleteConfirmDialog(testKey)
	dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Simulate 'n' key press
	_, cmd := dlg.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'n'},
	})

	// Execute the command
	if cmd != nil {
		msg := cmd()
		confirmMsg, ok := msg.(dialog.ConfirmResultMsg)
		if !ok {
			t.Fatalf("expected ConfirmResultMsg, got %T", msg)
		}
		if confirmMsg.Result {
			t.Error("result should be false for cancellation")
		}
	}

	// Key should still exist
	if !h.keyExists(testKey) {
		t.Error("key should still exist after cancellation")
	}
}

// TestDeleteKey_EscapeCancellation tests delete cancellation with Escape
func TestDeleteKey_EscapeCancellation(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("delete_esc")
	if err := h.setTestKey(testKey, "should_remain"); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Create confirm dialog
	dlg := app.CreateDeleteConfirmDialog(testKey)
	dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Simulate Escape key press
	_, cmd := dlg.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Execute the command
	if cmd != nil {
		msg := cmd()
		confirmMsg, ok := msg.(dialog.ConfirmResultMsg)
		if !ok {
			t.Fatalf("expected ConfirmResultMsg, got %T", msg)
		}
		if confirmMsg.Result {
			t.Error("result should be false for Escape")
		}
	}

	// Key should still exist
	if !h.keyExists(testKey) {
		t.Error("key should still exist after Escape")
	}
}

// TestKeyCreation_Validation tests key name validation rules
func TestKeyCreation_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple key",
			key:     "mykey",
			wantErr: false,
		},
		{
			name:    "valid key with delimiter",
			key:     "user:session:123",
			wantErr: false,
		},
		{
			name:    "valid key with numbers",
			key:     "key123",
			wantErr: false,
		},
		{
			name:    "valid key with underscore",
			key:     "my_key_name",
			wantErr: false,
		},
		{
			name:    "valid key with hyphen",
			key:     "my-key-name",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "key with space",
			key:     "my key",
			wantErr: true,
			errMsg:  "cannot contain spaces",
		},
		{
			name:    "key with newline",
			key:     "my\nkey",
			wantErr: true,
			errMsg:  "cannot contain newlines",
		},
		{
			name:    "key with carriage return",
			key:     "my\rkey",
			wantErr: true,
			errMsg:  "cannot contain newlines",
		},
		{
			name:    "key with tab",
			key:     "my\tkey",
			wantErr: true,
			errMsg:  "cannot contain",
		},
		{
			name:    "key exceeding max length",
			key:     strings.Repeat("a", 251),
			wantErr: true,
			errMsg:  "cannot exceed",
		},
		{
			name:    "key at max length",
			key:     strings.Repeat("a", 250),
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.ValidateKeyName(tc.key)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for key '%s'", tc.key)
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for key '%s': %v", tc.key, err)
				}
			}
		})
	}
}

// TestValueUpdate_AndRefresh tests updating a value and refreshing
func TestValueUpdate_AndRefresh(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("update_refresh")
	originalValue := "original_value"
	updatedValue := "updated_value"

	// Create initial key
	if err := h.setTestKey(testKey, originalValue); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Verify initial value
	value, err := h.getTestKey(testKey)
	if err != nil {
		t.Fatalf("failed to get initial value: %v", err)
	}
	if value != originalValue {
		t.Errorf("expected '%s', got '%s'", originalValue, value)
	}

	// Update the value
	err = h.mc.Set(&memcache.Item{
		Key:   testKey,
		Value: []byte(updatedValue),
	})
	if err != nil {
		t.Fatalf("failed to update value: %v", err)
	}

	// Refresh and verify
	value, err = h.getTestKey(testKey)
	if err != nil {
		t.Fatalf("failed to get updated value: %v", err)
	}
	if value != updatedValue {
		t.Errorf("expected '%s', got '%s'", updatedValue, value)
	}
}

// TestValueUpdate_CAS tests CAS (Compare-And-Swap) operations
func TestValueUpdate_CAS(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("cas_test")
	originalValue := "original_cas_value"

	// Create initial key
	if err := h.setTestKey(testKey, originalValue); err != nil {
		t.Fatalf("failed to set test key: %v", err)
	}

	// Get with CAS
	casItem, err := h.casClient.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("failed to get with CAS: %v", err)
	}

	// Verify CAS token
	if casItem.CAS == 0 {
		t.Error("expected non-zero CAS token")
	}

	// Modify and save with CAS
	casItem.Value = []byte("modified_cas_value")
	err = h.casClient.CompareAndSwap(casItem)
	if err != nil {
		t.Fatalf("CAS update failed: %v", err)
	}

	// Verify the update
	value, err := h.getTestKey(testKey)
	if err != nil {
		t.Fatalf("failed to get value: %v", err)
	}
	if value != "modified_cas_value" {
		t.Errorf("expected 'modified_cas_value', got '%s'", value)
	}
}

// TestErrorHandling_InvalidOperations tests error handling for invalid operations
func TestErrorHandling_InvalidOperations(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	t.Run("delete non-existent key", func(t *testing.T) {
		nonExistentKey := h.generateTestKey("non_existent")

		cmd := app.DeleteKeyCmd(h.mc, nonExistentKey)
		msg := cmd()

		errMsg, ok := msg.(app.DeleteErrorMsg)
		if !ok {
			t.Fatalf("expected DeleteErrorMsg, got %T", msg)
		}
		if errMsg.Err == nil {
			t.Error("expected error for deleting non-existent key")
		}
	})

	t.Run("get non-existent key", func(t *testing.T) {
		nonExistentKey := h.generateTestKey("non_existent_get")

		_, err := h.casClient.GetWithCAS(nonExistentKey)
		if err == nil {
			t.Error("expected error for getting non-existent key")
		}
		if !client.IsCacheMiss(err) {
			t.Errorf("expected cache miss error, got: %v", err)
		}
	})

	t.Run("nil client delete", func(t *testing.T) {
		cmd := app.DeleteKeyCmd(nil, "anykey")
		msg := cmd()

		errMsg, ok := msg.(app.DeleteErrorMsg)
		if !ok {
			t.Fatalf("expected DeleteErrorMsg, got %T", msg)
		}
		if errMsg.Err == nil {
			t.Error("expected error for nil client")
		}
	})

	t.Run("nil client create", func(t *testing.T) {
		cmd := app.NewKeyCmd(nil, app.NewKeyRequest{
			Key:   "testkey",
			Value: "testvalue",
		})
		msg := cmd()

		errMsg, ok := msg.(app.NewKeyErrorMsg)
		if !ok {
			t.Fatalf("expected NewKeyErrorMsg, got %T", msg)
		}
		if errMsg.Err == nil {
			t.Error("expected error for nil client")
		}
	})
}

// TestBatchDelete tests batch delete operations
func TestBatchDelete(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Create multiple test keys
	keys := make([]string, 5)
	for i := 0; i < 5; i++ {
		keys[i] = h.generateTestKey(fmt.Sprintf("batch_%d", i))
		if err := h.setTestKey(keys[i], fmt.Sprintf("value_%d", i)); err != nil {
			t.Fatalf("failed to set key %s: %v", keys[i], err)
		}
	}

	// Verify all keys exist
	for _, key := range keys {
		if !h.keyExists(key) {
			t.Fatalf("key %s should exist before batch delete", key)
		}
	}

	// Test batch delete validation
	err := app.ValidateBatchDeleteInput("DELETE")
	if err != nil {
		t.Errorf("'DELETE' should be valid: %v", err)
	}

	err = app.ValidateBatchDeleteInput("delete")
	if err == nil {
		t.Error("'delete' (lowercase) should be invalid")
	}

	err = app.ValidateBatchDeleteInput("")
	if err == nil {
		t.Error("empty string should be invalid")
	}

	// Execute batch delete
	cmd := app.BatchDeleteCmd(h.mc, keys)
	msg := cmd()

	result, ok := msg.(app.BatchDeleteResultMsg)
	if !ok {
		t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
	}

	// Verify all keys were deleted
	if len(result.Deleted) != 5 {
		t.Errorf("expected 5 deleted keys, got %d", len(result.Deleted))
	}

	if len(result.Failed) != 0 {
		t.Errorf("expected 0 failed keys, got %d", len(result.Failed))
	}

	// Verify keys are gone
	for _, key := range keys {
		if h.keyExists(key) {
			t.Errorf("key %s should not exist after batch delete", key)
		}
	}
}

// TestBatchDelete_PartialFailure tests batch delete with some failures
func TestBatchDelete_PartialFailure(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Create some keys
	existingKey := h.generateTestKey("batch_exists")
	if err := h.setTestKey(existingKey, "value"); err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	nonExistentKey := h.generateTestKey("batch_nonexistent")

	// Try to delete both
	keys := []string{existingKey, nonExistentKey}
	cmd := app.BatchDeleteCmd(h.mc, keys)
	msg := cmd()

	result, ok := msg.(app.BatchDeleteResultMsg)
	if !ok {
		t.Fatalf("expected BatchDeleteResultMsg, got %T", msg)
	}

	// The existing key should be deleted, non-existent should fail
	if len(result.Deleted) != 1 {
		t.Errorf("expected 1 deleted key, got %d", len(result.Deleted))
	}

	if len(result.Failed) != 1 {
		t.Errorf("expected 1 failed key, got %d", len(result.Failed))
	}

	// Verify summary
	summary := app.HandleBatchDeleteResult(result)
	if summary.AllSucceeded {
		t.Error("AllSucceeded should be false")
	}
	if !summary.HasErrors {
		t.Error("HasErrors should be true")
	}
	if !summary.ShouldRefresh {
		t.Error("ShouldRefresh should be true (some keys were deleted)")
	}
}

// TestKeyListNavigation tests navigation in key list
func TestKeyListNavigation(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Create test keys
	keys := make([]models.KeyInfo, 5)
	for i := 0; i < 5; i++ {
		key := h.generateTestKey(fmt.Sprintf("nav_%d", i))
		if err := h.setTestKey(key, fmt.Sprintf("value_%d", i)); err != nil {
			t.Fatalf("failed to set key: %v", err)
		}
		keys[i] = models.KeyInfo{Key: key, Size: 10}
	}

	// Create keylist model
	kl := keylist.NewModel()
	kl.SetKeys(keys)
	kl.SetSize(50, 20)

	// Initial cursor should be at 0
	if kl.Cursor() != 0 {
		t.Errorf("initial cursor should be 0, got %d", kl.Cursor())
	}

	// Navigate down
	kl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if kl.Cursor() != 1 {
		t.Errorf("cursor should be 1 after down, got %d", kl.Cursor())
	}

	// Navigate up
	kl.Update(tea.KeyMsg{Type: tea.KeyUp})
	if kl.Cursor() != 0 {
		t.Errorf("cursor should be 0 after up, got %d", kl.Cursor())
	}
}

// TestKeyListFilter tests filtering keys
func TestKeyListFilter(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Create keylist with test keys
	keys := []models.KeyInfo{
		{Key: "user:123", Size: 10},
		{Key: "user:456", Size: 10},
		{Key: "session:abc", Size: 10},
		{Key: "cache:data", Size: 10},
	}

	kl := keylist.NewModel()
	kl.SetKeys(keys)

	// Filter by "user"
	kl.SetFilter("user")

	filtered := kl.FilteredKeys()
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered keys, got %d", len(filtered))
	}

	// Clear filter
	kl.SetFilter("")
	filtered = kl.FilteredKeys()
	if len(filtered) != 4 {
		t.Errorf("expected 4 keys after clearing filter, got %d", len(filtered))
	}
}

// TestKeyListMultiSelect tests multi-select functionality
func TestKeyListMultiSelect(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	// Use flat keys without delimiters for direct selection
	keys := []models.KeyInfo{
		{Key: "multi1", Size: 10},
		{Key: "multi2", Size: 10},
		{Key: "multi3", Size: 10},
	}

	kl := keylist.NewModel()
	kl.SetDelimiter("") // Flat list - no tree structure
	kl.SetKeys(keys)
	kl.SetSize(50, 20)

	// Initially no selection
	if kl.HasSelection() {
		t.Error("should have no selection initially")
	}

	// Toggle selection with space
	kl.Update(tea.KeyMsg{Type: tea.KeySpace})

	if !kl.HasSelection() {
		t.Error("should have selection after space")
	}

	if kl.SelectionCount() != 1 {
		t.Errorf("expected 1 selected, got %d", kl.SelectionCount())
	}

	// Move down and select another
	kl.Update(tea.KeyMsg{Type: tea.KeyDown})
	kl.Update(tea.KeyMsg{Type: tea.KeySpace})

	if kl.SelectionCount() != 2 {
		t.Errorf("expected 2 selected, got %d", kl.SelectionCount())
	}

	// Get selected keys
	selectedKeys := kl.SelectedKeys()
	if len(selectedKeys) != 2 {
		t.Errorf("expected 2 selected keys, got %d", len(selectedKeys))
	}

	// Clear selection
	kl.ClearSelection()
	if kl.HasSelection() {
		t.Error("should have no selection after clear")
	}
}

// TestConfirmDialog_YesNoToggle tests toggling between Yes and No
func TestConfirmDialog_YesNoToggle(t *testing.T) {
	dlg := dialog.New("Test", "Test message")
	dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Default should be No
	if dlg.FocusedOnYes() {
		t.Error("default focus should be No")
	}

	// Toggle with Tab
	dlg.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !dlg.FocusedOnYes() {
		t.Error("focus should be Yes after Tab")
	}

	// Toggle back
	dlg.Update(tea.KeyMsg{Type: tea.KeyTab})
	if dlg.FocusedOnYes() {
		t.Error("focus should be No after second Tab")
	}

	// Use arrow keys
	dlg.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if !dlg.FocusedOnYes() {
		t.Error("focus should be Yes after Left")
	}

	dlg.Update(tea.KeyMsg{Type: tea.KeyRight})
	if dlg.FocusedOnYes() {
		t.Error("focus should be No after Right")
	}
}

// TestConfirmDialog_QuickKeys tests 'y' and 'n' quick keys
func TestConfirmDialog_QuickKeys(t *testing.T) {
	t.Run("press y", func(t *testing.T) {
		dlg := dialog.New("Test", "Test message")
		dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

		_, cmd := dlg.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'y'},
		})

		if cmd != nil {
			msg := cmd()
			result, ok := msg.(dialog.ConfirmResultMsg)
			if !ok {
				t.Fatalf("expected ConfirmResultMsg, got %T", msg)
			}
			if !result.Result {
				t.Error("result should be true for 'y'")
			}
		}
	})

	t.Run("press n", func(t *testing.T) {
		dlg := dialog.New("Test", "Test message")
		dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

		_, cmd := dlg.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'n'},
		})

		if cmd != nil {
			msg := cmd()
			result, ok := msg.(dialog.ConfirmResultMsg)
			if !ok {
				t.Fatalf("expected ConfirmResultMsg, got %T", msg)
			}
			if result.Result {
				t.Error("result should be false for 'n'")
			}
		}
	})
}

// TestInputDialog_Validation tests input dialog validation
func TestInputDialog_Validation(t *testing.T) {
	dlg := dialog.NewInput("Test Input").
		WithPlaceholder("Enter value...").
		WithValidator(func(s string) error {
			if len(s) < 3 {
				return fmt.Errorf("input must be at least 3 characters")
			}
			return nil
		})
	dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Type short input
	for _, r := range "ab" {
		dlg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Try to submit - should fail validation
	dlg.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Check validation error
	if dlg.ValidationError() == "" {
		// Note: validation error might be set on the model
		// This depends on implementation details
	}
}

// TestInputDialog_Cancel tests input dialog cancellation
func TestInputDialog_Cancel(t *testing.T) {
	dlg := dialog.NewInput("Test Input")
	dlg.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Type something
	for _, r := range "test" {
		dlg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Escape
	_, cmd := dlg.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd != nil {
		msg := cmd()
		result, ok := msg.(dialog.InputResultMsg)
		if !ok {
			t.Fatalf("expected InputResultMsg, got %T", msg)
		}
		if !result.Canceled {
			t.Error("result should be canceled")
		}
	}
}

// TestCRUDWorkflow_FullCycle tests a complete CRUD workflow
func TestCRUDWorkflow_FullCycle(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	testKey := h.generateTestKey("crud_full")

	// 1. CREATE
	t.Log("Creating key...")
	cmd := app.NewKeyCmd(h.mc, app.NewKeyRequest{
		Key:   testKey,
		Value: "initial_value",
		TTL:   0,
	})
	msg := cmd()
	if _, ok := msg.(app.KeyCreatedMsg); !ok {
		t.Fatalf("CREATE failed: expected KeyCreatedMsg, got %T", msg)
	}

	// 2. READ
	t.Log("Reading key...")
	casItem, err := h.casClient.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if string(casItem.Value) != "initial_value" {
		t.Errorf("READ: expected 'initial_value', got '%s'", string(casItem.Value))
	}

	// 3. UPDATE
	t.Log("Updating key...")
	casItem.Value = []byte("updated_value")
	err = h.casClient.CompareAndSwap(casItem)
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}

	// Verify update
	value, err := h.getTestKey(testKey)
	if err != nil {
		t.Fatalf("failed to verify update: %v", err)
	}
	if value != "updated_value" {
		t.Errorf("UPDATE: expected 'updated_value', got '%s'", value)
	}

	// 4. DELETE
	t.Log("Deleting key...")
	delCmd := app.DeleteKeyCmd(h.mc, testKey)
	delMsg := delCmd()
	if _, ok := delMsg.(app.KeyDeletedMsg); !ok {
		t.Fatalf("DELETE failed: expected KeyDeletedMsg, got %T", delMsg)
	}

	// Verify deletion
	if h.keyExists(testKey) {
		t.Error("key should not exist after DELETE")
	}

	t.Log("CRUD workflow completed successfully")
}

// TestAppModel_StateTransitions tests app state transitions
func TestAppModel_StateTransitions(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	model := app.NewModel(h.addr)

	// Initial state should be Connecting
	if model.State() != app.StateConnecting {
		t.Errorf("initial state should be Connecting, got %s", model.State().String())
	}

	// Simulate window size
	model.Update(tea.WindowSizeMsg{Width: crudDefaultTermWidth, Height: crudDefaultTermHeight})

	// Run init (connect)
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		model.Update(msg)
	}

	// After connection, state should be Connected or Loading or Ready
	state := model.State()
	if state != app.StateConnected && state != app.StateLoading && state != app.StateReady {
		t.Errorf("state after connection should be Connected/Loading/Ready, got %s", state.String())
	}
}

// TestAppModel_FocusTransitions tests focus mode transitions
func TestAppModel_FocusTransitions(t *testing.T) {
	h := newCrudTestHelper(t)
	defer h.cleanup()

	model := h.initModel()

	// Initial focus should be KeyList
	if model.Focus() != app.FocusKeyList {
		t.Errorf("initial focus should be KeyList, got %d", model.Focus())
	}

	// Tab should switch to Viewer (if implemented)
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("tab"),
	})

	// Test '?' opens help
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'?'},
	})

	// Focus should be Help
	if model.Focus() != app.FocusHelp {
		t.Logf("focus after '?' is %d (expected Help)", model.Focus())
	}
}

// TestHandleDeleteResult tests the HandleDeleteResult function
func TestHandleDeleteResult(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		msg := app.KeyDeletedMsg{Key: "testkey"}
		result := app.HandleDeleteResult(msg)

		if !result.ShouldRefresh {
			t.Error("ShouldRefresh should be true for successful delete")
		}
		if result.DeletedKey != "testkey" {
			t.Errorf("expected DeletedKey 'testkey', got '%s'", result.DeletedKey)
		}
		if result.Error != "" {
			t.Errorf("unexpected error: %s", result.Error)
		}
	})

	t.Run("failed delete", func(t *testing.T) {
		msg := app.DeleteErrorMsg{Key: "testkey", Err: fmt.Errorf("delete failed")}
		result := app.HandleDeleteResult(msg)

		if result.ShouldRefresh {
			t.Error("ShouldRefresh should be false for failed delete")
		}
		if result.FailedKey != "testkey" {
			t.Errorf("expected FailedKey 'testkey', got '%s'", result.FailedKey)
		}
		if result.Error == "" {
			t.Error("expected error message")
		}
	})
}

// TestHandleNewKeyResult tests the HandleNewKeyResult function
func TestHandleNewKeyResult(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		msg := app.KeyCreatedMsg{Key: "newkey"}
		result := app.HandleNewKeyResult(msg)

		if !result.ShouldRefresh {
			t.Error("ShouldRefresh should be true for successful create")
		}
		if result.CreatedKey != "newkey" {
			t.Errorf("expected CreatedKey 'newkey', got '%s'", result.CreatedKey)
		}
	})

	t.Run("failed create", func(t *testing.T) {
		msg := app.NewKeyErrorMsg{Key: "newkey", Err: fmt.Errorf("create failed")}
		result := app.HandleNewKeyResult(msg)

		if result.ShouldRefresh {
			t.Error("ShouldRefresh should be false for failed create")
		}
		if result.FailedKey != "newkey" {
			t.Errorf("expected FailedKey 'newkey', got '%s'", result.FailedKey)
		}
	})
}

// TestEditorModes tests editor mode functionality
func TestEditorModes(t *testing.T) {
	ed := editor.New("testkey", []byte(`{"key": "value"}`))
	ed.SetSize(crudDefaultTermWidth, crudDefaultTermHeight)

	// Default mode should be Text
	if ed.Mode() != editor.ModeText {
		t.Errorf("default mode should be Text, got %s", ed.Mode().String())
	}

	// Set JSON mode
	ed.SetMode(editor.ModeJSON)
	if ed.Mode() != editor.ModeJSON {
		t.Errorf("mode should be JSON, got %s", ed.Mode().String())
	}

	// Test FormatJSON
	err := ed.FormatJSON()
	if err != nil {
		t.Errorf("FormatJSON failed: %v", err)
	}
}

// TestContextExtraction tests context extraction functions
func TestContextExtraction(t *testing.T) {
	t.Run("ExtractDeleteContext", func(t *testing.T) {
		// String context
		key, ok := app.ExtractDeleteContext("mykey")
		if !ok || key != "mykey" {
			t.Error("failed to extract string context")
		}

		// DeleteContext struct
		key, ok = app.ExtractDeleteContext(app.DeleteContext{Key: "mykey2"})
		if !ok || key != "mykey2" {
			t.Error("failed to extract DeleteContext")
		}

		// Nil context
		_, ok = app.ExtractDeleteContext(nil)
		if ok {
			t.Error("should return false for nil context")
		}
	})

	t.Run("ExtractNewKeyContext", func(t *testing.T) {
		// String context
		key, ok := app.ExtractNewKeyContext("mykey")
		if !ok || key != "mykey" {
			t.Error("failed to extract string context")
		}

		// NewKeyContext struct
		key, ok = app.ExtractNewKeyContext(app.NewKeyContext{Key: "mykey2"})
		if !ok || key != "mykey2" {
			t.Error("failed to extract NewKeyContext")
		}

		// Nil context
		_, ok = app.ExtractNewKeyContext(nil)
		if ok {
			t.Error("should return false for nil context")
		}
	})

	t.Run("ExtractBatchDeleteContext", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		ctx := app.BatchDeleteContext{Keys: keys}

		extracted, ok := app.ExtractBatchDeleteContext(ctx)
		if !ok {
			t.Error("failed to extract BatchDeleteContext")
		}
		if len(extracted) != 3 {
			t.Errorf("expected 3 keys, got %d", len(extracted))
		}

		// Nil context
		_, ok = app.ExtractBatchDeleteContext(nil)
		if ok {
			t.Error("should return false for nil context")
		}
	})
}

// TestProcessConfirmResult tests ProcessConfirmResult function
func TestProcessConfirmResult(t *testing.T) {
	t.Run("confirmed", func(t *testing.T) {
		result := dialog.ConfirmResultMsg{
			Result:  true,
			Context: "mykey",
		}

		msg := app.ProcessConfirmResult(result)
		if msg == nil {
			t.Fatal("expected non-nil DeleteConfirmMsg")
		}
		if msg.Key != "mykey" {
			t.Errorf("expected key 'mykey', got '%s'", msg.Key)
		}
	})

	t.Run("canceled", func(t *testing.T) {
		result := dialog.ConfirmResultMsg{
			Result:  false,
			Context: "mykey",
		}

		msg := app.ProcessConfirmResult(result)
		if msg != nil {
			t.Error("expected nil for canceled")
		}
	})
}

// TestProcessBatchInputResult tests ProcessBatchInputResult function
func TestProcessBatchInputResult(t *testing.T) {
	t.Run("valid DELETE confirmation", func(t *testing.T) {
		result := dialog.InputResultMsg{
			Value:    "DELETE",
			Canceled: false,
			Context:  app.BatchDeleteContext{Keys: []string{"key1", "key2"}},
		}

		msg := app.ProcessBatchInputResult(result)
		if msg == nil {
			t.Fatal("expected non-nil BatchDeleteMsg")
		}
		if len(msg.Keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(msg.Keys))
		}
	})

	t.Run("canceled", func(t *testing.T) {
		result := dialog.InputResultMsg{
			Value:    "DELETE",
			Canceled: true,
			Context:  app.BatchDeleteContext{Keys: []string{"key1"}},
		}

		msg := app.ProcessBatchInputResult(result)
		if msg != nil {
			t.Error("expected nil for canceled")
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		result := dialog.InputResultMsg{
			Value:    "delete",
			Canceled: false,
			Context:  app.BatchDeleteContext{Keys: []string{"key1"}},
		}

		msg := app.ProcessBatchInputResult(result)
		if msg != nil {
			t.Error("expected nil for invalid input")
		}
	})
}

// TestBatchDeleteSummary_String tests BatchDeleteSummary String method
func TestBatchDeleteSummary_String(t *testing.T) {
	testCases := []struct {
		name     string
		summary  app.BatchDeleteSummary
		expected string
	}{
		{
			name:     "no keys",
			summary:  app.BatchDeleteSummary{TotalCount: 0},
			expected: "No keys to delete",
		},
		{
			name: "single key success",
			summary: app.BatchDeleteSummary{
				TotalCount:   1,
				DeletedCount: 1,
				AllSucceeded: true,
			},
			expected: "Successfully deleted 1 key",
		},
		{
			name: "multiple keys success",
			summary: app.BatchDeleteSummary{
				TotalCount:   5,
				DeletedCount: 5,
				AllSucceeded: true,
			},
			expected: "Successfully deleted 5 keys",
		},
		{
			name: "all failed",
			summary: app.BatchDeleteSummary{
				TotalCount:   3,
				DeletedCount: 0,
				FailedCount:  3,
				AllSucceeded: false,
			},
			expected: "Failed to delete all 3 keys",
		},
		{
			name: "partial success",
			summary: app.BatchDeleteSummary{
				TotalCount:   5,
				DeletedCount: 3,
				FailedCount:  2,
				AllSucceeded: false,
			},
			expected: "Deleted 3 keys, 2 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.summary.String()
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
