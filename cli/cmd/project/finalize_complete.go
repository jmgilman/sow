package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/statechart"
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
func newFinalizeCompleteCmd(accessor SowFSAccessor) *cobra.Command {
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

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Verify project exists
			_, err := sowFS.Project()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first: %w", err)
			}

			// === STATECHART INTEGRATION START ===

			// Load machine
			machine, err := statechart.Load()
			if err != nil {
				return fmt.Errorf("failed to load statechart: %w", err)
			}

			// Validate current state and determine event based on subphase
			currentState := machine.State()
			var event statechart.Event
			var expectedState statechart.State

			if subphase == "documentation" {
				expectedState = statechart.FinalizeDocumentation
				event = statechart.EventDocumentationDone
			} else { // checks
				expectedState = statechart.FinalizeChecks
				event = statechart.EventChecksDone
			}

			// Verify we're in the expected state
			if currentState != expectedState {
				return fmt.Errorf("cannot complete %s in current state: %s (expected %s)", subphase, currentState, expectedState)
			}

			// Fire event (validates transition, outputs prompt)
			if err := machine.Fire(event); err != nil {
				return fmt.Errorf("failed to complete %s subphase: %w", subphase, err)
			}

			// Save state
			if err := machine.Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			// === STATECHART INTEGRATION END ===

			cmd.Printf("\n✓ Completed %s subphase\n", subphase)

			// Provide next step guidance
			if subphase == "documentation" {
				cmd.Println("\n→ Next: Run final checks (tests, linters, build)")
			} else if subphase == "checks" {
				cmd.Println("\n→ Next: Delete project directory with 'sow project delete'")
			}

			return nil
		},
	}

	return cmd
}
