package projects

import (
	"time"

	p "github.com/jmgilman/sow/cli/schemas/phases"
)

// ============================================================================
// WARNING: This file has a corresponding hand-written Go type file.
// When modifying this schema, you MUST manually update projects/standard.go
// to keep the Go types in sync. Do not rely on code generation.
//
// WHY HAND-WRITTEN:
// CUE's gengotypes tool cannot properly handle the type unification patterns
// used below (e.g., `p.#Phase & {status: p.#GenericStatus}`). It generates
// inline anonymous structs instead of preserving type references, which breaks
// the composition pattern. Therefore, the Go types are manually maintained.
// ============================================================================

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
	// Status is constrained to GenericStatus via unification
	phases: {
		// Phase 1: Planning (required, human-led)
		planning: p.#Phase & {
			status: p.#GenericStatus
		}

		// Phase 2: Implementation (required, AI-autonomous)
		implementation: p.#Phase & {
			status: p.#GenericStatus

			// Whether task list has been approved by human
			tasks_approved?: bool @go(,optional=nillable)
		}

		// Phase 3: Review (required, AI-autonomous)
		review: p.#Phase & {
			status: p.#GenericStatus

			// Current iteration number (increments on fail → reimplementation)
			iteration?: int @go(,optional=nillable)
		}

		// Phase 4: Finalize (required, AI-autonomous)
		finalize: p.#Phase & {
			status: p.#GenericStatus

			// Whether project directory has been deleted
			project_deleted?: bool @go(,optional=nillable)

			// URL of created pull request
			pr_url?: string @go(,optional=nillable)

			// Documentation updates made
			documentation_updates?: [...string] @go(,optional=nillable)
		}
	}
}

// ProjectState is the root discriminated union for all project types.
// The discriminator field is project.type.
#ProjectState: #StandardProjectState | #ExplorationProjectState | #DesignProjectState | #BreakdownProjectState
