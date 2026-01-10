package keylist_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

func TestNewModel(t *testing.T) {
	m := keylist.NewModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestModel_SetKeys(t *testing.T) {
	m := keylist.NewModel()
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
		{Key: "session:abc", Size: 50},
	}

	m.SetKeys(keys)

	if m.KeyCount() != 3 {
		t.Errorf("expected 3 keys, got %d", m.KeyCount())
	}
}

func TestModel_BuildTree(t *testing.T) {
	m := keylist.NewModel()
	keys := []models.KeyInfo{
		{Key: "user:profile:1", Size: 100},
		{Key: "user:profile:2", Size: 200},
		{Key: "user:session:abc", Size: 50},
		{Key: "cache:data", Size: 300},
	}

	m.SetKeys(keys)
	m.SetDelimiter(":")

	tree := m.Tree()
	if tree == nil {
		t.Fatal("expected non-nil tree")
	}

	// Should have 2 root nodes: "user" and "cache"
	if len(tree.Children) != 2 {
		t.Errorf("expected 2 root nodes, got %d", len(tree.Children))
	}
}

func TestModel_Navigation(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	keys := []models.KeyInfo{
		{Key: "a", Size: 100},
		{Key: "b", Size: 200},
		{Key: "c", Size: 300},
	}
	m.SetKeys(keys)

	// Initial cursor at 0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor at 0, got %d", m.Cursor())
	}

	// Move down
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", m.Cursor())
	}

	// Move down again
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor at 2, got %d", m.Cursor())
	}

	// Move up
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", m.Cursor())
	}
}

func TestModel_NavigationBounds(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	keys := []models.KeyInfo{
		{Key: "a", Size: 100},
		{Key: "b", Size: 200},
	}
	m.SetKeys(keys)

	// Try to move up from top
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor() != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.Cursor())
	}

	// Move to bottom
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Try to move past bottom
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 1 {
		t.Errorf("cursor should stay at 1, got %d", m.Cursor())
	}
}

func TestModel_Selection(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list, no tree
	keys := []models.KeyInfo{
		{Key: "user1", Size: 100},
		{Key: "user2", Size: 200},
	}
	m.SetKeys(keys)

	// Select first item (which is a leaf)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected command on selection")
	}
}

func TestModel_Filter(t *testing.T) {
	m := keylist.NewModel()
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
		{Key: "session:abc", Size: 50},
		{Key: "cache:data", Size: 300},
	}
	m.SetKeys(keys)

	m.SetFilter("user")

	filtered := m.FilteredKeys()
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered keys, got %d", len(filtered))
	}
}

func TestModel_FilterEmpty(t *testing.T) {
	m := keylist.NewModel()
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
	}
	m.SetKeys(keys)

	m.SetFilter("")

	filtered := m.FilteredKeys()
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered keys (empty filter shows all), got %d", len(filtered))
	}
}

func TestModel_SelectedKey(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list, no tree
	keys := []models.KeyInfo{
		{Key: "user1", Size: 100},
		{Key: "user2", Size: 200},
	}
	m.SetKeys(keys)

	selected := m.SelectedKey()
	if selected == nil {
		t.Fatal("expected selected key")
	}
	if selected.Key != "user1" {
		t.Errorf("expected 'user1', got '%s'", selected.Key)
	}

	// Move down and check again
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected = m.SelectedKey()
	if selected.Key != "user2" {
		t.Errorf("expected 'user2', got '%s'", selected.Key)
	}
}

func TestModel_View(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	keys := []models.KeyInfo{
		{Key: "user:1", Size: 100},
		{Key: "user:2", Size: 200},
	}
	m.SetKeys(keys)

	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestModel_EmptyView(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)

	view := m.View()
	// Should show some message for empty state
	if view == "" {
		t.Error("view should show empty state message")
	}
}

func TestTreeNode_Basic(t *testing.T) {
	node := keylist.NewTreeNode("test", false)
	if node.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", node.Name)
	}
	if node.IsLeaf {
		t.Error("expected non-leaf node")
	}
}

func TestTreeNode_Leaf(t *testing.T) {
	ki := &models.KeyInfo{Key: "user:1", Size: 100}
	node := keylist.NewLeafNode(ki)

	if !node.IsLeaf {
		t.Error("expected leaf node")
	}
	if node.KeyInfo != ki {
		t.Error("KeyInfo should be set")
	}
}

// Multi-select tests

