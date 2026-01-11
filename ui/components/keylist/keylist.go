package keylist

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nnnkkk7/memtui/models"
)

// TreeNode represents a node in the key tree
type TreeNode struct {
	Name     string
	IsLeaf   bool
	Expanded bool
	KeyInfo  *models.KeyInfo
	Children []*TreeNode
	Parent   *TreeNode
	depth    int
}

// NewTreeNode creates a new non-leaf tree node
func NewTreeNode(name string, expanded bool) *TreeNode {
	return &TreeNode{
		Name:     name,
		IsLeaf:   false,
		Expanded: expanded,
		Children: make([]*TreeNode, 0),
	}
}

// NewLeafNode creates a new leaf tree node
func NewLeafNode(ki *models.KeyInfo) *TreeNode {
	return &TreeNode{
		Name:    ki.Key,
		IsLeaf:  true,
		KeyInfo: ki,
	}
}

// KeySelectedMsg is sent when a key is selected
type KeySelectedMsg struct {
	Key models.KeyInfo
}

// Model represents the key list component
type Model struct {
	keys      []models.KeyInfo
	filtered  []models.KeyInfo
	filter    string
	delimiter string
	tree      *TreeNode
	cursor    int
	offset    int
	width     int
	height    int
	flatNodes []*TreeNode // Flattened visible nodes for navigation

	// Multi-select support
	selected    map[string]bool // Map of selected key names
	multiSelect bool            // Whether multi-select mode is enabled

	// Styles
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	folderStyle   lipgloss.Style
	leafStyle     lipgloss.Style
	markedStyle   lipgloss.Style // Style for multi-selected items
}

// NewModel creates a new key list model
func NewModel() *Model {
	return &Model{
		delimiter: ":",
		tree:      NewTreeNode("root", true),
		selected:  make(map[string]bool),
		normalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true),
		folderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		leafStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		markedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("136")).
			Foreground(lipgloss.Color("230")).
			Bold(true),
	}
}

// SetKeys sets the keys and rebuilds the tree
func (m *Model) SetKeys(keys []models.KeyInfo) {
	m.keys = keys
	m.filtered = keys
	m.rebuildTree()
}

// SetDelimiter sets the delimiter for tree building
func (m *Model) SetDelimiter(d string) {
	m.delimiter = d
	m.rebuildTree()
}

// SetFilter sets the filter pattern
func (m *Model) SetFilter(pattern string) {
	m.filter = pattern
	if pattern == "" {
		m.filtered = m.keys
	} else {
		m.filtered = nil
		for _, ki := range m.keys {
			if strings.Contains(ki.Key, pattern) {
				m.filtered = append(m.filtered, ki)
			}
		}
	}
	m.rebuildTree()
}

// SetSize sets the component size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// KeyCount returns the number of keys
func (m *Model) KeyCount() int {
	return len(m.keys)
}

// Cursor returns the current cursor position
func (m *Model) Cursor() int {
	return m.cursor
}

// Tree returns the tree root
func (m *Model) Tree() *TreeNode {
	return m.tree
}

// FilteredKeys returns the filtered keys
func (m *Model) FilteredKeys() []models.KeyInfo {
	return m.filtered
}

// SelectedKey returns the currently selected key
func (m *Model) SelectedKey() *models.KeyInfo {
	if m.cursor < 0 || m.cursor >= len(m.flatNodes) {
		return nil
	}
	node := m.flatNodes[m.cursor]
	return node.KeyInfo
}

// Multi-select methods

// ToggleSelection toggles the selection state of the current cursor item
func (m *Model) ToggleSelection() {
	if m.cursor < 0 || m.cursor >= len(m.flatNodes) {
		return
	}
	node := m.flatNodes[m.cursor]
	if !node.IsLeaf || node.KeyInfo == nil {
		return
	}

	key := node.KeyInfo.Key
	if m.selected[key] {
		delete(m.selected, key)
	} else {
		m.selected[key] = true
	}
}

// IsSelected returns true if the given key is selected
func (m *Model) IsSelected(key string) bool {
	return m.selected[key]
}

// SelectionCount returns the number of selected items
func (m *Model) SelectionCount() int {
	return len(m.selected)
}

// SelectedKeys returns a list of all selected key names
func (m *Model) SelectedKeys() []string {
	keys := make([]string, 0, len(m.selected))
	for key := range m.selected {
		keys = append(keys, key)
	}
	return keys
}

// ClearSelection clears all selections
func (m *Model) ClearSelection() {
	m.selected = make(map[string]bool)
}

// SelectAll selects all visible leaf keys
func (m *Model) SelectAll() {
	for _, node := range m.flatNodes {
		if node.IsLeaf && node.KeyInfo != nil {
			m.selected[node.KeyInfo.Key] = true
		}
	}
}

// HasSelection returns true if any items are selected
func (m *Model) HasSelection() bool {
	return len(m.selected) > 0
}

// SetMultiSelectMode enables or disables multi-select mode
func (m *Model) SetMultiSelectMode(enabled bool) {
	m.multiSelect = enabled
	if !enabled {
		m.ClearSelection()
	}
}

