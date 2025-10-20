package project

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"fmt"

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
func newArtifactAddCmd() *cobra.Command {
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

			// Artifact can only be added to discovery or design
			if phaseName != "discovery" && phaseName != "design" {
				return fmt.Errorf("artifacts can only be added to discovery or design phases, got: %s", phaseName)
			}

			// Get Sow from context
			s := cmdutil.SowFromContext(cmd.Context())

			// Get project
			project, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Add artifact (handles validation, auto-save)
			if err := project.AddArtifact(phaseName, artifactPath, approved); err != nil {
				return fmt.Errorf("failed to add artifact: %w", err)
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
