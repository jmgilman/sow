// Package statechart implements a state machine for project lifecycle management.
package statechart

// Event represents a trigger that causes state transitions.
type Event string

const (
	// EventProjectInit is triggered when `sow project init` is called.
	// Transitions: NoProject → PlanningActive.
	EventProjectInit Event = "project_init"

	// EventCompletePlanning is triggered when planning phase is completed.
	// Requires guard: task list artifact is approved.
	// Transitions: PlanningActive → ImplementationPlanning.
	EventCompletePlanning Event = "complete_planning"

	// EventTaskCreated is triggered when `sow task init` creates at least one task.
	// Requires guard: at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTaskCreated Event = "task_created"

	// EventTasksApproved is triggered when `sow project phase approve implementation` is called.
	// Requires guard: tasks_approved flag set and at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTasksApproved Event = "tasks_approved"

	// EventAllTasksComplete is an internal auto-transition when all tasks are done.
	// Requires guard: all tasks completed or abandoned.
	// Transitions: ImplementationExecuting → ReviewActive.
	EventAllTasksComplete Event = "all_tasks_complete"

	// EventReviewFail is triggered when `sow project review add-report --assessment fail` is called.
	// This loops back to implementation with incremented iteration counter.
	// Transitions: ReviewActive → ImplementationExecuting.
	EventReviewFail Event = "review_fail"

	// EventReviewPass is triggered when `sow project review add-report --assessment pass` is called.
	// Transitions: ReviewActive → FinalizeDocumentation.
	EventReviewPass Event = "review_pass"

	// EventDocumentationDone is an internal auto-transition when documentation is assessed.
	// Requires guard: documentation updated or not needed.
	// Transitions: FinalizeDocumentation → FinalizeChecks.
	EventDocumentationDone Event = "documentation_done"

	// EventChecksDone is an internal auto-transition when checks are complete.
	// Requires guard: checks passed or not needed.
	// Transitions: FinalizeChecks → FinalizeDelete.
	EventChecksDone Event = "checks_done"

	// EventProjectDelete is triggered when `sow project delete` is called.
	// Requires guard: project_deleted flag set to true.
	// Transitions: FinalizeDelete → NoProject.
	EventProjectDelete Event = "project_delete"
)

// String returns the string representation of the event.
func (e Event) String() string {
	return string(e)
}
