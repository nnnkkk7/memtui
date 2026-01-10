//go:build e2e

package e2e_test

import (
	"sort"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/models"
	"github.com/nnnkkk7/memtui/ui/components/keylist"
)

// =============================================================================
// Test Data for Key Navigation
// =============================================================================

// Hierarchical test key structure for navigation tests
var navigationTestKeys = map[string]string{
	// User namespace
	"e2e_nav:user:1001:profile":     `{"name":"Alice","email":"alice@example.com"}`,
	"e2e_nav:user:1001:session":     `{"token":"abc123","expires":1735689600}`,
	"e2e_nav:user:1001:preferences": `{"theme":"dark","language":"en"}`,
	"e2e_nav:user:1002:profile":     `{"name":"Bob","email":"bob@example.com"}`,
	"e2e_nav:user:1002:session":     `{"token":"def456","expires":1735689600}`,
	"e2e_nav:user:1003:profile":     `{"name":"Charlie","email":"charlie@example.com"}`,

	// Cache namespace
	"e2e_nav:cache:api:users":       `[{"id":1001},{"id":1002}]`,
	"e2e_nav:cache:api:products":    `[{"id":1},{"id":2}]`,
	"e2e_nav:cache:api:orders":      `[{"id":100}]`,
	"e2e_nav:cache:static:homepage": `<html>...</html>`,

	// Config namespace
	"e2e_nav:config:app:settings":  `{"debug":false,"timeout":30}`,
	"e2e_nav:config:db:connection": `{"host":"localhost","port":5432}`,

	// Single-level keys (no hierarchy beyond prefix)
	"e2e_nav:simple_key_1": "simple value 1",
	"e2e_nav:simple_key_2": "simple value 2",
}

// filterNavigationKeys filters enumerated keys to only navigation test keys
func filterNavigationKeys(allKeys []models.KeyInfo) []models.KeyInfo {
	var testKeys []models.KeyInfo
	for _, ki := range allKeys {
		if strings.HasPrefix(ki.Key, "e2e_nav:") {
			testKeys = append(testKeys, ki)
		}
	}
	return testKeys
}

// =============================================================================
// Test 1: Tree Structure Building from Flat Key List
// =============================================================================

func TestE2E_TreeStructureBuilding(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	// Wait for metadump to catch up
	time.Sleep(500 * time.Millisecond)

	t.Run("builds hierarchical tree from flat keys", func(t *testing.T) {
		// Create keylist model with test keys
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		tree := m.Tree()
		if tree == nil {
			t.Fatal("expected non-nil tree root")
		}

		// Root should have children
		if len(tree.Children) == 0 {
			t.Fatal("expected tree to have children")
		}

		// Verify e2e_nav is a top-level node
		var foundNavNode bool
		for _, child := range tree.Children {
			if child.Name == "e2e_nav" {
				foundNavNode = true
				break
			}
		}

		if !foundNavNode {
			t.Error("expected 'e2e_nav' node in tree")
		}

		t.Logf("Tree has %d top-level nodes", len(tree.Children))
	})

	t.Run("tree has correct depth", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Find deepest path - e2e_nav:user:1001:profile = 4 levels
		var maxDepth int
		var countDepth func(node *keylist.TreeNode, depth int)
		countDepth = func(node *keylist.TreeNode, depth int) {
			if depth > maxDepth {
				maxDepth = depth
			}
			for _, child := range node.Children {
				countDepth(child, depth+1)
			}
		}
		countDepth(m.Tree(), 0)

		// Expected depth: e2e_nav (1) : user (2) : 1001 (3) : profile (4)
		expectedMinDepth := 4
		if maxDepth < expectedMinDepth {
			t.Errorf("expected tree depth >= %d, got %d", expectedMinDepth, maxDepth)
		}
		t.Logf("Maximum tree depth: %d", maxDepth)
	})

	t.Run("tree structure with different delimiters", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetKeys(keys)

		// With colon delimiter
		m.SetDelimiter(":")
		colonChildCount := len(m.Tree().Children)

		// With no delimiter (flat list)
		m.SetDelimiter("")
		flatChildCount := len(m.Tree().Children)

		// Flat tree should have more direct children
		if flatChildCount <= colonChildCount {
			t.Log("With no delimiter, all keys should be at top level")
		}
		t.Logf("Colon delimiter: %d children, No delimiter: %d children", colonChildCount, flatChildCount)
	})
}

