package phases

// ImplementationPhase represents the implementation phase
#ImplementationPhase: {
	#Phase

	// Always enabled
	enabled: true

	// Whether planner agent was used
	planner_used?: null | bool @go(,optional=nillable)

	// Approved task list (gap-numbered)
	tasks: [...#Task]

	// Human approval of task plan before autonomous execution
	tasks_approved: bool

	// Tasks awaiting human approval before execution
	pending_task_additions?: null | [...#Task] @go(,optional=nillable)
}
