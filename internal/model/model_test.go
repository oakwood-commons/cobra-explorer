package model_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/model"
	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// fixtureRoot builds a cobra tree with a runnable command that has flags
// (deploy) and one without flags (status).
func fixtureRoot() *cobra.Command {
	root := &cobra.Command{Use: "mycli", Short: "My CLI"}

	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the app",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
	deploy.Flags().StringP("name", "n", "", "deployment name")
	_ = deploy.MarkFlagRequired("name")
	deploy.Flags().BoolP("verbose", "v", false, "verbose output")
	deploy.Flags().String("region", "us-east", "target region")

	status := &cobra.Command{
		Use:   "status",
		Short: "Show status",
		RunE:  func(cmd *cobra.Command, _ []string) error { cmd.Print("all good"); return nil },
	}

	root.AddCommand(deploy, status)
	return root
}

func findNode(t *testing.T, root *cobra.Command, name string) *tree.CommandNode {
	t.Helper()
	treeRoot := tree.BuildTree(root, tree.BuildOptions{ExcludeNames: []string{"explore"}})
	for _, ch := range treeRoot.Children {
		if ch.Name == name {
			return ch
		}
	}
	t.Fatalf("node %q not found", name)
	return nil
}

// ready constructs a model and sends an initial window size so it is renderable.
func ready(root *cobra.Command, opts model.Options) model.Model {
	m := model.New(root, opts)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return tm.(model.Model)
}

func key(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case " ":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	default:
		r := []rune(s)
		return tea.KeyPressMsg{Code: r[0], Text: s}
	}
}

func TestNew_UsesRootNameAsBinary(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	assert.Contains(t, m.View().Content, "mycli")
}

func TestNew_BinaryNameOverride(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{BinaryName: "custom"})
	assert.Contains(t, m.View().Content, "custom")
}

func TestInit_ReturnsNil(t *testing.T) {
	m := model.New(fixtureRoot(), model.Options{})
	assert.Nil(t, m.Init())
}

func TestView_BeforeReady(t *testing.T) {
	m := model.New(fixtureRoot(), model.Options{})
	assert.Equal(t, "Initializing...", m.View().Content)
}

func TestWindowSize_MakesReady(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	assert.NotEqual(t, "Initializing...", m.View().Content)
	assert.Contains(t, m.View().Content, "Commands")
}

func TestCtrlC_Quits(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	_, cmd := m.Update(key("ctrl+c"))
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestQ_QuitsFromTreeZone(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	_, cmd := m.Update(key("q"))
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestQ_QuitsFromFlagsZone(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)
	// Selecting a command with flags focuses the flags zone.
	require.Contains(t, m.View().Content, "Enter: edit")

	_, cmd := m.Update(key("q"))
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestTab_CyclesZones(t *testing.T) {
	// Select the runnable "deploy" so multiple zones exist.
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// After selecting a command with flags, footer shows the flags-zone hint.
	assert.Contains(t, m.View().Content, "Enter: edit")

	// Tab to the next zone (command bar) → footer hint changes.
	tm, _ = m.Update(key("tab"))
	m = tm.(model.Model)
	assert.Contains(t, m.View().Content, "copy")
}

func TestSelectRunnableWithoutFlags_FocusesCommandZone(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "status")})
	m = tm.(model.Model)
	// Command bar should show the assembled command.
	assert.Contains(t, m.View().Content, "mycli status")
}

func TestCommandZone_EnterEmitsExitCommand(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	// status has no required flags → immediately valid; zone becomes command.
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "status")})
	m = tm.(model.Model)

	_, cmd := m.Update(key("enter"))
	require.NotNil(t, cmd)
	exitMsg, ok := cmd().(model.ExitCommandMsg)
	require.True(t, ok, "expected ExitCommandMsg")
	assert.Equal(t, "mycli status", exitMsg.Command)

	// Feeding ExitCommandMsg back sets ExitCommand and quits.
	tm, quitCmd := m.Update(exitMsg)
	m = tm.(model.Model)
	assert.Equal(t, "mycli status", m.ExitCommand())
	require.NotNil(t, quitCmd)
	assert.IsType(t, tea.QuitMsg{}, quitCmd())
}

