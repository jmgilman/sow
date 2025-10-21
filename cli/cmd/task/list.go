package task

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	sowpkg "github.com/jmgilman/sow/cli/internal/sow"
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

	// Get project
	proj, err := projectpkg.Load(ctx)
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Get state
	state := proj.State()

	// Get tasks from implementation phase
	tasks := state.Phases.Implementation.Tasks

	// Output based on format
	if format == "json" {
		// JSON output: serialize tasks array
		jsonData, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
		}
		cmd.Println(string(jsonData))
	} else {
		// Text output: use formatted display
		output := sowpkg.FormatTaskList(tasks)
		cmd.Print(output)
	}

	return nil
}
