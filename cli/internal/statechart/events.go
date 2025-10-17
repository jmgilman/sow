// Package statechart implements a state machine for project lifecycle management.
package statechart

// Event represents a trigger that causes state transitions.
type Event string

const (
	// EventProjectInit is triggered when `sow project init` is called.
	// Transitions: NoProject → DiscoveryDecision.
	EventProjectInit Event = "project_init"

	// EventEnableDiscovery is triggered when `sow project phase enable discovery` is called.
	// Transitions: DiscoveryDecision → DiscoveryActive.
	EventEnableDiscovery Event = "enable_discovery"

	// EventSkipDiscovery is an internal auto-transition when discovery is not needed.
	// Transitions: DiscoveryDecision → DesignDecision.
	EventSkipDiscovery Event = "skip_discovery"

	// EventCompleteDiscovery is triggered when `sow project phase complete discovery` is called.
	// Requires guard: all artifacts approved or no artifacts exist.
	// Transitions: DiscoveryActive → DesignDecision.
	EventCompleteDiscovery Event = "complete_discovery"

	// EventEnableDesign is triggered when `sow project phase enable design` is called.
	// Transitions: DesignDecision → DesignActive.
	EventEnableDesign Event = "enable_design"

	// EventSkipDesign is an internal auto-transition when design is not needed.
	// Transitions: DesignDecision → ImplementationPlanning.
	EventSkipDesign Event = "skip_design"

	// EventCompleteDesign is triggered when `sow project phase complete design` is called.
	// Requires guard: all artifacts approved or no artifacts exist.
	// Transitions: DesignActive → ImplementationPlanning.
	EventCompleteDesign Event = "complete_design"

	// EventTaskCreated is triggered when `sow task init` creates at least one task.
	// Requires guard: at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTaskCreated Event = "task_created"

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