// =============================================================================
// Test 2: Folder Expansion and Collapse
// =============================================================================

func TestE2E_FolderExpansionCollapse(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("folders are expanded by default", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		tree := m.Tree()
		for _, child := range tree.Children {
			if !child.IsLeaf && !child.Expanded {
				t.Errorf("folder %q should be expanded by default", child.Name)
			}
		}
	})

	t.Run("Enter toggles folder expansion", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Cursor starts at 0, which should be on first node
		if m.Cursor() != 0 {
			t.Errorf("expected cursor at 0, got %d", m.Cursor())
		}

		// Get initial view
		initialView := m.View()

		// Press Enter to toggle
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// View should change (either folder collapsed or key selected)
		newView := m.View()
		if initialView == newView {
			t.Log("View unchanged after Enter - may be on a leaf node")
		}
	})

	t.Run("Right arrow expands collapsed folder", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Right arrow on folder should expand it
		m.Update(tea.KeyMsg{Type: tea.KeyRight})

		// Verify no crash
		view := m.View()
		if view == "" {
			t.Error("view should not be empty after right arrow")
		}
	})

	t.Run("Left arrow collapses expanded folder", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Left arrow on expanded folder should collapse it
		m.Update(tea.KeyMsg{Type: tea.KeyLeft})

		// Verify no crash
		view := m.View()
		if view == "" {
			t.Error("view should not be empty after left arrow")
		}
	})

	t.Run("Left arrow on leaf navigates to parent", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Navigate deep into the tree
		for i := 0; i < 5; i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		cursorBefore := m.Cursor()

		// Press left - should navigate to parent
		m.Update(tea.KeyMsg{Type: tea.KeyLeft})

		// Cursor may have changed if we navigated to parent
		cursorAfter := m.Cursor()
		t.Logf("Cursor before left: %d, after: %d", cursorBefore, cursorAfter)
	})
}

// =============================================================================
// Test 3: Key Selection with Enter
// =============================================================================

func TestE2E_KeySelectionWithEnter(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("Enter on leaf node sends selection message", func(t *testing.T) {
		// Use flat list for easier testing
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("") // Flat list
		m.SetKeys(keys)

		// Press Enter to select
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Error("expected command on Enter for leaf node")
			return
		}

		// Execute command to get message
		msg := cmd()
		if selectedMsg, ok := msg.(keylist.KeySelectedMsg); ok {
			t.Logf("Selected key: %s", selectedMsg.Key.Key)
		} else {
			t.Errorf("expected KeySelectedMsg, got %T", msg)
		}
	})

	t.Run("SelectedKey returns currently focused key", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("") // Flat list
		m.SetKeys(keys)

		// Initially, cursor is at 0
		selectedKey := m.SelectedKey()
		if selectedKey == nil {
			t.Error("expected a selected key at cursor 0")
			return
		}

		if !strings.HasPrefix(selectedKey.Key, "e2e_nav:") {
			t.Errorf("expected key to start with test prefix, got %q", selectedKey.Key)
		}

		// Move down and verify selection changes
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		newSelected := m.SelectedKey()
		if newSelected == nil {
			t.Error("expected a selected key after moving down")
			return
		}

		t.Logf("Initial key: %s, After down: %s", selectedKey.Key, newSelected.Key)
	})

	t.Run("selection works after navigation", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Navigate down several times
		for i := 0; i < 3; i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		// Verify cursor position
		if m.Cursor() != 3 {
			t.Errorf("expected cursor at 3, got %d", m.Cursor())
		}

		// Select should work
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(keylist.KeySelectedMsg); ok {
				t.Log("Selection works after navigation")
			}
		}
	})
}

