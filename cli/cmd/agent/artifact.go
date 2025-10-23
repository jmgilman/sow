package agent

import "github.com/spf13/cobra"

// NewArtifactCmd creates the parent artifact command.
//
// Usage:
//
//	sow agent artifact <subcommand>
//
// Subcommands operate on the currently active phase.
func NewArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage artifacts for the active phase",
		Long: `Manage artifacts for the currently active phase.

Artifacts track important files created during discovery and design phases,
such as research documents, architecture diagrams, and design notes.

All artifact commands operate on the currently active phase - no need to specify --phase.

Available subcommands:
  add       Add an artifact to the active phase
  approve   Approve an artifact
  list      List all artifacts`,
	}

	// Add subcommands
	cmd.AddCommand(NewArtifactAddCmd())
	cmd.AddCommand(NewArtifactApproveCmd())
	cmd.AddCommand(NewArtifactListCmd())

	return cmd
}
