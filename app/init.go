package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/config"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/command"
	"github.com/nnnkkk7/memtui/ui/components/help"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
	"github.com/nnnkkk7/memtui/ui/components/viewer"
	"github.com/nnnkkk7/memtui/ui/styles"
)

// NewModel creates a new app model with default settings
func NewModel(addr string) *Model {
	return NewModelWithConfig(addr, nil)
}

// NewModelWithConfig creates a new app model with the given configuration
func NewModelWithConfig(addr string, cfg *config.Config) *Model {
	// Use defaults if no config provided
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Select theme based on config
	theme := &styles.DarkTheme
	if cfg.UI.Theme == "light" {
		theme = &styles.LightTheme
	}

	// Create keylist with delimiter from config
	kl := keylist.NewModel()
	kl.SetDelimiter(cfg.UI.KeyDelimiter)

	return &Model{
		addr:           addr,
		state:          StateConnecting,
		styles:         NewStylesFromTheme(theme),
		keyList:        kl,
		viewer:         viewer.NewModel(),
		commandPalette: command.New(command.DefaultCommands()),
		help:           help.NewModel(),
		focus:          FocusKeyList,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return m.connectCmd()
}

// State returns the current state
func (m *Model) State() State {
	return m.state
}

// Width returns the terminal width
func (m *Model) Width() int {
	return m.width
}

// Height returns the terminal height
func (m *Model) Height() int {
	return m.height
}

// Keys returns the loaded keys
func (m *Model) Keys() []models.KeyInfo {
	return m.keys
}

// Error returns the error message
func (m *Model) Error() string {
	return m.err
}

// Focus returns the current focus mode
func (m *Model) Focus() FocusMode {
	return m.focus
}

// SetFocus sets the focus mode (for testing)
func (m *Model) SetFocus(focus FocusMode) {
	m.focus = focus
}

// updateComponentSizes updates all component sizes based on terminal dimensions
func (m *Model) updateComponentSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Calculate layout: key list on left (30%), viewer on right (70%)
	keyListWidth := m.width * 30 / 100
	viewerWidth := m.width - keyListWidth - 4

	contentHeight := m.height - 4 // Reserve for status bar and padding

	m.keyList.SetSize(keyListWidth-2, contentHeight-2)
	m.viewer.SetSize(viewerWidth-2, contentHeight-2)

	if m.commandPalette != nil {
		m.commandPalette.SetSize(m.width, m.height)
	}
	if m.help != nil {
		m.help.SetSize(m.width, m.height)
	}
}
