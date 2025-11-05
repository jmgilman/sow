package phases

import "time"

// GenericStatus defines the common phase status values that most project types use.
// Project types can use this for their phases or define custom status enums per phase.
#GenericStatus: "pending" | "in_progress" | "completed" | "skipped" | "failed"

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
	failed_at?:    time.Time @go(,optional=nillable)

	// Iteration tracking for phases that go through multiple cycles
	// Starts at 1, increments on re-entry after failure
	iteration?: int & >=1 @go(,optional=nillable)

	// Generic collections (used by phases that need them)
	artifacts: [...#Artifact] // Used by discovery, design, review
	tasks: [...#Task] // Used by implementation

	// Input sources that inform this phase's work (design, breakdown)
	inputs?: [...#Artifact] @go(,optional=nillable)

	// Escape hatch for unanticipated fields
	metadata?: {[string]: _} @go(,optional=nillable)
}

// Artifact represents a phase artifact requiring human approval
#Artifact: {
	// Path relative to .sow/project/
	path: string

	// Human approval status (optional for exploration artifacts)
	approved?: bool @go(,optional=nillable)

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

	// References to artifacts created for this task (exploration topics)
	refs?: [...#Artifact] @go(,optional=nillable)

	// Task-specific metadata (e.g., github_issue_url for breakdown tasks)
	metadata?: {[string]: _} @go(,optional=nillable)
}
