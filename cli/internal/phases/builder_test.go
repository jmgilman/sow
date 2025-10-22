package phases

import (
	"context"
	"testing"

	"github.com/qmuntal/stateless"
)

// MockPhase is a test implementation of the Phase interface
type MockPhase struct {
	name         string
	entryState   State
	states       []State
	addedToMachine bool
	nextEntry    State
}

func NewMockPhase(name string, entryState State, states []State) *MockPhase {
	return &MockPhase{
		name:       name,
		entryState: entryState,
		states:     states,
	}
}

func (m *MockPhase) EntryState() State {
	return m.entryState
}

func (m *MockPhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry State) {
	m.addedToMachine = true
	m.nextEntry = nextPhaseEntry

	// Configure each state in this mock phase
	for _, state := range m.states {
		sm.Configure(state)
	}
}

func (m *MockPhase) Metadata() PhaseMetadata {
	return PhaseMetadata{
		Name:   m.name,
		States: m.states,
	}
}

func TestBuildPhaseChain_EmptyPhases(t *testing.T) {
	sm := stateless.NewStateMachine(NoProject)
	phases := []Phase{}

	phaseMap := BuildPhaseChain(sm, phases)

	if len(phaseMap) != 0 {
		t.Errorf("Expected empty phase map, got %d phases", len(phaseMap))
	}
}

func TestBuildPhaseChain_SinglePhase(t *testing.T) {
	sm := stateless.NewStateMachine(NoProject)
	mockPhase := NewMockPhase("test", DiscoveryDecision, []State{DiscoveryDecision, DiscoveryActive})
	phases := []Phase{mockPhase}

	phaseMap := BuildPhaseChain(sm, phases)

	// Verify phase was added to machine
	if !mockPhase.addedToMachine {
		t.Error("Expected phase to be added to machine")
	}

	// Verify next entry is NoProject (last phase loops back)
	if mockPhase.nextEntry != NoProject {
		t.Errorf("Expected next entry to be NoProject, got %s", mockPhase.nextEntry)
	}

	// Verify phase map
	if len(phaseMap) != 1 {
		t.Errorf("Expected 1 phase in map, got %d", len(phaseMap))
	}

	if phaseMap["test"] != mockPhase {
		t.Error("Expected phase to be in map")
	}

	// Verify NoProject can transition to first phase
	canFire, _ := sm.CanFire(EventProjectInit)
	if !canFire {
		t.Error("Expected NoProject to permit EventProjectInit")
	}
}

func TestBuildPhaseChain_MultiplePhases(t *testing.T) {
	sm := stateless.NewStateMachine(NoProject)

	phase1 := NewMockPhase("phase1", DiscoveryDecision, []State{DiscoveryDecision, DiscoveryActive})
	phase2 := NewMockPhase("phase2", DesignDecision, []State{DesignDecision, DesignActive})
	phase3 := NewMockPhase("phase3", ImplementationPlanning, []State{ImplementationPlanning})

	phases := []Phase{phase1, phase2, phase3}

	phaseMap := BuildPhaseChain(sm, phases)

	// Verify all phases were added
	if !phase1.addedToMachine || !phase2.addedToMachine || !phase3.addedToMachine {
		t.Error("Expected all phases to be added to machine")
	}

	// Verify phase chaining
	if phase1.nextEntry != phase2.EntryState() {
		t.Errorf("Expected phase1 next to be %s, got %s", phase2.EntryState(), phase1.nextEntry)
	}

	if phase2.nextEntry != phase3.EntryState() {
		t.Errorf("Expected phase2 next to be %s, got %s", phase3.EntryState(), phase2.nextEntry)
	}

	if phase3.nextEntry != NoProject {
		t.Errorf("Expected phase3 next to be NoProject, got %s", phase3.nextEntry)
	}

	// Verify phase map contains all phases
	if len(phaseMap) != 3 {
		t.Errorf("Expected 3 phases in map, got %d", len(phaseMap))
	}

	if phaseMap["phase1"] != phase1 {
		t.Error("Expected phase1 in map")
	}
	if phaseMap["phase2"] != phase2 {
		t.Error("Expected phase2 in map")
	}
	if phaseMap["phase3"] != phase3 {
		t.Error("Expected phase3 in map")
	}
}

