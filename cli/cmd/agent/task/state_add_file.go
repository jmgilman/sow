package task

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"fmt"

	"github.com/spf13/cobra"
)

// newStateAddFileCmd creates the state add-file command.
func newStateAddFileCmd() *cobra.Command {
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
			return runStateAddFile(cmd, args)
		},
	}

	return cmd
}

func runStateAddFile(cmd *cobra.Command, args []string) error {
	// First arg is always the file path
	filePath := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Get project
	proj, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Resolve task ID (from args or infer)
	taskID, err := resolveTaskID(proj, taskIDArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve task ID: %w", err)
	}

	// Get task
	t, err := proj.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task '%s' not found: %w", taskID, err)
	}

	// Check if already exists (for better messaging)
	taskState, err := t.State()
	if err != nil {
		return fmt.Errorf("failed to read task state: %w", err)
	}
	alreadyExists := false
	for _, file := range taskState.Task.Files_modified {
		if file == filePath {
			alreadyExists = true
			break
		}
	}

	// Add file (auto-saves)
	if err := t.AddFile(filePath); err != nil {
		return err
	}

	// Print success message
	if alreadyExists {
		cmd.Printf("✓ File already tracked: %s\n", filePath)
	} else {
		cmd.Printf("✓ Added modified file: %s\n", filePath)
	}

	return nil
}
