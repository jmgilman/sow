package task

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// newFeedbackAddCmd creates the feedback add command.
func newFeedbackAddCmd(accessor SowFSAccessor) *cobra.Command {
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
			return runFeedbackAdd(cmd, args, accessor)
		},
	}

	// Flags
	cmd.Flags().BoolP("increment-iteration", "i", false, "Automatically increment task iteration after adding feedback")

	return cmd
}

func runFeedbackAdd(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	// First arg is always the feedback message
	message := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	incrementIteration, _ := cmd.Flags().GetBool("increment-iteration")

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

	// Generate feedback ID
	feedbackID := task.GenerateNextFeedbackID(taskState)

	// Add feedback to state
	if err := task.AddFeedback(taskState, feedbackID); err != nil {
		return fmt.Errorf("failed to add feedback: %w", err)
	}

	// Increment iteration if requested
	if incrementIteration {
		if err := task.IncrementTaskIteration(taskState); err != nil {
			return fmt.Errorf("failed to increment iteration: %w", err)
		}
	}

	// Write updated state
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Create feedback file
	feedbackFilename := fmt.Sprintf("%s.md", feedbackID)
	feedbackContent := fmt.Sprintf(`# Feedback %s

**Created:** %s
**Status:** pending

## Issue

%s

## Required Changes

(Describe specific changes needed)

## Context

(Optional: Add any additional context or references)
`, feedbackID, time.Now().Format(time.RFC3339), message)

	if err := taskFS.WriteFeedback(feedbackFilename, feedbackContent); err != nil {
		return fmt.Errorf("failed to write feedback file: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Created feedback %s\n", feedbackID)
	feedbackPath := fmt.Sprintf(".sow/project/phases/implementation/tasks/%s/feedback/%s", taskID, feedbackFilename)
	cmd.Printf("  Feedback file: %s\n", feedbackPath)

	if incrementIteration {
		cmd.Printf("  Iteration: %d → %d\n", taskState.Task.Iteration-1, taskState.Task.Iteration)
	}

	return nil
}
