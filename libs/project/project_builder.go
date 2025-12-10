package project

import (
	"github.com/jmgilman/sow/libs/project/state"
)

// ProjectTypeConfigBuilder provides a fluent API for building project type configurations.
// It allows defining phases, transitions, prompts, and initialization logic.
//
//nolint:revive // "Project" prefix distinguishes this from generic machine builders
type ProjectTypeConfigBuilder struct {
	name               string
	initialState       State
	phaseConfigs       map[string]*PhaseConfig
	transitions        []TransitionConfig
	onAdvance          map[State]EventDeterminer
	prompts            map[State]PromptGenerator
	orchestratorPrompt PromptGenerator
	initializer        Initializer
	branches           map[State]*BranchConfig
}

// NewProjectTypeConfigBuilder creates a new builder for a project type configuration.
func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder {
	return &ProjectTypeConfigBuilder{
		name:         name,
		phaseConfigs: make(map[string]*PhaseConfig),
		onAdvance:    make(map[State]EventDeterminer),
		prompts:      make(map[State]PromptGenerator),
		branches:     make(map[State]*BranchConfig),
	}
}

// SetInitialState sets the initial state for new projects of this type.
func (b *ProjectTypeConfigBuilder) SetInitialState(state State) *ProjectTypeConfigBuilder {
	b.initialState = state
	return b
}

// WithPhase adds a phase configuration with the given name and options.
func (b *ProjectTypeConfigBuilder) WithPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder {
	pc := &PhaseConfig{name: name}
	for _, opt := range opts {
		opt(pc)
	}
	b.phaseConfigs[name] = pc
	return b
}

// AddTransition adds a state transition with optional configuration.
func (b *ProjectTypeConfigBuilder) AddTransition(from, to State, event Event, opts ...ProjectTransitionOption) *ProjectTypeConfigBuilder {
	tc := TransitionConfig{
		From:  from,
		To:    to,
		Event: event,
	}
	for _, opt := range opts {
		opt(&tc)
	}
	b.transitions = append(b.transitions, tc)
	return b
}

// OnAdvance sets an event determiner for a state.
// The determiner is called during Advance() to automatically select the event to fire.
func (b *ProjectTypeConfigBuilder) OnAdvance(state State, determiner EventDeterminer) *ProjectTypeConfigBuilder {
	b.onAdvance[state] = determiner
	return b
}

// AddBranch adds a state-determined branch point.
// Use BranchOn to specify the discriminator and When to define branch paths.
func (b *ProjectTypeConfigBuilder) AddBranch(from State, opts ...BranchOption) *ProjectTypeConfigBuilder {
	bc := &BranchConfig{
		from:     from,
		branches: make(map[string]*BranchPath),
	}
	for _, opt := range opts {
		opt(bc)
	}

	// Store the branch config
	b.branches[from] = bc

	// Generate transitions from branch paths
	for _, path := range bc.branches {
		tc := TransitionConfig{
			From:          from,
			To:            path.to,
			Event:         path.event,
			description:   path.description,
			guardTemplate: path.guardTemplate,
			onEntry:       path.onEntry,
			onExit:        path.onExit,
			failedPhase:   path.failedPhase,
		}
		b.transitions = append(b.transitions, tc)
	}

	// Set up event determiner for branching state
	if bc.discriminator != nil {
		b.onAdvance[from] = func(p *state.Project) (Event, error) {
			value := bc.discriminator(p)
			if path, exists := bc.branches[value]; exists {
				return path.event, nil
			}
			return "", &ErrBranchNotFound{State: from, Value: value}
		}
	}

	return b
}

// WithPrompt sets a prompt generator for a state.
func (b *ProjectTypeConfigBuilder) WithPrompt(state State, gen PromptGenerator) *ProjectTypeConfigBuilder {
	b.prompts[state] = gen
	return b
}

// WithOrchestratorPrompt sets the orchestrator prompt generator.
func (b *ProjectTypeConfigBuilder) WithOrchestratorPrompt(gen PromptGenerator) *ProjectTypeConfigBuilder {
	b.orchestratorPrompt = gen
	return b
}

// WithInitializer sets the project initializer function.
func (b *ProjectTypeConfigBuilder) WithInitializer(init Initializer) *ProjectTypeConfigBuilder {
	b.initializer = init
	return b
}

// Build creates the ProjectTypeConfig from the builder configuration.
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig {
	// Copy maps to ensure builder can be reused
	phaseConfigs := make(map[string]*PhaseConfig, len(b.phaseConfigs))
	for k, v := range b.phaseConfigs {
		phaseConfigs[k] = v
	}

	onAdvance := make(map[State]EventDeterminer, len(b.onAdvance))
	for k, v := range b.onAdvance {
		onAdvance[k] = v
	}

	prompts := make(map[State]PromptGenerator, len(b.prompts))
	for k, v := range b.prompts {
		prompts[k] = v
	}

	branches := make(map[State]*BranchConfig, len(b.branches))
	for k, v := range b.branches {
		branches[k] = v
	}

	// Copy transitions slice
	transitions := make([]TransitionConfig, len(b.transitions))
	copy(transitions, b.transitions)

	return &ProjectTypeConfig{
		name:               b.name,
		initialState:       b.initialState,
		phaseConfigs:       phaseConfigs,
		transitions:        transitions,
		onAdvance:          onAdvance,
		prompts:            prompts,
		orchestratorPrompt: b.orchestratorPrompt,
		initializer:        b.initializer,
		branches:           branches,
	}
}