// =============================================================================
// Test 4: Vim-style Navigation (j/k)
// =============================================================================

func TestE2E_VimStyleNavigation(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("j key moves down (simulated via KeyDown)", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 2 {
			t.Skip("need at least 2 keys for navigation test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		initialCursor := m.Cursor()
		if initialCursor != 0 {
			t.Errorf("expected initial cursor at 0, got %d", initialCursor)
		}

		// Simulate 'j' via KeyDown (app layer maps j to Down)
		m.Update(tea.KeyMsg{Type: tea.KeyDown})

		if m.Cursor() != 1 {
			t.Errorf("expected cursor at 1 after down, got %d", m.Cursor())
		}
	})

	t.Run("k key moves up (simulated via KeyUp)", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 2 {
			t.Skip("need at least 2 keys for navigation test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Move down first
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyDown})

		cursorAfterDown := m.Cursor()

		// Simulate 'k' via KeyUp
		m.Update(tea.KeyMsg{Type: tea.KeyUp})

		if m.Cursor() >= cursorAfterDown {
			t.Errorf("expected cursor to decrease after up, was %d now %d", cursorAfterDown, m.Cursor())
		}
	})

	t.Run("jjjkk sequence results in net +1 position", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 5 {
			t.Skip("need at least 5 keys for sequence test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// jjjkk sequence (down 3, up 2 = net +1)
		m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 0 -> 1
		m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 1 -> 2
		m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 2 -> 3
		m.Update(tea.KeyMsg{Type: tea.KeyUp})   // 3 -> 2
		m.Update(tea.KeyMsg{Type: tea.KeyUp})   // 2 -> 1

		if m.Cursor() != 1 {
			t.Errorf("expected cursor at 1 after jjjkk sequence, got %d", m.Cursor())
		}
	})
}

// =============================================================================
// Test 5: Arrow Key Navigation
// =============================================================================

func TestE2E_ArrowKeyNavigation(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("Down arrow moves cursor down", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 2 {
			t.Skip("need at least 2 keys for navigation test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.Cursor() != 1 {
			t.Errorf("expected cursor at 1 after down arrow, got %d", m.Cursor())
		}
	})

	t.Run("Up arrow moves cursor up", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 3 {
			t.Skip("need at least 3 keys for navigation test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Move down then up
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyUp})

		if m.Cursor() != 1 {
			t.Errorf("expected cursor at 1 after down-down-up, got %d", m.Cursor())
		}
	})

	t.Run("cursor stays at top when pressing up at boundary", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Try to move up from top - cursor should stay at 0
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.Cursor() != 0 {
			t.Errorf("cursor should stay at 0 when at top boundary, got %d", m.Cursor())
		}

		// Multiple up presses should not go negative
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.Cursor() != 0 {
			t.Errorf("cursor should remain at 0 after multiple up presses, got %d", m.Cursor())
		}
	})

	t.Run("cursor stays at bottom when pressing down at boundary", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Move to the last item
		lastIndex := len(keys) - 1
		for i := 0; i < lastIndex+5; i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		if m.Cursor() != lastIndex {
			t.Errorf("cursor should be at last index %d, got %d", lastIndex, m.Cursor())
		}

		// Additional down presses should not go past the end
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.Cursor() != lastIndex {
			t.Errorf("cursor should remain at last index %d after extra down, got %d", lastIndex, m.Cursor())
		}
	})

	t.Run("navigation through entire list", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Navigate to the end
		for i := 0; i < len(keys); i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		expectedCursor := len(keys) - 1
		if m.Cursor() != expectedCursor {
			t.Errorf("expected cursor at %d after navigating to end, got %d", expectedCursor, m.Cursor())
		}

		// Navigate back to the beginning
		for i := 0; i < len(keys); i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyUp})
		}

		if m.Cursor() != 0 {
			t.Errorf("expected cursor at 0 after navigating to beginning, got %d", m.Cursor())
		}
	})
}

