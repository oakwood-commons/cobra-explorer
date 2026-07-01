package theme

import "github.com/charmbracelet/lipgloss"

// Palette defines the semantic colors of a theme. Contributors add new theme
// presets by constructing a Palette and passing it to New — every lipgloss
// style in the resulting Theme is derived from these colors, so a preset only
// needs to specify its palette. See registry.go for how presets are registered
// and made selectable by name.
type Palette struct {
	// Name is the unique identifier used to select the theme (e.g. "dark").
	Name string

	// Background is the fill color for the whole UI and every panel interior.
	// Leave it empty (the zero value) to make the surface transparent so the
	// host terminal's own background shows through — see TerminalPalette.
	Background lipgloss.Color
	// Foreground is the default text color used across panels.
	Foreground lipgloss.Color
	// Muted is used for secondary/dim text (hints, placeholders).
	Muted lipgloss.Color

	// Border is the color of unfocused panel borders.
	Border lipgloss.Color
	// BorderActive is the color of the focused panel border.
	BorderActive lipgloss.Color

	// Accent highlights titles, the tree cursor, and subheadings.
	Accent lipgloss.Color
	// AccentText is the text color drawn on top of Accent-colored backgrounds.
	AccentText lipgloss.Color
	// Selection is the tree cursor background while the tree is unfocused.
	Selection lipgloss.Color

	// Runnable colors runnable commands in the tree.
	Runnable lipgloss.Color

	// PreviewText and PreviewBackground style the assembled command preview.
	PreviewText       lipgloss.Color
	PreviewBackground lipgloss.Color

	// Scrollbar colors scrollbar thumbs and tracks.
	Scrollbar lipgloss.Color

	// Semantic status colors.
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
}

// Theme defines all visual styles used by the TUI. Build one from a Palette
// with New; the individual styles are derived so they share a consistent
// background and text color.
type Theme struct {
	// Name identifies the theme preset this was built from.
	Name string

	// Base fills the UI background and provides the default text color. It is
	// used to paint gaps and whitespace so the surface color is uniform.
	Base lipgloss.Style
	// Body styles scrollable body text (descriptions, command output).
	Body lipgloss.Style

	// Panel chrome
	PanelBorder        lipgloss.Style
	PanelBorderFocused lipgloss.Style
	PanelTitle         lipgloss.Style
	Title              lipgloss.Style

	// Tree
	TreeNormal          lipgloss.Style
	TreeCursor          lipgloss.Style
	TreeCursorUnfocused lipgloss.Style
	TreeRunnable        lipgloss.Style

	// Text
	Heading    lipgloss.Style
	Subheading lipgloss.Style
	Dim        lipgloss.Style

	// Command builder
	CommandPreview lipgloss.Style
	KeyHint        lipgloss.Style

	// Search
	SearchPrompt lipgloss.Style

	// Scrollbar
	Scrollbar lipgloss.Style

	// Status
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
}

// BackgroundColor returns the palette background color, suitable for
// lipgloss.WithWhitespaceBackground when filling layout gaps.
func (t Theme) BackgroundColor() lipgloss.TerminalColor {
	return t.Base.GetBackground()
}

// New builds a Theme from a Palette. Every style inherits the palette's
// background so the UI renders with a consistent surface color.
func New(p Palette) Theme {
	base := lipgloss.NewStyle().
		Background(p.Background).
		Foreground(p.Foreground)

	// fg returns a style with the given foreground on the palette background.
	fg := func(c lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(c).Background(p.Background)
	}

	// panel returns a bordered box whose border and interior share the bg.
	panel := func(border lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			BorderBackground(p.Background).
			Background(p.Background)
	}

	onAccent := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.AccentText).
		Background(p.Accent).
		Padding(0, 1)

	return Theme{
		Name: p.Name,
		Base: base,
		Body: base,

		PanelBorder:        panel(p.Border),
		PanelBorderFocused: panel(p.BorderActive),
		PanelTitle:         onAccent,
		Title:              onAccent,

		TreeNormal: fg(p.Foreground),
		TreeCursor: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.AccentText).
			Background(p.Accent),
		TreeCursorUnfocused: lipgloss.NewStyle().
			Foreground(p.Foreground).
			Background(p.Selection),
		TreeRunnable: fg(p.Runnable),

		Heading:    fg(p.Foreground).Bold(true),
		Subheading: fg(p.Accent).Bold(true),
		Dim:        fg(p.Muted),

		CommandPreview: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.PreviewText).
			Background(p.PreviewBackground).
			Padding(0, 1),
		KeyHint: fg(p.Muted),

		SearchPrompt: fg(p.Warning).Bold(true),

		Scrollbar: fg(p.Scrollbar),

		Success: fg(p.Success),
		Warning: fg(p.Warning),
		Error:   fg(p.Error).Bold(true),
	}
}

// Default returns the built-in dark theme.
func Default() Theme { return New(DarkPalette) }

// Light returns the built-in light theme.
func Light() Theme { return New(LightPalette) }

// DarkPalette is the default dark theme palette.
var DarkPalette = Palette{
	Name:              "dark",
	Background:        lipgloss.Color("235"),
	Foreground:        lipgloss.Color("252"),
	Muted:             lipgloss.Color("244"),
	Border:            lipgloss.Color("62"),
	BorderActive:      lipgloss.Color("75"),
	Accent:            lipgloss.Color("62"),
	AccentText:        lipgloss.Color("231"),
	Selection:         lipgloss.Color("240"),
	Runnable:          lipgloss.Color("114"),
	PreviewText:       lipgloss.Color("48"),
	PreviewBackground: lipgloss.Color("236"),
	Scrollbar:         lipgloss.Color("62"),
	Success:           lipgloss.Color("48"),
	Warning:           lipgloss.Color("214"),
	Error:             lipgloss.Color("203"),
}

