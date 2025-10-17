package sowfs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jmgilman/sow/cli/schemas"
)

// Gap-numbered task ID regex (010, 020, 030, etc.)
// Pattern: at least 3 digits.
// Auto-generated IDs use increments of 10, but manual IDs can be any 3+ digit number.
var taskIDPattern = regexp.MustCompile(`^[0-9]{3,}$`)

// ProjectFS provides access to the .sow/project/ directory.
//
// This domain handles active project state including:
//   - state.yaml - Project state across all 5 phases
//   - log.md - Orchestrator action log
//   - context/ - Project-specific context files
//   - phases/ - Phase-specific directories with tasks
//
// All paths are relative to .sow/project/.
type ProjectFS interface {
	// State reads and validates the project state file.
	// Returns the parsed and validated ProjectState struct.
	// File: .sow/project/state.yaml
	State() (*schemas.ProjectState, error)

	// WriteState validates and writes the project state file.
	// Validates against CUE schema before writing.
	// File: .sow/project/state.yaml
	WriteState(state *schemas.ProjectState) error

	// AppendLog appends an entry to the project log.
	// Entry should be formatted markdown.
	// File: .sow/project/log.md
	AppendLog(entry string) error

	// ReadLog reads the entire project log.
	// File: .sow/project/log.md
	ReadLog() (string, error)

	// Task returns a TaskFS for a specific task.
	// taskID should be gap-numbered (e.g., "010", "020")
	// Returns ErrTaskNotFound if task doesn't exist
	// Returns ErrInvalidTaskID if taskID format is invalid
	Task(taskID string) (*TaskFSImpl, error)

	// TaskUnchecked returns a TaskFS for a specific task without checking existence.
	// This should be used when creating a new task.
	// taskID should be gap-numbered (e.g., "010", "020")
	// Returns ErrInvalidTaskID if taskID format is invalid
	TaskUnchecked(taskID string) (*TaskFSImpl, error)

	// Tasks returns TaskFS instances for all tasks.
	// Only includes tasks in the implementation phase.
	Tasks() ([]*TaskFSImpl, error)

	// Exists checks if the project directory exists.
	// Returns false if no active project.
	Exists() (bool, error)

	// ReadContext reads a context file from .sow/project/context/
	// Path is relative to .sow/project/context/
	ReadContext(path string) ([]byte, error)

	// WriteContext writes a context file to .sow/project/context/
	// Path is relative to .sow/project/context/
	WriteContext(path string, data []byte) error

	// ListContextFiles lists all files in .sow/project/context/
	// Returns paths relative to .sow/project/context/
	ListContextFiles() ([]string, error)

	// Delete removes the entire project directory.
	// This is typically called during the finalize phase before creating a PR.
	// The design requires no project files to be present in merged code.
	// Returns an error if deletion fails.
	Delete() error
}

// ProjectFSImpl is the concrete implementation of ProjectFS.
type ProjectFSImpl struct {
	// sowFS is the parent SowFS (for accessing chrooted .sow filesystem)
	sowFS *SowFSImpl

	// validator provides CUE schema validation
	validator *schemas.CUEValidator
}

// Ensure ProjectFSImpl implements ProjectFS.
var _ ProjectFS = (*ProjectFSImpl)(nil)

// NewProjectFS creates a new ProjectFS instance.
func NewProjectFS(sowFS *SowFSImpl, validator *schemas.CUEValidator) *ProjectFSImpl {
	return &ProjectFSImpl{
		sowFS:     sowFS,
		validator: validator,
	}
}

// State reads the project state.
func (p *ProjectFSImpl) State() (*schemas.ProjectState, error) {
	statePath := "project/state.yaml"

	// Read state file
	data, err := p.sowFS.fs.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to read project state: %w", err)
	}

	// Validate against CUE schema
	if err := p.validator.ValidateProjectState(data); err != nil {
		return nil, fmt.Errorf("project state validation failed: %w", err)
	}

	// Parse YAML
	var state schemas.ProjectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse project state YAML: %w", err)
	}

	return &state, nil
}

// WriteState writes the project state.
func (p *ProjectFSImpl) WriteState(state *schemas.ProjectState) error {
	statePath := "project/state.yaml"

	// Encode to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal project state: %w", err)
	}

	// Validate against CUE schema
	if err := p.validator.ValidateProjectState(data); err != nil {
		return fmt.Errorf("project state validation failed: %w", err)
	}

	// Ensure project directory exists
	if err := p.sowFS.fs.MkdirAll("project", 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Write to file
	if err := p.sowFS.fs.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project state: %w", err)
	}

	return nil
}

