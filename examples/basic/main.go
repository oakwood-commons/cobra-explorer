package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/oakwood-commons/cobra-explorer/explore"
)

func main() {
	root := buildDemoCLI()

	// The theme defaults to "dark". End users can override it at runtime with
	// the explore command's --theme flag (e.g. `demo explore --theme dracula`)
	// or list the available themes with `demo explore --list-themes`.
	root.AddCommand(explore.NewCommand(root, explore.WithExecution(true)))

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildDemoCLI() *cobra.Command {
	root := &cobra.Command{
		Use:   "demo",
		Short: "A demo CLI to showcase cobra-explorer",
	}

	// run command with subcommands
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run various tasks",
	}

	runCmd.AddCommand(&cobra.Command{
		Use:     "server",
		Short:   "Start the development server",
		Long:    "Start the development server with hot-reload support. Serves on the configured port.",
		Example: "  demo run server --port 8080\n  demo run server --host 0.0.0.0",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			host, _ := cmd.Flags().GetString("host")
			fmt.Printf("Starting server on %s:%d\n", host, port)
			return nil
		},
	})
	runCmd.Commands()[0].Flags().IntP("port", "p", 8080, "Port to listen on")
	runCmd.Commands()[0].Flags().String("host", "localhost", "Host to bind to")
	runCmd.Commands()[0].Flags().Bool("tls", false, "Enable TLS")

	runCmd.AddCommand(&cobra.Command{
		Use:   "worker",
		Short: "Start a background worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			fmt.Printf("Starting worker with concurrency=%d\n", concurrency)
			return nil
		},
	})
	runCmd.Commands()[1].Flags().Int("concurrency", 4, "Number of concurrent workers")
	runCmd.Commands()[1].Flags().StringSlice("queues", []string{"default"}, "Queues to process")

	root.AddCommand(runCmd)

	// config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}
	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("config: default settings")
			return nil
		},
	})
	configCmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			fmt.Printf("Set %s = %s\n", args[0], args[1])
			return nil
		},
	})
	root.AddCommand(configCmd)

	// version command
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("demo v0.1.0")
			return nil
		},
	})

	// Global flags
	root.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	root.PersistentFlags().String("config", "", "Config file path")

	return root
}
