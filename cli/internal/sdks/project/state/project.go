package state

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/project"
)

// Project wraps the CUE-generated ProjectState with runtime behavior.
// It embeds ProjectState for serialization and adds runtime-only fields
// that are not persisted (config, machine, and ctx).
type Project struct {
	project.ProjectState

	// Runtime-only fields (not serialized)
	config  *ProjectTypeConfig
	machine *Machine
	ctx     *sow.Context // Context for FS operations
}

// Config returns the project type configuration for this project.
// This provides access to phase configurations, transitions, and other
// type-specific behavior.
func (p *Project) Config() *ProjectTypeConfig {
	return p.config
}

// PhaseConfig holds configuration for a single phase in a project type.
// It defines validation rules for artifacts and metadata.
type PhaseConfig struct {
	// allowedInputTypes are the artifact types allowed as inputs
	// Empty slice means all types are allowed
	allowedInputTypes []string

	// allowedOutputTypes are the artifact types allowed as outputs
	// Empty slice means all types are allowed
	allowedOutputTypes []string

	// supportsTasks indicates whether the phase can have tasks
	supportsTasks bool

	// metadataSchema is an embedded CUE schema for metadata validation
	// Empty string means no metadata validation
	metadataSchema string
}

// Event represents a trigger that causes state transitions.
// This type is defined in the state machine SDK but redeclared here
// for convenience and to avoid import cycles.
type Event string

// GuardTemplate is a function that checks if a transition should be allowed.
// It receives the project and returns true if the transition should proceed.
type GuardTemplate func(*Project) bool

// ActionTemplate is a function that performs an action during a transition.
// It receives the project and can mutate its state.
type ActionTemplate func(*Project) error

// EventDeterminer is a function that determines which event to fire from a state.
// It examines the project state and returns the appropriate event to advance.
type EventDeterminer func(*Project) (Event, error)

// Initializer is a function that initializes a newly created project.
// It is called during Create() to set up phases, metadata, and any
// project-type-specific initial state.
type Initializer func(*Project) error

// TransitionConfig defines a state transition with guards and actions.
type TransitionConfig struct {
	From          State
	To            State
	Event         Event
	guardTemplate GuardTemplate  // Optional guard function
	onEntry       ActionTemplate // Optional action on entering To state
	onExit        ActionTemplate // Optional action on exiting From state
}

// ProjectTypeConfig holds project-type-specific configuration.
// This will be expanded in future tasks to include OnAdvance handlers,
// guards, and other project-type-specific behavior.
type ProjectTypeConfig struct {
	// phaseConfigs are the phase configurations indexed by phase name
	phaseConfigs map[string]*PhaseConfig

	// initialState is the starting state for new projects of this type
	initialState State

	// transitions define the state machine transitions
	transitions []TransitionConfig

	// onAdvance maps states to event determiners for the Advance() method
	onAdvance map[State]EventDeterminer

	// initializer is called during Create() to initialize the project
	// with phases, metadata, and any type-specific initial state
	initializer Initializer
}

// InitialState returns the configured initial state for this project type.
func (ptc *ProjectTypeConfig) InitialState() State {
	return ptc.initialState
}

// Initialize calls the configured initializer function if present.
// Returns nil if no initializer is configured.
func (ptc *ProjectTypeConfig) Initialize(project *Project) error {
	if ptc.initializer == nil {
		return nil
	}
	return ptc.initializer(project)
}

// Machine represents the state machine for the project.
// This is a placeholder for the state machine that will be
// integrated in future tasks.
type Machine struct {
	currentState State
	transitions  map[State]map[Event]*TransitionConfig
	project      *Project
}

// State returns the current state of the machine.
// This is a stub that will be implemented in Unit 3.
func (m *Machine) State() State {
	if m == nil {
		return State("")
	}
	return m.currentState
}