// AppendLog appends to project log.
func (p *ProjectFSImpl) AppendLog(entry string) error {
	logPath := "project/log.md"

	// Read existing log
	existingLog := ""
	data, err := p.sowFS.fs.ReadFile(logPath)
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

	// Ensure project directory exists
	if err := p.sowFS.fs.MkdirAll("project", 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Write updated log
	if err := p.sowFS.fs.WriteFile(logPath, []byte(newLog), 0644); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// ReadLog reads the project log.
func (p *ProjectFSImpl) ReadLog() (string, error) {
	logPath := "project/log.md"

	data, err := p.sowFS.fs.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty string if log doesn't exist yet
			return "", nil
		}
		return "", fmt.Errorf("failed to read log: %w", err)
	}

	return string(data), nil
}

// Task returns a TaskFS for a specific task.
func (p *ProjectFSImpl) Task(taskID string) (*TaskFSImpl, error) {
	// Validate task ID format (gap-numbered)
	if !taskIDPattern.MatchString(taskID) {
		return nil, ErrInvalidTaskID
	}

	// Check if task directory exists
	taskPath := filepath.Join("project/phases/implementation/tasks", taskID)
	exists, err := p.sowFS.fs.Exists(taskPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check task directory: %w", err)
	}
	if !exists {
		return nil, ErrTaskNotFound
	}

	// Return TaskFS instance
	return NewTaskFS(p.sowFS, taskID, p.validator), nil
}

// TaskUnchecked returns a TaskFS for a specific task without checking if it exists.
// This should be used when creating a new task.
func (p *ProjectFSImpl) TaskUnchecked(taskID string) (*TaskFSImpl, error) {
	// Validate task ID format (gap-numbered)
	if !taskIDPattern.MatchString(taskID) {
		return nil, ErrInvalidTaskID
	}

	// Return TaskFS instance without existence check
	return NewTaskFS(p.sowFS, taskID, p.validator), nil
}

// Tasks returns all TaskFS instances.
func (p *ProjectFSImpl) Tasks() ([]*TaskFSImpl, error) {
	tasksPath := "project/phases/implementation/tasks"

	// Check if tasks directory exists
	exists, err := p.sowFS.fs.Exists(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check tasks directory: %w", err)
	}
	if !exists {
		// No tasks directory means no tasks
		return []*TaskFSImpl{}, nil
	}

	// Read directory entries
	entries, err := p.sowFS.fs.ReadDir(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	// Filter for directories with valid task IDs
	var tasks []*TaskFSImpl
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		if taskIDPattern.MatchString(taskID) {
			tasks = append(tasks, NewTaskFS(p.sowFS, taskID, p.validator))
		}
	}

	return tasks, nil
}

// Exists checks if project exists.
func (p *ProjectFSImpl) Exists() (bool, error) {
	statePath := "project/state.yaml"
	exists, err := p.sowFS.fs.Exists(statePath)
	if err != nil {
		return false, fmt.Errorf("failed to check if project exists: %w", err)
	}
	return exists, nil
}

// ReadContext reads a context file.
func (p *ProjectFSImpl) ReadContext(path string) ([]byte, error) {
	contextPath := filepath.Join("project/context", path)
	data, err := p.sowFS.fs.ReadFile(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read context file %s: %w", path, err)
	}
	return data, nil
}

// WriteContext writes a context file.
func (p *ProjectFSImpl) WriteContext(path string, data []byte) error {
	contextPath := filepath.Join("project/context", path)

	// Create parent directories if needed
	dir := filepath.Dir(contextPath)
	if err := p.sowFS.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create context directories: %w", err)
	}

	if err := p.sowFS.fs.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file %s: %w", path, err)
	}
	return nil
}

// ListContextFiles lists all context files.
func (p *ProjectFSImpl) ListContextFiles() ([]string, error) {
	contextPath := "project/context"

	// Check if context directory exists
	exists, err := p.sowFS.fs.Exists(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check context directory: %w", err)
	}
	if !exists {
		// No context directory means no files
		return []string{}, nil
	}

	// Walk the context directory to collect all files
	var files []string
	err = p.sowFS.fs.Walk(contextPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only collect files
		if d.IsDir() {
			return nil
		}

		// Make path relative to context directory
		relPath, err := filepath.Rel(contextPath, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %s: %w", path, err)
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk context directory: %w", err)
	}

	return files, nil
}

// Delete removes the entire project directory.
func (p *ProjectFSImpl) Delete() error {
	projectPath := "project"

	// Check if project exists
	exists, err := p.sowFS.fs.Exists(projectPath)
	if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}
	if !exists {
		return ErrProjectNotFound
	}

	// Remove the entire project directory
	if err := p.sowFS.fs.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("failed to delete project directory: %w", err)
	}

	return nil
}
