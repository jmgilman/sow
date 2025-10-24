package projects

import (
	"time"

	p "github.com/jmgilman/sow/cli/schemas/phases"
)

// StandardProjectState defines the schema for a standard project type.
//
// This project type follows the 4-phase model:
// Planning → Implementation → Review → Finalize
#StandardProjectState: {
	// Statechart metadata (tracks state machine position)
	statechart: {
		// Current state in the lifecycle state machine
		current_state: "NoProject" | "PlanningActive" | "ImplementationPlanning" | "ImplementationExecuting" | "ReviewActive" | "FinalizeDocumentation" | "FinalizeChecks" | "FinalizeDelete"
	}

	// Project metadata
	project: {
		// Project type identifier
		type: "standard"

		// Kebab-case project identifier
		name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

		// Git branch name this project belongs to
		branch: string & !=""

		// Human-readable project description
		description: string

		// Optional GitHub issue number this project is linked to
		github_issue?: (int & >0) @go(,optional=nillable)

		// ISO 8601 timestamps
		created_at: time.Time
		updated_at: time.Time
	}

	// 4-phase structure (composing reusable phase definitions)
	phases: {
		// Phase 1: Planning (required, human-led)
		planning: p.#Phase

		// Phase 2: Implementation (required, AI-autonomous)
		implementation: p.#Phase

		// Phase 3: Review (required, AI-autonomous)
		review: p.#Phase

		// Phase 4: Finalize (required, AI-autonomous)
		finalize: p.#Phase
	}
}

// ProjectState is the root discriminated union for all project types.
// For MVP, only StandardProjectState exists. Future project types
// (DesignProjectState, SpikeProjectState, etc.) will be added here.
#ProjectState: #StandardProjectState
