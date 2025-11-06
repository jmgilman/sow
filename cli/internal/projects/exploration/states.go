package exploration

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Exploration project states for the research workflow.

const (
	// Active indicates active research phase.
	Active = state.State("Active")

	// Summarizing indicates synthesis/summarizing phase.
	Summarizing = state.State("Summarizing")

	// Finalizing indicates finalization in progress.
	Finalizing = state.State("Finalizing")

	// Completed indicates exploration finished.
	Completed = state.State("Completed")
)
