package theme_test

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/theme"
)

func TestRegistry_HasBuiltinPresets(t *testing.T) {
	for _, name := range []string{"dark", "light", "dracula", "nord"} {
		th, ok := theme.Get(name)
		require.True(t, ok, "preset %q should be registered", name)
		assert.Equal(t, name, th.Name)
	}
}

func TestGet_UnknownReturnsFalse(t *testing.T) {
	_, ok := theme.Get("does-not-exist")
	assert.False(t, ok)
}

func TestNames_SortedAndContainsBuiltins(t *testing.T) {
	names := theme.Names()

	require.NotEmpty(t, names)
	// Sorted ascending.
	for i := 1; i < len(names); i++ {
		assert.LessOrEqual(t, names[i-1], names[i], "Names should be sorted")
	}
	assert.Contains(t, names, "dark")
	assert.Contains(t, names, "light")
}

func TestRegister_MakesThemeSelectable(t *testing.T) {
	custom := theme.New(theme.Palette{
		Name:       "test-custom",
		Background: lipgloss.Color("55"),
		Foreground: lipgloss.Color("15"),
		Accent:     lipgloss.Color("5"),
	})
	theme.Register(custom)

	got, ok := theme.Get("test-custom")
	require.True(t, ok)
	assert.Equal(t, "test-custom", got.Name)
	assert.Contains(t, theme.Names(), "test-custom")
}
