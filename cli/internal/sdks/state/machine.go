package state

import (
	"context"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/qmuntal/stateless"
)

// Machine wraps the stateless state machine with project-specific context.
type Machine struct {
	sm                *stateless.StateMachine
	projectState      *schemas.ProjectState
	fs                sow.FS // Optional filesystem for testability
	guardDescriptions map[transitionKey]string
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

// setupUnhandledTriggerHandler configures custom error messages for guard failures.
// It replaces the default error message (which shows generic func names like "func1")
// with our custom guard descriptions when available.
func (m *Machine) setupUnhandledTriggerHandler() {
	m.sm.OnUnhandledTrigger(func(_ context.Context, state, trigger any, unmetGuards []string) error {
		// Try to get our custom description
		currentState, _ := state.(State)
		event, _ := trigger.(Event)
		key := transitionKey{from: currentState, event: event}

		if desc, exists := m.guardDescriptions[key]; exists {
			// Use our custom description
			return fmt.Errorf("stateless: trigger '%v' is valid for transition from state '%v' but guard condition is not met: %s", trigger, state, desc)
		}

		// Fallback to default error with generic guard descriptions
		return fmt.Errorf("stateless: trigger '%v' is valid for transition from state '%v' but guard conditions are not met. Guard descriptions: '%v'", trigger, state, unmetGuards)
	})
}
