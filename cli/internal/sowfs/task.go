package sowfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jmgilman/sow/cli/schemas"
)

// TaskFS provides access to a specific task directory.
//
// Task directory structure:
//
//	.sow/project/phases/implementation/tasks/{id}/
//	├── state.yaml       - Task metadata
//	├── description.md   - Task requirements
//	├── log.md           - Worker action log
//	└── feedback/        - Human corrections (optional)
//	    ├── 001.md
//	    └── 002.md
//
// All paths are relative to the task directory.
type TaskFS interface {
	// ID returns the task ID (gap-numbered, e.g., "010")
	ID() string

	// State reads and validates the task state file.
	// Returns the parsed and validated TaskState struct.
	// File: state.yaml
	State() (*schemas.TaskState, error)

	// WriteState validates and writes the task state file.
	// Validates against CUE schema before writing.
	// File: state.yaml
	WriteState(state *schemas.TaskState) error

	// AppendLog appends an entry to the task log.
	// Entry should be formatted markdown.
	// File: log.md
	AppendLog(entry string) error

	// ReadLog reads the entire task log.
	// File: log.md
	ReadLog() (string, error)

	// ReadDescription reads the task description.
	// File: description.md
	ReadDescription() (string, error)

	// WriteDescription writes the task description.
	// File: description.md
	WriteDescription(content string) error

	// ListFeedback lists all feedback file names.
	// Returns filenames like ["001.md", "002.md"]
	// Directory: feedback/
	ListFeedback() ([]string, error)

	// ReadFeedback reads a specific feedback file.
	// Filename should be just the file name (e.g., "001.md")
	// Directory: feedback/
	ReadFeedback(filename string) (string, error)

	// WriteFeedback writes a feedback file.
	// Filename should be just the file name (e.g., "001.md")
	// Creates feedback/ directory if needed.
	// Directory: feedback/
	WriteFeedback(filename string, content string) error

	// Path returns the absolute path to the task directory.
	// Format: {repoRoot}/.sow/project/phases/implementation/tasks/{id}
	Path() string
}

// TaskFSImpl is the concrete implementation of TaskFS.
type TaskFSImpl struct {
	// sowFS is the parent SowFS (for accessing chrooted .sow filesystem)
	sowFS *SowFSImpl

	// taskID is the gap-numbered task identifier (e.g., "010")
	taskID string

	// taskPath is the path relative to .sow/
	// Format: project/phases/implementation/tasks/{id}
	taskPath string

	// validator provides CUE schema validation
	validator *schemas.CUEValidator
}

// Ensure TaskFSImpl implements TaskFS.
var _ TaskFS = (*TaskFSImpl)(nil)

// NewTaskFS creates a new TaskFS instance.
func NewTaskFS(sowFS *SowFSImpl, taskID string, validator *schemas.CUEValidator) *TaskFSImpl {
	return &TaskFSImpl{
		sowFS:     sowFS,
		taskID:    taskID,
		taskPath:  "project/phases/implementation/tasks/" + taskID,
		validator: validator,
	}
}

// ID returns the task ID.
func (t *TaskFSImpl) ID() string {
	return t.taskID
}

// State reads the task state.
func (t *TaskFSImpl) State() (*schemas.TaskState, error) {
	statePath := filepath.Join(t.taskPath, "state.yaml")

	// Read state file
	data, err := t.sowFS.fs.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to read task state: %w", err)
	}

	// Validate against CUE schema
	if err := t.validator.ValidateTaskState(data); err != nil {
		return nil, fmt.Errorf("task state validation failed: %w", err)
	}

	// Parse YAML
	var state schemas.TaskState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse task state YAML: %w", err)
	}

	return &state, nil
}

// WriteState writes the task state.
func (t *TaskFSImpl) WriteState(state *schemas.TaskState) error {
	statePath := filepath.Join(t.taskPath, "state.yaml")

	// Encode to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal task state: %w", err)
	}

	// Validate against CUE schema
	if err := t.validator.ValidateTaskState(data); err != nil {
		return fmt.Errorf("task state validation failed: %w", err)
	}

	// Ensure task directory exists
	if err := t.sowFS.fs.MkdirAll(t.taskPath, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Write to file
	if err := t.sowFS.fs.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	return nil
}

