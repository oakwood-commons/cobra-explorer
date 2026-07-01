package layout_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oakwood-commons/cobra-explorer/internal/layout"
)

func TestContentRows(t *testing.T) {
	assert.Equal(t, 0, layout.ContentRows(0))
	assert.Equal(t, 0, layout.ContentRows(1))
	assert.Equal(t, 4, layout.ContentRows(5))
	// Negative inner height clamps to 0.
	assert.Equal(t, 0, layout.ContentRows(-3))
}

func TestCalculate_ClampsToMinimums(t *testing.T) {
	ly := layout.Calculate(10, 3, layout.ContentHints{})
	assert.Equal(t, layout.MinTermW, ly.TermW)
	assert.Equal(t, layout.MinTermH, ly.TermH)
}

func TestCalculate_WidthsSumToTermWidth(t *testing.T) {
	ly := layout.Calculate(120, 40, layout.ContentHints{DescContentRows: 5, FlagContentRows: 3})

	treeOuterW := ly.TreeInnerW + layout.BorderSize
	rightOuterW := ly.DescInnerW + layout.BorderSize
	assert.Equal(t, ly.TermW, treeOuterW+rightOuterW)

	// Desc and flags share the same inner width.
	assert.Equal(t, ly.DescInnerW, ly.FlagsInnerW)
}

func TestCalculate_RightPanelHeightsFitBody(t *testing.T) {
	termH := 40
	ly := layout.Calculate(120, termH, layout.ContentHints{DescContentRows: 5, FlagContentRows: 3})

	bodyH := termH - layout.HeaderH - layout.FooterH - (layout.CmdBarInnerH + layout.BorderSize)

	// desc outer + flags outer must equal bodyH.
	descOuterH := ly.DescInnerH + layout.BorderSize
	flagsOuterH := ly.FlagsInnerH + layout.BorderSize
	assert.Equal(t, bodyH, descOuterH+flagsOuterH)

	// Tree outer height also equals bodyH.
	assert.Equal(t, bodyH, ly.TreeInnerH+layout.BorderSize)
}

func TestCalculate_CmdBar(t *testing.T) {
	ly := layout.Calculate(100, 30, layout.ContentHints{})
	assert.Equal(t, layout.CmdBarInnerH, ly.CmdBarInnerH)
	assert.Equal(t, ly.TermW-layout.BorderSize, ly.CmdBarInnerW)
}

func TestCalculate_TinyTerminalStillValid(t *testing.T) {
	ly := layout.Calculate(layout.MinTermW, layout.MinTermH, layout.ContentHints{})
	// All inner dimensions must remain positive.
	assert.GreaterOrEqual(t, ly.TreeInnerW, 1)
	assert.GreaterOrEqual(t, ly.TreeInnerH, 1)
	assert.GreaterOrEqual(t, ly.DescInnerW, 1)
	assert.GreaterOrEqual(t, ly.DescInnerH, layout.ContentRows(2))
	assert.GreaterOrEqual(t, ly.FlagsInnerW, 1)
	assert.GreaterOrEqual(t, ly.CmdBarInnerW, 1)
}

func TestCalculate_LargeContentHintsProportional(t *testing.T) {
	// When content exceeds available space, panels split proportionally
	// while respecting minimums; totals still fit.
	termH := 20
	ly := layout.Calculate(100, termH, layout.ContentHints{DescContentRows: 100, FlagContentRows: 100})

	bodyH := termH - layout.HeaderH - layout.FooterH - (layout.CmdBarInnerH + layout.BorderSize)
	descOuterH := ly.DescInnerH + layout.BorderSize
	flagsOuterH := ly.FlagsInnerH + layout.BorderSize
	assert.Equal(t, bodyH, descOuterH+flagsOuterH)
	assert.GreaterOrEqual(t, ly.DescInnerH, 2)
	assert.GreaterOrEqual(t, ly.FlagsInnerH, 2)
}

func TestCalculate_TreeWidthApproxPercent(t *testing.T) {
	ly := layout.Calculate(200, 40, layout.ContentHints{})
	treeOuterW := ly.TreeInnerW + layout.BorderSize
	expected := int(200 * layout.TreeWidthPct)
	assert.Equal(t, expected, treeOuterW)
}
