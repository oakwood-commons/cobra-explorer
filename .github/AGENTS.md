# Agents

## Feature Developer

When implementing a new feature for cobra-explorer:

1. Read the relevant design documents in `design/` first
2. Check if the feature affects the public API (`explore/`) — if so, keep changes minimal
3. Implement in the appropriate `internal/` package
4. Follow Bubble Tea patterns (Model/Update/View, message passing)
5. Write tests that verify model state transitions
6. Update CHANGELOG.md under `[Unreleased]`
7. If adding a new option, update the README options table

## Bug Fixer

When fixing a bug:

1. Write a failing test that reproduces the issue
2. Fix the minimal code to make the test pass
3. Verify no regressions with `task ci`
4. Add a changelog entry under `### Fixed`

## TUI Component Author

When creating a new TUI component:

1. Read `design/initial-design-document.md` for the original architecture and component conventions (historical snapshot — verify against current code)
2. Create a package under `internal/` with its own model
3. Follow these files as reference patterns:
   - `internal/tree/model.go` — complex component with navigation
   - `internal/flaginput/toggle.go` — simple widget
4. Define typed messages for inter-component communication
5. Integrate into `internal/model/model.go`
6. Never use `lipgloss.NewStyle()` in View — use the theme

## Reviewer

When reviewing changes:

1. Ensure no new exports in `explore/` without justification
2. Verify Cobra command tree is never mutated
3. Check that `tea.Cmd` is used for side effects, not goroutines
4. Confirm lipgloss usage follows the panel sizing rules in copilot-instructions.md
5. Verify DCO sign-off on all commits
