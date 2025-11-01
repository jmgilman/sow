// ============================================================================
// WARNING: This file has a corresponding CUE schema file (standard.cue).
// When modifying these Go types, you MUST manually update standard.cue
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
// These types must be kept in sync with the CUE schemas in standard.cue.
package projects

import (
	"time"

	p "github.com/jmgilman/sow/cli/schemas/phases"
)

// StandardProjectState defines the schema for a standard project type.
//
// This project type follows the 4-phase model:
// Planning → Implementation → Review → Finalize.
type StandardProjectState struct {
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

		// Optional GitHub issue number this project is linked to
		Github_issue *int64 `json:"github_issue,omitempty"`

		// ISO 8601 timestamps
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
	} `json:"project"`

	// 4-phase structure (composing reusable phase definitions)
	// Status is constrained to GenericStatus via unification
	Phases struct {
		// Phase 1: Planning (required, human-led)
		Planning PlanningPhase `json:"planning"`

		// Phase 2: Implementation (required, AI-autonomous)
		Implementation ImplementationPhase `json:"implementation"`

		// Phase 3: Review (required, AI-autonomous)
		Review ReviewPhase `json:"review"`

		// Phase 4: Finalize (required, AI-autonomous)
		Finalize FinalizePhase `json:"finalize"`
	} `json:"phases"`
}

// PlanningPhase is the planning phase for standard projects.
type PlanningPhase struct {
	p.Phase

	// Note: Planning phase uses only common Phase fields
}

// ImplementationPhase is the implementation phase for standard projects.
type ImplementationPhase struct {
	p.Phase

	// Whether task list has been approved by human
	Tasks_approved *bool `json:"tasks_approved,omitempty"`
}

// ReviewPhase is the review phase for standard projects.
type ReviewPhase struct {
	p.Phase

	// Current iteration number (increments on fail → reimplementation)
	Iteration *int `json:"iteration,omitempty"`
}

// FinalizePhase is the finalize phase for standard projects.
type FinalizePhase struct {
	p.Phase

	// Whether project directory has been deleted
	Project_deleted *bool `json:"project_deleted,omitempty"`

	// URL of created pull request
	Pr_url *string `json:"pr_url,omitempty"`

	// Documentation updates made
	Documentation_updates []string `json:"documentation_updates,omitempty"`
}

// ProjectState is the root discriminated union for all project types.
// The discriminator field is project.type.
//
// Valid project types:
//   - StandardProjectState (project.type == "standard")
//   - ExplorationProjectState (project.type == "exploration")
//   - DesignProjectState (project.type == "design")
//   - BreakdownProjectState (project.type == "breakdown")
//
// Note: Go does not support true discriminated unions like CUE does.
// This type alias points to StandardProjectState for backward compatibility.
// Code that needs to handle all project types should use type switches on the
// project.type field and cast appropriately.
type ProjectState = StandardProjectState
