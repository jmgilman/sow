package project

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// BranchConfig represents a state-determined branch point in the state machine.
// It captures a discriminator function that examines project state and returns
// a string value, plus a map of branch paths that define what happens for each value.
//
// BranchConfig is used internally by AddBranch to auto-generate transitions and
// event determiners. It's stored in ProjectTypeConfig for introspection.
type BranchConfig struct {
	from          sdkstate.State                           // Source state
	discriminator func(*state.Project) string              // Returns branch value
	branches      map[string]*BranchPath                   // value -> branch path
}

// BranchPath represents one possible branch destination.
// Each path maps a discriminator value to a transition configuration.
//
// Example: discriminator returns "pass" → fire EventReviewPass → go to FinalizeChecks
type BranchPath struct {
	value       string              // Discriminator value that triggers this path
	event       sdkstate.Event      // Event to fire
	to          sdkstate.State      // Target state
	description string              // Human-readable description

	// Standard transition configuration (forwarded from When options)
	guardTemplate GuardTemplate     // Optional guard (in addition to discriminator)
	onEntry      Action             // OnEntry action
	onExit       Action             // OnExit action
	failedPhase  string             // Phase to mark as failed
}

// BranchOption configures a BranchConfig.
type BranchOption func(*BranchConfig)

// BranchOn specifies the discriminator function for a branch configuration.
//
// The discriminator examines project state and returns a string value that
// determines which branch path to take. This value is matched against the
// values defined in When() clauses.
//
// The discriminator is called during Advance() to automatically determine
// which event to fire. It should be a pure function that examines project
// state but does not modify it.
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
// Standard transition options (WithGuard, WithOnEntry, WithDescription, etc.) can
// be passed to configure the generated transition.
//
// Parameters:
//   - value - Discriminator value to match (e.g., "pass", "fail", "staging")
//   - event - Event to fire when this branch is taken
//   - to - Target state for this branch
//   - opts - Standard TransitionOption functions
//
// Example:
//
//	When("pass",
//	    sdkstate.Event(EventReviewPass),
//	    sdkstate.State(FinalizeChecks),
//	    WithGuard("review passed", func(p *state.Project) bool {
//	        return getReviewAssessment(p) == "pass"
//	    }),
//	    WithDescription("Review approved - proceed to finalization"),
//	)
func When(
	value string,
	event sdkstate.Event,
	to sdkstate.State,
	opts ...TransitionOption,
) BranchOption {
	return func(bc *BranchConfig) {
		// Create BranchPath
		path := &BranchPath{
			value: value,
			event: event,
			to:    to,
		}

		// Apply transition options to extract configuration
		// This is a temporary TransitionConfig used to capture option values
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
