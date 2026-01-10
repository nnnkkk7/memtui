// Package command provides a command palette component for the memtui application.
// It implements a VS Code-style command palette with fuzzy search functionality.
package command

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents a single command that can be executed from the palette.
type Command struct {
	Name        string         // Display name of the command
	Description string         // Brief description of what the command does
	Shortcut    string         // Keyboard shortcut (e.g., "r", "Ctrl+S")
	Action      func() tea.Msg // Function to execute when command is selected
}

// CommandExecuteMsg is sent when a command is executed from the palette.
type CommandExecuteMsg struct {
	Command Command
}

// CommandCancelMsg is sent when the command palette is canceled (Escape).
type CommandCancelMsg struct{}

// CommandPalette is a VS Code-style command palette component.
type CommandPalette struct {
	input    textinput.Model
	commands []Command
	filtered []Command
	selected int
	visible  bool
	width    int
	height   int

	// Styles
	overlayStyle     lipgloss.Style
	titleStyle       lipgloss.Style
	inputStyle       lipgloss.Style
	itemStyle        lipgloss.Style
	selectedStyle    lipgloss.Style
	shortcutStyle    lipgloss.Style
	descriptionStyle lipgloss.Style
	emptyStyle       lipgloss.Style
	hintStyle        lipgloss.Style
}

// New creates a new CommandPalette with the given commands.
func New(commands []Command) *CommandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type to search commands..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	palette := &CommandPalette{
		input:    ti,
		commands: commands,
		filtered: make([]Command, len(commands)),
		selected: 0,
		visible:  false, // Hidden by default, shown with Ctrl+P

		overlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#528BFF")).
			Padding(1, 2).
			Background(lipgloss.Color("#282C34")),

		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E5C07B")).
			MarginBottom(1),

		inputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF")),

		itemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF")).
			PaddingLeft(1),

		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#528BFF")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			PaddingLeft(1),

		shortcutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			Width(8).
			Align(lipgloss.Right),

		descriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			Italic(true),

		emptyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			Italic(true).
			PaddingLeft(1),

		hintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6370")).
			MarginTop(1).
			Italic(true),
	}

	// Copy commands to filtered list
	copy(palette.filtered, commands)

	return palette
}

// DefaultCommands returns the standard set of commands for memtui.
func DefaultCommands() []Command {
	return []Command{
		{
			Name:        "Refresh keys",
			Description: "Reload the key list from Memcached",
			Shortcut:    "r",
			Action:      func() tea.Msg { return RefreshKeysMsg{} },
		},
		{
			Name:        "Delete key",
			Description: "Delete the selected key",
			Shortcut:    "d",
			Action:      func() tea.Msg { return DeleteKeyMsg{} },
		},
		{
			Name:        "New key",
			Description: "Create a new key-value pair",
			Shortcut:    "n",
			Action:      func() tea.Msg { return NewKeyMsg{} },
		},
		{
			Name:        "Edit value",
			Description: "Edit the selected key's value",
			Shortcut:    "e",
			Action:      func() tea.Msg { return EditValueMsg{} },
		},
		{
			Name:        "Show stats",
			Description: "Display Memcached server statistics",
			Shortcut:    "s",
			Action:      func() tea.Msg { return ShowStatsMsg{} },
		},
		{
			Name:        "Toggle theme",
			Description: "Switch between dark and light themes",
			Shortcut:    "",
			Action:      func() tea.Msg { return ToggleThemeMsg{} },
		},
		{
			Name:        "Show help",
			Description: "Display keyboard shortcuts help",
			Shortcut:    "?",
			Action:      func() tea.Msg { return ShowHelpMsg{} },
		},
		{
			Name:        "Quit",
			Description: "Exit the application",
			Shortcut:    "q",
			Action:      func() tea.Msg { return QuitMsg{} },
		},
		{
			Name:        "Filter keys",
			Description: "Enter key filter/search mode",
			Shortcut:    "/",
			Action:      func() tea.Msg { return FilterKeysMsg{} },
		},
		{
			Name:        "Copy value",
			Description: "Copy the selected value to clipboard",
			Shortcut:    "c",
			Action:      func() tea.Msg { return CopyValueMsg{} },
		},
	}
}

// RefreshKeysMsg requests the key list to be refreshed.
type RefreshKeysMsg struct{}

// DeleteKeyMsg requests deletion of the selected key.
type DeleteKeyMsg struct{}

// NewKeyMsg requests creation of a new key.
type NewKeyMsg struct{}

// EditValueMsg requests editing of the current value.
type EditValueMsg struct{}

// ShowStatsMsg requests display of server statistics.
type ShowStatsMsg struct{}

// ToggleThemeMsg requests toggling between light and dark themes.
type ToggleThemeMsg struct{}

// ShowHelpMsg requests display of the help screen.
type ShowHelpMsg struct{}

// QuitMsg requests application exit.
type QuitMsg struct{}

// FilterKeysMsg requests key filtering mode.
type FilterKeysMsg struct{}

// CopyValueMsg requests copying the current value to clipboard.
type CopyValueMsg struct{}

// Visible returns whether the command palette is currently visible.
func (p *CommandPalette) Visible() bool {
	return p.visible
}

// Show makes the command palette visible.
func (p *CommandPalette) Show() {
	p.visible = true
	p.input.Focus()
	p.input.SetValue("")
	p.filterCommands("")
	p.selected = 0
}

