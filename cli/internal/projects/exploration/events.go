package exploration

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Exploration project events trigger state transitions.

const (
	// EventBeginSummarizing transitions from Active to Summarizing
	// Fired when all research topics are resolved
	EventBeginSummarizing = state.Event("begin_summarizing")

	// EventCompleteSummarizing transitions from Summarizing to Finalizing
	// Fired when all summary artifacts are approved
	EventCompleteSummarizing = state.Event("complete_summarizing")

	// EventCompleteFinalization transitions from Finalizing to Completed
	// Fired when all finalization tasks are completed
	EventCompleteFinalization = state.Event("complete_finalization")
)
