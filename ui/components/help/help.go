package help

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Category represents the category of a keybinding
type Category string

const (
	CategoryGlobal  Category = "Global"
	CategoryKeyList Category = "Key List"
	CategoryViewer  Category = "Viewer"
)

// KeyBinding represents a single keybinding entry
type KeyBinding struct {
	Key      string
	Action   string
	Category Category
}

// Model represents the help overlay component
type Model struct {
	visible  bool
	width    int
	height   int
	bindings []KeyBinding

	// Styles
	overlayStyle lipgloss.Style
	titleStyle   lipgloss.Style
	sectionStyle lipgloss.Style
	keyStyle     lipgloss.Style
	actionStyle  lipgloss.Style
	footerStyle  lipgloss.Style
}

// NewModel creates a new help overlay model
func NewModel() *Model {
	m := &Model{
		visible: false,
		overlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Background(lipgloss.Color("235")),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")).
			MarginBottom(1),
		sectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginTop(1),
		keyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Width(15),
		actionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		footerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1).
			Italic(true),
	}
	m.bindings = m.initKeyBindings()
	return m
}

// initKeyBindings initializes the keybinding list based on design document Appendix A
func (m *Model) initKeyBindings() []KeyBinding {
	return []KeyBinding{
		// Global keybindings
		{Key: "q, Ctrl+C", Action: "Quit", Category: CategoryGlobal},
		{Key: "?", Action: "Toggle Help", Category: CategoryGlobal},
		{Key: "Tab", Action: "Switch Pane", Category: CategoryGlobal},
		{Key: "r", Action: "Refresh", Category: CategoryGlobal},
		{Key: "s", Action: "Show Stats", Category: CategoryGlobal},

		// Key List pane keybindings
		{Key: "Up, k", Action: "Move Up", Category: CategoryKeyList},
		{Key: "Down, j", Action: "Move Down", Category: CategoryKeyList},
		{Key: "Enter, l", Action: "Select / Expand", Category: CategoryKeyList},
		{Key: "h", Action: "Collapse / Go to Parent", Category: CategoryKeyList},
		{Key: "/", Action: "Search Mode", Category: CategoryKeyList},
		{Key: "d", Action: "Delete Key", Category: CategoryKeyList},
		{Key: "n", Action: "Create New Key", Category: CategoryKeyList},
		{Key: "m", Action: "Load More", Category: CategoryKeyList},

		// Viewer pane keybindings
		{Key: "e", Action: "Edit Mode", Category: CategoryViewer},
		{Key: "J", Action: "JSON View", Category: CategoryViewer},
		{Key: "H", Action: "Hex View", Category: CategoryViewer},
		{Key: "T", Action: "Text View", Category: CategoryViewer},
		{Key: "A", Action: "Auto Detect", Category: CategoryViewer},
		{Key: "c", Action: "Copy Value", Category: CategoryViewer},
		{Key: "PageUp", Action: "Page Up", Category: CategoryViewer},
		{Key: "PageDown", Action: "Page Down", Category: CategoryViewer},
	}
}

// Visible returns whether the help overlay is visible
func (m *Model) Visible() bool {
	return m.visible
}

// Show shows the help overlay
func (m *Model) Show() {
	m.visible = true
}

// Hide hides the help overlay
func (m *Model) Hide() {
	m.visible = false
}

// Toggle toggles the help overlay visibility
func (m *Model) Toggle() {
	m.visible = !m.visible
}

// SetSize sets the overlay size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// KeyBindings returns all keybindings
func (m *Model) KeyBindings() []KeyBinding {
	return m.bindings
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.visible {
				m.Hide()
			}
		case tea.KeyRunes:
			if len(msg.Runes) > 0 && msg.Runes[0] == '?' {
				m.Toggle()
			}
		}
	}
	return m, nil
}

// View renders the help overlay
func (m *Model) View() string {
	if !m.visible {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(m.titleStyle.Render("Keybindings Help"))
	b.WriteString("\n")

	// Render bindings by category
	categories := []Category{CategoryGlobal, CategoryKeyList, CategoryViewer}

	for _, cat := range categories {
		b.WriteString(m.sectionStyle.Render(string(cat)))
		b.WriteString("\n")

		for _, binding := range m.bindings {
			if binding.Category == cat {
				key := m.keyStyle.Render(binding.Key)
				action := m.actionStyle.Render(binding.Action)
				b.WriteString(key + action + "\n")
			}
		}
	}

	// Footer with close hint
	b.WriteString(m.footerStyle.Render("Press ? or Esc to close"))

	content := b.String()

	// Apply overlay style with centering if size is set
	if m.width > 0 && m.height > 0 {
		// Calculate overlay dimensions
		overlayWidth := 50
		if overlayWidth > m.width-4 {
			overlayWidth = m.width - 4
		}

		m.overlayStyle = m.overlayStyle.Width(overlayWidth)
		return m.overlayStyle.Render(content)
	}

	return m.overlayStyle.Render(content)
}
