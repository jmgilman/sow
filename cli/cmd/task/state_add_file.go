package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// newStateAddFileCmd creates the state add-file command.
func newStateAddFileCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-file <path> [task-id]",
		Short: "Track a file that was modified during task execution",
		Long: `Track a file that was modified during task execution.

Workers use this command to record which files they changed while working
on a task. The orchestrator uses this information for:
  - Review phase context (what changed)
  - Dependency tracking (which tasks touch which files)
  - Impact analysis (understanding scope of changes)

Duplicates are automatically filtered - adding the same file twice
has no effect.

Paths should be relative to the repository root:
  ✓ src/auth/jwt.py
  ✓ docs/architecture/auth.md
  ✗ /abs/path/to/file.py
  ✗ ../outside/repo/file.py

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task state add-file src/auth/jwt.py 010
  sow task state add-file src/auth/jwt_test.py
  sow task state add-file docs/api/auth.md`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateAddFile(cmd, args, accessor)
		},
	}

	return cmd
}

func runStateAddFile(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	// First arg is always the file path
	filePath := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Resolve task ID (either from args or inferred)
	taskID, err := taskutil.ResolveTaskIDFromArgs(sowFS, taskIDArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve task ID: %w", err)
	}

	// Validate task ID format
	if err := task.ValidateTaskID(taskID); err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	// Get project (must exist)
	projectFS, err := sowFS.Project()
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Get TaskFS
	taskFS, err := projectFS.Task(taskID)
	if err != nil {
		return fmt.Errorf("task '%s' not found: %w", taskID, err)
	}

	// Read task state
	taskState, err := taskFS.State()
	if err != nil {
		return fmt.Errorf("failed to read task state: %w", err)
	}

	// Check if already exists (for better messaging)
	alreadyExists := false
	for _, file := range taskState.Task.Files_modified {
		if file == filePath {
			alreadyExists = true
			break
		}
	}

	// Add file
	if err := task.AddModifiedFile(taskState, filePath); err != nil {
		return fmt.Errorf("failed to add modified file: %w", err)
	}

	// Write updated state
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Print success message
	if alreadyExists {
		cmd.Printf("✓ File already tracked: %s\n", filePath)
	} else {
		cmd.Printf("✓ Added modified file: %s\n", filePath)
	}

	return nil
}