// Hide hides the command palette.
func (p *CommandPalette) Hide() {
	p.visible = false
	p.input.Blur()
}

// Toggle toggles the command palette visibility.
func (p *CommandPalette) Toggle() {
	if p.visible {
		p.Hide()
	} else {
		p.Show()
	}
}

// SetSize sets the dimensions for the command palette.
func (p *CommandPalette) SetSize(width, height int) {
	p.width = width
	p.height = height

	// Adjust input width based on palette width
	if width > 0 {
		inputWidth := width - 10
		if inputWidth > 50 {
			inputWidth = 50
		}
		if inputWidth < 20 {
			inputWidth = 20
		}
		p.input.Width = inputWidth
	}
}

// SelectedCommand returns the currently selected command, or nil if none.
func (p *CommandPalette) SelectedCommand() *Command {
	if len(p.filtered) == 0 || p.selected >= len(p.filtered) {
		return nil
	}
	return &p.filtered[p.selected]
}

// Init implements tea.Model.
func (p *CommandPalette) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (p *CommandPalette) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return p, p.execute()

		case tea.KeyEsc:
			return p, p.cancel()

		case tea.KeyUp:
			p.moveUp()
			return p, nil

		case tea.KeyDown:
			p.moveDown()
			return p, nil

		case tea.KeyRunes:
			// Handle vim-style navigation (but only if not typing in input)
			if len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'k':
					// Check if this is likely navigation vs typing
					if p.input.Value() == "" || !p.input.Focused() {
						p.moveUp()
						return p, nil
					}
				case 'j':
					if p.input.Value() == "" || !p.input.Focused() {
						p.moveDown()
						return p, nil
					}
				}
			}
		}
	}

	// Update the text input
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)

	// Update filtered commands based on input
	p.filterCommands(p.input.Value())

	return p, cmd
}

// moveUp moves the selection up with wrapping.
func (p *CommandPalette) moveUp() {
	if len(p.filtered) == 0 {
		return
	}
	p.selected--
	if p.selected < 0 {
		p.selected = len(p.filtered) - 1
	}
}

// moveDown moves the selection down with wrapping.
func (p *CommandPalette) moveDown() {
	if len(p.filtered) == 0 {
		return
	}
	p.selected++
	if p.selected >= len(p.filtered) {
		p.selected = 0
	}
}

// filterCommands filters the command list based on the query.
func (p *CommandPalette) filterCommands(query string) {
	if query == "" {
		p.filtered = make([]Command, len(p.commands))
		copy(p.filtered, p.commands)
		p.selected = 0
		return
	}

	// Use fuzzy matching to filter and rank commands
	p.filtered = RankCommands(p.commands, query)
	p.selected = 0
}

// execute executes the currently selected command.
func (p *CommandPalette) execute() tea.Cmd {
	cmd := p.SelectedCommand()
	if cmd == nil {
		return nil
	}

	p.Hide()

	return func() tea.Msg {
		return CommandExecuteMsg{Command: *cmd}
	}
}

// cancel cancels the command palette.
func (p *CommandPalette) cancel() tea.Cmd {
	p.Hide()

	return func() tea.Msg {
		return CommandCancelMsg{}
	}
}

// View implements tea.Model.
func (p *CommandPalette) View() string {
	if !p.visible {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(p.titleStyle.Render("Command Palette"))
	b.WriteString("\n\n")

	// Input field
	b.WriteString(p.input.View())
	b.WriteString("\n\n")

	// Command list
	if len(p.filtered) == 0 {
		b.WriteString(p.emptyStyle.Render("No matching commands"))
		b.WriteString("\n")
	} else {
		maxItems := 10 // Maximum items to display
		if len(p.filtered) < maxItems {
			maxItems = len(p.filtered)
		}

		for i := 0; i < maxItems; i++ {
			cmd := p.filtered[i]

			// Build the command line
			var line strings.Builder

			// Shortcut (right-aligned)
			if cmd.Shortcut != "" {
				line.WriteString(p.shortcutStyle.Render("[" + cmd.Shortcut + "]"))
				line.WriteString(" ")
			} else {
				line.WriteString(p.shortcutStyle.Render(""))
				line.WriteString(" ")
			}

			// Command name
			line.WriteString(cmd.Name)

			// Apply selected or normal style
			if i == p.selected {
				b.WriteString(p.selectedStyle.Render(line.String()))
			} else {
				b.WriteString(p.itemStyle.Render(line.String()))
			}
			b.WriteString("\n")
		}

		// Show indicator if there are more items
		if len(p.filtered) > maxItems {
			remaining := len(p.filtered) - maxItems
			b.WriteString(p.emptyStyle.Render(strings.Repeat(" ", 10) + "... and " + itoa(remaining) + " more"))
			b.WriteString("\n")
		}
	}

	// Hints
	b.WriteString(p.hintStyle.Render("Enter: execute  Esc: cancel  Up/Down: navigate"))

	content := b.String()

	// Apply overlay style
	if p.width > 0 {
		overlayWidth := 60
		if overlayWidth > p.width-4 {
			overlayWidth = p.width - 4
		}
		p.overlayStyle = p.overlayStyle.Width(overlayWidth)
	}

	return p.overlayStyle.Render(content)
}

// itoa converts an integer to a string (simple helper to avoid strconv import).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
