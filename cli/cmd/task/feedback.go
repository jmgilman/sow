package task

import (
	"github.com/spf13/cobra"
)

// newFeedbackCmd creates the feedback command for managing task feedback.
//
// Usage:
//   sow task feedback <subcommand>
//
// Subcommands:
//   - add: Create new feedback for a task
//   - mark-addressed: Mark feedback as addressed
func newFeedbackCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Manage human feedback and corrections",
		Long: `Manage human feedback and corrections for tasks.

Feedback enables human-in-the-loop corrections during task execution.
When a worker makes a mistake or needs guidance, humans can provide
feedback that guides the next iteration.

Feedback workflow:
  1. Human provides feedback (sow task feedback add)
  2. Orchestrator increments iteration (sow task state increment)
  3. Worker reads feedback from feedback/ directory
  4. Worker addresses feedback and marks it (sow task feedback mark-addressed)

Each feedback item has:
  - A unique ID (001, 002, 003...)
  - A status (pending, addressed, superseded)
  - A markdown file with the feedback content
  - Timestamps for tracking

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")`,
	}

	// Add subcommands
	cmd.AddCommand(newFeedbackAddCmd(accessor))
	cmd.AddCommand(newFeedbackMarkAddressedCmd(accessor))

	return cmd
}
