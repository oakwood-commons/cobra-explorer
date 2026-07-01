package flaginput

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// Toggle handles boolean flags.
type Toggle struct {
	flag  tree.FlagInfo
	value bool
}

// NewToggle creates a boolean flag toggle.
func NewToggle(f tree.FlagInfo) FlagInput {
	return &Toggle{
		flag:  f,
		value: f.DefValue == "true",
	}
}

func (t *Toggle) Flag() tree.FlagInfo { return t.flag }
func (t *Toggle) IsFocused() bool     { return false }

func (t *Toggle) Value() string {
	if t.value {
		return "true"
	}
	return ""
}

func (t *Toggle) SetValue(v string) FlagInput {
	t.value = (v == "true")
	return t
}

func (t *Toggle) Focus() FlagInput { return t }
func (t *Toggle) Blur() FlagInput  { return t }

func (t *Toggle) HandleKey(msg tea.KeyMsg) FlagInput {
	if msg.String() == " " || msg.String() == "enter" {
		t.value = !t.value
	}
	return t
}

func (t *Toggle) Validate() error { return nil }

func (t *Toggle) Render(isCursor, _ bool, maxWidth int) string {
	f := t.flag

	var label string
	if f.Shorthand != "" {
		label = fmt.Sprintf("-%s, --%-16s", f.Shorthand, f.Name)
	} else {
		label = fmt.Sprintf("    --%-16s", f.Name)
	}

	check := "[ ]"
	if t.value {
		check = "[x]"
	}

	line := label + " " + check

	if isCursor {
		line = "> " + line
	} else {
		line = "  " + line
	}

	if maxWidth > 0 && len(line) > maxWidth {
		line = line[:maxWidth-3] + "..."
	}

	return line
}
