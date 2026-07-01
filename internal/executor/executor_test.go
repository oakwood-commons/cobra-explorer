package executor_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/executor"
)

// runCmd invokes the tea.Cmd and returns the resulting ExecutionDoneMsg.
func runCmd(t *testing.T, cmd tea.Cmd) executor.ExecutionDoneMsg {
	t.Helper()
	require.NotNil(t, cmd)
	msg := cmd()
	done, ok := msg.(executor.ExecutionDoneMsg)
	require.True(t, ok, "expected ExecutionDoneMsg, got %T", msg)
	return done
}

func TestNewInlineExecCommand_CapturesStdout(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use: "greet",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Print("hello world")
			return nil
		},
	}
	root.AddCommand(sub)

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"greet"}))
	assert.NoError(t, done.Err)
	assert.Contains(t, done.Output, "hello world")
}

func TestNewInlineExecCommand_CapturesStderr(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use: "warn",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.PrintErr("a warning")
			return nil
		},
	}
	root.AddCommand(sub)

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"warn"}))
	assert.NoError(t, done.Err)
	assert.Contains(t, done.Output, "a warning")
}

func TestNewInlineExecCommand_CombinesStdoutAndStderr(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use: "both",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Print("OUT")
			cmd.PrintErr("ERR")
			return nil
		},
	}
	root.AddCommand(sub)

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"both"}))
	assert.Contains(t, done.Output, "OUT")
	assert.Contains(t, done.Output, "ERR")
	// stdout comes before stderr.
	assert.Less(t, indexOf(done.Output, "OUT"), indexOf(done.Output, "ERR"))
}

func TestNewInlineExecCommand_ReturnsError(t *testing.T) {
	wantErr := errors.New("boom")
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use:           "fail",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          func(_ *cobra.Command, _ []string) error { return wantErr },
	}
	root.AddCommand(sub)

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"fail"}))
	require.Error(t, done.Err)
	assert.Contains(t, done.Err.Error(), "boom")
}

func TestNewInlineExecCommand_PassesFlags(t *testing.T) {
	var gotName string
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use: "deploy",
		RunE: func(cmd *cobra.Command, _ []string) error {
			gotName, _ = cmd.Flags().GetString("name")
			return nil
		},
	}
	sub.Flags().String("name", "", "name")
	root.AddCommand(sub)

	done := runCmd(t, executor.NewInlineExecCommand(root, []string{"deploy", "--name", "prod"}))
	assert.NoError(t, done.Err)
	assert.Equal(t, "prod", gotName)
}

func TestNewInlineExecCommand_ResetsFlagStateBetweenRuns(t *testing.T) {
	var seen []string
	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{
		Use: "deploy",
		RunE: func(cmd *cobra.Command, _ []string) error {
			v, _ := cmd.Flags().GetString("name")
			seen = append(seen, v)
			return nil
		},
	}
	sub.Flags().String("name", "", "name")
	root.AddCommand(sub)

	// First run sets the flag.
	runCmd(t, executor.NewInlineExecCommand(root, []string{"deploy", "--name", "prod"}))
	// Second run omits it; state should reset to default (empty).
	runCmd(t, executor.NewInlineExecCommand(root, []string{"deploy"}))

	require.Len(t, seen, 2)
	assert.Equal(t, "prod", seen[0])
	assert.Equal(t, "", seen[1], "flag state should reset between executions")
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
