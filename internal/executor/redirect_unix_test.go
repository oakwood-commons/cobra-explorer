//go:build !windows

package executor_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/oakwood-commons/cobra-explorer/internal/executor"
)

// TestNewInlineExecCommand_CapturesCompletionScript is the definitive
// regression test for the "(no output)" bug. Cobra's generated `completion`
// command captures os.Stdout once at InitDefaultCompletionCmd time and writes
// the script to that saved reference — bypassing cmd.SetOut entirely. Only
// file-descriptor-level redirection captures it. This test drives the real
// Cobra completion command, not a simulation.
func TestNewInlineExecCommand_CapturesCompletionScript(t *testing.T) {
	root := &cobra.Command{Use: "demo"}
	root.AddCommand(&cobra.Command{
		Use: "sub",
		Run: func(*cobra.Command, []string) {},
	})
	// Create the default completion command now, capturing os.Stdout at init
	// time exactly as a real CLI's first Execute() would.
	root.InitDefaultCompletionCmd()

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"completion", "bash"}))
	assert.NoError(t, done.Err)
	assert.NotEmpty(t, done.Output, "completion script must be captured, not lost to fd 1")
	assert.Contains(t, done.Output, "bash completion V2 for demo")
}

// TestNewInlineExecCommand_CapturesCachedStdoutReference covers the general
// pattern behind the completion bug: a handler that writes to a copy of
// os.Stdout taken before the run. Swapping the os.Stdout variable would miss
// this; redirecting the file descriptor captures it.
func TestNewInlineExecCommand_CapturesCachedStdoutReference(t *testing.T) {
	cached := os.Stdout // captured before execution, like InitDefaultCompletionCmd

	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use: "gen",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Fprintln(cached, "written to a cached os.Stdout")
			return nil
		},
	}
	root.AddCommand(sub)

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"gen"}))
	assert.NoError(t, done.Err)
	assert.Contains(t, done.Output, "written to a cached os.Stdout")
}
