// Package models defines the data structures used throughout memtui.
package models

import "time"

// Item represents an item stored in Memcached.
type Item struct {
	Key   string
	Value []byte
	Flags uint32
	CAS   uint64

	// TTL-related fields (from metadump)
	Expiration int64 // Unix timestamp (0 = permanent)
	LastAccess int64 // Last access time

	// Computed fields
	TTLRemaining int64 // Remaining TTL in seconds (calculated in real-time)
	IsExpired    bool  // Expiration flag
}

// UpdateTTL updates TTL-related fields based on the current time.
func (i *Item) UpdateTTL() {
	now := time.Now().Unix()
	if i.Expiration == 0 {
		i.TTLRemaining = -1 // permanent
		i.IsExpired = false
	} else if i.Expiration <= now {
		i.TTLRemaining = 0
		i.IsExpired = true
	} else {
		i.TTLRemaining = i.Expiration - now
		i.IsExpired = false
	}
}
