// ============================================================================
// WARNING: This file has a corresponding CUE schema file (exploration.cue).
// When modifying these Go types, you MUST manually update exploration.cue
// to keep the schemas in sync. Do not rely on code generation.
//
// WHY HAND-WRITTEN:
// CUE's gengotypes tool cannot properly handle the type unification patterns
// used in project schemas (e.g., `p.#Phase & {status: p.#GenericStatus}`).
// It generates inline anonymous structs instead of preserving type references,
// which breaks the composition pattern and causes type compatibility issues.
// Therefore, these types must be manually maintained to match the CUE schema.
// ============================================================================

// Package projects provides hand-written Go types for project schemas.
// These types must be kept in sync with the CUE schemas.
package projects

import (
	"time"

	p "github.com/jmgilman/sow/cli/schemas/phases"
)

// ExplorationProjectState defines the schema for an exploration project type.
//
// This project type follows the 2-phase model:
// Exploration → Finalization.
type ExplorationProjectState struct {
	// Statechart metadata (tracks state machine position)
	Statechart struct {
		// Current state in the lifecycle state machine
		Current_state string `json:"current_state"`
	} `json:"statechart"`

	// Project metadata
	Project struct {
		// Project type identifier
		Type string `json:"type"`

		// Kebab-case project identifier
		Name string `json:"name"`

		// Git branch name this project belongs to
		Branch string `json:"branch"`

		// Human-readable project description
		Description string `json:"description"`

		// ISO 8601 timestamps
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
	} `json:"project"`

	// 2-phase structure (composing reusable phase definitions)
	Phases struct {
		// Phase 1: Exploration
		Exploration ExplorationPhase `json:"exploration"`

		// Phase 2: Finalization
		Finalization ExplorationFinalizationPhase `json:"finalization"`
	} `json:"phases"`
}

// ExplorationPhase is the exploration phase for exploration projects.
type ExplorationPhase struct {
	p.Phase

	// Note: Exploration phase uses common Phase fields plus status constraint
}

// ExplorationFinalizationPhase is the finalization phase for exploration projects.
type ExplorationFinalizationPhase struct {
	p.Phase

	// Note: Finalization phase uses common Phase fields
}
