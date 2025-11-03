package project

import "time"

// TaskState represents a discrete unit of work within a phase.
// Tasks are the atomic units of project execution, each with a unique ID,
// status tracking, iteration support for refinement cycles, and collections
// of input and output artifacts.
#TaskState: {
	// id is the unique three-digit identifier for this task.
	// Must match pattern: exactly 3 digits (e.g., "001", "010", "042", "999")
	// IDs are typically gap-numbered (010, 020, 030) to allow insertion.
	id: string & =~"^[0-9]{3}$"

	// name is the human-readable name of the task.
	// Must be non-empty. Example: "Implement JWT middleware", "Write unit tests"
	name: string & !=""

	// phase identifies which phase this task belongs to.
	// Must be non-empty. Example: "implementation", "testing", "documentation"
	phase: string & !=""

	// status indicates the current state of the task.
	// Must be one of: "pending" (not started), "in_progress" (actively working),
	// "completed" (successfully finished), "abandoned" (cancelled/obsolete).
	status: "pending" | "in_progress" | "completed" | "abandoned"

	// created_at is the timestamp when this task was created.
	created_at: time.Time

	// started_at is the optional timestamp when work on this task began.
	// Only set when the task transitions to "in_progress" status.
	started_at?: time.Time

	// updated_at is the timestamp of the last modification to this task.
	// Updated whenever task state changes (status, iteration, artifacts, etc.).
	updated_at: time.Time

	// completed_at is the optional timestamp when this task finished.
	// Only set when the task transitions to "completed" or "abandoned" status.
	completed_at?: time.Time

	// iteration is the current iteration number for this task.
	// Must be >= 1. Starts at 1, increments when task is sent back for revisions.
	// Supports refinement cycles (e.g., review → revise → review again).
	iteration: int & >=1

	// assigned_agent identifies the type of agent responsible for this task.
	// Must be non-empty. Examples: "implementer", "architect", "reviewer"
	// Determines which agent type is spawned to execute the task.
	assigned_agent: string & !=""

	// inputs is the list of artifacts that this task consumes.
	// These provide context and requirements for completing the task.
	// Examples: design documents, specifications, feedback files.
	inputs: [...#ArtifactState]

	// outputs is the list of artifacts that this task produces.
	// These represent the work product of completing the task.
	// Examples: code files, test files, documentation.
	outputs: [...#ArtifactState]

	// metadata holds task-specific data.
	// Structure and validation are defined by the project type.
	// Example uses: complexity rating, dependencies, custom properties.
	metadata?: {...}
}
