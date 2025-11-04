package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Standard project states for the 5-phase workflow.

const (
	// NoProject indicates no active project (initial and final state).
	NoProject = state.State("NoProject")

	// PlanningActive indicates planning phase in progress.
	PlanningActive = state.State("PlanningActive")

	// ImplementationPlanning indicates implementation planning step.
	ImplementationPlanning = state.State("ImplementationPlanning")

	// ImplementationExecuting indicates task execution.
	ImplementationExecuting = state.State("ImplementationExecuting")

	// ReviewActive indicates review phase in progress.
	ReviewActive = state.State("ReviewActive")

	// FinalizeDocumentation indicates documentation update step.
	FinalizeDocumentation = state.State("FinalizeDocumentation")

	// FinalizeChecks indicates final validation checks.
	FinalizeChecks = state.State("FinalizeChecks")

	// FinalizeDelete indicates project cleanup step.
	FinalizeDelete = state.State("FinalizeDelete")
)
