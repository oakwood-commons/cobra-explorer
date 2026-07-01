package clipboard_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/clipboard"
)

func TestNew_ReturnsImplementation(t *testing.T) {
	c := clipboard.New()
	require.NotNil(t, c)
}

func TestClipboard_AvailableIsBoolean(t *testing.T) {
	c := clipboard.New()
	// Available must not panic and returns deterministically.
	_ = c.Available()
}

func TestClipboard_WriteWhenAvailable(t *testing.T) {
	c := clipboard.New()
	if !c.Available() {
		t.Skip("clipboard not available in this environment")
	}
	// Writing should succeed on a system with clipboard support.
	err := c.Write("cobra-explorer-test")
	assert.NoError(t, err)
}

func TestErrUnavailable(t *testing.T) {
	require.Error(t, clipboard.ErrUnavailable)
	assert.Contains(t, clipboard.ErrUnavailable.Error(), "clipboard")
}
