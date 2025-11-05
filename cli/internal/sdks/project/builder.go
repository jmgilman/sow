// Package project provides a builder SDK for declaratively defining project types
// with phases, transitions, validation, and state machine integration.
package project

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// ProjectTypeConfigBuilder provides a fluent API for building ProjectTypeConfig instances.
// It accumulates configuration through chainable method calls and can be reused
// for building multiple configs via repeated Build() calls.
type ProjectTypeConfigBuilder struct {
	name               string
	phaseConfigs       map[string]*PhaseConfig
	initialState       sdkstate.State
	transitions        []TransitionConfig
	onAdvance          map[sdkstate.State]EventDeterminer
	prompts            map[sdkstate.State]PromptGenerator
	orchestratorPrompt PromptGenerator
	initializer        state.Initializer
}

// NewProjectTypeConfigBuilder creates a new builder for a project type config.
// The name parameter identifies the project type.
// All internal collections are initialized as empty but non-nil.
func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder {
	return &ProjectTypeConfigBuilder{
		name:         name,
		phaseConfigs: make(map[string]*PhaseConfig),
		transitions:  make([]TransitionConfig, 0),
		onAdvance:    make(map[sdkstate.State]EventDeterminer),
		prompts:      make(map[sdkstate.State]PromptGenerator),
	}
}

// WithPhase adds a phase configuration to the builder.
// The phase is created with the given name and all provided options are applied.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) WithPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder {
	pc := &PhaseConfig{
		name: name,
	}

	// Apply all options to the phase config
	for _, opt := range opts {
		opt(pc)
	}

	b.phaseConfigs[name] = pc
	return b
}

// SetInitialState sets the initial state for the state machine.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) SetInitialState(state sdkstate.State) *ProjectTypeConfigBuilder {
	b.initialState = state
	return b
}

// AddTransition adds a state transition to the builder.
// The transition is configured with from/to states, triggering event, and any provided options.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) AddTransition(
	from, to sdkstate.State,
	event sdkstate.Event,
	opts ...TransitionOption,
) *ProjectTypeConfigBuilder {
	tc := TransitionConfig{
		From:  from,
		To:    to,
		Event: event,
	}

	// Apply all options to the transition config
	for _, opt := range opts {
		opt(&tc)
	}

	b.transitions = append(b.transitions, tc)
	return b
}

// OnAdvance registers an event determiner function for the given state.
// The determiner is called during Advance() to determine which event to fire.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) OnAdvance(
	state sdkstate.State,
	determiner EventDeterminer,
) *ProjectTypeConfigBuilder {
	b.onAdvance[state] = determiner
	return b
}

// WithPrompt registers a prompt generator function for the given state.
// The generator creates contextual prompts for users in that state.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) WithPrompt(
	state sdkstate.State,
	generator PromptGenerator,
) *ProjectTypeConfigBuilder {
	b.prompts[state] = generator
	return b
}

// WithOrchestratorPrompt registers a prompt generator for orchestrator guidance.
// This prompt explains how the project type works and how the orchestrator
// should coordinate work through phases. It is shown when projects are created
// or continued, providing workflow-specific context.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) WithOrchestratorPrompt(
	generator PromptGenerator,
) *ProjectTypeConfigBuilder {
	b.orchestratorPrompt = generator
	return b
}

// WithInitializer registers an initializer function for project creation.
// The initializer is called during Create() to set up phases, metadata,
// and any project-type-specific initial state.
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) WithInitializer(
	initializer state.Initializer,
) *ProjectTypeConfigBuilder {
	b.initializer = initializer
	return b
}

// Build creates a new ProjectTypeConfig from the builder's accumulated state.
// The builder is NOT reset after calling Build(), allowing it to be reused
// for creating multiple configs or for building incrementally.
//
// All data is copied into the new config:
// - Maps are copied to new map instances
// - Slices are copied to new slice instances
// - Function references are copied (functions themselves are immutable).
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig {
	// Copy phase configs map
	phaseConfigs := make(map[string]*PhaseConfig, len(b.phaseConfigs))
	for k, v := range b.phaseConfigs {
		phaseConfigs[k] = v
	}

	// Copy transitions slice
	transitions := make([]TransitionConfig, len(b.transitions))
	copy(transitions, b.transitions)

	// Copy onAdvance map
	onAdvance := make(map[sdkstate.State]EventDeterminer, len(b.onAdvance))
	for k, v := range b.onAdvance {
		onAdvance[k] = v
	}

	// Copy prompts map
	prompts := make(map[sdkstate.State]PromptGenerator, len(b.prompts))
	for k, v := range b.prompts {
		prompts[k] = v
	}

	return &ProjectTypeConfig{
		name:               b.name,
		phaseConfigs:       phaseConfigs,
		initialState:       b.initialState,
		transitions:        transitions,
		onAdvance:          onAdvance,
		prompts:            prompts,
		orchestratorPrompt: b.orchestratorPrompt,
		initializer:        b.initializer,
	}
}
