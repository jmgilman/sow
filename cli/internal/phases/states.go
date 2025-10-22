package phases

// State represents a state in the project lifecycle state machine.
// These are shared across all phase implementations.
type State string

const (
	// NoProject indicates no active project exists in the repository.
	NoProject State = "NoProject"

	// DiscoveryDecision is the decision gate for whether discovery phase is needed.
	DiscoveryDecision State = "DiscoveryDecision"

	// DiscoveryActive indicates the discovery phase is in progress (subservient mode).
	DiscoveryActive State = "DiscoveryActive"

	// DesignDecision is the decision gate for whether design phase is needed.
	DesignDecision State = "DesignDecision"

	// DesignActive indicates the design phase is in progress (subservient mode).
	DesignActive State = "DesignActive"

	// ImplementationPlanning indicates the implementation phase planning step.
	ImplementationPlanning State = "ImplementationPlanning"

	// ImplementationExecuting indicates tasks are being executed (autonomous mode).
	ImplementationExecuting State = "ImplementationExecuting"

	// ReviewActive indicates the review phase is in progress (autonomous mode).
	ReviewActive State = "ReviewActive"

	// FinalizeDocumentation indicates the documentation update step of finalization.
	FinalizeDocumentation State = "FinalizeDocumentation"

	// FinalizeChecks indicates the final checks step of finalization.
	FinalizeChecks State = "FinalizeChecks"

	// FinalizeDelete indicates the project deletion step of finalization.
	FinalizeDelete State = "FinalizeDelete"
)

// String returns the string representation of the state.
func (s State) String() string {
	return string(s)
}

// IsSubservientMode returns true if the state requires subservient mode (human-led).
func (s State) IsSubservientMode() bool {
	switch s {
	case DiscoveryDecision, DiscoveryActive, DesignDecision, DesignActive:
		return true
	default:
		return false
	}
}

// IsAutonomousMode returns true if the state requires autonomous mode (AI-led).
func (s State) IsAutonomousMode() bool {
	switch s {
	case ImplementationPlanning, ImplementationExecuting, ReviewActive,
		FinalizeDocumentation, FinalizeChecks, FinalizeDelete:
		return true
	default:
		return false
	}
}
