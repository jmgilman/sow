package breakdown

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// TestEventsAreCorrectType verifies all event constants use state.Event type.
func TestEventsAreCorrectType(t *testing.T) {
	tests := []struct {
		name  string
		event state.Event
	}{
		{"EventBeginPublishing", EventBeginPublishing},
		{"EventCompleteBreakdown", EventCompleteBreakdown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the constant can be assigned to state.Event type
			e := tt.event
			if e == "" {
				t.Errorf("Event %s should not be empty", tt.name)
			}
		})
	}
}

// TestEventValues verifies event constants have correct string values.
func TestEventValues(t *testing.T) {
	tests := []struct {
		name     string
		event    state.Event
		expected string
	}{
		{"EventBeginPublishing", EventBeginPublishing, "begin_publishing"},
		{"EventCompleteBreakdown", EventCompleteBreakdown, "complete_breakdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("Expected event value %q, got %q", tt.expected, string(tt.event))
			}
		})
	}
}

// TestAllEventsAreDifferent verifies no duplicate event values.
func TestAllEventsAreDifferent(t *testing.T) {
	events := []state.Event{
		EventBeginPublishing,
		EventCompleteBreakdown,
	}
	seen := make(map[state.Event]bool)

	for _, e := range events {
		if seen[e] {
			t.Errorf("Duplicate event found: %s", e)
		}
		seen[e] = true
	}

	if len(seen) != 2 {
		t.Errorf("Expected 2 unique events, got %d", len(seen))
	}
}

// TestEventNamingConvention verifies events use snake_case.
func TestEventNamingConvention(t *testing.T) {
	tests := []struct {
		name  string
		event state.Event
	}{
		{"EventBeginPublishing uses snake_case", EventBeginPublishing},
		{"EventCompleteBreakdown uses snake_case", EventCompleteBreakdown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventStr := string(tt.event)
			// Check that it contains underscore (snake_case indicator)
			hasUnderscore := false
			for _, c := range eventStr {
				if c == '_' {
					hasUnderscore = true
					break
				}
			}
			if !hasUnderscore {
				t.Errorf("Event %s should use snake_case (contain underscore)", tt.event)
			}
		})
	}
}
