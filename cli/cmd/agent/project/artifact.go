package project

import (
	"github.com/spf13/cobra"
)

// newArtifactCmd creates the artifact command for managing phase artifacts.
//
// Usage:
//   sow project artifact <subcommand>
//
// Subcommands:
//   - add: Add an artifact to a phase
//   - approve: Approve an existing artifact
//   - list: List artifacts for a phase
func newArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage phase artifacts",
		Long: `Manage artifacts in discovery and design phases.

Artifacts are outputs from discovery and design phases that require human approval
before the phase can be completed. Examples include:
  - Research reports from discovery
  - Architecture Decision Records (ADRs) from design
  - Design documents and specifications from design

All artifacts must be approved before their phase can be marked complete.`,
	}

	// Add subcommands
	cmd.AddCommand(newArtifactAddCmd())
	cmd.AddCommand(newArtifactApproveCmd())
	cmd.AddCommand(newArtifactListCmd())

	return cmd
}
