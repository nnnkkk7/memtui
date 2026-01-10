// Package styles provides theme and style definitions for the memtui application.
// It uses lipgloss for terminal styling and supports both dark and light themes.
package styles

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the application.
// It supports adaptive colors that adjust based on terminal background.
type Theme struct {
	Name string

	// Basic colors
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Success   lipgloss.AdaptiveColor
	Warning   lipgloss.AdaptiveColor
	Error     lipgloss.AdaptiveColor

	// Background colors
	Background lipgloss.AdaptiveColor
	Surface    lipgloss.AdaptiveColor

	// Text colors
	Text      lipgloss.AdaptiveColor
	TextMuted lipgloss.AdaptiveColor

	// Border colors
	Border      lipgloss.AdaptiveColor
	BorderFocus lipgloss.AdaptiveColor
}

// DarkTheme is the default dark color theme.
var DarkTheme = Theme{
	Name:        "dark",
	Primary:     lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"},
	Secondary:   lipgloss.AdaptiveColor{Light: "#56B6C2", Dark: "#56B6C2"},
	Success:     lipgloss.AdaptiveColor{Light: "#98C379", Dark: "#98C379"},
	Warning:     lipgloss.AdaptiveColor{Light: "#E5C07B", Dark: "#E5C07B"},
	Error:       lipgloss.AdaptiveColor{Light: "#E06C75", Dark: "#E06C75"},
	Background:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282C34"},
	Surface:     lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#3E4451"},
	Text:        lipgloss.AdaptiveColor{Light: "#24292E", Dark: "#ABB2BF"},
	TextMuted:   lipgloss.AdaptiveColor{Light: "#6A737D", Dark: "#5C6370"},
	Border:      lipgloss.AdaptiveColor{Light: "#E1E4E8", Dark: "#4B5263"},
	BorderFocus: lipgloss.AdaptiveColor{Light: "#0366D6", Dark: "#528BFF"},
}

// LightTheme is the light color theme.
var LightTheme = Theme{
	Name:        "light",
	Primary:     lipgloss.AdaptiveColor{Light: "#5B4FCF", Dark: "#5B4FCF"},
	Secondary:   lipgloss.AdaptiveColor{Light: "#0C969B", Dark: "#0C969B"},
	Success:     lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#2E7D32"},
	Warning:     lipgloss.AdaptiveColor{Light: "#F57C00", Dark: "#F57C00"},
	Error:       lipgloss.AdaptiveColor{Light: "#D32F2F", Dark: "#D32F2F"},
	Background:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"},
	Surface:     lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#F5F5F5"},
	Text:        lipgloss.AdaptiveColor{Light: "#24292E", Dark: "#24292E"},
	TextMuted:   lipgloss.AdaptiveColor{Light: "#6A737D", Dark: "#6A737D"},
	Border:      lipgloss.AdaptiveColor{Light: "#E1E4E8", Dark: "#E1E4E8"},
	BorderFocus: lipgloss.AdaptiveColor{Light: "#0366D6", Dark: "#0366D6"},
}

// DefaultTheme returns the default theme (dark).
func DefaultTheme() Theme {
	return DarkTheme
}

// GetTheme returns a theme by name. Returns dark theme for unknown names.
func GetTheme(name string) Theme {
	switch name {
	case "light":
		return LightTheme
	case "dark":
		return DarkTheme
	default:
		return DarkTheme
	}
}

// KeyListStyles contains styles for the key list component.
type KeyListStyles struct {
	Normal   lipgloss.Style
	Selected lipgloss.Style
	Folder   lipgloss.Style
	Leaf     lipgloss.Style
}

// ViewerStyles contains styles for the viewer component.
type ViewerStyles struct {
	Header  lipgloss.Style
	Content lipgloss.Style
	Meta    lipgloss.Style
}

// HelpStyles contains styles for the help overlay component.
type HelpStyles struct {
	Overlay lipgloss.Style
	Title   lipgloss.Style
	Section lipgloss.Style
	Key     lipgloss.Style
	Action  lipgloss.Style
	Footer  lipgloss.Style
}

// PanelStyles contains styles for panel borders.
type PanelStyles struct {
	Normal  lipgloss.Style
	Focused lipgloss.Style
}

// Styles contains all application styles.
type Styles struct {
	Theme Theme

	// Component styles
	Title     lipgloss.Style
	StatusBar lipgloss.Style
	Error     lipgloss.Style
	Cursor    lipgloss.Style

	// Grouped component styles
	KeyList KeyListStyles
	Viewer  ViewerStyles
	Help    HelpStyles
	Panel   PanelStyles
}

// NewStyles creates a new Styles instance with the given theme.
func NewStyles(theme Theme) *Styles {
	return &Styles{
		Theme: theme,

		// Title style - bold with primary color
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary),

		// StatusBar style
		StatusBar: lipgloss.NewStyle().
			Background(theme.Surface).
			Foreground(theme.TextMuted).
			Padding(0, 1),

		// Error style
		Error: lipgloss.NewStyle().
			Foreground(theme.Error).
			Bold(true),

		// Cursor style
		Cursor: lipgloss.NewStyle().
			Background(theme.Primary).
			Foreground(lipgloss.Color("#FFFFFF")),

		// KeyList styles
		KeyList: KeyListStyles{
			Normal: lipgloss.NewStyle().
				Foreground(theme.Text),
			Selected: lipgloss.NewStyle().
				Background(theme.Primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true),
			Folder: lipgloss.NewStyle().
				Foreground(theme.Warning),
			Leaf: lipgloss.NewStyle().
				Foreground(theme.Text),
		},

		// Viewer styles
		Viewer: ViewerStyles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Secondary),
			Content: lipgloss.NewStyle().
				Foreground(theme.Text),
			Meta: lipgloss.NewStyle().
				Foreground(theme.TextMuted),
		},

		// Help styles
		Help: HelpStyles{
			Overlay: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.BorderFocus).
				Padding(1, 2).
				Background(theme.Surface),
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Warning).
				MarginBottom(1),
			Section: lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Secondary).
				MarginTop(1),
			Key: lipgloss.NewStyle().
				Foreground(theme.Warning).
				Width(15),
			Action: lipgloss.NewStyle().
				Foreground(theme.Text),
			Footer: lipgloss.NewStyle().
				Foreground(theme.TextMuted).
				MarginTop(1).
				Italic(true),
		},

		// Panel styles
		Panel: PanelStyles{
			Normal: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Border),
			Focused: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.BorderFocus),
		},
	}
}

// DefaultStyles returns styles with the default theme.
func DefaultStyles() *Styles {
	return NewStyles(DefaultTheme())
}
