package phases

// Event represents a trigger that causes state transitions.
type Event string

const (
	// EventProjectInit is triggered when a new project is initialized.
	// Transitions: NoProject → (first phase entry state).
	EventProjectInit Event = "project_init"

	// EventEnableDiscovery is triggered when discovery phase is enabled.
	// Transitions: DiscoveryDecision → DiscoveryActive.
	EventEnableDiscovery Event = "enable_discovery"

	// EventSkipDiscovery is triggered when discovery phase is skipped.
	// Transitions: DiscoveryDecision → (next phase entry state).
	EventSkipDiscovery Event = "skip_discovery"

	// EventCompleteDiscovery is triggered when discovery phase is completed.
	// Requires guard: all artifacts approved or no artifacts exist.
	// Transitions: DiscoveryActive → (next phase entry state).
	EventCompleteDiscovery Event = "complete_discovery"

	// EventEnableDesign is triggered when design phase is enabled.
	// Transitions: DesignDecision → DesignActive.
	EventEnableDesign Event = "enable_design"

	// EventSkipDesign is triggered when design phase is skipped.
	// Transitions: DesignDecision → (next phase entry state).
	EventSkipDesign Event = "skip_design"

	// EventCompleteDesign is triggered when design phase is completed.
	// Requires guard: all artifacts approved or no artifacts exist.
	// Transitions: DesignActive → (next phase entry state).
	EventCompleteDesign Event = "complete_design"

	// EventTaskCreated is triggered when a task is created during implementation planning.
	// Requires guard: at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTaskCreated Event = "task_created"

	// EventTasksApproved is triggered when implementation tasks are approved.
	// Requires guard: tasks_approved flag set and at least 1 task exists.
	// Transitions: ImplementationPlanning → ImplementationExecuting.
	EventTasksApproved Event = "tasks_approved"

	// EventAllTasksComplete is triggered when all tasks are completed.
	// Requires guard: all tasks completed or abandoned.
	// Transitions: ImplementationExecuting → (next phase entry state).
	EventAllTasksComplete Event = "all_tasks_complete"

	// EventReviewFail is triggered when review fails.
	// Transitions: ReviewActive → ImplementationPlanning (backward loop).
	EventReviewFail Event = "review_fail"

	// EventReviewPass is triggered when review passes.
	// Transitions: ReviewActive → (next phase entry state).
	EventReviewPass Event = "review_pass"

	// EventDocumentationDone is triggered when documentation is completed.
	// Requires guard: documentation updated or not needed.
	// Transitions: FinalizeDocumentation → FinalizeChecks.
	EventDocumentationDone Event = "documentation_done"

	// EventChecksDone is triggered when final checks are completed.
	// Requires guard: checks passed or not needed.
	// Transitions: FinalizeChecks → FinalizeDelete.
	EventChecksDone Event = "checks_done"

	// EventProjectDelete is triggered when project is deleted.
	// Requires guard: project_deleted flag set to true.
	// Transitions: FinalizeDelete → NoProject.
	EventProjectDelete Event = "project_delete"
)

// String returns the string representation of the event.
func (e Event) String() string {
	return string(e)
}
