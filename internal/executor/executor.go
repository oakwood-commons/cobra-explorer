package executor

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
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
//
// Output is captured at the OS level for the duration of the run. On Unix this
// redirects the stdout/stderr file descriptors (see redirectStdio), so it
// captures every write to fd 1/2: Cobra's own writer (cmd.Print, usage, help),
// direct writes from handlers via fmt.Println / fmt.Fprintln(os.Stdout, ...),
// and even writers that cached os.Stdout before the run — most notably Cobra's
// generated `completion` command, which grabs os.Stdout once at init time.
func NewInlineExecCommand(root *cobra.Command, args []string) tea.Cmd {
	return func() tea.Msg {
		stdout, stderr, err := runCaptured(root, args)

		// Combine output: stdout first, then stderr if present.
		var combined strings.Builder
		if len(stdout) > 0 {
			combined.Write(stdout)
		}
		if len(stderr) > 0 {
			if combined.Len() > 0 {
				combined.WriteString("\n")
			}
			combined.Write(stderr)
		}

		return ExecutionDoneMsg{
			Output: combined.String(),
			Err:    err,
		}
	}
}

// runCaptured executes the command while capturing everything written to
// os.Stdout and os.Stderr, as well as Cobra's own writers. It returns the raw
// stdout bytes, stderr bytes, and the execution error.
func runCaptured(root *cobra.Command, args []string) (stdoutBytes, stderrBytes []byte, execErr error) {
	// Reset all flag state before execution to avoid stale values.
	resetFlagState(root)
	root.SetArgs(args)

	outR, outW, outErr := os.Pipe()
	errR, errW, errErr := os.Pipe()

	// If the pipes or the OS-level redirect can't be set up, fall back to
	// buffer-only capture so at least Cobra's own writer output is returned.
	if outErr != nil || errErr != nil {
		closeIfNotNil(outR, outW, errR, errW)
		return bufferFallback(root)
	}
	restore, rerr := redirectStdio(outW, errW)
	if rerr != nil {
		closeIfNotNil(outR, outW, errR, errW)
		return bufferFallback(root)
	}

	// Also point Cobra's own writer at the pipes so commands that use
	// cmd.OutOrStdout() are captured uniformly on every platform.
	root.SetOut(outW)
	root.SetErr(errW)

	// Drain pipes concurrently to avoid deadlock when output exceeds the pipe
	// buffer (~64KB) before the writers are closed.
	var outBuf, errBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&outBuf, outR)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&errBuf, errR)
	}()

	execErr = root.Execute()

	// Reset Cobra writers to their defaults, restore the OS streams, then close
	// the pipe writers so the drain goroutines observe EOF, and wait for them.
	root.SetOut(nil)
	root.SetErr(nil)
	restore()
	_ = outW.Close()
	_ = errW.Close()
	wg.Wait()
	_ = outR.Close()
	_ = errR.Close()

	return outBuf.Bytes(), errBuf.Bytes(), execErr
}

// bufferFallback runs the command capturing only Cobra's own writer output. It
// is used when OS-level stream redirection is unavailable.
func bufferFallback(root *cobra.Command) (stdoutBytes, stderrBytes []byte, execErr error) {
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	execErr = root.Execute()
	root.SetOut(nil)
	root.SetErr(nil)
	return outBuf.Bytes(), errBuf.Bytes(), execErr
}

// closeIfNotNil closes any non-nil files, ignoring errors. Used for cleanup on
// the pipe-creation failure path.
func closeIfNotNil(files ...*os.File) {
	for _, f := range files {
		if f != nil {
			_ = f.Close()
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
