package sowfs

import (
	"fmt"
	"strings"
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

// AddError appends a pre-constructed validation error.
func (r *ValidationResult) AddError(verr ValidationError) {
	r.Errors = append(r.Errors, verr)
}

// Merge combines errors from another ValidationResult into this one.
func (r *ValidationResult) Merge(other *ValidationResult) {
	if other != nil {
		r.Errors = append(r.Errors, other.Errors...)
	}
}
