# Cobra Integration Patterns

## Description

Guidelines for working with Cobra command trees in cobra-explorer.

## Core Principle

**cobra-explorer NEVER mutates the Cobra command tree.** All access is read-only introspection.

## Tree Building

The tree is built once at startup in `internal/tree/builder.go`:

```go
tree.BuildTree(root, tree.BuildOptions{
    ShowHidden:   opts.ShowHidden,
    ExcludeNames: []string{"explore"}, // exclude ourselves
})
```

## Accessing Command Data

```go
// Subcommands
children := cmd.Commands()

// Local flags (defined on this command)
cmd.LocalFlags().VisitAll(func(f *pflag.Flag) { ... })

// Inherited flags (from parent PersistentFlags)
cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) { ... })

// Check if runnable
cmd.Runnable() // has Run or RunE

// Required flag detection
f.Annotations[cobra.BashCompOneRequiredFlag]
```

## In-Process Execution

When executing a built command (`internal/executor/`):

1. Construct args from `BuiltCommand`
2. Set `root.SetArgs(args)`
3. Capture stdout/stderr with `root.SetOut()` / `root.SetErr()`
4. Call `root.Execute()`
5. Reset output writers after execution

## Adding New Cobra Data to the Tree

If you need to expose additional Cobra metadata:

1. Add the field to `tree.CommandNode` in `internal/tree/node.go`
2. Populate it in `BuildTree()` in `internal/tree/builder.go`
3. Display it in the appropriate panel (usually detail panel)
