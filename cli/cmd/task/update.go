package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/statechart"
	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the task update command.
func NewUpdateCmd(accessor SowFSAccessor) *cobra.Command {
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
			return runUpdate(cmd, args, accessor)
		},
	}

	// Flags
	cmd.Flags().StringP("status", "s", "", "New task status (pending, in_progress, completed, abandoned)")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	statusFlag, _ := cmd.Flags().GetString("status")

	// At least one flag must be provided
	if statusFlag == "" {
		return fmt.Errorf("at least one update flag must be specified (--status)")
	}

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

	// Read project state
	projectState, err := projectFS.State()
	if err != nil {
		return fmt.Errorf("failed to read project state: %w", err)
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

	// Track what was updated
	updated := []string{}

	// Update status if provided
	if statusFlag != "" {
		oldStatus := taskState.Task.Status

		// Update detailed task state (with timestamps)
		if err := task.UpdateTaskStatus(taskState, statusFlag); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		// Update lightweight project state entry
		if err := task.UpdateTaskStatusInProject(projectState, taskID, statusFlag); err != nil {
			return fmt.Errorf("failed to update project state: %w", err)
		}

		updated = append(updated, fmt.Sprintf("status: %s → %s", oldStatus, statusFlag))
	}

	// Write updated states
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	if err := projectFS.WriteState(projectState); err != nil {
		return fmt.Errorf("failed to write project state: %w", err)
	}

	// === STATECHART INTEGRATION: Auto-fire EventAllTasksComplete ===
	// Check if status was updated to completed or abandoned
	if statusFlag == "completed" || statusFlag == "abandoned" {
		// Check if ALL tasks are now complete
		if statechart.AllTasksComplete(projectState) {
			// Load statechart machine
			machine, err := statechart.Load()
			if err != nil {
				return fmt.Errorf("failed to load statechart: %w", err)
			}

			// Verify we're in ImplementationExecuting state
			currentState := machine.State()
			if currentState == statechart.ImplementationExecuting {
				// Fire event to transition to ReviewActive
				if err := machine.Fire(statechart.EventAllTasksComplete); err != nil {
					return fmt.Errorf("failed to transition to review: %w", err)
				}

				// Save state with new statechart state
				if err := machine.Save(); err != nil {
					return fmt.Errorf("failed to save statechart state: %w", err)
				}

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
