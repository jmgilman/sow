package project

// TransitionConfig holds configuration for a state machine transition.
// It defines the source state, target state, triggering event, and
// optional guards, actions, and phase status handling.
type TransitionConfig struct {
	// From is the source state
	From State

	// To is the target state
	To State

	// Event is the event that triggers the transition
	Event Event

	// guardTemplate is a function template that becomes a bound guard
	guardTemplate GuardTemplate

	// onEntry is an action to execute when entering the target state
	onEntry Action

	// onExit is an action to execute when exiting the source state
	onExit Action

	// failedPhase optionally specifies a phase to mark as "failed" instead of "completed"
	// when exiting its end state on this transition. Used for error/failure paths.
	failedPhase string

	// description is a human-readable explanation of what this transition does.
	// Context-specific: same event from different states can have different meanings.
	description string
}

// GuardDescription returns the guard description.
// Returns empty string if no guard or guard has no description.
func (tc *TransitionConfig) GuardDescription() string {
	return tc.guardTemplate.Description
}

// Description returns the transition description.
// This provides a human-readable explanation of what this transition does.
func (tc *TransitionConfig) Description() string {
	return tc.description
}

// FailedPhase returns the phase name to mark as failed when this transition fires.
// Returns empty string if no phase should be marked as failed.
func (tc *TransitionConfig) FailedPhase() string {
	return tc.failedPhase
}