func TestCommandZone_CopyShowsFeedback(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "status")})
	m = tm.(model.Model)

	tm, _ = m.Update(key("c"))
	m = tm.(model.Model)
	assert.Contains(t, m.View().Content, "Copied")
}

// TestCommandZone_CopyBadgeVisibleWhenOverflowed verifies that the copied
// confirmation stays visible even when the command overflows the bar and is
// scrolled to the start (the badge is pinned to the right, not appended to the
// scrollable text).
func TestCommandZone_CopyBadgeVisibleWhenOverflowed(t *testing.T) {
	m := model.New(fixtureRoot(), model.Options{})
	// Narrow terminal so the command string overflows the command bar.
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	m = tm.(model.Model)

	// Select deploy, which has a required flag and therefore a long
	// "[missing: name]" suffix that overflows a 40-column bar.
	tm, _ = m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// Focus the command zone.
	for !strings.Contains(m.View().Content, "c: copy") {
		tm, _ = m.Update(key("tab"))
		m = tm.(model.Model)
	}

	// Ensure we are scrolled to the very start.
	tm, _ = m.Update(key("0"))
	m = tm.(model.Model)

	tm, _ = m.Update(key("c"))
	m = tm.(model.Model)

	assert.Contains(t, m.View().Content, "Copied", "copied badge should be visible while scrolled to start")
}

func TestCommandZone_EnterBlockedWhenRequiredFlagMissing(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	// deploy requires --name. Select then tab to command zone.
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)
	tm, _ = m.Update(key("tab")) // flags → command
	m = tm.(model.Model)

	_, cmd := m.Update(key("enter"))
	// Invalid command → no exit command emitted.
	assert.Nil(t, cmd)

	// The command bar warns about the missing flag.
	assert.Contains(t, m.View().Content, "missing")
}

func TestFlagEdit_SetsValue(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// Flags zone, cursor on required "name" (required flags sort first).
	// Enter to begin editing a text flag.
	tm, _ = m.Update(key("enter"))
	m = tm.(model.Model)

	for _, r := range "prod" {
		tm, _ = m.Update(key(string(r)))
		m = tm.(model.Model)
	}
	// Commit with enter.
	tm, _ = m.Update(key("enter"))
	m = tm.(model.Model)

	// Command bar reflects the new flag value using the flag's shorthand.
	assert.Contains(t, m.View().Content, "-n prod")
}

func TestFlagEdit_EscCancels(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	tm, _ = m.Update(key("enter")) // start editing name
	m = tm.(model.Model)
	for _, r := range "temp" {
		tm, _ = m.Update(key(string(r)))
		m = tm.(model.Model)
	}
	tm, _ = m.Update(key("esc")) // cancel
	m = tm.(model.Model)

	// name remains unset → still flagged as missing.
	assert.Contains(t, m.View().Content, "missing")
	assert.NotContains(t, m.View().Content, "--name temp")
}

