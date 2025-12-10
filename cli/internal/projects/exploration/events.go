// Package exploration provides the exploration project type implementation.
// This package defines states, events, guards, and configuration for exploration projects.
package exploration

import (
	"github.com/jmgilman/sow/libs/project"
)

// Exploration project events trigger state transitions.

const (
	// EventBeginSummarizing transitions from Active to Summarizing.
	// Fired when all research topics are resolved.
	EventBeginSummarizing = project.Event("begin_summarizing")

	// EventCompleteSummarizing transitions from Summarizing to Finalizing.
	// Fired when all summary artifacts are approved.
	EventCompleteSummarizing = project.Event("complete_summarizing")

	// EventCompleteFinalization transitions from Finalizing to Completed.
	// Fired when all finalization tasks are completed.
	EventCompleteFinalization = project.Event("complete_finalization")
)