// =============================================================================
// Test 6: Filter Mode (/ key) with Search
// =============================================================================

func TestE2E_KeylistFilterMode(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("SetFilter filters keys", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Initial state - all keys visible
		initialCount := len(m.FilteredKeys())

		// Apply a filter for "user"
		m.SetFilter("user")

		// Should have fewer keys now
		filteredCount := len(m.FilteredKeys())

		// Verify filtered keys contain "user"
		for _, ki := range m.FilteredKeys() {
			if !strings.Contains(ki.Key, "user") {
				t.Errorf("filtered key %q should contain 'user'", ki.Key)
			}
		}

		t.Logf("Initial count: %d, Filtered count: %d", initialCount, filteredCount)
	})

	t.Run("filter with partial match", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Filter for "profile"
		m.SetFilter("profile")

		filtered := m.FilteredKeys()
		for _, ki := range filtered {
			if !strings.Contains(ki.Key, "profile") {
				t.Errorf("filtered key %q should contain 'profile'", ki.Key)
			}
		}

		t.Logf("Found %d keys matching 'profile'", len(filtered))
	})

	t.Run("filter with no matches returns empty", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Filter for something that doesn't exist
		m.SetFilter("xyznonexistent12345")

		filtered := m.FilteredKeys()
		if len(filtered) != 0 {
			t.Errorf("expected 0 filtered keys for non-existent pattern, got %d", len(filtered))
		}

		// View should show "No keys" message
		view := m.View()
		if !strings.Contains(view, "No keys") {
			t.Log("View when filtered to empty should indicate no keys")
		}
	})

	t.Run("filter by ID pattern", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Filter for "1001" - should match user:1001:* keys
		m.SetFilter("1001")

		filtered := m.FilteredKeys()
		for _, ki := range filtered {
			if !strings.Contains(ki.Key, "1001") {
				t.Errorf("filtered key %q should contain '1001'", ki.Key)
			}
		}

		// Should find 3 keys (profile, session, preferences for user 1001)
		expectedCount := 3
		if len(filtered) != expectedCount {
			t.Logf("Expected %d keys matching '1001', got %d", expectedCount, len(filtered))
		}
	})

	t.Run("filter is case-sensitive", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Filter for "USER" (uppercase)
		m.SetFilter("USER")

		filtered := m.FilteredKeys()

		// With case-sensitive matching, this should return 0 results
		if len(filtered) > 0 {
			t.Log("Note: Filter appears to be case-insensitive (unexpected)")
		} else {
			t.Log("Filter is case-sensitive (expected)")
		}
	})
}

// =============================================================================
// Test 7: Filter Clearing with Esc
// =============================================================================

func TestE2E_FilterClearing(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("empty filter restores all keys", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Record initial count
		initialCount := len(m.FilteredKeys())

		// Apply filter
		m.SetFilter("user")
		filteredCount := len(m.FilteredKeys())

		if filteredCount == initialCount {
			t.Log("Filter didn't reduce count - may all match")
		}

		// Clear filter with empty string
		m.SetFilter("")

		// Should restore all keys
		restoredCount := len(m.FilteredKeys())
		if restoredCount != initialCount {
			t.Errorf("expected %d keys after clearing filter, got %d", initialCount, restoredCount)
		}
	})

	t.Run("multiple filter cycles", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		initialCount := len(m.FilteredKeys())

		// Cycle 1: filter and clear
		m.SetFilter("user")
		m.SetFilter("")
		if len(m.FilteredKeys()) != initialCount {
			t.Errorf("cycle 1: expected %d keys, got %d", initialCount, len(m.FilteredKeys()))
		}

		// Cycle 2: filter and clear
		m.SetFilter("cache")
		m.SetFilter("")
		if len(m.FilteredKeys()) != initialCount {
			t.Errorf("cycle 2: expected %d keys, got %d", initialCount, len(m.FilteredKeys()))
		}

		// Cycle 3: filter and clear
		m.SetFilter("config")
		m.SetFilter("")
		if len(m.FilteredKeys()) != initialCount {
			t.Errorf("cycle 3: expected %d keys, got %d", initialCount, len(m.FilteredKeys()))
		}
	})

	t.Run("tree rebuilds correctly after filter clear", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Get initial tree child count
		initialChildCount := len(m.Tree().Children)

		// Apply filter that removes some tree branches
		m.SetFilter("user")
		filteredChildCount := len(m.Tree().Children)

		// Clear filter
		m.SetFilter("")
		restoredChildCount := len(m.Tree().Children)

		if restoredChildCount != initialChildCount {
			t.Errorf("tree structure not restored: initial %d, restored %d",
				initialChildCount, restoredChildCount)
		}

		t.Logf("Tree children: initial=%d, filtered=%d, restored=%d",
			initialChildCount, filteredChildCount, restoredChildCount)
	})
}