func TestModel_ToggleSelection(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
		{Key: "key2", Size: 200},
		{Key: "key3", Size: 300},
	}
	m.SetKeys(keys)

	// Initially no selection
	if m.SelectionCount() != 0 {
		t.Errorf("expected 0 selections, got %d", m.SelectionCount())
	}

	// Toggle selection on first item
	m.ToggleSelection()
	if m.SelectionCount() != 1 {
		t.Errorf("expected 1 selection, got %d", m.SelectionCount())
	}
	if !m.IsSelected("key1") {
		t.Error("expected key1 to be selected")
	}

	// Toggle off
	m.ToggleSelection()
	if m.SelectionCount() != 0 {
		t.Errorf("expected 0 selections, got %d", m.SelectionCount())
	}
	if m.IsSelected("key1") {
		t.Error("expected key1 to be deselected")
	}
}

func TestModel_ToggleSelectionViaSpace(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
		{Key: "key2", Size: 200},
	}
	m.SetKeys(keys)

	// Press space to toggle selection
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.SelectionCount() != 1 {
		t.Errorf("expected 1 selection after space, got %d", m.SelectionCount())
	}

	// Move down and select another
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.SelectionCount() != 2 {
		t.Errorf("expected 2 selections, got %d", m.SelectionCount())
	}
}

func TestModel_SelectedKeys(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
		{Key: "key2", Size: 200},
		{Key: "key3", Size: 300},
	}
	m.SetKeys(keys)

	// Select first and third
	m.ToggleSelection() // key1
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.ToggleSelection() // key3

	selected := m.SelectedKeys()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected keys, got %d", len(selected))
	}

	// Check that both keys are in the list (order not guaranteed)
	hasKey1, hasKey3 := false, false
	for _, k := range selected {
		if k == "key1" {
			hasKey1 = true
		}
		if k == "key3" {
			hasKey3 = true
		}
	}
	if !hasKey1 || !hasKey3 {
		t.Errorf("expected key1 and key3 in selected keys, got %v", selected)
	}
}

func TestModel_ClearSelection(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
		{Key: "key2", Size: 200},
	}
	m.SetKeys(keys)

	// Select both
	m.ToggleSelection()
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.ToggleSelection()

	if m.SelectionCount() != 2 {
		t.Errorf("expected 2 selections, got %d", m.SelectionCount())
	}

	// Clear
	m.ClearSelection()
	if m.SelectionCount() != 0 {
		t.Errorf("expected 0 selections after clear, got %d", m.SelectionCount())
	}
}

func TestModel_SelectAll(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
		{Key: "key2", Size: 200},
		{Key: "key3", Size: 300},
	}
	m.SetKeys(keys)

	m.SelectAll()

	if m.SelectionCount() != 3 {
		t.Errorf("expected 3 selections, got %d", m.SelectionCount())
	}
	if !m.IsSelected("key1") || !m.IsSelected("key2") || !m.IsSelected("key3") {
		t.Error("expected all keys to be selected")
	}
}

func TestModel_HasSelection(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
	}
	m.SetKeys(keys)

	if m.HasSelection() {
		t.Error("expected no selection initially")
	}

	m.ToggleSelection()
	if !m.HasSelection() {
		t.Error("expected selection after toggle")
	}
}

func TestModel_MultiSelectMode(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
	}
	m.SetKeys(keys)

	// Enable multi-select mode
	m.SetMultiSelectMode(true)
	if !m.IsMultiSelectMode() {
		t.Error("expected multi-select mode to be enabled")
	}

	// Add selection
	m.ToggleSelection()
	if m.SelectionCount() != 1 {
		t.Errorf("expected 1 selection, got %d", m.SelectionCount())
	}

	// Disable multi-select mode should clear selection
	m.SetMultiSelectMode(false)
	if m.IsMultiSelectMode() {
		t.Error("expected multi-select mode to be disabled")
	}
	if m.SelectionCount() != 0 {
		t.Errorf("expected 0 selections after disabling mode, got %d", m.SelectionCount())
	}
}

func TestModel_ToggleSelectionOnFolder(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	keys := []models.KeyInfo{
		{Key: "user:profile:1", Size: 100},
		{Key: "user:profile:2", Size: 200},
	}
	m.SetKeys(keys)
	m.SetDelimiter(":")

	// Cursor should be on "user" folder
	// Toggling should do nothing for folders
	m.ToggleSelection()
	if m.SelectionCount() != 0 {
		t.Errorf("expected 0 selections on folder toggle, got %d", m.SelectionCount())
	}
}

func TestModel_ViewWithSelection(t *testing.T) {
	m := keylist.NewModel()
	m.SetSize(40, 20)
	m.SetDelimiter("") // Flat list
	keys := []models.KeyInfo{
		{Key: "key1", Size: 100},
		{Key: "key2", Size: 200},
	}
	m.SetKeys(keys)

	// Select first key
	m.ToggleSelection()

	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
	// View should contain selection markers
	if !contains(view, "[x]") {
		t.Error("view should contain [x] for selected item")
	}
	if !contains(view, "[ ]") {
		t.Error("view should contain [ ] for unselected item")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
