package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// newPhaseApproveCmd creates the command to approve a phase plan.
//
// Usage:
//   sow project phase approve <phase>
//
// Currently only supports:
//   - implementation: Approves task plan and transitions to execution mode
func newPhaseApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <phase>",
		Short: "Approve a phase plan",
		Long: `Approve a phase plan and transition to the next state.

Currently supported phases:
  - implementation: Approve the task plan created during planning.
                   This transitions from planning to autonomous execution.

Human approval gates ensure critical decisions are reviewed before
autonomous execution begins.

Example:
  # Approve implementation task plan
  sow project phase approve implementation`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Get Sow from context
			s := cmdutil.SowFromContext(cmd.Context())

			// Get project
			project, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Only implementation phase supports approval currently
			if phase != "implementation" {
				return fmt.Errorf("only 'implementation' phase supports approval, got: %s", phase)
			}

			// Approve tasks (handles validation, state machine transitions)
			if err := project.ApproveTasks(); err != nil {
				return err
			}

			cmd.Printf("\n✓ Approved %s phase plan\n", phase)
			cmd.Printf("→ Transitioning to autonomous execution\n")
			return nil
		},
	}

	return cmd
}
