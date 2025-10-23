package statechart

import (
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/projects"
	"github.com/qmuntal/stateless"
)

// NewMachineFromPhases creates a Machine from a pre-configured stateless.StateMachine.
//
// Example usage:
//
//	sm := stateless.NewStateMachine(initialState)
//	// Phases configure sm via BuildPhaseChain...
//	machine := NewMachineFromPhases(sm, projectState)
//
// Parameters:
//   - sm: Pre-configured stateless state machine with all states and transitions
//   - projectState: Project state for persistence and guard evaluation
//
// Returns:
//   - *Machine: Wrapped machine ready for use
func NewMachineFromPhases(sm *stateless.StateMachine, projectState *projects.ProjectState) *Machine {
	// Convert projects.ProjectState to schemas.ProjectState (they're type aliases)
	schemasState := (*schemas.ProjectState)(projectState)

	return &Machine{
		sm:              sm,
		projectState:    schemasState,
		suppressPrompts: false, // Show prompts by default
	}
}
