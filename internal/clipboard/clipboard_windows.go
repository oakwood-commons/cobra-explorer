//go:build windows

package clipboard

import (
	"context"
	"os/exec"
	"strings"
)

type windowsClipboard struct{}

func newPlatformClipboard() Clipboard {
	return &windowsClipboard{}
}

func (c *windowsClipboard) Write(text string) error {
	cmd := exec.CommandContext(context.Background(), "clip.exe")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (c *windowsClipboard) Available() bool {
	_, err := exec.LookPath("clip.exe")
	return err == nil
}
