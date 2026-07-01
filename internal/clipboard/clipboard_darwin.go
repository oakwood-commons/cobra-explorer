//go:build darwin

package clipboard

import (
	"context"
	"os/exec"
	"strings"
)

type darwinClipboard struct{}

func newPlatformClipboard() Clipboard {
	return &darwinClipboard{}
}

func (c *darwinClipboard) Write(text string) error {
	cmd := exec.CommandContext(context.Background(), "pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (c *darwinClipboard) Available() bool {
	_, err := exec.LookPath("pbcopy")
	return err == nil
}
