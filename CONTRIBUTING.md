# Contributing to cobra-explorer

Thank you for your interest in contributing! This document provides guidelines and information for contributors.

## Developer Certificate of Origin (DCO)

This project uses the [Developer Certificate of Origin](https://developercertificate.org/) (DCO). All commits must be signed off to certify that you have the right to submit the contribution under the project's license.

Sign off your commits with:

```bash
git commit -s -m "Your commit message"
```

This adds a `Signed-off-by` line to your commit message. If you forget, you can amend:

```bash
git commit --amend -s
```

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/<your-username>/cobra-explorer.git
   cd cobra-explorer
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests to verify your setup:
   ```bash
   task test
   ```

## Development Workflow

### Branch Naming

- `feature/short-description` — new features
- `fix/short-description` — bug fixes
- `docs/short-description` — documentation changes

### Making Changes

1. Create a branch from `main`
2. Make your changes
3. Write or update tests as appropriate
4. Run the full CI check locally:
   ```bash
   task ci
   ```
5. Commit with a clear message and sign off (`-s` flag)
6. Push and open a pull request

### Running CI Locally

This project uses [Task](https://taskfile.dev) as its task runner:

```bash
task ci          # Run all checks (fmt, vet, lint, test, build)
task test        # Run tests with race detection
task lint        # Run golangci-lint
task cover       # Generate coverage report
task fmt         # Format code
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add fuzzy search to tree navigator
fix: correct flag value parsing for duration type
docs: update README with new options
refactor: extract layout calculation into helper
test: add integration tests for executor
```

## Code Guidelines

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective-go) and the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Run `gofmt` and `golangci-lint` before committing (`task ci`)
- Keep the public API surface minimal — use `internal/` for implementation details

### Architecture

- The public API is defined in the root package (`explore.go`, `options.go`, `doc.go`)
- All implementation lives under `internal/`
- TUI components follow the [Bubble Tea](https://github.com/charmbracelet/bubbletea) Elm architecture (Model, Update, View)
- The Cobra command tree is introspected but never mutated

### Testing

- Unit tests live alongside the code they test (`*_test.go`)
- Use `testdata/` for test fixtures
- Test TUI components by calling `Update()` with messages and asserting model state
- Avoid testing `View()` output directly (too brittle) — test the model state instead

## Project Structure

```
cobra-explorer/
├── explore/            # Public API: Run(), NewCommand(), Options
│   ├── explore.go
│   ├── options.go
│   └── doc.go
├── internal/
│   ├── model/          # Root Bubble Tea model
│   ├── tree/           # Command tree data structure + navigation
│   ├── builder/        # Command builder (BuiltCommand)
│   ├── flaginput/      # Type-aware flag input widgets
│   ├── executor/       # In-process command execution
│   ├── clipboard/      # OS-specific clipboard support
│   ├── theme/          # Visual themes and styles
│   ├── layout/         # Panel sizing and responsive layout
│   └── scrollbar/      # Custom scrollbar component
├── examples/
│   └── basic/          # Example integration
└── design/             # Design documents
```

## Reporting Issues

- Use GitHub Issues
- Include Go version, OS, and terminal emulator
- For TUI rendering issues, include a screenshot or recording

## Questions?

Open a Discussion on GitHub or comment on an existing issue.
