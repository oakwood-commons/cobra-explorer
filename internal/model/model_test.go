package model_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oakwood-commons/cobra-explorer/internal/executor"
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

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestNew_UsesRootNameAsBinary(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	assert.Contains(t, m.View(), "mycli")
}

func TestNew_BinaryNameOverride(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{BinaryName: "custom"})
	assert.Contains(t, m.View(), "custom")
}

func TestInit_ReturnsNil(t *testing.T) {
	m := model.New(fixtureRoot(), model.Options{})
	assert.Nil(t, m.Init())
}

func TestView_BeforeReady(t *testing.T) {
	m := model.New(fixtureRoot(), model.Options{})
	assert.Equal(t, "Initializing...", m.View())
}

func TestWindowSize_MakesReady(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	assert.NotEqual(t, "Initializing...", m.View())
	assert.Contains(t, m.View(), "Commands")
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

func TestTab_CyclesZones(t *testing.T) {
	// Select the runnable "deploy" so multiple zones exist.
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// After selecting a command with flags, footer shows the flags-zone hint.
	assert.Contains(t, m.View(), "Enter: edit")

	// Tab to the next zone (command bar) → footer hint changes.
	tm, _ = m.Update(key("tab"))
	m = tm.(model.Model)
	assert.Contains(t, m.View(), "copy")
}

func TestSelectRunnableWithoutFlags_FocusesCommandZone(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "status")})
	m = tm.(model.Model)
	// Command bar should show the assembled command.
	assert.Contains(t, m.View(), "mycli status")
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
	assert.Contains(t, m.View(), "Copied")
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
	for !strings.Contains(m.View(), "c: copy") {
		tm, _ = m.Update(key("tab"))
		m = tm.(model.Model)
	}

	// Ensure we are scrolled to the very start.
	tm, _ = m.Update(key("0"))
	m = tm.(model.Model)

	tm, _ = m.Update(key("c"))
	m = tm.(model.Model)

	assert.Contains(t, m.View(), "Copied", "copied badge should be visible while scrolled to start")
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
	assert.Contains(t, m.View(), "missing")
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
	assert.Contains(t, m.View(), "-n prod")
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
	assert.Contains(t, m.View(), "missing")
	assert.NotContains(t, m.View(), "--name temp")
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

	assert.Contains(t, m.View(), "--verbose")
}

func TestExecutionResult_ShowAndDismiss(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})

	tm, _ := m.Update(executor.ExecutionDoneMsg{Output: "run output here"})
	m = tm.(model.Model)
	view := m.View()
	assert.Contains(t, view, "run output here")
	assert.Contains(t, view, "Output")

	// Dismiss with q.
	tm, _ = m.Update(key("q"))
	m = tm.(model.Model)
	assert.NotContains(t, m.View(), "run output here")
}

func TestExecutionResult_ShowsError(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(executor.ExecutionDoneMsg{Output: "", Err: assertErr{}})
	m = tm.(model.Model)
	view := m.View()
	assert.Contains(t, view, "failed")
	assert.Contains(t, view, "boom")
}

func TestExecutionEnabled_EnterRunsCommand(t *testing.T) {
	root := fixtureRoot()
	m := ready(root, model.Options{ExecutionEnabled: true})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, root, "status")})
	m = tm.(model.Model)

	// Footer should indicate "run" rather than paste.
	assert.Contains(t, m.View(), "Enter: run")

	_, cmd := m.Update(key("enter"))
	require.NotNil(t, cmd)
	// Running the command yields an ExecutionDoneMsg with captured output.
	done, ok := cmd().(executor.ExecutionDoneMsg)
	require.True(t, ok)
	assert.Contains(t, done.Output, "all good")
}

func TestHighlight_UpdatesDescription(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandHighlightedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)
	assert.Contains(t, m.View(), "Deploy the app")
}

func TestView_ContainsFooterHints(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	// Tree zone is default; footer mentions navigation and quit.
	view := m.View()
	assert.True(t, strings.Contains(view, "navigate") || strings.Contains(view, "quit"))
}

func TestFlagsZone_ShowsFlagDetail(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	tm, _ := m.Update(tree.CommandSelectedMsg{Node: findNode(t, fixtureRoot(), "deploy")})
	m = tm.(model.Model)

	// Cursor starts on the required "name" flag; detail panel shows its metadata.
	view := m.View()
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
	assert.Contains(t, m.View(), "Default: us-east")

	// Move to inherited "config" flag.
	tm, _ = m.Update(key("j"))
	m = tm.(model.Model)
	assert.Contains(t, m.View(), "Inherited: yes")
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
	assert.NotEmpty(t, m.View())
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
	assert.NotEmpty(t, m.View())
}

func TestExecResult_ScrollKeys(t *testing.T) {
	m := ready(fixtureRoot(), model.Options{})
	long := strings.Repeat("output line\n", 100)
	tm, _ := m.Update(executor.ExecutionDoneMsg{Output: long})
	m = tm.(model.Model)

	for _, k := range []string{"j", "k", "d", "u", "g", "G"} {
		tm, _ = m.Update(key(k))
		m = tm.(model.Model)
	}
	assert.Contains(t, m.View(), "output line")

	// Esc dismisses.
	tm, _ = m.Update(key("esc"))
	m = tm.(model.Model)
	assert.NotContains(t, m.View(), "output line")
}

// assertErr is a stub error used to exercise the failure path.
type assertErr struct{}

func (assertErr) Error() string { return "boom" }
