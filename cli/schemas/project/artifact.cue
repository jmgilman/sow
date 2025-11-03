package project

import "time"

// ArtifactState represents a file or document produced or consumed by a phase or task.
// Artifacts track important work products such as design documents, task lists,
// review reports, and other phase outputs. Each artifact has a type, path, approval
// status, and optional metadata.
#ArtifactState: {
	// type identifies the kind of artifact.
	// Must be non-empty. Examples: "task_list", "review", "design_doc", "adr"
	// The type determines how the artifact is interpreted and used.
	type: string & !=""

	// path is the relative path to the artifact file from .sow/project/.
	// Must be non-empty. Example: "phases/planning/task-breakdown.md"
	path: string & !=""

	// approved indicates whether this artifact has been reviewed and approved.
	// Used by state machines to gate transitions (e.g., can't proceed until artifact approved).
	approved: bool

	// created_at is the timestamp when this artifact was created or added.
	created_at: time.Time

	// metadata holds artifact-type-specific data.
	// Structure and validation are defined by the artifact type.
	// Example uses: review assessment (pass/fail), document category, custom properties.
	metadata?: {...}
}
