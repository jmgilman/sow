package project

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the project command with wizard as primary interface.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create or continue a project (interactive)",
		Long: `Interactive wizard for creating or continuing projects.

The wizard guides you through:
  - Creating new projects from GitHub issues or branch names
  - Continuing existing projects
  - Selecting project types and providing descriptions

Examples:
  sow project                    # Launch interactive wizard
  sow project -- --model opus    # Launch wizard, pass flags to Claude`,
		RunE: runWizard,
		Args: cobra.NoArgs,
	}

	// Keep set and delete subcommands
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
