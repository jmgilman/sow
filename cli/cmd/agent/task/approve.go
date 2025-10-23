package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewApproveCmd creates the command to approve the task plan.
//
// Usage:
//   sow agent task approve
//
// Approves the implementation task plan created during the planning phase.
// This transitions from planning to autonomous execution mode.
func NewApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve the task plan",
		Long: `Approve the implementation task plan.

This command approves the task plan created during the implementation planning
phase. After approval, the project transitions from planning to autonomous
execution mode where workers can begin implementing tasks.

Human approval gates ensure critical decisions are reviewed before
autonomous execution begins.

Example:
  # Approve the task plan
  sow agent task approve`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get Sow from context
			ctx := cmdutil.GetContext(cmd.Context())

			// Get project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent init' first")
			}

			// Approve tasks (handles validation, state machine transitions)
			if err := project.ApproveTasks(); err != nil {
				return err
			}

			cmd.Printf("\n✓ Approved task plan\n")
			cmd.Printf("→ Transitioning to autonomous execution\n")
			return nil
		},
	}

	return cmd
}
