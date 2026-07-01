package layout

/*
Geometry (all values in terminal cells):

	┌─────────────────────────────── termW ──────────────────────────────┐
	│ Header (1 row, plain text, no border)                              │
	├────────────────┬───────────────────────────────────────────────────┤
	│                │  ╭─ desc panel ───────────────────────╮           │
	│ ╭─ tree ─────╮│  │ title (1 row)                      │           │
	│ │ title      ││  │ viewport content ...                │           │
	│ │ tree view  ││  ╰─────────────────────────────────────╯           │
	│ │            ││  ╭─ flags panel ──────────────────────╮           │
	│ │            ││  │ title (1 row)                      │           │
	│ │            ││  │ flag rows ...                       │           │
	│ ╰────────────╯│  ╰─────────────────────────────────────╯           │
	├────────────────┴───────────────────────────────────────────────────┤
	│ ╭─ command bar ───────────────────────────────────────────────────╮│
	│ │ > command preview                                                ││
	│ ╰─────────────────────────────────────────────────────────────────╯│
	│ Footer (1 row, plain text, no border)                              │
	└────────────────────────────────────────────────────────────────────┘

	Vertical budget:  termH = headerH + bodyH + cmdBarOuterH + footerH
	                  termH = 1 + bodyH + 3 + 1
	                  bodyH = termH - 5

	Body left side:   tree outer = bodyH tall, treeOuterW wide
	Body right side:  desc outer + flags outer = bodyH tall, rightOuterW wide
	                  (stacked vertically, must sum to bodyH exactly)

	Horizontal:       treeOuterW + rightOuterW = termW

	For any bordered panel:
	  outerW = innerW + 2  (left border + right border)
	  outerH = innerH + 2  (top border + bottom border)

	Content inside a panel's inner area:
	  Row 0: title line (bold heading)
	  Rows 1..innerH-1: scrollable content (innerH - 1 rows)
*/

// Layout holds INNER dimensions for each panel.
// These are passed directly to lipgloss.Place(innerW, innerH, ...).
// The rendered panel (after border) will be innerW+2 × innerH+2.
type Layout struct {
	TreeInnerW int
	TreeInnerH int

	DescInnerW int
	DescInnerH int

	FlagsInnerW int
	FlagsInnerH int

	CmdBarInnerW int
	CmdBarInnerH int

	TermW int
	TermH int
}

// ContentRows returns the number of scrollable content rows available
// below the title line for a given panel inner height.
func ContentRows(innerH int) int {
	r := innerH - 1 // subtract title row
	if r < 0 {
		return 0
	}
	return r
}

// ContentHints tells Calculate how many content rows each panel needs
// (NOT including the title line — that's handled internally).
type ContentHints struct {
	DescContentRows int // lines of desc text the viewport will display
	FlagContentRows int // number of flag input rows
}

const (
	MinTermW = 60
	MinTermH = 12

	TreeWidthPct = 0.30

	HeaderH      = 1
	FooterH      = 1
	CmdBarInnerH = 1 // command bar always 1 inner row
	BorderSize   = 2 // 1 char on each side (top+bottom or left+right)
)

// Calculate computes inner dimensions from terminal size and content hints.
func Calculate(termW, termH int, hints ContentHints) Layout {
	if termW < MinTermW {
		termW = MinTermW
	}
	if termH < MinTermH {
		termH = MinTermH
	}

	// --- Widths ---
	treeOuterW := int(float64(termW) * TreeWidthPct)
	rightOuterW := termW - treeOuterW

	treeInnerW := treeOuterW - BorderSize
	if treeInnerW < 1 {
		treeInnerW = 1
	}
	rightInnerW := rightOuterW - BorderSize
	if rightInnerW < 1 {
		rightInnerW = 1
	}

	// --- Heights ---
	cmdBarOuterH := CmdBarInnerH + BorderSize // = 3
	bodyH := termH - HeaderH - FooterH - cmdBarOuterH
	if bodyH < 6 {
		bodyH = 6
	}

	// Tree takes full body height.
	treeInnerH := bodyH - BorderSize
	if treeInnerH < 1 {
		treeInnerH = 1
	}

	// Right side: desc and flags stack vertically within bodyH.
	// descOuterH + flagsOuterH = bodyH
	// descInnerH + 2 + flagsInnerH + 2 = bodyH
	// descInnerH + flagsInnerH = bodyH - 4
	rightInnerTotal := bodyH - 2*BorderSize // total inner rows split between desc and flags
	if rightInnerTotal < 2 {
		rightInnerTotal = 2
	}

	// Each panel needs at minimum: 1 title row + 1 content row = 2 inner rows.
	const minPanelInner = 2

	// Desired inner heights: title + content
	descInnerWant := 1 + hints.DescContentRows
	if descInnerWant < minPanelInner {
		descInnerWant = minPanelInner
	}
	flagsInnerWant := 1 + hints.FlagContentRows
	if flagsInnerWant < minPanelInner {
		flagsInnerWant = minPanelInner
	}

	totalWant := descInnerWant + flagsInnerWant

	var descInnerH, flagsInnerH int
	if totalWant <= rightInnerTotal {
		// Both fit. Give each what they want, split surplus equally.
		surplus := rightInnerTotal - totalWant
		descInnerH = descInnerWant + surplus/2
		flagsInnerH = rightInnerTotal - descInnerH
	} else {
		// Not enough space — divide proportionally, respect minimums.
		descInnerH = rightInnerTotal * descInnerWant / totalWant
		if descInnerH < minPanelInner {
			descInnerH = minPanelInner
		}
		flagsInnerH = rightInnerTotal - descInnerH
		if flagsInnerH < minPanelInner {
			flagsInnerH = minPanelInner
			descInnerH = rightInnerTotal - flagsInnerH
		}
	}

	cmdBarInnerW := termW - BorderSize
	if cmdBarInnerW < 1 {
		cmdBarInnerW = 1
	}

	return Layout{
		TreeInnerW:   treeInnerW,
		TreeInnerH:   treeInnerH,
		DescInnerW:   rightInnerW,
		DescInnerH:   descInnerH,
		FlagsInnerW:  rightInnerW,
		FlagsInnerH:  flagsInnerH,
		CmdBarInnerW: cmdBarInnerW,
		CmdBarInnerH: CmdBarInnerH,
		TermW:        termW,
		TermH:        termH,
	}
}
