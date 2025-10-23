package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Determine active phase
			state := project.State()
			activePhase, phaseStatus := projectpkg.DetermineActivePhase(state)

			if activePhase == "unknown" {
				return fmt.Errorf("no active phase found - project may be complete")
			}

			// Validate we're in an active state (not a decision state)
			// Only optional phases (discovery, design) have decision states
			if phaseStatus == "pending" && (activePhase == "discovery" || activePhase == "design") {
				return fmt.Errorf("phase %s is in decision state - enable or skip it first", activePhase)
			}

			// Complete the phase
			if err := project.CompletePhase(activePhase); err != nil {
				return fmt.Errorf("failed to complete %s phase: %w", activePhase, err)
			}

			cmd.Printf("\n✓ Completed %s phase\n", activePhase)
			return nil
		},
	}

	return cmd
}
