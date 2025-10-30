package agent

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewCompleteCmd creates the command to complete the active phase.
//
// Usage:
//
//	sow agent complete
//
// This command automatically detects the current active phase and completes it.
// No need to specify which phase - it's implicit based on the project state.
func NewCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete",
		Short: "Complete the active phase",
		Long: `Complete the currently active phase.

This command automatically detects which phase is currently active and marks it as complete.
You don't need to specify the phase name - it's determined from the project state.

The command will fail if:
  - No project exists
  - No phase is currently active
  - The active phase cannot be completed (e.g., missing required artifacts or tasks)

Example:
  # Complete whichever phase is currently active
  sow agent complete`,
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

			// Complete the phase via Phase interface
			result, err := phase.Complete()
			if err != nil {
				return fmt.Errorf("failed to complete phase: %w", err)
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

			cmd.Printf("\n✓ Completed %s phase\n", phase.Name())
			return nil
		},
	}

	return cmd
}
