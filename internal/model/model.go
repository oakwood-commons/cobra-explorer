package model

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spf13/cobra"

	"github.com/oakwood-commons/cobra-explorer/internal/builder"
	"github.com/oakwood-commons/cobra-explorer/internal/clipboard"
	"github.com/oakwood-commons/cobra-explorer/internal/executor"
	"github.com/oakwood-commons/cobra-explorer/internal/flaginput"
	"github.com/oakwood-commons/cobra-explorer/internal/layout"
	"github.com/oakwood-commons/cobra-explorer/internal/scrollbar"
	"github.com/oakwood-commons/cobra-explorer/internal/theme"
	"github.com/oakwood-commons/cobra-explorer/internal/tree"
)

// Focus zones.
const (
	ZoneTree    = 0
	ZoneDesc    = 1
	ZoneFlags   = 2
	ZoneCommand = 3
)

// ExitCommandMsg is sent when the user presses Enter in the command bar.
type ExitCommandMsg struct {
	Command string
}

// Model is the top-level Bubble Tea model.
type Model struct {
	root     *cobra.Command
	treeRoot *tree.CommandNode
	treeM    tree.Model
	descVP   viewport.Model
	clip     clipboard.Clipboard
	ly       layout.Layout
	theme    theme.Theme

	// Current command state.
	built       *builder.BuiltCommand
	flagInputs  []flaginput.FlagInput
	flagCursor  int
	flagScroll  int
	cmdScroll   int
	editing     bool
	descContent string

	// Focus.
	zone  int
	zones []int

	// Status.
	binaryName       string
	showHidden       bool
	executionEnabled bool
	copied           bool
	exitCommand      string
	lastExecOutput   string
	lastExecErr      error
	showExecResult   bool
	execVP           viewport.Model
	width            int
	height           int
	ready            bool
}

// Options configures the model.
type Options struct {
	BinaryName       string
	Theme            theme.Theme
	ThemeSet         bool
	ShowHidden       bool
	ExecutionEnabled bool
}

// New creates a new Model.
func New(root *cobra.Command, opts Options) Model {
	treeRoot := tree.BuildTree(root, tree.BuildOptions{
		ShowHidden:   opts.ShowHidden,
		ExcludeNames: []string{"explore"},
	})
	th := opts.Theme
	if !opts.ThemeSet {
		th = theme.Default()
	}
	binaryName := opts.BinaryName
	if binaryName == "" {
		binaryName = root.Name()
	}

	m := Model{
		root:             root,
		treeRoot:         treeRoot,
		treeM:            tree.NewModel(treeRoot, th),
		descVP:           viewport.New(viewport.WithWidth(80), viewport.WithHeight(10)),
		clip:             clipboard.New(),
		theme:            th,
		binaryName:       binaryName,
		showHidden:       opts.ShowHidden,
		executionEnabled: opts.ExecutionEnabled,
		zone:             ZoneTree,
		zones:            []int{ZoneTree, ZoneCommand},
	}
	m.descVP.Style = m.theme.Body
	m.selectNode(treeRoot)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.reflow()
		if m.showExecResult {
			m.execVP.SetWidth(m.width)
			m.execVP.SetHeight(m.height - 4)
		}
		m.ready = true
		return m, nil

	case tea.KeyPressMsg:
		// Execution result screen: scroll or dismiss.
		if m.showExecResult {
			switch msg.String() {
			case "q", "esc", "enter":
				m.showExecResult = false
				m.lastExecOutput = ""
				m.lastExecErr = nil
			case "j", "down":
				m.execVP.ScrollDown(1)
			case "k", "up":
				m.execVP.ScrollUp(1)
			case "d":
				m.execVP.HalfPageDown()
			case "u":
				m.execVP.HalfPageUp()
			case "g":
				m.execVP.GotoTop()
			case "G":
				m.execVP.GotoBottom()
			}
			return m, nil
		}
		if m.editing {
			return m.handleFlagEdit(msg)
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.zone != ZoneFlags {
				return m, tea.Quit
			}
		case "tab":
			m.cycleZone(1)
			return m, nil
		case "shift+tab":
			m.cycleZone(-1)
			return m, nil
		}
		switch m.zone {
		case ZoneTree:
			return m.handleTree(msg)
		case ZoneDesc:
			return m.handleDesc(msg)
		case ZoneFlags:
			return m.handleFlags(msg)
		case ZoneCommand:
			return m.handleCmd(msg)
		}

	case tree.CommandHighlightedMsg:
		m.selectNode(msg.Node)
		return m, nil

	case tree.CommandSelectedMsg:
		m.selectNode(msg.Node)
		if len(m.flagInputs) > 0 {
			m.zone = ZoneFlags
		} else {
			m.zone = ZoneCommand
		}
		return m, nil

	case ExitCommandMsg:
		m.exitCommand = msg.Command
		return m, tea.Quit

	case executor.ExecutionDoneMsg:
		m.lastExecErr = msg.Err
		m.lastExecOutput = msg.Output
		m.showExecResult = true
		// Set up the execution result viewport. Full width so the themed
		// background covers the entire row; height leaves room for the header,
		// two blank spacer rows, and the footer.
		m.execVP = viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.height-4))
		m.execVP.Style = m.theme.Body
		content := msg.Output
		if msg.Err != nil {
			if content != "" {
				content += "\n"
			}
			content += "Error: " + msg.Err.Error()
		}
		if content == "" {
			content = "(no output)"
		}
		m.execVP.SetContent(content)
		return m, nil
	}
	return m, nil
}

