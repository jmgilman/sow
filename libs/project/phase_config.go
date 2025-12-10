package project

// PhaseConfig holds configuration for a single phase in a project type.
// It defines the phase's start and end states, allowed artifact types,
// and whether tasks are supported.
type PhaseConfig struct {
	// name is the phase identifier (e.g., "planning", "implementation")
	name string

	// startState is the state when the phase begins
	startState State

	// endState is the state when the phase ends
	endState State

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

// Name returns the phase name (e.g., "planning", "implementation").
func (pc *PhaseConfig) Name() string {
	return pc.name
}

// StartState returns the phase's start state.
func (pc *PhaseConfig) StartState() State {
	return pc.startState
}

// EndState returns the phase's end state.
func (pc *PhaseConfig) EndState() State {
	return pc.endState
}

// AllowedInputTypes returns the allowed input artifact types.
// Empty slice means all types are allowed.
func (pc *PhaseConfig) AllowedInputTypes() []string {
	return pc.allowedInputTypes
}

// AllowedOutputTypes returns the allowed output artifact types.
// Empty slice means all types are allowed.
func (pc *PhaseConfig) AllowedOutputTypes() []string {
	return pc.allowedOutputTypes
}

// SupportsTasks returns whether the phase supports task management.
func (pc *PhaseConfig) SupportsTasks() bool {
	return pc.supportsTasks
}

// MetadataSchema returns the CUE schema for phase metadata validation.
// Returns empty string if no schema is configured.
func (pc *PhaseConfig) MetadataSchema() string {
	return pc.metadataSchema
}
