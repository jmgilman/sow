package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/config"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sow",
		Short: "AI-powered system of work",
		Long: `sow - Structured software development with AI agents

sow is a command-line tool that manages .sow/ directory structures,
validates project and task states against embedded CUE schemas, and
provides utilities for AI agents to work with structured projects.`,
		Version: config.Version,
		SilenceUsage: true,
	}

	// Global flags
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress output")
	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colors")

	// Add subcommands
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}

// NewVersionCmd creates the version command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "sow %s\n", config.Version)
			if config.BuildDate != "unknown" {
				fmt.Fprintf(cmd.OutOrStdout(), "Built: %s\n", config.BuildDate)
			}
			if config.Commit != "none" {
				fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", config.Commit)
			}
		},
	}
}