// --- Key handlers ---

func (m Model) handleTree(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.treeM, cmd = m.treeM.Update(tea.Msg(msg))
	return m, cmd
}

func (m Model) handleDesc(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.descVP.ScrollDown(1)
	case "k", "up":
		m.descVP.ScrollUp(1)
	case "d":
		m.descVP.HalfPageDown()
	case "u":
		m.descVP.HalfPageUp()
	}
	return m, nil
}

func (m Model) handleFlags(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if len(m.flagInputs) == 0 {
		return m, nil
	}
	switch msg.String() {
	case "j", "down":
		if m.flagCursor < len(m.flagInputs)-1 {
			m.flagCursor++
			m.scrollFlags()
		}
	case "k", "up":
		if m.flagCursor > 0 {
			m.flagCursor--
			m.scrollFlags()
		}
	case "enter":
		fi := m.flagInputs[m.flagCursor]
		if fi.Flag().Type == "bool" {
			m.toggleBool(m.flagCursor)
		} else {
			m.editing = true
			m.flagInputs[m.flagCursor] = fi.Focus()
		}
	case "space":
		if m.flagInputs[m.flagCursor].Flag().Type == "bool" {
			m.toggleBool(m.flagCursor)
		}
	}
	return m, nil
}

func (m Model) handleFlagEdit(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "tab":
		fi := m.flagInputs[m.flagCursor].Blur()
		m.flagInputs[m.flagCursor] = fi
		m.built.SetFlag(fi.Flag().Name, fi.Value())
		m.editing = false
	case "esc":
		fi := m.flagInputs[m.flagCursor]
		oldVal := ""
		if m.built != nil {
			oldVal = m.built.FlagValues[fi.Flag().Name]
		}
		m.flagInputs[m.flagCursor] = fi.Blur().SetValue(oldVal)
		m.editing = false
	default:
		fi := m.flagInputs[m.flagCursor]
		m.flagInputs[m.flagCursor] = fi.HandleKey(msg)
	}
	return m, nil
}

func (m Model) handleCmd(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.built != nil && m.built.IsValid() {
			if m.executionEnabled {
				return m, executor.NewInlineExecCommand(m.root, m.built.ToArgs())
			}
			return m, func() tea.Msg { return ExitCommandMsg{Command: m.built.String()} }
		}
	case "c":
		if m.built != nil {
			_ = m.clip.Write(m.built.String())
			m.copied = true
		}
	case "l", "right":
		m.cmdScroll++
		m.clampCmdScroll()
	case "h", "left":
		if m.cmdScroll > 0 {
			m.cmdScroll--
		}
	case "0", "home":
		m.cmdScroll = 0
	case "$", "end":
		m.cmdScroll = m.cmdMaxScroll()
	}
	return m, nil
}

// cmdBarAvail returns the number of columns available for command text inside
// the command bar, after reserving space for the "> " prefix (2 columns) and
// the CommandPreview style's horizontal padding (1 column on each side).
func (m Model) cmdBarAvail() int {
	avail := m.ly.CmdBarInnerW - 2 - 2
	if avail < 1 {
		avail = 1
	}
	return avail
}