// CanFire checks if an event can be fired from the current state.
// This is a stub that will be implemented in Unit 3.
func (m *Machine) CanFire(event Event) (bool, error) {
	// Check if transition exists
	if m.transitions == nil {
		return false, nil
	}
	stateTransitions, ok := m.transitions[m.currentState]
	if !ok {
		return false, nil
	}
	tc, ok := stateTransitions[event]
	if !ok {
		return false, nil
	}

	// Check guard if present
	if tc.guardTemplate != nil {
		return tc.guardTemplate(m.project), nil
	}

	return true, nil
}

// Fire triggers an event, causing a state transition if valid.
// This is a stub that will be implemented in Unit 3.
func (m *Machine) Fire(event Event) error {
	// Check if transition exists
	if m.transitions == nil {
		return fmt.Errorf("no transitions configured")
	}
	stateTransitions, ok := m.transitions[m.currentState]
	if !ok {
		return fmt.Errorf("no transitions from state %s", m.currentState)
	}
	tc, ok := stateTransitions[event]
	if !ok {
		return fmt.Errorf("no transition for event %s from state %s", event, m.currentState)
	}

	// Execute onExit action
	if tc.onExit != nil {
		if err := tc.onExit(m.project); err != nil {
			return fmt.Errorf("onExit failed: %w", err)
		}
	}

	// Change state
	m.currentState = tc.To

	// Execute onEntry action
	if tc.onEntry != nil {
		if err := tc.onEntry(m.project); err != nil {
			return fmt.Errorf("onEntry failed: %w", err)
		}
	}

	return nil
}

// Helper methods for common guard patterns

// PhaseOutputApproved checks if a phase has an approved artifact of the given type.
// Returns false if phase not found, artifact not found, or artifact not approved.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseOutputApproved(phaseName, outputType string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false // Phase not found
	}

	for _, artifact := range phase.Outputs {
		if artifact.Type == outputType && artifact.Approved {
			return true
		}
	}

	return false
}

// PhaseMetadataBool reads a boolean value from phase metadata.
// Returns false if phase not found, key not found, or value is not a boolean.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseMetadataBool(phaseName, key string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false // Phase not found
	}

	if phase.Metadata == nil {
		return false // No metadata
	}

	value, ok := phase.Metadata[key]
	if !ok {
		return false // Key not found
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false // Wrong type
	}

	return boolValue
}

// AllTasksComplete checks if all tasks across all phases are completed.
// Returns true if all tasks have status "completed", or if no tasks exist (vacuous truth).
// This is a read-only method used in state machine guards.
func (p *Project) AllTasksComplete() bool {
	for _, phase := range p.Phases {
		for _, task := range phase.Tasks {
			if task.Status != "completed" {
				return false
			}
		}
	}
	return true // All tasks completed (or no tasks exist)
}

// Advance progresses the project to its next state using configured event determination.
//
// The Advance flow:
//  1. Get current state from machine
//  2. Look up event determiner for current state (from OnAdvance config)
//  3. Call determiner to get next event
//  4. Check if transition is allowed (guard evaluation via machine.CanFire)
//  5. Fire event if allowed (executes OnExit, transition, OnEntry)
//
// Returns error if:
//   - No event determiner configured for current state
//   - Determiner fails to determine event
//   - Guard prevents transition (CanFire returns false)
//   - Event firing fails
func (p *Project) Advance() error {
	// 1. Get current state
	currentState := p.machine.State()

	// 2. Get event determiner for current state
	determiner := p.config.GetEventDeterminer(currentState)
	if determiner == nil {
		return fmt.Errorf("no event determiner for state: %s", currentState)
	}

	// 3. Determine next event based on project state
	event, err := determiner(p)
	if err != nil {
		return fmt.Errorf("failed to determine event: %w", err)
	}

	// 4. Check if transition is allowed (guard evaluation)
	can, err := p.machine.CanFire(event)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot fire event %s from state %s", event, currentState)
	}

	// 5. Fire the event (executes OnExit, transition, OnEntry)
	if err := p.machine.Fire(event); err != nil {
		return fmt.Errorf("failed to fire event: %w", err)
	}

	return nil
}
