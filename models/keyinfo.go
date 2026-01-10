package models

import (
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// KeyInfo represents metadata for a Memcached key from lru_crawler metadump
type KeyInfo struct {
	Key        string // Key name
	Expiration int64  // Unix timestamp (0 = permanent)
	LastAccess int64  // Last access timestamp
	CAS        uint64 // CAS value
	Fetch      bool   // Whether fetch has been performed
	SlabClass  int    // Slab class ID
	Size       int    // Size in bytes
}

// ErrInvalidMetadumpLine is returned when a metadump line cannot be parsed
var ErrInvalidMetadumpLine = errors.New("invalid metadump line")

// ParseMetadumpLine parses a single line from lru_crawler metadump output
// Format: key=<key> exp=<exp> la=<la> cas=<cas> fetch=<yes|no> cls=<cls> size=<size>
func ParseMetadumpLine(line string) (KeyInfo, error) {
	// Handle trailing \r from Memcached protocol
	line = strings.TrimSuffix(line, "\r")
	line = strings.TrimSpace(line)

	if line == "" || line == "END" {
		return KeyInfo{}, ErrInvalidMetadumpLine
	}

	ki := KeyInfo{}
	fields := strings.Fields(line)

	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "key":
			// Memcached metadump URL-encodes keys, so decode them
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				ki.Key = value // Use original if decode fails
			} else {
				ki.Key = decoded
			}
		case "exp":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				ki.Expiration = v
			}
		case "la":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				ki.LastAccess = v
			}
		case "cas":
			if v, err := strconv.ParseUint(value, 10, 64); err == nil {
				ki.CAS = v
			}
		case "fetch":
			ki.Fetch = (value == "yes")
		case "cls":
			if v, err := strconv.Atoi(value); err == nil {
				ki.SlabClass = v
			}
		case "size":
			if v, err := strconv.Atoi(value); err == nil {
				ki.Size = v
			}
		}
	}

	if ki.Key == "" {
		return KeyInfo{}, ErrInvalidMetadumpLine
	}

	return ki, nil
}

// IsExpiredAt checks if the key is expired at the given timestamp
// Returns false for permanent keys (Expiration == 0)
func (ki KeyInfo) IsExpiredAt(now int64) bool {
	if ki.Expiration == 0 {
		return false
	}
	return ki.Expiration < now
}

// SortOrder represents the sort order for KeyInfo slices
type SortOrder int

// Sort order options
const (
	// SortByKey sorts alphabetically by key name
	SortByKey SortOrder = iota
	// SortBySize sorts by size (ascending)
	SortBySize
)

// SortKeyInfos returns a sorted copy of the KeyInfo slice
func SortKeyInfos(keys []KeyInfo, order SortOrder) []KeyInfo {
	result := make([]KeyInfo, len(keys))
	copy(result, keys)

	switch order {
	case SortByKey:
		sort.Slice(result, func(i, j int) bool {
			return result[i].Key < result[j].Key
		})
	case SortBySize:
		sort.Slice(result, func(i, j int) bool {
			return result[i].Size < result[j].Size
		})
	}

	return result
}

// FilterKeyInfos returns keys that contain the given pattern
func FilterKeyInfos(keys []KeyInfo, pattern string) []KeyInfo {
	if pattern == "" {
		result := make([]KeyInfo, len(keys))
		copy(result, keys)
		return result
	}

	var result []KeyInfo
	for _, ki := range keys {
		if strings.Contains(ki.Key, pattern) {
			result = append(result, ki)
		}
	}
	return result
}
