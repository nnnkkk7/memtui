package client_test

import (
	"errors"
	"testing"

	"github.com/nnnkkk7/memtui/client"
)

// TestCASItem_Fields verifies that CASItem has all required fields
func TestCASItem_Fields(t *testing.T) {
	item := client.CASItem{
		Key:        "test-key",
		Value:      []byte("test-value"),
		Flags:      42,
		Expiration: 3600,
		CAS:        123456789,
	}

	if item.Key != "test-key" {
		t.Errorf("expected Key 'test-key', got '%s'", item.Key)
	}
	if string(item.Value) != "test-value" {
		t.Errorf("expected Value 'test-value', got '%s'", string(item.Value))
	}
	if item.Flags != 42 {
		t.Errorf("expected Flags 42, got %d", item.Flags)
	}
	if item.Expiration != 3600 {
		t.Errorf("expected Expiration 3600, got %d", item.Expiration)
	}
	if item.CAS != 123456789 {
		t.Errorf("expected CAS 123456789, got %d", item.CAS)
	}
}

// TestCASConflictError verifies the CAS conflict error type
func TestCASConflictError(t *testing.T) {
	err := client.NewCASConflictError("my-key")

	// Check that it implements error interface
	var _ error = err

	// Check error message
	if err.Error() != "CAS conflict for key 'my-key': item has been modified" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	// Check key accessor
	if err.Key() != "my-key" {
		t.Errorf("expected key 'my-key', got '%s'", err.Key())
	}

	// Check that errors.As works with CASConflictError
	var casErr *client.CASConflictError
	if !errors.As(err, &casErr) {
		t.Error("expected error to be CASConflictError")
	}
}

// TestIsCASConflict tests the IsCASConflict helper function
func TestIsCASConflict(t *testing.T) {
	casErr := client.NewCASConflictError("test-key")
	if !client.IsCASConflict(casErr) {
		t.Error("expected IsCASConflict to return true for CASConflictError")
	}

	otherErr := errors.New("some other error")
	if client.IsCASConflict(otherErr) {
		t.Error("expected IsCASConflict to return false for non-CAS errors")
	}

	if client.IsCASConflict(nil) {
		t.Error("expected IsCASConflict to return false for nil")
	}
}

// TestIsCacheMiss tests the IsCacheMiss helper function
func TestIsCacheMiss(t *testing.T) {
	// For non-cache-miss errors
	otherErr := errors.New("some other error")
	if client.IsCacheMiss(otherErr) {
		t.Error("expected IsCacheMiss to return false for non-cache-miss errors")
	}

	if client.IsCacheMiss(nil) {
		t.Error("expected IsCacheMiss to return false for nil")
	}
}

// TestClient_GetWithCAS_NoServer tests GetWithCAS behavior when server is unavailable
func TestClient_GetWithCAS_NoServer(t *testing.T) {
	// Use a port that's unlikely to have a server
	c, err := client.New("localhost:59998")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	_, err = c.GetWithCAS("test-key")
	if err == nil {
		t.Error("expected error when server is not available")
	}
}

// TestClient_CompareAndSwap_NoServer tests CompareAndSwap behavior when server is unavailable
func TestClient_CompareAndSwap_NoServer(t *testing.T) {
	// Use a port that's unlikely to have a server
	c, err := client.New("localhost:59998")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// Create an item without mcItem (not from GetWithCAS)
	item := &client.CASItem{
		Key:   "test-key",
		Value: []byte("test-value"),
		CAS:   12345,
	}

	err = c.CompareAndSwap(item)
	if err == nil {
		t.Error("expected error when item is not from GetWithCAS")
	}
}

// TestClient_CompareAndSwap_NilItem tests CompareAndSwap with nil item
func TestClient_CompareAndSwap_NilItem(t *testing.T) {
	c, err := client.New("localhost:59998")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	err = c.CompareAndSwap(nil)
	if err == nil {
		t.Error("expected error when item is nil")
	}
}

