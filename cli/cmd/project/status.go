package project

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the project status command.
func NewStatusCmd() *cobra.Command {
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
			return runStatus(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("format", "f", "text", "Output format (text or json)")

	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", format)
	}

	// Get Sow from context
	sow := sowFromContext(cmd.Context())

	// Get project
	proj, err := sow.GetProject()
	if err != nil {
		return fmt.Errorf("no active project found - run 'sow project init' to create one")
	}

	// Get project state
	state := proj.State()

	// Output based on format
	if format == "json" {
		// JSON output: serialize the entire state
		jsonData, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal project state to JSON: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	} else {
		// Text output: use formatted display
		output := project.FormatStatus(state)
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	return nil
}
