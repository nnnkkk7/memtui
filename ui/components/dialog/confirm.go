// Package dialog provides dialog components for user interactions.
package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmResultMsg is the message returned when the dialog is closed.
// Result is true if confirmed, false if canceled.
type ConfirmResultMsg struct {
	Result  bool
	Context interface{}
}

// ConfirmDialog is a confirmation dialog component for destructive operations.
type ConfirmDialog struct {
	title      string
	message    string
	focusedYes bool
	context    interface{}
	width      int
	height     int

	// Styles
	overlayStyle       lipgloss.Style
	titleStyle         lipgloss.Style
	messageStyle       lipgloss.Style
	buttonStyle        lipgloss.Style
	buttonFocusedStyle lipgloss.Style
	hintStyle          lipgloss.Style
}

// New creates a new confirmation dialog with the given title and message.
func New(title, message string) *ConfirmDialog {
	return NewWithContext(title, message, nil)
}

// NewWithContext creates a new confirmation dialog with context data.
// The context will be returned in the ConfirmResultMsg.
func NewWithContext(title, message string, context interface{}) *ConfirmDialog {
	return &ConfirmDialog{
		title:      title,
		message:    message,
		focusedYes: false, // Default to No for safety
		context:    context,
		overlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#E06C75")).
			Padding(1, 2).
			Background(lipgloss.Color("#282C34")),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E06C75")).
			MarginBottom(1),
		messageStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF")).
			MarginBottom(1),
		buttonStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF")).
			Background(lipgloss.Color("#3E4451")).
			Padding(0, 2).
			MarginRight(1),
		buttonFocusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#528BFF")).
			Bold(true).
			Padding(0, 2).
			MarginRight(1),
		hintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			MarginTop(1).
			Italic(true),
	}
}

// Title returns the dialog title.
func (d *ConfirmDialog) Title() string {
	return d.title
}

// Message returns the dialog message.
func (d *ConfirmDialog) Message() string {
	return d.message
}

// FocusedOnYes returns true if Yes button is focused.
func (d *ConfirmDialog) FocusedOnYes() bool {
	return d.focusedYes
}

// SetSize sets the dialog size for layout calculations.
func (d *ConfirmDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// Init implements tea.Model.
func (d *ConfirmDialog) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (d *ConfirmDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			d.focusedYes = !d.focusedYes
		case tea.KeyLeft:
			d.focusedYes = true
		case tea.KeyRight:
			d.focusedYes = false
		case tea.KeyEnter:
			return d, d.confirm(d.focusedYes)
		case tea.KeyEsc:
			return d, d.confirm(false)
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'y', 'Y':
					return d, d.confirm(true)
				case 'n', 'N':
					return d, d.confirm(false)
				}
			}
		}
	}
	return d, nil
}

// confirm creates a command that returns a ConfirmResultMsg.
func (d *ConfirmDialog) confirm(result bool) tea.Cmd {
	return func() tea.Msg {
		return ConfirmResultMsg{
			Result:  result,
			Context: d.context,
		}
	}
}

// View implements tea.Model.
func (d *ConfirmDialog) View() string {
	var b strings.Builder

	// Title
	b.WriteString(d.titleStyle.Render(d.title))
	b.WriteString("\n")

	// Message
	b.WriteString(d.messageStyle.Render(d.message))
	b.WriteString("\n\n")

	// Buttons
	var yesButton, noButton string
	if d.focusedYes {
		yesButton = d.buttonFocusedStyle.Render("[ Yes ]")
		noButton = d.buttonStyle.Render("[ No ]")
	} else {
		yesButton = d.buttonStyle.Render("[ Yes ]")
		noButton = d.buttonFocusedStyle.Render("[ No ]")
	}
	b.WriteString(yesButton + "  " + noButton)
	b.WriteString("\n")

	// Hint
	b.WriteString(d.hintStyle.Render("y/n: quick select  Tab: switch  Enter: confirm  Esc: cancel"))

	content := b.String()

	// Apply overlay style
	if d.width > 0 {
		overlayWidth := 50
		if overlayWidth > d.width-4 {
			overlayWidth = d.width - 4
		}
		d.overlayStyle = d.overlayStyle.Width(overlayWidth)
	}

	return d.overlayStyle.Render(content)
}
