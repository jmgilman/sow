package schemas

import "time"

// ProjectState defines the schema for .sow/project/state.yaml
//
// This file tracks the state of an active project across all 5 phases.
// All projects have the same 5 phases; the 'enabled' flag controls
// which phases actually execute.
#ProjectState: {
	// Statechart metadata (tracks state machine position)
	statechart: {
		// Current state in the lifecycle state machine
		current_state: "NoProject" | "DiscoveryDecision" | "DiscoveryActive" | "DesignDecision" | "DesignActive" | "ImplementationPlanning" | "ImplementationExecuting" | "ReviewActive" | "FinalizeDocumentation" | "FinalizeChecks" | "FinalizeDelete"
	}

	// Project metadata
	project: {
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

	// 5-phase structure (fixed phases, enabled flag controls execution)
	phases: {
		// Phase 1: Discovery (optional, human-led)
		discovery: #DiscoveryPhase

		// Phase 2: Design (optional, human-led)
		design: #DesignPhase

		// Phase 3: Implementation (required, AI-autonomous)
		implementation: #ImplementationPhase

		// Phase 4: Review (required, AI-autonomous)
		review: #ReviewPhase

		// Phase 5: Finalize (required, AI-autonomous)
		finalize: #FinalizePhase
	}
}

// Phase represents common phase fields
#Phase: {
	// Phase execution status
	status: "skipped" | "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:   time.Time
	started_at:   *null | time.Time
	completed_at: *null | time.Time
}

// DiscoveryPhase represents the discovery phase
#DiscoveryPhase: {
	#Phase

	// Can be disabled
	enabled: bool

	// Discovery type categorization
	discovery_type: *null | "bug" | "feature" | "docs" | "refactor" | "general"

	// Discovery artifacts requiring approval
	artifacts: [...#Artifact]
}

// DesignPhase represents the design phase
#DesignPhase: {
	#Phase

	// Can be disabled
	enabled: bool

	// Whether architect agent was used
	architect_used: *null | bool

	// Design artifacts requiring approval (ADRs, design docs)
	artifacts: [...#Artifact]
}

// ImplementationPhase represents the implementation phase
#ImplementationPhase: {
	#Phase

	// Always enabled
	enabled: true

	// Whether planner agent was used
	planner_used: *null | bool

	// Approved task list (gap-numbered)
	tasks: [...#Task]

	// Human approval of task plan before autonomous execution
	tasks_approved: bool

	// Tasks awaiting human approval before execution
	pending_task_additions: *null | [...#Task]
}

// ReviewPhase represents the review phase
#ReviewPhase: {
	#Phase

	// Always enabled
	enabled: true

	// Current review iteration (increments on loop-back)
	iteration: int & >=1

	// Review reports (numbered 001, 002, 003...)
	reports: [...#ReviewReport]
}

// FinalizePhase represents the finalize phase
#FinalizePhase: {
	#Phase

	// Always enabled
	enabled: true

	// Documentation files updated
	documentation_updates: *null | [...string]

	// Design artifacts moved to knowledge (from→to pairs)
	artifacts_moved: *null | [...{
		from: string
		to:   string
	}]

	// Critical gate: must be true before phase completion
	project_deleted: bool

	// Pull request URL (created during finalize)
	pr_url: *null | string
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
	dependencies: *null | [...string]
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
