package standard

// Metadata schema for implementation phase
{
	// Planning approval flag
	// Set to true when user approves task descriptions in .sow/project/context/tasks/
	// Used by guard to allow transition from ImplementationPlanning -> ImplementationExecuting
	planning_approved?: bool

	// Future: Could add fields like:
	// - planning_iteration?: int (for rework cycles)
	// - planner_agent_id?: string (for tracking which agent created plans)
}
