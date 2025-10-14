package context

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Context types
const (
	ContextTypeNone    = "none"
	ContextTypeProject = "project"
	ContextTypeTask    = "task"
)

// Context represents the current working context (task/project/none)
type Context struct {
	Type      string // "task", "project", or "none"
	SowRoot   string // Path to .sow/ directory
	TaskID    string // Task ID (if in task context)
	Phase     string // Phase name (if in task context)
	Iteration int    // Iteration number (if in task context)
	AgentRole string // Agent role (if in task context)
}

// TaskState represents minimal task state needed for context detection
type taskState struct {
	Task struct {
		ID            string `yaml:"id"`
		Phase         string `yaml:"phase"`
		Iteration     int    `yaml:"iteration"`
		AssignedAgent string `yaml:"assigned_agent"`
	} `yaml:"task"`
}

// DetectContext detects the current working context by walking up directory tree
func DetectContext() (*Context, error) {
	// Find .sow/ directory by walking up
	sowRoot, err := findSowRoot()
	if err != nil {
		// Not in a sow repository
		return &Context{Type: ContextTypeNone}, nil
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if we're in a task directory
	taskInfo, err := findTaskContext(cwd, sowRoot)
	if err != nil {
		return nil, err
	}

	if taskInfo != nil {
		return taskInfo, nil
	}

	// Otherwise, we're at project level
	return &Context{
		Type:    ContextTypeProject,
		SowRoot: sowRoot,
	}, nil
}

// findSowRoot walks up the directory tree to find .sow/
func findSowRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		sowPath := filepath.Join(dir, ".sow")
		if info, err := os.Stat(sowPath); err == nil && info.IsDir() {
			return sowPath, nil
		}

		// Move up one level
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", fmt.Errorf(".sow directory not found")
		}
		dir = parent
	}
}

// findTaskContext checks if current directory is within a task
func findTaskContext(cwd, sowRoot string) (*Context, error) {
	// Pattern to match task directory: .sow/project/phases/{phase}/tasks/{taskID}
	taskPattern := regexp.MustCompile(`/\.sow/project/phases/([^/]+)/tasks/([^/]+)(?:/|$)`)

	// Check if cwd matches task pattern
	matches := taskPattern.FindStringSubmatch(cwd)
	if matches == nil {
		return nil, nil
	}

	phase := matches[1]
	taskID := matches[2]

	// Read task state.yaml to get iteration and agent role
	taskDir := filepath.Join(sowRoot, "project", "phases", phase, "tasks", taskID)
	stateFile := filepath.Join(taskDir, "state.yaml")

	data, err := os.ReadFile(stateFile)
	if err != nil {
		// Task directory exists but no state.yaml - not a valid task context
		return nil, nil
	}

	var state taskState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse task state: %w", err)
	}

	return &Context{
		Type:      ContextTypeTask,
		SowRoot:   sowRoot,
		TaskID:    state.Task.ID,
		Phase:     state.Task.Phase,
		Iteration: state.Task.Iteration,
		AgentRole: state.Task.AssignedAgent,
	}, nil
}

// GetAgentID constructs agent ID from role and iteration (e.g., "implementer-1")
func (c *Context) GetAgentID() string {
	if c.Type != ContextTypeTask {
		return ""
	}
	return fmt.Sprintf("%s-%d", c.AgentRole, c.Iteration)
}

// GetLogPath returns the path to the appropriate log.md file
func (c *Context) GetLogPath() string {
	if c.Type == ContextTypeTask {
		return filepath.Join(c.SowRoot, "project", "phases", c.Phase, "tasks", c.TaskID, "log.md")
	}
	return filepath.Join(c.SowRoot, "project", "log.md")
}
