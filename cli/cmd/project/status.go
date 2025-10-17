package project

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the project status command.
func NewStatusCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		Long: `Display current project status including phases and tasks.

Shows:
  - Project metadata (name, branch, description)
  - Phase status (enabled/disabled, current state)
  - Task summary (counts by status)

Output format:
  - Default: Human-readable formatted text
  - --format json: Machine-readable JSON

Example:
  sow project status
  sow project status --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, accessor)
		},
	}

	// Flags
	cmd.Flags().StringP("format", "f", "text", "Output format (text or json)")

	return cmd
}

func runStatus(cmd *cobra.Command, _ []string, accessor SowFSAccessor) error {
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", format)
	}

	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Get ProjectFS (will error if no project exists)
	projectFS, err := sowFS.Project()
	if err != nil {
		return fmt.Errorf("no active project found - run 'sow project init' to create one")
	}

	// Read project state
	state, err := projectFS.State()
	if err != nil {
		return fmt.Errorf("failed to read project state: %w", err)
	}

	// Output based on format
	if format == "json" {
		// JSON output: serialize the entire state
		jsonData, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal project state to JSON: %w", err)
		}
		cmd.Println(string(jsonData))
	} else {
		// Text output: use formatted display
		output := project.FormatStatus(state)
		cmd.Print(output)
	}

	return nil
}
