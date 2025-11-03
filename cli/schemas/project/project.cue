package project

import "time"

// ProjectState represents the complete state of a project in the sow system.
// It includes project identification, timestamps, all phases, and state machine position.
// This is the root type for all project state files stored at .sow/project/state.yaml.
#ProjectState: {
	// name is the unique identifier for the project.
	// Must be lowercase alphanumeric with hyphens allowed (not at start/end).
	// Example: "my-project", "auth-implementation", "bug-fix-123"
	name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

	// type identifies the project type (e.g., "standard", "exploration", "design").
	// Must be lowercase alphanumeric with underscores allowed.
	// Each type defines its own phase lifecycle and behavior.
	type: string & =~"^[a-z0-9_]+$"

	// branch is the git branch associated with this project.
	// Must be non-empty. Example: "feat/add-authentication"
	branch: string & !=""

	// description is an optional human-readable description of the project.
	description?: string

	// created_at is the timestamp when the project was initialized.
	created_at: time.Time

	// updated_at is the timestamp of the last project state modification.
	updated_at: time.Time

	// phases is a map of phase names to phase state.
	// Keys are phase identifiers (e.g., "planning", "implementation", "review").
	// Values contain the state and collections for each phase.
	phases: [string]: #PhaseState

	// statechart tracks the current position in the project state machine.
	// This determines which operations are valid at any given time.
	statechart: #StatechartState
}

// StatechartState represents the current position in a project's state machine.
// It tracks which state the project is in and when it last transitioned.
#StatechartState: {
	// current_state is the name of the current state in the state machine.
	// Must be non-empty. Example: "PlanningActive", "ImplementationExecuting"
	current_state: string & !=""

	// updated_at is the timestamp of the last state transition.
	updated_at: time.Time
}
