// Package dialog provides dialog components for user interactions.
package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputResultMsg is the message returned when the input dialog is closed.
// Value contains the entered text, Canceled is true if the user pressed Escape.
type InputResultMsg struct {
	Value    string
	Canceled bool
	Context  interface{}
}

// ValidatorFunc is a function that validates input and returns an error if invalid.
type ValidatorFunc func(string) error

// InputDialog is an input dialog component for text entry.
type InputDialog struct {
	title           string
	textInput       textinput.Model
	validator       ValidatorFunc
	context         interface{}
	validationError string
	width           int
	height          int

	// Styles
	overlayStyle lipgloss.Style
	titleStyle   lipgloss.Style
	inputStyle   lipgloss.Style
	errorStyle   lipgloss.Style
	hintStyle    lipgloss.Style
}

// NewInput creates a new input dialog with the given title.
func NewInput(title string) *InputDialog {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	return &InputDialog{
		title:     title,
		textInput: ti,
		overlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#528BFF")).
			Padding(1, 2).
			Background(lipgloss.Color("#282C34")),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#528BFF")).
			MarginBottom(1),
		inputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			MarginTop(1),
		hintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			MarginTop(1).
			Italic(true),
	}
}

// Title returns the dialog title.
func (d *InputDialog) Title() string {
	return d.title
}

// Value returns the current input value.
func (d *InputDialog) Value() string {
	return d.textInput.Value()
}

// ValidationError returns the current validation error message, if any.
func (d *InputDialog) ValidationError() string {
	return d.validationError
}

// WithPlaceholder sets the placeholder text for the input field.
func (d *InputDialog) WithPlaceholder(placeholder string) *InputDialog {
	d.textInput.Placeholder = placeholder
	return d
}

// WithValue sets the initial value of the input field.
func (d *InputDialog) WithValue(value string) *InputDialog {
	d.textInput.SetValue(value)
	return d
}

// WithValidator sets a validation function for the input.
func (d *InputDialog) WithValidator(validator ValidatorFunc) *InputDialog {
	d.validator = validator
	return d
}

// WithContext sets context data that will be returned in the result message.
func (d *InputDialog) WithContext(context interface{}) *InputDialog {
	d.context = context
	return d
}

// SetSize sets the dialog size for layout calculations.
func (d *InputDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	// Adjust text input width based on dialog width
	if width > 0 {
		inputWidth := width - 10 // Account for padding and borders
		if inputWidth > 60 {
			inputWidth = 60
		}
		if inputWidth < 20 {
			inputWidth = 20
		}
		d.textInput.Width = inputWidth
	}
}

// Init implements tea.Model. It initializes the text input and returns its blink command.
func (d *InputDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model. It handles keyboard input for the dialog.
func (d *InputDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return d, d.submit()
		case tea.KeyEsc:
			return d, d.cancel()
		}
	}

	// Pass other messages to the text input
	var cmd tea.Cmd
	d.textInput, cmd = d.textInput.Update(msg)

	// Clear validation error when user types
	d.validationError = ""

	return d, cmd
}

// submit validates and submits the input value.
func (d *InputDialog) submit() tea.Cmd {
	value := d.textInput.Value()

	// Run validation if validator is set
	if d.validator != nil {
		if err := d.validator(value); err != nil {
			d.validationError = err.Error()
			return nil // Don't submit, show error
		}
	}

	return func() tea.Msg {
		return InputResultMsg{
			Value:    value,
			Canceled: false,
			Context:  d.context,
		}
	}
}

// cancel cancels the dialog and returns an empty result.
func (d *InputDialog) cancel() tea.Cmd {
	return func() tea.Msg {
		return InputResultMsg{
			Value:    "",
			Canceled: true,
			Context:  d.context,
		}
	}
}

// View implements tea.Model. It renders the dialog.
func (d *InputDialog) View() string {
	var b strings.Builder

	// Title
	b.WriteString(d.titleStyle.Render(d.title))
	b.WriteString("\n\n")

	// Text input
	b.WriteString(d.textInput.View())
	b.WriteString("\n")

	// Validation error
	if d.validationError != "" {
		b.WriteString(d.errorStyle.Render(d.validationError))
		b.WriteString("\n")
	}

	// Hint
	b.WriteString(d.hintStyle.Render("Enter: submit  Esc: cancel"))

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
