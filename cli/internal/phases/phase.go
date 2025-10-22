package phases

import (
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/qmuntal/stateless"
)

// Phase represents a reusable phase in the project lifecycle.
// Phases are "lego blocks" that add states and transitions to the state machine.
//
// Each phase:
// - Owns its states and transitions
// - Defines its entry point (entry state)
// - Configures the state machine by adding its states
// - Links to the next phase's entry state
// - Provides metadata for CLI validation
//
// Example usage:
//
//	discovery := discovery.New(true, &state.Phases.Discovery)
//	design := design.New(true, &state.Phases.Design)
//
//	phases := []Phase{discovery, design, ...}
//	phaseMap := BuildPhaseChain(sm, phases)
type Phase interface {
	// AddToMachine adds this phase's states and transitions to the state machine.
	// The nextPhaseEntry parameter is the entry state of the next phase in the chain,
	// allowing this phase to transition forward when complete.
	//
	// For the last phase in a chain, nextPhaseEntry should be NoProject.
	AddToMachine(sm *stateless.StateMachine, nextPhaseEntry statechart.State)

	// EntryState returns the state where this phase begins.
	// This is used by BuildPhaseChain to wire phases together.
	EntryState() statechart.State

	// Metadata returns descriptive information about this phase.
	// Used by the CLI for validation and introspection.
	Metadata() PhaseMetadata
}

// PhaseMap is a convenience type for looking up phases by name.
// Returned by BuildPhaseChain for project-specific customization.
type PhaseMap map[string]Phase