// NightPalette is a darker variant of the dark theme with a near-black
// background for low-light environments and OLED displays.
var NightPalette = Palette{
	Name:              "night",
	Background:        lipgloss.Color("232"),
	Foreground:        lipgloss.Color("252"),
	Muted:             lipgloss.Color("244"),
	Border:            lipgloss.Color("62"),
	BorderActive:      lipgloss.Color("75"),
	Accent:            lipgloss.Color("62"),
	AccentText:        lipgloss.Color("231"),
	Selection:         lipgloss.Color("240"),
	Runnable:          lipgloss.Color("114"),
	PreviewText:       lipgloss.Color("48"),
	PreviewBackground: lipgloss.Color("234"),
	Scrollbar:         lipgloss.Color("62"),
	Success:           lipgloss.Color("48"),
	Warning:           lipgloss.Color("214"),
	Error:             lipgloss.Color("203"),
}

// LightPalette is the built-in light theme palette. It uses a white background
// with dark text so it reads clearly as a light theme.
var LightPalette = Palette{
	Name:              "light",
	Background:        lipgloss.Color("231"),
	Foreground:        lipgloss.Color("236"),
	Muted:             lipgloss.Color("242"),
	Border:            lipgloss.Color("250"),
	BorderActive:      lipgloss.Color("33"),
	Accent:            lipgloss.Color("63"),
	AccentText:        lipgloss.Color("231"),
	Selection:         lipgloss.Color("253"),
	Runnable:          lipgloss.Color("28"),
	PreviewText:       lipgloss.Color("22"),
	PreviewBackground: lipgloss.Color("254"),
	Scrollbar:         lipgloss.Color("63"),
	Success:           lipgloss.Color("28"),
	Warning:           lipgloss.Color("130"),
	Error:             lipgloss.Color("160"),
}

// TerminalPalette is a transparent preset that inherits the host terminal's
// background. It leaves Background empty so no surface color is painted, and is
// tuned with light text for dark terminals. The tree cursor, unfocused cursor,
// and command preview keep explicit fills so they stay legible over any
// background.
var TerminalPalette = Palette{
	Name:              "terminal",
	Background:        lipgloss.Color(""),
	Foreground:        lipgloss.Color("252"),
	Muted:             lipgloss.Color("244"),
	Border:            lipgloss.Color("62"),
	BorderActive:      lipgloss.Color("75"),
	Accent:            lipgloss.Color("62"),
	AccentText:        lipgloss.Color("231"),
	Selection:         lipgloss.Color("240"),
	Runnable:          lipgloss.Color("114"),
	PreviewText:       lipgloss.Color("48"),
	PreviewBackground: lipgloss.Color("236"),
	Scrollbar:         lipgloss.Color("62"),
	Success:           lipgloss.Color("48"),
	Warning:           lipgloss.Color("214"),
	Error:             lipgloss.Color("203"),
}

// TerminalLightPalette is the light-terminal counterpart of TerminalPalette. It
// is also transparent (empty Background) but tuned with dark text and borders
// so it reads clearly on terminals with a light background.
var TerminalLightPalette = Palette{
	Name:              "terminal-light",
	Background:        lipgloss.Color(""),
	Foreground:        lipgloss.Color("236"),
	Muted:             lipgloss.Color("242"),
	Border:            lipgloss.Color("250"),
	BorderActive:      lipgloss.Color("33"),
	Accent:            lipgloss.Color("63"),
	AccentText:        lipgloss.Color("231"),
	Selection:         lipgloss.Color("253"),
	Runnable:          lipgloss.Color("28"),
	PreviewText:       lipgloss.Color("22"),
	PreviewBackground: lipgloss.Color("254"),
	Scrollbar:         lipgloss.Color("63"),
	Success:           lipgloss.Color("28"),
	Warning:           lipgloss.Color("130"),
	Error:             lipgloss.Color("160"),
}

// DraculaPalette is a preset based on the popular Dracula color scheme.
var DraculaPalette = Palette{
	Name:              "dracula",
	Background:        lipgloss.Color("236"),
	Foreground:        lipgloss.Color("253"),
	Muted:             lipgloss.Color("61"),
	Border:            lipgloss.Color("61"),
	BorderActive:      lipgloss.Color("141"),
	Accent:            lipgloss.Color("141"),
	AccentText:        lipgloss.Color("236"),
	Selection:         lipgloss.Color("238"),
	Runnable:          lipgloss.Color("84"),
	PreviewText:       lipgloss.Color("84"),
	PreviewBackground: lipgloss.Color("235"),
	Scrollbar:         lipgloss.Color("61"),
	Success:           lipgloss.Color("84"),
	Warning:           lipgloss.Color("215"),
	Error:             lipgloss.Color("203"),
}

// NordPalette is a preset based on the Nord color scheme.
var NordPalette = Palette{
	Name:              "nord",
	Background:        lipgloss.Color("236"),
	Foreground:        lipgloss.Color("188"),
	Muted:             lipgloss.Color("103"),
	Border:            lipgloss.Color("67"),
	BorderActive:      lipgloss.Color("110"),
	Accent:            lipgloss.Color("110"),
	AccentText:        lipgloss.Color("236"),
	Selection:         lipgloss.Color("239"),
	Runnable:          lipgloss.Color("108"),
	PreviewText:       lipgloss.Color("108"),
	PreviewBackground: lipgloss.Color("235"),
	Scrollbar:         lipgloss.Color("67"),
	Success:           lipgloss.Color("108"),
	Warning:           lipgloss.Color("222"),
	Error:             lipgloss.Color("131"),
}
