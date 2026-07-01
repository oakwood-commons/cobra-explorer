---
description: "Add a new flag input type to handle a specific pflag type"
---

# Add Flag Input Type

## Context

You are adding a new flag input widget to cobra-explorer. Flag inputs implement the `FlagInput` interface and provide type-aware editing for specific pflag types.

## Steps

1. Check `internal/flaginput/flaginput.go` for the `FlagInput` interface definition
2. Create a new file in `internal/flaginput/<type>.go` implementing:
   - Model struct with state for editing
   - `New(flag FlagInfo) *<Type>Input` constructor
   - `Update(msg tea.Msg) tea.Cmd` — handle keystrokes during editing
   - `View() string` — render the current value with editing state
   - `Value() string` — return the current value as a string
   - `Editing() bool` — whether this input is currently being edited
   - `Focus()` / `Blur()` — enter/exit editing mode
3. Register the type in the dispatcher (`NewFlagInput` or similar) in `flaginput.go`
4. Write table-driven tests covering:
   - Initial state from flag default
   - Key handling during editing
   - Value output after editing
   - Edge cases for the specific type

## Reference Implementations

- `toggle.go` — simplest (bool, Space to toggle)
- `text.go` — text editing with cursor
- `choice.go` — selection from predefined options
- `stepper.go` — numeric increment/decrement
- `slice.go` — multi-value input