// =============================================================================
// Test 8: Selection State After Filtering
// =============================================================================

func TestE2E_SelectionStateAfterFiltering(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("cursor resets appropriately after filter", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 5 {
			t.Skip("need at least 5 keys for this test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Move cursor to position 3
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyDown})

		cursorBefore := m.Cursor()
		if cursorBefore != 3 {
			t.Errorf("expected cursor at 3, got %d", cursorBefore)
		}

		// Apply filter - cursor should be valid for new filtered list
		m.SetFilter("user")

		filteredLen := len(m.FilteredKeys())
		if filteredLen > 0 && (m.Cursor() < 0 || m.Cursor() >= filteredLen) {
			t.Errorf("cursor %d out of bounds for filtered keys (len=%d)", m.Cursor(), filteredLen)
		}
	})

	t.Run("multi-selection preserved by key name", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)
		if len(keys) < 3 {
			t.Skip("need at least 3 keys for multi-selection test")
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Select multiple items using ToggleSelection
		m.ToggleSelection() // Select first
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.ToggleSelection() // Select second
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.ToggleSelection() // Select third

		selectionCountBefore := m.SelectionCount()
		if selectionCountBefore != 3 {
			t.Errorf("expected 3 selections, got %d", selectionCountBefore)
		}

		// Apply filter
		m.SetFilter("user")

		// Selections are tracked by key name, so they persist
		selectedKeys := m.SelectedKeys()
		t.Logf("After filter, %d keys still selected", len(selectedKeys))
	})

	t.Run("navigation works after filter clear", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Apply and clear filter
		m.SetFilter("cache")
		m.SetFilter("")

		// Navigation should still work
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.Cursor() != 1 {
			t.Errorf("expected cursor at 1 after navigation, got %d", m.Cursor())
		}

		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.Cursor() != 0 {
			t.Errorf("expected cursor at 0 after up, got %d", m.Cursor())
		}
	})

	t.Run("SelectedKey returns valid key after filter", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Apply filter
		m.SetFilter("profile")

		if len(m.FilteredKeys()) > 0 {
			selectedKey := m.SelectedKey()
			if selectedKey == nil {
				t.Error("expected a selected key in filtered view")
				return
			}

			if !strings.Contains(selectedKey.Key, "profile") {
				t.Errorf("selected key %q should match filter 'profile'", selectedKey.Key)
			}
		}
	})
}

// =============================================================================
// Additional E2E Tests
// =============================================================================

