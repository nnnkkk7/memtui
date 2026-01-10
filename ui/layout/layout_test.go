package layout_test

import (
	"testing"

	"github.com/nnnkkk7/memtui/ui/layout"
)

// TestNewLayout tests the Layout constructor
func TestNewLayout(t *testing.T) {
	l := layout.New()
	if l == nil {
		t.Fatal("expected non-nil layout")
	}
}

// TestLayoutDefault tests default values
func TestLayoutDefault(t *testing.T) {
	l := layout.New()

	// Default values should be set
	if l.Width() != 0 {
		t.Errorf("expected default width 0, got %d", l.Width())
	}
	if l.Height() != 0 {
		t.Errorf("expected default height 0, got %d", l.Height())
	}
}

// TestLayoutSetSize tests setting terminal size
func TestLayoutSetSize(t *testing.T) {
	l := layout.New()
	l.SetSize(120, 40)

	if l.Width() != 120 {
		t.Errorf("expected width 120, got %d", l.Width())
	}
	if l.Height() != 40 {
		t.Errorf("expected height 40, got %d", l.Height())
	}
}

// TestCalculate tests the Calculate method
func TestCalculate(t *testing.T) {
	l := layout.New()
	l.SetSize(120, 40)
	l.Calculate()

	// After Calculate, pane sizes should be computed
	if l.KeyListWidth() <= 0 {
		t.Errorf("expected positive KeyListWidth, got %d", l.KeyListWidth())
	}
	if l.ViewerWidth() <= 0 {
		t.Errorf("expected positive ViewerWidth, got %d", l.ViewerWidth())
	}
}

// TestCalculateTwoPaneLayout tests 2-pane layout calculation
func TestCalculateTwoPaneLayout(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// KeyListWidth + ViewerWidth should equal Width (minus border)
	// Allow for border character (1 char between panes)
	totalWidth := l.KeyListWidth() + l.ViewerWidth() + 1
	if totalWidth != 100 {
		t.Errorf("expected total width 100, got %d (KeyList: %d, Viewer: %d)",
			totalWidth, l.KeyListWidth(), l.ViewerWidth())
	}
}

// TestCalculateKeyListRatio tests KeyList width ratio (30%)
func TestCalculateKeyListRatio(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// KeyList should be approximately 30% of width
	expectedKeyListWidth := 30
	if l.KeyListWidth() != expectedKeyListWidth {
		t.Errorf("expected KeyListWidth %d (30%%), got %d", expectedKeyListWidth, l.KeyListWidth())
	}
}

// TestCalculateViewerRatio tests Viewer width ratio (70% - border)
func TestCalculateViewerRatio(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// Viewer should be approximately 70% of width minus border
	expectedViewerWidth := 69 // 100 - 30 - 1 (border)
	if l.ViewerWidth() != expectedViewerWidth {
		t.Errorf("expected ViewerWidth %d (70%% - border), got %d", expectedViewerWidth, l.ViewerWidth())
	}
}

// TestCalculateContentHeight tests content area height
func TestCalculateContentHeight(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// Content height should account for header (2 lines) and statusbar (1 line)
	expectedContentHeight := 27 // 30 - 2 (header) - 1 (statusbar)
	if l.ContentHeight() != expectedContentHeight {
		t.Errorf("expected ContentHeight %d, got %d", expectedContentHeight, l.ContentHeight())
	}
}

// TestCalculateHeaderHeight tests header height
func TestCalculateHeaderHeight(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	if l.HeaderHeight() != 2 {
		t.Errorf("expected HeaderHeight 2, got %d", l.HeaderHeight())
	}
}

// TestCalculateStatusBarHeight tests status bar height
func TestCalculateStatusBarHeight(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	if l.StatusBarHeight() != 1 {
		t.Errorf("expected StatusBarHeight 1, got %d", l.StatusBarHeight())
	}
}

// TestCalculateMinimumSize tests minimum size handling
func TestCalculateMinimumSize(t *testing.T) {
	l := layout.New()
	// Use a size that can accommodate both minimums
	minWidth := layout.MinKeyListWidth + layout.MinViewerWidth + 1 // +1 for border
	l.SetSize(minWidth, 10)
	l.Calculate()

	// Should still produce valid layout with minimum sizes
	if l.KeyListWidth() < layout.MinKeyListWidth {
		t.Errorf("KeyListWidth should not be less than minimum %d, got %d",
			layout.MinKeyListWidth, l.KeyListWidth())
	}
	if l.ViewerWidth() < layout.MinViewerWidth {
		t.Errorf("ViewerWidth should not be less than minimum %d, got %d",
			layout.MinViewerWidth, l.ViewerWidth())
	}
}

