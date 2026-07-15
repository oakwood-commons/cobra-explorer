package builder

import (
	"fmt"
	"strings"

	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// BuiltCommand holds the user's assembled command state.
type BuiltCommand struct {
	Node       *tree.CommandNode
	FlagValues map[string]string
	Args       []string
}

// NewBuiltCommand creates a BuiltCommand for the given node with no values set.
func NewBuiltCommand(node *tree.CommandNode) *BuiltCommand {
	return &BuiltCommand{
		Node:       node,
		FlagValues: make(map[string]string),
	}
}

// SetFlag sets a flag value. Empty string clears it.
func (bc *BuiltCommand) SetFlag(name, value string) {
	if value == "" {
		delete(bc.FlagValues, name)
	} else {
		bc.FlagValues[name] = value
	}
}

// ToArgs produces the argument slice for cobra (without the binary name).
func (bc *BuiltCommand) ToArgs() []string {
	var args []string

	// Command path (skip binary name which is index 0)
	if len(bc.Node.FullPath) > 1 {
		args = append(args, bc.Node.FullPath[1:]...)
	}

	// Flags in deterministic order (follow AllFlags order)
	allFlags := bc.Node.AllFlags()
	for _, f := range allFlags {
		val, ok := bc.FlagValues[f.Name]
		if !ok || val == "" {
			continue
		}

		// Prefer the shorthand form when the flag defines one.
		flagArg := "--" + f.Name
		if f.Shorthand != "" {
			flagArg = "-" + f.Shorthand
		}

		switch f.Type {
		case "bool":
			if val == "true" {
				args = append(args, flagArg)
			}
		case "stringSlice", "stringArray":
			for _, v := range strings.Split(val, ",") {
				v = strings.TrimSpace(v)
				if v != "" {
					args = append(args, flagArg, v)
				}
			}
		default:
			args = append(args, flagArg, val)
		}
	}

	// Positional args
	args = append(args, bc.Args...)

	return args
}

// String renders the full command as a copy-pasteable string.
func (bc *BuiltCommand) String() string {
	args := bc.ToArgs()
	parts := make([]string, 0, 1+len(args))
	parts = append(parts, bc.Node.FullPath[0])
	parts = append(parts, args...)

	var quoted []string
	for _, p := range parts {
		if strings.ContainsAny(p, " \t\"'\\") {
			quoted = append(quoted, fmt.Sprintf("%q", p))
		} else {
			quoted = append(quoted, p)
		}
	}
	return strings.Join(quoted, " ")
}

// UnsetRequiredFlags returns names of required flags that have no value.
func (bc *BuiltCommand) UnsetRequiredFlags() []string {
	var missing []string
	for _, f := range bc.Node.Flags {
		if f.Required {
			if val, ok := bc.FlagValues[f.Name]; !ok || val == "" {
				missing = append(missing, f.Name)
			}
		}
	}
	return missing
}

// IsValid returns true if all required flags are set.
func (bc *BuiltCommand) IsValid() bool {
	return len(bc.UnsetRequiredFlags()) == 0
}
