package project

import (
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
}
