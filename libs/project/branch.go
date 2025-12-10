package project

import (
	"github.com/jmgilman/sow/libs/project/state"
)

// BranchConfig represents a state-determined branch point in the state machine.
// It captures a discriminator function that examines project state and returns
// a string value, plus a map of branch paths that define what happens for each value.
//
// BranchConfig is used internally by AddBranch to auto-generate transitions and
// event determiners. It's stored in ProjectTypeConfig for introspection.
type BranchConfig struct {
	from          State                       // Source state
	discriminator func(*state.Project) string // Returns branch value
	branches      map[string]*BranchPath      // value -> branch path
}

// BranchPath represents one possible branch destination.
// Each path maps a discriminator value to a transition configuration.
//
// Example: discriminator returns "pass" -> fire EventReviewPass -> go to FinalizeChecks.
type BranchPath struct {
	value       string // Discriminator value that triggers this path
	event       Event  // Event to fire
	to          State  // Target state
	description string // Human-readable description

	// Standard transition configuration (forwarded from When options)
	guardTemplate GuardTemplate // Optional guard (in addition to discriminator)
	onEntry       Action        // OnEntry action
	onExit        Action        // OnExit action
	failedPhase   string        // Phase to mark as failed
}
