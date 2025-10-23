package project

import "errors"

var (
	// ErrNotSupported indicates an operation is not supported by this phase.
	ErrNotSupported = errors.New("operation not supported by this phase")

	// ErrPhaseNotFound indicates a phase with the given name doesn't exist.
	ErrPhaseNotFound = errors.New("phase not found")

	// ErrNoProject indicates no project exists.
	ErrNoProject = errors.New("no project exists")

	// ErrProjectExists indicates a project already exists.
	ErrProjectExists = errors.New("project already exists")

	// ErrInvalidPhase indicates an invalid phase name.
	ErrInvalidPhase = errors.New("invalid phase")

	// ErrPhaseNotEnabled indicates the phase is not enabled.
	ErrPhaseNotEnabled = errors.New("phase not enabled")

	// ErrNoTask indicates no task with that ID exists.
	ErrNoTask = errors.New("task not found")

	// ErrArtifactNotFound indicates no artifact with that path exists.
	ErrArtifactNotFound = errors.New("artifact not found")

	// ErrInvalidMetadata indicates invalid metadata format.
	ErrInvalidMetadata = errors.New("invalid metadata")
)
