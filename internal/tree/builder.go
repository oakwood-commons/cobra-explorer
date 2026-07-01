package tree

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// BuildOptions controls what is included during introspection.
type BuildOptions struct {
	ShowHidden     bool
	ShowDeprecated bool
	BinaryName     string
	ExcludeNames   []string // Command names to exclude from the tree
}

// BuildTree walks the cobra.Command hierarchy and produces an immutable
// CommandNode tree. This is called once at startup.
func BuildTree(root *cobra.Command, opts BuildOptions) *CommandNode {
	binaryName := opts.BinaryName
	if binaryName == "" {
		binaryName = root.Name()
	}
	return buildNode(root, nil, []string{binaryName}, 0, opts)
}

func buildNode(
	cmd *cobra.Command,
	parent *CommandNode,
	path []string,
	depth int,
	opts BuildOptions,
) *CommandNode {
	node := &CommandNode{
		Name:        cmd.Name(),
		FullPath:    append([]string{}, path...),
		Short:       cmd.Short,
		Long:        cmd.Long,
		Example:     cmd.Example,
		Aliases:     cmd.Aliases,
		Deprecated:  cmd.Deprecated,
		Hidden:      cmd.Hidden,
		Runnable:    cmd.Runnable(),
		GroupID:     cmd.GroupID,
		Annotations: cmd.Annotations,
		Parent:      parent,
		Depth:       depth,
	}

	// Extract local flags
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !opts.ShowHidden && f.Hidden {
			return
		}
		node.Flags = append(node.Flags, flagFromPflag(f, false))
	})

	// Extract inherited persistent flags
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !opts.ShowHidden && f.Hidden {
			return
		}
		node.InheritedFlags = append(node.InheritedFlags, flagFromPflag(f, true))
	})

	// Recurse into subcommands
	for _, child := range cmd.Commands() {
		if !opts.ShowHidden && child.Hidden {
			continue
		}
		if !opts.ShowDeprecated && child.Deprecated != "" {
			continue
		}
		if isExcluded(child.Name(), opts.ExcludeNames) {
			continue
		}
		childPath := make([]string, len(path)+1)
		copy(childPath, path)
		childPath[len(path)] = child.Name()
		childNode := buildNode(child, node, childPath, depth+1, opts)
		node.Children = append(node.Children, childNode)
	}

	node.TotalLeaves = countLeaves(node)

	return node
}

func flagFromPflag(f *pflag.Flag, inherited bool) FlagInfo {
	fi := FlagInfo{
		Name:        f.Name,
		Shorthand:   f.Shorthand,
		Usage:       f.Usage,
		DefValue:    f.DefValue,
		Type:        f.Value.Type(),
		Required:    isRequired(f),
		Deprecated:  f.Deprecated,
		Hidden:      f.Hidden,
		Inherited:   inherited,
		NoOptDefVal: f.NoOptDefVal,
	}

	// Check for registered valid values via completion annotations.
	if ann, ok := f.Annotations[cobra.BashCompCustom]; ok {
		fi.ValidValues = ann
	}

	return fi
}

func isRequired(f *pflag.Flag) bool {
	ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
	return ok && len(ann) > 0 && ann[0] == "true"
}

func countLeaves(node *CommandNode) int {
	if node.IsLeaf() {
		if node.Runnable {
			return 1
		}
		return 0
	}
	count := 0
	for _, child := range node.Children {
		count += child.TotalLeaves
	}
	if node.Runnable {
		count++
	}
	return count
}

func isExcluded(name string, excludes []string) bool {
	for _, ex := range excludes {
		if name == ex {
			return true
		}
	}
	return false
}
