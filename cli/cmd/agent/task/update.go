package task

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"fmt"

	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the task update command.
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [task-id]",
		Short: "Update task properties",
		Long: `Update properties of an existing task.

Currently supports updating:
  - Status (pending, in_progress, completed, abandoned)

Status transitions automatically update timestamps:
  - in_progress: Sets started_at if not already set
  - completed/abandoned: Sets completed_at and started_at if not set

Both the lightweight task entry (in project state) and the detailed
task state (in task directory) are updated together.

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task update 010 --status in_progress  # Update task 010
  sow task update --status completed        # Update inferred task`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("status", "s", "", "New task status (pending, in_progress, completed, abandoned)")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	statusFlag, _ := cmd.Flags().GetString("status")

	// At least one flag must be specified
	if statusFlag == "" {
		return fmt.Errorf("at least one update flag must be specified (--status)")
	}

	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Get project
	proj, err := projectpkg.Load(ctx)
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

	// Track what was updated
	updated := []string{}

	// Update status if provided
	if statusFlag != "" {
		oldStatus := t.Status()

		// Update status (handles project state, task state, timestamps, and state machine transitions)
		if err := t.SetStatus(statusFlag); err != nil {
			return err
		}

		updated = append(updated, fmt.Sprintf("status: %s → %s", oldStatus, statusFlag))

		// Check if we transitioned to review phase
		if statusFlag == "completed" || statusFlag == "abandoned" {
			// The SetStatus method already handles the state machine transition if all tasks are complete
			// Just provide user feedback if we're now in review
			currentState := proj.Machine().State()
			if currentState.String() == "ReviewActive" {
				cmd.Printf("\n✓ All tasks complete. Transitioning to review phase.\n")
			}
		}
	}

	// Print success message
	cmd.Printf("✓ Updated task %s\n", taskID)
	for _, update := range updated {
		cmd.Printf("  %s\n", update)
	}

	return nil
}
