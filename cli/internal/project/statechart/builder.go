package statechart

import (
	"context"
	"fmt"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/qmuntal/stateless"
)

// MachineBuilder provides a fluent API for constructing state machines.
// It enables project types to define their own state machines without duplicating
// common infrastructure patterns.
type MachineBuilder struct {
	sm              *stateless.StateMachine
	projectState    *schemas.ProjectState
	promptGenerator PromptGenerator
	suppressPrompts bool
}

// NewBuilder creates a new MachineBuilder starting at the specified initial state.
// The promptGenerator is used to generate contextual prompts for state entry actions.
//
// Example:
//
//	builder := statechart.NewBuilder(
//	    statechart.PlanningActive,
//	    projectState,
//	    NewStandardPromptGenerator(ctx),
//	)
func NewBuilder(
	initialState State,
	projectState *schemas.ProjectState,
	promptGenerator PromptGenerator,
) *MachineBuilder {
	sm := stateless.NewStateMachine(initialState)
	return &MachineBuilder{
		sm:              sm,
		projectState:    projectState,
		promptGenerator: promptGenerator,
		suppressPrompts: false,
	}
}

// TransitionOption configures a state transition.
type TransitionOption func(*transitionConfig)

// transitionConfig holds configuration for a single transition.
type transitionConfig struct {
	guard        GuardFunc
	entryActions []func(context.Context, ...any) error
	exitActions  []func(context.Context, ...any) error
}

// GuardFunc is a function that determines if a transition is permitted.
// It should return true if the transition should be allowed, false otherwise.
// Guards are closures that can capture project state and other context.
type GuardFunc func() bool

// WithGuard adds a guard condition to a transition.
// The transition will only be permitted if the guard returns true.
//
// Example:
//
//	builder.AddTransition(
//	    PlanningActive,
//	    ImplementationPlanning,
//	    EventCompletePlanning,
//	    statechart.WithGuard(func() bool {
//	        return PlanningComplete(project.state.Phases.Planning)
//	    }),
//	)
func WithGuard(guard GuardFunc) TransitionOption {
	return func(c *transitionConfig) {
		c.guard = guard
	}
}

// WithOnEntry adds a custom entry action to execute when entering the target state.
// Multiple entry actions can be added and will execute in the order they are added,
// after the built-in prompt generation action.
//
// Entry actions receive the context and can mutate shared state through closures.
// The project state will be saved after all entry actions complete.
//
// Example:
//
//	builder.AddTransition(
//	    PlanningActive,
//	    ImplementationPlanning,
//	    EventCompletePlanning,
//	    statechart.WithOnEntry(func(_ context.Context, _ ...any) error {
//	        project.state.Phases.Implementation.Status = "in_progress"
//	        return nil
//	    }),
//	)
func WithOnEntry(action func(context.Context, ...any) error) TransitionOption {
	return func(c *transitionConfig) {
		c.entryActions = append(c.entryActions, action)
	}
}

// WithOnExit adds a custom exit action to execute when leaving the source state.
// Multiple exit actions can be added and will execute in the order they are added,
// before the state transition occurs.
//
// Exit actions receive the context and can mutate shared state through closures.
// The project state will be saved after the transition completes.
//
// Example:
//
//	builder.AddTransition(
//	    PlanningActive,
//	    ImplementationPlanning,
//	    EventCompletePlanning,
//	    statechart.WithOnExit(func(_ context.Context, _ ...any) error {
//	        project.state.Phases.Planning.Status = "completed"
//	        return nil
//	    }),
//	)
func WithOnExit(action func(context.Context, ...any) error) TransitionOption {
	return func(c *transitionConfig) {
		c.exitActions = append(c.exitActions, action)
	}
}

// AddTransition adds a state transition with optional configuration.
// Transitions without a guard (no WithGuard option) are unconditional.
//
// Example unconditional transition:
//
//	builder.AddTransition(
//	    NoProject,
//	    PlanningActive,
//	    EventProjectInit,
//	)
//
// Example conditional transition:
//
//	builder.AddTransition(
//	    PlanningActive,
//	    ImplementationPlanning,
//	    EventCompletePlanning,
//	    statechart.WithGuard(func() bool {
//	        return PlanningComplete(project.state.Phases.Planning)
//	    }),
//	)
func (b *MachineBuilder) AddTransition(
	from, to State,
	event Event,
	opts ...TransitionOption,
) *MachineBuilder {
	// Apply options
	config := &transitionConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Configure the source state for exit actions
	cfgFrom := b.sm.Configure(from)

	// Add user-defined exit actions (run before leaving source state)
	for _, action := range config.exitActions {
		cfgFrom.OnExit(action)
	}

	// Configure the transition
	if config.guard != nil {
		// Conditional transition with guard
		// Wrap guard to match stateless signature
		cfgFrom.Permit(event, to, func(_ context.Context, _ ...any) bool {
			return config.guard()
		})
	} else {
		// Unconditional transition
		cfgFrom.Permit(event, to)
	}

	// Configure the target state for entry actions
	cfgTo := b.sm.Configure(to)

	// Add built-in prompt generation entry action
	cfgTo.OnEntry(b.onEntry(from))

	// Add user-defined entry actions (run after entering target state)
	for _, action := range config.entryActions {
		cfgTo.OnEntry(action)
	}

	return b
}

// ConfigureState provides direct access to the stateless StateConfiguration
// for advanced use cases not covered by AddTransition.
//
// Example:
//
//	builder.ConfigureState(ReviewActive).
//	    OnExit(func(ctx context.Context, args ...any) error {
//	        // Custom exit action
//	        return nil
//	    })
func (b *MachineBuilder) ConfigureState(state State) *stateless.StateConfiguration {
	return b.sm.Configure(state)
}

// SuppressPrompts disables prompt generation for all state entry actions.
// This is useful for tests and non-interactive CLI commands.
func (b *MachineBuilder) SuppressPrompts(suppress bool) *MachineBuilder {
	b.suppressPrompts = suppress
	return b
}

// Build creates the final Machine instance with all configured transitions.
// This should be called after all transitions have been added.
//
// Example:
//
//	machine := builder.
//	    AddTransition(NoProject, PlanningActive, EventProjectInit).
//	    AddTransition(PlanningActive, ImplementationPlanning, EventCompletePlanning, WithGuard(...)).
//	    Build()
func (b *MachineBuilder) Build() *Machine {
	return &Machine{
		sm:              b.sm,
		projectState:    b.projectState,
		suppressPrompts: b.suppressPrompts,
	}
}

// onEntry creates an entry action that generates and prints a contextual prompt.
func (b *MachineBuilder) onEntry(state State) func(context.Context, ...any) error {
	return func(_ context.Context, _ ...any) error {
		// Skip entirely if prompts are suppressed
		if b.suppressPrompts {
			return nil
		}

		prompt, err := b.promptGenerator.GeneratePrompt(state, b.projectState)
		if err != nil {
			return fmt.Errorf("failed to generate prompt for state %s: %w", state, err)
		}

		fmt.Println(prompt)
		return nil
	}
}
