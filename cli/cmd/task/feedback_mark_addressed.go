package task

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"fmt"

	"github.com/spf13/cobra"
)

// newFeedbackMarkAddressedCmd creates the feedback mark-addressed command.
func newFeedbackMarkAddressedCmd() *cobra.Command {
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
			return runFeedbackMarkAddressed(cmd, args)
		},
	}

	return cmd
}

func runFeedbackMarkAddressed(cmd *cobra.Command, args []string) error {
	// First arg is always the feedback ID
	feedbackID := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	// Get Sow from context
	s := cmdutil.SowFromContext(cmd.Context())

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

	// Mark feedback as addressed (auto-saves)
	if err := t.MarkFeedbackAddressed(feedbackID); err != nil {
		return err
	}

	// Print success message
	cmd.Printf("✓ Marked feedback %s as addressed\n", feedbackID)

	return nil
}
