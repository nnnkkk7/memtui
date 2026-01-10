package stats

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/models"
)

// createTestStats creates a Stats struct with test data
func createTestStats() *models.Stats {
	return &models.Stats{
		PID:                12345,
		Uptime:             86400 + 3600 + 120 + 5, // 1d 1h 2m 5s
		Version:            "1.6.18",
		CurrentConnections: 10,
		TotalConnections:   1000,
		CurrentItems:       500,
		TotalItems:         10000,
		Bytes:              52428800,     // 50 MB
		LimitMaxBytes:      104857600,    // 100 MB
		GetHits:            8000,
		GetMisses:          2000,
		Evictions:          100,
		HitRate:            80.0,
		BytesRead:          1073741824,   // 1 GB
		BytesWritten:       536870912,    // 512 MB
		Raw:                make(map[string]string),
	}
}

func TestStatsView_New(t *testing.T) {
	t.Run("creates new StatsView with default state", func(t *testing.T) {
		view := New()

		if view == nil {
			t.Fatal("New() returned nil")
		}

		if view.stats != nil {
			t.Error("expected stats to be nil initially")
		}

		if view.width != 0 {
			t.Errorf("expected width to be 0, got %d", view.width)
		}

		if view.height != 0 {
			t.Errorf("expected height to be 0, got %d", view.height)
		}
	})
}

func TestStatsView_SetStats(t *testing.T) {
	t.Run("sets stats data", func(t *testing.T) {
		view := New()
		stats := createTestStats()

		view.SetStats(stats)

		if view.stats == nil {
			t.Fatal("SetStats() did not set stats")
		}

		if view.stats.Version != "1.6.18" {
			t.Errorf("expected version '1.6.18', got '%s'", view.stats.Version)
		}

		if view.stats.PID != 12345 {
			t.Errorf("expected PID 12345, got %d", view.stats.PID)
		}
	})

	t.Run("handles nil stats", func(t *testing.T) {
		view := New()
		view.SetStats(nil)

		if view.stats != nil {
			t.Error("expected stats to be nil")
		}
	})
}

func TestStatsView_SetSize(t *testing.T) {
	t.Run("sets width and height", func(t *testing.T) {
		view := New()
		view.SetSize(80, 24)

		if view.width != 80 {
			t.Errorf("expected width 80, got %d", view.width)
		}

		if view.height != 24 {
			t.Errorf("expected height 24, got %d", view.height)
		}
	})
}

func TestStatsView_View(t *testing.T) {
	t.Run("renders placeholder when no stats", func(t *testing.T) {
		view := New()
		view.SetSize(80, 24)

		output := view.View()

		if !strings.Contains(output, "No statistics available") {
			t.Errorf("expected 'No statistics available' message, got: %s", output)
		}
	})

	t.Run("renders all metrics properly", func(t *testing.T) {
		view := New()
		view.SetSize(80, 40)
		stats := createTestStats()
		view.SetStats(stats)

		output := view.View()

		// Check section headers
		if !strings.Contains(output, "Server Info") {
			t.Error("expected 'Server Info' section header")
		}
		if !strings.Contains(output, "Performance") {
			t.Error("expected 'Performance' section header")
		}
		if !strings.Contains(output, "Memory") {
			t.Error("expected 'Memory' section header")
		}

		// Check key metrics
		if !strings.Contains(output, "1.6.18") {
			t.Error("expected version '1.6.18' in output")
		}
		if !strings.Contains(output, "1d 1h 2m 5s") {
			t.Error("expected formatted uptime '1d 1h 2m 5s' in output")
		}
		if !strings.Contains(output, "10") {
			t.Error("expected current connections '10' in output")
		}
		if !strings.Contains(output, "500") {
			t.Error("expected current items '500' in output")
		}
		if !strings.Contains(output, "80.00%") {
			t.Error("expected hit rate '80.00%' in output")
		}
		if !strings.Contains(output, "50.00%") {
			t.Error("expected memory usage '50.00%' in output")
		}
	})

	t.Run("handles zero size", func(t *testing.T) {
		view := New()
		stats := createTestStats()
		view.SetStats(stats)

		output := view.View()

		// Should still render something even with zero size
		if output == "" {
			t.Error("expected non-empty output even with zero size")
		}
	})
}

