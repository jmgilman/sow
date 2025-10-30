package task

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
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

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				return fmt.Errorf("no active phase found")
			}

			// Approve tasks via Phase interface
			result, err := phase.ApproveTasks()
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("phase %s does not support task approval", phase.Name())
			}
			if err != nil {
				return fmt.Errorf("failed to approve tasks: %w", err)
			}

			// Fire event if phase returned one
			if result.Event != "" {
				machine := proj.Machine()
				if err := machine.Fire(result.Event); err != nil {
					return fmt.Errorf("failed to fire event %s: %w", result.Event, err)
				}
				// Save after transition
				if err := proj.Save(); err != nil {
					return fmt.Errorf("failed to save project state: %w", err)
				}
			}

			cmd.Printf("\n✓ Approved task plan\n")
			cmd.Printf("→ Transitioning to autonomous execution\n")
			return nil
		},
	}

	return cmd
}
