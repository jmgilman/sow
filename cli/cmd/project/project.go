// Package project provides commands for managing project lifecycle.
package project

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the root project command.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project lifecycle",
		Long: `Manage project lifecycle and state.

The project commands provide access to project initialization, status checking,
and cleanup operations. These commands are primarily used by the orchestrator
agent but can be invoked manually for debugging or intervention.

All projects follow the 5-phase model:
  1. Discovery (optional, human-led)
  2. Design (optional, human-led)
  3. Implementation (required, AI-autonomous)
  4. Review (required, AI-autonomous)
  5. Finalize (required, AI-autonomous)`,
	}

	// Add subcommands
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(newPhaseCmd())
	cmd.AddCommand(newArtifactCmd())
	cmd.AddCommand(newReviewCmd())
	cmd.AddCommand(newFinalizeCmd())

	return cmd
}
