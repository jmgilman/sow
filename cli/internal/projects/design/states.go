package design

import "github.com/jmgilman/sow/libs/project"

// Design project states for the design workflow.

const (
	// Active indicates active design phase.
	Active = project.State("Active")

	// Finalizing indicates finalization in progress.
	Finalizing = project.State("Finalizing")

	// Completed indicates design finished.
	Completed = project.State("Completed")
)
