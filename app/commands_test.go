package app

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/models"
)

// mockMemcachedClient implements client.MemcachedClient for testing
type mockMemcachedClient struct {
	lastSetItem *memcache.Item
	lastCASItem *client.CASItem
	setErr      error
	casErr      error
	getWithCAS  *client.CASItem
	getWithErr  error
}

func (m *mockMemcachedClient) Get(key string) (*memcache.Item, error) {
	return nil, nil
}

func (m *mockMemcachedClient) GetWithCAS(key string) (*client.CASItem, error) {
	if m.getWithErr != nil {
		return nil, m.getWithErr
	}
	return m.getWithCAS, nil
}

func (m *mockMemcachedClient) Set(item *memcache.Item) error {
	m.lastSetItem = item
	return m.setErr
}

func (m *mockMemcachedClient) CompareAndSwap(item *client.CASItem) error {
	m.lastCASItem = item
	return m.casErr
}

func (m *mockMemcachedClient) Delete(key string) error {
	return nil
}

func (m *mockMemcachedClient) Close() error {
	return nil
}

func (m *mockMemcachedClient) Address() string {
	return "localhost:11211"
}

// TestCalculateRemainingTTL tests the TTL calculation from absolute expiration timestamps
func TestCalculateRemainingTTL(t *testing.T) {
	t.Run("zero expiration returns zero", func(t *testing.T) {
		got := calculateRemainingTTL(0)
		if got != 0 {
			t.Errorf("calculateRemainingTTL(0) = %d, want 0", got)
		}
	})

	t.Run("past expiration returns zero", func(t *testing.T) {
		pastExp := time.Now().Unix() - 100
		got := calculateRemainingTTL(pastExp)
		if got != 0 {
			t.Errorf("calculateRemainingTTL(%d) = %d, want 0", pastExp, got)
		}
	})

	t.Run("current time returns zero", func(t *testing.T) {
		now := time.Now().Unix()
		got := calculateRemainingTTL(now)
		if got != 0 {
			t.Errorf("calculateRemainingTTL(%d) = %d, want 0", now, got)
		}
	})

	t.Run("future expiration returns positive TTL", func(t *testing.T) {
		futureExp := time.Now().Unix() + 300
		got := calculateRemainingTTL(futureExp)

		// Allow some tolerance for test execution time
		if got < 295 || got > 305 {
			t.Errorf("calculateRemainingTTL(%d) = %d, want ~300", futureExp, got)
		}
	})

	t.Run("far future expiration returns positive TTL", func(t *testing.T) {
		futureExp := time.Now().Unix() + 86400 // 1 day
		got := calculateRemainingTTL(futureExp)

		if got < 86395 || got > 86405 {
			t.Errorf("calculateRemainingTTL(%d) = %d, want ~86400", futureExp, got)
		}
	})

	t.Run("very large expiration is capped at max int32", func(t *testing.T) {
		veryFarFuture := time.Now().Unix() + int64(math.MaxInt32) + 1000
		got := calculateRemainingTTL(veryFarFuture)

		if got != math.MaxInt32 {
			t.Errorf("calculateRemainingTTL with overflow = %d, want %d", got, math.MaxInt32)
		}
	})
}

