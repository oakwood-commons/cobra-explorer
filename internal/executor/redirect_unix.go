//go:build !windows

package executor

import (
	"os"

	"golang.org/x/sys/unix"
)

// redirectStdio points the process-level stdout and stderr file descriptors at
// the given pipe writers, returning a function that restores the originals.
//
// Because this operates at the file-descriptor level (fd 1 and fd 2), it
// captures every write to those descriptors: direct fmt.Println calls, Cobra's
// writer, and — critically — writers that captured os.Stdout before the
// redirect (such as Cobra's generated `completion` command). A plain swap of
// the os.Stdout Go variable would miss those cached references.
func redirectStdio(outW, errW *os.File) (restore func(), err error) {
	savedOut, err := unix.Dup(unix.Stdout)
	if err != nil {
		return nil, err
	}
	savedErr, err := unix.Dup(unix.Stderr)
	if err != nil {
		_ = unix.Close(savedOut)
		return nil, err
	}

	if err = dup2(int(outW.Fd()), unix.Stdout); err != nil {
		_ = unix.Close(savedOut)
		_ = unix.Close(savedErr)
		return nil, err
	}
	if err = dup2(int(errW.Fd()), unix.Stderr); err != nil {
		_ = dup2(savedOut, unix.Stdout)
		_ = unix.Close(savedOut)
		_ = unix.Close(savedErr)
		return nil, err
	}

	return func() {
		_ = dup2(savedOut, unix.Stdout)
		_ = dup2(savedErr, unix.Stderr)
		_ = unix.Close(savedOut)
		_ = unix.Close(savedErr)
	}, nil
}
