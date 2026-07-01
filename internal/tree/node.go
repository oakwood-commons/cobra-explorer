package tree

import "strings"

// CommandNode is the internal model of a command, decoupled from cobra.Command.
// Immutable after construction.
type CommandNode struct {
	// Identity
	Name     string
	FullPath []string // e.g., ["mycli", "run", "solution"]

	// Documentation
	Short      string
	Long       string
	Example    string
	Aliases    []string
	Deprecated string

	// Metadata
	Hidden   bool
	Runnable bool // true if the command has Run or RunE set
	GroupID  string

	// Flags (local to this command)
	Flags []FlagInfo

	// Persistent flags inherited from ancestors
	InheritedFlags []FlagInfo

	// Flag relationship metadata (Cobra 1.8+)
	RequiredTogether  [][]string
	MutuallyExclusive [][]string

	// Hierarchy
	Children []*CommandNode
	Parent   *CommandNode

	// Annotations from cobra (pass-through)
	Annotations map[string]string

	// Computed fields
	Depth       int
	TotalLeaves int
}

// FlagInfo is a snapshot of a pflag.Flag, decoupled from the live flag set.
type FlagInfo struct {
	Name        string
	Shorthand   string
	Usage       string
	DefValue    string
	Type        string // pflag type name: "string", "bool", "int", "stringSlice", etc.
	Required    bool
	Deprecated  string
	Hidden      bool
	Inherited   bool
	NoOptDefVal string

	// Enum support: if the flag has registered completion with fixed values.
	ValidValues []string
}

// IsLeaf returns true if the node has no visible children.
func (n *CommandNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// CommandString returns the full command path as a string.
func (n *CommandNode) CommandString() string {
	return strings.Join(n.FullPath, " ")
}

// AllFlags returns local flags + inherited flags in display order:
// required first, then local, then inherited.
func (n *CommandNode) AllFlags() []FlagInfo {
	var required, local []FlagInfo
	inherited := make([]FlagInfo, 0, len(n.InheritedFlags))
	for _, f := range n.Flags {
		if f.Hidden {
			continue
		}
		if f.Required {
			required = append(required, f)
		} else {
			local = append(local, f)
		}
	}
	for _, f := range n.InheritedFlags {
		if f.Hidden {
			continue
		}
		inherited = append(inherited, f)
	}
	result := make([]FlagInfo, 0, len(required)+len(local)+len(inherited))
	result = append(result, required...)
	result = append(result, local...)
	result = append(result, inherited...)
	return result
}
