//go:build linux

package clipboard

import (
	"context"
	"os/exec"
	"strings"
)

type linuxClipboard struct {
	cmd string
}

func newPlatformClipboard() Clipboard {
	for _, candidate := range []string{"wl-copy", "xclip", "xsel"} {
		if _, err := exec.LookPath(candidate); err == nil {
			return &linuxClipboard{cmd: candidate}
		}
	}
	return &linuxClipboard{}
}

func (c *linuxClipboard) Write(text string) error {
	if c.cmd == "" {
		return ErrUnavailable
	}

	var cmd *exec.Cmd
	ctx := context.Background()
	switch c.cmd {
	case "wl-copy":
		cmd = exec.CommandContext(ctx, "wl-copy")
	case "xclip":
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard")
	case "xsel":
		cmd = exec.CommandContext(ctx, "xsel", "--clipboard", "--input")
	default:
		return ErrUnavailable
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (c *linuxClipboard) Available() bool {
	return c.cmd != ""
}
