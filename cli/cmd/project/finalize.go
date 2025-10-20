package project

import (
	"github.com/spf13/cobra"
)

// newFinalizeCmd creates the finalize command for managing the finalize phase.
//
// Usage:
//   sow project finalize <subcommand>
//
// Subcommands:
//   - complete: Complete a finalize subphase (documentation or checks)
//   - add-document: Track a documentation file update
//   - move-artifact: Record an artifact moved to knowledge
func newFinalizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize",
		Short: "Manage finalize phase",
		Long: `Manage the finalize phase.

The finalize phase prepares work for merge through:
  - Documentation updates (tracked via add-document)
  - Moving design artifacts to knowledge directory (tracked via move-artifact)
  - Final validation checks
  - Project deletion (mandatory before completion)
  - Pull request creation

These commands track changes made during finalization for audit purposes.`,
	}

	// Add subcommands
	cmd.AddCommand(newFinalizeCompleteCmd())
	cmd.AddCommand(newFinalizeAddDocumentCmd())
	cmd.AddCommand(newFinalizeMoveArtifactCmd())

	return cmd
}
