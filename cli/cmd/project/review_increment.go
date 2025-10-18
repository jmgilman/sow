package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newReviewIncrementCmd creates the command to increment review iteration.
//
// Usage:
//   sow project review increment
//
// This command increments the review iteration counter when looping back
// from review to implementation. The counter tracks how many review cycles
// have occurred.
func newReviewIncrementCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increment",
		Short: "Increment review iteration counter",
		Long: `Increment the review iteration counter.

This command is used when looping back from review to implementation to address
issues found during review. Each increment represents one review cycle.

The iteration counter starts at 1 and increments each time the project returns
to implementation phase from review.

Example:
  # Increment review iteration before loop-back
  sow project review increment`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			oldIteration := state.Phases.Review.Iteration

			// Increment review iteration
			if err := project.IncrementReviewIteration(state); err != nil {
				return fmt.Errorf("failed to increment review iteration: %w", err)
			}

			// Write updated state
			if err := projectFS.WriteState(state); err != nil {
				return fmt.Errorf("failed to write project state: %w", err)
			}

			cmd.Printf("✓ Incremented review iteration: %d → %d\n", oldIteration, state.Phases.Review.Iteration)

			return nil
		},
	}

	return cmd
}
