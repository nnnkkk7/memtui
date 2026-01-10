// Package statusbar provides a status bar component for displaying connection status.
package statusbar

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Status represents the connection status
type Status int

// Status bar states
const (
	// StatusConnecting indicates the connection is in progress
	StatusConnecting Status = iota
	// StatusConnected indicates successful connection
	StatusConnected
	// StatusLoading indicates data is loading
	StatusLoading
	// StatusReady indicates the app is ready
	StatusReady
	// StatusError indicates an error state
	StatusError
)

// String returns the string representation of the status
func (s Status) String() string {
	switch s {
	case StatusConnecting:
		return "Connecting"
	case StatusConnected:
		return "Connected"
	case StatusLoading:
		return "Loading"
	case StatusReady:
		return "Ready"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Model represents the status bar component
type Model struct {
	address  string
	version  string
	keyCount int
	status   Status
	error    string
	width    int

	// Styles
	style       lipgloss.Style
	errorStyle  lipgloss.Style
	statusStyle lipgloss.Style
}

// NewModel creates a new status bar model
func NewModel() *Model {
	return &Model{
		style: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
	}
}

// SetAddress sets the server address
func (m *Model) SetAddress(addr string) {
	m.address = addr
}

// SetVersion sets the server version
func (m *Model) SetVersion(version string) {
	m.version = version
}

// SetKeyCount sets the number of keys
func (m *Model) SetKeyCount(count int) {
	m.keyCount = count
}

// SetStatus sets the current status
func (m *Model) SetStatus(status Status) {
	m.status = status
}

// SetError sets the error message
func (m *Model) SetError(err string) {
	m.error = err
	m.status = StatusError
}

// SetWidth sets the width of the status bar
func (m *Model) SetWidth(width int) {
	m.width = width
}

// View renders the status bar
func (m *Model) View() string {
	var content string

	// Status indicator
	statusText := m.statusStyle.Render(m.status.String())

	if m.status == StatusError && m.error != "" {
		content = fmt.Sprintf(" %s | %s ", statusText, m.errorStyle.Render(m.error))
	} else {
		// Normal status bar
		parts := []string{statusText}

		if m.address != "" {
			parts = append(parts, m.address)
		}

		if m.version != "" {
			parts = append(parts, fmt.Sprintf("v%s", m.version))
		}

		if m.keyCount > 0 || m.status == StatusReady {
			parts = append(parts, fmt.Sprintf("%d keys", m.keyCount))
		}

		content = " "
		for i, part := range parts {
			if i > 0 {
				content += " | "
			}
			content += part
		}
		content += " "
	}

	// Apply width if set
	if m.width > 0 {
		m.style = m.style.Width(m.width)
	}

	return m.style.Render(content)
}
