package flaginput

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// Choice handles flags with a fixed set of valid values.
type Choice struct {
	flag     tree.FlagInfo
	options  []string
	selected int // index into options; -1 = none selected
	focused  bool
}

// NewChoice creates a choice/enum flag input.
func NewChoice(f tree.FlagInfo) FlagInput {
	return &Choice{
		flag:     f,
		options:  f.ValidValues,
		selected: -1,
	}
}

func (c *Choice) Flag() tree.FlagInfo { return c.flag }
func (c *Choice) IsFocused() bool     { return c.focused }

func (c *Choice) Value() string {
	if c.selected < 0 || c.selected >= len(c.options) {
		return ""
	}
	return c.options[c.selected]
}

func (c *Choice) SetValue(v string) FlagInput {
	for i, opt := range c.options {
		if opt == v {
			c.selected = i
			return c
		}
	}
	c.selected = -1
	return c
}

func (c *Choice) Focus() FlagInput {
	c.focused = true
	return c
}

func (c *Choice) Blur() FlagInput {
	c.focused = false
	return c
}

func (c *Choice) HandleKey(msg tea.KeyMsg) FlagInput {
	switch msg.String() {
	case "up", "k":
		if c.selected > 0 {
			c.selected--
		}
	case "down", "j":
		if c.selected < len(c.options)-1 {
			c.selected++
		} else if c.selected < 0 && len(c.options) > 0 {
			c.selected = 0
		}
	}
	return c
}

func (c *Choice) Validate() error { return nil }

func (c *Choice) Render(isCursor, isEditing bool, maxWidth int) string {
	f := c.flag

	var label string
	if f.Shorthand != "" {
		label = fmt.Sprintf("-%s, --%-16s", f.Shorthand, f.Name)
	} else {
		label = fmt.Sprintf("    --%-16s", f.Name)
	}

	val := c.Value()
	if val == "" && f.DefValue != "" {
		val = f.DefValue
	}

	var valueDisplay string
	switch {
	case isEditing:
		valueDisplay = "[" + val + " v]"
	case val != "":
		valueDisplay = "[" + val + "]"
	default:
		valueDisplay = "[        v]"
	}

	line := label + " " + valueDisplay

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
