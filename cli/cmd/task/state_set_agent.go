package task

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newStateSetAgentCmd creates the state set-agent command.
func newStateSetAgentCmd() *cobra.Command {
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
			return runStateSetAgent(cmd, args)
		},
	}

	return cmd
}

func runStateSetAgent(cmd *cobra.Command, args []string) error {
	// First arg is always the agent name
	agent := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	// Get Sow from context
	s := sowFromContext(cmd.Context())

	// Get project
	proj, err := s.GetProject()
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Resolve task ID (from args or infer)
	taskID, err := resolveTaskID(proj, taskIDArgs)
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
	oldAgent := taskState.Task.Assigned_agent

	// Set agent (auto-saves)
	if err := t.SetAgent(agent); err != nil {
		return err
	}

	// Print success message
	cmd.Printf("✓ Updated assigned agent: %s → %s\n", oldAgent, agent)

	return nil
}
