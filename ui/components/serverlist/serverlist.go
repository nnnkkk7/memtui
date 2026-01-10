// Package serverlist provides a Bubble Tea component for displaying and managing
// a list of Memcached servers with connection status.
package serverlist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ServerItem represents a single server in the list with its connection status
type ServerItem struct {
	Name      string // Human-readable server name
	Address   string // Server address in host:port format
	Status    string // Connection status message (e.g., "Connected", "Disconnected", "Error")
	Connected bool   // Whether currently connected to this server
	Default   bool   // Whether this is the default server
}

// Messages

// ServerSelectedMsg is sent when a server is selected from the list
type ServerSelectedMsg struct {
	Server ServerItem
}

// AddServerRequestMsg is sent when user requests to add a new server
type AddServerRequestMsg struct{}

// DeleteServerRequestMsg is sent when user requests to delete a server
type DeleteServerRequestMsg struct {
	Server ServerItem
}

// SetDefaultRequestMsg is sent when user requests to set a server as default
type SetDefaultRequestMsg struct {
	Server ServerItem
}

// CloseServerListMsg is sent when user requests to close the server list
type CloseServerListMsg struct{}

// Model represents the server list component
type Model struct {
	servers []ServerItem
	cursor  int
	width   int
	height  int
	focused bool

	// Styles
	normalStyle    lipgloss.Style
	selectedStyle  lipgloss.Style
	connectedStyle lipgloss.Style
	defaultStyle   lipgloss.Style
	titleStyle     lipgloss.Style
	helpStyle      lipgloss.Style
}

// New creates a new server list model with the given servers
func New(servers []ServerItem) *Model {
	if servers == nil {
		servers = []ServerItem{}
	}

	return &Model{
		servers: servers,
		cursor:  0,
		focused: true,
		normalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true),
		connectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")), // Green
		defaultStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")), // Orange
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1),
	}
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

// handleKeyMsg processes keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.cursor < len(m.servers)-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyEnter:
		if len(m.servers) == 0 {
			return m, nil
		}
		selected := m.servers[m.cursor]
		return m, func() tea.Msg {
			return ServerSelectedMsg{Server: selected}
		}

	case tea.KeyEscape:
		return m, func() tea.Msg {
			return CloseServerListMsg{}
		}

	case tea.KeyRunes:
		return m.handleRuneKey(msg)
	}

	return m, nil
}

// handleRuneKey processes single character key presses
func (m *Model) handleRuneKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if len(msg.Runes) == 0 {
		return m, nil
	}

	switch msg.Runes[0] {
	case 'j': // vim-style down
		if m.cursor < len(m.servers)-1 {
			m.cursor++
		}

	case 'k': // vim-style up
		if m.cursor > 0 {
			m.cursor--
		}

	case 'a': // add server
		return m, func() tea.Msg {
			return AddServerRequestMsg{}
		}

	case 'd': // delete server
		if len(m.servers) == 0 {
			return m, nil
		}
		selected := m.servers[m.cursor]
		return m, func() tea.Msg {
			return DeleteServerRequestMsg{Server: selected}
		}

	case 's': // set default
		if len(m.servers) == 0 {
			return m, nil
		}
		selected := m.servers[m.cursor]
		return m, func() tea.Msg {
			return SetDefaultRequestMsg{Server: selected}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *Model) View() string {
	if len(m.servers) == 0 {
		return m.renderEmpty()
	}

	var b strings.Builder

	// Title
	b.WriteString(m.titleStyle.Render("Servers"))
	b.WriteString("\n")

	// Server list
	for i, server := range m.servers {
		line := m.renderServerItem(server, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Help text
	b.WriteString(m.helpStyle.Render("Enter: select | a: add | d: delete | s: set default | Esc: close"))

	return b.String()
}

// renderEmpty renders the empty state message
func (m *Model) renderEmpty() string {
	var b strings.Builder

	b.WriteString(m.titleStyle.Render("Servers"))
	b.WriteString("\n")
	b.WriteString(m.normalStyle.Render("No servers configured"))
	b.WriteString("\n\n")
	b.WriteString(m.helpStyle.Render("a: add server | Esc: close"))

	return b.String()
}

// renderServerItem renders a single server item
func (m *Model) renderServerItem(server ServerItem, isSelected bool) string {
	// Build indicators
	var indicators []string

	if server.Connected {
		indicators = append(indicators, m.connectedStyle.Render("*"))
	}
	if server.Default {
		indicators = append(indicators, m.defaultStyle.Render("[default]"))
	}

	// Build main content
	name := server.Name
	address := server.Address
	status := server.Status

	// Format: name (address) - status [indicators]
	line := fmt.Sprintf("  %s (%s)", name, address)
	if status != "" {
		line += fmt.Sprintf(" - %s", status)
	}
	if len(indicators) > 0 {
		line += " " + strings.Join(indicators, " ")
	}

	// Apply selection style if needed
	if isSelected {
		// Use cursor indicator
		line = "> " + line[2:] // Replace leading spaces with cursor
		return m.selectedStyle.Render(line)
	}

	// Apply connected style if connected but not selected
	if server.Connected {
		return m.connectedStyle.Render(line)
	}

	return m.normalStyle.Render(line)
}

// SetServers updates the server list
func (m *Model) SetServers(servers []ServerItem) {
	if servers == nil {
		servers = []ServerItem{}
	}
	m.servers = servers

	// Reset cursor if out of bounds
	if m.cursor >= len(m.servers) {
		m.cursor = 0
	}
}

// Selected returns the currently selected server, or nil if none
func (m *Model) Selected() *ServerItem {
	if len(m.servers) == 0 || m.cursor < 0 || m.cursor >= len(m.servers) {
		return nil
	}
	return &m.servers[m.cursor]
}

// SetSize sets the component dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Cursor returns the current cursor position
func (m *Model) Cursor() int {
	return m.cursor
}

// ServerCount returns the number of servers in the list
func (m *Model) ServerCount() int {
	return len(m.servers)
}

// SetFocused sets the focused state of the component
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the component is focused
func (m *Model) IsFocused() bool {
	return m.focused
}

// Servers returns the current list of servers
func (m *Model) Servers() []ServerItem {
	return m.servers
}
