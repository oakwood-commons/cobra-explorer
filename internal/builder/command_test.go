package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/builder"
	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// leafNode builds a runnable node with the given flags for testing.
func leafNode(flags ...tree.FlagInfo) *tree.CommandNode {
	return &tree.CommandNode{
		Name:     "deploy",
		FullPath: []string{"mycli", "run", "deploy"},
		Runnable: true,
		Flags:    flags,
	}
}

func TestNewBuiltCommand(t *testing.T) {
	node := leafNode()
	bc := builder.NewBuiltCommand(node)

	require.NotNil(t, bc)
	assert.Equal(t, node, bc.Node)
	assert.NotNil(t, bc.FlagValues)
	assert.Empty(t, bc.FlagValues)
	assert.Empty(t, bc.Args)
}

func TestBuiltCommand_SetFlag(t *testing.T) {
	bc := builder.NewBuiltCommand(leafNode())

	bc.SetFlag("name", "prod")
	assert.Equal(t, "prod", bc.FlagValues["name"])

	// Overwrite.
	bc.SetFlag("name", "staging")
	assert.Equal(t, "staging", bc.FlagValues["name"])

	// Empty string clears the flag.
	bc.SetFlag("name", "")
	_, ok := bc.FlagValues["name"]
	assert.False(t, ok)
}

func TestBuiltCommand_ToArgs_Path(t *testing.T) {
	bc := builder.NewBuiltCommand(leafNode())
	// Binary name (index 0) is stripped, sub-path kept.
	assert.Equal(t, []string{"run", "deploy"}, bc.ToArgs())
}

func TestBuiltCommand_ToArgs_SingleElementPath(t *testing.T) {
	node := &tree.CommandNode{Name: "mycli", FullPath: []string{"mycli"}, Runnable: true}
	bc := builder.NewBuiltCommand(node)
	assert.Empty(t, bc.ToArgs())
}

func TestBuiltCommand_ToArgs_FlagTypes(t *testing.T) {
	node := leafNode(
		tree.FlagInfo{Name: "name", Type: "string"},
		tree.FlagInfo{Name: "verbose", Type: "bool"},
		tree.FlagInfo{Name: "count", Type: "int"},
		tree.FlagInfo{Name: "tags", Type: "stringSlice"},
	)
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("name", "prod")
	bc.SetFlag("verbose", "true")
	bc.SetFlag("count", "3")
	bc.SetFlag("tags", "a, b ,c")

	args := bc.ToArgs()
	assert.Equal(t, []string{
		"run", "deploy",
		"--name", "prod",
		"--verbose",
		"--count", "3",
		"--tags", "a",
		"--tags", "b",
		"--tags", "c",
	}, args)
}

func TestBuiltCommand_ToArgs_BoolFalseOmitted(t *testing.T) {
	node := leafNode(tree.FlagInfo{Name: "verbose", Type: "bool"})
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("verbose", "false")
	assert.Equal(t, []string{"run", "deploy"}, bc.ToArgs())
}

func TestBuiltCommand_ToArgs_PrefersShorthand(t *testing.T) {
	node := leafNode(
		tree.FlagInfo{Name: "name", Shorthand: "n", Type: "string"},
		tree.FlagInfo{Name: "verbose", Shorthand: "v", Type: "bool"},
		tree.FlagInfo{Name: "tags", Shorthand: "t", Type: "stringSlice"},
		tree.FlagInfo{Name: "region", Type: "string"}, // no shorthand
	)
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("name", "prod")
	bc.SetFlag("verbose", "true")
	bc.SetFlag("tags", "a,b")
	bc.SetFlag("region", "us-east")

	assert.Equal(t, []string{
		"run", "deploy",
		"-n", "prod",
		"-v",
		"-t", "a",
		"-t", "b",
		"--region", "us-east",
	}, bc.ToArgs())

	assert.Equal(t, "mycli run deploy -n prod -v -t a -t b --region us-east", bc.String())
}

func TestBuiltCommand_ToArgs_EmptyFlagSkipped(t *testing.T) {
	node := leafNode(tree.FlagInfo{Name: "name", Type: "string"})
	bc := builder.NewBuiltCommand(node)
	// Never set — should be skipped.
	assert.Equal(t, []string{"run", "deploy"}, bc.ToArgs())
}

func TestBuiltCommand_ToArgs_SliceSkipsEmptyElements(t *testing.T) {
	node := leafNode(tree.FlagInfo{Name: "tags", Type: "stringArray"})
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("tags", "a,,  ,b")
	assert.Equal(t, []string{"run", "deploy", "--tags", "a", "--tags", "b"}, bc.ToArgs())
}

func TestBuiltCommand_ToArgs_PositionalArgs(t *testing.T) {
	node := leafNode(tree.FlagInfo{Name: "name", Type: "string"})
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("name", "prod")
	bc.Args = []string{"pos1", "pos2"}
	assert.Equal(t, []string{"run", "deploy", "--name", "prod", "pos1", "pos2"}, bc.ToArgs())
}

func TestBuiltCommand_String(t *testing.T) {
	node := leafNode(tree.FlagInfo{Name: "name", Type: "string"})
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("name", "prod")
	assert.Equal(t, "mycli run deploy --name prod", bc.String())
}

func TestBuiltCommand_String_QuotesValuesWithSpaces(t *testing.T) {
	node := leafNode(tree.FlagInfo{Name: "msg", Type: "string"})
	bc := builder.NewBuiltCommand(node)
	bc.SetFlag("msg", "hello world")
	assert.Equal(t, `mycli run deploy --msg "hello world"`, bc.String())
}

func TestBuiltCommand_UnsetRequiredFlags(t *testing.T) {
	node := leafNode(
		tree.FlagInfo{Name: "name", Type: "string", Required: true},
		tree.FlagInfo{Name: "region", Type: "string", Required: true},
		tree.FlagInfo{Name: "verbose", Type: "bool"},
	)
	bc := builder.NewBuiltCommand(node)

	assert.ElementsMatch(t, []string{"name", "region"}, bc.UnsetRequiredFlags())
	assert.False(t, bc.IsValid())

	bc.SetFlag("name", "prod")
	assert.Equal(t, []string{"region"}, bc.UnsetRequiredFlags())
	assert.False(t, bc.IsValid())

	bc.SetFlag("region", "us")
	assert.Empty(t, bc.UnsetRequiredFlags())
	assert.True(t, bc.IsValid())
}

func TestBuiltCommand_IsValid_NoRequiredFlags(t *testing.T) {
	bc := builder.NewBuiltCommand(leafNode(tree.FlagInfo{Name: "opt", Type: "string"}))
	assert.True(t, bc.IsValid())
}
