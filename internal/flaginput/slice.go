package flaginput

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// SliceInput handles stringSlice and stringArray flags.
type SliceInput struct {
	flag    tree.FlagInfo
	values  []string
	buffer  string
	cursor  int
	focused bool
}

// NewSliceInput creates a multi-value flag input.
func NewSliceInput(f tree.FlagInfo) FlagInput {
	return &SliceInput{
		flag: f,
	}
}

func (s *SliceInput) Flag() tree.FlagInfo { return s.flag }
func (s *SliceInput) IsFocused() bool     { return s.focused }

func (s *SliceInput) Value() string {
	if len(s.values) == 0 {
		return ""
	}
	return strings.Join(s.values, ",")
}

func (s *SliceInput) SetValue(v string) FlagInput {
	if v == "" {
		s.values = nil
	} else {
		s.values = strings.Split(v, ",")
	}
	return s
}

func (s *SliceInput) Focus() FlagInput {
	s.focused = true
	s.buffer = ""
	s.cursor = 0
	return s
}

func (s *SliceInput) Blur() FlagInput {
	if s.buffer != "" {
		s.values = append(s.values, strings.TrimSpace(s.buffer))
		s.buffer = ""
	}
	s.focused = false
	return s
}

func (s *SliceInput) HandleKey(msg tea.KeyMsg) FlagInput {
	switch msg.String() {
	case "backspace":
		if s.cursor > 0 && len(s.buffer) > 0 {
			s.buffer = s.buffer[:s.cursor-1] + s.buffer[s.cursor:]
			s.cursor--
		} else if s.buffer == "" && len(s.values) > 0 {
			// Remove last value
			s.values = s.values[:len(s.values)-1]
		}
	case ",":
		// Commit current buffer as a value
		v := strings.TrimSpace(s.buffer)
		if v != "" {
			s.values = append(s.values, v)
		}
		s.buffer = ""
		s.cursor = 0
	case "left":
		if s.cursor > 0 {
			s.cursor--
		}
	case "right":
		if s.cursor < len(s.buffer) {
			s.cursor++
		}
	default:
		if len(msg.String()) == 1 {
			ch := msg.String()
			s.buffer = s.buffer[:s.cursor] + ch + s.buffer[s.cursor:]
			s.cursor++
		}
	}
	return s
}

func (s *SliceInput) Validate() error { return nil }

func (s *SliceInput) Render(isCursor, isEditing bool, maxWidth int) string {
	f := s.flag

	var label string
	if f.Shorthand != "" {
		label = fmt.Sprintf("-%s, --%-16s", f.Shorthand, f.Name)
	} else {
		label = fmt.Sprintf("    --%-16s", f.Name)
	}

	var valueDisplay string
	switch {
	case isEditing:
		existing := strings.Join(s.values, ", ")
		if existing != "" {
			existing += ", "
		}
		valueDisplay = "[" + existing + s.buffer + "|]"
	case len(s.values) > 0:
		valueDisplay = "[" + strings.Join(s.values, ", ") + "]"
	default:
		valueDisplay = "[        ]"
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
