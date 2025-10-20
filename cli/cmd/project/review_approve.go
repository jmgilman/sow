package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// newReviewApproveCmd creates the command to approve a review report.
//
// Usage:
//   sow project review approve <report-id>
//
// Approves the orchestrator's review assessment and triggers the state transition
// based on the assessment (pass → finalize, fail → back to implementation).
func newReviewApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <report-id>",
		Short: "Approve a review report",
		Long: `Approve the orchestrator's review assessment.

After the orchestrator creates a review report with an assessment (pass/fail),
human approval is required before the state transition occurs.

Transitions based on assessment:
  - pass: Proceeds to finalize phase
  - fail: Loops back to implementation planning (with incremented iteration)

Example:
  # Approve the latest review report
  sow project review approve 001

  # Approve a specific review report
  sow project review approve 002`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportID := args[0]

			// Get Sow from context
			s := cmdutil.SowFromContext(cmd.Context())

			// Get project
			project, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Approve review (handles validation, state machine transitions)
			if err := project.ApproveReview(reportID); err != nil {
				return err
			}

			cmd.Printf("\n✓ Approved review report %s\n", reportID)
			return nil
		},
	}

	return cmd
}
