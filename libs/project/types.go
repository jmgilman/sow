package project

import (
	"github.com/jmgilman/sow/libs/project/state"
)

// State represents a state in the project lifecycle state machine.
// States are string constants defined by project type configurations.
// All project types share the NoProject state for when no project exists.
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

// Event represents a trigger that causes state transitions.
// Events are defined by project type configurations and fired
// when advancing through the project lifecycle.
type Event string

// String returns the string representation of the event.
func (e Event) String() string {
	return string(e)
}

// Guard is a condition function that determines if a transition is allowed.
// Returns true if the transition should proceed, false otherwise.
type Guard func() bool

// GuardTemplate is a template function that gets bound to a project instance
// via closure. It receives the project and returns whether transition is allowed.
// The Description provides a human-readable explanation of what the guard checks,
// which appears in error messages when the guard fails.
//
// Note: The state.Project type is a placeholder that will be fully implemented
// in Task 040.
type GuardTemplate struct {
	Description string
	Func        func(*state.Project) bool
}

// Action is a function that mutates project state during transitions.
// Returns error if action fails.
//
// Note: The state.Project type is a placeholder that will be fully implemented
// in Task 040.
type Action func(*state.Project) error

// EventDeterminer examines project state and determines the next event
// for the generic Advance() command. Returns the event or error if unable
// to determine (e.g., missing required state).
//
// Note: The state.Project type is a placeholder that will be fully implemented
// in Task 040.
type EventDeterminer func(*state.Project) (Event, error)

// PromptGenerator creates a contextual prompt for a given state.
// Returns markdown-formatted string to display to user.
//
// Note: The state.Project type is a placeholder that will be fully implemented
// in Task 040.
type PromptGenerator func(*state.Project) string

// PromptFunc generates a prompt for a state during transitions.
// Used by the state machine builder.
type PromptFunc func(State) string