// IsMultiSelectMode returns true if multi-select mode is enabled
func (m *Model) IsMultiSelectMode() bool {
	return m.multiSelect
}

// rebuildTree rebuilds the tree from filtered keys
func (m *Model) rebuildTree() {
	m.tree = NewTreeNode("root", true)

	for i := range m.filtered {
		m.insertKey(&m.filtered[i])
	}

	m.flattenTree()
}

// insertKey inserts a key into the tree
func (m *Model) insertKey(ki *models.KeyInfo) {
	if m.delimiter == "" || !strings.Contains(ki.Key, m.delimiter) {
		// No delimiter or key has no delimiter, flat list
		leaf := NewLeafNode(ki)
		leaf.Name = ki.Key
		leaf.Parent = m.tree
		leaf.depth = 0
		m.tree.Children = append(m.tree.Children, leaf)
		return
	}

	parts := strings.Split(ki.Key, m.delimiter)
	current := m.tree

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - create leaf
			leaf := NewLeafNode(ki)
			leaf.Name = part
			leaf.Parent = current
			leaf.depth = i + 1
			current.Children = append(current.Children, leaf)
		} else {
			// Find or create folder node
			var found *TreeNode
			for _, child := range current.Children {
				if !child.IsLeaf && child.Name == part {
					found = child
					break
				}
			}
			if found == nil {
				found = NewTreeNode(part, true) // Default expanded
				found.Parent = current
				found.depth = i + 1
				current.Children = append(current.Children, found)
			}
			current = found
		}
	}
}

// flattenTree flattens the tree for navigation
func (m *Model) flattenTree() {
	m.flatNodes = nil
	m.flattenNode(m.tree)
}

func (m *Model) flattenNode(node *TreeNode) {
	for _, child := range node.Children {
		m.flatNodes = append(m.flatNodes, child)
		if !child.IsLeaf && child.Expanded {
			m.flattenNode(child)
		}
	}
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle vim-style navigation (j/k)
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'j':
				if m.cursor < len(m.flatNodes)-1 {
					m.cursor++
				}
				return m, nil
			case 'k':
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			}
		}

		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
			}
		case tea.KeySpace:
			// Toggle selection of current item (multi-select)
			m.ToggleSelection()
		case tea.KeyEnter:
			if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if node.IsLeaf && node.KeyInfo != nil {
					return m, func() tea.Msg {
						return KeySelectedMsg{Key: *node.KeyInfo}
					}
				} else {
					// Toggle folder
					node.Expanded = !node.Expanded
					m.flattenTree()
				}
			}
		case tea.KeyRight:
			// Expand folder
			if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if !node.IsLeaf && !node.Expanded {
					node.Expanded = true
					m.flattenTree()
				}
			}
		case tea.KeyLeft:
			// Collapse folder or go to parent
			if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if !node.IsLeaf && node.Expanded {
					node.Expanded = false
					m.flattenTree()
				} else if node.Parent != nil && node.Parent != m.tree {
					// Find parent in flat list
					for i, n := range m.flatNodes {
						if n == node.Parent {
							m.cursor = i
							break
						}
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the key list
func (m *Model) View() string {
	if len(m.flatNodes) == 0 {
		return m.normalStyle.Render("No keys")
	}

	var b strings.Builder

	// Calculate visible range
	visibleHeight := m.height
	if visibleHeight <= 0 {
		visibleHeight = 20
	}

	// Adjust offset to keep cursor visible
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+visibleHeight {
		m.offset = m.cursor - visibleHeight + 1
	}

	// Render visible nodes
	for i := m.offset; i < len(m.flatNodes) && i < m.offset+visibleHeight; i++ {
		node := m.flatNodes[i]
		line := m.renderNode(node, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) renderNode(node *TreeNode, cursorOnThis bool) string {
	indent := strings.Repeat("  ", node.depth)

	// Determine if this node is multi-selected (marked for batch operation)
	isMarked := false
	if node.IsLeaf && node.KeyInfo != nil {
		isMarked = m.selected[node.KeyInfo.Key]
	}

	// Build prefix with selection marker
	var prefix string
	if node.IsLeaf {
		if isMarked {
			prefix = "[x] "
		} else {
			prefix = "[ ] "
		}
	} else if node.Expanded {
		prefix = "▼ "
	} else {
		prefix = "▶ "
	}

	name := node.Name
	if node.IsLeaf && node.KeyInfo != nil {
		name = node.Name
	}

	line := indent + prefix + name

	// Truncate if needed
	if m.width > 0 && len(line) > m.width {
		line = line[:m.width-3] + "..."
	}

	// Apply styles based on cursor position and selection state
	if cursorOnThis && isMarked {
		// Both cursor and marked - use a combined style
		return m.selectedStyle.Render(line)
	} else if cursorOnThis {
		return m.selectedStyle.Render(line)
	} else if isMarked {
		return m.markedStyle.Render(line)
	}

	if node.IsLeaf {
		return m.leafStyle.Render(line)
	}
	return m.folderStyle.Render(line)
}