// AppendLog appends to task log.
func (t *TaskFSImpl) AppendLog(entry string) error {
	logPath := filepath.Join(t.taskPath, "log.md")

	// Read existing log
	existingLog := ""
	data, err := t.sowFS.fs.ReadFile(logPath)
	if err == nil {
		existingLog = string(data)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing log: %w", err)
	}

	// Append new entry
	newLog := existingLog
	if existingLog != "" && !strings.HasSuffix(existingLog, "\n") {
		newLog += "\n"
	}
	newLog += entry
	if !strings.HasSuffix(entry, "\n") {
		newLog += "\n"
	}

	// Ensure task directory exists
	if err := t.sowFS.fs.MkdirAll(t.taskPath, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Write updated log
	if err := t.sowFS.fs.WriteFile(logPath, []byte(newLog), 0644); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// ReadLog reads the task log.
func (t *TaskFSImpl) ReadLog() (string, error) {
	logPath := filepath.Join(t.taskPath, "log.md")

	data, err := t.sowFS.fs.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty string if log doesn't exist yet
			return "", nil
		}
		return "", fmt.Errorf("failed to read log: %w", err)
	}

	return string(data), nil
}

// ReadDescription reads the task description.
func (t *TaskFSImpl) ReadDescription() (string, error) {
	descPath := filepath.Join(t.taskPath, "description.md")

	data, err := t.sowFS.fs.ReadFile(descPath)
	if err != nil {
		return "", fmt.Errorf("failed to read description: %w", err)
	}

	return string(data), nil
}

// WriteDescription writes the task description.
func (t *TaskFSImpl) WriteDescription(content string) error {
	descPath := filepath.Join(t.taskPath, "description.md")

	// Ensure task directory exists
	if err := t.sowFS.fs.MkdirAll(t.taskPath, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	if err := t.sowFS.fs.WriteFile(descPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write description: %w", err)
	}
	return nil
}

// ListFeedback lists feedback files.
func (t *TaskFSImpl) ListFeedback() ([]string, error) {
	feedbackPath := filepath.Join(t.taskPath, "feedback")

	// Check if feedback directory exists
	exists, err := t.sowFS.fs.Exists(feedbackPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check feedback directory: %w", err)
	}
	if !exists {
		// No feedback directory means no feedback files
		return []string{}, nil
	}

	// Read directory entries
	entries, err := t.sowFS.fs.ReadDir(feedbackPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feedback directory: %w", err)
	}

	// Filter for files only
	var feedbackFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			feedbackFiles = append(feedbackFiles, entry.Name())
		}
	}

	return feedbackFiles, nil
}

// ReadFeedback reads a feedback file.
func (t *TaskFSImpl) ReadFeedback(filename string) (string, error) {
	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)
	feedbackPath := filepath.Join(t.taskPath, "feedback", filename)

	data, err := t.sowFS.fs.ReadFile(feedbackPath)
	if err != nil {
		return "", fmt.Errorf("failed to read feedback: %w", err)
	}

	return string(data), nil
}

// WriteFeedback writes a feedback file.
func (t *TaskFSImpl) WriteFeedback(filename string, content string) error {
	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)
	feedbackPath := filepath.Join(t.taskPath, "feedback", filename)

	// Create feedback directory if needed
	feedbackDir := filepath.Join(t.taskPath, "feedback")
	if err := t.sowFS.fs.MkdirAll(feedbackDir, 0755); err != nil {
		return fmt.Errorf("failed to create feedback directory: %w", err)
	}

	if err := t.sowFS.fs.WriteFile(feedbackPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write feedback: %w", err)
	}
	return nil
}

// Path returns the absolute path to the task directory.
func (t *TaskFSImpl) Path() string {
	return t.sowFS.repoRoot + "/.sow/" + t.taskPath
}
