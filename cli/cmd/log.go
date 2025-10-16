package cmd

import (
	"github.com/spf13/cobra"
)

// NewLogCmd creates the log command.
func NewLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Create a log entry",
		Long: `Create a log entry in the appropriate log file.

Auto-detects context (task vs project) and writes to the correct log:
  - Task context: .sow/project/phases/implementation/tasks/<id>/log.md
  - Project context: .sow/project/log.md

Log entries are formatted markdown with timestamp, agent ID, action,
result, files modified, and optional notes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("Log entry created")
			return nil
		},
	}

	// Flags
	cmd.Flags().StringP("action", "a", "", "Action performed (required)")
	cmd.Flags().StringP("result", "r", "", "Result of action (required)")
	cmd.Flags().StringSliceP("files", "f", []string{}, "Files modified")
	cmd.Flags().StringP("notes", "n", "", "Additional notes")
	cmd.Flags().Bool("project", false, "Force project-level log (ignore task context)")

	_ = cmd.MarkFlagRequired("action")
	_ = cmd.MarkFlagRequired("result")

	return cmd
}
