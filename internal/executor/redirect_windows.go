//go:build windows

package executor

import "os"

// redirectStdio swaps the os.Stdout and os.Stderr variables to the given pipe
// writers, returning a function that restores the originals.
//
// Unlike the Unix implementation this works at the Go-variable level rather
// than the file-descriptor level, so it cannot capture writers that cached the
// original os.Stdout before the redirect (e.g. Cobra's generated `completion`
// command). It does capture the common cases: fmt.Println and Cobra's writer.
func redirectStdio(outW, errW *os.File) (restore func(), err error) {
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	return func() {
		os.Stdout, os.Stderr = origOut, origErr
	}, nil
}
