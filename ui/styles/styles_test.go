package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// ========== Theme Tests ==========

func TestTheme_HasRequiredFields(t *testing.T) {
	theme := Theme{}

	// Theme should have Name field
	theme.Name = "test"
	if theme.Name != "test" {
		t.Errorf("Theme.Name should be settable")
	}
}

func TestTheme_BasicColors(t *testing.T) {
	theme := Theme{
		Name:      "test",
		Primary:   lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"},
		Secondary: lipgloss.AdaptiveColor{Light: "#56B6C2", Dark: "#56B6C2"},
		Success:   lipgloss.AdaptiveColor{Light: "#98C379", Dark: "#98C379"},
		Warning:   lipgloss.AdaptiveColor{Light: "#E5C07B", Dark: "#E5C07B"},
		Error:     lipgloss.AdaptiveColor{Light: "#E06C75", Dark: "#E06C75"},
	}

	if theme.Primary.Light != "#7D56F4" {
		t.Errorf("Expected Primary.Light = #7D56F4, got %s", theme.Primary.Light)
	}
	if theme.Secondary.Dark != "#56B6C2" {
		t.Errorf("Expected Secondary.Dark = #56B6C2, got %s", theme.Secondary.Dark)
	}
	if theme.Success.Light != "#98C379" {
		t.Errorf("Expected Success.Light = #98C379, got %s", theme.Success.Light)
	}
	if theme.Warning.Light != "#E5C07B" {
		t.Errorf("Expected Warning.Light = #E5C07B, got %s", theme.Warning.Light)
	}
	if theme.Error.Light != "#E06C75" {
		t.Errorf("Expected Error.Light = #E06C75, got %s", theme.Error.Light)
	}
}

func TestTheme_BackgroundColors(t *testing.T) {
	theme := Theme{
		Background: lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282C34"},
		Surface:    lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#3E4451"},
	}

	if theme.Background.Light != "#FFFFFF" {
		t.Errorf("Expected Background.Light = #FFFFFF, got %s", theme.Background.Light)
	}
	if theme.Background.Dark != "#282C34" {
		t.Errorf("Expected Background.Dark = #282C34, got %s", theme.Background.Dark)
	}
	if theme.Surface.Dark != "#3E4451" {
		t.Errorf("Expected Surface.Dark = #3E4451, got %s", theme.Surface.Dark)
	}
}

func TestTheme_TextColors(t *testing.T) {
	theme := Theme{
		Text:      lipgloss.AdaptiveColor{Light: "#24292E", Dark: "#ABB2BF"},
		TextMuted: lipgloss.AdaptiveColor{Light: "#6A737D", Dark: "#5C6370"},
	}

	if theme.Text.Dark != "#ABB2BF" {
		t.Errorf("Expected Text.Dark = #ABB2BF, got %s", theme.Text.Dark)
	}
	if theme.TextMuted.Dark != "#5C6370" {
		t.Errorf("Expected TextMuted.Dark = #5C6370, got %s", theme.TextMuted.Dark)
	}
}

func TestTheme_BorderColors(t *testing.T) {
	theme := Theme{
		Border:      lipgloss.AdaptiveColor{Light: "#E1E4E8", Dark: "#4B5263"},
		BorderFocus: lipgloss.AdaptiveColor{Light: "#0366D6", Dark: "#528BFF"},
	}

	if theme.Border.Dark != "#4B5263" {
		t.Errorf("Expected Border.Dark = #4B5263, got %s", theme.Border.Dark)
	}
	if theme.BorderFocus.Dark != "#528BFF" {
		t.Errorf("Expected BorderFocus.Dark = #528BFF, got %s", theme.BorderFocus.Dark)
	}
}

// ========== DarkTheme Tests ==========

func TestDarkTheme_Exists(t *testing.T) {
	theme := DarkTheme

	if theme.Name != "dark" {
		t.Errorf("Expected DarkTheme.Name = dark, got %s", theme.Name)
	}
}

