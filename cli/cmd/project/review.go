package project

import (
	"github.com/spf13/cobra"
)

// newReviewCmd creates the review command for managing the review phase.
//
// Usage:
//   sow project review <subcommand>
//
// Subcommands:
//   - increment: Increment review iteration counter
//   - add-report: Add a review report with assessment
func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Manage review phase",
		Long: `Manage the review phase.

The review phase validates that implementation meets expected outcomes.
When issues are found, the review iteration counter is incremented and
the project loops back to implementation. Each iteration produces a
numbered review report (001, 002, 003...) with a pass/fail assessment.

Review loop-back workflow:
  1. Review finds issues → report created with assessment=fail
  2. Increment iteration counter → sow project review increment
  3. Loop back to implementation to fix issues
  4. Automatic return to review
  5. Create new review report → sow project review add-report
  6. If assessment=pass → proceed to finalize`,
	}

	// Add subcommands
	cmd.AddCommand(newReviewIncrementCmd())
	cmd.AddCommand(newReviewAddReportCmd())

	return cmd
}
