package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	switch m.state {
	case StateConnecting:
		return m.viewConnecting()
	case StateLoading:
		return m.viewLoading()
	case StateError:
		return m.viewError()
	default:
		return m.viewMain()
	}
}

func (m *Model) viewConnecting() string {
	return m.styles.Connecting.Render(fmt.Sprintf("Connecting to %s...", m.addr))
}

func (m *Model) viewLoading() string {
	return m.styles.Connecting.Render("Loading keys...")
}

func (m *Model) viewError() string {
	return m.styles.Error.Render(fmt.Sprintf("Error: %s\n\nPress 'q' to quit, 'r' to retry.", m.err))
}

func (m *Model) viewMain() string {
	// Overlay dialogs if active
	if m.confirmDialog != nil {
		return m.renderWithOverlay(m.confirmDialog.View())
	}
	if m.inputDialog != nil {
		return m.renderWithOverlay(m.inputDialog.View())
	}
	if m.editor != nil {
		return m.editor.View()
	}
	if m.commandPalette.Visible() {
		return m.renderWithOverlay(m.commandPalette.View())
	}
	if m.help.Visible() {
		return m.help.View()
	}

	// Main two-pane layout
	keyListWidth := m.width * 30 / 100
	viewerWidth := m.width - keyListWidth - 4

	// Colors for focused/unfocused panes
	focusedColor := lipgloss.Color("#98C379")   // Bright green for focused
	unfocusedColor := lipgloss.Color("#4B5263") // Dim gray for unfocused

	// Style for focused/unfocused panes
	var keyListStyle, viewerStyle lipgloss.Style
	var keyListTitle, viewerTitle string

	if m.focus == FocusKeyList {
		// KeyList is focused - use thick border and bright color
		keyListStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(focusedColor)
		keyListTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(focusedColor).
			Render("▶ Keys")

		// Viewer is unfocused - use rounded border and dim color
		viewerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(unfocusedColor)
		viewerTitle = lipgloss.NewStyle().
			Foreground(unfocusedColor).
			Render("  Value")
	} else {
		// KeyList is unfocused
		keyListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(unfocusedColor)
		keyListTitle = lipgloss.NewStyle().
			Foreground(unfocusedColor).
			Render("  Keys")

		// Viewer is focused - use thick border and bright color
		viewerStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(focusedColor)
		viewerTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(focusedColor).
			Render("▶ Value")
	}

	// Render key list with title
	keyListContent := m.keyList.View()
	keyListBox := keyListStyle.
		Width(keyListWidth - 2).
		Height(m.height - 5).
		Render(keyListContent)
	keyListPanel := lipgloss.JoinVertical(lipgloss.Left, keyListTitle, keyListBox)

	// Render viewer with title
	viewerContent := m.viewer.View()
	viewerBox := viewerStyle.
		Width(viewerWidth - 2).
		Height(m.height - 5).
		Render(viewerContent)
	viewerPanel := lipgloss.JoinVertical(lipgloss.Left, viewerTitle, viewerBox)

	// Join horizontally
	main := lipgloss.JoinHorizontal(lipgloss.Top, keyListPanel, viewerPanel)

	// Status bar
	filterStatus := ""
	if m.filtering {
		filterStatus = fmt.Sprintf(" | Filter: %s_", m.filterInput)
	}

	keyCount := len(m.keyList.FilteredKeys())
	statusText := fmt.Sprintf(" %s | %d keys | %s%s ", m.addr, keyCount, m.version, filterStatus)
	status := m.styles.StatusBar.Width(m.width).Render(statusText)

	// Help bar
	helpText := "q:quit r:refresh /:filter d:delete e:edit n:new ?:help Tab/Esc:switch Ctrl+P:commands"
	helpBar := m.styles.Help.Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left, main, status, helpBar)
}

func (m *Model) renderWithOverlay(overlay string) string {
	// Center the overlay on top of a dimmed background
	overlayHeight := lipgloss.Height(overlay)
	overlayWidth := lipgloss.Width(overlay)

	// Calculate padding
	topPadding := (m.height - overlayHeight) / 2
	leftPadding := (m.width - overlayWidth) / 2

	if topPadding < 0 {
		topPadding = 0
	}
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Create centered overlay
	centeredOverlay := lipgloss.NewStyle().
		MarginTop(topPadding).
		MarginLeft(leftPadding).
		Render(overlay)

	return centeredOverlay
}
