package agent

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewAdvanceCmd creates the command to advance to the next state within the current phase.
//
// Usage:
//
//	sow agent advance
//
// This command advances to the next state within the current phase (intra-phase transition).
// It is used by project types that have multiple states within a single phase.
func NewAdvanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance",
		Short: "Advance to next state within current phase",
		Long: `Advance to the next state within the current phase (intra-phase transition).

This command is used by project types that have multiple states within a single phase.
For example, an exploration phase might have "active" and "summarizing" states.

Standard project phases don't have internal states, so this command will return an error
when called on a standard project. Use 'sow agent complete' to move between phases instead.

The command will fail if:
  - No project exists
  - No phase is currently active
  - The active phase does not support state advancement (e.g., standard project phases)

Example:
  # Advance within the current phase (if supported)
  sow agent advance`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
				return fmt.Errorf("no active phase found - project may be complete")
			}

			// Advance the phase
			err = phase.Advance()
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("phase %s does not support state advancement", phase.Name())
			}
			if err != nil {
				return fmt.Errorf("failed to advance: %w", err)
			}

			cmd.Println("✓ Advanced to next state")
			return nil
		},
	}

	return cmd
}
