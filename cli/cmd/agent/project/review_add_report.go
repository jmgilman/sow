package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newReviewAddReportCmd creates the command to add a review report.
//
// Usage:
//   sow project review add-report <path> --assessment <assessment>
//
// Arguments:
//   <path>: Path to report file (relative to .sow/project/phases/review/)
//
// Flags:
//   --assessment: Assessment result (pass or fail, required)
func newReviewAddReportCmd() *cobra.Command {
	var assessment string

	cmd := &cobra.Command{
		Use:   "add-report <path>",
		Short: "Add a review report with assessment",
		Long: `Add a review report to the review phase.

Review reports document the outcome of a review iteration. Each report must have
an assessment of either "pass" or "fail":
  - pass: Implementation meets requirements, can proceed to finalize
  - fail: Issues found, must loop back to implementation

Reports are automatically numbered (001, 002, 003...) based on the order they
are added.

Examples:
  # Add passing review report
  sow project review add-report reports/001-review.md --assessment pass

  # Add failing review report (requires loop-back)
  sow project review add-report reports/002-review.md --assessment fail`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportPath := args[0]

			// Validate assessment
			if assessment != "pass" && assessment != "fail" {
				return fmt.Errorf("invalid assessment '%s': must be 'pass' or 'fail'", assessment)
			}

			// Get Sow from context
			ctx := cmdutil.GetContext(cmd.Context())

			// Get project
			proj, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Add review report (handles state machine transition based on assessment)
			if err := proj.AddReviewReport(reportPath, assessment); err != nil {
				return err
			}

			// Get the report ID from updated state
			state := proj.State()
			reportID := state.Phases.Review.Reports[len(state.Phases.Review.Reports)-1].Id

			cmd.Printf("✓ Added review report %s (%s)\n", reportID, assessment)
			cmd.Printf("  %s\n", reportPath)

			if assessment == "pass" {
				cmd.Println("\n→ Review passed. Transitioning to finalize phase.")
			} else {
				cmd.Println("\n→ Review failed. Looping back to implementation planning.")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&assessment, "assessment", "a", "", "Assessment result (pass or fail, required)")
	_ = cmd.MarkFlagRequired("assessment")

	return cmd
}
