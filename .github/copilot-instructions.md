# Copilot Instructions for cobra-explorer

## Project Overview

cobra-explorer is a Go library that provides an interactive TUI (Terminal User Interface) for any Cobra-based CLI application. It allows users to visually browse command trees, inspect flags, build commands interactively, and execute them.

## Architecture

### Public API (`explore/`)

The public API surface is intentionally minimal and lives in `explore/`:

- `explore/explore.go` — `Run()` and `NewCommand()` entry points
- `explore/options.go` — Functional options (`WithBinaryName`, `WithThemeName`, `WithShowHidden`, `WithExecution`, `WithLightTheme`)
- `explore/doc.go` — Package-level documentation

Import path: `github.com/oakwood-commons/cobra-explorer/explore`

**Rule:** Never add new exported symbols to `explore/` without explicit discussion. The public API should stay minimal.

### Internal Packages (`internal/`)

All implementation lives under `internal/` to prevent consumers from depending on implementation details:

| Package | Responsibility |
|---------|---------------|
| `internal/model` | Root Bubble Tea model (Init, Update, View), focus management |
| `internal/tree` | Command tree data structure, Cobra introspection, navigation |
| `internal/builder` | BuiltCommand type, command string serialization |
| `internal/flaginput` | Type-aware flag input widgets (text, toggle, choice, slice, stepper) |
| `internal/executor` | In-process Cobra command execution |
| `internal/clipboard` | OS-specific clipboard support (build tags) |
| `internal/theme` | Theme struct and built-in themes (dark/light) |
| `internal/layout` | Panel sizing and responsive layout calculations |
| `internal/scrollbar` | Custom scrollbar rendering component |

### Design Document

The `design/` directory contains the initial design spec:
- `initial-design-document.md` — The original architecture, package layout, key types, focus zones, key bindings, flag inputs, and execution model. This is a historical snapshot of the initial design and is **not** kept in sync with ongoing changes, so it may diverge from current behavior.

**Consult the initial design document** for architectural background and intent, but treat the code and `README.md` as the source of truth for current behavior.

## Coding Conventions

### Go Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `golangci-lint` with the project's `.golangci.yml` configuration
- Prefer returning errors over panicking
- Use `fmt.Errorf("cobra-explorer: %w", err)` for error wrapping

### Bubble Tea Patterns

- Every TUI component follows the Elm architecture: Model struct + `Init()`, `Update(msg)`, `View()` methods
- Components communicate via typed messages (`tea.Msg` implementations) — never share mutable state
- Use `tea.Cmd` for async operations, never use goroutines directly in Update
- Keep `View()` functions pure — they should only read model state and return a string
- Handle `tea.WindowSizeMsg` in every component that does layout

### Cobra Integration

- The library NEVER mutates the Cobra command tree — it's read-only introspection
- Use `cmd.Commands()` to get subcommands, `cmd.LocalFlags()` and `cmd.InheritedFlags()` for flags
- Required flag detection uses `f.Annotations[cobra.BashCompOneRequiredFlag]`
- The "explore" command itself is excluded from the tree to avoid recursion

### Lipgloss / Styling

- All styles are defined in `internal/theme/theme.go` — never use inline `lipgloss.NewStyle()` in View functions
- Use `lipgloss.Place()` for positioning, but always truncate output to panel height after calling it (it pads but doesn't truncate)
- Panel dimensions: `lipgloss.Style.Width(w).Height(h)` sets INNER content dimensions; total size includes borders

### Testing

- Test model state, not View() string output (too brittle)
- Use table-driven tests for input handling across flag types
- Test tree building with Cobra command fixtures
- Platform-specific clipboard code uses build tags — test with mocks

### File Organization

- New TUI components go in their own package under `internal/`
- Messages shared between components belong in the component that sends them
- Keep the root package files minimal — route to internal packages immediately

### Documentation Sync

- Whenever a change alters how the application is interacted with or how it behaves (key bindings, flag input behavior, panel navigation, options, execution flow, visible output, etc.), check the change against `README.md`.
- If the observable behavior described in `README.md` no longer matches, update `README.md` in the same change so the docs stay accurate.
- This applies to behavior changes even when not explicitly asked to update docs — treat the doc update as part of completing the behavior change.

## Common Tasks

### Adding a new functional option

1. Add the field to `config` struct in `explore/options.go`
2. Create `With<Name>` function returning `Option`
3. Handle the new config field in `explore/explore.go` when building `model.Options`
4. Update README.md options table

### Adding a new flag input type

1. Implement the `FlagInput` interface in `internal/flaginput/`
2. Add the type detection logic in `flaginput.go` dispatcher
3. Add tests with representative flag definitions

### Adding a new panel/component

1. Create a new package under `internal/`
2. Implement `Init()`, `Update()`, `View()` methods
3. Define message types for communication with parent model
4. Integrate in `internal/model/model.go`'s Update and View

## Dependencies

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/bubbles` — Reusable TUI components
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/spf13/cobra` — The CLI framework being explored (peer dependency)
- `github.com/spf13/pflag` — Flag handling (transitive via Cobra)

Keep the dependency footprint minimal. Only add new dependencies from the charmbracelet ecosystem unless there's a strong reason otherwise.
