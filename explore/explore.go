package explore

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/oakwood-commons/cobra-explorer/internal/model"
	"github.com/oakwood-commons/cobra-explorer/internal/theme"
)

// Run launches the interactive TUI explorer for the given cobra root command.
// When the user selects a command and presses Enter, the built command string
// is printed to stdout and the function returns nil. The calling shell wrapper
// can capture this output and place it in the readline buffer.
func Run(root *cobra.Command, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	th, themeSet := theme.Default(), false
	if cfg.themeName != "" {
		if t, ok := theme.Get(cfg.themeName); ok {
			th, themeSet = t, true
		}
	}

	m := model.New(root, model.Options{
		BinaryName:       cfg.binaryName,
		Theme:            th,
		ThemeSet:         themeSet,
		ShowHidden:       cfg.showHidden,
		ExecutionEnabled: cfg.executionEnabled,
	})

	// AltScreen is requested declaratively by the model's View(); NewProgram no
	// longer takes a WithAltScreen option in bubbletea v2.
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if user exited with a command to emit
	if msg, ok := finalModel.(exitCommander); ok {
		if cmd := msg.ExitCommand(); cmd != "" {
			fmt.Fprintln(os.Stdout, cmd)
		}
	}

	return nil
}

// exitCommander is implemented by the model to retrieve the exit command.
type exitCommander interface {
	ExitCommand() string
}

// NewCommand creates a *cobra.Command that launches the explorer when run.
// Embed this as a subcommand of your CLI's root command:
//
//	rootCmd.AddCommand(explore.NewCommand(rootCmd))
//
// The command exposes a --theme flag so end users can pick a theme preset at
// runtime. Any theme the developer configured via WithThemeName acts as the
// default; when the user passes --theme, their choice overrides it.
func NewCommand(root *cobra.Command, opts ...Option) *cobra.Command {
	var (
		themeName  string
		listThemes bool
	)

	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Interactively explore available commands",
		Long:  "Launch an interactive TUI to browse commands, view documentation, and build commands.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if listThemes {
				fmt.Fprintln(cmd.OutOrStdout(), "Available themes:")
				for _, name := range theme.Names() {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", name)
				}
				return nil
			}
			runOpts := opts
			if themeName != "" {
				if _, ok := theme.Get(themeName); !ok {
					return fmt.Errorf("cobra-explorer: unknown theme %q (available: %s)",
						themeName, strings.Join(theme.Names(), ", "))
				}
				// The user's choice is appended last so it overrides any theme
				// the developer configured via WithThemeName.
				runOpts = append(runOpts, WithThemeName(themeName))
			}
			return Run(root, runOpts...)
		},
	}

	cmd.Flags().StringVar(&themeName, "theme", "",
		"color theme to use (see --list-themes for available names)")
	cmd.Flags().BoolVar(&listThemes, "list-themes", false,
		"list the available color themes and exit")
	_ = cmd.RegisterFlagCompletionFunc("theme",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return theme.Names(), cobra.ShellCompDirectiveNoFileComp
		})

	return cmd
}
