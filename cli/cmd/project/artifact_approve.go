package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
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
func newArtifactApproveCmd(accessor SowFSAccessor) *cobra.Command {
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

			// Validate phase name
			if err := project.ValidatePhase(phaseName); err != nil {
				return fmt.Errorf("phase validation failed: %w", err)
			}

			// Artifact can only exist in discovery or design
			if phaseName != project.PhaseDiscovery && phaseName != project.PhaseDesign {
				return fmt.Errorf("artifacts can only exist in discovery or design phases, got: %s", phaseName)
			}

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Get project filesystem
			projectFS, err := sowFS.Project()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first: %w", err)
			}

			// Read current state
			state, err := projectFS.State()
			if err != nil {
				return fmt.Errorf("failed to read project state: %w", err)
			}

			// Approve the artifact
			if err := project.ApproveArtifact(state, phaseName, artifactPath); err != nil {
				return fmt.Errorf("failed to approve artifact: %w", err)
			}

			// Write updated state
			if err := projectFS.WriteState(state); err != nil {
				return fmt.Errorf("failed to write project state: %w", err)
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
