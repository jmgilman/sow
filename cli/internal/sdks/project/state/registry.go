package state

// Registry maps project type names to their configurations.
// This is a global registry that is populated during initialization.
// Full implementation will be provided in Unit 3 (SDK Builder).
var Registry = make(map[string]*ProjectTypeConfig)

// State represents a state machine state.
// This is a placeholder type that will be replaced with the actual
// state machine implementation in future tasks.
type State string

// String returns the string representation of the state.
func (s State) String() string {
	return string(s)
}

// BuildMachine builds a state machine for the project.
// This is a stub that will be implemented in Unit 3.
func (ptc *ProjectTypeConfig) BuildMachine(_ *Project, _ State) *Machine {
	// Stub - returns nil for now
	// Full implementation in Unit 3
	return nil
}

// Validate validates project metadata against embedded schemas.
// This is a stub that will be implemented in Unit 3.
func (ptc *ProjectTypeConfig) Validate(_ *Project) error {
	// Stub - no validation for now
	// Full implementation in Unit 3
	return nil
}
