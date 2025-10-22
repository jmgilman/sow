package phases

import (
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/qmuntal/stateless"
)

// BuildPhaseChain wires up a sequence of phases into a state machine.
//
// This meta-level helper:
// 1. Configures NoProject → first phase transition
// 2. Chains phases together by linking each phase's completion to the next phase's entry
// 3. Returns a PhaseMap for project-specific customization (e.g., adding backward transitions)
//
// Example usage:
//
//	sm := stateless.NewStateMachine(NoProject)
//	phases := []Phase{
//	    discovery.New(true, &state.Phases.Discovery),
//	    design.New(true, &state.Phases.Design),
//	    implementation.New(&state.Phases.Implementation),
//	}
//	phaseMap := BuildPhaseChain(sm, phases)
//
//	// Add exceptional backward transition
//	reviewPhase := phaseMap["review"]
//	implPhase := phaseMap["implementation"]
//	sm.Configure(ReviewActive).
//	    Permit(EventReviewFail, implPhase.EntryState(), guardFunc)
//
// Parameters:
//   - sm: The stateless state machine to configure
//   - phases: Ordered list of phases to chain together
//
// Returns:
//   - PhaseMap: Map of phase name → Phase for customization
func BuildPhaseChain(sm *stateless.StateMachine, phases []Phase) PhaseMap {
	if len(phases) == 0 {
		return make(PhaseMap)
	}

	phaseMap := make(PhaseMap)

	// Configure initial transition: NoProject → first phase
	sm.Configure(statechart.NoProject).
		Permit(statechart.EventProjectInit, phases[0].EntryState())

	// Chain phases together
	for i, phase := range phases {
		// Store in map for later lookup
		phaseMap[phase.Metadata().Name] = phase

		// Determine next phase entry state
		var nextEntry statechart.State
		if i < len(phases)-1 {
			// Link to next phase
			nextEntry = phases[i+1].EntryState()
		} else {
			// Last phase loops back to NoProject
			nextEntry = statechart.NoProject
		}

		// Let the phase configure its states and transitions
		phase.AddToMachine(sm, nextEntry)
	}

	return phaseMap
}
