// Package standard implements the standard project type using the SDK builder.
package standard

import (
	"github.com/jmgilman/sow/libs/project"
)

// Standard project events trigger state transitions.

const (
	// EventProjectInit creates new project.
	// Transition: NoProject → ImplementationPlanning.
	EventProjectInit = project.Event("project_init")

	// EventPlanningComplete completes planning and task breakdown.
	// Transition: ImplementationPlanning → ImplementationDraftPRCreation.
	// Guard: all task description files approved.
	EventPlanningComplete = project.Event("planning_complete")

	// EventDraftPRCreated indicates draft PR has been created.
	// Transition: ImplementationDraftPRCreation → ImplementationExecuting.
	// Guard: draft PR created and metadata stored.
	EventDraftPRCreated = project.Event("draft_pr_created")

	// EventAllTasksComplete indicates all tasks done.
	// Transition: ImplementationExecuting → ReviewActive.
	// Guard: all tasks completed or abandoned.
	EventAllTasksComplete = project.Event("all_tasks_complete")

	// EventReviewPass passes review assessment.
	// Transition: ReviewActive → FinalizeChecks.
	// Guard: review artifact approved with assessment=pass.
	EventReviewPass = project.Event("review_pass")

	// EventReviewFail fails review (rework loop).
	// Transition: ReviewActive → ImplementationPlanning.
	// Guard: review artifact approved with assessment=fail.
	EventReviewFail = project.Event("review_fail")

	// EventChecksDone completes final checks.
	// Transition: FinalizeChecks → FinalizePRReady.
	EventChecksDone = project.Event("checks_done")

	// EventPRReady indicates PR body updated and marked ready for review.
	// Transition: FinalizePRReady → FinalizePRChecks.
	// Guard: pr_body artifact approved.
	EventPRReady = project.Event("pr_ready")

	// EventPRChecksPass indicates all PR checks have passed.
	// Transition: FinalizePRChecks → FinalizeCleanup.
	// Guard: pr_checks_passed metadata flag.
	EventPRChecksPass = project.Event("pr_checks_pass")

	// EventCleanupComplete completes project cleanup.
	// Transition: FinalizeCleanup → NoProject.
	// Guard: project_deleted metadata flag.
	EventCleanupComplete = project.Event("cleanup_complete")
)
