// Project State Schema
// Location: .sow/project/state.yaml
//
// Central planning document containing all project metadata, phases, and tasks.
// This schema validates project state files and enforces required fields and constraints.

package sow

import "time"

// ProjectState defines the complete project structure
#ProjectState: {
	project: #Project
	phases:  [...#Phase]
}

// Project metadata and complexity assessment
#Project: {
	// Project identifier (kebab-case recommended)
	name: string & =~"^[a-z0-9]+(-[a-z0-9]+)*$"

	// Git branch this project belongs to
	branch: string & =~"^[a-zA-Z0-9/_-]+$"

	// Timestamps (ISO 8601 format)
	created_at: string & time.Format(time.RFC3339)
	updated_at: string & time.Format(time.RFC3339)

	// Human-readable description
	description: string

	// Complexity assessment from initial planning
	complexity: {
		// Complexity rating: 1=simple, 2=moderate, 3=complex
		rating: int & >=1 & <=3

		metrics: {
			// Estimated number of files affected
			estimated_files: int & >=0

			// Cross-cutting concerns exist
			cross_cutting: bool

			// New dependencies required
			new_dependencies: bool
		}
	}

	// Name of currently active phase (must exist in phases array)
	active_phase: string & =~"^(discovery|design|implement|test|review|deploy|document)$"
}

// Phase within a project
#Phase: {
	// Phase name (standard lifecycle phase)
	name: string & =~"^(discovery|design|implement|test|review|deploy|document)$"

	// Phase status
	status: "pending" | "in_progress" | "completed"

	// When phase was created
	created_at: string & time.Format(time.RFC3339)

	// When phase completed (null if not done)
	completed_at: null | (string & time.Format(time.RFC3339))

	// Tasks within this phase
	tasks: [...#TaskSummary]
}

// Task summary (lightweight version for project state)
#TaskSummary: {
	// Gap-numbered ID (e.g., "010", "020", "030")
	id: string & =~"^[0-9]{3}$"

	// Task name/description
	name: string

	// Task status
	status: "pending" | "in_progress" | "completed" | "abandoned"

	// Can run in parallel with other parallel tasks
	parallel: bool

	// Assigned agent role
	assigned_agent: "architect" | "implementer" | "integration-tester" | "reviewer" | "documenter"
}