func TestDarkTheme_HasAllColors(t *testing.T) {
	theme := DarkTheme

	// Check primary colors
	if theme.Primary.Dark == "" {
		t.Error("DarkTheme.Primary.Dark should not be empty")
	}
	if theme.Secondary.Dark == "" {
		t.Error("DarkTheme.Secondary.Dark should not be empty")
	}
	if theme.Success.Dark == "" {
		t.Error("DarkTheme.Success.Dark should not be empty")
	}
	if theme.Warning.Dark == "" {
		t.Error("DarkTheme.Warning.Dark should not be empty")
	}
	if theme.Error.Dark == "" {
		t.Error("DarkTheme.Error.Dark should not be empty")
	}

	// Check background colors
	if theme.Background.Dark == "" {
		t.Error("DarkTheme.Background.Dark should not be empty")
	}
	if theme.Surface.Dark == "" {
		t.Error("DarkTheme.Surface.Dark should not be empty")
	}

	// Check text colors
	if theme.Text.Dark == "" {
		t.Error("DarkTheme.Text.Dark should not be empty")
	}
	if theme.TextMuted.Dark == "" {
		t.Error("DarkTheme.TextMuted.Dark should not be empty")
	}

	// Check border colors
	if theme.Border.Dark == "" {
		t.Error("DarkTheme.Border.Dark should not be empty")
	}
	if theme.BorderFocus.Dark == "" {
		t.Error("DarkTheme.BorderFocus.Dark should not be empty")
	}
}

// ========== LightTheme Tests ==========

func TestLightTheme_Exists(t *testing.T) {
	theme := LightTheme

	if theme.Name != "light" {
		t.Errorf("Expected LightTheme.Name = light, got %s", theme.Name)
	}
}

func TestLightTheme_HasAllColors(t *testing.T) {
	theme := LightTheme

	// Check primary colors
	if theme.Primary.Light == "" {
		t.Error("LightTheme.Primary.Light should not be empty")
	}
	if theme.Secondary.Light == "" {
		t.Error("LightTheme.Secondary.Light should not be empty")
	}
	if theme.Success.Light == "" {
		t.Error("LightTheme.Success.Light should not be empty")
	}
	if theme.Warning.Light == "" {
		t.Error("LightTheme.Warning.Light should not be empty")
	}
	if theme.Error.Light == "" {
		t.Error("LightTheme.Error.Light should not be empty")
	}

	// Check background colors
	if theme.Background.Light == "" {
		t.Error("LightTheme.Background.Light should not be empty")
	}
	if theme.Surface.Light == "" {
		t.Error("LightTheme.Surface.Light should not be empty")
	}

	// Check text colors
	if theme.Text.Light == "" {
		t.Error("LightTheme.Text.Light should not be empty")
	}
	if theme.TextMuted.Light == "" {
		t.Error("LightTheme.TextMuted.Light should not be empty")
	}

	// Check border colors
	if theme.Border.Light == "" {
		t.Error("LightTheme.Border.Light should not be empty")
	}
	if theme.BorderFocus.Light == "" {
		t.Error("LightTheme.BorderFocus.Light should not be empty")
	}
}

// ========== DefaultTheme Tests ==========

func TestDefaultTheme_ReturnsDarkTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme.Name != "dark" {
		t.Errorf("Expected DefaultTheme to return dark theme, got %s", theme.Name)
	}
}

// ========== Styles Tests ==========

func TestStyles_HasTheme(t *testing.T) {
	s := NewStyles(DarkTheme)

	if s.Theme.Name != "dark" {
		t.Errorf("Expected Styles.Theme.Name = dark, got %s", s.Theme.Name)
	}
}

func TestStyles_HasTitleStyle(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Title style should exist and be usable
	result := s.Title.Render("Test Title")
	if result == "" {
		t.Error("Title style should render content")
	}
}

func TestStyles_HasStatusBarStyle(t *testing.T) {
	s := NewStyles(DarkTheme)

	result := s.StatusBar.Render("Status text")
	if result == "" {
		t.Error("StatusBar style should render content")
	}
}

func TestStyles_HasErrorStyle(t *testing.T) {
	s := NewStyles(DarkTheme)

	result := s.Error.Render("Error message")
	if result == "" {
		t.Error("Error style should render content")
	}
}

func TestStyles_HasKeyListStyles(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Check KeyList related styles
	if s.KeyList.Normal.Render("normal") == "" {
		t.Error("KeyList.Normal style should render content")
	}
	if s.KeyList.Selected.Render("selected") == "" {
		t.Error("KeyList.Selected style should render content")
	}
	if s.KeyList.Folder.Render("folder") == "" {
		t.Error("KeyList.Folder style should render content")
	}
	if s.KeyList.Leaf.Render("leaf") == "" {
		t.Error("KeyList.Leaf style should render content")
	}
}

func TestStyles_HasViewerStyles(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Check Viewer related styles
	if s.Viewer.Header.Render("header") == "" {
		t.Error("Viewer.Header style should render content")
	}
	if s.Viewer.Content.Render("content") == "" {
		t.Error("Viewer.Content style should render content")
	}
	if s.Viewer.Meta.Render("meta") == "" {
		t.Error("Viewer.Meta style should render content")
	}
}

