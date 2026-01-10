package models

import (
	"fmt"
	"strconv"
	"strings"
)

// Stats represents Memcached server statistics from the 'stats' command.
type Stats struct {
	// Process info
	PID    int   // Process ID
	Uptime int64 // Seconds since server start

	// Version
	Version string // Memcached version string

	// Connections
	CurrentConnections int   // Current open connections
	TotalConnections   int64 // Total connections since start

	// Items
	CurrentItems int64 // Current number of items stored
	TotalItems   int64 // Total items stored since start

	// Memory
	Bytes         int64 // Current bytes used for storage
	LimitMaxBytes int64 // Maximum bytes allowed (limit_maxbytes)

	// Cache performance
	GetHits   int64   // Cache hits
	GetMisses int64   // Cache misses
	Evictions int64   // Items evicted to free memory
	HitRate   float64 // Calculated hit rate percentage

	// Network I/O
	BytesRead    int64 // Total bytes read from network
	BytesWritten int64 // Total bytes written to network

	// Raw contains all STAT key-value pairs
	Raw map[string]string
}

// SlabItemStats represents stats for items in a slab class (from 'stats items').
type SlabItemStats struct {
	SlabID    int   // Slab class ID
	Number    int64 // Number of items in this slab class
	Age       int64 // Age of oldest item in seconds
	Evicted   int64 // Items evicted from this slab class
	EvictedNZ int64 // Items evicted with non-zero TTL
	Outofmem  int64 // Times memory allocation failed
}

// SlabStats represents stats for a slab class (from 'stats slabs').
type SlabStats struct {
	SlabID     int   // Slab class ID
	ChunkSize  int64 // Bytes per chunk
	Chunks     int64 // Total chunks in this slab class
	UsedChunks int64 // Chunks currently in use
	FreeChunks int64 // Chunks not in use
	MemReq     int64 // Memory requested for this slab class
}

// SlabsStats represents overall slab statistics.
type SlabsStats struct {
	ActiveSlabs   int                // Number of active slab classes
	TotalMalloced int64              // Total memory allocated for slabs
	Slabs         map[int]*SlabStats // Per-slab statistics
}

// ParseStatsResponse parses the response from a Memcached 'stats' command.
// The response format is:
//
//	STAT <key> <value>\r\n
//	...
//	END\r\n
func ParseStatsResponse(response string) (*Stats, error) {
	stats := &Stats{
		Raw: make(map[string]string),
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		// Handle \r\n line endings from Memcached protocol
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimSpace(line)

		if line == "" || line == "END" {
			continue
		}

		// Parse "STAT key value" format
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 || parts[0] != "STAT" {
			continue
		}

		key := parts[1]
		value := strings.TrimSuffix(parts[2], "\r") // Remove any trailing \r

		// Store in raw map
		stats.Raw[key] = value

		// Parse known fields
		switch key {
		case "pid":
			stats.PID, _ = strconv.Atoi(value)
		case "uptime":
			stats.Uptime, _ = strconv.ParseInt(value, 10, 64)
		case "version":
			stats.Version = value
		case "curr_connections":
			stats.CurrentConnections, _ = strconv.Atoi(value)
		case "total_connections":
			stats.TotalConnections, _ = strconv.ParseInt(value, 10, 64)
		case "curr_items":
			stats.CurrentItems, _ = strconv.ParseInt(value, 10, 64)
		case "total_items":
			stats.TotalItems, _ = strconv.ParseInt(value, 10, 64)
		case "bytes":
			stats.Bytes, _ = strconv.ParseInt(value, 10, 64)
		case "limit_maxbytes":
			stats.LimitMaxBytes, _ = strconv.ParseInt(value, 10, 64)
		case "get_hits":
			stats.GetHits, _ = strconv.ParseInt(value, 10, 64)
		case "get_misses":
			stats.GetMisses, _ = strconv.ParseInt(value, 10, 64)
		case "evictions":
			stats.Evictions, _ = strconv.ParseInt(value, 10, 64)
		case "bytes_read":
			stats.BytesRead, _ = strconv.ParseInt(value, 10, 64)
		case "bytes_written":
			stats.BytesWritten, _ = strconv.ParseInt(value, 10, 64)
		}
	}

	// Calculate hit rate
	total := stats.GetHits + stats.GetMisses
	if total > 0 {
		stats.HitRate = float64(stats.GetHits) / float64(total) * 100
	}

	return stats, nil
}

// MemoryUsagePercent returns the percentage of memory used.
func (s *Stats) MemoryUsagePercent() float64 {
	if s.LimitMaxBytes == 0 {
		return 0.0
	}
	return float64(s.Bytes) / float64(s.LimitMaxBytes) * 100
}

// UptimeFormatted returns the uptime in a human-readable format.
func (s *Stats) UptimeFormatted() string {
	uptime := s.Uptime

	days := uptime / 86400
	uptime %= 86400

	hours := uptime / 3600
	uptime %= 3600

	minutes := uptime / 60
	seconds := uptime % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// BytesFormatted returns the bytes used in a human-readable format.
func (s *Stats) BytesFormatted() string {
	return FormatBytes(s.Bytes)
}

// FormatBytes formats bytes in human-readable units.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
