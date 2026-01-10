package viewer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nnnkkk7/memtui/models"
	viewerPkg "github.com/nnnkkk7/memtui/viewer"
)

// ViewMode represents the display mode
type ViewMode int

const (
	ViewModeAuto ViewMode = iota
	ViewModeJSON
	ViewModeHex
	ViewModeText
)

// String returns the string representation of the view mode
func (v ViewMode) String() string {
	switch v {
	case ViewModeAuto:
		return "Auto"
	case ViewModeJSON:
		return "JSON"
	case ViewModeHex:
		return "Hex"
	case ViewModeText:
		return "Text"
	default:
		return "Unknown"
	}
}

// Model represents the value viewer component
type Model struct {
	value        []byte
	keyInfo      models.KeyInfo
	viewMode     ViewMode
	detectedType viewerPkg.DataType
	content      string
	scrollOffset int
	width        int
	height       int

	// Formatters
	jsonFormatter *viewerPkg.JSONFormatter
	hexFormatter  *viewerPkg.HexFormatter
	textFormatter *viewerPkg.TextFormatter
	autoFormatter *viewerPkg.AutoFormatter

	// Styles
	headerStyle  lipgloss.Style
	contentStyle lipgloss.Style
	metaStyle    lipgloss.Style
}

// NewModel creates a new viewer model
func NewModel() *Model {
	return &Model{
		viewMode:      ViewModeAuto,
		jsonFormatter: viewerPkg.NewJSONFormatter(),
		hexFormatter:  viewerPkg.NewHexFormatter(),
		textFormatter: viewerPkg.NewTextFormatter(),
		autoFormatter: viewerPkg.NewAutoFormatter(),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")),
		contentStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		metaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}

// SetValue sets the value to display
func (m *Model) SetValue(value []byte) {
	m.value = value
	m.detectedType = viewerPkg.DetectType(value)
	m.formatContent()
	m.scrollOffset = 0
}

// SetKeyInfo sets the key info
func (m *Model) SetKeyInfo(ki models.KeyInfo) {
	m.keyInfo = ki
}

// SetViewMode sets the view mode
func (m *Model) SetViewMode(mode ViewMode) {
	m.viewMode = mode
	m.formatContent()
}

// SetSize sets the component size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ViewMode returns the current view mode
func (m *Model) ViewMode() ViewMode {
	return m.viewMode
}

// Content returns the formatted content
func (m *Model) Content() string {
	return m.content
}

// DetectedType returns the detected data type as a string
func (m *Model) DetectedType() string {
	return m.detectedType.String()
}

// ScrollOffset returns the current scroll offset
func (m *Model) ScrollOffset() int {
	return m.scrollOffset
}

// formatContent formats the value based on view mode
func (m *Model) formatContent() {
	if len(m.value) == 0 {
		m.content = ""
		return
	}

	var err error
	switch m.viewMode {
	case ViewModeJSON:
		m.content, err = m.jsonFormatter.Format(m.value)
		if err != nil {
			m.content = string(m.value)
		}
	case ViewModeHex:
		m.content, _ = m.hexFormatter.Format(m.value)
	case ViewModeText:
		m.content, _ = m.textFormatter.Format(m.value)
	case ViewModeAuto:
		m.content, _ = m.autoFormatter.Format(m.value)
	}
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case tea.KeyDown:
			m.scrollOffset++
		case tea.KeyPgUp:
			pageSize := m.height - 4
			if pageSize < 1 {
				pageSize = 10
			}
			m.scrollOffset -= pageSize
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
		case tea.KeyPgDown:
			pageSize := m.height - 4
			if pageSize < 1 {
				pageSize = 10
			}
			m.scrollOffset += pageSize
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "J":
				m.SetViewMode(ViewModeJSON)
			case "H":
				m.SetViewMode(ViewModeHex)
			case "T":
				m.SetViewMode(ViewModeText)
			case "A":
				m.SetViewMode(ViewModeAuto)
			}
		}
	}

	return m, nil
}

// View renders the viewer
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "No size set"
	}

	var b strings.Builder

	// Header with key info
	if m.keyInfo.Key != "" {
		header := m.headerStyle.Render(m.keyInfo.Key)
		b.WriteString(header)
		b.WriteString("\n")

		// Metadata line
		meta := fmt.Sprintf("Size: %d bytes | Type: %s | Mode: %s",
			m.keyInfo.Size, m.detectedType.String(), m.viewMode.String())
		b.WriteString(m.metaStyle.Render(meta))
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", min(m.width, 60)))
		b.WriteString("\n")
	}

	// Content
	if len(m.value) == 0 {
		b.WriteString(m.metaStyle.Render("No value loaded"))
		return b.String()
	}

	lines := strings.Split(m.content, "\n")
	contentHeight := m.height - 4 // Reserve space for header/footer
	if contentHeight < 1 {
		contentHeight = 10
	}

	// Clamp scroll offset
	maxOffset := len(lines) - contentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}

	// Render visible lines
	endLine := m.scrollOffset + contentHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	for i := m.scrollOffset; i < endLine; i++ {
		line := lines[i]
		// Truncate long lines
		if m.width > 0 && len(line) > m.width {
			line = line[:m.width-3] + "..."
		}
		b.WriteString(m.contentStyle.Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
