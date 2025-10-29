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
	guard GuardFunc
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

	// Configure the transition
	cfg := b.sm.Configure(from)

	if config.guard != nil {
		// Conditional transition with guard
		// Wrap guard to match stateless signature
		cfg.Permit(event, to, func(_ context.Context, _ ...any) bool {
			return config.guard()
		})
	} else {
		// Unconditional transition
		cfg.Permit(event, to)
	}

	// Add entry action to generate prompt on state entry
	cfg.OnEntry(b.onEntry(from))

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
