package task

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newStateIncrementCmd creates the state increment command.
func newStateIncrementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increment [task-id]",
		Short: "Increment the task iteration counter",
		Long: `Increment the task's iteration counter.

The iteration counter tracks how many times a worker has attempted this task.
It's typically incremented when:
  - Human provides feedback requiring a retry
  - Task fails and needs reattempting
  - Orchestrator reassigns the task

Incrementing the iteration counter updates the task's state.yaml file
and sets the updated_at timestamp.

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task state increment 010     # Increment task 010
  sow task state increment         # Increment inferred task`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateIncrement(cmd, args)
		},
	}

	return cmd
}

func runStateIncrement(cmd *cobra.Command, args []string) error {
	// Get Sow from context
	s := sowFromContext(cmd.Context())

	// Get project
	proj, err := s.GetProject()
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Resolve task ID (from args or infer)
	taskID, err := resolveTaskID(proj, args)
	if err != nil {
		return fmt.Errorf("failed to resolve task ID: %w", err)
	}

	// Get task
	t, err := proj.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task '%s' not found: %w", taskID, err)
	}

	// Store old value for output
	taskState, err := t.State()
	if err != nil {
		return fmt.Errorf("failed to read task state: %w", err)
	}
	oldIteration := taskState.Task.Iteration

	// Increment iteration (auto-saves)
	if err := t.IncrementIteration(); err != nil {
		return err
	}

	// Get new state
	newTaskState, err := t.State()
	if err != nil {
		return fmt.Errorf("failed to read updated task state: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Incremented iteration: %d → %d\n", oldIteration, newTaskState.Task.Iteration)

	return nil
}
