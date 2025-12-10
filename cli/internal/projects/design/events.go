// Package design provides the design project type implementation.
// This package defines states, events, guards, and configuration for design projects.
package design

import "github.com/jmgilman/sow/libs/project"

// Design project events trigger state transitions.

const (
	// EventCompleteDesign transitions from Active to Finalizing.
	// Fired when all documents are approved.
	EventCompleteDesign = project.Event("complete_design")

	// EventCompleteFinalization transitions from Finalizing to Completed.
	// Fired when all finalization tasks are completed.
	EventCompleteFinalization = project.Event("complete_finalization")
)
