package flaginput_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/flaginput"
	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// key builds a tea.KeyMsg from a string for tests.
func key(s string) tea.KeyMsg {
	switch s {
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "ctrl+a":
		return tea.KeyMsg{Type: tea.KeyCtrlA}
	case "ctrl+e":
		return tea.KeyMsg{Type: tea.KeyCtrlE}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// typeString feeds each rune of s as a key into the input.
func typeString(fi flaginput.FlagInput, s string) flaginput.FlagInput {
	for _, r := range s {
		fi = fi.HandleKey(key(string(r)))
	}
	return fi
}

func TestNew_Dispatch(t *testing.T) {
	tests := []struct {
		name     string
		flagType string
		valid    []string
		// probe verifies the returned input behaves like the expected concrete type.
		probe func(t *testing.T, fi flaginput.FlagInput)
	}{
		{"bool→Toggle", "bool", nil, func(t *testing.T, fi flaginput.FlagInput) {
			// Toggle never enters focus.
			assert.False(t, fi.Focus().IsFocused())
			assert.Equal(t, "true", fi.HandleKey(key(" ")).Value())
		}},
		{"int→Text", "int", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.Error(t, fi.SetValue("x").Validate())
		}},
		{"float→Text", "float64", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.Error(t, fi.SetValue("x").Validate())
		}},
		{"duration→Text", "duration", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.Error(t, fi.SetValue("x").Validate())
		}},
		{"count→Stepper", "count", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.Equal(t, "1", fi.HandleKey(key("up")).Value())
		}},
		{"stringSlice→Slice", "stringSlice", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.Equal(t, "a,b", fi.SetValue("a,b").Value())
		}},
		{"stringArray→Slice", "stringArray", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.Equal(t, "a,b", fi.SetValue("a,b").Value())
		}},
		{"string→Text", "string", nil, func(t *testing.T, fi flaginput.FlagInput) {
			assert.True(t, fi.Focus().IsFocused())
		}},
		{"enum→Choice", "string", []string{"a", "b"}, func(t *testing.T, fi flaginput.FlagInput) {
			// Choice selects first option on "down".
			assert.Equal(t, "a", fi.HandleKey(key("down")).Value())
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := flaginput.New(tree.FlagInfo{Name: "f", Type: tt.flagType, ValidValues: tt.valid})
			require.NotNil(t, fi)
			tt.probe(t, fi)
		})
	}
}

func TestValidateInt(t *testing.T) {
	assert.NoError(t, flaginput.ValidateInt(""))
	assert.NoError(t, flaginput.ValidateInt("42"))
	assert.NoError(t, flaginput.ValidateInt("-7"))
	assert.Error(t, flaginput.ValidateInt("abc"))
	assert.Error(t, flaginput.ValidateInt("3.14"))
}

func TestValidateFloat(t *testing.T) {
	assert.NoError(t, flaginput.ValidateFloat(""))
	assert.NoError(t, flaginput.ValidateFloat("3.14"))
	assert.NoError(t, flaginput.ValidateFloat("-2"))
	assert.Error(t, flaginput.ValidateFloat("abc"))
}

func TestValidateDuration(t *testing.T) {
	assert.NoError(t, flaginput.ValidateDuration(""))
	assert.NoError(t, flaginput.ValidateDuration("30s"))
	assert.NoError(t, flaginput.ValidateDuration("5m"))
	assert.Error(t, flaginput.ValidateDuration("later"))
}

// --- TextInput ---

func TestTextInput_TypeAndBlur(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "name", Type: "string"})
	assert.Equal(t, "", fi.Value())
	assert.False(t, fi.IsFocused())

	fi = fi.Focus()
	assert.True(t, fi.IsFocused())
	fi = typeString(fi, "prod")
	// Value only commits on Blur.
	assert.Equal(t, "", fi.Value())

	fi = fi.Blur()
	assert.False(t, fi.IsFocused())
	assert.Equal(t, "prod", fi.Value())
}

func TestTextInput_Editing(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "name", Type: "string"}).Focus()
	fi = typeString(fi, "helo")
	// Move left once, insert 'l' to make "hello".
	fi = fi.HandleKey(key("left"))
	fi = fi.HandleKey(key("l"))
	fi = fi.Blur()
	assert.Equal(t, "hello", fi.Value())
}

func TestTextInput_Backspace(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "n", Type: "string"}).Focus()
	fi = typeString(fi, "abc")
	fi = fi.HandleKey(key("backspace"))
	fi = fi.Blur()
	assert.Equal(t, "ab", fi.Value())
}

func TestTextInput_Delete(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "n", Type: "string"}).Focus()
	fi = typeString(fi, "abc")
	fi = fi.HandleKey(key("home"))
	fi = fi.HandleKey(key("delete"))
	fi = fi.Blur()
	assert.Equal(t, "bc", fi.Value())
}

func TestTextInput_HomeEnd(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "n", Type: "string"}).Focus()
	fi = typeString(fi, "xyz")
	fi = fi.HandleKey(key("home"))
	fi = fi.HandleKey(key("A")) // insert at start
	fi = fi.HandleKey(key("end"))
	fi = fi.HandleKey(key("Z")) // insert at end
	fi = fi.Blur()
	assert.Equal(t, "AxyzZ", fi.Value())
}

