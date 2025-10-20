package project

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newArtifactApproveCmd creates the command to approve an artifact.
//
// Usage:
//   sow project artifact approve <path> --phase <phase>
//
// Arguments:
//   <path>: Path to artifact to approve
//
// Flags:
//   --phase: Phase name (discovery or design, required)
func newArtifactApproveCmd() *cobra.Command {
	var phaseName string

	cmd := &cobra.Command{
		Use:   "approve <path>",
		Short: "Approve an existing artifact",
		Long: `Approve an artifact in the discovery or design phase.

Artifacts must be approved before their phase can be marked complete.
This command marks an existing artifact as approved.

Examples:
  # Approve research report
  sow project artifact approve phases/discovery/research/001-jwt-libraries.md --phase discovery

  # Approve ADR
  sow project artifact approve phases/design/adrs/001-use-jwt-rs256.md --phase design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			artifactPath := args[0]

			// Artifact can only exist in discovery or design
			if phaseName != "discovery" && phaseName != "design" {
				return fmt.Errorf("artifacts can only exist in discovery or design phases, got: %s", phaseName)
			}

			// Get Sow from context
			s := sowFromContext(cmd.Context())

			// Get project
			project, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Approve artifact (handles validation, auto-save)
			if err := project.ApproveArtifact(phaseName, artifactPath); err != nil {
				return err
			}

			cmd.Printf("✓ Approved artifact in %s phase\n", phaseName)
			cmd.Printf("  %s\n", artifactPath)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&phaseName, "phase", "", "Phase name (discovery or design, required)")
	_ = cmd.MarkFlagRequired("phase")

	return cmd
}
