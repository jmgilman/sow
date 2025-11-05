// Package standard implements the standard project type using the SDK builder.
package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Standard project events trigger state transitions.

const (
	// EventProjectInit creates new project.
	// Transition: NoProject → ImplementationPlanning.
	EventProjectInit = state.Event("project_init")

	// EventPlanningComplete completes planning and task breakdown.
	// Transition: ImplementationPlanning → ImplementationExecuting.
	// Guard: all task description files approved.
	EventPlanningComplete = state.Event("planning_complete")

	// EventAllTasksComplete indicates all tasks done.
	// Transition: ImplementationExecuting → ReviewActive.
	// Guard: all tasks completed or abandoned.
	EventAllTasksComplete = state.Event("all_tasks_complete")

	// EventReviewPass passes review assessment.
	// Transition: ReviewActive → FinalizeChecks.
	// Guard: review artifact approved with assessment=pass.
	EventReviewPass = state.Event("review_pass")

	// EventReviewFail fails review (rework loop).
	// Transition: ReviewActive → ImplementationPlanning.
	// Guard: review artifact approved with assessment=fail.
	EventReviewFail = state.Event("review_fail")

	// EventChecksDone completes final checks.
	// Transition: FinalizeChecks → FinalizePRCreation.
	EventChecksDone = state.Event("checks_done")

	// EventPRCreated indicates PR created and approved.
	// Transition: FinalizePRCreation → FinalizeCleanup.
	// Guard: pr_body artifact approved.
	EventPRCreated = state.Event("pr_created")

	// EventCleanupComplete completes project cleanup.
	// Transition: FinalizeCleanup → NoProject.
	// Guard: project_deleted metadata flag.
	EventCleanupComplete = state.Event("cleanup_complete")
)
