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
// This method does not validate the configuration. For validation,
// use BuildWithValidation which returns an error for invalid configurations.
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig {
	config, _ := b.build()
	return config
}

// build is the internal implementation that creates the config without validation.
func (b *ProjectTypeConfigBuilder) build() (*ProjectTypeConfig, error) {
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
	}, nil
}

// BuildWithValidation creates the ProjectTypeConfig from the builder configuration.
// It returns an error if validation fails.
//
// Validation checks:
//   - Initial state must be set
//   - Phases with start/end states must have both set
//   - Transitions must reference states that belong to configured phases (if phases are defined)
func (b *ProjectTypeConfigBuilder) BuildWithValidation() (*ProjectTypeConfig, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	return b.build()
}

// validate checks the builder configuration for common mistakes.
func (b *ProjectTypeConfigBuilder) validate() error {
	var issues []string

	// Validate initial state is set
	if b.initialState == "" {
		issues = append(issues, "initial state not set (use SetInitialState)")
	}

	// Validate phase configurations
	issues = b.validatePhases(issues)

	// Collect all states from phases for transition validation
	phaseStates := b.collectPhaseStates()

	// Validate transitions and initial state reference known phase states
	issues = b.validateTransitions(phaseStates, issues)

	if len(issues) > 0 {
		return &ErrConfigValidation{Issues: issues}
	}

	return nil
}

// validatePhases checks that phases have both start and end states if either is set.
func (b *ProjectTypeConfigBuilder) validatePhases(issues []string) []string {
	for name, phase := range b.phaseConfigs {
		if phase.startState != "" && phase.endState == "" {
			issues = append(issues, "phase \""+name+"\" has start state but no end state")
		}
		if phase.endState != "" && phase.startState == "" {
			issues = append(issues, "phase \""+name+"\" has end state but no start state")
		}
	}
	return issues
}

// collectPhaseStates returns a map of all states to their owning phase names.
func (b *ProjectTypeConfigBuilder) collectPhaseStates() map[State]string {
	phaseStates := make(map[State]string)
	for name, phase := range b.phaseConfigs {
		if phase.startState != "" {
			phaseStates[phase.startState] = name
		}
		if phase.endState != "" {
			phaseStates[phase.endState] = name
		}
	}
	return phaseStates
}

// validateTransitions checks that transitions reference known states.
func (b *ProjectTypeConfigBuilder) validateTransitions(phaseStates map[State]string, issues []string) []string {
	// Only validate if phases are defined
	if len(phaseStates) == 0 {
		return issues
	}

	for _, tc := range b.transitions {
		issues = b.validateTransitionState(tc.From, "from", phaseStates, issues)
		issues = b.validateTransitionState(tc.To, "to", phaseStates, issues)
	}

	// Validate initial state is in a phase
	if b.initialState != "" && b.initialState != NoProject {
		if _, ok := phaseStates[b.initialState]; !ok {
			issues = append(issues, "initial state \""+string(b.initialState)+"\" not in any phase")
		}
	}

	return issues
}

// validateTransitionState checks if a state is known in the phase configuration.
func (b *ProjectTypeConfigBuilder) validateTransitionState(
	s State,
	direction string,
	phaseStates map[State]string,
	issues []string,
) []string {
	// Skip empty states and NoProject sentinel
	if s == "" || s == NoProject {
		return issues
	}
	if _, ok := phaseStates[s]; !ok {
		issues = append(issues, "transition "+direction+" state \""+string(s)+"\" not in any phase")
	}
	return issues
}
