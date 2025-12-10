package project

import (
	"fmt"

	"github.com/qmuntal/stateless"
)

// Machine wraps a qmuntal/stateless state machine with project-specific behavior.
// It provides type-safe access to state transitions and optional prompt generation.
type Machine struct {
	fsm       *stateless.StateMachine
	promptGen PromptFunc
}

// NewMachine creates a new Machine wrapper.
// This is typically called by MachineBuilder.Build().
func NewMachine(fsm *stateless.StateMachine, promptGen PromptFunc) *Machine {
	return &Machine{
		fsm:       fsm,
		promptGen: promptGen,
	}
}

// State returns the current state.
func (m *Machine) State() State {
	state := m.fsm.MustState()
	s, ok := state.(string)
	if !ok {
		// State should always be a string when properly configured
		return State("")
	}
	return State(s)
}

// Fire triggers a state transition with the given event.
// Returns an error if the transition is not allowed.
func (m *Machine) Fire(event Event) error {
	if err := m.fsm.Fire(string(event)); err != nil {
		return fmt.Errorf("transition not allowed: cannot fire '%s' from state '%s': %w", event, m.State(), err)
	}
	return nil
}

// CanFire returns true if the given event can be fired from the current state.
func (m *Machine) CanFire(event Event) bool {
	can, _ := m.fsm.CanFire(string(event))
	return can
}

// PermittedTriggers returns all events that can be fired from the current state.
func (m *Machine) PermittedTriggers() []Event {
	triggers, _ := m.fsm.PermittedTriggers()
	events := make([]Event, 0, len(triggers))
	for _, t := range triggers {
		if s, ok := t.(string); ok {
			events = append(events, Event(s))
		}
	}
	return events
}

// Prompt returns the prompt for the current state.
// Returns empty string if no prompt is configured for the state.
func (m *Machine) Prompt() string {
	if m.promptGen == nil {
		return ""
	}
	return m.promptGen(m.State())
}

// FSM returns the underlying stateless.StateMachine.
// This is used when the raw machine is needed (e.g., for interface implementations).
func (m *Machine) FSM() *stateless.StateMachine {
	return m.fsm
}