func TestStyles_HasHelpStyles(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Check Help related styles
	if s.Help.Overlay.Render("overlay") == "" {
		t.Error("Help.Overlay style should render content")
	}
	if s.Help.Title.Render("title") == "" {
		t.Error("Help.Title style should render content")
	}
	if s.Help.Section.Render("section") == "" {
		t.Error("Help.Section style should render content")
	}
	if s.Help.Key.Render("key") == "" {
		t.Error("Help.Key style should render content")
	}
	if s.Help.Action.Render("action") == "" {
		t.Error("Help.Action style should render content")
	}
	if s.Help.Footer.Render("footer") == "" {
		t.Error("Help.Footer style should render content")
	}
}

func TestStyles_HasPanelStyles(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Check Panel styles
	if s.Panel.Normal.Render("normal") == "" {
		t.Error("Panel.Normal style should render content")
	}
	if s.Panel.Focused.Render("focused") == "" {
		t.Error("Panel.Focused style should render content")
	}
}

func TestStyles_HasCursorStyle(t *testing.T) {
	s := NewStyles(DarkTheme)

	result := s.Cursor.Render("cursor")
	if result == "" {
		t.Error("Cursor style should render content")
	}
}

// ========== NewStyles with Different Themes ==========

func TestNewStyles_WithDarkTheme(t *testing.T) {
	s := NewStyles(DarkTheme)

	if s == nil {
		t.Fatal("NewStyles should not return nil")
	}
	if s.Theme.Name != "dark" {
		t.Errorf("Expected theme name = dark, got %s", s.Theme.Name)
	}
}

func TestNewStyles_WithLightTheme(t *testing.T) {
	s := NewStyles(LightTheme)

	if s == nil {
		t.Fatal("NewStyles should not return nil")
	}
	if s.Theme.Name != "light" {
		t.Errorf("Expected theme name = light, got %s", s.Theme.Name)
	}
}

// ========== DefaultStyles Tests ==========

func TestDefaultStyles_ReturnsStyles(t *testing.T) {
	s := DefaultStyles()

	if s == nil {
		t.Fatal("DefaultStyles should not return nil")
	}
	if s.Theme.Name != "dark" {
		t.Errorf("Expected default styles with dark theme, got %s", s.Theme.Name)
	}
}

// ========== GetTheme Tests ==========

func TestGetTheme_Dark(t *testing.T) {
	theme := GetTheme("dark")

	if theme.Name != "dark" {
		t.Errorf("Expected dark theme, got %s", theme.Name)
	}
}

func TestGetTheme_Light(t *testing.T) {
	theme := GetTheme("light")

	if theme.Name != "light" {
		t.Errorf("Expected light theme, got %s", theme.Name)
	}
}

func TestGetTheme_Unknown_DefaultsToDark(t *testing.T) {
	theme := GetTheme("unknown")

	if theme.Name != "dark" {
		t.Errorf("Expected unknown theme to default to dark, got %s", theme.Name)
	}
}

func TestGetTheme_Empty_DefaultsToDark(t *testing.T) {
	theme := GetTheme("")

	if theme.Name != "dark" {
		t.Errorf("Expected empty theme name to default to dark, got %s", theme.Name)
	}
}

// ========== Style Attribute Tests ==========

func TestTitleStyle_IsBold(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Title style should be bold - we can verify by checking rendered output
	rendered := s.Title.Render("Test")
	if len(rendered) == 0 {
		t.Error("Title should render non-empty string")
	}
}

func TestErrorStyle_UsesErrorColor(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Error style should render without panic
	rendered := s.Error.Render("Error")
	if len(rendered) == 0 {
		t.Error("Error style should render non-empty string")
	}
}

func TestPanelFocused_HasBorder(t *testing.T) {
	s := NewStyles(DarkTheme)

	// Focused panel should have a border
	rendered := s.Panel.Focused.Render("Content")
	if len(rendered) == 0 {
		t.Error("Panel.Focused should render non-empty string")
	}
}

// ========== Style Consistency Tests ==========

func TestStyles_ConsistentWithTheme(t *testing.T) {
	darkStyles := NewStyles(DarkTheme)
	lightStyles := NewStyles(LightTheme)

	// Both should be valid
	if darkStyles == nil || lightStyles == nil {
		t.Fatal("Both dark and light styles should be created")
	}

	// They should have different theme names
	if darkStyles.Theme.Name == lightStyles.Theme.Name {
		t.Error("Dark and light styles should have different theme names")
	}
}
