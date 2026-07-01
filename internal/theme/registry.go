package theme

import (
	"sort"
	"sync"
)

// registry holds the theme presets that can be selected by name. It is seeded
// with the built-in presets and can be extended at startup via Register, which
// lets contributors add their own themes without touching the model code.
var (
	registryMu sync.RWMutex
	registry   = builtinRegistry()
)

// builtinRegistry returns the presets that ship with the library.
func builtinRegistry() map[string]Theme {
	m := make(map[string]Theme)
	for _, p := range []Palette{DarkPalette, NightPalette, LightPalette, DraculaPalette, NordPalette, TerminalPalette, TerminalLightPalette} {
		t := New(p)
		m[t.Name] = t
	}
	return m
}

// Register adds (or overrides) a named theme preset so it can be selected by
// name. Contributors call this to make a custom theme available. It is safe
// for concurrent use.
func Register(t Theme) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[t.Name] = t
}

// Get returns the registered theme with the given name and whether it exists.
func Get(name string) (Theme, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	t, ok := registry[name]
	return t, ok
}

// Names returns the sorted names of all registered theme presets.
func Names() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
