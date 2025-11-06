package breakdown

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Breakdown project states for the breakdown workflow.

const (
	// Active indicates active breakdown phase where work units are decomposed, specified, and reviewed.
	Active = state.State("Active")

	// Publishing indicates GitHub issue creation is in progress.
	Publishing = state.State("Publishing")

	// Completed indicates breakdown is finished and all issues have been published.
	Completed = state.State("Completed")
)
