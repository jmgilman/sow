// Package breakdown provides the breakdown project type implementation.
// This package defines states, events, guards, and configuration for breakdown projects.
package breakdown

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Breakdown project events trigger state transitions.

const (
	// EventBeginPublishing transitions from Active to Publishing.
	// Fired when all work units are approved and dependencies are valid.
	EventBeginPublishing = state.Event("begin_publishing")

	// EventCompleteBreakdown transitions from Publishing to Completed.
	// Fired when all work units are successfully published to GitHub as issues.
	EventCompleteBreakdown = state.Event("complete_breakdown")
)
