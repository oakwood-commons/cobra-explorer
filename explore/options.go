package explore

// Option configures the explorer behavior.
type Option func(*config)

type config struct {
	binaryName string
	themeName  string
	showHidden bool
}

func defaultConfig() config {
	return config{}
}

// WithBinaryName overrides the binary name displayed in the TUI.
// By default, the root command's Name() is used.
func WithBinaryName(name string) Option {
	return func(c *config) {
		c.binaryName = name
	}
}

// WithThemeName selects a registered theme preset by name (for example
// "dark", "light", "dracula", or "nord"). Unknown names fall back to the
// default dark theme. Additional presets can be registered by contributors.
func WithThemeName(name string) Option {
	return func(c *config) {
		c.themeName = name
	}
}

// WithLightTheme uses the built-in light theme. It is shorthand for
// WithThemeName("light").
func WithLightTheme() Option {
	return func(c *config) {
		c.themeName = "light"
	}
}

// WithShowHidden makes hidden commands visible in the tree.
func WithShowHidden(show bool) Option {
	return func(c *config) {
		c.showHidden = show
	}
}
