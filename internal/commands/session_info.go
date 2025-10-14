package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/config"
	"github.com/your-org/sow/internal/context"
)

// SessionInfo represents the JSON output structure
type SessionInfo struct {
	Context    string  `json:"context"`
	TaskID     *string `json:"task_id,omitempty"`
	Phase      *string `json:"phase,omitempty"`
	CLIVersion string  `json:"cli_version"`
}

// NewSessionInfoCmd creates the session-info command
func NewSessionInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session-info",
		Short: "Display current session information",
		Long: `Detect and report session context (task/project/none).

Used by SessionStart hooks to display context information.
Very fast detection (<100ms).

Outputs JSON for easy parsing by hooks and scripts.`,
		Example: `  # Display session info
  sow session-info

  # Parse output in scripts
  INFO=$(sow session-info)
  CONTEXT=$(echo $INFO | jq -r '.context')`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSessionInfo(cmd)
		},
	}

	return cmd
}

func runSessionInfo(cmd *cobra.Command) error {
	// Detect context
	ctx, err := context.DetectContext()
	if err != nil {
		return fmt.Errorf("failed to detect context: %w", err)
	}

	// Build output structure
	info := SessionInfo{
		Context:    ctx.Type,
		CLIVersion: config.Version,
	}

	// Add task-specific fields if in task context
	if ctx.Type == context.ContextTypeTask {
		info.TaskID = &ctx.TaskID
		info.Phase = &ctx.Phase
	}

	// Marshal to JSON
	output, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to stdout
	fmt.Fprintln(cmd.OutOrStdout(), string(output))

	return nil
}
