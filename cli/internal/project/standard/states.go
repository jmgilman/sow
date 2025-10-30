package standard

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// Standard project states - these are specific to the standard 5-phase project workflow.
// The NoProject state remains in the statechart package as it's shared across all project types.

const (
	// PlanningActive indicates the planning phase is in progress (subservient mode).
	// The orchestrator gathers context, confirms requirements, creates task list, and gets approval.
	PlanningActive = statechart.State("PlanningActive")

	// ImplementationPlanning indicates the implementation phase planning step.
	// The orchestrator (or planner agent) creates the task breakdown.
	ImplementationPlanning = statechart.State("ImplementationPlanning")

	// ImplementationExecuting indicates tasks are being executed (autonomous mode).
	// The orchestrator spawns implementer agents to work on tasks.
	ImplementationExecuting = statechart.State("ImplementationExecuting")

	// ReviewActive indicates the review phase is in progress (autonomous mode).
	// The orchestrator or reviewer agent validates the implementation.
	ReviewActive = statechart.State("ReviewActive")

	// FinalizeDocumentation indicates the documentation update step of finalization.
	// The orchestrator checks if documentation needs updates and performs them.
	FinalizeDocumentation = statechart.State("FinalizeDocumentation")

	// FinalizeChecks indicates the final checks step of finalization.
	// The orchestrator runs tests, linters, and other validation checks.
	FinalizeChecks = statechart.State("FinalizeChecks")

	// FinalizeDelete indicates the project deletion step of finalization.
	// The orchestrator must delete the project folder before completing.
	FinalizeDelete = statechart.State("FinalizeDelete")
)