// TestSaveValueCmd_PreservesTTLAndFlags tests that editing preserves TTL and flags
func TestSaveValueCmd_PreservesTTLAndFlags(t *testing.T) {
	t.Run("preserves flags from CASItem and TTL from KeyInfo", func(t *testing.T) {
		mock := &mockMemcachedClient{}
		m := &Model{mcClient: mock}

		futureExp := time.Now().Unix() + 600 // 10 minutes
		keyInfo := &models.KeyInfo{
			Key:        "test-key",
			Expiration: futureExp,
		}
		casItem := &client.CASItem{
			Key:   "test-key",
			Flags: 42,
		}

		cmd := m.saveValueCmd("test-key", []byte("new value"), keyInfo, casItem)
		msg := cmd()

		if _, ok := msg.(KeyCreatedMsg); !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T: %v", msg, msg)
		}

		if mock.lastSetItem == nil {
			t.Fatal("expected Set to be called")
		}

		if mock.lastSetItem.Key != "test-key" {
			t.Errorf("expected key 'test-key', got '%s'", mock.lastSetItem.Key)
		}

		if string(mock.lastSetItem.Value) != "new value" {
			t.Errorf("expected value 'new value', got '%s'", string(mock.lastSetItem.Value))
		}

		if mock.lastSetItem.Flags != 42 {
			t.Errorf("expected flags 42, got %d", mock.lastSetItem.Flags)
		}

		// TTL should be approximately 600 seconds
		if mock.lastSetItem.Expiration < 595 || mock.lastSetItem.Expiration > 605 {
			t.Errorf("expected expiration ~600, got %d", mock.lastSetItem.Expiration)
		}
	})

	t.Run("handles nil keyInfo and casItem gracefully", func(t *testing.T) {
		mock := &mockMemcachedClient{}
		m := &Model{mcClient: mock}

		cmd := m.saveValueCmd("test-key", []byte("value"), nil, nil)
		msg := cmd()

		if _, ok := msg.(KeyCreatedMsg); !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T: %v", msg, msg)
		}

		if mock.lastSetItem.Flags != 0 {
			t.Errorf("expected flags 0, got %d", mock.lastSetItem.Flags)
		}

		if mock.lastSetItem.Expiration != 0 {
			t.Errorf("expected expiration 0, got %d", mock.lastSetItem.Expiration)
		}
	})

	t.Run("returns zero TTL for expired keys", func(t *testing.T) {
		mock := &mockMemcachedClient{}
		m := &Model{mcClient: mock}

		pastExp := time.Now().Unix() - 100
		keyInfo := &models.KeyInfo{
			Key:        "test-key",
			Expiration: pastExp,
		}

		cmd := m.saveValueCmd("test-key", []byte("value"), keyInfo, nil)
		cmd()

		if mock.lastSetItem.Expiration != 0 {
			t.Errorf("expected expiration 0 for expired key, got %d", mock.lastSetItem.Expiration)
		}
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		m := &Model{mcClient: nil}

		cmd := m.saveValueCmd("test-key", []byte("value"), nil, nil)
		msg := cmd()

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err != "client not connected" {
			t.Errorf("expected 'client not connected', got '%s'", errMsg.Err)
		}
	})

	t.Run("returns error when Set fails", func(t *testing.T) {
		mock := &mockMemcachedClient{
			setErr: errors.New("connection refused"),
		}
		m := &Model{mcClient: mock}

		cmd := m.saveValueCmd("test-key", []byte("value"), nil, nil)
		msg := cmd()

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err == "" {
			t.Error("expected non-empty error message")
		}
	})
}

// TestSaveValueWithCASCmd_PreservesTTLAndFlags tests CAS update preserves TTL and flags
func TestSaveValueWithCASCmd_PreservesTTLAndFlags(t *testing.T) {
	t.Run("preserves flags and TTL during CAS update", func(t *testing.T) {
		mockCASItem := &client.CASItem{
			Key:   "test-key",
			Value: []byte("old value"),
			Flags: 0,
		}
		mock := &mockMemcachedClient{getWithCAS: mockCASItem}
		m := &Model{mcClient: mock}

		originalCAS := &client.CASItem{
			Key:   "test-key",
			Flags: 99,
		}
		futureExp := time.Now().Unix() + 1800 // 30 minutes
		keyInfo := &models.KeyInfo{
			Key:        "test-key",
			Expiration: futureExp,
		}

		cmd := m.saveValueWithCASCmd("test-key", []byte("updated value"), originalCAS, keyInfo)
		msg := cmd()

		if _, ok := msg.(KeyCreatedMsg); !ok {
			t.Fatalf("expected KeyCreatedMsg, got %T: %v", msg, msg)
		}

		if mock.lastCASItem == nil {
			t.Fatal("expected CompareAndSwap to be called")
		}

		if string(mock.lastCASItem.Value) != "updated value" {
			t.Errorf("expected value 'updated value', got '%s'", string(mock.lastCASItem.Value))
		}

		if mock.lastCASItem.Flags != 99 {
			t.Errorf("expected flags 99, got %d", mock.lastCASItem.Flags)
		}

		// TTL should be approximately 1800 seconds
		if mock.lastCASItem.Expiration < 1795 || mock.lastCASItem.Expiration > 1805 {
			t.Errorf("expected expiration ~1800, got %d", mock.lastCASItem.Expiration)
		}
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		m := &Model{mcClient: nil}

		cmd := m.saveValueWithCASCmd("test-key", []byte("value"), nil, nil)
		msg := cmd()

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err != "client not connected" {
			t.Errorf("expected 'client not connected', got '%s'", errMsg.Err)
		}
	})

	t.Run("returns error when GetWithCAS fails", func(t *testing.T) {
		mock := &mockMemcachedClient{
			getWithErr: errors.New("key not found"),
		}
		m := &Model{mcClient: mock}

		cmd := m.saveValueWithCASCmd("test-key", []byte("value"), nil, nil)
		msg := cmd()

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("returns CAS conflict error when CompareAndSwap fails", func(t *testing.T) {
		mockCASItem := &client.CASItem{
			Key:   "test-key",
			Value: []byte("value"),
		}
		mock := &mockMemcachedClient{
			getWithCAS: mockCASItem,
			casErr:     client.NewCASConflictError("test-key"),
		}
		m := &Model{mcClient: mock}

		cmd := m.saveValueWithCASCmd("test-key", []byte("new value"), nil, nil)
		msg := cmd()

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err == "" {
			t.Error("expected non-empty error message for CAS conflict")
		}
	})
}
