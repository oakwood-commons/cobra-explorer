//go:build linux

package executor

import "golang.org/x/sys/unix"

// dup2 duplicates oldfd onto newfd. Linux arm64/riscv64 lack the dup2 syscall,
// so Dup3 (available on all Linux arches) is used uniformly.
func dup2(oldfd, newfd int) error { return unix.Dup3(oldfd, newfd, 0) }
