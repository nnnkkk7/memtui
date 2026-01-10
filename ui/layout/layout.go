// Package layout provides terminal layout management for memtui.
// It calculates pane sizes for the 2-pane layout (KeyList + Viewer).
package layout

import "fmt"

// Constants for layout constraints
const (
	// MinKeyListWidth is the minimum width for the key list pane
	MinKeyListWidth = 10
	// MinViewerWidth is the minimum width for the viewer pane
	MinViewerWidth = 20
	// DefaultKeyListRatio is the default width ratio for the key list (30%)
	DefaultKeyListRatio = 0.30
	// HeaderLines is the number of lines used by the header
	HeaderLines = 2
	// StatusBarLines is the number of lines used by the status bar
	StatusBarLines = 1
	// BorderWidth is the width of the border between panes
	BorderWidth = 1
)

// Bounds represents a rectangular area in the terminal
type Bounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Right returns the right edge X coordinate
func (b Bounds) Right() int {
	return b.X + b.Width
}

// Bottom returns the bottom edge Y coordinate
func (b Bounds) Bottom() int {
	return b.Y + b.Height
}

// Contains checks if a point is within the bounds
func (b Bounds) Contains(x, y int) bool {
	return x >= b.X && x < b.Right() && y >= b.Y && y < b.Bottom()
}

// Layout manages the terminal layout for memtui
type Layout struct {
	// Terminal dimensions
	width  int
	height int

	// Configuration
	keyListRatio float64

	// Calculated values
	keyListWidth    int
	viewerWidth     int
	contentHeight   int
	headerHeight    int
	statusBarHeight int

	// Calculated positions
	keyListX int
	viewerX  int
	contentY int

	// State
	calculated bool
}

// New creates a new Layout with default settings
func New() *Layout {
	return &Layout{
		keyListRatio:    DefaultKeyListRatio,
		headerHeight:    HeaderLines,
		statusBarHeight: StatusBarLines,
	}
}

// SetSize sets the terminal dimensions
func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.calculated = false
}

// SetKeyListRatio sets the ratio of width for the key list pane
// Value should be between 0.1 and 0.9
func (l *Layout) SetKeyListRatio(ratio float64) {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	l.keyListRatio = ratio
	l.calculated = false
}

// Calculate computes all layout dimensions based on current terminal size
func (l *Layout) Calculate() {
	// Handle zero or very small sizes
	if l.width <= 0 || l.height <= 0 {
		l.keyListWidth = 0
		l.viewerWidth = 0
		l.contentHeight = 0
		l.calculated = false
		return
	}

	// Calculate content height (total - header - statusbar)
	l.contentHeight = l.height - HeaderLines - StatusBarLines
	if l.contentHeight < 0 {
		l.contentHeight = 0
	}

	// Calculate key list width based on ratio
	availableWidth := l.width - BorderWidth
	l.keyListWidth = int(float64(l.width) * l.keyListRatio)

	// Ensure minimum widths
	if l.keyListWidth < MinKeyListWidth {
		l.keyListWidth = MinKeyListWidth
	}

	// Calculate viewer width
	l.viewerWidth = availableWidth - l.keyListWidth
	if l.viewerWidth < MinViewerWidth {
		// If viewer is too small, reduce key list
		l.viewerWidth = MinViewerWidth
		l.keyListWidth = availableWidth - l.viewerWidth
		if l.keyListWidth < MinKeyListWidth {
			l.keyListWidth = MinKeyListWidth
		}
	}

	// Handle case where total width is less than minimums
	if l.width < MinKeyListWidth+MinViewerWidth+BorderWidth {
		// Split evenly
		halfWidth := l.width / 2
		l.keyListWidth = halfWidth
		l.viewerWidth = l.width - l.keyListWidth - BorderWidth
		if l.viewerWidth < 0 {
			l.viewerWidth = 0
		}
	}

	// Calculate positions
	l.keyListX = 0
	l.viewerX = l.keyListWidth + BorderWidth
	l.contentY = HeaderLines

	l.calculated = true
}

// Width returns the terminal width
func (l *Layout) Width() int {
	return l.width
}

// Height returns the terminal height
func (l *Layout) Height() int {
	return l.height
}

// KeyListWidth returns the width of the key list pane
func (l *Layout) KeyListWidth() int {
	return l.keyListWidth
}

// ViewerWidth returns the width of the viewer pane
func (l *Layout) ViewerWidth() int {
	return l.viewerWidth
}

// ContentHeight returns the height of the content area
func (l *Layout) ContentHeight() int {
	return l.contentHeight
}

// HeaderHeight returns the height of the header
func (l *Layout) HeaderHeight() int {
	return l.headerHeight
}

// StatusBarHeight returns the height of the status bar
func (l *Layout) StatusBarHeight() int {
	return l.statusBarHeight
}

// KeyListX returns the X position of the key list pane
func (l *Layout) KeyListX() int {
	return l.keyListX
}

// ViewerX returns the X position of the viewer pane
func (l *Layout) ViewerX() int {
	return l.viewerX
}

// ContentY returns the Y position of the content area (after header)
func (l *Layout) ContentY() int {
	return l.contentY
}

// KeyListBounds returns the full bounds of the key list pane
func (l *Layout) KeyListBounds() Bounds {
	return Bounds{
		X:      l.keyListX,
		Y:      l.contentY,
		Width:  l.keyListWidth,
		Height: l.contentHeight,
	}
}

// ViewerBounds returns the full bounds of the viewer pane
func (l *Layout) ViewerBounds() Bounds {
	return Bounds{
		X:      l.viewerX,
		Y:      l.contentY,
		Width:  l.viewerWidth,
		Height: l.contentHeight,
	}
}

// HeaderBounds returns the bounds of the header area
func (l *Layout) HeaderBounds() Bounds {
	return Bounds{
		X:      0,
		Y:      0,
		Width:  l.width,
		Height: l.headerHeight,
	}
}

// StatusBarBounds returns the bounds of the status bar
func (l *Layout) StatusBarBounds() Bounds {
	return Bounds{
		X:      0,
		Y:      l.height - l.statusBarHeight,
		Width:  l.width,
		Height: l.statusBarHeight,
	}
}

// IsValid returns true if the layout has been calculated and is valid
func (l *Layout) IsValid() bool {
	return l.calculated && l.width > 0 && l.height > 0
}

// String returns a string representation of the layout for debugging
func (l *Layout) String() string {
	return fmt.Sprintf(
		"Layout{Width:%d, Height:%d, KeyList:%d@%d, Viewer:%d@%d, Content:%d@Y%d}",
		l.width, l.height,
		l.keyListWidth, l.keyListX,
		l.viewerWidth, l.viewerX,
		l.contentHeight, l.contentY,
	)
}
