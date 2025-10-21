package sow

import (
	"fmt"
	"os"
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
func Validate(repoRoot string) (*ValidationResult, error) {
	// Load CUE validator
	validator, err := schemas.NewCUEValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to load CUE schemas: %w", err)
	}

	result := &ValidationResult{}

	// Validate refs committed index (if exists)
	validateRefsCommittedIndex(repoRoot, result, validator)

	// Validate refs local index (if exists)
	validateRefsLocalIndex(repoRoot, result, validator)

	// Validate project (if exists)
	validateProject(repoRoot, result, validator)

	return result, nil
}

// validateRefsCommittedIndex validates the committed refs index.
func validateRefsCommittedIndex(repoRoot string, result *ValidationResult, validator *schemas.CUEValidator) {
	relPath := filepath.Join(".sow", "refs", "index.json")
	absPath := filepath.Join(repoRoot, relPath)

	// Check if file exists
	_, err := os.Stat(absPath)
	if err != nil {
		// File is optional - not an error
		return
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		result.Add(relPath, "refs-committed", fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Validate against schema
	if err := validator.ValidateRefsCommittedIndex(data); err != nil {
		result.Add(relPath, "refs-committed", err)
	}
}

// validateRefsLocalIndex validates the local refs index.
func validateRefsLocalIndex(repoRoot string, result *ValidationResult, validator *schemas.CUEValidator) {
	relPath := filepath.Join(".sow", "refs", "index.local.json")
	absPath := filepath.Join(repoRoot, relPath)

	// Check if file exists
	_, err := os.Stat(absPath)
	if err != nil {
		// File is optional - not an error
		return
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		result.Add(relPath, "refs-local", fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Validate against schema
	if err := validator.ValidateRefsLocalIndex(data); err != nil {
		result.Add(relPath, "refs-local", err)
	}
}

// validateProject validates the project directory and all its contents.
func validateProject(repoRoot string, result *ValidationResult, validator *schemas.CUEValidator) {
	relPath := filepath.Join(".sow", "project", "state.yaml")
	absPath := filepath.Join(repoRoot, relPath)

	// Check if project exists
	_, err := os.Stat(absPath)
	if err != nil {
		// No active project - not an error
		return
	}

	// Validate project state
	data, err := os.ReadFile(absPath)
	if err != nil {
		result.Add(relPath, "project-state", fmt.Errorf("failed to read file: %w", err))
		return
	}

	if err := validator.ValidateProjectState(data); err != nil {
		result.Add(relPath, "project-state", err)
	}

	// Validate all tasks
	validateTasks(repoRoot, result, validator)
}

// validateTasks validates all task state files.
func validateTasks(repoRoot string, result *ValidationResult, validator *schemas.CUEValidator) {
	relTasksDir := filepath.Join(".sow", "project", "phases", "implementation", "tasks")
	absTasksDir := filepath.Join(repoRoot, relTasksDir)

	// Check if tasks directory exists
	_, err := os.Stat(absTasksDir)
	if err != nil {
		// No tasks yet - not an error
		return
	}

	// Read directory entries
	entries, err := os.ReadDir(absTasksDir)
	if err != nil {
		result.Add(relTasksDir, "tasks", fmt.Errorf("failed to read directory: %w", err))
		return
	}

	// Validate each task directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		relTaskStatePath := filepath.Join(relTasksDir, taskID, "state.yaml")
		absTaskStatePath := filepath.Join(absTasksDir, taskID, "state.yaml")

		// Check if state file exists
		_, err := os.Stat(absTaskStatePath)
		if err != nil {
			result.Add(relTaskStatePath, "task-state", fmt.Errorf("task directory exists but state.yaml is missing"))
			continue
		}

		// Read state file
		data, err := os.ReadFile(absTaskStatePath)
		if err != nil {
			result.Add(relTaskStatePath, "task-state", fmt.Errorf("failed to read file: %w", err))
			continue
		}

		// Validate against schema
		if err := validator.ValidateTaskState(data); err != nil {
			result.Add(relTaskStatePath, "task-state", err)
		}
	}
}
