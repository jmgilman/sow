// Package breakdown provides the breakdown project type implementation.
// This package defines states, events, guards, and configuration for breakdown projects.
package breakdown

import "github.com/jmgilman/sow/libs/project"

// Breakdown project events trigger state transitions.

const (
	// EventBeginActive transitions from Discovery to Active.
	// Fired when discovery document is approved and work unit identification can begin.
	EventBeginActive = project.Event("begin_active")

	// EventBeginPublishing transitions from Active to Publishing.
	// Fired when all work units are approved and dependencies are valid.
	EventBeginPublishing = project.Event("begin_publishing")

	// EventCompleteBreakdown transitions from Publishing to Completed.
	// Fired when all work units are successfully published to GitHub as issues.
	EventCompleteBreakdown = project.Event("complete_breakdown")
)