// cmdMaxScroll returns the largest valid horizontal scroll offset for the
// command bar given the current command string and panel width.
func (m Model) cmdMaxScroll() int {
	if m.built == nil {
		return 0
	}
	plain, _ := m.builtCmdParts()
	maxScroll := len([]rune(plain)) - (m.cmdBarAvail() - m.copiedBadgeWidth())
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// clampCmdScroll keeps cmdScroll within valid bounds.
func (m *Model) clampCmdScroll() {
	maxScroll := m.cmdMaxScroll()
	if m.cmdScroll > maxScroll {
		m.cmdScroll = maxScroll
	}
	if m.cmdScroll < 0 {
		m.cmdScroll = 0
	}
}

// --- State management ---

// selectNode updates all state for a newly highlighted/selected node.
func (m *Model) selectNode(node *tree.CommandNode) {
	m.descContent = m.buildDescription(node)
	if node.Runnable {
		m.built = builder.NewBuiltCommand(node)
		m.flagInputs = createFlagInputs(node)
	} else {
		m.built = nil
		m.flagInputs = nil
	}
	m.flagCursor = 0
	m.flagScroll = 0
	m.cmdScroll = 0
	m.editing = false
	m.copied = false

	if m.width > 0 {
		m.reflow()
	}
	m.recomputeZones()
	if !m.zoneActive(m.zone) {
		m.zone = ZoneTree
	}
}

// reflow recalculates layout and syncs all child sizes.
// This is the ONLY place layout and child dimensions are set.
func (m *Model) reflow() {
	// 1. Compute viewport width (horizontal layout is content-independent).
	treeOuterW := int(float64(m.width) * layout.TreeWidthPct)
	rightOuterW := m.width - treeOuterW
	rightInnerW := rightOuterW - layout.BorderSize
	if rightInnerW < 1 {
		rightInnerW = 1
	}
	// Viewport gets the full inner width minus 1 col for scrollbar.
	vpW := rightInnerW - 1
	if vpW < 1 {
		vpW = 1
	}

	// 2. Set viewport width + content so line wrapping is accurate.
	m.descVP.SetWidth(vpW)
	m.descVP.SetContent(m.descContent)

	// 3. Compute layout with actual wrapped line count.
	hints := layout.ContentHints{
		DescContentRows: m.descVP.TotalLineCount(),
		FlagContentRows: len(m.flagInputs),
	}
	m.ly = layout.Calculate(m.width, m.height, hints)

	// 4. Sync child component sizes from layout.
	// Tree content area = innerH - 1 title row. Reserve 1 col for scrollbar.
	treeContentW := m.ly.TreeInnerW - 1
	if treeContentW < 1 {
		treeContentW = 1
	}
	m.treeM = m.treeM.SetSize(treeContentW, layout.ContentRows(m.ly.TreeInnerH))
	// Viewport height = desc content rows (innerH - 1 title row).
	m.descVP.SetHeight(layout.ContentRows(m.ly.DescInnerH))
	m.descVP.GotoTop()
}

// --- Zone management ---

func (m *Model) recomputeZones() {
	m.zones = []int{ZoneTree}
	if m.descVP.TotalLineCount() > m.descVP.Height() {
		m.zones = append(m.zones, ZoneDesc)
	}
	if len(m.flagInputs) > 0 {
		m.zones = append(m.zones, ZoneFlags)
	}
	if m.built != nil {
		m.zones = append(m.zones, ZoneCommand)
	}
}

func (m *Model) cycleZone(dir int) {
	idx := m.zoneIndex()
	idx = (idx + dir + len(m.zones)) % len(m.zones)
	m.zone = m.zones[idx]
}

func (m Model) zoneIndex() int {
	for i, z := range m.zones {
		if z == m.zone {
			return i
		}
	}
	return 0
}

func (m Model) zoneActive(z int) bool {
	for _, a := range m.zones {
		if a == z {
			return true
		}
	}
	return false
}

// --- Helpers ---

func (m *Model) toggleBool(idx int) {
	fi := m.flagInputs[idx]
	if fi.Value() == "true" {
		fi = fi.SetValue("")
	} else {
		fi = fi.SetValue("true")
	}
	m.flagInputs[idx] = fi
	m.built.SetFlag(fi.Flag().Name, fi.Value())
}

func (m *Model) scrollFlags() {
	vis := layout.ContentRows(m.ly.FlagsInnerH)
	if vis < 1 {
		vis = 1
	}
	if m.flagCursor < m.flagScroll {
		m.flagScroll = m.flagCursor
	}
	if m.flagCursor >= m.flagScroll+vis {
		m.flagScroll = m.flagCursor - vis + 1
	}
}

func (m Model) buildDescription(node *tree.CommandNode) string {
	var sb strings.Builder
	if node.Short != "" {
		sb.WriteString(node.Short + "\n\n")
	}
	if node.Long != "" {
		sb.WriteString(node.Long + "\n\n")
	}
	if node.Deprecated != "" {
		sb.WriteString("DEPRECATED: " + node.Deprecated + "\n\n")
	}
	if node.Runnable {
		sb.WriteString("Usage: " + node.CommandString() + " [flags]\n\n")
	}
	if len(node.Aliases) > 0 {
		sb.WriteString("Aliases: " + strings.Join(node.Aliases, ", ") + "\n\n")
	}
	if node.Example != "" {
		sb.WriteString("Examples:\n" + node.Example + "\n\n")
	}
	if len(node.Children) > 0 {
		sb.WriteString("Subcommands:\n")
		for _, ch := range node.Children {
			sb.WriteString("  " + ch.Name)
			if ch.Short != "" {
				sb.WriteString(" - " + ch.Short)
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// --- View ---

// newAltView wraps a rendered frame in a tea.View that requests the alternate
// screen buffer (full-window mode). In bubbletea v2, AltScreen is set on the
// View returned from View() rather than via a tea.NewProgram option.
func newAltView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) View() tea.View {
	if !m.ready {
		return newAltView("Initializing...")
	}

	// Full-screen execution result view.
	if m.showExecResult {
		return newAltView(m.execResultView())
	}

	header := m.headerBar()

	treePanel := m.borderedPanel(m.zone == ZoneTree, m.ly.TreeInnerW, m.ly.TreeInnerH,
		m.theme.Subheading.Render("Commands")+"\n"+m.treeBody())

	descPanel := m.borderedPanel(m.zone == ZoneDesc, m.ly.DescInnerW, m.ly.DescInnerH,
		m.descTitle()+"\n"+m.descBody())

	flagsPanel := m.borderedPanel(m.zone == ZoneFlags, m.ly.FlagsInnerW, m.ly.FlagsInnerH,
		m.theme.Subheading.Render("Flags")+"\n"+m.flagsBody())

	right := lipgloss.JoinVertical(lipgloss.Left, descPanel, flagsPanel)
	body := lipgloss.JoinHorizontal(lipgloss.Top, treePanel, right)

	cmdBar := m.borderedPanel(m.zone == ZoneCommand, m.ly.CmdBarInnerW, m.ly.CmdBarInnerH,
		m.cmdBarContent())

	footer := m.theme.Dim.Width(m.width).Render("  " + m.footerHints())

	view := lipgloss.JoinVertical(lipgloss.Left, header, body, cmdBar, footer)
	return newAltView(m.theme.Base.Width(m.width).Height(m.height).Render(view))
}

// headerBar renders the title pill filled to the full terminal width so the
// theme background covers the whole header row.
func (m Model) headerBar() string {
	pill := m.theme.Title.Render(" " + m.binaryName + " explore ")
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, pill,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(m.theme.BackgroundColor())))
}

func (m Model) execResultView() string {
	// Header with status.
	status := m.theme.Success.Render(" Command completed successfully ")
	if m.lastExecErr != nil {
		status = m.theme.Warning.Render(" Command failed ")
	}
	header := lipgloss.PlaceHorizontal(m.width, lipgloss.Left,
		m.theme.Title.Render(" Output ")+"  "+status,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(m.theme.BackgroundColor())))

	// Scrollable output viewport.
	content := m.execVP.View()

	// Scroll indicator.
	scrollInfo := ""
	if m.execVP.TotalLineCount() > m.execVP.Height() {
		pct := 0
		if m.execVP.TotalLineCount()-m.execVP.Height() > 0 {
			pct = m.execVP.YOffset() * 100 / (m.execVP.TotalLineCount() - m.execVP.Height())
		}
		scrollInfo = m.theme.Dim.Render(fmt.Sprintf("  %d%%", pct))
	}

	footer := m.theme.Dim.Width(m.width).Render("  ↑/↓: scroll  q/Esc/Enter: back" + scrollInfo)

	view := lipgloss.JoinVertical(lipgloss.Left, header, "", content, "", footer)
	return m.theme.Base.Width(m.width).Height(m.height).Render(view)
}

// borderedPanel sizes content to exact inner dimensions, then wraps in a border.
// lipgloss.Place pads but does NOT truncate, so we truncate to exactly h lines.
func (m Model) borderedPanel(focused bool, w, h int, content string) string {
	placed := lipgloss.Place(w, h, lipgloss.Left, lipgloss.Top, content,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(m.theme.BackgroundColor())))

	// Truncate to exactly h lines (Place doesn't clip overflow).
	lines := strings.Split(placed, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	placed = strings.Join(lines, "\n")

	style := m.theme.PanelBorder
	if focused {
		style = m.theme.PanelBorderFocused
	}
	return style.Render(placed)
}

func (m Model) descTitle() string {
	// Show flag name when flags zone is focused.
	if m.zone == ZoneFlags && len(m.flagInputs) > 0 {
		f := m.flagInputs[m.flagCursor].Flag()
		title := "--" + f.Name
		if f.Shorthand != "" {
			title = "-" + f.Shorthand + ", " + title
		}
		return m.theme.Subheading.Render(title)
	}
	title := "Details"
	if c := m.treeM.Cursor(); c != nil {
		title = c.CommandString()
	}
	return m.theme.Subheading.Render(title)
}

func (m Model) descBodyContent() string {
	// When flags zone is focused, show flag details instead of command description.
	if m.zone == ZoneFlags && len(m.flagInputs) > 0 {
		return m.buildFlagDetail(m.flagInputs[m.flagCursor].Flag())
	}
	return ""
}

func (m Model) buildFlagDetail(f tree.FlagInfo) string {
	var sb strings.Builder

	if f.Usage != "" {
		sb.WriteString(f.Usage + "\n\n")
	}
	sb.WriteString("Type: " + f.Type + "\n")
	if f.DefValue != "" && f.DefValue != "0" && f.DefValue != "false" && f.DefValue != "[]" {
		sb.WriteString("Default: " + f.DefValue + "\n")
	}
	if f.Required {
		sb.WriteString("Required: yes\n")
	}
	if f.Inherited {
		sb.WriteString("Inherited: yes (from parent command)\n")
	}
	if f.Deprecated != "" {
		sb.WriteString("\nDEPRECATED: " + f.Deprecated + "\n")
	}
	if len(f.ValidValues) > 0 {
		sb.WriteString("\nValid values:\n")
		for _, v := range f.ValidValues {
			sb.WriteString("  - " + v + "\n")
		}
	}
	return sb.String()
}

func (m Model) treeBody() string {
	content := m.treeM.View()
	if m.treeM.TotalCount() > m.treeM.Height() {
		// Pad content to fill panel width minus scrollbar column.
		colW := m.ly.TreeInnerW - 1
		content = m.theme.Body.Width(colW).Render(content)
		bar := scrollbar.Render(m.treeM.Height(), m.treeM.TotalCount(), m.treeM.ScrollOffset(),
			m.theme.Scrollbar)
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, bar)
	}
	return content
}

