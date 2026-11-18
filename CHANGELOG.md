# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- The flags panel no longer swallows global shortcuts: `q` (and every other
  shortcut) now works while the flags panel is focused. Keys are captured by a
  field only after `Enter` starts editing it, and while editing only `Enter` /
  `Tab` (commit), `Esc` (cancel) and `Ctrl+C` (quit) are handled.

### Removed

- **Breaking:** in-TUI command execution. The `WithExecution` option and the
  internal `executor` package are gone, along with the full-screen output view.
  `Enter` in the command bar now always hands the built command back to the
  shell. Remove `explore.WithExecution(...)` from your `explore.NewCommand` /
  `explore.Run` calls.

### Fixed

- A flag with a long usage string no longer overflows the details panel and
  breaks the layout while the flags panel is focused; the text now wraps to the
  panel width.
- The details panel wraps long description text on word boundaries again instead
  of cutting it off at the panel edge, so the end of a long sentence is no
  longer lost (a regression from the Bubble Tea v2 migration).

## [0.2.0] - 2026-08-05

### Changed

- Migrated to the Bubble Tea v2 ecosystem (`charm.land/bubbletea/v2`,
  `charm.land/lipgloss/v2`, `charm.land/bubbles/v2`). This is an internal
  change with no intended difference in behavior or key bindings.
- Minimum Go version raised to 1.25 (required by `charm.land/bubbles/v2`).

## [0.1.1] - 2026-08-05

### Fixed

- In-process execution now captures output written directly to `os.Stdout` and
  `os.Stderr` by command handlers (e.g. `fmt.Println`), not just output routed
  through Cobra's writer. On Unix, file-descriptor-level redirection also
  captures commands that cache `os.Stdout` before running — most notably
  Cobra's generated `completion` command — which previously produced no output.
- Execution result view now spans the full panel width so the themed background
  covers the entire row.

## [0.1.0] - 2026-07-15

### Added

- Interactive TUI command tree navigator with expand/collapse and keyboard navigation
- Detail panel with command documentation (short, long, usage, aliases, examples, subcommands)
- Flag inspector panel showing per-flag metadata (type, default, required, inherited, valid values)
- Command builder with type-aware flag inputs (text, toggle, choice, slice, stepper)
- In-process command execution via Cobra with a scrollable full-screen output view
- Required-flag validation with a live "missing flags" indicator in the command bar
- Clipboard copy support (macOS, Linux, Windows) via build-tagged implementations
- Responsive multi-panel layout with Tab/Shift+Tab focus zone cycling
- Custom proportional scrollbars for the tree, detail, and flags panels
- Theme preset registry (`internal/theme`) with `Register`, `Get`, and `Names`,
  letting contributors add selectable themes from a small color `Palette`
- Built-in theme presets: `dark`, `night`, `light`, `dracula`, `nord`,
  `terminal`, and `terminal-light`
- `night` theme preset — a darker variant of the dark theme with a near-black
  background for low-light environments and OLED displays
- `terminal` and `terminal-light` theme presets — transparent themes that
  inherit the host terminal's background (tuned with light and dark text for
  dark and light terminals respectively)
- Theme showcase in `doc/themes/` with a screenshot of every built-in theme,
  plus a VHS tape (`doc/themes/themes.tape`) to regenerate them
- `Run()` for direct TUI launch
- `NewCommand()` for embedding as a Cobra subcommand
- Functional options: `WithBinaryName`, `WithLightTheme`, `WithShowHidden`,
  `WithExecution`, and `WithThemeName`
- Comprehensive test suite across all packages (builder, clipboard, executor, flaginput, layout, model, scrollbar, theme, tree, and the public `explore` API)

[Unreleased]: https://github.com/oakwood-commons/cobra-explorer/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/oakwood-commons/cobra-explorer/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/oakwood-commons/cobra-explorer/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/oakwood-commons/cobra-explorer/releases/tag/v0.1.0
