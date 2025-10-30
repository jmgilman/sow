package phases

import "time"

// GenericStatus defines the common phase status values that most project types use.
// Project types can use this for their phases or define custom status enums per phase.
#GenericStatus: "pending" | "in_progress" | "completed" | "skipped"

// Phase is the universal schema for all phases in all project types.
// What makes a phase unique is its guards, prompts, and which operations it supports.
// NOTE: Project types will typically define their own phase types with additional fields.
#Phase: {
	// Common metadata
	status:  string
	enabled: bool

	// Timestamps
	created_at:    time.Time
	started_at?:   time.Time @go(,optional=nillable)
	completed_at?: time.Time @go(,optional=nillable)

	// Generic collections (used by phases that need them)
	artifacts: [...#Artifact] // Used by discovery, design, review
	tasks: [...#Task] // Used by implementation

	// Escape hatch for unanticipated fields
	metadata?: {[string]: _} @go(,optional=nillable)
}

// Artifact represents a phase artifact requiring human approval
#Artifact: {
	// Path relative to .sow/project/
	path: string

	// Human approval status
	approved: bool

	// When artifact was created
	created_at: time.Time

	// Artifact type (e.g., "task_list", "review", "documentation")
	type?: string @go(,optional=nillable)

	// Review assessment result ("pass" or "fail")
	assessment?: string @go(,optional=nillable)

	// Escape hatch for unanticipated fields
	metadata?: {[string]: _} @go(,optional=nillable)
}

// Task represents an implementation task
#Task: {
	// Gap-numbered ID (010, 020, 030...)
	id: string & =~"^[0-9]{3,}$"

	// Task name
	name: string & !=""

	// Task status
	status: "pending" | "in_progress" | "needs_review" | "completed" | "abandoned"

	// Can run in parallel with other tasks
	parallel: bool

	// Task IDs this task depends on
	dependencies?: [...string] @go(,optional=nillable)
}