func (m Model) descBody() string {
	// Show flag details directly when flags zone is focused.
	if flagContent := m.descBodyContent(); flagContent != "" {
		return m.theme.Body.Render(flagContent)
	}
	content := m.descVP.View()
	if m.descVP.TotalLineCount() > m.descVP.Height() {
		bar := scrollbar.Render(m.descVP.Height(), m.descVP.TotalLineCount(), m.descVP.YOffset(),
			m.theme.Scrollbar)
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, bar)
	}
	return content
}

func (m Model) flagsBody() string {
	if len(m.flagInputs) == 0 {
		if m.built == nil {
			return m.theme.Dim.Render("  Select a runnable command")
		}
		return m.theme.Dim.Render("  No flags")
	}

	vis := layout.ContentRows(m.ly.FlagsInnerH)
	if vis < 1 {
		vis = 1
	}
	end := m.flagScroll + vis
	if end > len(m.flagInputs) {
		end = len(m.flagInputs)
	}

	var sb strings.Builder
	for i := m.flagScroll; i < end; i++ {
		fi := m.flagInputs[i]
		cur := i == m.flagCursor && m.zone == ZoneFlags
		edit := m.editing && cur
		sb.WriteString("  " + fi.Render(cur, edit, m.ly.FlagsInnerW-3) + "\n")
	}

	content := sb.String()
	if len(m.flagInputs) > vis {
		// Pad content to fill panel width minus scrollbar column so bar is at far right.
		colW := m.ly.FlagsInnerW - 1
		content = m.theme.Body.Width(colW).Render(content)
		bar := scrollbar.Render(vis, len(m.flagInputs), m.flagScroll,
			m.theme.Scrollbar)
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, bar)
	}
	return content
}

