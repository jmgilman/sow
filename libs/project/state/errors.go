package state

import "errors"

var (
	// ErrNotFound indicates the project state does not exist in storage.
	ErrNotFound = errors.New("project state not found")

	// ErrInvalidState indicates the stored state is invalid or corrupted.
	ErrInvalidState = errors.New("invalid project state")

	// ErrValidationFailed indicates project state validation failed.
	ErrValidationFailed = errors.New("validation failed")

	// ErrInvalidArtifactType indicates an artifact type is not allowed.
	ErrInvalidArtifactType = errors.New("invalid artifact type")
)
