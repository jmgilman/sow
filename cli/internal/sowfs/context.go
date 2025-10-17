package sowfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContextType represents the type of workspace context.
type ContextType int

const (
	// ContextNone should never occur in SowFS (we're always in .sow).
	// Kept for completeness but SowFS requires .sow to exist.
	ContextNone ContextType = iota

	// ContextProject indicates we're at the project level.
	// (in .sow/ directory but not in a task).
	ContextProject

	// ContextTask indicates we're inside a task directory.
	ContextTask
)

// String returns a string representation of the context type.
func (c ContextType) String() string {
	switch c {
	case ContextNone:
		return "none"
	case ContextProject:
		return "project"
	case ContextTask:
		return "task"
	default:
		return "unknown"
	}
}

// WorkspaceContext represents the detected workspace context.
type WorkspaceContext struct {
	// Type indicates what kind of context this is
	Type ContextType

	// TaskID is populated when Type == ContextTask
	// Format: gap-numbered (e.g., "010", "020")
	TaskID string
}

// ContextFS provides workspace context detection.
//
// This domain helps determine where commands are running from:
// - Inside a task directory.
// - At the project root.
type ContextFS interface {
	// Detect determines the workspace context from the current working directory.
	// Returns WorkspaceContext with detected context information.
	Detect() (*WorkspaceContext, error)

	// InferTaskID infers the task ID for task commands that make ID optional.
	//
	// Inference strategy (in order):
	//   1. Current directory: If in task directory, use that task ID
	//   2. Active task: If exactly one task has status "in_progress", use that ID
	//   3. Error: If neither applies or multiple tasks in progress
	//
	// Returns:
	//   - Task ID if successfully inferred
	//   - Error with helpful message if inference fails
	InferTaskID() (string, error)
}

// ContextFSImpl is the concrete implementation of ContextFS.
type ContextFSImpl struct {
	// sowFS is the parent SowFS (for accessing repo root and sow dir)
	sowFS *SowFSImpl
}

// Ensure ContextFSImpl implements ContextFS.
var _ ContextFS = (*ContextFSImpl)(nil)

// NewContextFS creates a new ContextFS instance.
func NewContextFS(sowFS *SowFSImpl) *ContextFSImpl {
	return &ContextFSImpl{
		sowFS: sowFS,
	}
}

// Detect determines the workspace context from the current working directory.
func (c *ContextFSImpl) Detect() (*WorkspaceContext, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if we're inside a task directory
	taskID, inTask := c.isInTaskDirectory(cwd)
	if inTask {
		return &WorkspaceContext{
			Type:   ContextTask,
			TaskID: taskID,
		}, nil
	}

	// We're in the repository with .sow/ but not in a task
	return &WorkspaceContext{
		Type: ContextProject,
	}, nil
}

// isInTaskDirectory checks if the current path is inside a task directory.
//
// Task directories have the pattern:
// {repo}/.sow/project/phases/implementation/tasks/{id}/
//
// Returns:
//   - taskID: the gap-numbered task ID (e.g., "010")
//   - bool: true if inside a task directory
func (c *ContextFSImpl) isInTaskDirectory(currentPath string) (string, bool) {
	// Get path relative to .sow directory
	relPath, err := filepath.Rel(c.sowFS.SowDir(), currentPath)
	if err != nil {
		return "", false
	}

	// Check if path starts with project/phases/implementation/tasks/
	const taskPrefix = "project/phases/implementation/tasks/"
	if !strings.HasPrefix(relPath, taskPrefix) {
		return "", false
	}

	// Extract the path after the tasks/ directory
	remainder := strings.TrimPrefix(relPath, taskPrefix)

	// The task ID is the first path component
	parts := strings.Split(remainder, string(filepath.Separator))
	if len(parts) == 0 {
		return "", false
	}

	taskID := parts[0]

	// Validate task ID format (gap-numbered)
	if !taskIDPattern.MatchString(taskID) {
		return "", false
	}

	return taskID, true
}

// InferTaskID infers the task ID for commands with optional task ID.
func (c *ContextFSImpl) InferTaskID() (string, error) {
	// Step 1: Check current directory
	ctx, err := c.Detect()
	if err != nil {
		return "", fmt.Errorf("failed to detect context: %w", err)
	}

	// If we're in a task directory, use that task ID
	if ctx.Type == ContextTask {
		return ctx.TaskID, nil
	}

	// Step 2: Check for active task in project state
	// Get project filesystem
	projectFS, err := c.sowFS.Project()
	if err != nil {
		return "", fmt.Errorf("cannot infer task ID: no active project - run 'sow project init' first")
	}

	// Read project state
	state, err := projectFS.State()
	if err != nil {
		return "", fmt.Errorf("failed to read project state: %w", err)
	}

	// Find tasks with status "in_progress"
	var inProgressTasks []string
	for _, task := range state.Phases.Implementation.Tasks {
		if task.Status == "in_progress" {
			inProgressTasks = append(inProgressTasks, task.Id)
		}
	}

	// Step 3: Validate result
	if len(inProgressTasks) == 0 {
		return "", fmt.Errorf("cannot infer task ID: no task currently in progress - use --id flag or run from task directory")
	}

	if len(inProgressTasks) > 1 {
		// Format list of IDs
		idList := strings.Join(inProgressTasks, ", ")
		return "", fmt.Errorf("cannot infer task ID: multiple tasks in progress (%s) - use --id flag to specify", idList)
	}

	// Exactly one in_progress task
	return inProgressTasks[0], nil
}
