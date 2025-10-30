package statechart

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/qmuntal/stateless"
)

// Machine wraps the stateless state machine with project-specific context.
type Machine struct {
	sm              *stateless.StateMachine
	projectState    *schemas.ProjectState
	fs              sow.FS // Optional filesystem for testability
	suppressPrompts bool   // Suppress prompt printing (useful for tests and CLI commands)
}

// NewMachine creates a new state machine for project lifecycle management.
// The initial state is determined from the project state (or NoProject if nil).
//
// Deprecated: Use MachineBuilder to construct state machines. This function is only
// used by tests and may be removed in the future.
func NewMachine(projectState *schemas.ProjectState) *Machine {
	sm := stateless.NewStateMachine(NoProject)
	return &Machine{
		sm:              sm,
		projectState:    projectState,
		suppressPrompts: false,
	}
}


// ProjectState returns the machine's project state for modification.
func (m *Machine) ProjectState() *schemas.ProjectState {
	return m.projectState
}

// SetProjectState sets the machine's project state.
func (m *Machine) SetProjectState(state *schemas.ProjectState) {
	m.projectState = state
}

// SetFilesystem sets the filesystem for persistence operations.
func (m *Machine) SetFilesystem(fs sow.FS) {
	m.fs = fs
}

// SuppressPrompts disables prompt output (useful for tests and non-interactive CLI commands).
func (m *Machine) SuppressPrompts(suppress bool) {
	m.suppressPrompts = suppress
}


// Fire triggers an event, causing a state transition if valid.
func (m *Machine) Fire(event Event) error {
	if err := m.sm.Fire(event); err != nil {
		return fmt.Errorf("failed to fire event %s: %w", event, err)
	}
	return nil
}

// State returns the current state.
func (m *Machine) State() State {
	state := m.sm.MustState()
	if s, ok := state.(State); ok {
		return s
	}
	// This should never happen if the state machine is properly configured
	return NoProject
}

// CanFire checks if an event can be fired from the current state.
func (m *Machine) CanFire(event Event) (bool, error) {
	can, err := m.sm.CanFire(event)
	if err != nil {
		return false, fmt.Errorf("failed to check if event %s can fire: %w", event, err)
	}
	return can, nil
}

// PermittedTriggers returns all events that can be fired from the current state.
func (m *Machine) PermittedTriggers() ([]Event, error) {
	triggers, err := m.sm.PermittedTriggers()
	if err != nil {
		return nil, fmt.Errorf("failed to get permitted triggers: %w", err)
	}
	events := make([]Event, 0, len(triggers))
	for _, t := range triggers {
		if e, ok := t.(Event); ok {
			events = append(events, e)
		}
	}
	return events, nil
}
