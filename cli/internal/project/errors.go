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

	// ErrUnexpectedState indicates the phase is in a state not handled by Advance() logic.
	// This typically indicates state file corruption or a programming error.
	ErrUnexpectedState = errors.New("unexpected state for advance")

	// ErrCannotAdvance indicates the transition prerequisites are not met.
	// This occurs when guards fail, meaning the user called advance prematurely.
	ErrCannotAdvance = errors.New("cannot advance: prerequisites not met")

	// Note: State save errors wrap underlying I/O errors, not sentinel errors.
)
