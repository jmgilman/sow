package task

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/jmgilman/sow/cli/internal/taskutil"
	"github.com/spf13/cobra"
)

// newStateAddReferenceCmd creates the state add-reference command.
func newStateAddReferenceCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-reference <path> [task-id]",
		Short: "Add a context reference to a task",
		Long: `Add a context reference path that workers should read when executing the task.

References are paths relative to .sow/ that provide context for task execution.
Common reference types:
  - refs/: External knowledge (style guides, conventions)
  - knowledge/: Repository knowledge (ADRs, architecture docs)
  - project/: Project-specific context (discovery notes, design docs)

The orchestrator uses references to compile context for workers. Workers
read all referenced files before starting task execution.

Duplicates are automatically filtered - adding the same reference twice
has no effect.

Paths should be relative to .sow/ directory:
  ✓ refs/python-style/conventions.md
  ✓ knowledge/adrs/001-use-jwt.md
  ✗ /abs/path/to/file.md
  ✗ ../outside/sow/file.md

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")

Examples:
  sow task state add-reference refs/python-style/conventions.md 010
  sow task state add-reference knowledge/adrs/001-use-jwt.md`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateAddReference(cmd, args, accessor)
		},
	}

	return cmd
}

func runStateAddReference(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	// First arg is always the reference path
	referencePath := args[0]

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
	for _, ref := range taskState.Task.References {
		if ref == referencePath {
			alreadyExists = true
			break
		}
	}

	// Add reference
	if err := task.AddTaskReference(taskState, referencePath); err != nil {
		return fmt.Errorf("failed to add reference: %w", err)
	}

	// Write updated state
	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Print success message
	if alreadyExists {
		cmd.Printf("✓ Reference already exists: %s\n", referencePath)
	} else {
		cmd.Printf("✓ Added reference: %s\n", referencePath)
	}

	return nil
}
