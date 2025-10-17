// Package cmd provides the CLI commands for the sow tool.
package cmd

import (
	"fmt"
	"os"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/sow/cli/cmd/project"
	"github.com/jmgilman/sow/cli/cmd/refs"
	"github.com/jmgilman/sow/cli/cmd/task"
	"github.com/jmgilman/sow/cli/internal/sowfs"
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

			// Create filesystem rooted at current working directory
			baseFS := billy.NewLocal()
			fs, err := baseFS.Chroot(cwd)
			if err != nil {
				return fmt.Errorf("failed to chroot filesystem: %w", err)
			}

			ctx := WithFilesystem(cmd.Context(), fs)

			// Try to initialize SowFS (optional - will be nil if not in .sow directory)
			// This allows commands to optionally use SowFS when available
			sfs, err := sowfs.NewSowFS()
			if err == nil {
				ctx = WithSowFS(ctx, sfs)
			}
			// Silently ignore errors - not all commands require SowFS

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
	cmd.AddCommand(refs.NewRefsCmd(SowFSFromContext))
	cmd.AddCommand(project.NewProjectCmd(SowFSFromContext))
	cmd.AddCommand(task.NewTaskCmd(SowFSFromContext))

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
