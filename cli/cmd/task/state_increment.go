package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// newStateIncrementCmd creates the state increment command.
func newStateIncrementCmd(accessor SowFSAccessor) *cobra.Command {
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
			return runStateIncrement(cmd, args, accessor)
		},
	}

	return cmd
}

func runStateIncrement(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Resolve task ID (either from args or inferred)
	taskID, err := taskutil.ResolveTaskIDFromArgs(sowFS, args)
	if err != nil {
		return fmt.Errorf("failed to resolve task ID: %w", err)
	}

	// Validate task ID format
	if err := task.ValidateTaskID(taskID); err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	// Get project (must exist)
	projectFS, err := sowFS.Project()
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Get TaskFS
	taskFS, err := projectFS.Task(taskID)
	if err != nil {
		return fmt.Errorf("task '%s' not found: %w", taskID, err)
	}

	// Read task state
	taskState, err := taskFS.State()
	if err != nil {
		return fmt.Errorf("failed to read task state: %w", err)
	}

	// Store old value for output
	oldIteration := taskState.Task.Iteration

	// Increment iteration
	if err := task.IncrementTaskIteration(taskState); err != nil {
		return fmt.Errorf("failed to increment iteration: %w", err)
	}

	// Write updated state
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Incremented iteration: %d → %d\n", oldIteration, taskState.Task.Iteration)

	return nil
}
