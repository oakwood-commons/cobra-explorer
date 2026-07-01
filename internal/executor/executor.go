package executor

import (
	"bytes"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ExecutionDoneMsg is sent after command execution completes.
type ExecutionDoneMsg struct {
	Output string // combined stdout + stderr
	Err    error
}

// NewInlineExecCommand creates a tea.Cmd that executes the command using Cobra
// in-process, captures all output, and returns it in ExecutionDoneMsg.
func NewInlineExecCommand(root *cobra.Command, args []string) tea.Cmd {
	return func() tea.Msg {
		var stdout, stderr bytes.Buffer

		// Reset all flag state before execution to avoid stale values.
		resetFlagState(root)

		root.SetArgs(args)
		root.SetOut(&stdout)
		root.SetErr(&stderr)

		err := root.Execute()

		// Combine output: stdout first, then stderr if present.
		var combined strings.Builder
		if stdout.Len() > 0 {
			combined.WriteString(stdout.String())
		}
		if stderr.Len() > 0 {
			if combined.Len() > 0 {
				combined.WriteString("\n")
			}
			combined.WriteString(stderr.String())
		}

		return ExecutionDoneMsg{
			Output: combined.String(),
			Err:    err,
		}
	}
}

// resetFlagState recursively resets all flags in the command tree to their
// default values and marks them as unchanged. This prevents stale values
// from previous executions.
func resetFlagState(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})
	for _, child := range cmd.Commands() {
		resetFlagState(child)
	}
}
