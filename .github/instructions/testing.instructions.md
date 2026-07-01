# Testing Conventions

## Description

Testing patterns and conventions for cobra-explorer.

## General Rules

- Test model state transitions, NOT View() string output
- Use table-driven tests for multiple input scenarios
- Create Cobra command fixtures for integration tests
- Platform-specific code (clipboard) uses build tags and mocks

## Test Structure

```go
func TestTreeModel_Navigation(t *testing.T) {
    // Arrange: build a fixture command tree
    root := &cobra.Command{Use: "root"}
    child1 := &cobra.Command{Use: "child1", Short: "First child"}
    child2 := &cobra.Command{Use: "child2", Short: "Second child"}
    root.AddCommand(child1, child2)

    treeRoot := tree.BuildTree(root, tree.BuildOptions{})
    m := tree.NewModel(treeRoot, theme.Default())

    // Act: send a key message
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})

    // Assert: verify state
    assert.Equal(t, "child2", m.Selected().Name)
}
```

## Table-Driven Tests

```go
func TestFlagInput_Types(t *testing.T) {
    tests := []struct {
        name     string
        flagType string
        input    tea.KeyMsg
        expected string
    }{
        {"bool toggle", "bool", tea.KeyMsg{Type: tea.KeySpace}, "true"},
        {"int input", "int", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}, "5"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Running Tests

```bash
# All tests
make test

# Specific package
go test -v ./internal/tree/...

# With coverage
make cover

# Race detection (always in CI)
go test -race ./...
```

## What to Test

| Component | Test Focus |
|-----------|-----------|
| Tree model | Navigation state, expand/collapse, cursor position |
| Flag inputs | Value changes from keystrokes, validation |
| Builder | Command string serialization |
| Executor | Args construction, output capture |
| Layout | Panel dimensions at various terminal sizes |

## What NOT to Test

- Exact View() string output (too brittle, changes with styling)
- Lipgloss rendering (trust the library)
- Platform-specific clipboard operations (mock at interface boundary)
