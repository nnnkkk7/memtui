package client

import (
	"errors"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

// CASItem represents a Memcached item with CAS (Compare-And-Swap) support.
// The CAS token is used for optimistic locking - it ensures that the item
// has not been modified by another client between read and write operations.
type CASItem struct {
	// Key is the item's key (max 250 bytes)
	Key string

	// Value is the item's value
	Value []byte

	// Flags are server-opaque flags whose semantics are defined by the client
	Flags uint32

	// Expiration is the cache expiration time, in seconds:
	// 0 means no expiration, values up to 30 days are interpreted as relative,
	// larger values are interpreted as absolute Unix timestamps
	Expiration int32

	// CAS is the compare-and-swap token returned by GetWithCAS.
	// This value is informational only - the actual CAS token is stored
	// internally in the mcItem field for use with CompareAndSwap.
	CAS uint64

	// mcItem holds the original memcache.Item with its internal casid.
	// This is necessary because gomemcache's casid field is unexported.
	mcItem *memcache.Item
}

// CASConflictError is returned when a CompareAndSwap operation fails
// because the item has been modified by another client since it was read.
type CASConflictError struct {
	key string
}

// NewCASConflictError creates a new CASConflictError for the given key.
func NewCASConflictError(key string) *CASConflictError {
	return &CASConflictError{key: key}
}

// Error implements the error interface.
func (e *CASConflictError) Error() string {
	return fmt.Sprintf("CAS conflict for key '%s': item has been modified", e.key)
}

// Key returns the key that caused the conflict.
func (e *CASConflictError) Key() string {
	return e.key
}

// NewCASItem creates a new CASItem with the specified values.
// Note: Items created this way cannot be used with CompareAndSwap
// as they don't have a valid CAS token from the server.
func NewCASItem(key string, value []byte, flags uint32, expiration int32, cas uint64) *CASItem {
	return &CASItem{
		Key:        key,
		Value:      value,
		Flags:      flags,
		Expiration: expiration,
		CAS:        cas,
		mcItem:     nil,
	}
}

// ToMemcacheItem converts a CASItem to a gomemcache Item.
// Note: The CAS value is handled separately as gomemcache uses internal casid field.
func (item *CASItem) ToMemcacheItem() *memcache.Item {
	return &memcache.Item{
		Key:        item.Key,
		Value:      item.Value,
		Flags:      item.Flags,
		Expiration: item.Expiration,
	}
}

// GetWithCAS retrieves an item by key along with its CAS token.
// The returned CASItem can be modified and passed to CompareAndSwap
// to perform an atomic update.
//
// Returns:
//   - *CASItem: The item with its current value and CAS token
//   - error: ErrCacheMiss if the key doesn't exist, or other errors
func (c *Client) GetWithCAS(key string) (*CASItem, error) {
	// Use GetMulti to retrieve the item with CAS token populated
	// gomemcache populates the internal casid field when using GetMulti
	items, err := c.mc.GetMulti([]string{key})
	if err != nil {
		return nil, err
	}

	mcItem, ok := items[key]
	if !ok {
		return nil, memcache.ErrCacheMiss
	}

	// Create a CASItem that holds the original memcache.Item
	// This preserves the internal casid for use with CompareAndSwap
	return &CASItem{
		Key:        mcItem.Key,
		Value:      mcItem.Value,
		Flags:      mcItem.Flags,
		Expiration: mcItem.Expiration,
		CAS:        extractCASValue(mcItem),
		mcItem:     mcItem,
	}, nil
}

// extractCASValue attempts to extract the CAS value from the memcache.Item.
// Since casid is unexported, we can't read it directly.
// We return a non-zero placeholder to indicate CAS is available.
// The actual CAS comparison is done by gomemcache internally.
func extractCASValue(item *memcache.Item) uint64 {
	// The memcache.Item stores CAS in an unexported 'casid' field.
	// We can't access it directly, but gomemcache handles it internally
	// during CompareAndSwap operations.
	// Return 1 as a placeholder to indicate that CAS is available.
	if item != nil {
		return 1 // Placeholder indicating CAS token is present
	}
	return 0
}

// CompareAndSwap atomically updates an item only if it hasn't been modified
// since it was last read (i.e., the CAS token still matches).
//
// The typical usage pattern is:
//  1. Call GetWithCAS to get the current value and CAS token
//  2. Modify the Value field of the returned CASItem
//  3. Call CompareAndSwap with the modified CASItem
//
// Returns:
//   - nil: The update was successful
//   - *CASConflictError: Another client modified the item (CAS mismatch)
//   - other errors: Network or server errors
func (c *Client) CompareAndSwap(item *CASItem) error {
	if item == nil {
		return errors.New("CASItem cannot be nil")
	}

	// Check if this item has a valid CAS token from GetWithCAS
	if item.mcItem == nil {
		// Item was not retrieved via GetWithCAS, so we can't do a proper CAS
		// We need to fetch it first to get a CAS token
		return errors.New("CASItem must be retrieved via GetWithCAS before CompareAndSwap")
	}

	// Update the internal memcache.Item with the new values
	// while preserving the casid for the CAS operation
	item.mcItem.Value = item.Value
	item.mcItem.Flags = item.Flags
	item.mcItem.Expiration = item.Expiration

	// Perform the CAS operation using gomemcache
	err := c.mc.CompareAndSwap(item.mcItem)
	if err != nil {
		if errors.Is(err, memcache.ErrCASConflict) {
			return NewCASConflictError(item.Key)
		}
		if errors.Is(err, memcache.ErrNotStored) {
			// This can happen if the item was deleted or expired
			return NewCASConflictError(item.Key)
		}
		return err
	}

	return nil
}

// SetWithExpiration is a convenience method to set a key with expiration.
// This is useful for setting up test data before CAS operations.
func (c *Client) SetWithExpiration(key string, value []byte, flags uint32, expiration int32) error {
	return c.mc.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Flags:      flags,
		Expiration: expiration,
	})
}

// IsCacheMiss checks if the error is a cache miss error.
func IsCacheMiss(err error) bool {
	return errors.Is(err, memcache.ErrCacheMiss)
}

// IsCASConflict checks if the error is a CAS conflict error.
func IsCASConflict(err error) bool {
	var casErr *CASConflictError
	return errors.As(err, &casErr)
}
