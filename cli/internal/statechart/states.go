package statechart

// State represents a state in the project lifecycle state machine.
type State string

const (
	// NoProject indicates no active project exists in the repository.
	NoProject State = "NoProject"

	// DiscoveryDecision is the decision gate for whether discovery phase is needed.
	// The orchestrator applies rubrics or asks the user to determine if discovery is warranted.
	DiscoveryDecision State = "DiscoveryDecision"

	// DiscoveryActive indicates the discovery phase is in progress (subservient mode).
	// The orchestrator facilitates research, creates artifacts, and gets human approval.
	DiscoveryActive State = "DiscoveryActive"

	// DesignDecision is the decision gate for whether design phase is needed.
	// The orchestrator applies rubrics or asks the user to determine if design is warranted.
	DesignDecision State = "DesignDecision"

	// DesignActive indicates the design phase is in progress (subservient mode).
	// The orchestrator facilitates design alignment, creates artifacts, and gets human approval.
	DesignActive State = "DesignActive"

	// ImplementationPlanning indicates the implementation phase planning step.
	// The orchestrator (or planner agent) creates the task breakdown.
	ImplementationPlanning State = "ImplementationPlanning"

	// ImplementationExecuting indicates tasks are being executed (autonomous mode).
	// The orchestrator spawns implementer agents to work on tasks.
	ImplementationExecuting State = "ImplementationExecuting"

	// ReviewActive indicates the review phase is in progress (autonomous mode).
	// The orchestrator or reviewer agent validates the implementation.
	ReviewActive State = "ReviewActive"

	// FinalizeDocumentation indicates the documentation update step of finalization.
	// The orchestrator checks if documentation needs updates and performs them.
	FinalizeDocumentation State = "FinalizeDocumentation"

	// FinalizeChecks indicates the final checks step of finalization.
	// The orchestrator runs tests, linters, and other validation checks.
	FinalizeChecks State = "FinalizeChecks"

	// FinalizeDelete indicates the project deletion step of finalization.
	// The orchestrator must delete the project folder before completing.
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
