package task

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/spf13/cobra"
)

// NewListCmd creates the task list command.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks in the implementation phase",
		Long: `List all tasks in the current project's implementation phase.

Shows all tasks with their ID, status, and name. Tasks are sorted by ID.

Output format:
  - Default: Human-readable formatted table
  - --format json: Machine-readable JSON

Example:
  sow task list
  sow task list --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("format", "f", "text", "Output format (text or json)")

	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", format)
	}

	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Load project via loader to get interface
	proj, err := loader.Load(ctx)
	if err != nil {
		if errors.Is(err, project.ErrNoProject) {
			return fmt.Errorf("no active project - run 'sow agent init' first")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Get current phase
	phase := proj.CurrentPhase()
	if phase == nil {
		return fmt.Errorf("no active phase found")
	}

	// Get tasks via Phase interface
	domainTasks := phase.ListTasks()

	// Convert domain.Task to phases.Task for formatting
	phaseTasks := make([]phases.Task, len(domainTasks))
	for i, t := range domainTasks {
		// Get task state to extract schema fields
		taskState, err := t.State()
		if err != nil {
			return fmt.Errorf("failed to get task %s state: %w", t.ID, err)
		}
		phaseTasks[i] = phases.Task{
			Id:           taskState.Task.Id,
			Name:         taskState.Task.Name,
			Status:       taskState.Task.Status,
			Parallel:     false, // Not stored in TaskState currently
			Dependencies: nil,   // Not stored in TaskState currently
		}
	}

	// Output based on format
	if format == "json" {
		// JSON output: serialize tasks array
		jsonData, err := json.MarshalIndent(phaseTasks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
		}
		cmd.Println(string(jsonData))
	} else {
		// Text output: use formatted display
		output := formatTaskList(phaseTasks)
		cmd.Print(output)
	}

	return nil
}
