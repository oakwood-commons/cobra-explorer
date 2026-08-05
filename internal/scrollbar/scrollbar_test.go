package scrollbar_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/oakwood-commons/cobra-explorer/internal/scrollbar"
)

func TestRender_ZeroHeight(t *testing.T) {
	assert.Equal(t, "", scrollbar.Render(0, 100, 0, lipgloss.NewStyle()))
	assert.Equal(t, "", scrollbar.Render(-1, 100, 0, lipgloss.NewStyle()))
}

func TestRender_NoScrollNeeded(t *testing.T) {
	// totalLines <= viewHeight → blank column with viewHeight rows.
	out := scrollbar.Render(5, 5, 0, lipgloss.NewStyle())
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 5)
	for _, l := range lines {
		assert.NotContains(t, l, "┃")
		assert.NotContains(t, l, "│")
	}
}

func TestRender_ProducesCorrectHeight(t *testing.T) {
	out := scrollbar.Render(10, 100, 0, lipgloss.NewStyle())
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 10)
}

func TestRender_ThumbAtTopWhenOffsetZero(t *testing.T) {
	out := scrollbar.Render(10, 100, 0, lipgloss.NewStyle())
	lines := strings.Split(out, "\n")
	// First line should be a thumb char.
	assert.Contains(t, lines[0], "┃")
	// Last line should be a track char.
	assert.Contains(t, lines[len(lines)-1], "│")
}

func TestRender_ThumbAtBottomWhenFullyScrolled(t *testing.T) {
	view, total := 10, 100
	out := scrollbar.Render(view, total, total-view, lipgloss.NewStyle())
	lines := strings.Split(out, "\n")
	assert.Contains(t, lines[len(lines)-1], "┃")
	assert.Contains(t, lines[0], "│")
}

func TestRender_ThumbHasMinimumSize(t *testing.T) {
	// Very large total → thumb clamped to at least 1 char.
	out := scrollbar.Render(10, 100000, 0, lipgloss.NewStyle())
	assert.Contains(t, out, "┃")
}

func TestRender_ContainsBothTrackAndThumb(t *testing.T) {
	out := scrollbar.Render(20, 200, 50, lipgloss.NewStyle())
	assert.Contains(t, out, "┃", "should have a thumb")
	assert.Contains(t, out, "│", "should have track")
}
