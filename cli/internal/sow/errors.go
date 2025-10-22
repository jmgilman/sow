package sow

import "errors"

// Domain-specific errors for the sow system.
var (
	// ErrNoProject indicates no active project exists in the repository.
	ErrNoProject = errors.New("no active project exists")

	// ErrProjectExists indicates a project already exists when trying to create one.
	ErrProjectExists = errors.New("project already exists")

	// ErrNoTask indicates a task does not exist or cannot be found.
	ErrNoTask = errors.New("task not found")

	// ErrInvalidPhase indicates an invalid phase name was provided.
	ErrInvalidPhase = errors.New("invalid phase name")

	// ErrPhaseNotEnabled indicates an operation was attempted on a disabled phase.
	ErrPhaseNotEnabled = errors.New("phase not enabled")

	// ErrInvalidStatus indicates an invalid task status was provided.
	ErrInvalidStatus = errors.New("invalid task status")

	// ErrNotInitialized indicates sow has not been initialized in the repository.
	ErrNotInitialized = errors.New("sow not initialized (run 'sow init')")
)
