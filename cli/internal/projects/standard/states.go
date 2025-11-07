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

	// ImplementationDraftPRCreation indicates draft PR creation before task execution.
	ImplementationDraftPRCreation = state.State("ImplementationDraftPRCreation")

	// ImplementationExecuting indicates task execution.
	ImplementationExecuting = state.State("ImplementationExecuting")

	// ReviewActive indicates review phase in progress.
	ReviewActive = state.State("ReviewActive")

	// FinalizeChecks indicates final validation checks.
	FinalizeChecks = state.State("FinalizeChecks")

	// FinalizePRReady indicates updating PR body and marking ready for review.
	FinalizePRReady = state.State("FinalizePRReady")

	// FinalizePRChecks indicates PR checks monitoring and fixing.
	FinalizePRChecks = state.State("FinalizePRChecks")

	// FinalizeCleanup indicates project cleanup step.
	FinalizeCleanup = state.State("FinalizeCleanup")
)
