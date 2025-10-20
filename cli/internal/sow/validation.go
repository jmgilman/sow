package sow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/cli/schemas"
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	// Path to the file that failed validation (relative to .sow/)
	Path string

	// Schema type being validated against
	SchemaType string

	// The underlying validation error
	Err error
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s (%s): %v", v.Path, v.SchemaType, v.Err)
}

// ValidationResult aggregates multiple validation errors.
type ValidationResult struct {
	Errors []ValidationError
}

// HasErrors returns true if there are any validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Error implements the error interface, formatting all errors.
func (r *ValidationResult) Error() string {
	if !r.HasErrors() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("validation failed with %d error(s):\n", len(r.Errors)))
	for i, err := range r.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

// Add appends a validation error to the result.
func (r *ValidationResult) Add(path, schemaType string, err error) {
	if err != nil {
		r.Errors = append(r.Errors, ValidationError{
			Path:       path,
			SchemaType: schemaType,
			Err:        err,
		})
	}
}

// Validate performs comprehensive validation of the entire .sow directory structure.
//
// This validates all files that exist without fail-fast behavior, collecting
// all validation errors into a single result. Optional components (project, tasks)
// are only validated if they exist.
//
// Validated components:
//   - .sow/refs/index.json (if exists)
//   - .sow/refs/index.local.json (if exists)
//   - .sow/project/state.yaml (if exists)
//   - .sow/project/phases/implementation/tasks/*/state.yaml (if exist)
//
// Returns a ValidationResult containing all validation errors found.
// If no errors are found, ValidationResult.HasErrors() returns false.
func (s *Sow) Validate() (*ValidationResult, error) {
	// Load CUE validator
	validator, err := schemas.NewCUEValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to load CUE schemas: %w", err)
	}

	result := &ValidationResult{}

	// Validate refs committed index (if exists)
	s.validateRefsCommittedIndex(result, validator)

	// Validate refs local index (if exists)
	s.validateRefsLocalIndex(result, validator)

	// Validate project (if exists)
	s.validateProject(result, validator)

	return result, nil
}

// validateRefsCommittedIndex validates the committed refs index.
func (s *Sow) validateRefsCommittedIndex(result *ValidationResult, validator *schemas.CUEValidator) {
	path := filepath.Join(".sow", "refs", "index.json")

	// Check if file exists
	_, err := s.fs.Stat(path)
	if err != nil {
		// File is optional - not an error
		return
	}

	// Read file
	data, err := s.readFile(path)
	if err != nil {
		result.Add(path, "refs-committed", fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Validate against schema
	if err := validator.ValidateRefsCommittedIndex(data); err != nil {
		result.Add(path, "refs-committed", err)
	}
}

// validateRefsLocalIndex validates the local refs index.
func (s *Sow) validateRefsLocalIndex(result *ValidationResult, validator *schemas.CUEValidator) {
	path := filepath.Join(".sow", "refs", "index.local.json")

	// Check if file exists
	_, err := s.fs.Stat(path)
	if err != nil {
		// File is optional - not an error
		return
	}

	// Read file
	data, err := s.readFile(path)
	if err != nil {
		result.Add(path, "refs-local", fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Validate against schema
	if err := validator.ValidateRefsLocalIndex(data); err != nil {
		result.Add(path, "refs-local", err)
	}
}

// validateProject validates the project directory and all its contents.
func (s *Sow) validateProject(result *ValidationResult, validator *schemas.CUEValidator) {
	path := filepath.Join(".sow", "project", "state.yaml")

	// Check if project exists
	_, err := s.fs.Stat(path)
	if err != nil {
		// No active project - not an error
		return
	}

	// Validate project state
	data, err := s.readFile(path)
	if err != nil {
		result.Add(path, "project-state", fmt.Errorf("failed to read file: %w", err))
		return
	}

	if err := validator.ValidateProjectState(data); err != nil {
		result.Add(path, "project-state", err)
	}

	// Validate all tasks
	s.validateTasks(result, validator)
}

// validateTasks validates all task state files.
func (s *Sow) validateTasks(result *ValidationResult, validator *schemas.CUEValidator) {
	tasksDir := filepath.Join(".sow", "project", "phases", "implementation", "tasks")

	// Check if tasks directory exists
	_, err := s.fs.Stat(tasksDir)
	if err != nil {
		// No tasks yet - not an error
		return
	}

	// Read directory entries
	entries, err := s.fs.ReadDir(tasksDir)
	if err != nil {
		result.Add(tasksDir, "tasks", fmt.Errorf("failed to read directory: %w", err))
		return
	}

	// Validate each task directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		taskStatePath := filepath.Join(tasksDir, taskID, "state.yaml")

		// Check if state file exists
		_, err := s.fs.Stat(taskStatePath)
		if err != nil {
			result.Add(taskStatePath, "task-state", fmt.Errorf("task directory exists but state.yaml is missing"))
			continue
		}

		// Read state file
		data, err := s.readFile(taskStatePath)
		if err != nil {
			result.Add(taskStatePath, "task-state", fmt.Errorf("failed to read file: %w", err))
			continue
		}

		// Validate against schema
		if err := validator.ValidateTaskState(data); err != nil {
			result.Add(taskStatePath, "task-state", err)
		}
	}
}
