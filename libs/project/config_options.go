package project

import (
	"github.com/jmgilman/sow/libs/project/state"
)

// PhaseOpt is a function that modifies a PhaseConfig.
type PhaseOpt func(*PhaseConfig)

// WithStartState sets the phase start state.
func WithStartState(state State) PhaseOpt {
	return func(pc *PhaseConfig) {
		pc.startState = state
	}
}

// WithEndState sets the phase end state.
func WithEndState(state State) PhaseOpt {
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

// ProjectTransitionOption is a function that modifies a TransitionConfig.
// Named "ProjectTransitionOption" to distinguish from the base machine TransitionOption.
//
//nolint:revive // "Project" prefix required to distinguish from base TransitionOption
type ProjectTransitionOption func(*TransitionConfig)

// WithProjectGuard sets the guard template function for a transition.
// The description should explain what condition must be met for the transition,
// and will appear in error messages when the guard fails.
//
// Example:
//
//	WithProjectGuard("all tasks complete", func(p *state.Project) bool {
//	    return p.AllTasksComplete()
//	})
func WithProjectGuard(description string, guardFunc func(*state.Project) bool) ProjectTransitionOption {
	return func(tc *TransitionConfig) {
		tc.guardTemplate = GuardTemplate{
			Description: description,
			Func:        guardFunc,
		}
	}
}

// WithProjectOnEntry sets the entry action for a transition.
func WithProjectOnEntry(action Action) ProjectTransitionOption {
	return func(tc *TransitionConfig) {
		tc.onEntry = action
	}
}

// WithProjectOnExit sets the exit action for a transition.
func WithProjectOnExit(action Action) ProjectTransitionOption {
	return func(tc *TransitionConfig) {
		tc.onExit = action
	}
}

// WithProjectFailedPhase marks a phase as "failed" instead of "completed" when exiting
// its end state on this transition. Used for error/failure paths.
//
// Example:
//
//	AddTransition(
//	    ReviewActive,
//	    ImplementationPlanning,
//	    EventReviewFail,
//	    WithProjectFailedPhase("review"), // Mark review as failed, not completed
//	)
func WithProjectFailedPhase(phaseName string) ProjectTransitionOption {
	return func(tc *TransitionConfig) {
		tc.failedPhase = phaseName
	}
}

// WithProjectDescription adds a human-readable description to a transition.
//
// Descriptions are:
// - Context-specific (same event from different states can have different meanings)
// - Co-located with guards and actions
// - Used by CLI --list to show what each transition does
// - Visible to orchestrators for decision-making
//
// Example:
//
//	WithProjectDescription("Review approved - proceed to finalization")
func WithProjectDescription(description string) ProjectTransitionOption {
	return func(tc *TransitionConfig) {
		tc.description = description
	}
}

// BranchOption configures a BranchConfig.
type BranchOption func(*BranchConfig)

// BranchOn specifies the discriminator function for a branch configuration.
//
// The discriminator examines project state and returns a string value that
// determines which branch path to take. This value is matched against the
// values defined in When() clauses.
//
// Example:
//
//	BranchOn(func(p *state.Project) string {
//	    // Get review assessment from latest approved review artifact
//	    phase := p.Phases["review"]
//	    for i := len(phase.Outputs) - 1; i >= 0; i-- {
//	        artifact := phase.Outputs[i]
//	        if artifact.Type == "review" && artifact.Approved {
//	            if assessment, ok := artifact.Metadata["assessment"].(string); ok {
//	                return assessment  // "pass" or "fail"
//	            }
//	        }
//	    }
//	    return ""  // No approved review yet
//	})
func BranchOn(discriminator func(*state.Project) string) BranchOption {
	return func(bc *BranchConfig) {
		bc.discriminator = discriminator
	}
}

// When defines a branch path based on a discriminator value.
//
// Each When clause creates one possible branch destination. When the discriminator
// returns the specified value, the corresponding event is fired and the state
// machine transitions to the target state.
//
// Standard transition options (WithProjectGuard, WithProjectOnEntry, WithProjectDescription, etc.)
// can be passed to configure the generated transition.
//
// Parameters:
//   - value - Discriminator value to match (e.g., "pass", "fail", "staging")
//   - event - Event to fire when this branch is taken
//   - to - Target state for this branch
//   - opts - ProjectTransitionOption functions
//
// Example:
//
//	When("pass",
//	    EventReviewPass,
//	    FinalizeChecks,
//	    WithProjectGuard("review passed", func(p *state.Project) bool {
//	        return getReviewAssessment(p) == "pass"
//	    }),
//	    WithProjectDescription("Review approved - proceed to finalization"),
//	)
func When(value string, event Event, to State, opts ...ProjectTransitionOption) BranchOption {
	return func(bc *BranchConfig) {
		// Create BranchPath
		path := &BranchPath{
			value: value,
			event: event,
			to:    to,
		}

		// Apply transition options to extract configuration
		tc := TransitionConfig{}
		for _, opt := range opts {
			opt(&tc)
		}

		// Copy transition config fields to branch path
		path.description = tc.description
		path.guardTemplate = tc.guardTemplate
		path.onEntry = tc.onEntry
		path.onExit = tc.onExit
		path.failedPhase = tc.failedPhase

		// Initialize branches map if needed
		if bc.branches == nil {
			bc.branches = make(map[string]*BranchPath)
		}

		// Store branch path (last one wins if duplicate value)
		bc.branches[value] = path
	}
}