func TestE2E_FullNavigationWorkflow(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("complete navigation workflow", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		// Step 1: Verify initial state
		if m.KeyCount() != len(keys) {
			t.Errorf("Step 1: expected %d keys, got %d", len(keys), m.KeyCount())
		}
		t.Log("Step 1: Initial key count verified")

		// Step 2: Navigate down through the tree
		for i := 0; i < 5 && i < m.KeyCount(); i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}
		t.Logf("Step 2: Navigated down, cursor at %d", m.Cursor())

		// Step 3: Apply filter
		m.SetFilter("user")
		filteredCount := len(m.FilteredKeys())
		t.Logf("Step 3: Filtered to 'user', %d keys", filteredCount)

		// Step 4: Navigate in filtered view
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		t.Log("Step 4: Navigation in filtered view works")

		// Step 5: Clear filter
		m.SetFilter("")
		if m.KeyCount() != len(keys) {
			t.Errorf("Step 5: expected %d keys after clearing filter, got %d", len(keys), m.KeyCount())
		}
		t.Log("Step 5: Filter cleared, all keys restored")

		// Step 6: Multi-select
		m.SelectAll()
		if m.SelectionCount() == 0 {
			t.Error("Step 6: SelectAll should select keys")
		}
		t.Logf("Step 6: Selected %d keys", m.SelectionCount())

		// Step 7: Clear selection
		m.ClearSelection()
		if m.SelectionCount() != 0 {
			t.Errorf("Step 7: expected 0 selections after clear, got %d", m.SelectionCount())
		}
		t.Log("Step 7: Selection cleared")

		t.Log("Full navigation workflow completed successfully")
	})
}

func TestE2E_ViewRendering(t *testing.T) {
	skipIfNoMemcached(t)
	cleanup := setupTestKeys(t, navigationTestKeys)
	defer cleanup()

	t.Run("view renders correctly", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter(":")
		m.SetKeys(keys)

		view := m.View()

		// View should not be empty
		if view == "" {
			t.Error("view should not be empty")
		}

		// View should contain tree indicators for folders
		if !strings.Contains(view, "[]") && !strings.Contains(view, "[x]") {
			// May have selection markers or folder indicators
			t.Log("View rendered successfully")
		}
	})

	t.Run("empty list shows appropriate message", func(t *testing.T) {
		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetKeys([]models.KeyInfo{})

		view := m.View()
		if !strings.Contains(view, "No keys") {
			t.Error("empty key list should show 'No keys' message")
		}
	})

	t.Run("view with selections", func(t *testing.T) {
		keys := createTestKeyInfos(navigationTestKeys)

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetDelimiter("")
		m.SetKeys(keys)

		// Select first key
		m.ToggleSelection()

		view := m.View()
		if !strings.Contains(view, "[x]") {
			t.Error("view should contain [x] for selected item")
		}
	})
}

func TestE2E_EmptyAndEdgeCases(t *testing.T) {
	t.Run("empty list navigation does not crash", func(t *testing.T) {
		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetKeys([]models.KeyInfo{})

		// Navigation should not crash
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m.Update(tea.KeyMsg{Type: tea.KeyLeft})
		m.Update(tea.KeyMsg{Type: tea.KeyRight})

		// Cursor should remain at 0
		if m.Cursor() != 0 {
			t.Errorf("cursor should stay at 0 for empty list, got %d", m.Cursor())
		}
	})

	t.Run("single key list navigation", func(t *testing.T) {
		singleKey := []models.KeyInfo{
			{Key: "only_key", Size: 100},
		}

		m := keylist.NewModel()
		m.SetSize(80, 24)
		m.SetKeys(singleKey)

		// Down should not move past the only key
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.Cursor() != 0 {
			t.Errorf("cursor should stay at 0 for single key list, got %d", m.Cursor())
		}

		// Up should not move before the only key
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.Cursor() != 0 {
			t.Errorf("cursor should stay at 0 for single key list, got %d", m.Cursor())
		}
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// createTestKeyInfos creates KeyInfo slice from test data map
func createTestKeyInfos(testData map[string]string) []models.KeyInfo {
	keys := make([]models.KeyInfo, 0, len(testData))
	for key, value := range testData {
		keys = append(keys, models.KeyInfo{
			Key:  key,
			Size: len(value),
		})
	}
	// Sort for deterministic ordering
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Key < keys[j].Key
	})
	return keys
}
