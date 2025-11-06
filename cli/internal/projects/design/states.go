package design

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Design project states for the design workflow.

const (
	// Active indicates active design phase.
	Active = state.State("Active")

	// Finalizing indicates finalization in progress.
	Finalizing = state.State("Finalizing")

	// Completed indicates design finished.
	Completed = state.State("Completed")
)
