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
	guardDescriptions map[guardDescKey]string
}

// guardDescKey uniquely identifies a guard description by source, target, and event.
// Including 'to' allows different guard descriptions for branching transitions
// that share the same source state and event but go to different targets.
type guardDescKey struct {
	from  State
	to    State
	event Event
}

// transitionDef holds the configuration for a single transition.
type transitionDef struct {
	from    State
	to      State
	event   Event
	options []TransitionOption
}


// NewBuilder creates a new MachineBuilder with the given initial state.
// The promptFunc parameter is optional and provides state-specific prompts.
func NewBuilder(initialState State, promptFunc PromptFunc) *MachineBuilder {
	return &MachineBuilder{
		initialState:      initialState,
		promptFunc:        promptFunc,
		transitions:       make([]transitionDef, 0),
		guardDescriptions: make(map[guardDescKey]string),
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
//
// OnEntry and OnExit actions are composed when multiple transitions share the same
// target or source state respectively. The stateless library only supports one OnEntry
// and one OnExit per state, so this builder automatically chains multiple actions.
func (b *MachineBuilder) Build() *Machine {
	fsm := stateless.NewStateMachine(string(b.initialState))

	// Collect all onExit actions per source state and onEntry actions per target state.
	// We compose these into single actions because stateless only supports one per state.
	onExitActions := make(map[State][]func(context.Context, ...any) error)
	onEntryActions := make(map[State][]func(context.Context, ...any) error)

	// First pass: collect all actions and guard descriptions
	for _, t := range b.transitions {
		config := &transitionConfig{}
		for _, opt := range t.options {
			opt(config)
		}

		// Store guard description with full key (from, to, event)
		if config.guard != nil && config.guardDescription != "" {
			key := guardDescKey{from: t.from, to: t.to, event: t.event}
			b.guardDescriptions[key] = config.guardDescription
		}

		// Collect onExit actions for the source state
		if config.onExit != nil {
			onExitActions[t.from] = append(onExitActions[t.from], config.onExit)
		}

		// Collect onEntry actions for the target state
		if config.onEntry != nil {
			onEntryActions[t.to] = append(onEntryActions[t.to], config.onEntry)
		}
	}

	// Track which states have been configured
	configuredStates := make(map[State]bool)

	// Second pass: configure transitions and register composed actions
	for _, t := range b.transitions {
		config := &transitionConfig{}
		for _, opt := range t.options {
			opt(config)
		}

		cfgFrom := fsm.Configure(string(t.from))

		// Register composed onExit action once per source state
		if !configuredStates[t.from] {
			if actions := onExitActions[t.from]; len(actions) > 0 {
				cfgFrom.OnExit(composeActions(actions))
			}
			configuredStates[t.from] = true
		}

		// Configure the transition
		if config.guard != nil {
			guardFunc := func(_ context.Context, _ ...any) bool {
				return config.guard()
			}
			cfgFrom.Permit(stateless.Trigger(string(t.event)), string(t.to), guardFunc)
		} else {
			cfgFrom.Permit(stateless.Trigger(string(t.event)), string(t.to))
		}

		// Register composed onEntry action once per target state
		if !configuredStates[t.to] {
			if actions := onEntryActions[t.to]; len(actions) > 0 {
				cfgTo := fsm.Configure(string(t.to))
				cfgTo.OnEntry(composeActions(actions))
			}
			configuredStates[t.to] = true
		}
	}

	machine := NewMachine(fsm, b.promptFunc)
	b.setupUnhandledTriggerHandler(machine)

	return machine
}

// composeActions creates a single action that runs multiple actions in sequence.
// If any action returns an error, subsequent actions are not run.
func composeActions(actions []func(context.Context, ...any) error) func(context.Context, ...any) error {
	if len(actions) == 1 {
		return actions[0]
	}
	return func(ctx context.Context, args ...any) error {
		for _, action := range actions {
			if err := action(ctx, args...); err != nil {
				return err
			}
		}
		return nil
	}
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

		// Find guard descriptions for this (from, event) pair.
		// There may be multiple if this is a branching transition (same from+event, different to).
		// We collect all matching descriptions to provide a helpful error message.
		var descriptions []string
		for key, desc := range b.guardDescriptions {
			if key.from == currentState && key.event == event {
				descriptions = append(descriptions, desc)
			}
		}

		if len(descriptions) == 1 {
			return fmt.Errorf("guard '%s' failed for event '%s' from state '%s'", descriptions[0], event, currentState)
		}
		if len(descriptions) > 1 {
			return fmt.Errorf("guards failed for event '%s' from state '%s': %v", event, currentState, descriptions)
		}

		if len(unmetGuards) > 0 {
			return fmt.Errorf("guard conditions not met for event '%s' from state '%s': %v", event, currentState, unmetGuards)
		}

		return fmt.Errorf("trigger '%s' is not valid from state '%s'", event, currentState)
	})
}
