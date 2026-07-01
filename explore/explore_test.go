package explore_test

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/explore"
)

func TestNewCommand(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	cmd := explore.NewCommand(root)

	require.NotNil(t, cmd)
	assert.Equal(t, "explore", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestNewCommand_WithOptions(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	// Options should be accepted without panicking; command is constructed lazily.
	cmd := explore.NewCommand(root,
		explore.WithBinaryName("custom"),
		explore.WithLightTheme(),
		explore.WithShowHidden(true),
		explore.WithExecution(true),
	)
	require.NotNil(t, cmd)
	assert.Equal(t, "explore", cmd.Use)
}

func TestOptions_AreFunctional(t *testing.T) {
	// Each option returns a non-nil Option func.
	opts := []explore.Option{
		explore.WithBinaryName("x"),
		explore.WithThemeName("dracula"),
		explore.WithLightTheme(),
		explore.WithShowHidden(true),
		explore.WithExecution(true),
	}
	for i, o := range opts {
		assert.NotNil(t, o, "option %d should not be nil", i)
	}
}

func TestNewCommand_EmbeddableAsSubcommand(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	root.AddCommand(explore.NewCommand(root))

	found, _, err := root.Find([]string{"explore"})
	require.NoError(t, err)
	assert.Equal(t, "explore", found.Name())
}

func TestNewCommand_HasThemeFlag(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	cmd := explore.NewCommand(root)

	flag := cmd.Flags().Lookup("theme")
	require.NotNil(t, flag, "explore command should expose a --theme flag")
	assert.Empty(t, flag.DefValue, "theme flag should default to empty so it does not override developer config")
}

func TestNewCommand_UnknownThemeReturnsError(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	cmd := explore.NewCommand(root)
	require.NoError(t, cmd.Flags().Set("theme", "does-not-exist"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown theme")
	assert.Contains(t, err.Error(), "does-not-exist")
}

func TestNewCommand_HasListThemesFlag(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	cmd := explore.NewCommand(root)

	flag := cmd.Flags().Lookup("list-themes")
	require.NotNil(t, flag, "explore command should expose a --list-themes flag")
}

func TestNewCommand_ListThemesFlagListsThemes(t *testing.T) {
	root := &cobra.Command{Use: "mycli"}
	root.AddCommand(explore.NewCommand(root))
	root.SetArgs([]string{"explore", "--list-themes"})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	require.NoError(t, root.Execute())

	got := out.String()
	assert.Contains(t, got, "Available themes:")
	assert.Contains(t, got, "dark")
	assert.Contains(t, got, "dracula")
}
