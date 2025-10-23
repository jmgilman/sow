package agent

import (
	"github.com/spf13/cobra"
)

// NewFinalizeCmd creates the finalize command for managing the finalize phase.
//
// Usage:
//   sow agent finalize <subcommand>
//
// Subcommands:
//   - complete: Complete a finalize subphase (documentation or checks)
//   - doc: Track a documentation file update
//   - move: Record an artifact moved to knowledge
func NewFinalizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize",
		Short: "Manage finalize phase",
		Long: `Manage the finalize phase.

The finalize phase prepares work for merge through:
  - Documentation updates (tracked via doc)
  - Moving design artifacts to knowledge directory (tracked via move)
  - Final validation checks
  - Project deletion (mandatory before completion)
  - Pull request creation

These commands track changes made during finalization for audit purposes.

Available subcommands:
  complete    Complete a finalize subphase
  doc         Track a documentation update
  move        Record an artifact move`,
	}

	// Add subcommands
	cmd.AddCommand(NewFinalizeCompleteCmd())
	cmd.AddCommand(NewFinalizeDocCmd())
	cmd.AddCommand(NewFinalizeMoveCmd())

	return cmd
}
