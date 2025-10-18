package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newArtifactAddCmd creates the command to add an artifact to a phase.
//
// Usage:
//   sow project artifact add <path> --phase <phase> [--approved]
//
// Arguments:
//   <path>: Path to artifact relative to .sow/project/
//
// Flags:
//   --phase: Phase name (discovery or design, required)
//   --approved: Mark artifact as approved immediately (optional)
func newArtifactAddCmd(accessor SowFSAccessor) *cobra.Command {
	var phaseName string
	var approved bool

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add an artifact to a phase",
		Long: `Add an artifact to the discovery or design phase.

Artifacts are tracked in project state and must be approved before the phase
can be completed. By default, artifacts are added with approved=false.

Examples:
  # Add research report to discovery phase
  sow project artifact add phases/discovery/research/001-jwt-libraries.md --phase discovery

  # Add ADR to design phase (pre-approved)
  sow project artifact add phases/design/adrs/001-use-jwt-rs256.md --phase design --approved`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			artifactPath := args[0]

			// Validate phase name
			if err := project.ValidatePhase(phaseName); err != nil {
				return fmt.Errorf("phase validation failed: %w", err)
			}

			// Artifact can only be added to discovery or design
			if phaseName != project.PhaseDiscovery && phaseName != project.PhaseDesign {
				return fmt.Errorf("artifacts can only be added to discovery or design phases, got: %s", phaseName)
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

			// Add the artifact
			if err := project.AddArtifact(state, phaseName, artifactPath, approved); err != nil {
				return fmt.Errorf("failed to add artifact: %w", err)
			}

			// Write updated state
			if err := projectFS.WriteState(state); err != nil {
				return fmt.Errorf("failed to write project state: %w", err)
			}

			approvalStatus := "pending approval"
			if approved {
				approvalStatus = "approved"
			}

			cmd.Printf("✓ Added artifact to %s phase (%s)\n", phaseName, approvalStatus)
			cmd.Printf("  %s\n", artifactPath)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&phaseName, "phase", "", "Phase name (discovery or design, required)")
	_ = cmd.MarkFlagRequired("phase")
	cmd.Flags().BoolVar(&approved, "approved", false, "Mark artifact as approved immediately")

	return cmd
}