// TestFlagEdit_QTypesInsteadOfQuitting verifies that once Enter starts editing
// a field, global shortcuts such as "q" belong to the field, and that they work
// again as soon as the edit is committed.
func TestFlagEdit_QTypesInsteadOfQuitting(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	tm, _ = m.Update(key("enter")) // start editing "name"
	m = tm.(model.Model)

	tm, cmd := m.Update(key("q"))
	m = tm.(model.Model)
	assert.Nil(t, cmd, "q must not quit while a field is being edited")

	tm, _ = m.Update(key("enter")) // commit
	m = tm.(model.Model)
	assert.Contains(t, m.View().Content, "-n q", "q should have been typed into the field")

	// The edit is over, so shortcuts work again.
	_, cmd = m.Update(key("q"))
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestFlagEdit_CtrlCQuits(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	tm, _ = m.Update(key("enter")) // start editing "name"
	m = tm.(model.Model)

	_, cmd := m.Update(key("ctrl+c"))
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

// TestFlagEdit_ShortcutsBlockedOnlyWhileEditing verifies the rule for every
// input widget: Enter puts text, numeric, slice, choice and count flags into
// edit mode (where "q" is not a shortcut), while a bool flag is toggled by
// Enter and never enters edit mode (so "q" still quits).
func TestFlagEdit_ShortcutsBlockedOnlyWhileEditing(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(c *cobra.Command)
		wantEditing bool
	}{
		{"text", func(c *cobra.Command) { c.Flags().String("f", "", "text flag") }, true},
		{"int", func(c *cobra.Command) { c.Flags().Int("f", 0, "int flag") }, true},
		{"slice", func(c *cobra.Command) { c.Flags().StringSlice("f", nil, "slice flag") }, true},
		{"count", func(c *cobra.Command) { c.Flags().CountP("f", "f", "count flag") }, true},
		{"choice", func(c *cobra.Command) {
			c.Flags().String("f", "", "choice flag")
			_ = c.Flags().SetAnnotation("f", cobra.BashCompCustom, []string{"a", "b"})
		}, true},
		{"bool", func(c *cobra.Command) { c.Flags().Bool("f", false, "bool flag") }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "mycli"}
			run := &cobra.Command{
				Use:   "run",
				Short: "Run",
				RunE:  func(_ *cobra.Command, _ []string) error { return nil },
			}
			tt.setup(run)
			root.AddCommand(run)

			m := ready(root, model.Options{})
			tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, root, "run")})
			m = tm.(model.Model)

			tm, _ = m.Update(key("enter"))
			m = tm.(model.Model)

			tm, cmd := m.Update(key("q"))
			m = tm.(model.Model)
			if !tt.wantEditing {
				require.NotNil(t, cmd, "q should quit when no field is being edited")
				assert.IsType(t, tea.QuitMsg{}, cmd())
				return
			}
			assert.Nil(t, cmd, "q must not quit while a field is being edited")
			assert.Contains(t, m.View().Content, "Enter/Tab: confirm", "field should still be in edit mode")
		})
	}
}

func TestFlagToggle_Bool(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// Move cursor down to the "verbose" bool flag, then toggle with space.
	// Flags are alphabetical: name(required), region, verbose.
	tm, _ = m.Update(key("j"))
	m = tm.(model.Model)
	tm, _ = m.Update(key("j"))
	m = tm.(model.Model)
	tm, _ = m.Update(key(" "))
	m = tm.(model.Model)

	assert.Contains(t, m.View().Content, "--verbose")
}

func TestHighlight_UpdatesDescription(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandHighlightedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)
	assert.Contains(t, m.View().Content, "Deploy the app")
}

func TestView_ContainsFooterHints(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	// Tree zone is default; footer mentions navigation and quit.
	view := m.View().Content
	assert.True(t, strings.Contains(view, "navigate") || strings.Contains(view, "quit"))
}

func TestFlagsZone_ShowsFlagDetail(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// Cursor starts on the required "name" flag; detail panel shows its metadata.
	view := m.View().Content
	assert.Contains(t, view, "deployment name") // usage
	assert.Contains(t, view, "Required: yes")
}

func TestFlagsZone_DetailShowsDefaultAndInherited(t *testing.T) {
	// Dedicated root with a persistent (inherited) flag on the parent.
	root := &cobra.Command{Use: "mycli"}
	root.PersistentFlags().String("config", "", "config file path")
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
	deploy.Flags().String("region", "us-east", "target region")
	root.AddCommand(deploy)

	m := ready(root, model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, root, "deploy")})
	m = tm.(model.Model)

	// Flags (alphabetical, inherited last): region(0), config(1, inherited).
	// Cursor starts on "region" which has a default value.
	assert.Contains(t, m.View().Content, "Default: us-east")

	// Move to inherited "config" flag.
	tm, _ = m.Update(key("j"))
	m = tm.(model.Model)
	assert.Contains(t, m.View().Content, "Inherited: yes")
}

