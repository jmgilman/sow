package project

import "fmt"

// ErrNoDeterminer indicates no event determiner is configured for a state.
type ErrNoDeterminer struct {
	State State
}

// Error returns the error message.
func (e *ErrNoDeterminer) Error() string {
	return fmt.Sprintf("no event determiner configured for state %s", e.State)
}

// ErrBranchNotFound indicates a branch discriminator returned an unrecognized value.
type ErrBranchNotFound struct {
	State State
	Value string
}

// Error returns the error message.
func (e *ErrBranchNotFound) Error() string {
	return fmt.Sprintf("no branch path found for value %q in state %s", e.Value, e.State)
}

// ErrTransitionFailed indicates a state machine transition failed.
type ErrTransitionFailed struct {
	Cause error
}

// Error returns the error message.
func (e *ErrTransitionFailed) Error() string {
	return fmt.Sprintf("transition failed: %v", e.Cause)
}

// Unwrap returns the underlying error.
func (e *ErrTransitionFailed) Unwrap() error {
	return e.Cause
}

// ErrPhaseStatusUpdate indicates a phase status update failed.
type ErrPhaseStatusUpdate struct {
	Phase     string
	Operation string
	Cause     error
}

// Error returns the error message.
func (e *ErrPhaseStatusUpdate) Error() string {
	return fmt.Sprintf("failed to %s phase %s: %v", e.Operation, e.Phase, e.Cause)
}

// Unwrap returns the underlying error.
func (e *ErrPhaseStatusUpdate) Unwrap() error {
	return e.Cause
}

// ErrInvalidOutputType indicates an output artifact has a type not allowed for the phase.
type ErrInvalidOutputType struct {
	Phase        string
	ArtifactType string
	AllowedTypes []string
}

// Error returns the error message.
func (e *ErrInvalidOutputType) Error() string {
	return fmt.Sprintf("phase %q does not allow output type %q (allowed: %v)", e.Phase, e.ArtifactType, e.AllowedTypes)
}

// ErrInvalidInputType indicates an input artifact has a type not allowed for the phase.
type ErrInvalidInputType struct {
	Phase        string
	ArtifactType string
	AllowedTypes []string
}

// Error returns the error message.
func (e *ErrInvalidInputType) Error() string {
	return fmt.Sprintf("phase %q does not allow input type %q (allowed: %v)", e.Phase, e.ArtifactType, e.AllowedTypes)
}
