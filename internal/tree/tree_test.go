package tree_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

func buildTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "testcli", Short: "Test CLI"}
	sub := &cobra.Command{Use: "sub", Short: "A subcommand"}
	leaf := &cobra.Command{
		Use:   "leaf",
		Short: "A leaf command",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
	leaf.Flags().StringP("name", "n", "", "Name flag")
	leaf.Flags().BoolP("verbose", "v", false, "Verbose flag")
	_ = leaf.MarkFlagRequired("name")
	sub.AddCommand(leaf)
	root.AddCommand(sub)

	hidden := &cobra.Command{Use: "hidden", Short: "Hidden cmd", Hidden: true}
	root.AddCommand(hidden)

	return root
}

func TestBuildTree(t *testing.T) {
	root := buildTestRoot()
	node := tree.BuildTree(root, tree.BuildOptions{})

	assert.Equal(t, "testcli", node.Name)
	assert.Equal(t, []string{"testcli"}, node.FullPath)
	assert.False(t, node.Runnable)
	assert.Equal(t, 0, node.Depth)
}

func TestBuildTree_Children(t *testing.T) {
	root := buildTestRoot()
	node := tree.BuildTree(root, tree.BuildOptions{})

	// sub (hidden is filtered out by default)
	require.GreaterOrEqual(t, len(node.Children), 1)

	var sub *tree.CommandNode
	for _, ch := range node.Children {
		if ch.Name == "sub" {
			sub = ch
			break
		}
	}
	require.NotNil(t, sub)
	assert.Equal(t, []string{"testcli", "sub"}, sub.FullPath)
	assert.Equal(t, 1, sub.Depth)
	assert.Equal(t, node, sub.Parent)
}

func TestBuildTree_LeafFlags(t *testing.T) {
	root := buildTestRoot()
	node := tree.BuildTree(root, tree.BuildOptions{})

	var leaf *tree.CommandNode
	for _, ch := range node.Children {
		if ch.Name == "sub" {
			for _, grandchild := range ch.Children {
				if grandchild.Name == "leaf" {
					leaf = grandchild
				}
			}
		}
	}
	require.NotNil(t, leaf)
	assert.True(t, leaf.Runnable)
	assert.Equal(t, 2, leaf.Depth)

	// Should have "name" and "verbose" in local flags
	var nameFlag, verboseFlag *tree.FlagInfo
	for i := range leaf.Flags {
		switch leaf.Flags[i].Name {
		case "name":
			nameFlag = &leaf.Flags[i]
		case "verbose":
			verboseFlag = &leaf.Flags[i]
		}
	}
	require.NotNil(t, nameFlag)
	assert.Equal(t, "n", nameFlag.Shorthand)
	assert.Equal(t, "string", nameFlag.Type)
	assert.True(t, nameFlag.Required)

	require.NotNil(t, verboseFlag)
	assert.Equal(t, "bool", verboseFlag.Type)
	assert.False(t, verboseFlag.Required)
}

func TestCommandNode_AllFlags(t *testing.T) {
	root := buildTestRoot()
	node := tree.BuildTree(root, tree.BuildOptions{})

	var leaf *tree.CommandNode
	for _, ch := range node.Children {
		if ch.Name == "sub" {
			for _, grandchild := range ch.Children {
				if grandchild.Name == "leaf" {
					leaf = grandchild
				}
			}
		}
	}
	require.NotNil(t, leaf)

	all := leaf.AllFlags()
	// name is required, should be first
	require.GreaterOrEqual(t, len(all), 2)
	assert.Equal(t, "name", all[0].Name)
}

func TestCommandNode_CommandString(t *testing.T) {
	node := &tree.CommandNode{
		Name:     "deploy",
		FullPath: []string{"mycli", "run", "deploy"},
	}
	assert.Equal(t, "mycli run deploy", node.CommandString())
}