func TestTextInput_CtrlU_Clears(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "n", Type: "string"}).Focus()
	fi = typeString(fi, "clearme")
	fi = fi.HandleKey(key("ctrl+u"))
	fi = fi.Blur()
	assert.Equal(t, "", fi.Value())
}

func TestTextInput_SetValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "n", Type: "string"})
	fi = fi.SetValue("preset")
	assert.Equal(t, "preset", fi.Value())
}

func TestTextInput_Validate(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "n", Type: "int"})
	fi = fi.SetValue("abc")
	assert.Error(t, fi.Validate())
	fi = fi.SetValue("5")
	assert.NoError(t, fi.Validate())
}

func TestTextInput_Render(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "name", Shorthand: "n", Type: "string"})
	fi = fi.SetValue("prod")
	out := fi.Render(true, false, 80)
	assert.Contains(t, out, "name")
	assert.Contains(t, out, "prod")
	assert.Contains(t, out, ">")
}

// --- Toggle ---

func TestToggle_ToggleValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "verbose", Type: "bool"})
	assert.Equal(t, "", fi.Value())

	fi = fi.HandleKey(key(" "))
	assert.Equal(t, "true", fi.Value())

	fi = fi.HandleKey(key("enter"))
	assert.Equal(t, "", fi.Value())
}

func TestToggle_DefaultTrue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "bool", DefValue: "true"})
	assert.Equal(t, "true", fi.Value())
}

func TestToggle_SetValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "bool"})
	fi = fi.SetValue("true")
	assert.Equal(t, "true", fi.Value())
	fi = fi.SetValue("")
	assert.Equal(t, "", fi.Value())
}

func TestToggle_NeverFocused(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "bool"})
	assert.False(t, fi.IsFocused())
	assert.False(t, fi.Focus().IsFocused())
	assert.NoError(t, fi.Validate())
}

func TestToggle_Render(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "verbose", Type: "bool"})
	assert.Contains(t, fi.Render(false, false, 80), "[ ]")
	fi = fi.SetValue("true")
	assert.Contains(t, fi.Render(false, false, 80), "[x]")
}

// --- Choice ---

func TestChoice_Navigation(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "level", Type: "string", ValidValues: []string{"low", "mid", "high"}})
	assert.Equal(t, "", fi.Value())

	// Down from unselected selects first.
	fi = fi.HandleKey(key("down"))
	assert.Equal(t, "low", fi.Value())

	fi = fi.HandleKey(key("down"))
	assert.Equal(t, "mid", fi.Value())

	fi = fi.HandleKey(key("up"))
	assert.Equal(t, "low", fi.Value())

	// Can't go above first.
	fi = fi.HandleKey(key("up"))
	assert.Equal(t, "low", fi.Value())
}

func TestChoice_ClampsAtEnd(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "l", Type: "string", ValidValues: []string{"a", "b"}})
	fi = fi.HandleKey(key("down"))
	fi = fi.HandleKey(key("down"))
	fi = fi.HandleKey(key("down"))
	assert.Equal(t, "b", fi.Value())
}

func TestChoice_SetValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "l", Type: "string", ValidValues: []string{"a", "b", "c"}})
	fi = fi.SetValue("c")
	assert.Equal(t, "c", fi.Value())
	// Unknown value resets to none.
	fi = fi.SetValue("zzz")
	assert.Equal(t, "", fi.Value())
}

func TestChoice_FocusBlur(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "l", Type: "string", ValidValues: []string{"a"}})
	assert.False(t, fi.IsFocused())
	fi = fi.Focus()
	assert.True(t, fi.IsFocused())
	fi = fi.Blur()
	assert.False(t, fi.IsFocused())
	assert.NoError(t, fi.Validate())
}

// --- Slice ---

func TestSlice_AddValues(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"}).Focus()
	fi = typeString(fi, "a")
	fi = fi.HandleKey(key(",")) // commit "a"
	fi = typeString(fi, "b")
	fi = fi.Blur() // commits "b"
	assert.Equal(t, "a,b", fi.Value())
}

func TestSlice_BackspaceRemovesLastValueWhenBufferEmpty(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"})
	fi = fi.SetValue("a,b,c")
	fi = fi.Focus() // buffer empty
	fi = fi.HandleKey(key("backspace"))
	fi = fi.Blur()
	assert.Equal(t, "a,b", fi.Value())
}

func TestSlice_SetValueEmptyClears(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"})
	fi = fi.SetValue("x,y")
	assert.Equal(t, "x,y", fi.Value())
	fi = fi.SetValue("")
	assert.Equal(t, "", fi.Value())
}

func TestSlice_Render(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"})
	fi = fi.SetValue("a,b")
	out := fi.Render(false, false, 80)
	assert.Contains(t, out, "a, b")
	assert.NoError(t, fi.Validate())
}

// --- Stepper ---

