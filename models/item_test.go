package models

import (
	"testing"
	"time"
)

func TestItem_Struct(t *testing.T) {
	t.Run("Item has required fields", func(t *testing.T) {
		item := Item{
			Key:          "test:key",
			Value:        []byte("test value"),
			Flags:        0x0001,
			CAS:          1234567890,
			Expiration:   time.Now().Unix() + 3600,
			LastAccess:   time.Now().Unix(),
			TTLRemaining: 3600,
			IsExpired:    false,
		}

		if item.Key != "test:key" {
			t.Errorf("Key = %v, want %v", item.Key, "test:key")
		}
		if string(item.Value) != "test value" {
			t.Errorf("Value = %v, want %v", string(item.Value), "test value")
		}
		if item.Flags != 0x0001 {
			t.Errorf("Flags = %v, want %v", item.Flags, 0x0001)
		}
		if item.CAS != 1234567890 {
			t.Errorf("CAS = %v, want %v", item.CAS, 1234567890)
		}
	})

	t.Run("Item with nil Value", func(t *testing.T) {
		item := Item{
			Key:   "empty:key",
			Value: nil,
		}

		if item.Value != nil {
			t.Errorf("Value should be nil, got %v", item.Value)
		}
	})

	t.Run("Item with empty Value", func(t *testing.T) {
		item := Item{
			Key:   "empty:key",
			Value: []byte{},
		}

		if len(item.Value) != 0 {
			t.Errorf("Value should be empty, got %v", item.Value)
		}
	})
}

func TestItem_UpdateTTL(t *testing.T) {
	t.Run("permanent item (Expiration = 0)", func(t *testing.T) {
		item := &Item{
			Key:        "permanent:key",
			Expiration: 0, // permanent
		}

		item.UpdateTTL()

		if item.TTLRemaining != -1 {
			t.Errorf("TTLRemaining = %v, want -1 for permanent item", item.TTLRemaining)
		}
		if item.IsExpired {
			t.Errorf("IsExpired = %v, want false for permanent item", item.IsExpired)
		}
	})

	t.Run("valid TTL item (future expiration)", func(t *testing.T) {
		now := time.Now().Unix()
		futureExp := now + 3600 // 1 hour from now

		item := &Item{
			Key:        "future:key",
			Expiration: futureExp,
		}

		item.UpdateTTL()

		// TTLRemaining should be close to 3600 (allow 1 second tolerance)
		if item.TTLRemaining < 3599 || item.TTLRemaining > 3600 {
			t.Errorf("TTLRemaining = %v, want ~3600", item.TTLRemaining)
		}
		if item.IsExpired {
			t.Errorf("IsExpired = %v, want false for future expiration", item.IsExpired)
		}
	})

	t.Run("expired item (past expiration)", func(t *testing.T) {
		now := time.Now().Unix()
		pastExp := now - 100 // 100 seconds ago

		item := &Item{
			Key:        "expired:key",
			Expiration: pastExp,
		}

		item.UpdateTTL()

		if item.TTLRemaining != 0 {
			t.Errorf("TTLRemaining = %v, want 0 for expired item", item.TTLRemaining)
		}
		if !item.IsExpired {
			t.Errorf("IsExpired = %v, want true for expired item", item.IsExpired)
		}
	})

	t.Run("item expiring now (Expiration = now)", func(t *testing.T) {
		now := time.Now().Unix()

		item := &Item{
			Key:        "expiring:now",
			Expiration: now, // exactly now
		}

		item.UpdateTTL()

		// Expiration <= now means expired
		if item.TTLRemaining != 0 {
			t.Errorf("TTLRemaining = %v, want 0 for item expiring now", item.TTLRemaining)
		}
		if !item.IsExpired {
			t.Errorf("IsExpired = %v, want true for item expiring now", item.IsExpired)
		}
	})

	t.Run("UpdateTTL can be called multiple times", func(t *testing.T) {
		now := time.Now().Unix()
		item := &Item{
			Key:        "multi:update",
			Expiration: now + 100,
		}

		// First update
		item.UpdateTTL()
		if item.IsExpired {
			t.Error("Should not be expired on first update")
		}

		// Update again - should still work correctly
		item.UpdateTTL()
		if item.IsExpired {
			t.Error("Should not be expired on second update")
		}
	})

	t.Run("TTL calculation accuracy", func(t *testing.T) {
		now := time.Now().Unix()
		testCases := []struct {
			name          string
			expiration    int64
			wantTTLMin    int64
			wantTTLMax    int64
			wantIsExpired bool
		}{
			{"1 second remaining", now + 1, 0, 1, false},
			{"1 minute remaining", now + 60, 59, 60, false},
			{"1 hour remaining", now + 3600, 3599, 3600, false},
			{"1 day remaining", now + 86400, 86399, 86400, false},
			{"30 days remaining", now + 2592000, 2591999, 2592000, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				item := &Item{
					Key:        "test:key",
					Expiration: tc.expiration,
				}

				item.UpdateTTL()

				if item.TTLRemaining < tc.wantTTLMin || item.TTLRemaining > tc.wantTTLMax {
					t.Errorf("TTLRemaining = %v, want between %v and %v",
						item.TTLRemaining, tc.wantTTLMin, tc.wantTTLMax)
				}
				if item.IsExpired != tc.wantIsExpired {
					t.Errorf("IsExpired = %v, want %v", item.IsExpired, tc.wantIsExpired)
				}
			})
		}
	})
}

func TestItem_EdgeCases(t *testing.T) {
	t.Run("zero value item", func(t *testing.T) {
		item := &Item{}

		// Expiration is 0, so it's permanent
		item.UpdateTTL()

		if item.TTLRemaining != -1 {
			t.Errorf("TTLRemaining = %v, want -1 for zero value item", item.TTLRemaining)
		}
		if item.IsExpired {
			t.Errorf("IsExpired = %v, want false for zero value item", item.IsExpired)
		}
	})

	t.Run("max int64 expiration", func(t *testing.T) {
		item := &Item{
			Key:        "max:exp",
			Expiration: 1<<63 - 1, // max int64
		}

		item.UpdateTTL()

		// Should have very large TTL, not expired
		if item.IsExpired {
			t.Error("Should not be expired with max int64 expiration")
		}
		if item.TTLRemaining <= 0 {
			t.Errorf("TTLRemaining should be positive, got %v", item.TTLRemaining)
		}
	})

	t.Run("negative expiration treated as past", func(t *testing.T) {
		item := &Item{
			Key:        "negative:exp",
			Expiration: -1,
		}

		item.UpdateTTL()

		// Negative expiration is before now, so it's expired
		if item.TTLRemaining != 0 {
			t.Errorf("TTLRemaining = %v, want 0 for negative expiration", item.TTLRemaining)
		}
		if !item.IsExpired {
			t.Error("Should be expired with negative expiration")
		}
	})
}
