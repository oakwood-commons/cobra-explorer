package theme_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"

	"github.com/oakwood-commons/cobra-explorer/internal/theme"
)

func TestDefault_RendersStyles(t *testing.T) {
	th := theme.Default()

	// Every style should render its input without dropping the text.
	assert.Contains(t, th.Title.Render("hello"), "hello")
	assert.Contains(t, th.Heading.Render("head"), "head")
	assert.Contains(t, th.TreeCursor.Render("cur"), "cur")
	assert.Contains(t, th.CommandPreview.Render("cmd"), "cmd")
	assert.Contains(t, th.Success.Render("ok"), "ok")
	assert.Contains(t, th.Warning.Render("warn"), "warn")
	assert.Contains(t, th.Error.Render("err"), "err")
}

func TestLight_RendersStyles(t *testing.T) {
	th := theme.Light()

	assert.Contains(t, th.Title.Render("hello"), "hello")
	assert.Contains(t, th.TreeRunnable.Render("run"), "run")
	assert.Contains(t, th.SearchPrompt.Render("search"), "search")
}

func TestDefaultAndLight_DifferInColors(t *testing.T) {
	dark := theme.Default()
	light := theme.Light()

	// Tree normal foreground differs between themes.
	assert.NotEqual(t,
		dark.TreeNormal.GetForeground(),
		light.TreeNormal.GetForeground(),
	)
}

func TestThemes_HaveBorders(t *testing.T) {
	for name, th := range map[string]theme.Theme{"dark": theme.Default(), "light": theme.Light()} {
		t.Run(name, func(t *testing.T) {
			assert.True(t, th.PanelBorder.GetBorderTop(), "PanelBorder should have a border")
			assert.True(t, th.PanelBorderFocused.GetBorderTop(), "PanelBorderFocused should have a border")
		})
	}
}

func TestNew_AppliesBackgroundToStyles(t *testing.T) {
	p := theme.Palette{
		Name:       "test",
		Background: lipgloss.Color("21"),
		Foreground: lipgloss.Color("15"),
		Accent:     lipgloss.Color("5"),
	}
	th := theme.New(p)

	// Base and body-text styles must carry the palette background so the
	// surface color is consistent across panels.
	for label, style := range map[string]lipgloss.Style{
		"Base":       th.Base,
		"Body":       th.Body,
		"TreeNormal": th.TreeNormal,
		"Dim":        th.Dim,
		"Heading":    th.Heading,
	} {
		assert.Equal(t, lipgloss.Color("21"), style.GetBackground(),
			"%s should use the palette background", label)
	}
	assert.Equal(t, "test", th.Name)
}

func TestLight_UsesLightBackground(t *testing.T) {
	light := theme.Light()
	dark := theme.Default()

	// The light theme must define an explicit background distinct from the dark
	// theme so it actually renders as a light surface.
	assert.Equal(t, theme.LightPalette.Background, light.Base.GetBackground())
	assert.NotEqual(t, dark.Base.GetBackground(), light.Base.GetBackground())
}

func TestBackgroundColor_MatchesBase(t *testing.T) {
	th := theme.Default()
	assert.Equal(t, th.Base.GetBackground(), th.BackgroundColor())
}

func TestTerminalThemes_HaveTransparentSurface(t *testing.T) {
	// Force a color profile so styles emit ANSI sequences we can inspect.
	lipgloss.SetColorProfile(termenv.ANSI256)

	for _, name := range []string{"terminal", "terminal-light"} {
		t.Run(name, func(t *testing.T) {
			th, ok := theme.Get(name)
			assert.True(t, ok, "%s theme should be registered", name)

			// The surface styles must not paint a background so the terminal's
			// own background shows through. A background SGR uses the "48;"
			// prefix; a transparent style must omit it.
			for label, style := range map[string]lipgloss.Style{
				"Base":       th.Base,
				"Body":       th.Body,
				"TreeNormal": th.TreeNormal,
				"Dim":        th.Dim,
			} {
				out := style.Render("x")
				assert.NotContains(t, out, "48;",
					"%s.%s should render no background", name, label)
			}

			// The cursor and command preview keep explicit fills so they stay
			// legible over any terminal background.
			assert.Contains(t, th.TreeCursor.Render("x"), "48;",
				"%s TreeCursor should keep its accent background", name)
			assert.Contains(t, th.CommandPreview.Render("x"), "48;",
				"%s CommandPreview should keep its preview background", name)
		})
	}
}

func TestTerminalThemes_UseTerminalDefaultBackground(t *testing.T) {
	for _, name := range []string{"terminal", "terminal-light"} {
		th, ok := theme.Get(name)
		assert.True(t, ok)
		// BackgroundColor is what fills layout gaps; for a transparent theme it
		// must not resolve to an opaque color like the dark/light surfaces.
		assert.NotEqual(t, theme.DarkPalette.Background, th.BackgroundColor(),
			"%s should not use the opaque dark background", name)
		assert.NotEqual(t, theme.LightPalette.Background, th.BackgroundColor(),
			"%s should not use the opaque light background", name)
	}
}
