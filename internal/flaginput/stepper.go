package flaginput

import (
	"fmt"
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// Stepper handles count/int flags with increment/decrement.
type Stepper struct {
	flag  tree.FlagInfo
	value int
}

// NewStepper creates a stepper flag input.
func NewStepper(f tree.FlagInfo) FlagInput {
	v, _ := strconv.Atoi(f.DefValue)
	return &Stepper{
		flag:  f,
		value: v,
	}
}

func (s *Stepper) Flag() tree.FlagInfo { return s.flag }
func (s *Stepper) IsFocused() bool     { return false }

func (s *Stepper) Value() string {
	if s.value == 0 {
		return ""
	}
	return strconv.Itoa(s.value)
}

func (s *Stepper) SetValue(v string) FlagInput {
	n, _ := strconv.Atoi(v)
	s.value = n
	return s
}

func (s *Stepper) Focus() FlagInput { return s }
func (s *Stepper) Blur() FlagInput  { return s }

func (s *Stepper) HandleKey(msg tea.KeyPressMsg) FlagInput {
	switch msg.String() {
	case "up", "k", "+":
		s.value++
	case "down", "j", "-":
		if s.value > 0 {
			s.value--
		}
	}
	return s
}

func (s *Stepper) Validate() error { return nil }

func (s *Stepper) Render(isCursor, _ bool, maxWidth int) string {
	f := s.flag

	var label string
	if f.Shorthand != "" {
		label = fmt.Sprintf("-%s, --%-16s", f.Shorthand, f.Name)
	} else {
		label = fmt.Sprintf("    --%-16s", f.Name)
	}

	valueDisplay := fmt.Sprintf("[%d +/-]", s.value)

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
