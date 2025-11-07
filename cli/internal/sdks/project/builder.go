// Package project provides a builder SDK for declaratively defining project types
// with phases, transitions, validation, and state machine integration.
package project

import (
	"fmt"
	"sort"
	"strings"

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
	branches           map[sdkstate.State]*BranchConfig
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
		branches:     make(map[sdkstate.State]*BranchConfig),
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

// AddBranch configures state-determined branching from a state.
//
// This method provides a declarative API for defining multi-way branches where
// the next state can be determined by examining project state. It auto-generates:
//   1. Transitions for each When clause
//   2. Event determiner using the discriminator function
//
// The discriminator examines project state and returns a string value. This value
// is matched against the values defined in When() clauses to determine which
// event to fire and which state to transition to.
//
// Example (binary branch):
//   AddBranch(
//       sdkstate.State(ReviewActive),
//       project.BranchOn(func(p *state.Project) string {
//           // Get review assessment from latest approved review
//           assessment := getReviewAssessment(p)
//           return assessment  // "pass" or "fail"
//       }),
//       project.When("pass",
//           sdkstate.Event(EventReviewPass),
//           sdkstate.State(FinalizeChecks),
//           project.WithDescription("Review approved - proceed to finalization"),
//       ),
//       project.When("fail",
//           sdkstate.Event(EventReviewFail),
//           sdkstate.State(ImplementationPlanning),
//           project.WithDescription("Review failed - return to planning for rework"),
//           project.WithFailedPhase("review"),
//       ),
//   )
//
// This auto-generates:
//   - Two AddTransition calls (one per When clause)
//   - One OnAdvance determiner that calls discriminator and maps value to event
//
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) AddBranch(
	from sdkstate.State,
	opts ...BranchOption,
) *ProjectTypeConfigBuilder {
	// Create BranchConfig from options
	bc := &BranchConfig{
		from: from,
	}

	// Apply all options to build the branch config
	for _, opt := range opts {
		opt(bc)
	}

	// Validation 1: Discriminator is required
	if bc.discriminator == nil {
		panic(fmt.Sprintf(
			"AddBranch for state %s: no discriminator provided - use BranchOn() to specify discriminator function",
			from))
	}

	// Validation 2: At least one branch path is required
	if len(bc.branches) == 0 {
		panic(fmt.Sprintf(
			"AddBranch for state %s: no branch paths provided - use When() to define at least one branch path",
			from))
	}

	// Validation 3: Warn if state already has OnAdvance
	if _, exists := b.onAdvance[from]; exists {
		panic(fmt.Sprintf(
			"AddBranch for state %s: state already has OnAdvance determiner - cannot use both AddBranch and OnAdvance on the same state",
			from))
	}

	// Validation 4: Empty discriminator values not allowed
	for value := range bc.branches {
		if value == "" {
			panic(fmt.Sprintf(
				"AddBranch for state %s: empty string is not allowed as a discriminator value",
				from))
		}
	}

	// Generate AddTransition calls for each branch path
	// Sort by value for deterministic transition order
	values := make([]string, 0, len(bc.branches))
	for value := range bc.branches {
		values = append(values, value)
	}
	sort.Strings(values)

	for _, value := range values {
		path := bc.branches[value]

		// Collect transition options from branch path
		var transOpts []TransitionOption

		if path.description != "" {
			transOpts = append(transOpts, WithDescription(path.description))
		}
		if path.guardTemplate.Func != nil {
			transOpts = append(transOpts, WithGuard(path.guardTemplate.Description, path.guardTemplate.Func))
		}
		if path.onEntry != nil {
			transOpts = append(transOpts, WithOnEntry(path.onEntry))
		}
		if path.onExit != nil {
			transOpts = append(transOpts, WithOnExit(path.onExit))
		}
		if path.failedPhase != "" {
			transOpts = append(transOpts, WithFailedPhase(path.failedPhase))
		}

		// Generate transition
		b.AddTransition(from, path.to, path.event, transOpts...)
	}

	// Generate OnAdvance determiner
	b.OnAdvance(from, func(p *state.Project) (sdkstate.Event, error) {
		// Call discriminator to get branch value
		value := bc.discriminator(p)

		// Look up branch path for this value
		path, exists := bc.branches[value]
		if !exists {
			// Build helpful error message with available values
			availableValues := make([]string, 0, len(bc.branches))
			for v := range bc.branches {
				availableValues = append(availableValues, fmt.Sprintf("%q", v))
			}

			// Sort for deterministic output
			sort.Strings(availableValues)

			return "", fmt.Errorf(
				"no branch defined for discriminator value %q from state %s (available values: %s)",
				value, from, strings.Join(availableValues, ", "))
		}

		// Return the event for this branch
		return path.event, nil
	})

	// Store branch config for introspection
	b.branches[from] = bc

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

	// Copy branches map
	branches := make(map[sdkstate.State]*BranchConfig, len(b.branches))
	for k, v := range b.branches {
		branches[k] = v
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
		branches:           branches,
	}
}
