package agent

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
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

The active phase must support artifacts (discovery, design, and review phases do).
The artifact must already exist in the phase.

Example:
  sow agent artifact approve project/phases/discovery/research/001-jwt.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				return fmt.Errorf("no active phase found")
			}

			// Approve artifact via Phase interface
			result, err := phase.ApproveArtifact(path)
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("phase %s does not support artifacts", phase.Name())
			}
			if err != nil {
				return fmt.Errorf("failed to approve artifact: %w", err)
			}

			// Fire event if phase returned one
			if result.Event != "" {
				machine := proj.Machine()
				if err := machine.Fire(result.Event); err != nil {
					return fmt.Errorf("failed to fire event %s: %w", result.Event, err)
				}
				// Save after transition
				if err := proj.Save(); err != nil {
					return fmt.Errorf("failed to save project state: %w", err)
				}
			}

			cmd.Printf("\n✓ Approved artifact: %s\n", path)

			return nil
		},
	}

	return cmd
}
