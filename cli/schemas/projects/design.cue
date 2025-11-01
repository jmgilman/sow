package projects

import (
	"time"

	p "github.com/jmgilman/sow/cli/schemas/phases"
)

// ============================================================================
// WARNING: This file has a corresponding hand-written Go type file.
// When modifying this schema, you MUST manually update projects/design.go
// to keep the Go types in sync. Do not rely on code generation.
//
// WHY HAND-WRITTEN:
// CUE's gengotypes tool cannot properly handle the type unification patterns
// used below (e.g., `p.#Phase & {status: p.#GenericStatus}`). It generates
// inline anonymous structs instead of preserving type references, which breaks
// the composition pattern. Therefore, the Go types are manually maintained.
// ============================================================================

// DesignProjectState defines the schema for a design project type.
//
// This project type follows the 2-phase model:
// Design → Finalization
#DesignProjectState: {
	// Statechart metadata (tracks state machine position)
	statechart: {
		// Current state in the lifecycle state machine
		current_state: "DesignActive" | "DesignReview" | "FinalizationActive" | "Completed"
	}

	// Project metadata
	project: {
		// Project type identifier
		type: "design"

		// Kebab-case project identifier
		name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

		// Git branch name this project belongs to
		branch: string & !=""

		// Human-readable project description
		description: string

		// ISO 8601 timestamps
		created_at: time.Time
		updated_at: time.Time
	}

	// 2-phase structure (composing reusable phase definitions)
	phases: {
		// Phase 1: Design
		design: p.#Phase & {
			status: "active" | "review" | "completed"
		}

		// Phase 2: Finalization
		finalization: p.#Phase & {
			status: p.#GenericStatus
		}
	}
}
