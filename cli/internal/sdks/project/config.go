package project

import (
	"fmt"
	"sort"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// PhaseConfig holds configuration for a single phase in a project type.
type PhaseConfig struct {
	// name is the phase identifier
	name string

	// startState is the state when the phase begins
	startState sdkstate.State

	// endState is the state when the phase ends
	endState sdkstate.State

	// allowedInputTypes are the artifact types allowed as inputs
	// Empty slice means all types are allowed
	allowedInputTypes []string

	// allowedOutputTypes are the artifact types allowed as outputs
	// Empty slice means all types are allowed
	allowedOutputTypes []string

	// supportsTasks indicates whether the phase can have tasks
	supportsTasks bool

	// metadataSchema is an embedded CUE schema for metadata validation
	metadataSchema string
}

// TransitionConfig holds configuration for a state machine transition.
type TransitionConfig struct {
	// From is the source state
	From sdkstate.State

	// To is the target state
	To sdkstate.State

	// Event is the event that triggers the transition
	Event sdkstate.Event

	// guardTemplate is a function template that becomes a bound guard
	guardTemplate GuardTemplate

	// onEntry is an action to execute when entering the target state
	onEntry Action

	// onExit is an action to execute when exiting the source state
	onExit Action

	// failedPhase optionally specifies a phase to mark as "failed" instead of "completed"
	// when exiting its end state on this transition. Used for error/failure paths.
	failedPhase string
}

// ProjectTypeConfig holds the complete configuration for a project type.
type ProjectTypeConfig struct {
	// name is the project type identifier
	name string

	// phaseConfigs are the phase configurations indexed by phase name
	phaseConfigs map[string]*PhaseConfig

	// initialState is the starting state of the state machine
	initialState sdkstate.State

	// transitions are all state transitions for this project type
	transitions []TransitionConfig

	// onAdvance are event determiners mapped by state
	// These determine which event to use for the Advance() command
	onAdvance map[sdkstate.State]EventDeterminer

	// prompts are prompt generators mapped by state
	// These generate contextual prompts for users in each state
	prompts map[sdkstate.State]PromptGenerator

	// orchestratorPrompt generates project-type-specific orchestrator guidance
	// This explains how the project type works and how orchestrator coordinates work
	orchestratorPrompt PromptGenerator

	// initializer is called during Create() to initialize the project
	// with phases, metadata, and any type-specific initial state
	initializer state.Initializer
}

// InitialState returns the configured initial state for this project type.
func (ptc *ProjectTypeConfig) InitialState() sdkstate.State {
	return ptc.initialState
}

// Initialize calls the configured initializer function if present.
// Returns nil if no initializer is configured.
func (ptc *ProjectTypeConfig) Initialize(project *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	if ptc.initializer == nil {
		return nil
	}
	return ptc.initializer(project, initialInputs)
}

// OrchestratorPrompt returns the orchestrator prompt for this project type.
// This explains how the project type works and how the orchestrator should coordinate work.
// Returns empty string if no orchestrator prompt is configured.
func (ptc *ProjectTypeConfig) OrchestratorPrompt(project *state.Project) string {
	if ptc.orchestratorPrompt == nil {
		return ""
	}
	return ptc.orchestratorPrompt(project)
}

// GetStatePrompt returns the prompt for a specific state.
// Returns empty string if no prompt is configured for the state.
func (ptc *ProjectTypeConfig) GetStatePrompt(state sdkstate.State, project *state.Project) string {
	gen, exists := ptc.prompts[state]
	if !exists {
		return ""
	}
	return gen(project)
}

// GetTaskSupportingPhases returns the names of all phases that support tasks.
// Returns an empty slice if no phases support tasks.
// Phase names are returned in sorted order for deterministic behavior.
func (ptc *ProjectTypeConfig) GetTaskSupportingPhases() []string {
	var phases []string
	for name, config := range ptc.phaseConfigs {
		if config.supportsTasks {
			phases = append(phases, name)
		}
	}
	// Sort for deterministic ordering
	sort.Strings(phases)
	return phases
}

// PhaseSupportsTasks checks if a specific phase supports tasks.
// Returns false if the phase doesn't exist or doesn't support tasks.
func (ptc *ProjectTypeConfig) PhaseSupportsTasks(phaseName string) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.supportsTasks
}