// TestCalculateVerySmallSize tests behavior with terminal smaller than minimums
func TestCalculateVerySmallSize(t *testing.T) {
	l := layout.New()
	l.SetSize(20, 10) // Smaller than MinKeyListWidth + MinViewerWidth + border
	l.Calculate()

	// Should not panic and should produce some layout
	// Total should still equal width
	totalWidth := l.KeyListWidth() + l.ViewerWidth() + 1
	if totalWidth != 20 {
		t.Errorf("expected total width 20, got %d", totalWidth)
	}
}

// TestKeyListX tests KeyList X position
func TestKeyListX(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// KeyList should start at X=0
	if l.KeyListX() != 0 {
		t.Errorf("expected KeyListX 0, got %d", l.KeyListX())
	}
}

// TestViewerX tests Viewer X position
func TestViewerX(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// Viewer should start after KeyList + border
	expectedViewerX := l.KeyListWidth() + 1
	if l.ViewerX() != expectedViewerX {
		t.Errorf("expected ViewerX %d, got %d", expectedViewerX, l.ViewerX())
	}
}

// TestContentY tests content area Y position
func TestContentY(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	// Content area should start after header
	if l.ContentY() != 2 {
		t.Errorf("expected ContentY 2, got %d", l.ContentY())
	}
}

// TestKeyListBounds tests full KeyList bounds
func TestKeyListBounds(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	bounds := l.KeyListBounds()

	if bounds.X != 0 {
		t.Errorf("expected KeyList X 0, got %d", bounds.X)
	}
	if bounds.Y != 2 {
		t.Errorf("expected KeyList Y 2 (after header), got %d", bounds.Y)
	}
	if bounds.Width != 30 {
		t.Errorf("expected KeyList Width 30, got %d", bounds.Width)
	}
	if bounds.Height != 27 {
		t.Errorf("expected KeyList Height 27, got %d", bounds.Height)
	}
}

// TestViewerBounds tests full Viewer bounds
func TestViewerBounds(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	bounds := l.ViewerBounds()

	if bounds.X != 31 {
		t.Errorf("expected Viewer X 31, got %d", bounds.X)
	}
	if bounds.Y != 2 {
		t.Errorf("expected Viewer Y 2 (after header), got %d", bounds.Y)
	}
	if bounds.Width != 69 {
		t.Errorf("expected Viewer Width 69, got %d", bounds.Width)
	}
	if bounds.Height != 27 {
		t.Errorf("expected Viewer Height 27, got %d", bounds.Height)
	}
}

// TestHeaderBounds tests header bounds
func TestHeaderBounds(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	bounds := l.HeaderBounds()

	if bounds.X != 0 {
		t.Errorf("expected Header X 0, got %d", bounds.X)
	}
	if bounds.Y != 0 {
		t.Errorf("expected Header Y 0, got %d", bounds.Y)
	}
	if bounds.Width != 100 {
		t.Errorf("expected Header Width 100, got %d", bounds.Width)
	}
	if bounds.Height != 2 {
		t.Errorf("expected Header Height 2, got %d", bounds.Height)
	}
}

// TestStatusBarBounds tests status bar bounds
func TestStatusBarBounds(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	bounds := l.StatusBarBounds()

	if bounds.X != 0 {
		t.Errorf("expected StatusBar X 0, got %d", bounds.X)
	}
	if bounds.Y != 29 {
		t.Errorf("expected StatusBar Y 29, got %d", bounds.Y)
	}
	if bounds.Width != 100 {
		t.Errorf("expected StatusBar Width 100, got %d", bounds.Width)
	}
	if bounds.Height != 1 {
		t.Errorf("expected StatusBar Height 1, got %d", bounds.Height)
	}
}

// TestSetKeyListRatio tests custom KeyList ratio
func TestSetKeyListRatio(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.SetKeyListRatio(0.4) // 40%
	l.Calculate()

	expectedKeyListWidth := 40
	if l.KeyListWidth() != expectedKeyListWidth {
		t.Errorf("expected KeyListWidth %d (40%%), got %d", expectedKeyListWidth, l.KeyListWidth())
	}
}

