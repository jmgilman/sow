package phases

import (
	"context"
	"testing"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/qmuntal/stateless"
)

// MockPhase is a test implementation of the Phase interface.
type MockPhase struct {
	name           string
	entryState     statechart.State
	states         []statechart.State
	addedToMachine bool
	nextEntry      statechart.State
}

func NewMockPhase(name string, entryState statechart.State, states []statechart.State) *MockPhase {
	return &MockPhase{
		name:       name,
		entryState: entryState,
		states:     states,
	}
}

func (m *MockPhase) EntryState() statechart.State {
	return m.entryState
}

func (m *MockPhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry statechart.State) {
	m.addedToMachine = true
	m.nextEntry = nextPhaseEntry

	// Configure basic transitions for each state in this mock phase
	// This simulates what real phases do (configure skip/complete events)
	for _, state := range m.states {
		// Configure a forward transition to next phase entry
		// Use appropriate events based on which state this is
		switch state {
		case statechart.DiscoveryDecision:
			sm.Configure(state).
				Permit(statechart.EventSkipDiscovery, nextPhaseEntry)
		case statechart.DesignDecision:
			sm.Configure(state).
				Permit(statechart.EventSkipDesign, nextPhaseEntry)
		default:
			// Generic state - just configure it
			sm.Configure(state)
		}
	}
}

func (m *MockPhase) Metadata() PhaseMetadata {
	return PhaseMetadata{
		Name:   m.name,
		States: m.states,
	}
}

func TestBuildPhaseChain_EmptyPhases(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.NoProject)
	phases := []Phase{}

	phaseMap := BuildPhaseChain(sm, phases)

	if len(phaseMap) != 0 {
		t.Errorf("Expected empty phase map, got %d phases", len(phaseMap))
	}
}

func TestBuildPhaseChain_SinglePhase(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.NoProject)
	mockPhase := NewMockPhase("test", statechart.DiscoveryDecision, []statechart.State{statechart.DiscoveryDecision})

	phases := []Phase{mockPhase}
	phaseMap := BuildPhaseChain(sm, phases)

	if !mockPhase.addedToMachine {
		t.Error("Phase should have been added to machine")
	}

	if mockPhase.nextEntry != statechart.NoProject {
		t.Errorf("Last phase should link to NoProject, got %v", mockPhase.nextEntry)
	}

	if len(phaseMap) != 1 {
		t.Errorf("Expected 1 phase in map, got %d", len(phaseMap))
	}

	if phaseMap["test"] != mockPhase {
		t.Error("Phase should be in map with correct name")
	}
}

func TestBuildPhaseChain_MultiplePhases(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.NoProject)

	phase1 := NewMockPhase("phase1", statechart.DiscoveryDecision, []statechart.State{statechart.DiscoveryDecision})
	phase2 := NewMockPhase("phase2", statechart.DesignDecision, []statechart.State{statechart.DesignDecision})
	phase3 := NewMockPhase("phase3", statechart.ImplementationPlanning, []statechart.State{statechart.ImplementationPlanning})

	phases := []Phase{phase1, phase2, phase3}
	phaseMap := BuildPhaseChain(sm, phases)

	// Check all phases added
	if !phase1.addedToMachine || !phase2.addedToMachine || !phase3.addedToMachine {
		t.Error("All phases should have been added to machine")
	}

	// Check forward chaining
	if phase1.nextEntry != phase2.EntryState() {
		t.Errorf("Phase1 should link to phase2, got %v", phase1.nextEntry)
	}

	if phase2.nextEntry != phase3.EntryState() {
		t.Errorf("Phase2 should link to phase3, got %v", phase2.nextEntry)
	}

	if phase3.nextEntry != statechart.NoProject {
		t.Errorf("Phase3 should link to NoProject, got %v", phase3.nextEntry)
	}

	// Check phase map
	if len(phaseMap) != 3 {
		t.Errorf("Expected 3 phases in map, got %d", len(phaseMap))
	}
}

func TestBuildPhaseChain_InitialTransition(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.NoProject)
	mockPhase := NewMockPhase("test", statechart.DiscoveryDecision, []statechart.State{statechart.DiscoveryDecision})

	phases := []Phase{mockPhase}
	BuildPhaseChain(sm, phases)

	// Fire project init event
	err := sm.Fire(statechart.EventProjectInit)
	if err != nil {
		t.Errorf("Should be able to fire EventProjectInit: %v", err)
	}

	// Check we transitioned to first phase
	if sm.MustState() != statechart.DiscoveryDecision {
		t.Errorf("Should transition to DiscoveryDecision, got %v", sm.MustState())
	}
}

func TestBuildPhaseChain_PhaseMapLookup(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.NoProject)

	phase1 := NewMockPhase("discovery", statechart.DiscoveryDecision, []statechart.State{statechart.DiscoveryDecision})
	phase2 := NewMockPhase("design", statechart.DesignDecision, []statechart.State{statechart.DesignDecision})

	phases := []Phase{phase1, phase2}
	phaseMap := BuildPhaseChain(sm, phases)

	// Test lookup by name
	discoveryPhase, ok := phaseMap["discovery"]
	if !ok {
		t.Error("Should find discovery phase in map")
	}
	if discoveryPhase != phase1 {
		t.Error("Should return correct phase instance")
	}

	designPhase, ok := phaseMap["design"]
	if !ok {
		t.Error("Should find design phase in map")
	}
	if designPhase != phase2 {
		t.Error("Should return correct phase instance")
	}
}

func TestBuildPhaseChain_PostBuildCustomization(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.NoProject)

	phase1 := NewMockPhase("phase1", statechart.DiscoveryDecision, []statechart.State{statechart.DiscoveryDecision})
	phase2 := NewMockPhase("phase2", statechart.DesignDecision, []statechart.State{statechart.DesignDecision})

	phases := []Phase{phase1, phase2}
	phaseMap := BuildPhaseChain(sm, phases)

	// After building chain, add a custom backward transition
	// This simulates what a project type would do
	customGuard := func(_ context.Context, _ ...any) bool {
		return true
	}

	sm.Configure(statechart.DesignDecision).
		Permit(statechart.EventReviewFail, phase1.EntryState(), customGuard)

	// The custom transition should work
	err := sm.Fire(statechart.EventProjectInit)
	if err != nil {
		t.Errorf("Should transition to first phase: %v", err)
	}

	err = sm.Fire(statechart.EventSkipDiscovery)
	if err != nil {
		t.Errorf("Should transition to second phase: %v", err)
	}

	// Test backward transition (custom)
	err = sm.Fire(statechart.EventReviewFail)
	if err != nil {
		t.Errorf("Should allow custom backward transition: %v", err)
	}

	// Verify we used the phase map correctly
	if phaseMap["phase1"] != phase1 {
		t.Error("Phase map should contain correct phase references")
	}
}
