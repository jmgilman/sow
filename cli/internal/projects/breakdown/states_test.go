package breakdown

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// TestStatesAreCorrectType verifies all state constants use state.State type.
func TestStatesAreCorrectType(t *testing.T) {
	tests := []struct {
		name  string
		state state.State
	}{
		{"Active", Active},
		{"Publishing", Publishing},
		{"Completed", Completed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the constant can be assigned to state.State type
			s := tt.state
			if s == "" {
				t.Errorf("State %s should not be empty", tt.name)
			}
		})
	}
}

// TestStateValues verifies state constants have correct string values.
func TestStateValues(t *testing.T) {
	tests := []struct {
		name     string
		state    state.State
		expected string
	}{
		{"Active state", Active, "Active"},
		{"Publishing state", Publishing, "Publishing"},
		{"Completed state", Completed, "Completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.state) != tt.expected {
				t.Errorf("Expected state value %q, got %q", tt.expected, string(tt.state))
			}
		})
	}
}

// TestAllStatesAreDifferent verifies no duplicate state values.
func TestAllStatesAreDifferent(t *testing.T) {
	states := []state.State{Active, Publishing, Completed}
	seen := make(map[state.State]bool)

	for _, s := range states {
		if seen[s] {
			t.Errorf("Duplicate state found: %s", s)
		}
		seen[s] = true
	}

	if len(seen) != 3 {
		t.Errorf("Expected 3 unique states, got %d", len(seen))
	}
}