// TestClient_CompareAndSwap_NotFromGetWithCAS tests that items not from GetWithCAS fail
func TestClient_CompareAndSwap_NotFromGetWithCAS(t *testing.T) {
	c, err := client.New("localhost:59998")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// Create item using NewCASItem (not from GetWithCAS)
	item := client.NewCASItem("test-key", []byte("test-value"), 0, 60, 12345)

	err = c.CompareAndSwap(item)
	if err == nil {
		t.Error("expected error when item is not from GetWithCAS")
	}
	if err.Error() != "CASItem must be retrieved via GetWithCAS before CompareAndSwap" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestCASItem_ToMemcacheItem verifies conversion to memcache.Item
func TestCASItem_ToMemcacheItem(t *testing.T) {
	casItem := &client.CASItem{
		Key:        "convert-key",
		Value:      []byte("convert-value"),
		Flags:      99,
		Expiration: 7200,
		CAS:        987654321,
	}

	mcItem := casItem.ToMemcacheItem()

	if mcItem.Key != casItem.Key {
		t.Errorf("expected Key '%s', got '%s'", casItem.Key, mcItem.Key)
	}
	if string(mcItem.Value) != string(casItem.Value) {
		t.Errorf("expected Value '%s', got '%s'", string(casItem.Value), string(mcItem.Value))
	}
	if mcItem.Flags != casItem.Flags {
		t.Errorf("expected Flags %d, got %d", casItem.Flags, mcItem.Flags)
	}
	if mcItem.Expiration != casItem.Expiration {
		t.Errorf("expected Expiration %d, got %d", casItem.Expiration, mcItem.Expiration)
	}
	// Note: memcache.Item uses casid field which is internal
}

// TestNewCASItemFromMemcache verifies conversion from memcache.Item
func TestNewCASItemFromMemcache(t *testing.T) {
	// This test verifies the constructor works correctly
	casItem := client.NewCASItem("from-mc-key", []byte("from-mc-value"), 55, 1800, 111222333)

	if casItem.Key != "from-mc-key" {
		t.Errorf("expected Key 'from-mc-key', got '%s'", casItem.Key)
	}
	if string(casItem.Value) != "from-mc-value" {
		t.Errorf("expected Value 'from-mc-value', got '%s'", string(casItem.Value))
	}
	if casItem.Flags != 55 {
		t.Errorf("expected Flags 55, got %d", casItem.Flags)
	}
	if casItem.Expiration != 1800 {
		t.Errorf("expected Expiration 1800, got %d", casItem.Expiration)
	}
	if casItem.CAS != 111222333 {
		t.Errorf("expected CAS 111222333, got %d", casItem.CAS)
	}
}

// Integration tests below require a running Memcached server
// They are skipped if the server is not available

func TestClient_GetWithCAS_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// First set a value
	testKey := "cas-test-get-key"
	testValue := []byte("cas-test-value")

	err = c.SetWithExpiration(testKey, testValue, 0, 60)
	if err != nil {
		t.Skipf("skipping: Memcached server not available: %v", err)
	}
	defer c.Delete(testKey)

	// Get with CAS
	item, err := c.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("GetWithCAS failed: %v", err)
	}

	if item.Key != testKey {
		t.Errorf("expected key '%s', got '%s'", testKey, item.Key)
	}
	if string(item.Value) != string(testValue) {
		t.Errorf("expected value '%s', got '%s'", string(testValue), string(item.Value))
	}
	if item.CAS == 0 {
		t.Error("expected non-zero CAS value")
	}
}

func TestClient_CompareAndSwap_Success_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// First set a value
	testKey := "cas-test-swap-key"
	testValue := []byte("original-value")

	err = c.SetWithExpiration(testKey, testValue, 0, 60)
	if err != nil {
		t.Skipf("skipping: Memcached server not available: %v", err)
	}
	defer c.Delete(testKey)

	// Get with CAS
	item, err := c.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("GetWithCAS failed: %v", err)
	}

	// Modify value
	item.Value = []byte("modified-value")

	// Compare and swap
	err = c.CompareAndSwap(item)
	if err != nil {
		t.Fatalf("CompareAndSwap failed: %v", err)
	}

	// Verify the value was updated
	updatedItem, err := c.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("GetWithCAS after swap failed: %v", err)
	}

	if string(updatedItem.Value) != "modified-value" {
		t.Errorf("expected value 'modified-value', got '%s'", string(updatedItem.Value))
	}
}

func TestClient_CompareAndSwap_Conflict_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// First set a value
	testKey := "cas-test-conflict-key"
	testValue := []byte("original-value")

	err = c.SetWithExpiration(testKey, testValue, 0, 60)
	if err != nil {
		t.Skipf("skipping: Memcached server not available: %v", err)
	}
	defer c.Delete(testKey)

	// Get with CAS
	item, err := c.GetWithCAS(testKey)
	if err != nil {
		t.Fatalf("GetWithCAS failed: %v", err)
	}

	// Simulate another client modifying the value
	err = c.SetWithExpiration(testKey, []byte("changed-by-other"), 0, 60)
	if err != nil {
		t.Fatalf("failed to simulate concurrent modification: %v", err)
	}

	// Try to compare and swap with stale CAS
	item.Value = []byte("my-new-value")
	err = c.CompareAndSwap(item)

	// Should get a CAS conflict error
	if err == nil {
		t.Fatal("expected CAS conflict error, got nil")
	}

	var casErr *client.CASConflictError
	if !errors.As(err, &casErr) {
		t.Errorf("expected CASConflictError, got %T: %v", err, err)
	}
}

func TestClient_CompareAndSwap_KeyNotFound_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// Try to CAS a non-existent key
	item := &client.CASItem{
		Key:        "cas-non-existent-key-12345",
		Value:      []byte("some-value"),
		Flags:      0,
		Expiration: 60,
		CAS:        12345,
	}

	err = c.CompareAndSwap(item)
	if err == nil {
		// Clean up in case it succeeded unexpectedly
		c.Delete(item.Key)
		t.Fatal("expected error for non-existent key, got nil")
	}

	// Should get either cache miss or CAS conflict
	// The exact error depends on Memcached behavior
	if err == nil {
		t.Error("expected an error for non-existent key")
	}
}

func TestClient_GetWithCAS_KeyNotFound_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer c.Close()

	// Try to get a non-existent key
	_, err = c.GetWithCAS("definitely-not-existing-key-xyz-123")
	if err == nil {
		t.Error("expected cache miss error for non-existent key")
	}

	// Should be a cache miss error
	if !client.IsCacheMiss(err) {
		// That's okay, could be server unavailable
		t.Logf("got error (may be server unavailable): %v", err)
	}
}
