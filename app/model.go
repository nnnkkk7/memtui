package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/command"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
	"github.com/nnnkkk7/memtui/ui/components/editor"
	"github.com/nnnkkk7/memtui/ui/components/help"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
	"github.com/nnnkkk7/memtui/ui/components/viewer"
	"github.com/nnnkkk7/memtui/ui/styles"
)

// State represents the application state
type State int

// Application states
const (
	// StateConnecting indicates the app is connecting to the server
	StateConnecting State = iota
	// StateConnected indicates successful connection
	StateConnected
	// StateLoading indicates data is being loaded
	StateLoading
	// StateReady indicates the app is ready for use
	StateReady
	// StateError indicates an error occurred
	StateError
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateLoading:
		return "Loading"
	case StateReady:
		return "Ready"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// FocusMode represents which component has focus
type FocusMode int

// Focus modes for UI components
const (
	// FocusKeyList indicates the key list has focus
	FocusKeyList FocusMode = iota
	// FocusViewer indicates the value viewer has focus
	FocusViewer
	// FocusDialog indicates a dialog has focus
	FocusDialog
	// FocusEditor indicates the editor has focus
	FocusEditor
	// FocusCommandPalette indicates the command palette has focus
	FocusCommandPalette
	// FocusHelp indicates the help view has focus
	FocusHelp
	// FocusFilter indicates the filter input has focus
	FocusFilter
)

// Model is the main Bubble Tea model
type Model struct {
	addr    string
	state   State
	width   int
	height  int
	keys    []models.KeyInfo
	err     string
	version string
	focus   FocusMode

	// Memcached client (unified interface for all operations including CAS)
	mcClient client.MemcachedClient

	// Components
	keyList        *keylist.Model
	viewer         *viewer.Model
	commandPalette *command.CommandPalette
	confirmDialog  *dialog.ConfirmDialog
	inputDialog    *dialog.InputDialog
	editor         *editor.Editor
	help           *help.Model

	// Filter mode
	filterInput string
	filtering   bool

	// Current key being viewed/edited
	currentKey     *models.KeyInfo
	currentValue   []byte
	currentCASItem *client.CASItem

	// Server capabilities
	supportsMetadump bool

	// Styles
	styles Styles
}

// Styles holds lipgloss styles for the app, derived from a Theme
type Styles struct {
	Theme      styles.Theme
	Title      lipgloss.Style
	StatusBar  lipgloss.Style
	Error      lipgloss.Style
	KeyList    lipgloss.Style
	Viewer     lipgloss.Style
	Help       lipgloss.Style
	Connecting lipgloss.Style
	Border     lipgloss.Style
}

// DefaultStyles returns the default styles using the dark theme
func DefaultStyles() Styles {
	return NewStylesFromTheme(&styles.DarkTheme)
}

// NewStylesFromTheme creates Styles from a Theme
func NewStylesFromTheme(theme *styles.Theme) Styles {
	return Styles{
		Theme: *theme,
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.BorderFocus),
		StatusBar: lipgloss.NewStyle().
			Background(theme.Surface).
			Foreground(theme.TextMuted).
			Padding(0, 1),
		Error: lipgloss.NewStyle().
			Foreground(theme.Error).
			Bold(true),
		KeyList: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderFocus),
		Viewer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Secondary),
		Help: lipgloss.NewStyle().
			Foreground(theme.TextMuted),
		Connecting: lipgloss.NewStyle().
			Foreground(theme.Warning),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border),
	}
}
