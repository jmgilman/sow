package standard

import (
	"github.com/jmgilman/sow/libs/project"
)

// Standard project states for the 5-phase workflow.

const (
	// NoProject indicates no active project (initial and final state).
	NoProject = project.State("NoProject")

	// ImplementationPlanning indicates planning and task breakdown creation.
	ImplementationPlanning = project.State("ImplementationPlanning")

	// ImplementationDraftPRCreation indicates draft PR creation before task execution.
	ImplementationDraftPRCreation = project.State("ImplementationDraftPRCreation")

	// ImplementationExecuting indicates task execution.
	ImplementationExecuting = project.State("ImplementationExecuting")

	// ReviewActive indicates review phase in progress.
	ReviewActive = project.State("ReviewActive")

	// FinalizeChecks indicates final validation checks.
	FinalizeChecks = project.State("FinalizeChecks")

	// FinalizePRReady indicates updating PR body and marking ready for review.
	FinalizePRReady = project.State("FinalizePRReady")

	// FinalizePRChecks indicates PR checks monitoring and fixing.
	FinalizePRChecks = project.State("FinalizePRChecks")

	// FinalizeCleanup indicates project cleanup step.
	FinalizeCleanup = project.State("FinalizeCleanup")
)
