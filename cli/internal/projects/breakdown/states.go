package breakdown

import "github.com/jmgilman/sow/libs/project"

// Breakdown project states for the breakdown workflow.

const (
	// Discovery indicates initial exploration phase where codebase/design context is gathered.
	Discovery = project.State("Discovery")

	// Active indicates active breakdown phase where work units are decomposed, specified, and reviewed.
	Active = project.State("Active")

	// Publishing indicates GitHub issue creation is in progress.
	Publishing = project.State("Publishing")

	// Completed indicates breakdown is finished and all issues have been published.
	Completed = project.State("Completed")
)
