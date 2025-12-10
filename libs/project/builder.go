package project

import (
	"context"
	"fmt"

	"github.com/qmuntal/stateless"
)

// MachineBuilder provides a fluent API for building state machines.
type MachineBuilder struct {
	initialState      State
	promptFunc        PromptFunc
	transitions       []transitionDef
	guardDescriptions map[transitionKey]string
}

// transitionDef holds the configuration for a single transition.
type transitionDef struct {
	from    State
	to      State
	event   Event
	options []TransitionOption
}

// transitionKey uniquely identifies a transition by source state and event.
type transitionKey struct {
	from  State
	event Event
}

// NewBuilder creates a new MachineBuilder with the given initial state.
// The promptFunc parameter is optional and provides state-specific prompts.
func NewBuilder(initialState State, promptFunc PromptFunc) *MachineBuilder {
	return &MachineBuilder{
		initialState:      initialState,
		promptFunc:        promptFunc,
		transitions:       make([]transitionDef, 0),
		guardDescriptions: make(map[transitionKey]string),
	}
}

// AddTransition adds a transition from one state to another.
func (b *MachineBuilder) AddTransition(from, to State, event Event, opts ...TransitionOption) *MachineBuilder {
	b.transitions = append(b.transitions, transitionDef{
		from:    from,
		to:      to,
		event:   event,
		options: opts,
	})
	return b
}

// Build creates the state machine with all configured transitions.
func (b *MachineBuilder) Build() *Machine {
	fsm := stateless.NewStateMachine(string(b.initialState))

	// Track which states have been configured to avoid duplicate OnEntry/OnExit
	configuredStates := make(map[State]bool)

	for _, t := range b.transitions {
		// Apply options
		config := &transitionConfig{}
		for _, opt := range t.options {
			opt(config)
		}

		// Store guard description if provided
		if config.guard != nil && config.guardDescription != "" {
			key := transitionKey{from: t.from, event: t.event}
			b.guardDescriptions[key] = config.guardDescription
		}

		// Configure the source state
		cfgFrom := fsm.Configure(string(t.from))

		// Add exit action if provided and state not already configured
		if config.onExit != nil && !configuredStates[t.from] {
			cfgFrom.OnExit(config.onExit)
		}

		// Configure the transition
		if config.guard != nil {
			// Wrap guard to match stateless signature
			guardFunc := func(_ context.Context, _ ...any) bool {
				return config.guard()
			}
			cfgFrom.Permit(stateless.Trigger(string(t.event)), string(t.to), guardFunc)
		} else {
			cfgFrom.Permit(stateless.Trigger(string(t.event)), string(t.to))
		}

		// Configure the target state entry action if provided
		if config.onEntry != nil && !configuredStates[t.to] {
			cfgTo := fsm.Configure(string(t.to))
			cfgTo.OnEntry(config.onEntry)
		}

		configuredStates[t.from] = true
		if config.onEntry != nil {
			configuredStates[t.to] = true
		}
	}

	machine := NewMachine(fsm, b.promptFunc)

	// Set up custom unhandled trigger handler for better error messages
	b.setupUnhandledTriggerHandler(machine)

	return machine
}

// setupUnhandledTriggerHandler configures custom error messages for guard failures.
func (b *MachineBuilder) setupUnhandledTriggerHandler(m *Machine) {
	m.fsm.OnUnhandledTrigger(func(_ context.Context, state, trigger any, unmetGuards []string) error {
		// Handle type conversion for state - may be string or State type
		var currentState State
		switch s := state.(type) {
		case string:
			currentState = State(s)
		case State:
			currentState = s
		default:
			currentState = State(fmt.Sprintf("%v", state))
		}

		// Handle type conversion for trigger - may be string, Event, or stateless.Trigger
		var event Event
		switch t := trigger.(type) {
		case string:
			event = Event(t)
		case Event:
			event = t
		case stateless.Trigger:
			if s, ok := t.(string); ok {
				event = Event(s)
			} else {
				event = Event(fmt.Sprintf("%v", trigger))
			}
		default:
			event = Event(fmt.Sprintf("%v", trigger))
		}

		key := transitionKey{from: currentState, event: event}

		if desc, exists := b.guardDescriptions[key]; exists {
			return fmt.Errorf("guard '%s' failed for event '%s' from state '%s'", desc, event, currentState)
		}

		if len(unmetGuards) > 0 {
			return fmt.Errorf("guard conditions not met for event '%s' from state '%s': %v", event, currentState, unmetGuards)
		}

		return fmt.Errorf("trigger '%s' is not valid from state '%s'", event, currentState)
	})
}
