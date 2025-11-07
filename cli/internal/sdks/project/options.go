package project

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
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
// The description should explain what condition must be met for the transition,
// and will appear in error messages when the guard fails. This helps orchestrators
// understand what action is needed.
//
// Example:
//
//	WithGuard("all tasks complete", func(p *state.Project) bool {
//	    return allTasksComplete(p)
//	})
func WithGuard(description string, guardFunc func(*state.Project) bool) TransitionOption {
	return func(tc *TransitionConfig) {
		tc.guardTemplate = GuardTemplate{
			Description: description,
			Func:        guardFunc,
		}
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

// WithFailedPhase marks a phase as "failed" instead of "completed" when exiting
// its end state on this transition. Used for error/failure paths where a phase
// should be marked as failed rather than successfully completed.
//
// Example:
//
//	AddTransition(
//	    ReviewActive,
//	    ImplementationPlanning,
//	    EventReviewFail,
//	    WithFailedPhase("review"), // Mark review as failed, not completed
//	)
func WithFailedPhase(phaseName string) TransitionOption {
	return func(tc *TransitionConfig) {
		tc.failedPhase = phaseName
	}
}

// WithDescription adds a human-readable description to a transition.
//
// Descriptions are:
// - Context-specific (same event from different states can have different meanings)
// - Co-located with guards and actions
// - Used by CLI --list to show what each transition does
// - Visible to orchestrators for decision-making
//
// Best practice: Always add descriptions for transitions, especially in branching states.
//
// Example:
//
//	WithDescription("Review approved - proceed to finalization")
func WithDescription(description string) TransitionOption {
	return func(tc *TransitionConfig) {
		tc.description = description
	}
}
