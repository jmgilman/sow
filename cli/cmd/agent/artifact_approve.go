package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewArtifactApproveCmd creates the command to approve an artifact on the active phase.
//
// Usage:
//
//	sow agent artifact approve <path>
//
// Approves an artifact on the currently active phase.
func NewArtifactApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <path>",
		Short: "Approve an artifact on the active phase",
		Long: `Approve an artifact on the currently active phase.

The active phase must support artifacts (discovery and design phases do).
The artifact must already exist in the phase.

Example:
  sow agent artifact approve project/phases/discovery/research/001-jwt.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent init' first")
			}

			// Determine active phase
			state := project.State()
			activePhase, phaseStatus := projectpkg.DetermineActivePhase(state)

			if activePhase == "unknown" {
				return fmt.Errorf("no active phase found")
			}

			// Get phase metadata to validate artifacts are supported
			metadata, err := projectpkg.GetPhaseMetadata(activePhase)
			if err != nil {
				return fmt.Errorf("failed to get phase metadata: %w", err)
			}

			if !metadata.SupportsArtifacts {
				return fmt.Errorf("phase %s does not support artifacts", activePhase)
			}

			// Validate we're in an active state (not a decision state)
			if phaseStatus == "pending" {
				return fmt.Errorf("phase %s is in decision state - enable it first", activePhase)
			}

			// Approve artifact
			if err := project.ApproveArtifact(activePhase, path); err != nil {
				return fmt.Errorf("failed to approve artifact: %w", err)
			}

			cmd.Printf("\n✓ Approved artifact: %s\n", path)

			return nil
		},
	}

	return cmd
}
