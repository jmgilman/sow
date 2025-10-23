package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/jmgilman/sow/cli/internal/sow"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLog(cmd, args)
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

func runLog(cmd *cobra.Command, _ []string) error {
	// Get flags
	action, _ := cmd.Flags().GetString("action")
	result, _ := cmd.Flags().GetString("result")
	files, _ := cmd.Flags().GetStringSlice("files")
	notes, _ := cmd.Flags().GetString("notes")
	forceProject, _ := cmd.Flags().GetBool("project")

	// Build options
	var opts []domain.LogOption
	if len(files) > 0 {
		opts = append(opts, domain.WithFiles(files...))
	}
	if notes != "" {
		opts = append(opts, domain.WithNotes(notes))
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())
	contextType, taskID := sow.DetectContext(ctx.RepoRoot())

	// Load project
	proj, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Log to appropriate level
	if forceProject || contextType == "project" {
		return logToProject(cmd, proj, action, result, opts)
	}

	if contextType == "task" {
		return logToTask(cmd, proj, taskID, action, result, opts)
	}

	return fmt.Errorf("unknown context type: %s", contextType)
}

// logToProject logs an entry to the project log.
func logToProject(cmd *cobra.Command, proj domain.Project, action, result string, opts []domain.LogOption) error {
	if err := proj.Log(action, result, opts...); err != nil {
		return err
	}
	cmd.Println("✓ Log entry added to project log")
	return nil
}

// logToTask logs an entry to a task log.
func logToTask(cmd *cobra.Command, proj domain.Project, taskID, action, result string, opts []domain.LogOption) error {
	task, err := proj.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if err := task.Log(action, result, opts...); err != nil {
		return err
	}
	cmd.Println("✓ Log entry added to task log")
	return nil
}