// TestSetKeyListRatioClamp tests ratio clamping
func TestSetKeyListRatioClamp(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)

	// Test too low ratio
	l.SetKeyListRatio(0.05)
	l.Calculate()
	if l.KeyListWidth() < layout.MinKeyListWidth {
		t.Errorf("ratio should be clamped to ensure minimum KeyListWidth")
	}

	// Test too high ratio
	l.SetKeyListRatio(0.95)
	l.Calculate()
	if l.ViewerWidth() < layout.MinViewerWidth {
		t.Errorf("ratio should be clamped to ensure minimum ViewerWidth")
	}
}

// TestZeroSize tests handling of zero size
func TestZeroSize(t *testing.T) {
	l := layout.New()
	l.SetSize(0, 0)
	l.Calculate()

	// Should not panic and produce safe defaults
	if l.KeyListWidth() < 0 {
		t.Error("KeyListWidth should not be negative")
	}
	if l.ViewerWidth() < 0 {
		t.Error("ViewerWidth should not be negative")
	}
}

// TestLargeTerminal tests layout with large terminal
func TestLargeTerminal(t *testing.T) {
	l := layout.New()
	l.SetSize(300, 100)
	l.Calculate()

	// Should scale proportionally
	totalWidth := l.KeyListWidth() + l.ViewerWidth() + 1
	if totalWidth != 300 {
		t.Errorf("expected total width 300, got %d", totalWidth)
	}
}

// TestBoundsType tests the Bounds struct
func TestBoundsType(t *testing.T) {
	bounds := layout.Bounds{X: 10, Y: 20, Width: 30, Height: 40}

	if bounds.X != 10 {
		t.Errorf("expected X 10, got %d", bounds.X)
	}
	if bounds.Y != 20 {
		t.Errorf("expected Y 20, got %d", bounds.Y)
	}
	if bounds.Width != 30 {
		t.Errorf("expected Width 30, got %d", bounds.Width)
	}
	if bounds.Height != 40 {
		t.Errorf("expected Height 40, got %d", bounds.Height)
	}
}

// TestBoundsRight tests Bounds.Right method
func TestBoundsRight(t *testing.T) {
	bounds := layout.Bounds{X: 10, Y: 20, Width: 30, Height: 40}

	if bounds.Right() != 40 {
		t.Errorf("expected Right 40, got %d", bounds.Right())
	}
}

// TestBoundsBottom tests Bounds.Bottom method
func TestBoundsBottom(t *testing.T) {
	bounds := layout.Bounds{X: 10, Y: 20, Width: 30, Height: 40}

	if bounds.Bottom() != 60 {
		t.Errorf("expected Bottom 60, got %d", bounds.Bottom())
	}
}

// TestBoundsContains tests Bounds.Contains method
func TestBoundsContains(t *testing.T) {
	bounds := layout.Bounds{X: 10, Y: 20, Width: 30, Height: 40}

	// Inside
	if !bounds.Contains(15, 30) {
		t.Error("expected (15, 30) to be inside bounds")
	}

	// On edge
	if !bounds.Contains(10, 20) {
		t.Error("expected (10, 20) to be inside bounds (top-left corner)")
	}

	// Outside
	if bounds.Contains(5, 30) {
		t.Error("expected (5, 30) to be outside bounds")
	}
	if bounds.Contains(15, 70) {
		t.Error("expected (15, 70) to be outside bounds")
	}
}

// TestRecalculateOnResize tests recalculation after resize
func TestRecalculateOnResize(t *testing.T) {
	l := layout.New()

	l.SetSize(100, 30)
	l.Calculate()
	firstKeyListWidth := l.KeyListWidth()

	l.SetSize(200, 50)
	l.Calculate()
	secondKeyListWidth := l.KeyListWidth()

	if secondKeyListWidth <= firstKeyListWidth {
		t.Errorf("KeyListWidth should increase with larger terminal: %d -> %d",
			firstKeyListWidth, secondKeyListWidth)
	}
}

// TestIsValid tests layout validity check
func TestIsValid(t *testing.T) {
	l := layout.New()

	// Before Calculate, should be invalid
	if l.IsValid() {
		t.Error("layout should be invalid before Calculate")
	}

	l.SetSize(100, 30)
	l.Calculate()

	// After Calculate with valid size, should be valid
	if !l.IsValid() {
		t.Error("layout should be valid after Calculate with valid size")
	}
}

// TestString tests string representation
func TestString(t *testing.T) {
	l := layout.New()
	l.SetSize(100, 30)
	l.Calculate()

	str := l.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}
}