func TestStepper_IncrementDecrement(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "count"})
	assert.Equal(t, "", fi.Value()) // 0 renders as empty

	fi = fi.HandleKey(key("up"))
	assert.Equal(t, "1", fi.Value())
	fi = fi.HandleKey(key("up"))
	assert.Equal(t, "2", fi.Value())
	fi = fi.HandleKey(key("down"))
	assert.Equal(t, "1", fi.Value())
}

func TestStepper_DoesNotGoNegative(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "count"})
	fi = fi.HandleKey(key("down"))
	assert.Equal(t, "", fi.Value()) // stays at 0
}

func TestStepper_DefaultValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "count", DefValue: "3"})
	assert.Equal(t, "3", fi.Value())
}

func TestStepper_SetValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "v", Type: "count"})
	fi = fi.SetValue("5")
	assert.Equal(t, "5", fi.Value())
	assert.NoError(t, fi.Validate())
	assert.False(t, fi.IsFocused())
}

func TestStepper_Render(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "verbosity", Type: "count"})
	fi = fi.SetValue("2")
	out := fi.Render(true, false, 80)
	assert.Contains(t, out, "2")
	assert.Contains(t, out, "+/-")
}

// --- Accessors & shared rendering behavior ---

func TestFlag_AccessorReturnsMetadata(t *testing.T) {
	types := []struct {
		flagType string
		valid    []string
	}{
		{"string", nil},
		{"bool", nil},
		{"count", nil},
		{"stringSlice", nil},
		{"string", []string{"a", "b"}},
	}
	for _, tt := range types {
		t.Run(tt.flagType, func(t *testing.T) {
			fi := flaginput.New(tree.FlagInfo{Name: "myflag", Type: tt.flagType, ValidValues: tt.valid})
			assert.Equal(t, "myflag", fi.Flag().Name)
			assert.Equal(t, tt.flagType, fi.Flag().Type)
		})
	}
}

func TestRender_NoShorthand(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "config", Type: "string"})
	out := fi.Render(false, false, 80)
	assert.Contains(t, out, "--config")
	assert.NotContains(t, out, "-, ")
}

func TestRender_TruncatesToMaxWidth(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "averylongflagname", Shorthand: "a", Type: "string"})
	fi = fi.SetValue("some-long-value-here")
	out := fi.Render(true, false, 20)
	assert.LessOrEqual(t, len(out), 20)
	assert.Contains(t, out, "...")
}

func TestTextRender_ShowsDefaultValue(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "region", Type: "string", DefValue: "us-east"})
	out := fi.Render(false, false, 80)
	assert.Contains(t, out, "us-east")
}

func TestTextRender_EditingShortBuffer(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "name", Type: "string"}).Focus()
	fi = typeString(fi, "hi")
	out := fi.Render(true, true, 80)
	// Cursor marker should be visible in the edit buffer.
	assert.Contains(t, out, "|")
	assert.Contains(t, out, "hi")
}

func TestTextRender_EditingLongBufferScrolls(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "name", Type: "string"}).Focus()
	fi = typeString(fi, "abcdefghijklmnopqrstuvwxyz0123456789")
	// A width that shows the edit buffer but is narrow enough to force the
	// horizontal scrolling window with an overflow indicator.
	out := fi.Render(true, true, 60)
	assert.Contains(t, out, "|")
	assert.Contains(t, out, "<") // left overflow indicator when cursor at end
}

func TestChoiceRender_EditingShowsCaret(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "level", Type: "string", ValidValues: []string{"low", "high"}})
	fi = fi.SetValue("low")
	out := fi.Render(true, true, 80)
	assert.Contains(t, out, "low")
	assert.Contains(t, out, "v]") // dropdown caret in editing mode
}

func TestChoiceRender_DefaultAndEmpty(t *testing.T) {
	withDef := flaginput.New(tree.FlagInfo{Name: "l", Type: "string", ValidValues: []string{"a"}, DefValue: "a"})
	assert.Contains(t, withDef.Render(false, false, 80), "a")

	empty := flaginput.New(tree.FlagInfo{Name: "l", Type: "string", ValidValues: []string{"a"}})
	assert.Contains(t, empty.Render(false, false, 80), "[")
}

func TestSliceRender_EditingShowsBuffer(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"}).Focus()
	fi = typeString(fi, "a")
	fi = fi.HandleKey(key(",")) // commit "a"
	fi = typeString(fi, "bee")  // pending buffer
	out := fi.Render(true, true, 80)
	assert.Contains(t, out, "a")
	assert.Contains(t, out, "bee")
	assert.Contains(t, out, "|")
}

func TestSliceRender_Empty(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"})
	out := fi.Render(false, false, 80)
	assert.Contains(t, out, "[")
}

func TestSlice_CursorMovement(t *testing.T) {
	fi := flaginput.New(tree.FlagInfo{Name: "tags", Type: "stringSlice"}).Focus()
	fi = typeString(fi, "abc")
	fi = fi.HandleKey(key("left"))
	fi = fi.HandleKey(key("left"))
	fi = fi.HandleKey(key("right"))
	// Insert marker in the middle: buffer "abc", cursor at index 2.
	fi = typeString(fi, "X")
	fi = fi.Blur()
	assert.Equal(t, "abXc", fi.Value())
}
