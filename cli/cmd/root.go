package cmd

import (
	"fmt"
	"os"

	"github.com/jmgilman/go/fs/billy"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sow",
		Short: "AI-powered system of work",
		Long: `sow - Structured software development with AI agents

sow is a CLI tool and framework for managing AI-assisted software projects.
It provides structure, state management, and context compilation for
orchestrating multiple AI agents across a 5-phase development workflow.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize adapters
			fs := billy.NewLocal()

			// Add adapters to context
			ctx := WithFilesystem(cmd.Context(), fs)
			cmd.SetContext(ctx)

			return nil
		},
	}

	// Global flags
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	cmd.PersistentFlags().Bool("quiet", false, "Suppress non-error output")

	// Add subcommands
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewValidateCmd())
	cmd.AddCommand(NewSchemaCmd())
	cmd.AddCommand(NewLogCmd())
	cmd.AddCommand(NewSessionInfoCmd())
	cmd.AddCommand(NewRefsCmd())
	cmd.AddCommand(NewCacheCmd())

	return cmd
}

// Execute runs the root command
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
