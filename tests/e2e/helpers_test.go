//go:build e2e

package e2e_test

import (
	"os"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nnnkkk7/memtui/client"
)

// =============================================================================
// Shared Test Helpers for E2E Tests
// =============================================================================

// getMemcachedAddr returns the Memcached address for testing
func getMemcachedAddr() string {
	addr := os.Getenv("MEMCACHED_ADDR")
	if addr == "" {
		addr = "localhost:11211"
	}
	return addr
}

// skipIfNoMemcached skips the test if Memcached is not available
func skipIfNoMemcached(t *testing.T) {
	t.Helper()
	addr := getMemcachedAddr()
	detector := client.NewCapabilityDetector()
	_, err := detector.Detect(addr)
	if err != nil {
		t.Skipf("Memcached not available at %s: %v", addr, err)
	}
}

// setupTestKeys creates test keys in Memcached and returns a cleanup function
func setupTestKeys(t *testing.T, keys map[string]string) func() {
	t.Helper()
	addr := getMemcachedAddr()
	mc := memcache.New(addr)

	for key, value := range keys {
		err := mc.Set(&memcache.Item{
			Key:   key,
			Value: []byte(value),
		})
		if err != nil {
			t.Fatalf("failed to set test key %s: %v", key, err)
		}
	}

	// Wait for keys to be indexed by lru_crawler
	time.Sleep(100 * time.Millisecond)

	return func() {
		for key := range keys {
			mc.Delete(key)
		}
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
