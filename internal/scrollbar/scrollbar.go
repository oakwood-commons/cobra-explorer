package scrollbar

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Render produces a vertical scrollbar string of the given height.
// It shows a thumb proportional to the visible fraction and positioned
// according to the scroll offset. If all content fits (totalLines <= viewHeight),
// it returns an empty column (spaces) so layout is preserved.
func Render(viewHeight, totalLines, scrollOffset int, style lipgloss.Style) string {
	if viewHeight <= 0 {
		return ""
	}
	if totalLines <= viewHeight {
		// No scrolling needed — return blank column
		return strings.Repeat(" \n", viewHeight-1) + " "
	}

	// Thumb size: proportional to visible fraction, minimum 1
	thumbSize := viewHeight * viewHeight / totalLines
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize >= viewHeight {
		thumbSize = viewHeight - 1
	}

	// Thumb position
	maxOffset := totalLines - viewHeight
	if maxOffset < 1 {
		maxOffset = 1
	}
	trackSpace := viewHeight - thumbSize
	thumbPos := scrollOffset * trackSpace / maxOffset

	if thumbPos < 0 {
		thumbPos = 0
	}
	if thumbPos > trackSpace {
		thumbPos = trackSpace
	}

	var lines []string
	for i := range viewHeight {
		if i >= thumbPos && i < thumbPos+thumbSize {
			lines = append(lines, style.Render("┃"))
		} else {
			lines = append(lines, style.Render("│"))
		}
	}
	return strings.Join(lines, "\n")
}
