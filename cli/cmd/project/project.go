package project

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the project command with wizard as primary interface.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [-- <claude-flags>...]",
		Short: "Create or continue a project (interactive)",
		Long: `Interactive wizard for creating or continuing projects.

The wizard guides you through:
  - Creating new projects from GitHub issues or branch names
  - Continuing existing projects
  - Selecting project types and providing descriptions

Claude Code Flags:
  Pass additional flags to Claude Code using -- separator.
  All flags after -- are forwarded directly to the claude CLI.

Examples:
  sow project                                    # Launch interactive wizard
  sow project -- --model opus                    # Use specific model
  sow project -- --verbose                       # Enable verbose output
  sow project -- --model opus --verbose          # Multiple flags
  sow project -- --dangerously-skip-permissions  # Advanced flags`,
		RunE: runWizard,
		// No Args validator - allows pass-through flags after --
	}

	// Keep set and delete subcommands
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
