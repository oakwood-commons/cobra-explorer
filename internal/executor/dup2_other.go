//go:build !linux && !windows

package executor

import "golang.org/x/sys/unix"

// dup2 duplicates oldfd onto newfd on Darwin and the BSDs, which provide the
// dup2 syscall directly.
func dup2(oldfd, newfd int) error { return unix.Dup2(oldfd, newfd) }
