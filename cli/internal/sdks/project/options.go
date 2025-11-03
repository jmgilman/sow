package project

import (
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// PhaseOpt is a function that modifies a PhaseConfig.
type PhaseOpt func(*PhaseConfig)

// WithStartState sets the phase start state.
func WithStartState(state sdkstate.State) PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.startState = state
	}
}

// WithEndState sets the phase end state.
func WithEndState(state sdkstate.State) PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.endState = state
	}
}

// WithInputs sets the allowed input artifact types for the phase.
func WithInputs(types ...string) PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.allowedInputTypes = types
	}
}

// WithOutputs sets the allowed output artifact types for the phase.
func WithOutputs(types ...string) PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.allowedOutputTypes = types
	}
}

// WithTasks enables task support for the phase.
func WithTasks() PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.supportsTasks = true
	}
}

// WithMetadataSchema sets the embedded CUE metadata schema for the phase.
func WithMetadataSchema(schema string) PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.metadataSchema = schema
	}
}

// TransitionOption is a function that modifies a TransitionConfig.
type TransitionOption func(*TransitionConfig)

// WithGuard sets the guard template function for a transition.
func WithGuard(guardFunc GuardTemplate) TransitionOption {
	return func(tc *TransitionConfig) {
		tc.guardTemplate = guardFunc
	}
}

// WithOnEntry sets the entry action for a transition.
func WithOnEntry(action Action) TransitionOption {
	return func(tc *TransitionConfig) {
		tc.onEntry = action
	}
}

// WithOnExit sets the exit action for a transition.
func WithOnExit(action Action) TransitionOption {
	return func(tc *TransitionConfig) {
		tc.onExit = action
	}
}
