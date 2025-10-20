package project

import (
	"fmt"

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
func newReviewIncrementCmd() *cobra.Command {
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
			// Get Sow from context
			s := sowFromContext(cmd.Context())

			// Get project
			proj, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Get current iteration before incrementing
			oldIteration := proj.State().Phases.Review.Iteration

			// Increment review iteration (auto-saves)
			if err := proj.IncrementReviewIteration(); err != nil {
				return err
			}

			// Get new iteration after increment
			newIteration := proj.State().Phases.Review.Iteration

			cmd.Printf("✓ Incremented review iteration: %d → %d\n", oldIteration, newIteration)

			return nil
		},
	}

	return cmd
}
