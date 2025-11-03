package state

// State represents a state in the project lifecycle state machine.
type State string

const (
	// NoProject indicates no active project exists in the repository.
	// This is a shared state used by all project types.
	NoProject State = "NoProject"
)

// String returns the string representation of the state.
func (s State) String() string {
	return string(s)
}
