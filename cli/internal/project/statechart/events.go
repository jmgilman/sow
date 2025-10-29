// Package statechart implements a state machine for project lifecycle management.
package statechart

// Event represents a trigger that causes state transitions.
// Individual project types define their own event constants.
type Event string

// String returns the string representation of the event.
func (e Event) String() string {
	return string(e)
}
