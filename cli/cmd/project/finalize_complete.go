package project

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"fmt"

	"github.com/spf13/cobra"
)

// newFinalizeCompleteCmd creates the command to complete a finalize subphase.
//
// Usage:
//   sow project finalize complete <subphase>
//
// Arguments:
//   <subphase>: The subphase to complete (documentation or checks)
//
// The finalize phase has three subphases:
//   1. documentation - Update documentation files
//   2. checks - Run tests, linters, and build
//   3. delete - Delete project directory (uses `sow project delete`)
func newFinalizeCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <subphase>",
		Short: "Complete a finalize subphase",
		Long: `Complete a finalize subphase and advance to the next step.

The finalize phase has three subphases:
  1. documentation - Update documentation files (README, API docs, etc.)
  2. checks - Run final validation (tests, linters, build)
  3. delete - Delete project directory (use 'sow project delete')

This command is used to signal completion of documentation updates or final checks,
allowing the state machine to advance to the next finalize step.

Examples:
  # Complete documentation subphase
  sow project finalize complete documentation

  # Complete checks subphase
  sow project finalize complete checks`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subphase := args[0]

			// Validate subphase
			if subphase != "documentation" && subphase != "checks" {
				return fmt.Errorf("invalid subphase '%s': must be 'documentation' or 'checks'", subphase)
			}

			// Get Sow from context
			s := cmdutil.SowFromContext(cmd.Context())

			// Get project
			project, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Complete finalize subphase (handles validation, state machine transitions)
			if err := project.CompleteFinalizeSubphase(subphase); err != nil {
				return err
			}

			cmd.Printf("\n✓ Completed %s subphase\n", subphase)

			// Provide next step guidance
			switch subphase {
			case "documentation":
				cmd.Println("\n→ Next: Run final checks (tests, linters, build)")
			case "checks":
				cmd.Println("\n→ Next: Delete project directory with 'sow project delete'")
			}

			return nil
		},
	}

	return cmd
}
