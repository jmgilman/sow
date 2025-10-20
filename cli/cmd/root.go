// Package cmd provides the CLI commands for the sow tool.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/sow/cli/cmd/project"
	"github.com/jmgilman/sow/cli/cmd/refs"
	"github.com/jmgilman/sow/cli/cmd/task"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
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

			// Find repository root (walk up to find .git)
			repoRoot := findRepoRoot(cwd)
			if repoRoot == "" {
				repoRoot = cwd // Fallback to cwd if not in a git repo
			}

			// Create raw billy filesystem rooted at repo root
			rawBillyFS := osfs.New(repoRoot)

			// Create unified Sow instance
			sowInstance := sow.New(rawBillyFS)

			// Create wrapped filesystem for backwards compatibility
			baseFS := billy.NewLocal()
			wrappedFS, err := baseFS.Chroot(repoRoot)
			if err != nil {
				return fmt.Errorf("failed to chroot filesystem: %w", err)
			}

			// Add to context
			ctx := cmdutil.WithFilesystem(cmd.Context(), wrappedFS)
			ctx = cmdutil.WithSow(ctx, sowInstance)

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
	cmd.AddCommand(NewGreetCmd())
	cmd.AddCommand(NewStartCmd())
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

// findRepoRoot walks up the directory tree to find the git repository root.
// Returns the directory containing .git, or empty string if not found.
func findRepoRoot(start string) string {
	dir := start
	for {
		// Check if .git exists in this directory
		gitPath := dir + string(os.PathSeparator) + ".git"
		if _, err := os.Stat(gitPath); err == nil {
			return dir
		}

		// Move up one directory
		parent := dir + string(os.PathSeparator) + ".."
		absParent, err := os.Stat(parent)
		if err != nil {
			return "" // Can't stat parent, give up
		}

		// Get absolute path of parent
		absPath, err := filepath.Abs(parent)
		if err != nil {
			return ""
		}

		// Check if we've reached the root
		if absPath == dir {
			return "" // Reached filesystem root without finding .git
		}

		// Check if parent is the same as current (another way to detect root)
		currentStat, _ := os.Stat(dir)
		if os.SameFile(currentStat, absParent) {
			return ""
		}

		dir = absPath
	}
}
