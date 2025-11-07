package standard

// Metadata schema for implementation phase
{
	// Planning approval flag
	// Set to true when user approves task descriptions in .sow/project/context/tasks/
	// Used by guard to allow transition from ImplementationPlanning -> ImplementationDraftPRCreation
	planning_approved?: bool

	// Draft PR creation flag
	// Set to true when draft PR has been created and metadata stored
	// Used by guard to allow transition from ImplementationDraftPRCreation -> ImplementationExecuting
	draft_pr_created?: bool

	// PR URL from GitHub
	// Example: "https://github.com/owner/repo/pull/123"
	pr_url?: string

	// PR number extracted from URL
	// Example: 123
	pr_number?: int

	// Future: Could add fields like:
	// - planning_iteration?: int (for rework cycles)
	// - planner_agent_id?: string (for tracking which agent created plans)
}