// TestFlagsZone_LongUsageDoesNotOverflow guards against a regression where a
// flag with a usage string wider than the details panel was rendered unwrapped,
// widening the panel past the terminal and forcing the whole view to wrap onto
// far more rows than the terminal has ("blowing up" the layout).
func TestFlagsZone_LongUsageDoesNotOverflow(t *testing.T) {
	longUsage := "a comma separated list of key=value tags to attach to every " +
		"resource created during this deployment so that they can be tracked " +
		"and billed appropriately across every team in the organization"

	root := &cobra.Command{Use: "mycli"}
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
	deploy.Flags().StringSlice("tags", nil, longUsage)
	root.AddCommand(deploy)

	sizes := []struct{ w, h int }{{120, 40}, {100, 30}, {80, 24}}
	for _, sz := range sizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := model.New(root, model.Options{})
			tm, _ := m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			m = tm.(model.Model)
			// Selecting a runnable command with flags focuses the flags zone,
			// so the details panel renders the flag usage.
			tm, _ = m.Update(tree.CommandSelectedMsg{Node: findNode(t, root, "deploy")})
			m = tm.(model.Model)

			lines := strings.Split(m.View().Content, "\n")
			assert.LessOrEqualf(t, len(lines), sz.h,
				"view has %d rows, exceeds terminal height %d", len(lines), sz.h)
			for i, line := range lines {
				assert.LessOrEqualf(t, lipgloss.Width(line), sz.w,
					"row %d width %d exceeds terminal width %d", i, lipgloss.Width(line), sz.w)
			}
		})
	}
}

// TestDescription_LongLinesWrap guards against a regression where the details
// panel cut long description lines off at the panel edge instead of wrapping
// them, silently dropping the end of the sentence.
func TestDescription_LongLinesWrap(t *testing.T) {
	long := "This sentence is deliberately longer than the details panel so that " +
		"its final words, zebra quokka narwhal, are only visible if the text " +
		"wraps instead of being cut off at the panel edge."

	root := &cobra.Command{Use: "mycli"}
	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Serve",
		Long:  long,
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	})

	sizes := []struct{ w, h int }{{120, 40}, {100, 30}, {80, 24}}
	for _, sz := range sizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := model.New(root, model.Options{})
			tm, _ := m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			m = tm.(model.Model)
			tm, _ = m.Update(tree.CommandHighlightedMsg{Node: findNode(t, root, "serve")})
			m = tm.(model.Model)

			view := m.View().Content
			for _, word := range strings.Fields(long) {
				assert.Containsf(t, view, word, "word %q was cut off", word)
			}
		})
	}
}

func TestTreeZone_NavigationForwardsToTree(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	// In tree zone, "j" forwards to the tree model and emits a highlight cmd.
	_, cmd := m.Update(key("j"))
	require.NotNil(t, cmd)
	msg := cmd()
	hl, ok := msg.(tree.CommandHighlightedMsg)
	require.True(t, ok)

	// Feeding the highlight back updates the description panel.
	tm, _ := m.Update(hl)
	m = tm.(model.Model)
	assert.NotEmpty(t, m.View().Content)
}

func TestDescZone_ScrollingDoesNotPanic(t *testing.T) {
	// Command with a long description so the desc zone is scrollable/focusable.
	root := &cobra.Command{Use: "mycli"}
	long := strings.Repeat("This is a long line of documentation text.\n", 50)
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Docs command",
		Long:  long,
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
	root.AddCommand(cmd)

	m := ready(root, model.Options{})
	tm, _ := m.Update(tree.CommandHighlightedMsg{Node: findNode(t, root, "docs")})
	m = tm.(model.Model)

	// Tab until we reach the desc zone, then scroll.
	for i := 0; i < 3; i++ {
		tm, _ = m.Update(key("tab"))
		m = tm.(model.Model)
	}
	for _, k := range []string{"j", "k", "d", "u"} {
		tm, _ = m.Update(key(k))
		m = tm.(model.Model)
	}
	assert.NotEmpty(t, m.View().Content)
}
