package project

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// GuardTemplate is a template function that gets bound to a project instance
// via closure. It receives the project and returns whether transition is allowed.
// The Description provides a human-readable explanation of what the guard checks,
// which appears in error messages when the guard fails.
type GuardTemplate struct {
	Description string
	Func        func(*state.Project) bool
}

// Action is a function that mutates project state during transitions.
// Returns error if action fails.
type Action func(*state.Project) error

// EventDeterminer examines project state and determines the next event
// for the generic Advance() command. Returns the event or error if unable
// to determine (e.g., missing required state).
type EventDeterminer func(*state.Project) (sdkstate.Event, error)

// PromptGenerator creates a contextual prompt for a given state.
// Returns markdown-formatted string to display to user.
type PromptGenerator func(*state.Project) string