func (m Model) cmdBarContent() string {
	if m.built == nil {
		return m.theme.Dim.Render("  No command selected")
	}

	plain, styled := m.builtCmdParts()

	// Reserve space at the right edge for the copied confirmation so it stays
	// visible regardless of the horizontal scroll position.
	badge := ""
	if m.zone == ZoneCommand && m.copied {
		badge = "  " + m.theme.Success.Render("✓ Copied")
	}

	// Columns available for command text after the "> " prefix, the preview
	// style's horizontal padding, and the (optional) copied badge.
	avail := m.cmdBarAvail() - m.copiedBadgeWidth()
	if avail < 1 {
		avail = 1
	}

	var body string
	if len([]rune(plain)) <= avail {
		// Everything fits: render the fully styled version unchanged.
		body = styled
	} else {
		// Overflow: horizontally scroll the plain text so the whole command is
		// reachable via ←/→. Styling is dropped while scrolling since the
		// visible window is a substring.
		body = windowText(plain, m.cmdScroll, avail)
	}
	return m.theme.CommandPreview.Render("> " + body + badge)
}

// copiedBadgeWidth returns the number of columns reserved on the right of the
// command bar for the copied confirmation (0 when it is not showing).
func (m Model) copiedBadgeWidth() int {
	if m.zone == ZoneCommand && m.copied {
		return 2 + lipgloss.Width("✓ Copied")
	}
	return 0
}

