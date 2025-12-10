package state

import "errors"

var (
	// ErrNotFound indicates the project state does not exist in storage.
	ErrNotFound = errors.New("project state not found")

	// ErrInvalidState indicates the stored state is invalid or corrupted.
	ErrInvalidState = errors.New("invalid project state")
)
