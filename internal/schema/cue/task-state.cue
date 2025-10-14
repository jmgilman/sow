// Task State Schema
// Location: .sow/project/phases/<phase>/tasks/<id>/state.yaml
//
// Task-specific metadata including iteration tracking, context references,
// and feedback status. Each task has its own state file in its task directory.

package sow

import "time"

// TaskState defines the complete task metadata structure
#TaskState: {
	task: #Task
}

// Task metadata and tracking
#Task: {
	// Task ID (matches directory name, gap-numbered like "010", "020")
	id: string & =~"^[0-9]{3}$"

	// Task name/description
	name: string

	// Phase this task belongs to
	phase: string & =~"^(discovery|design|implement|test|review|deploy|document)$"

	// Task status
	status: "pending" | "in_progress" | "completed" | "abandoned"

	// Timestamps (ISO 8601 format)
	created_at: string & time.Format(time.RFC3339)

	// When work first began (null if not started)
	started_at: null | (string & time.Format(time.RFC3339))

	// Last modification timestamp
	updated_at: string & time.Format(time.RFC3339)

	// When task completed (null if not done)
	completed_at: null | (string & time.Format(time.RFC3339))

	// Attempt counter (managed by orchestrator, incremented on retry)
	// Must be >= 1
	iteration: int & >=1

	// Assigned agent role
	assigned_agent: "architect" | "implementer" | "integration-tester" | "reviewer" | "documenter"

	// Context references (file paths relative to .sow/ root)
	// Orchestrator compiles this list; worker reads all referenced files
	references: [...string]

	// Feedback tracking
	feedback: [...#Feedback]

	// Files modified during task execution (auto-populated by worker)
	files_modified: [...string]
}

// Feedback item
#Feedback: {
	// Feedback ID (sequential like "001", "002")
	id: string & =~"^[0-9]{3}$"

	// When feedback was created
	created_at: string & time.Format(time.RFC3339)

	// Feedback status
	status: "pending" | "addressed" | "superseded"
}
