package task

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newStateAddReferenceCmd creates the state add-reference command.
func newStateAddReferenceCmd() *cobra.Command {
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
			return runStateAddReference(cmd, args)
		},
	}

	return cmd
}

func runStateAddReference(cmd *cobra.Command, args []string) error {
	// First arg is always the reference path
	referencePath := args[0]

	// Remaining args are for task ID resolution
	taskIDArgs := args[1:]

	// Get Sow from context
	s := sowFromContext(cmd.Context())

	// Get project
	proj, err := s.GetProject()
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
	for _, ref := range taskState.Task.References {
		if ref == referencePath {
			alreadyExists = true
			break
		}
	}

	// Add reference (auto-saves)
	if err := t.AddReference(referencePath); err != nil {
		return err
	}

	// Print success message
	if alreadyExists {
		cmd.Printf("✓ Reference already exists: %s\n", referencePath)
	} else {
		cmd.Printf("✓ Added reference: %s\n", referencePath)
	}

	return nil
}
