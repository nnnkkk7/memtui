package client

import (
	"github.com/bradfitz/gomemcache/memcache"
)

// MemcachedClient defines the interface for memcached operations.
// This interface unifies basic operations and CAS (Compare-And-Swap) support.
type MemcachedClient interface {
	// Get retrieves an item by key
	Get(key string) (*memcache.Item, error)

	// GetWithCAS retrieves an item by key along with its CAS token
	// for optimistic locking
	GetWithCAS(key string) (*CASItem, error)

	// Set stores an item
	Set(item *memcache.Item) error

	// CompareAndSwap atomically updates an item only if it hasn't been
	// modified since it was last read
	CompareAndSwap(item *CASItem) error

	// Delete removes an item by key
	Delete(key string) error

	// Close closes the client connection
	Close() error

	// Address returns the server address
	Address() string
}

// Ensure Client implements MemcachedClient interface
var _ MemcachedClient = (*Client)(nil)
