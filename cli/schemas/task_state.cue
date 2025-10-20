package schemas

import "time"

// TaskState defines the schema for task state files at:
// .sow/project/phases/implementation/tasks/<id>/state.yaml
//
// This tracks metadata for individual implementation tasks.
#TaskState: {
	task: {
		// Task ID (matches directory name, gap-numbered)
		id: string & =~"^[0-9]{3,}$"

		// Task name
		name: string & !=""

		// Always "implementation" in the 5-phase model
		phase: "implementation"

		// Task execution status
		status: "pending" | "in_progress" | "completed" | "abandoned"

		// Timestamps
		created_at:   time.Time
		started_at:   *null | time.Time
		updated_at:   time.Time
		completed_at: *null | time.Time

		// Iteration counter (managed by orchestrator)
		// Used to construct agent ID: {assigned_agent}-{iteration}
		iteration: int & >=1

		// Agent assigned to this task (e.g., "implementer", "architect")
		// Used to construct agent ID: {assigned_agent}-{iteration}
		assigned_agent: string & !=""

		// Context file paths (relative to .sow/)
		// Compiled by orchestrator during context preparation
		references: [...string]

		// Human feedback/corrections
		feedback: [...#Feedback]

		// Files modified during task execution
		// Auto-populated by worker via 'sow log' command
		files_modified: [...string]
	}
}

// Feedback represents human corrections/guidance
#Feedback: {
	// Feedback ID (001, 002, 003...)
	id: string & =~"^[0-9]{3}$"

	// When feedback was provided
	created_at: time.Time

	// Feedback status
	status: "pending" | "addressed" | "superseded"
}
