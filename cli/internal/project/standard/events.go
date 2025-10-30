package standard

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// Standard project events - these trigger state transitions in the standard 5-phase workflow.

const (
	// EventProjectInit is triggered when `sow project init` is called.
	// Transitions: NoProject → PlanningActive.
	EventProjectInit = statechart.Event("project_init")

	// EventCompletePlanning is triggered when planning phase is completed.
	// Requires guard: task list artifact is approved.
	// Transitions: PlanningActive → ImplementationPlanning.
	EventCompletePlanning = statechart.Event("complete_planning")

	// EventTaskCreated is triggered when `sow task init` creates at least one task.
	// Requires guard: at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTaskCreated = statechart.Event("task_created")

	// EventTasksApproved is triggered when `sow project phase approve implementation` is called.
	// Requires guard: tasks_approved flag set and at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTasksApproved = statechart.Event("tasks_approved")

	// EventAllTasksComplete is an internal auto-transition when all tasks are done.
	// Requires guard: all tasks completed or abandoned.
	// Transitions: ImplementationExecuting → ReviewActive.
	EventAllTasksComplete = statechart.Event("all_tasks_complete")

	// EventReviewFail is triggered when `sow project review add-report --assessment fail` is called.
	// This loops back to implementation with incremented iteration counter.
	// Transitions: ReviewActive → ImplementationExecuting.
	EventReviewFail = statechart.Event("review_fail")

	// EventReviewPass is triggered when `sow project review add-report --assessment pass` is called.
	// Transitions: ReviewActive → FinalizeDocumentation.
	EventReviewPass = statechart.Event("review_pass")

	// EventDocumentationDone is an internal auto-transition when documentation is assessed.
	// Requires guard: documentation updated or not needed.
	// Transitions: FinalizeDocumentation → FinalizeChecks.
	EventDocumentationDone = statechart.Event("documentation_done")

	// EventChecksDone is an internal auto-transition when checks are complete.
	// Requires guard: checks passed or not needed.
	// Transitions: FinalizeChecks → FinalizeDelete.
	EventChecksDone = statechart.Event("checks_done")

	// EventProjectDelete is triggered when `sow project delete` is called.
	// Requires guard: project_deleted flag set to true.
	// Transitions: FinalizeDelete → NoProject.
	EventProjectDelete = statechart.Event("project_delete")
)
