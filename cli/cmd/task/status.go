package task

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	sowpkg "github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the task status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [task-id]",
		Short: "Show detailed status for a specific task",
		Long: `Display detailed status information for a specific task.

Shows comprehensive task information including:
  - Task metadata (ID, name, phase, status)
  - Timestamps (created, started, completed, updated)
  - Iteration and assigned agent
  - References to related resources
  - Feedback history count
  - Modified files

Output format:
  - Default: Human-readable formatted text
  - --format json: Machine-readable JSON

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task status 010          # Show status for task 010
  sow task status              # Show status for inferred task
  sow task status --format json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskStatus(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("format", "f", "text", "Output format (text or json)")

	return cmd
}

func runTaskStatus(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", format)
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

	// Get task state
	taskState, err := t.State()
	if err != nil {
		return fmt.Errorf("failed to read task state: %w", err)
	}

	// Output based on format
	if format == "json" {
		// JSON output: serialize task state
		jsonData, err := json.MarshalIndent(taskState, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal task state to JSON: %w", err)
		}
		cmd.Println(string(jsonData))
	} else {
		// Text output: use formatted display
		output := sowpkg.FormatTaskStatus(taskState)
		cmd.Print(output)
	}

	return nil
}
