package tree_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

func TestCommandNode_IsLeaf(t *testing.T) {
	leaf := &tree.CommandNode{Name: "leaf"}
	assert.True(t, leaf.IsLeaf())

	parent := &tree.CommandNode{
		Name:     "parent",
		Children: []*tree.CommandNode{leaf},
	}
	assert.False(t, parent.IsLeaf())
}

func TestCommandNode_AllFlags_Ordering(t *testing.T) {
	node := &tree.CommandNode{
		Name: "deploy",
		Flags: []tree.FlagInfo{
			{Name: "optional", Type: "string"},
			{Name: "required", Type: "string", Required: true},
			{Name: "hidden", Type: "string", Hidden: true},
		},
		InheritedFlags: []tree.FlagInfo{
			{Name: "inherited", Type: "string", Inherited: true},
			{Name: "hiddenInherited", Type: "string", Hidden: true},
		},
	}

	all := node.AllFlags()
	names := make([]string, len(all))
	for i, f := range all {
		names[i] = f.Name
	}

	// Order: required, then local (non-required), then inherited.
	// Hidden flags are excluded.
	assert.Equal(t, []string{"required", "optional", "inherited"}, names)
}

func TestCommandNode_AllFlags_Empty(t *testing.T) {
	node := &tree.CommandNode{Name: "x"}
	assert.Empty(t, node.AllFlags())
}

func TestBuildTree_ShowHidden(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	visible := &cobra.Command{Use: "visible"}
	hidden := &cobra.Command{Use: "hidden", Hidden: true}
	root.AddCommand(visible, hidden)

	// Default: hidden excluded.
	def := tree.BuildTree(root, tree.BuildOptions{})
	assert.Len(t, def.Children, 1)

	// ShowHidden: hidden included.
	shown := tree.BuildTree(root, tree.BuildOptions{ShowHidden: true})
	assert.Len(t, shown.Children, 2)
}

func TestBuildTree_ShowDeprecated(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	normal := &cobra.Command{Use: "normal"}
	deprecated := &cobra.Command{Use: "old", Deprecated: "use normal"}
	root.AddCommand(normal, deprecated)

	def := tree.BuildTree(root, tree.BuildOptions{})
	assert.Len(t, def.Children, 1)

	shown := tree.BuildTree(root, tree.BuildOptions{ShowDeprecated: true})
	assert.Len(t, shown.Children, 2)
}

func TestBuildTree_ExcludeNames(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	keep := &cobra.Command{Use: "keep"}
	drop := &cobra.Command{Use: "explore"}
	root.AddCommand(keep, drop)

	node := tree.BuildTree(root, tree.BuildOptions{ExcludeNames: []string{"explore"}})
	require.Len(t, node.Children, 1)
	assert.Equal(t, "keep", node.Children[0].Name)
}

func TestBuildTree_BinaryNameOverride(t *testing.T) {
	root := &cobra.Command{Use: "actualname"}
	node := tree.BuildTree(root, tree.BuildOptions{BinaryName: "mycli"})
	assert.Equal(t, []string{"mycli"}, node.FullPath)
}

func TestBuildTree_InheritedFlags(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("config", "", "config file")
	sub := &cobra.Command{Use: "sub", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	root.AddCommand(sub)

	node := tree.BuildTree(root, tree.BuildOptions{})
	require.Len(t, node.Children, 1)
	subNode := node.Children[0]

	var found bool
	for _, f := range subNode.InheritedFlags {
		if f.Name == "config" {
			found = true
			assert.True(t, f.Inherited)
		}
	}
	assert.True(t, found, "sub should inherit the persistent 'config' flag")
}

func TestBuildTree_TotalLeaves(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	group := &cobra.Command{Use: "group"}
	leaf1 := &cobra.Command{Use: "l1", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	leaf2 := &cobra.Command{Use: "l2", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	group.AddCommand(leaf1, leaf2)
	root.AddCommand(group)

	node := tree.BuildTree(root, tree.BuildOptions{})
	assert.Equal(t, 2, node.TotalLeaves)
}
