package tree_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/theme"
	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// navRoot builds a small tree: root → [a → [a1, a2], b(runnable)].
func navRoot() *tree.CommandNode {
	root := &cobra.Command{Use: "root"}
	a := &cobra.Command{Use: "a", Short: "group a"}
	a1 := &cobra.Command{Use: "a1", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	a2 := &cobra.Command{Use: "a2", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	a.AddCommand(a1, a2)
	b := &cobra.Command{Use: "b", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	root.AddCommand(a, b)
	return tree.BuildTree(root, tree.BuildOptions{})
}

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestModel_InitialState(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	require.NotNil(t, m.Cursor())
	assert.Equal(t, "root", m.Cursor().Name)
	// Root is expanded by default → root + a + b visible.
	assert.Equal(t, 3, m.TotalCount())
	assert.Equal(t, 0, m.ScrollOffset())
}

func TestModel_MoveDownUp(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())

	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, "a", m.Cursor().Name)

	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, "b", m.Cursor().Name)

	// Can't move past the end.
	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, "b", m.Cursor().Name)

	m, _ = m.Update(keyMsg("k"))
	assert.Equal(t, "a", m.Cursor().Name)
}

func TestModel_ExpandCollapse(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())

	// Move to 'a' and expand it (right/l descends to first child).
	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, "a", m.Cursor().Name)

	m, _ = m.Update(keyMsg("l"))
	assert.Equal(t, "a1", m.Cursor().Name)
	// Now a1, a2 also visible: root, a, a1, a2, b.
	assert.Equal(t, 5, m.TotalCount())

	// Left/h from a1 ascends back to parent 'a'.
	m, _ = m.Update(keyMsg("h"))
	assert.Equal(t, "a", m.Cursor().Name)

	// Left/h again collapses 'a'.
	m, _ = m.Update(keyMsg("h"))
	assert.Equal(t, 3, m.TotalCount())
}

func TestModel_HomeEnd(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	assert.Equal(t, "b", m.Cursor().Name)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
	assert.Equal(t, "root", m.Cursor().Name)
	assert.Equal(t, 0, m.ScrollOffset())
}

func TestModel_EnterOnRunnableEmitsSelected(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	// Navigate to 'b' (runnable).
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	require.Equal(t, "b", m.Cursor().Name)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	selected, ok := msg.(tree.CommandSelectedMsg)
	require.True(t, ok, "expected CommandSelectedMsg, got %T", msg)
	assert.Equal(t, "b", selected.Node.Name)
}

func TestModel_EnterOnGroupExpands(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	m, _ = m.Update(keyMsg("j")) // 'a', a non-runnable group
	require.Equal(t, "a", m.Cursor().Name)

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Should expand and descend rather than emit CommandSelectedMsg.
	require.NotNil(t, cmd)
	msg := cmd()
	_, isSelected := msg.(tree.CommandSelectedMsg)
	assert.False(t, isSelected)
	assert.Equal(t, "a1", m.Cursor().Name)
}

func TestModel_MoveEmitsHighlight(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	_, cmd := m.Update(keyMsg("j"))
	require.NotNil(t, cmd)
	msg := cmd()
	hl, ok := msg.(tree.CommandHighlightedMsg)
	require.True(t, ok)
	assert.Equal(t, "a", hl.Node.Name)
}

func TestModel_SetFocused(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	m = m.SetFocused(true)
	// View should render without panic when focused.
	m = m.SetSize(40, 10)
	assert.NotEmpty(t, m.View())
}

func TestModel_SetSizeAndView(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	m = m.SetSize(30, 5)
	assert.Equal(t, 5, m.Height())
	assert.NotEmpty(t, m.View())

	// Zero height renders empty.
	m = m.SetSize(30, 0)
	assert.Equal(t, "", m.View())
}

func TestModel_SelectNodeExpandsAncestors(t *testing.T) {
	root := navRoot()
	// Find a2 deep in the tree.
	var a2 *tree.CommandNode
	for _, ch := range root.Children {
		if ch.Name == "a" {
			for _, gc := range ch.Children {
				if gc.Name == "a2" {
					a2 = gc
				}
			}
		}
	}
	require.NotNil(t, a2)

	m := tree.NewModel(root, theme.Default())
	m = m.SetSize(40, 10)
	m = m.SelectNode(a2)
	assert.Equal(t, "a2", m.Cursor().Name)
	// Ancestor 'a' expanded → all 5 nodes visible.
	assert.Equal(t, 5, m.TotalCount())
}

func TestModel_ScrollWithSmallViewport(t *testing.T) {
	m := tree.NewModel(navRoot(), theme.Default())
	m = m.SetSize(40, 2) // only 2 rows visible

	// Move to end; scroll should follow the cursor.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	assert.Equal(t, "b", m.Cursor().Name)
	assert.Greater(t, m.ScrollOffset(), 0)
}
