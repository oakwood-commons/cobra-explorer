package flaginput

import (
	"fmt"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// FlagInput is the interface for type-aware flag value editing.
type FlagInput interface {
	// Flag returns the flag metadata.
	Flag() tree.FlagInfo

	// Value returns the current string value (empty = not set).
	Value() string

	// SetValue sets the value programmatically.
	SetValue(v string) FlagInput

	// Focus activates inline editing.
	Focus() FlagInput

	// Blur deactivates editing.
	Blur() FlagInput

	// IsFocused returns whether this input is actively being edited.
	IsFocused() bool

	// HandleKey processes a keypress while editing.
	HandleKey(msg tea.KeyPressMsg) FlagInput

	// Render produces the display line for this flag.
	Render(isCursor, isEditing bool, maxWidth int) string

	// Validate checks if the current value is valid for the flag type.
	Validate() error
}

// Validator is a function that validates a flag value string.
type Validator func(value string) error

// New creates the appropriate FlagInput for the given flag type.
func New(f tree.FlagInfo) FlagInput {
	switch f.Type {
	case "bool":
		return NewToggle(f)
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return NewTextInput(f, ValidateInt)
	case "float32", "float64":
		return NewTextInput(f, ValidateFloat)
	case "duration":
		return NewTextInput(f, ValidateDuration)
	case "count":
		return NewStepper(f)
	case "stringSlice", "stringArray":
		return NewSliceInput(f)
	default:
		if len(f.ValidValues) > 0 {
			return NewChoice(f)
		}
		return NewTextInput(f, nil)
	}
}

// ValidateInt checks if a string is a valid integer.
func ValidateInt(v string) error {
	if v == "" {
		return nil
	}
	_, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fmt.Errorf("expected integer, got %q", v)
	}
	return nil
}

// ValidateFloat checks if a string is a valid float.
func ValidateFloat(v string) error {
	if v == "" {
		return nil
	}
	_, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("expected number, got %q", v)
	}
	return nil
}

// ValidateDuration checks if a string is a valid Go duration.
func ValidateDuration(v string) error {
	if v == "" {
		return nil
	}
	_, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("expected duration (e.g., 30s, 5m), got %q", v)
	}
	return nil
}
