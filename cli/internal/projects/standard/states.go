package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Standard project states for the 5-phase workflow.

const (
	// NoProject indicates no active project (initial and final state).
	NoProject = state.State("NoProject")

	// ImplementationPlanning indicates planning and task breakdown creation.
	ImplementationPlanning = state.State("ImplementationPlanning")

	// ImplementationExecuting indicates task execution.
	ImplementationExecuting = state.State("ImplementationExecuting")

	// ReviewActive indicates review phase in progress.
	ReviewActive = state.State("ReviewActive")

	// FinalizeChecks indicates final validation checks.
	FinalizeChecks = state.State("FinalizeChecks")

	// FinalizePRCreation indicates PR creation and approval step.
	FinalizePRCreation = state.State("FinalizePRCreation")

	// FinalizeCleanup indicates project cleanup step.
	FinalizeCleanup = state.State("FinalizeCleanup")
)
