package phases

import "time"

// Phase represents common phase fields
#Phase: {
	// Phase execution status
	status: "skipped" | "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:    time.Time
	started_at?:   time.Time @go(,optional=nillable)
	completed_at?: time.Time @go(,optional=nillable)
}

// Artifact represents an artifact requiring human approval
#Artifact: {
	// Path relative to .sow/project/
	path: string

	// Human approval status
	approved: bool

	// When artifact was created
	created_at: time.Time
}

// Task represents an implementation task
#Task: {
	// Gap-numbered ID (010, 020, 030...)
	id: string & =~"^[0-9]{3,}$"

	// Task name
	name: string & !=""

	// Task status
	status: "pending" | "in_progress" | "completed" | "abandoned"

	// Can run in parallel with other tasks
	parallel: bool

	// Task IDs this task depends on
	dependencies?: [...string] @go(,optional=nillable)
}

// ReviewReport represents a review iteration report
#ReviewReport: {
	// Report ID (001, 002, 003...)
	id: string & =~"^[0-9]{3}$"

	// Path relative to .sow/project/phases/review/
	path: string

	// When report was created
	created_at: time.Time

	// Review assessment
	assessment: "pass" | "fail"

	// Human approval of orchestrator's review
	approved: bool
}
