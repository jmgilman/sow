package project

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newPhaseStatusCmd creates the command to show phase status.
//
// Usage:
//   sow project phase status [--format text|json]
//
// Displays the status of all 5 project phases.
func newPhaseStatusCmd(accessor SowFSAccessor) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show phase status",
		Long: `Show the status of all project phases.

Displays:
  - Which phases are enabled
  - Current phase status (pending, in_progress, completed, skipped)
  - Phase-specific details

Output formats:
  - text: Human-readable table (default)
  - json: Machine-readable JSON`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Validate format
			if format != "text" && format != "json" {
				return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", format)
			}

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Get project filesystem
			projectFS, err := sowFS.Project()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first: %w", err)
			}

			// Read current state
			state, err := projectFS.State()
			if err != nil {
				return fmt.Errorf("failed to read project state: %w", err)
			}

			// Output based on format
			switch format {
			case "text":
				output := project.FormatPhaseStatus(state)
				cmd.Print(output)

			case "json":
				// Marshal phases as JSON
				jsonBytes, err := json.MarshalIndent(state.Phases, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal phase status to JSON: %w", err)
				}
				cmd.Println(string(jsonBytes))
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")

	return cmd
}
