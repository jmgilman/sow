package domain

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// PhaseOperationResult represents the outcome of a phase operation.
// It encapsulates an optional event that should be fired after the operation completes.
// This enables phases to declaratively trigger state machine transitions from CLI operations
// while keeping the CLI layer generic across all project types.
type PhaseOperationResult struct {
	// Event is the optional state machine event to fire after the operation.
	// Empty string indicates no event should be fired.
	Event statechart.Event
}

// NoEvent returns a PhaseOperationResult with no event to fire.
// Use this when a phase operation completes successfully but should not
// trigger a state machine transition.
//
// Example:
//
//	func (p *ReviewPhase) ApproveArtifact(path string) (*PhaseOperationResult, error) {
//	    if err := p.artifacts.Approve(path); err != nil {
//	        return nil, err
//	    }
//	    // Individual artifact approval doesn't trigger transitions
//	    return NoEvent(), nil
//	}
func NoEvent() *PhaseOperationResult {
	return &PhaseOperationResult{
		Event: "",
	}
}

// WithEvent returns a PhaseOperationResult that will fire the given event.
// Use this when a phase operation should trigger a state machine transition
// after completing successfully.
//
// Example:
//
//	func (p *PlanningPhase) Complete() (*PhaseOperationResult, error) {
//	    p.state.Status = "completed"
//	    if err := p.project.Save(); err != nil {
//	        return nil, err
//	    }
//	    // Completion triggers transition to implementation
//	    return WithEvent(statechart.EventCompletePlanning), nil
//	}
func WithEvent(event statechart.Event) *PhaseOperationResult {
	return &PhaseOperationResult{
		Event: event,
	}
}
