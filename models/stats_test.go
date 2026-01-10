package models_test

import (
	"testing"

	"github.com/nnnkkk7/memtui/models"
)

func TestStats_Fields(t *testing.T) {
	stats := models.Stats{
		PID:                1234,
		Uptime:             86400,
		Version:            "1.6.22",
		CurrentConnections: 10,
		TotalConnections:   1000,
		CurrentItems:       500,
		TotalItems:         10000,
		Bytes:              1048576,
		LimitMaxBytes:      67108864,
		GetHits:            8000,
		GetMisses:          2000,
		Evictions:          50,
		BytesRead:          5000000,
		BytesWritten:       3000000,
	}

	if stats.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", stats.PID)
	}
	if stats.Version != "1.6.22" {
		t.Errorf("expected version '1.6.22', got '%s'", stats.Version)
	}
	if stats.CurrentItems != 500 {
		t.Errorf("expected CurrentItems 500, got %d", stats.CurrentItems)
	}
	if stats.LimitMaxBytes != 67108864 {
		t.Errorf("expected LimitMaxBytes 67108864, got %d", stats.LimitMaxBytes)
	}
}

func TestParseStatsResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(*testing.T, *models.Stats)
	}{
		{
			name: "basic stats response",
			input: `STAT pid 1234
STAT uptime 86400
STAT version 1.6.22
STAT curr_connections 10
STAT total_connections 1000
STAT curr_items 500
STAT total_items 10000
STAT bytes 1048576
STAT limit_maxbytes 67108864
STAT get_hits 8000
STAT get_misses 2000
STAT evictions 50
STAT bytes_read 5000000
STAT bytes_written 3000000
END
`,
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				if s.PID != 1234 {
					t.Errorf("PID: expected 1234, got %d", s.PID)
				}
				if s.Uptime != 86400 {
					t.Errorf("Uptime: expected 86400, got %d", s.Uptime)
				}
				if s.Version != "1.6.22" {
					t.Errorf("Version: expected '1.6.22', got '%s'", s.Version)
				}
				if s.CurrentConnections != 10 {
					t.Errorf("CurrentConnections: expected 10, got %d", s.CurrentConnections)
				}
				if s.TotalConnections != 1000 {
					t.Errorf("TotalConnections: expected 1000, got %d", s.TotalConnections)
				}
				if s.CurrentItems != 500 {
					t.Errorf("CurrentItems: expected 500, got %d", s.CurrentItems)
				}
				if s.TotalItems != 10000 {
					t.Errorf("TotalItems: expected 10000, got %d", s.TotalItems)
				}
				if s.Bytes != 1048576 {
					t.Errorf("Bytes: expected 1048576, got %d", s.Bytes)
				}
				if s.LimitMaxBytes != 67108864 {
					t.Errorf("LimitMaxBytes: expected 67108864, got %d", s.LimitMaxBytes)
				}
				if s.GetHits != 8000 {
					t.Errorf("GetHits: expected 8000, got %d", s.GetHits)
				}
				if s.GetMisses != 2000 {
					t.Errorf("GetMisses: expected 2000, got %d", s.GetMisses)
				}
				if s.Evictions != 50 {
					t.Errorf("Evictions: expected 50, got %d", s.Evictions)
				}
				if s.BytesRead != 5000000 {
					t.Errorf("BytesRead: expected 5000000, got %d", s.BytesRead)
				}
				if s.BytesWritten != 3000000 {
					t.Errorf("BytesWritten: expected 3000000, got %d", s.BytesWritten)
				}
			},
		},
		{
			name: "stats with CRLF line endings",
			input: "STAT pid 5678\r\nSTAT uptime 3600\r\nSTAT version 1.4.31\r\nSTAT curr_items 100\r\nEND\r\n",
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				if s.PID != 5678 {
					t.Errorf("PID: expected 5678, got %d", s.PID)
				}
				if s.Uptime != 3600 {
					t.Errorf("Uptime: expected 3600, got %d", s.Uptime)
				}
				if s.Version != "1.4.31" {
					t.Errorf("Version: expected '1.4.31', got '%s'", s.Version)
				}
				if s.CurrentItems != 100 {
					t.Errorf("CurrentItems: expected 100, got %d", s.CurrentItems)
				}
			},
		},
		{
			name: "stats with trailing CR in values",
			input: "STAT version 1.6.22\r\nSTAT curr_items 200\r\nEND\r\n",
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				// Values should not have trailing \r
				if s.Version != "1.6.22" {
					t.Errorf("Version: expected '1.6.22', got '%s' (len=%d)", s.Version, len(s.Version))
				}
				if s.CurrentItems != 200 {
					t.Errorf("CurrentItems: expected 200, got %d", s.CurrentItems)
				}
			},
		},
		{
			name:    "empty response",
			input:   "",
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				if s.Version != "" {
					t.Errorf("Version: expected empty, got '%s'", s.Version)
				}
			},
		},
		{
			name:    "END only",
			input:   "END\r\n",
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				if s.Version != "" {
					t.Errorf("Version: expected empty, got '%s'", s.Version)
				}
			},
		},
		{
			name: "malformed line - missing value",
			input: "STAT pid\nSTAT version 1.6.22\nEND\n",
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				// pid should be skipped, version should be parsed
				if s.Version != "1.6.22" {
					t.Errorf("Version: expected '1.6.22', got '%s'", s.Version)
				}
			},
		},
		{
			name: "non-STAT lines ignored",
			input: "SERVER_ERROR some error\nSTAT version 1.6.22\nEND\n",
			wantErr: false,
			validate: func(t *testing.T, s *models.Stats) {
				if s.Version != "1.6.22" {
					t.Errorf("Version: expected '1.6.22', got '%s'", s.Version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := models.ParseStatsResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if stats == nil {
				t.Error("expected non-nil stats")
				return
			}
			if tt.validate != nil {
				tt.validate(t, stats)
			}
		})
	}
}

