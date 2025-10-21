package cmd

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/logging"
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

	// Validate action and result early
	if err := logging.ValidateAction(action); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := logging.ValidateResult(result); err != nil {
		return fmt.Errorf("%w", err)
	}

	// Get contexts
	ctx := cmdutil.GetContext(cmd.Context())
	sowCtx := cmdutil.GetContext(cmd.Context())

	// Detect workspace context
	contextType, taskID := sow.DetectContext(ctx.RepoRoot())

	// Determine agent ID based on context
	agentID, err := determineAgentID(sowCtx, contextType, taskID, forceProject)
	if err != nil {
		return fmt.Errorf("failed to determine agent ID: %w", err)
	}

	// Create log entry
	entry := logging.LogEntry{
		Timestamp: time.Now(),
		AgentID:   agentID,
		Action:    action,
		Result:    result,
		Files:     files,
		Notes:     notes,
	}

	// Validate entry
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid log entry: %w", err)
	}

	// Format entry
	formatted := entry.Format()

	// Append to appropriate log
	if err := appendToLog(sowCtx, contextType, taskID, formatted, forceProject); err != nil {
		return fmt.Errorf("failed to append log entry: %w", err)
	}

	// Success message
	logType := "task"
	if contextType == "project" || forceProject {
		logType = "project"
	}
	cmd.Printf("✓ Log entry added to %s log\n", logType)

	return nil
}

// determineAgentID constructs the agent ID based on workspace context.
//
// For task context: reads iteration from task state and builds "{role}-{iteration}".
// For project context: uses "orchestrator".
func determineAgentID(sowCtx *sow.Context, contextType, taskID string, forceProject bool) (string, error) {
	// If forcing project level, always use orchestrator
	if forceProject || contextType == "project" {
		return "orchestrator", nil
	}

	// We're in a task context - need to read iteration from task state
	if contextType == "task" {
		// Get project
		proj, err := projectpkg.Load(sowCtx)
		if err != nil {
			return "", fmt.Errorf("failed to access project: %w", err)
		}

		// Get task
		task, err := proj.GetTask(taskID)
		if err != nil {
			return "", fmt.Errorf("failed to access task %s: %w", taskID, err)
		}

		// Read task state
		state, err := task.State()
		if err != nil {
			return "", fmt.Errorf("failed to read task state: %w", err)
		}

		// Extract agent and iteration from state
		agentRole := state.Task.Assigned_agent
		iteration := int(state.Task.Iteration)

		// Build agent ID: {assigned_agent}-{iteration}
		return logging.BuildAgentID(agentRole, iteration), nil
	}

	return "", fmt.Errorf("unknown context type: %v", contextType)
}

// appendToLog writes the formatted entry to the appropriate log file.
func appendToLog(sowCtx *sow.Context, contextType, taskID, entry string, forceProject bool) error {
	// Get project
	proj, err := projectpkg.Load(sowCtx)
	if err != nil {
		return fmt.Errorf("failed to access project: %w", err)
	}

	// Force project level or if context is project
	if forceProject || contextType == "project" {
		if err := proj.AppendLog(entry); err != nil {
			return fmt.Errorf("failed to append to project log: %w", err)
		}
		return nil
	}

	// Task level
	if contextType == "task" {
		// Get task
		task, err := proj.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to access task: %w", err)
		}

		if err := task.AppendLog(entry); err != nil {
			return fmt.Errorf("failed to append to task log: %w", err)
		}
		return nil
	}

	return fmt.Errorf("cannot append log in context type: %v", contextType)
}
