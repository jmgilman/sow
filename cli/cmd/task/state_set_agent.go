package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// newStateSetAgentCmd creates the state set-agent command.
func newStateSetAgentCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-agent <agent> [task-id]",
		Short: "Change the assigned agent for a task",
		Long: `Change which agent type should execute this task.

The assigned agent determines which type of worker will execute the task.
Common agent types:
  - implementer: For code implementation tasks
  - reviewer: For code review tasks
  - architect: For design and architecture tasks
  - planner: For task breakdown and planning

Changing the assigned agent updates the task's state.yaml file
and sets the updated_at timestamp.

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task state set-agent reviewer 010    # Assign reviewer to task 010
  sow task state set-agent implementer     # Assign implementer to inferred task`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateSetAgent(cmd, args, accessor)
		},
	}

	return cmd
}

func runStateSetAgent(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	// First arg is always the agent name
	agent := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Resolve task ID (either from args or inferred)
	taskID, err := taskutil.ResolveTaskIDFromArgs(sowFS, taskIDArgs)
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
	oldAgent := taskState.Task.Assigned_agent

	// Set agent
	if err := task.SetTaskAgent(taskState, agent); err != nil {
		return fmt.Errorf("failed to set agent: %w", err)
	}

	// Write updated state
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Updated assigned agent: %s → %s\n", oldAgent, agent)

	return nil
}
