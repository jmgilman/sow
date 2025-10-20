// Package cmd provides the CLI commands for the sow tool.
package cmd

import (
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/sow/cli/cmd/project"
	"github.com/jmgilman/sow/cli/cmd/refs"
	"github.com/jmgilman/sow/cli/cmd/task"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
)

// NewRootCmd creates the root command.
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
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Get current working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			// Create raw billy filesystem for Sow instance
			rawBillyFS := osfs.New(cwd)

			// Create unified Sow instance
			sowInstance := sow.New(rawBillyFS)

			// Create wrapped filesystem for backwards compatibility
			baseFS := billy.NewLocal()
			wrappedFS, err := baseFS.Chroot(cwd)
			if err != nil {
				return fmt.Errorf("failed to chroot filesystem: %w", err)
			}

			// Add to context
			ctx := WithFilesystem(cmd.Context(), wrappedFS)
			ctx = WithSow(ctx, sowInstance)

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
	cmd.AddCommand(NewLogCmd())
	cmd.AddCommand(NewSessionInfoCmd())
	cmd.AddCommand(refs.NewRefsCmd())
	cmd.AddCommand(project.NewProjectCmd())
	cmd.AddCommand(task.NewTaskCmd())

	return cmd
}

// Execute runs the root command.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
