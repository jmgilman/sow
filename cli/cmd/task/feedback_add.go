package task

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"fmt"

	"github.com/spf13/cobra"
)

// newFeedbackAddCmd creates the feedback add command.
func newFeedbackAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <message> [task-id]",
		Short: "Add feedback to a task",
		Long: `Add human feedback to guide the next task iteration.

Creates a new feedback entry in the task state and writes a feedback
file to the task's feedback/ directory. The feedback file uses a structured
markdown format for consistency.

Feedback flow:
  1. Human notices issue with task work
  2. Human adds feedback: sow task feedback add "Use RS256 not HS256" 010
  3. Orchestrator sees pending feedback, increments iteration
  4. Worker reads feedback/001.md and addresses the issue
  5. Worker marks feedback addressed: sow task feedback mark-addressed 001 010

The feedback ID is auto-generated (001, 002, 003...) based on existing
feedback for the task.

Feedback files are created at:
  .sow/project/phases/implementation/tasks/{id}/feedback/{feedback-id}.md

Use --increment-iteration to automatically bump the iteration counter
after adding feedback (common workflow).

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task feedback add "Use RS256 instead of HS256" 010
  sow task feedback add "Add error handling" --increment-iteration
  sow task feedback add "Follow PEP 8 style guide"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFeedbackAdd(cmd, args)
		},
	}

	// Flags
	cmd.Flags().BoolP("increment-iteration", "i", false, "Automatically increment task iteration after adding feedback")

	return cmd
}

func runFeedbackAdd(cmd *cobra.Command, args []string) error {
	// First arg is always the feedback message
	message := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	incrementIteration, _ := cmd.Flags().GetBool("increment-iteration")

	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Get project
	proj, err := projectpkg.Load(ctx)
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

	// Store old iteration if we're incrementing
	var oldIteration int64
	if incrementIteration {
		taskState, err := t.State()
		if err != nil {
			return fmt.Errorf("failed to read task state: %w", err)
		}
		oldIteration = taskState.Task.Iteration
	}

	// Add feedback (auto-saves, creates file, returns feedback ID)
	feedbackID, err := t.AddFeedback(message)
	if err != nil {
		return err
	}

	// Increment iteration if requested
	if incrementIteration {
		if err := t.IncrementIteration(); err != nil {
			return err
		}
	}

	// Print success message
	cmd.Printf("✓ Created feedback %s\n", feedbackID)
	feedbackPath := fmt.Sprintf(".sow/project/phases/implementation/tasks/%s/feedback/%s.md", taskID, feedbackID)
	cmd.Printf("  Feedback file: %s\n", feedbackPath)

	if incrementIteration {
		newTaskState, err := t.State()
		if err != nil {
			return fmt.Errorf("failed to read updated task state: %w", err)
		}
		cmd.Printf("  Iteration: %d → %d\n", oldIteration, newTaskState.Task.Iteration)
	}

	return nil
}
