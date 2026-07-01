package tree

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/oakwood-commons/cobra-explorer/internal/theme"
)

// Model is the tree navigator TUI sub-model.
type Model struct {
	root     *CommandNode
	cursor   *CommandNode
	expanded map[string]bool
	scroll   int
	width    int
	height   int
	theme    theme.Theme
	focused  bool

	// Flattened visible nodes (recomputed on expand/collapse).
	visible []*CommandNode
}

// NewModel creates a tree navigator for the given command tree.
func NewModel(root *CommandNode, th theme.Theme) Model {
	m := Model{
		root:     root,
		cursor:   root,
		expanded: map[string]bool{pathKey(root): true},
		theme:    th,
	}
	m.recomputeVisible()
	return m
}

// SetSize updates the viewport dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	m.ensureCursorVisible()
	return m
}

// SetFocused updates whether this component has keyboard focus.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// Cursor returns the currently highlighted node.
func (m Model) Cursor() *CommandNode {
	return m.cursor
}

// TotalCount returns the total number of visible nodes.
func (m Model) TotalCount() int {
	return len(m.visible)
}

// ScrollOffset returns the current scroll position.
func (m Model) ScrollOffset() int {
	return m.scroll
}

// Height returns the viewport height.
func (m Model) Height() int {
	return m.height
}

// Update handles key events when the tree is focused.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		m.moveCursor(-1)
		return m, m.emitHighlight()
	case "down", "j":
		m.moveCursor(1)
		return m, m.emitHighlight()
	case "right", "l":
		m.expandOrDescend()
		return m, m.emitHighlight()
	case "left", "h":
		m.collapseOrAscend()
		return m, m.emitHighlight()
	case "enter":
		if m.cursor.Runnable {
			return m, func() tea.Msg {
				return CommandSelectedMsg{Node: m.cursor}
			}
		}
		m.expandOrDescend()
		return m, m.emitHighlight()
	case "home":
		if len(m.visible) > 0 {
			m.cursor = m.visible[0]
			m.scroll = 0
		}
		return m, m.emitHighlight()
	case "end":
		if len(m.visible) > 0 {
			m.cursor = m.visible[len(m.visible)-1]
			m.ensureCursorVisible()
		}
		return m, m.emitHighlight()
	}
	return m, nil
}

// View renders the tree panel.
func (m Model) View() string {
	if m.height <= 0 {
		return ""
	}

	var sb strings.Builder

	visibleEnd := m.scroll + m.height
	if visibleEnd > len(m.visible) {
		visibleEnd = len(m.visible)
	}

	for i := m.scroll; i < visibleEnd; i++ {
		node := m.visible[i]
		line := m.renderNode(node)
		sb.WriteString(line)
		if i < visibleEnd-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m Model) renderNode(node *CommandNode) string {
	indent := strings.Repeat("  ", node.Depth)

	var icon string
	if len(node.Children) > 0 {
		if m.expanded[pathKey(node)] {
			icon = "v "
		} else {
			icon = "> "
		}
	} else {
		icon = "  "
	}

	label := node.Name

	style := m.theme.TreeNormal
	switch {
	case node == m.cursor && m.focused:
		style = m.theme.TreeCursor
	case node == m.cursor:
		style = m.theme.TreeCursorUnfocused
	case node.Runnable:
		style = m.theme.TreeRunnable
	}

	if node.Deprecated != "" {
		style = style.Strikethrough(true)
	}
	if !node.Runnable && node.IsLeaf() {
		style = style.Faint(true)
	}

	suffix := ""
	if node.Runnable && node != m.cursor {
		suffix = " *"
	}

	rendered := style.Render(indent + icon + label + suffix)
	// Pad the row to full width using the theme background so trailing space
	// matches the panel surface instead of the terminal default.
	return lipgloss.NewStyle().
		Background(m.theme.Base.GetBackground()).
		Width(m.width).
		MaxWidth(m.width).
		Render(rendered)
}

func (m *Model) moveCursor(delta int) {
	idx := m.cursorIndex()
	newIdx := idx + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(m.visible) {
		newIdx = len(m.visible) - 1
	}
	if newIdx >= 0 && newIdx < len(m.visible) {
		m.cursor = m.visible[newIdx]
		m.ensureCursorVisible()
	}
}

func (m *Model) expandOrDescend() {
	if len(m.cursor.Children) > 0 {
		if !m.expanded[pathKey(m.cursor)] {
			m.expanded[pathKey(m.cursor)] = true
			m.recomputeVisible()
		}
		// Move to first child
		m.cursor = m.cursor.Children[0]
		m.ensureCursorVisible()
	}
}

func (m *Model) collapseOrAscend() {
	if m.expanded[pathKey(m.cursor)] && len(m.cursor.Children) > 0 {
		m.expanded[pathKey(m.cursor)] = false
		m.recomputeVisible()
	} else if m.cursor.Parent != nil {
		m.cursor = m.cursor.Parent
		m.ensureCursorVisible()
	}
}

func (m *Model) recomputeVisible() {
	m.visible = m.visible[:0]
	m.walkVisible(m.root)
}

func (m *Model) walkVisible(node *CommandNode) {
	m.visible = append(m.visible, node)
	if m.expanded[pathKey(node)] {
		for _, child := range node.Children {
			m.walkVisible(child)
		}
	}
}

func (m *Model) cursorIndex() int {
	for i, n := range m.visible {
		if n == m.cursor {
			return i
		}
	}
	return 0
}

func (m *Model) ensureCursorVisible() {
	if m.height <= 0 {
		return
	}
	idx := m.cursorIndex()
	if idx < m.scroll {
		m.scroll = idx
	}
	if idx >= m.scroll+m.height {
		m.scroll = idx - m.height + 1
	}
}

func (m Model) emitHighlight() tea.Cmd {
	node := m.cursor
	return func() tea.Msg {
		return CommandHighlightedMsg{Node: node}
	}
}

// SelectNode navigates the tree so the given node is the cursor.
// It expands ancestors as needed.
func (m Model) SelectNode(node *CommandNode) Model {
	// Expand all ancestors
	cur := node.Parent
	for cur != nil {
		m.expanded[pathKey(cur)] = true
		cur = cur.Parent
	}
	m.recomputeVisible()
	m.cursor = node
	m.ensureCursorVisible()
	return m
}

func pathKey(node *CommandNode) string {
	return strings.Join(node.FullPath, ".")
}
