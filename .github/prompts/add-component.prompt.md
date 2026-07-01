---
description: "Add a new TUI component to the cobra-explorer project"
---

# Add TUI Component

## Context

You are adding a new Bubble Tea component to cobra-explorer. All components live under `internal/` and follow the Elm architecture.

## Steps

1. Read the initial design spec in `design/initial-design-document.md` for the original architecture and component conventions (historical snapshot — verify against current code)
2. Create a new package under `internal/<component-name>/`
3. Define the Model struct with all necessary state
4. Implement these methods:
   - `New(...)` — constructor
   - `Init() tea.Cmd` — initial command (usually nil)
   - `Update(msg tea.Msg) (Model, tea.Cmd)` — handle messages
   - `View() string` — render to string (pure function)
5. Define message types for communication with the parent model
6. Accept the theme as a parameter to `New()` for styling
7. Handle `tea.WindowSizeMsg` for responsive layout
8. Integrate into `internal/model/model.go`:
   - Add the component model as a field
   - Forward relevant messages in `Update()`
   - Call `View()` in the appropriate panel
9. Write tests that verify Update() state transitions

## Rules

- Never use `lipgloss.NewStyle()` in View — all styles come from theme
- Use `tea.Cmd` for async ops — never spawn goroutines in Update
- Keep View() pure — only read state, return string
- Messages flow up via tea.Cmd, down via direct Update() calls
