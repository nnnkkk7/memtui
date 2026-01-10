// Package stats provides a Bubble Tea component for displaying Memcached server statistics.
package stats

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/styles"
)

// RefreshStatsMsg is a message requesting a stats refresh
type RefreshStatsMsg struct{}

// StatsView is a Bubble Tea component for displaying Memcached statistics.
type StatsView struct {
	stats  *models.Stats
	width  int
	height int

	// Styles
	titleStyle   lipgloss.Style
	sectionStyle lipgloss.Style
	labelStyle   lipgloss.Style
	valueStyle   lipgloss.Style
	mutedStyle   lipgloss.Style
	goodStyle    lipgloss.Style
	warnStyle    lipgloss.Style
	badStyle     lipgloss.Style
}

// New creates a new StatsView component.
func New() *StatsView {
	theme := styles.DefaultTheme()

	return &StatsView{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1),
		sectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Secondary).
			MarginTop(1).
			MarginBottom(0),
		labelStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted).
			Width(20),
		valueStyle: lipgloss.NewStyle().
			Foreground(theme.Text),
		mutedStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted),
		goodStyle: lipgloss.NewStyle().
			Foreground(theme.Success),
		warnStyle: lipgloss.NewStyle().
			Foreground(theme.Warning),
		badStyle: lipgloss.NewStyle().
			Foreground(theme.Error),
	}
}

// SetStats sets the statistics to display.
func (s *StatsView) SetStats(stats *models.Stats) {
	s.stats = stats
}

// SetSize sets the component dimensions.
func (s *StatsView) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// Stats returns the current stats.
func (s *StatsView) Stats() *models.Stats {
	return s.stats
}

// Init initializes the component.
func (s *StatsView) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s *StatsView) Update(msg tea.Msg) (*StatsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "r", "R":
				return s, func() tea.Msg { return RefreshStatsMsg{} }
			}
		}
	}

	return s, nil
}

// View renders the component.
func (s *StatsView) View() string {
	if s.stats == nil {
		return s.mutedStyle.Render("No statistics available")
	}

	var b strings.Builder

	// Title
	b.WriteString(s.titleStyle.Render("Memcached Statistics"))
	b.WriteString("\n")

	// Server Info Section
	b.WriteString(s.renderSection("Server Info"))
	b.WriteString(s.renderRow("Version", s.stats.Version))
	b.WriteString(s.renderRow("PID", fmt.Sprintf("%d", s.stats.PID)))
	b.WriteString(s.renderRow("Uptime", s.formatUptime()))

	// Connections Section
	b.WriteString(s.renderSection("Connections"))
	b.WriteString(s.renderRow("Current", fmt.Sprintf("%d", s.stats.CurrentConnections)))
	b.WriteString(s.renderRow("Total", fmt.Sprintf("%d", s.stats.TotalConnections)))

	// Items Section
	b.WriteString(s.renderSection("Items"))
	b.WriteString(s.renderRow("Current Items", fmt.Sprintf("%d", s.stats.CurrentItems)))
	b.WriteString(s.renderRow("Total Items", fmt.Sprintf("%d", s.stats.TotalItems)))
	b.WriteString(s.renderRow("Evictions", s.formatEvictions()))

	// Memory Section
	b.WriteString(s.renderSection("Memory"))
	b.WriteString(s.renderRow("Used", s.stats.BytesFormatted()))
	b.WriteString(s.renderRow("Limit", models.FormatBytes(s.stats.LimitMaxBytes)))
	b.WriteString(s.renderRow("Usage", s.formatMemoryPercentColored()))

	// Performance Section
	b.WriteString(s.renderSection("Performance"))
	b.WriteString(s.renderRow("Hit Rate", s.formatHitRateColored()))
	b.WriteString(s.renderRow("Get Hits", fmt.Sprintf("%d", s.stats.GetHits)))
	b.WriteString(s.renderRow("Get Misses", fmt.Sprintf("%d", s.stats.GetMisses)))

	// Network I/O Section
	b.WriteString(s.renderSection("Network I/O"))
	b.WriteString(s.renderRow("Bytes Read", models.FormatBytes(s.stats.BytesRead)))
	b.WriteString(s.renderRow("Bytes Written", models.FormatBytes(s.stats.BytesWritten)))

	// Footer with refresh hint
	b.WriteString("\n")
	b.WriteString(s.mutedStyle.Render("Press 'r' to refresh"))

	return b.String()
}

// renderSection renders a section header.
func (s *StatsView) renderSection(title string) string {
	return "\n" + s.sectionStyle.Render(title) + "\n"
}

// renderRow renders a label-value row.
func (s *StatsView) renderRow(label, value string) string {
	return s.labelStyle.Render(label+":") + " " + s.valueStyle.Render(value) + "\n"
}

// formatMemoryPercent returns the memory usage as a percentage string.
func (s *StatsView) formatMemoryPercent() string {
	if s.stats == nil {
		return "0.00%"
	}
	percent := s.stats.MemoryUsagePercent()
	return fmt.Sprintf("%.2f%%", percent)
}

// formatMemoryPercentColored returns memory usage with color coding.
func (s *StatsView) formatMemoryPercentColored() string {
	if s.stats == nil {
		return s.mutedStyle.Render("0.00%")
	}

	percent := s.stats.MemoryUsagePercent()
	percentStr := fmt.Sprintf("%.2f%%", percent)

	switch {
	case percent >= 90:
		return s.badStyle.Render(percentStr)
	case percent >= 70:
		return s.warnStyle.Render(percentStr)
	default:
		return s.goodStyle.Render(percentStr)
	}
}

// formatHitRate returns the hit rate as a percentage string.
func (s *StatsView) formatHitRate() string {
	if s.stats == nil {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", s.stats.HitRate)
}

// formatHitRateColored returns hit rate with color coding.
func (s *StatsView) formatHitRateColored() string {
	if s.stats == nil {
		return s.mutedStyle.Render("0.00%")
	}

	hitRate := s.stats.HitRate
	hitRateStr := fmt.Sprintf("%.2f%%", hitRate)

	switch {
	case hitRate >= 90:
		return s.goodStyle.Render(hitRateStr)
	case hitRate >= 70:
		return s.warnStyle.Render(hitRateStr)
	default:
		return s.badStyle.Render(hitRateStr)
	}
}

// formatUptime returns the uptime in a human-readable format.
func (s *StatsView) formatUptime() string {
	if s.stats == nil {
		return "0s"
	}
	return s.stats.UptimeFormatted()
}

// formatEvictions returns evictions with color coding.
func (s *StatsView) formatEvictions() string {
	if s.stats == nil {
		return "0"
	}

	evictions := s.stats.Evictions
	evictStr := fmt.Sprintf("%d", evictions)

	switch {
	case evictions > 1000:
		return s.badStyle.Render(evictStr)
	case evictions > 100:
		return s.warnStyle.Render(evictStr)
	default:
		return s.valueStyle.Render(evictStr)
	}
}
