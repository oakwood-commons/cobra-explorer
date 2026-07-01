package clipboard

import "fmt"

// Clipboard is the interface for system clipboard access.
type Clipboard interface {
	// Write copies text to the system clipboard.
	Write(text string) error

	// Available reports whether clipboard access is supported.
	Available() bool
}

// New returns the platform-appropriate clipboard implementation.
func New() Clipboard {
	return newPlatformClipboard()
}

// ErrUnavailable is returned when no clipboard mechanism is found.
var ErrUnavailable = fmt.Errorf("clipboard: no clipboard mechanism available")