func TestParseStatsResponse_HitRate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantHitRate float64
	}{
		{
			name: "80% hit rate",
			input: `STAT get_hits 8000
STAT get_misses 2000
END
`,
			wantHitRate: 80.0,
		},
		{
			name: "100% hit rate",
			input: `STAT get_hits 1000
STAT get_misses 0
END
`,
			wantHitRate: 100.0,
		},
		{
			name: "0% hit rate",
			input: `STAT get_hits 0
STAT get_misses 1000
END
`,
			wantHitRate: 0.0,
		},
		{
			name: "no hits or misses (0 total)",
			input: `STAT get_hits 0
STAT get_misses 0
END
`,
			wantHitRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := models.ParseStatsResponse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stats.HitRate != tt.wantHitRate {
				t.Errorf("HitRate: expected %.2f, got %.2f", tt.wantHitRate, stats.HitRate)
			}
		})
	}
}

func TestParseStatsResponse_RawMap(t *testing.T) {
	input := `STAT pid 1234
STAT custom_stat some_value
STAT version 1.6.22
END
`
	stats, err := models.ParseStatsResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Raw map should contain all STAT values
	if stats.Raw == nil {
		t.Fatal("Raw map is nil")
	}

	if stats.Raw["pid"] != "1234" {
		t.Errorf("Raw[pid]: expected '1234', got '%s'", stats.Raw["pid"])
	}
	if stats.Raw["custom_stat"] != "some_value" {
		t.Errorf("Raw[custom_stat]: expected 'some_value', got '%s'", stats.Raw["custom_stat"])
	}
	if stats.Raw["version"] != "1.6.22" {
		t.Errorf("Raw[version]: expected '1.6.22', got '%s'", stats.Raw["version"])
	}
}

