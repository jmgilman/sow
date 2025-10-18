package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// newFeedbackMarkAddressedCmd creates the feedback mark-addressed command.
func newFeedbackMarkAddressedCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-addressed <feedback-id> [task-id]",
		Short: "Mark feedback as addressed",
		Long: `Mark a feedback item as addressed by a worker.

Updates the feedback status from "pending" to "addressed" in the task's
state.yaml file. This signals that the worker has incorporated the feedback
and the issue has been resolved.

Workers typically call this after:
  1. Reading the feedback file (feedback/001.md)
  2. Making the requested changes
  3. Verifying the changes work correctly

The feedback file itself remains unchanged - only the status in state.yaml
is updated. This maintains the historical record while tracking completion.

Feedback statuses:
  - pending: Feedback waiting to be addressed
  - addressed: Worker has incorporated the feedback
  - superseded: Feedback no longer relevant (replaced by newer feedback)

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task feedback mark-addressed 001 010    # Mark feedback 001 in task 010
  sow task feedback mark-addressed 002        # Mark feedback 002 in inferred task
  sow task feedback mark-addressed 001`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFeedbackMarkAddressed(cmd, args, accessor)
		},
	}

	return cmd
}

func runFeedbackMarkAddressed(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	// First arg is always the feedback ID
	feedbackID := args[0]

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

	// Mark feedback as addressed
	if err := task.MarkFeedbackAddressed(taskState, feedbackID); err != nil {
		return fmt.Errorf("failed to mark feedback as addressed: %w", err)
	}

	// Write updated state
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Marked feedback %s as addressed\n", feedbackID)

	return nil
}
