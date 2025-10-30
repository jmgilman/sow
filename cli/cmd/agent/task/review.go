package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewReviewCmd creates the task review command.
func NewReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review [task-id]",
		Short: "Review a task and approve or request changes",
		Long: `Review a task that has status "needs_review" and make a decision.

The orchestrator should:
1. Read task requirements from description.md
2. Check task state.yaml for files_modified list
3. Review actual changes using git diff
4. Write review to: project/phases/implementation/tasks/<id>/review.md

Then execute one of these commands:

APPROVE:
  Marks the task as completed and preserves review.md as approval record.

REQUEST CHANGES:
  Converts review.md into feedback, increments iteration counter,
  and returns task to "in_progress" status for worker to address.

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "needs_review")

Examples:
  sow agent task review 010 --approve           # Approve task 010
  sow agent task review --request-changes       # Request changes on inferred task`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReview(cmd, args)
		},
	}

	// Flags
	cmd.Flags().Bool("approve", false, "Approve the task review (marks task as completed)")
	cmd.Flags().Bool("request-changes", false, "Request changes (converts review to feedback, returns task to in_progress)")

	// Mutually exclusive flags
	cmd.MarkFlagsMutuallyExclusive("approve", "request-changes")

	return cmd
}

func runReview(cmd *cobra.Command, args []string) error {
	approveFlag, _ := cmd.Flags().GetBool("approve")
	requestChangesFlag, _ := cmd.Flags().GetBool("request-changes")

	// Exactly one flag must be specified
	if !approveFlag && !requestChangesFlag {
		return fmt.Errorf("exactly one flag must be specified: --approve or --request-changes")
	}

	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Get project
	proj, err := loader.Load(ctx)
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

	// Verify task is in needs_review status
	if t.Status() != "needs_review" {
		return fmt.Errorf("task '%s' is not in 'needs_review' status (current: %s)", taskID, t.Status())
	}

	// Verify review.md exists
	reviewPath := filepath.Join("project/phases/implementation/tasks", taskID, "review.md")
	reviewContent, err := proj.ReadFile(reviewPath)
	if err != nil {
		absPath := filepath.Join(".sow", reviewPath)
		return fmt.Errorf("review.md not found at %s - orchestrator must write review before approving/rejecting", absPath)
	}

	if approveFlag {
		return approveReview(cmd, proj, t, taskID)
	}
	return requestChanges(cmd, proj, t, taskID, reviewContent)
}

func approveReview(cmd *cobra.Command, proj domain.Project, t *domain.Task, taskID string) error {
	// Transition status: needs_review → completed
	// SetStatus handles timestamps automatically
	if err := t.SetStatus("completed"); err != nil {
		return fmt.Errorf("failed to set task status to completed: %w", err)
	}

	// Save project state
	if err := proj.Machine().Save(); err != nil {
		return fmt.Errorf("failed to save project state: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Task %s approved and marked as completed\n", taskID)
	cmd.Printf("  Review preserved at: project/phases/implementation/tasks/%s/review.md\n", taskID)

	// Check if we transitioned to review phase
	currentState := proj.Machine().State()
	if currentState.String() == "ReviewActive" {
		cmd.Printf("\n✓ All tasks complete. Transitioning to review phase.\n")
	}

	return nil
}

func requestChanges(cmd *cobra.Command, proj domain.Project, t *domain.Task, taskID string, reviewContent []byte) error {
	// Read review.md content (already validated to exist)
	reviewMessage := string(reviewContent)

	// Generate next feedback ID and create feedback entry
	feedbackID, err := t.AddFeedback(reviewMessage)
	if err != nil {
		return fmt.Errorf("failed to add feedback: %w", err)
	}

	// Move review.md to feedback/
	reviewPath := filepath.Join("project/phases/implementation/tasks", taskID, "review.md")
	feedbackPath := filepath.Join("project/phases/implementation/tasks", taskID, "feedback", feedbackID+".md")

	// Read and write to new location
	if err := proj.WriteFile(feedbackPath, reviewContent); err != nil {
		return fmt.Errorf("failed to write feedback file: %w", err)
	}

	// Delete review.md (we've moved it to feedback/)
	// Note: We need to use the full path from root
	reviewFullPath := filepath.Join(".sow", reviewPath)
	if err := os.Remove(reviewFullPath); err != nil {
		// Non-fatal - feedback was created successfully
		cmd.PrintErrf("Warning: failed to remove review.md: %v\n", err)
	}

	// Increment iteration counter
	if err := t.IncrementIteration(); err != nil {
		return fmt.Errorf("failed to increment iteration: %w", err)
	}

	// Transition status: needs_review → in_progress
	// SetStatus will automatically clear completed_at for the new iteration
	if err := t.SetStatus("in_progress"); err != nil {
		return fmt.Errorf("failed to set task status to in_progress: %w", err)
	}

	// Save project state
	if err := proj.Machine().Save(); err != nil {
		return fmt.Errorf("failed to save project state: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Changes requested for task %s\n", taskID)
	cmd.Printf("  Feedback added: project/phases/implementation/tasks/%s/feedback/%s.md\n", taskID, feedbackID)
	cmd.Printf("  Task returned to in_progress with iteration incremented\n")
	cmd.Printf("  Worker will be re-invoked to address feedback\n")

	return nil
}