func TestBuildPhaseChain_InitialTransition(t *testing.T) {
	sm := stateless.NewStateMachine(NoProject)

	phase1 := NewMockPhase("phase1", DiscoveryDecision, []State{DiscoveryDecision})
	phases := []Phase{phase1}

	BuildPhaseChain(sm, phases)

	// Verify NoProject → first phase entry transition exists
	canFire, err := sm.CanFire(EventProjectInit)
	if err != nil {
		t.Fatalf("Error checking if event can fire: %v", err)
	}

	if !canFire {
		t.Error("Expected EventProjectInit to be permitted from NoProject")
	}

	// Fire the transition and verify state change
	err = sm.Fire(EventProjectInit)
	if err != nil {
		t.Fatalf("Error firing EventProjectInit: %v", err)
	}

	currentState := sm.MustState().(State)
	if currentState != DiscoveryDecision {
		t.Errorf("Expected state to be DiscoveryDecision, got %s", currentState)
	}
}

func TestBuildPhaseChain_PhaseMapLookup(t *testing.T) {
	sm := stateless.NewStateMachine(NoProject)

	phase1 := NewMockPhase("discovery", DiscoveryDecision, []State{DiscoveryDecision})
	phase2 := NewMockPhase("design", DesignDecision, []State{DesignDecision})

	phases := []Phase{phase1, phase2}
	phaseMap := BuildPhaseChain(sm, phases)

	// Test lookup by name
	discoveryPhase, ok := phaseMap["discovery"]
	if !ok {
		t.Error("Expected to find discovery phase in map")
	}

	if discoveryPhase.EntryState() != DiscoveryDecision {
		t.Errorf("Expected discovery phase entry state to be DiscoveryDecision, got %s", discoveryPhase.EntryState())
	}

	designPhase, ok := phaseMap["design"]
	if !ok {
		t.Error("Expected to find design phase in map")
	}

	if designPhase.EntryState() != DesignDecision {
		t.Errorf("Expected design phase entry state to be DesignDecision, got %s", designPhase.EntryState())
	}

	// Test lookup of non-existent phase
	_, ok = phaseMap["nonexistent"]
	if ok {
		t.Error("Expected to not find nonexistent phase in map")
	}
}

// Test that phases can be customized after BuildPhaseChain
func TestBuildPhaseChain_PostBuildCustomization(t *testing.T) {
	sm := stateless.NewStateMachine(NoProject)

	phase1 := NewMockPhase("phase1", DiscoveryDecision, []State{DiscoveryDecision})
	phase2 := NewMockPhase("phase2", DesignDecision, []State{DesignDecision})

	phases := []Phase{phase1, phase2}
	phaseMap := BuildPhaseChain(sm, phases)

	// Add a custom backward transition using the phase map
	phase1Entry := phaseMap["phase1"].EntryState()

	// Configure a backward transition from phase2 to phase1
	sm.Configure(DesignDecision).
		Permit("custom_backward", phase1Entry, func(ctx context.Context, args ...any) bool {
			return true
		})

	// Verify the custom transition was added
	// Create a new state machine at DesignDecision to test the custom transition
	sm2 := stateless.NewStateMachine(DesignDecision)
	BuildPhaseChain(sm2, phases)
	sm2.Configure(DesignDecision).
		Permit("custom_backward", phase1Entry, func(ctx context.Context, args ...any) bool {
			return true
		})

	canFire, err := sm2.CanFire("custom_backward")
	if err != nil {
		t.Fatalf("Error checking if custom event can fire: %v", err)
	}

	if !canFire {
		t.Error("Expected custom backward transition to be permitted")
	}
}
