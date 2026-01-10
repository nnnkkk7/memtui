// Package editor provides a value editor component for editing Memcached values.
// It supports multi-line text editing with save/cancel functionality.
package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EditorMode represents the editor display/editing mode.
type EditorMode int

const (
	// ModeText is plain text editing mode.
	ModeText EditorMode = iota
	// ModeJSON is JSON editing mode with formatting support.
	ModeJSON
)

// String returns the string representation of the editor mode.
func (m EditorMode) String() string {
	switch m {
	case ModeText:
		return "Text"
	case ModeJSON:
		return "JSON"
	default:
		return "Unknown"
	}
}

// EditorSaveMsg is returned when the user saves the edited content.
type EditorSaveMsg struct {
	Key         string
	Value       []byte
	OriginalCAS uint64
}

// EditorCancelMsg is returned when the user cancels editing.
type EditorCancelMsg struct{}

// Editor is a component for editing Memcached values.
type Editor struct {
	key           string
	originalValue []byte
	cas           uint64
	textarea      textarea.Model
	mode          EditorMode
	dirty         bool
	width         int
	height        int

	// Track initial content for dirty detection
	initialContent string

	// Styles
	headerStyle   lipgloss.Style
	metaStyle     lipgloss.Style
	hintStyle     lipgloss.Style
	modifiedStyle lipgloss.Style
	borderStyle   lipgloss.Style
}

// New creates a new Editor with the given key and initial value.
func New(key string, value []byte) *Editor {
	ta := textarea.New()
	ta.SetValue(string(value))
	ta.Focus()
	ta.CharLimit = 0 // No character limit
	ta.ShowLineNumbers = true

	// Set a reasonable default size
	ta.SetWidth(60)
	ta.SetHeight(15)

	return &Editor{
		key:            key,
		originalValue:  value,
		textarea:       ta,
		mode:           ModeText,
		dirty:          false,
		initialContent: string(value),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#528BFF")),
		metaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")),
		hintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			Italic(true),
		modifiedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Bold(true),
		borderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#528BFF")),
	}
}

// Key returns the key being edited.
func (e *Editor) Key() string {
	return e.key
}

// OriginalValue returns the original value before any edits.
func (e *Editor) OriginalValue() []byte {
	return e.originalValue
}

// Value returns the current textarea content.
func (e *Editor) Value() string {
	return e.textarea.Value()
}

// SetContent sets the textarea content.
func (e *Editor) SetContent(content []byte) {
	e.textarea.SetValue(string(content))
	e.checkDirty()
}

// SetCAS sets the CAS value for optimistic locking.
func (e *Editor) SetCAS(cas uint64) {
	e.cas = cas
}

// Mode returns the current editor mode.
func (e *Editor) Mode() EditorMode {
	return e.mode
}

// SetMode sets the editor mode.
func (e *Editor) SetMode(mode EditorMode) {
	e.mode = mode
}

// IsDirty returns true if the content has been modified since loading.
func (e *Editor) IsDirty() bool {
	return e.dirty
}

// SetSize sets the editor dimensions.
func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height

	// Reserve space for header (3 lines) and hints (2 lines)
	taWidth := width - 4 // Account for border
	taHeight := height - 7

	if taWidth < 20 {
		taWidth = 20
	}
	if taHeight < 5 {
		taHeight = 5
	}

	e.textarea.SetWidth(taWidth)
	e.textarea.SetHeight(taHeight)
}

// FormatJSON formats the current content as indented JSON.
func (e *Editor) FormatJSON() error {
	content := e.textarea.Value()

	var buf bytes.Buffer
	err := json.Indent(&buf, []byte(content), "", "  ")
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	e.textarea.SetValue(buf.String())
	e.checkDirty()
	return nil
}

// checkDirty updates the dirty flag based on current content.
func (e *Editor) checkDirty() {
	e.dirty = e.textarea.Value() != e.initialContent
}

// Init initializes the editor and returns the initial command.
func (e *Editor) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages and returns the updated model and command.
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlS:
			// Save
			return e, e.save()
		case tea.KeyEsc:
			// Cancel
			return e, e.cancel()
		case tea.KeyCtrlF:
			// Format JSON (only in JSON mode)
			if e.mode == ModeJSON {
				_ = e.FormatJSON()
			}
			return e, nil
		}
	}

	// Pass messages to textarea
	var cmd tea.Cmd
	e.textarea, cmd = e.textarea.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Check if content changed
	e.checkDirty()

	return e, tea.Batch(cmds...)
}

// save creates a command that returns EditorSaveMsg.
func (e *Editor) save() tea.Cmd {
	return func() tea.Msg {
		return EditorSaveMsg{
			Key:         e.key,
			Value:       []byte(e.textarea.Value()),
			OriginalCAS: e.cas,
		}
	}
}

// cancel creates a command that returns EditorCancelMsg.
func (e *Editor) cancel() tea.Cmd {
	return func() tea.Msg {
		return EditorCancelMsg{}
	}
}

// View renders the editor.
func (e *Editor) View() string {
	if e.width == 0 || e.height == 0 {
		return "No size set"
	}

	var b strings.Builder

	// Header with key name
	header := e.headerStyle.Render(fmt.Sprintf("Editing: %s", e.key))
	b.WriteString(header)

	// Modified indicator
	if e.dirty {
		b.WriteString(" ")
		b.WriteString(e.modifiedStyle.Render("[Modified]"))
	}
	b.WriteString("\n")

	// Metadata line
	contentSize := len(e.textarea.Value())
	meta := fmt.Sprintf("Size: %d bytes | Mode: %s", contentSize, e.mode.String())
	b.WriteString(e.metaStyle.Render(meta))
	b.WriteString("\n")

	// Separator
	sepWidth := e.width - 4
	if sepWidth > 60 {
		sepWidth = 60
	}
	if sepWidth < 20 {
		sepWidth = 20
	}
	b.WriteString(strings.Repeat("─", sepWidth))
	b.WriteString("\n")

	// Textarea
	b.WriteString(e.textarea.View())
	b.WriteString("\n")

	// Hints
	hints := []string{
		"Ctrl+S: Save",
		"Esc: Cancel",
	}
	if e.mode == ModeJSON {
		hints = append(hints, "Ctrl+F: Format JSON")
	}
	b.WriteString(e.hintStyle.Render(strings.Join(hints, " | ")))

	return b.String()
}