func TestStatsView_Update(t *testing.T) {
	t.Run("handles resize message", func(t *testing.T) {
		view := New()
		view.SetSize(80, 24)

		resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
		updatedView, cmd := view.Update(resizeMsg)

		if cmd != nil {
			t.Error("expected no command from resize")
		}

		if updatedView.width != 120 {
			t.Errorf("expected width 120 after resize, got %d", updatedView.width)
		}

		if updatedView.height != 40 {
			t.Errorf("expected height 40 after resize, got %d", updatedView.height)
		}
	})

	t.Run("handles refresh key", func(t *testing.T) {
		view := New()
		view.SetSize(80, 24)

		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
		_, cmd := view.Update(keyMsg)

		// Should return a refresh command
		if cmd == nil {
			t.Error("expected refresh command from 'r' key")
		}
	})

	t.Run("handles R key for refresh", func(t *testing.T) {
		view := New()
		view.SetSize(80, 24)

		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
		_, cmd := view.Update(keyMsg)

		// Should return a refresh command
		if cmd == nil {
			t.Error("expected refresh command from 'R' key")
		}
	})

	t.Run("handles unknown key gracefully", func(t *testing.T) {
		view := New()
		view.SetSize(80, 24)

		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		updatedView, cmd := view.Update(keyMsg)

		if cmd != nil {
			t.Error("expected no command from unknown key")
		}

		if updatedView == nil {
			t.Error("expected non-nil view")
		}
	})
}

func TestStatsView_FormatMetrics(t *testing.T) {
	t.Run("formats memory percentage correctly", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			Bytes:         52428800,  // 50 MB
			LimitMaxBytes: 104857600, // 100 MB
		}
		view.SetStats(stats)

		result := view.formatMemoryPercent()

		if result != "50.00%" {
			t.Errorf("expected '50.00%%', got '%s'", result)
		}
	})

	t.Run("formats zero memory percentage", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			Bytes:         0,
			LimitMaxBytes: 104857600,
		}
		view.SetStats(stats)

		result := view.formatMemoryPercent()

		if result != "0.00%" {
			t.Errorf("expected '0.00%%', got '%s'", result)
		}
	})

	t.Run("handles zero max memory", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			Bytes:         1000,
			LimitMaxBytes: 0,
		}
		view.SetStats(stats)

		result := view.formatMemoryPercent()

		if result != "0.00%" {
			t.Errorf("expected '0.00%%' when max is 0, got '%s'", result)
		}
	})

	t.Run("formats hit rate correctly", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			HitRate: 85.5,
		}
		view.SetStats(stats)

		result := view.formatHitRate()

		if result != "85.50%" {
			t.Errorf("expected '85.50%%', got '%s'", result)
		}
	})

	t.Run("formats zero hit rate", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			HitRate: 0,
		}
		view.SetStats(stats)

		result := view.formatHitRate()

		if result != "0.00%" {
			t.Errorf("expected '0.00%%', got '%s'", result)
		}
	})

	t.Run("formats uptime correctly", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			Uptime: 90061, // 1d 1h 1m 1s
		}
		view.SetStats(stats)

		result := view.formatUptime()

		if result != "1d 1h 1m 1s" {
			t.Errorf("expected '1d 1h 1m 1s', got '%s'", result)
		}
	})

	t.Run("formats short uptime (hours only)", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			Uptime: 3661, // 1h 1m 1s
		}
		view.SetStats(stats)

		result := view.formatUptime()

		if result != "1h 1m 1s" {
			t.Errorf("expected '1h 1m 1s', got '%s'", result)
		}
	})

	t.Run("formats very short uptime (seconds only)", func(t *testing.T) {
		view := New()
		stats := &models.Stats{
			Uptime: 45,
		}
		view.SetStats(stats)

		result := view.formatUptime()

		if result != "45s" {
			t.Errorf("expected '45s', got '%s'", result)
		}
	})
}

func TestStatsView_Init(t *testing.T) {
	t.Run("returns nil command", func(t *testing.T) {
		view := New()
		cmd := view.Init()

		if cmd != nil {
			t.Error("expected Init() to return nil command")
		}
	})
}

// RefreshStatsMsg is the message type for refresh requests
func TestRefreshStatsMsg(t *testing.T) {
	t.Run("RefreshStatsMsg is defined", func(t *testing.T) {
		msg := RefreshStatsMsg{}
		_ = msg // Just verify it compiles
	})
}
