package exploration

import (
	"github.com/jmgilman/sow/libs/project"
)

// Exploration project states for the research workflow.

const (
	// Active indicates active research phase.
	Active = project.State("Active")

	// Summarizing indicates synthesis/summarizing phase.
	Summarizing = project.State("Summarizing")

	// Finalizing indicates finalization in progress.
	Finalizing = project.State("Finalizing")

	// Completed indicates exploration finished.
	Completed = project.State("Completed")
)