func TestStats_MemoryUsagePercent(t *testing.T) {
	tests := []struct {
		name          string
		bytes         int64
		limitMaxBytes int64
		wantPercent   float64
	}{
		{
			name:          "50% usage",
			bytes:         33554432,
			limitMaxBytes: 67108864,
			wantPercent:   50.0,
		},
		{
			name:          "100% usage",
			bytes:         67108864,
			limitMaxBytes: 67108864,
			wantPercent:   100.0,
		},
		{
			name:          "0% usage",
			bytes:         0,
			limitMaxBytes: 67108864,
			wantPercent:   0.0,
		},
		{
			name:          "limit is 0 (edge case)",
			bytes:         1000,
			limitMaxBytes: 0,
			wantPercent:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &models.Stats{
				Bytes:         tt.bytes,
				LimitMaxBytes: tt.limitMaxBytes,
			}
			got := stats.MemoryUsagePercent()
			if got != tt.wantPercent {
				t.Errorf("MemoryUsagePercent: expected %.2f, got %.2f", tt.wantPercent, got)
			}
		})
	}
}

func TestStats_UptimeFormatted(t *testing.T) {
	tests := []struct {
		name   string
		uptime int64
		want   string
	}{
		{
			name:   "less than a minute",
			uptime: 45,
			want:   "45s",
		},
		{
			name:   "minutes only",
			uptime: 300,
			want:   "5m 0s",
		},
		{
			name:   "hours and minutes",
			uptime: 3661,
			want:   "1h 1m 1s",
		},
		{
			name:   "days",
			uptime: 90061, // 1 day, 1 hour, 1 minute, 1 second
			want:   "1d 1h 1m 1s",
		},
		{
			name:   "zero",
			uptime: 0,
			want:   "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &models.Stats{Uptime: tt.uptime}
			got := stats.UptimeFormatted()
			if got != tt.want {
				t.Errorf("UptimeFormatted: expected '%s', got '%s'", tt.want, got)
			}
		})
	}
}

func TestStats_BytesFormatted(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 500,
			want:  "500 B",
		},
		{
			name:  "kilobytes",
			bytes: 2048,
			want:  "2.00 KB",
		},
		{
			name:  "megabytes",
			bytes: 1048576,
			want:  "1.00 MB",
		},
		{
			name:  "gigabytes",
			bytes: 1073741824,
			want:  "1.00 GB",
		},
		{
			name:  "zero",
			bytes: 0,
			want:  "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &models.Stats{Bytes: tt.bytes}
			got := stats.BytesFormatted()
			if got != tt.want {
				t.Errorf("BytesFormatted: expected '%s', got '%s'", tt.want, got)
			}
		})
	}
}

func TestSlabItemStats_Fields(t *testing.T) {
	s := models.SlabItemStats{
		SlabID:    1,
		Number:    100,
		Age:       3600,
		Evicted:   10,
		EvictedNZ: 5,
		Outofmem:  0,
	}

	if s.SlabID != 1 {
		t.Errorf("SlabID: expected 1, got %d", s.SlabID)
	}
	if s.Number != 100 {
		t.Errorf("Number: expected 100, got %d", s.Number)
	}
}

func TestSlabStats_Fields(t *testing.T) {
	s := models.SlabStats{
		SlabID:     1,
		ChunkSize:  96,
		Chunks:     10922,
		UsedChunks: 500,
		FreeChunks: 10422,
		MemReq:     48000,
	}

	if s.SlabID != 1 {
		t.Errorf("SlabID: expected 1, got %d", s.SlabID)
	}
	if s.ChunkSize != 96 {
		t.Errorf("ChunkSize: expected 96, got %d", s.ChunkSize)
	}
}

func TestSlabsStats_Fields(t *testing.T) {
	ss := models.SlabsStats{
		ActiveSlabs: 5,
		TotalMalloced: 67108864,
		Slabs: map[int]*models.SlabStats{
			1: {SlabID: 1, ChunkSize: 96},
			2: {SlabID: 2, ChunkSize: 120},
		},
	}

	if ss.ActiveSlabs != 5 {
		t.Errorf("ActiveSlabs: expected 5, got %d", ss.ActiveSlabs)
	}
	if ss.TotalMalloced != 67108864 {
		t.Errorf("TotalMalloced: expected 67108864, got %d", ss.TotalMalloced)
	}
	if len(ss.Slabs) != 2 {
		t.Errorf("Slabs count: expected 2, got %d", len(ss.Slabs))
	}
}
