package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/context"
)

// NewLogCmd creates the log command
func NewLogCmd() *cobra.Command {
	var files []string
	var action string
	var result string
	var forceProject bool

	cmd := &cobra.Command{
		Use:   "log [flags] <details>",
		Short: "Append structured log entry to project or task log",
		Long: `Append structured log entry to appropriate log.md file.

Auto-detects context (task vs project) by walking directory tree to find .sow/.
Supports both orchestrator and worker logs.

Fast performance (<1s, ideally <100ms).`,
		Example: `  # Task log (from task directory)
  sow log --action created_file --result success --file src/main.go "Created main file"

  # Multiple files
  sow log --action modified_file --result success \
    --file src/main.go --file src/utils.go "Updated files"

  # Project log (force project level)
  sow log --project --action started_phase --result success "Started implement phase"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			details := args[0]

			// Validate required flags
			if action == "" {
				return fmt.Errorf("--action is required")
			}
			if result == "" {
				return fmt.Errorf("--result is required")
			}

			return runLog(cmd, files, action, result, details, forceProject)
		},
	}

	cmd.Flags().StringSliceVar(&files, "file", nil, "File affected (can specify multiple)")
	cmd.Flags().StringVar(&action, "action", "", "Action type (required)")
	cmd.Flags().StringVar(&result, "result", "", "Result: success, error, partial (required)")
	cmd.Flags().BoolVar(&forceProject, "project", false, "Write to project log (default: auto-detect)")

	return cmd
}

func runLog(cmd *cobra.Command, files []string, action, result, details string, forceProject bool) error {
	// Detect context
	ctx, err := context.DetectContext()
	if err != nil {
		return fmt.Errorf("failed to detect context: %w", err)
	}

	if ctx.Type == context.ContextTypeNone {
		return fmt.Errorf("not in a sow repository (no .sow/ directory found)")
	}

	// Override context type if --project flag is set
	if forceProject {
		ctx.Type = context.ContextTypeProject
	}

	// Get log path
	logPath := ctx.GetLogPath()

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file does not exist: %s", logPath)
	}

	// Generate timestamp (ISO 8601)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Build log entry
	var entry strings.Builder
	entry.WriteString("\n")
	entry.WriteString(fmt.Sprintf("## %s", timestamp))

	// Add agent ID for task logs
	if ctx.Type == context.ContextTypeTask {
		agentID := ctx.GetAgentID()
		entry.WriteString(fmt.Sprintf(" - %s", agentID))
	}

	entry.WriteString("\n\n")
	entry.WriteString(fmt.Sprintf("**Action**: %s\n\n", action))
	entry.WriteString(fmt.Sprintf("**Result**: %s\n\n", result))

	// Add files if provided
	if len(files) > 0 {
		entry.WriteString("**Files**:\n")
		for _, file := range files {
			entry.WriteString(fmt.Sprintf("- %s\n", file))
		}
		entry.WriteString("\n")
	}

	entry.WriteString(fmt.Sprintf("**Details**: %s\n\n", details))
	entry.WriteString("---\n")

	// Append to log file (atomic operation)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry.String()); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	// Output confirmation
	fmt.Fprintf(cmd.OutOrStdout(), "Log entry added to %s\n", logPath)

	return nil
}
