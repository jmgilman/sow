// Package standard implements the standard project type using the SDK builder.
package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Standard project events trigger state transitions.

const (
	// EventProjectInit creates new project.
	// Transition: NoProject → PlanningActive.
	EventProjectInit = state.Event("project_init")

	// EventCompletePlanning completes planning phase.
	// Transition: PlanningActive → ImplementationPlanning.
	// Guard: task_list artifact approved.
	EventCompletePlanning = state.Event("complete_planning")

	// EventTasksApproved approves implementation tasks.
	// Transition: ImplementationPlanning → ImplementationExecuting.
	// Guard: tasks_approved metadata flag.
	EventTasksApproved = state.Event("tasks_approved")

	// EventAllTasksComplete indicates all tasks done.
	// Transition: ImplementationExecuting → ReviewActive.
	// Guard: all tasks completed or abandoned.
	EventAllTasksComplete = state.Event("all_tasks_complete")

	// EventReviewPass passes review assessment.
	// Transition: ReviewActive → FinalizeDocumentation.
	// Guard: review artifact approved with assessment=pass.
	EventReviewPass = state.Event("review_pass")

	// EventReviewFail fails review (rework loop).
	// Transition: ReviewActive → ImplementationPlanning.
	// Guard: review artifact approved with assessment=fail.
	EventReviewFail = state.Event("review_fail")

	// EventDocumentationDone completes documentation.
	// Transition: FinalizeDocumentation → FinalizeChecks.
	EventDocumentationDone = state.Event("documentation_done")

	// EventChecksDone completes final checks.
	// Transition: FinalizeChecks → FinalizeDelete.
	EventChecksDone = state.Event("checks_done")

	// EventProjectDelete deletes project.
	// Transition: FinalizeDelete → NoProject.
	// Guard: project_deleted metadata flag.
	EventProjectDelete = state.Event("project_delete")
)
