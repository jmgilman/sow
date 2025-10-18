// Package project provides commands for managing project lifecycle.
package project

import (
	"context"

	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/spf13/cobra"
)

// SowFSAccessor is a function type that retrieves SowFS from context.
// This allows commands to be tested with different SowFS implementations.
type SowFSAccessor func(ctx context.Context) sowfs.SowFS

// NewProjectCmd creates the root project command.
func NewProjectCmd(accessor SowFSAccessor) *cobra.Command {
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
	cmd.AddCommand(NewInitCmd(accessor))
	cmd.AddCommand(NewStatusCmd(accessor))
	cmd.AddCommand(NewDeleteCmd(accessor))
	cmd.AddCommand(newPhaseCmd(accessor))
	cmd.AddCommand(newArtifactCmd(accessor))
	cmd.AddCommand(newReviewCmd(accessor))
	cmd.AddCommand(newFinalizeCmd(accessor))

	return cmd
}
