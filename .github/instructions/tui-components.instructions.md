# cobra-explorer TUI Component Skill

## Description

This skill covers creating and modifying Bubble Tea TUI components within the cobra-explorer project.

## Key Patterns

### Model Structure

Every component follows this pattern:

```go
package mycomponent

import tea "charm.land/bubbletea/v2"

type Model struct {
    // State fields
    width  int
    height int
}

func New(/* dependencies */) Model {
    return Model{}
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}

func (m Model) View() string {
    // Pure rendering — only read state
    return ""
}
```

### Message Passing

Components communicate via typed messages:

```go
// Defined in the SENDING component's package
type SelectionChangedMsg struct {
    NodePath []string
}
```

The parent model (`internal/model`) routes messages between components.

### Styling Rules

- Import theme: `"github.com/oakwood-commons/cobra-explorer/internal/theme"`
- Pass theme to component constructor
- Use theme styles in View(), never create new styles inline
- `lipgloss.Place()` pads but doesn't truncate — always truncate after

### Focus Management

- Focus zones are defined in `internal/model/model.go` (ZoneTree, ZoneDesc, ZoneFlags, ZoneCommand)
- Only the focused zone handles arrow keys and Enter
- Tab cycles between zones
- Focused panels get `PanelBorderFocused` style, unfocused get `PanelBorder`

## Testing

```go
func TestMyComponent_Update(t *testing.T) {
    m := New(/* ... */)
    
    // Send a message
    m, cmd := m.Update(someMsg{})
    
    // Assert state, not View output
    assert.Equal(t, expected, m.someField)
}
```
