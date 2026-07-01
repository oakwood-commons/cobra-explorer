// Package explore provides an interactive TUI for Cobra-based CLI applications.
// It lets users visually navigate the full command tree, inspect flags and
// documentation, build commands interactively, and execute them.
//
// Integration is a single function call:
//
//	rootCmd.AddCommand(explore.NewCommand(rootCmd))
//
// Or launch directly:
//
//	explore.Run(rootCmd)
package explore