// builtCmdParts returns the command preview both as plain text (for width
// measurement and scrolling) and with styled annotations (for display).
func (m Model) builtCmdParts() (plain, styled string) {
	cmd := m.built.String()
	plain, styled = cmd, cmd
	if missing := m.built.UnsetRequiredFlags(); len(missing) > 0 {
		s := "[missing: " + strings.Join(missing, ", ") + "]"
		plain += "  " + s
		styled += "  " + m.theme.Warning.Render(s)
	}
	return plain, styled
}

// windowText returns a horizontally scrolled slice of s that fits within width
// columns, adding "<"/">" overflow markers when content extends beyond the
// visible window. scroll is the number of columns hidden on the left.
func windowText(s string, scroll, width int) string {
	runes := []rune(s)
	n := len(runes)
	if n <= width {
		return s
	}
	if width < 3 {
		width = 3
	}

	maxScroll := n - width
	if scroll < 0 {
		scroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	start := scroll
	end := start + width
	leftMark := start > 0
	rightMark := end < n

	// Reserve a column for each marker so the total width stays constant.
	if leftMark {
		start++
	}
	if rightMark {
		end--
	}
	if start > end {
		start = end
	}

	var b strings.Builder
	if leftMark {
		b.WriteString("<")
	}
	b.WriteString(string(runes[start:end]))
	if rightMark {
		b.WriteString(">")
	}
	return b.String()
}

func (m Model) footerHints() string {
	if m.zone == ZoneFlags && m.editing {
		return "Enter/Tab: confirm  │  Esc: cancel"
	}
	common := "Tab: next zone"
	switch m.zone {
	case ZoneTree:
		return common + "  │  ↑/↓: navigate  ←/→: collapse/expand  │  q: quit"
	case ZoneDesc:
		return common + "  │  ↑/↓: scroll  │  q: quit"
	case ZoneFlags:
		return common + "  │  ↑/↓: navigate  Enter: edit  Space: toggle bool"
	case ZoneCommand:
		if m.executionEnabled {
			return common + "  │  Enter: run  c: copy  ←/→: scroll  │  q: quit"
		}
		return common + "  │  Enter: paste to terminal  c: copy  ←/→: scroll  │  q: quit"
	}
	return common + "  │  q: quit"
}

// ExitCommand returns the command string from exit, or empty.
func (m Model) ExitCommand() string {
	return m.exitCommand
}

func createFlagInputs(node *tree.CommandNode) []flaginput.FlagInput {
	flags := node.AllFlags()
	inputs := make([]flaginput.FlagInput, 0, len(flags))
	for _, f := range flags {
		inputs = append(inputs, flaginput.New(f))
	}
	return inputs
}
