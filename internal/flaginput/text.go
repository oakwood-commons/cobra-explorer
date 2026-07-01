package flaginput

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// TextInput handles string, int, float, duration, and other text-based flags.
type TextInput struct {
	flag      tree.FlagInfo
	validator Validator
	value     string
	buffer    string // editing buffer
	cursor    int
	focused   bool
}

// NewTextInput creates a text-based flag input.
func NewTextInput(f tree.FlagInfo, v Validator) FlagInput {
	return &TextInput{
		flag:      f,
		validator: v,
	}
}

func (t *TextInput) Flag() tree.FlagInfo { return t.flag }
func (t *TextInput) Value() string       { return t.value }
func (t *TextInput) IsFocused() bool     { return t.focused }

func (t *TextInput) SetValue(v string) FlagInput {
	t.value = v
	return t
}

func (t *TextInput) Focus() FlagInput {
	t.focused = true
	t.buffer = t.value
	t.cursor = len(t.buffer)
	return t
}

func (t *TextInput) Blur() FlagInput {
	t.value = t.buffer
	t.focused = false
	return t
}

func (t *TextInput) HandleKey(msg tea.KeyMsg) FlagInput {
	switch msg.String() {
	case "backspace":
		if t.cursor > 0 && len(t.buffer) > 0 {
			t.buffer = t.buffer[:t.cursor-1] + t.buffer[t.cursor:]
			t.cursor--
		}
	case "delete":
		if t.cursor < len(t.buffer) {
			t.buffer = t.buffer[:t.cursor] + t.buffer[t.cursor+1:]
		}
	case "left":
		if t.cursor > 0 {
			t.cursor--
		}
	case "right":
		if t.cursor < len(t.buffer) {
			t.cursor++
		}
	case "home", "ctrl+a":
		t.cursor = 0
	case "end", "ctrl+e":
		t.cursor = len(t.buffer)
	case "ctrl+u":
		t.buffer = ""
		t.cursor = 0
	default:
		// Insert printable character
		if len(msg.String()) == 1 {
			ch := msg.String()
			t.buffer = t.buffer[:t.cursor] + ch + t.buffer[t.cursor:]
			t.cursor++
		}
	}
	return t
}

func (t *TextInput) Validate() error {
	if t.validator == nil {
		return nil
	}
	return t.validator(t.value)
}

func (t *TextInput) Render(isCursor, isEditing bool, maxWidth int) string {
	f := t.flag

	var label string
	if f.Shorthand != "" {
		label = fmt.Sprintf("-%s, --%-16s", f.Shorthand, f.Name)
	} else {
		label = fmt.Sprintf("    --%-16s", f.Name)
	}

	var valueDisplay string
	switch {
	case isEditing:
		// Overhead outside the value display is the 2-char cursor prefix
		// ("> "), the label, and the single space between them.
		avail := maxWidth - len(label) - 3
		if maxWidth <= 0 {
			avail = len(t.buffer) + 4
		}
		valueDisplay = t.renderEditBuffer(avail)
	case t.value != "":
		valueDisplay = "[" + t.value + "]"
	case f.DefValue != "" && f.DefValue != "0" && f.DefValue != "false" && f.DefValue != "[]":
		valueDisplay = "[" + f.DefValue + "]"
	default:
		valueDisplay = "[        ]"
	}

	line := label + " " + valueDisplay

	if isCursor {
		line = "> " + line
	} else {
		line = "  " + line
	}

	// While editing, renderEditBuffer already fits the value within the
	// available width, so truncating here would clobber the scrolled buffer
	// and cursor. Only truncate the non-editing display.
	if !isEditing && maxWidth > 0 && len(line) > maxWidth {
		line = line[:maxWidth-3] + "..."
	}

	return line
}

// renderEditBuffer returns the value display with horizontal scrolling so the
// cursor is always visible. It shows "<"/">" overflow indicators when content
// extends beyond the visible window. The returned string is guaranteed to fit
// within availWidth columns (including the surrounding brackets, cursor bar,
// and any overflow markers) so the caller never has to truncate it.
func (t *TextInput) renderEditBuffer(availWidth int) string {
	// Minimum to render "[|]".
	if availWidth < 3 {
		availWidth = 3
	}

	buf := t.buffer
	cur := t.cursor

	// Space available inside the surrounding brackets.
	innerWidth := availWidth - 2

	// If the whole buffer plus the cursor bar fits, show it all.
	if len(buf)+1 <= innerWidth {
		if cur < len(buf) {
			return "[" + buf[:cur] + "|" + buf[cur:] + "]"
		}
		return "[" + buf + "|]"
	}

	// A scrolling window is needed. The cursor bar always occupies one
	// column, leaving the rest for buffer text. Pick a window centered on
	// the cursor, then shrink it to make room for the overflow markers so
	// the total width never exceeds innerWidth.
	window := innerWidth - 1 // reserve the cursor bar
	if window < 0 {
		window = 0
	}

	clamp := func(w int) (start, end int) {
		start = cur - w/2
		if start < 0 {
			start = 0
		}
		end = start + w
		if end > len(buf) {
			end = len(buf)
		}
		start = end - w
		if start < 0 {
			start = 0
		}
		return start, end
	}

	start, end := clamp(window)
	if start > 0 {
		window--
	}
	if end < len(buf) {
		window--
	}
	if window < 0 {
		window = 0
	}
	start, end = clamp(window)

	visible := buf[start:end]
	localCur := cur - start
	if localCur < 0 {
		localCur = 0
	}
	if localCur > len(visible) {
		localCur = len(visible)
	}

	var result string
	if localCur < len(visible) {
		result = visible[:localCur] + "|" + visible[localCur:]
	} else {
		result = visible + "|"
	}

	// Add overflow indicators.
	prefix := "["
	suffix := "]"
	if start > 0 {
		prefix = "[<"
	}
	if end < len(buf) {
		suffix = ">]"
	}

	return prefix + result + suffix
}