// GetDefaultTaskPhase returns the default phase for task operations based on current state.
// Returns empty string if no phase supports tasks or state mapping is ambiguous.
//
// Logic:
//  1. Check if current state maps to a phase's start or end state
//  2. If that phase supports tasks, return it
//  3. Otherwise return first task-supporting phase alphabetically
func (ptc *ProjectTypeConfig) GetDefaultTaskPhase(currentState sdkstate.State) string {
	// Try to map current state to a phase
	for name, config := range ptc.phaseConfigs {
		if (config.startState == currentState || config.endState == currentState) && config.supportsTasks {
			return name
		}
	}

	// Fallback: return first task-supporting phase
	phases := ptc.GetTaskSupportingPhases()
	if len(phases) > 0 {
		return phases[0]
	}
	return ""
}

// Validate validates project state against project type configuration.
//
// Performs two-tier validation:
//  1. Artifact type validation - Checks inputs/outputs against allowed types
//  2. Metadata validation - Validates metadata against embedded CUE schemas
//
// Returns error describing first validation failure found.
func (ptc *ProjectTypeConfig) Validate(project *state.Project) error {
	// Validate each phase
	for phaseName, phaseConfig := range ptc.phaseConfigs {
		phase, exists := project.Phases[phaseName]
		if !exists {
			// Phase not in state - skip (may be optional/future phase)
			continue
		}

		// Validate artifact types using state package helpers
		if err := state.ValidateArtifactTypes(
			phase.Inputs,
			phaseConfig.allowedInputTypes,
			phaseName,
			"input",
		); err != nil {
			return fmt.Errorf("validating inputs: %w", err)
		}

		if err := state.ValidateArtifactTypes(
			phase.Outputs,
			phaseConfig.allowedOutputTypes,
			phaseName,
			"output",
		); err != nil {
			return fmt.Errorf("validating outputs: %w", err)
		}

		// Validate metadata against embedded schema (if schema provided)
		// Phases without schemas can have arbitrary metadata
		if phaseConfig.metadataSchema != "" {
			if err := state.ValidateMetadata(
				phase.Metadata,
				phaseConfig.metadataSchema,
			); err != nil {
				return fmt.Errorf("phase %s metadata: %w", phaseName, err)
			}
		}
	}

	return nil
}

// DetermineEvent determines which event to fire from the current state.
// Returns the event to fire, or an error if no determiner is configured.
func (ptc *ProjectTypeConfig) DetermineEvent(project *state.Project) (sdkstate.Event, error) {
	currentState := sdkstate.State(project.Statechart.Current_state)
	determiner, exists := ptc.onAdvance[currentState]
	if !exists {
		return "", fmt.Errorf("no event determiner configured for state %s", currentState)
	}
	return determiner(project)
}

// GetPhaseForState returns the phase name that contains the given state.
// Returns empty string if the state doesn't belong to any phase's start or end state.
// If multiple phases have the same state (which shouldn't happen in a well-designed
// project type), returns the first match in iteration order.
func (ptc *ProjectTypeConfig) GetPhaseForState(state sdkstate.State) string {
	for name, config := range ptc.phaseConfigs {
		if config.startState == state || config.endState == state {
			return name
		}
	}
	return ""
}

// IsPhaseStartState checks if the given state is the startState of the specified phase.
// Returns false if the phase doesn't exist or the state doesn't match.
func (ptc *ProjectTypeConfig) IsPhaseStartState(phaseName string, state sdkstate.State) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.startState == state
}

// IsPhaseEndState checks if the given state is the endState of the specified phase.
// Returns false if the phase doesn't exist or the state doesn't match.
func (ptc *ProjectTypeConfig) IsPhaseEndState(phaseName string, state sdkstate.State) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.endState == state
}

// GetTransition looks up a transition config by from state, event, and to state.
// Returns nil if no matching transition is found.
func (ptc *ProjectTypeConfig) GetTransition(from, to sdkstate.State, event sdkstate.Event) *TransitionConfig {
	for i := range ptc.transitions {
		tc := &ptc.transitions[i]
		if tc.From == from && tc.To == to && tc.Event == event {
			return tc
		}
	}
	return nil
}
