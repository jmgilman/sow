package state

import (
	"github.com/jmgilman/sow/libs/schemas/project"
)

// validateStructure validates the project state structure against CUE schemas.
// This is a placeholder that will be implemented in Task 080.
// For now, it performs basic validation of required fields.
func validateStructure(state *project.ProjectState) error {
	// Basic validation of required fields
	if state.Name == "" {
		return &ValidationError{Field: "name", Message: "required"}
	}
	if state.Type == "" {
		return &ValidationError{Field: "type", Message: "required"}
	}
	if state.Branch == "" {
		return &ValidationError{Field: "branch", Message: "required"}
	}

	// CUE validation will be added in Task 080
	return nil
}

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
