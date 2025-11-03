package project

import "time"

// PhaseState represents the state of a single phase within a project.
// Phases are major divisions of work (e.g., planning, implementation, review).
// Each phase has status tracking, timestamps, optional metadata, and collections
// of artifacts and tasks.
#PhaseState: {
	// status indicates the current status of the phase.
	// Must be non-empty. Common values: "pending", "in_progress", "completed"
	// Actual valid statuses are defined by the project type.
	status: string & !=""

	// enabled indicates whether this phase is enabled for the current project.
	// Some project types may have optional phases that can be skipped.
	enabled: bool

	// created_at is the timestamp when this phase was created.
	created_at: time.Time

	// started_at is the optional timestamp when work on this phase began.
	// Only set when the phase transitions to an active state.
	started_at?: time.Time

	// completed_at is the optional timestamp when this phase finished.
	// Only set when the phase transitions to a completed state.
	completed_at?: time.Time

	// metadata holds project-type-specific data for this phase.
	// Structure and validation are defined by the project type.
	// Example uses: complexity rating, approval flags, custom settings.
	metadata?: {...}

	// inputs is the list of artifacts that this phase consumes.
	// These artifacts are typically outputs from previous phases.
	inputs: [...#ArtifactState]

	// outputs is the list of artifacts that this phase produces.
	// These become available for use by subsequent phases.
	outputs: [...#ArtifactState]

	// tasks is the list of tasks belonging to this phase.
	// Each task represents a discrete unit of work to be completed.
	tasks: [...#TaskState]
}
