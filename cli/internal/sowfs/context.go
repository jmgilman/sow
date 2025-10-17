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
