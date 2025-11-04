package project

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
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
func (ptc *ProjectTypeConfig) Initialize(project *state.Project) error {
	if ptc.initializer == nil {
		return nil
	}
	return ptc.initializer(project)
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
	if len(phases) > 1 {
		for i := 0; i < len(phases)-1; i++ {
			for j := i + 1; j < len(phases); j++ {
				if phases[i] > phases[j] {
					phases[i], phases[j] = phases[j], phases[i]
				}
			}
		}
	}
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
