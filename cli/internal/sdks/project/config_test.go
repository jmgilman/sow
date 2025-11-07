package project

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

func TestGetTaskSupportingPhases(t *testing.T) {
	tests := []struct {
		name     string
		config   *ProjectTypeConfig
		expected []string
	}{
		{
			name: "single task-supporting phase",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"implementation": {supportsTasks: true},
					"planning":       {supportsTasks: false},
					"review":         {supportsTasks: false},
				},
			},
			expected: []string{"implementation"},
		},
		{
			name: "multiple task-supporting phases (sorted)",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"testing":        {supportsTasks: true},
					"implementation": {supportsTasks: true},
					"planning":       {supportsTasks: false},
					"design":         {supportsTasks: true},
				},
			},
			expected: []string{"design", "implementation", "testing"},
		},
		{
			name: "no task-supporting phases",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"exploration": {supportsTasks: false},
					"review":      {supportsTasks: false},
				},
			},
			expected: []string{},
		},
		{
			name: "empty phase configs",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTaskSupportingPhases()

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("GetTaskSupportingPhases() length = %d, expected %d", len(result), len(tt.expected))
				return
			}

			// Check each element
			for i, phase := range result {
				if phase != tt.expected[i] {
					t.Errorf("GetTaskSupportingPhases()[%d] = %s, expected %s", i, phase, tt.expected[i])
				}
			}
		})
	}
}

func TestPhaseSupportsTasks(t *testing.T) {
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"implementation": {supportsTasks: true},
			"planning":       {supportsTasks: false},
			"review":         {supportsTasks: false},
		},
	}

	tests := []struct {
		name      string
		phaseName string
		expected  bool
	}{
		{
			name:      "phase supports tasks",
			phaseName: "implementation",
			expected:  true,
		},
		{
			name:      "phase does not support tasks",
			phaseName: "planning",
			expected:  false,
		},
		{
			name:      "phase does not exist",
			phaseName: "nonexistent",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.PhaseSupportsTasks(tt.phaseName)
			if result != tt.expected {
				t.Errorf("PhaseSupportsTasks(%s) = %v, expected %v", tt.phaseName, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultTaskPhase(t *testing.T) {
	const (
		StateStart         = sdkstate.State("StateStart")
		StateImplPlanning  = sdkstate.State("StateImplPlanning")
		StateImplExecuting = sdkstate.State("StateImplExecuting")
		StateReview        = sdkstate.State("StateReview")
		StateUnknown       = sdkstate.State("StateUnknown")
	)

	tests := []struct {
		name         string
		config       *ProjectTypeConfig
		currentState sdkstate.State
		expected     string
	}{
		{
			name: "current state maps to task-supporting phase start",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"planning": {
						startState:    StateStart,
						supportsTasks: false,
					},
					"implementation": {
						startState:    StateImplPlanning,
						endState:      StateImplExecuting,
						supportsTasks: true,
					},
				},
			},
			currentState: StateImplPlanning,
			expected:     "implementation",
		},
		{
			name: "current state maps to task-supporting phase end",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"implementation": {
						startState:    StateImplPlanning,
						endState:      StateImplExecuting,
						supportsTasks: true,
					},
				},
			},
			currentState: StateImplExecuting,
			expected:     "implementation",
		},
		{
			name: "current state does not map, fallback to first alphabetically",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"testing": {
						startState:    StateReview,
						supportsTasks: true,
					},
					"implementation": {
						startState:    StateImplPlanning,
						supportsTasks: true,
					},
				},
			},
			currentState: StateUnknown,
			expected:     "implementation", // alphabetically first
		},
		{
			name: "no task-supporting phases",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"planning": {
						startState:    StateStart,
						supportsTasks: false,
					},
				},
			},
			currentState: StateStart,
			expected:     "",
		},
		{
			name: "multiple task-supporting phases, prefer state match",
			config: &ProjectTypeConfig{
				phaseConfigs: map[string]*PhaseConfig{
					"zzz-phase": {
						startState:    StateUnknown,
						supportsTasks: true,
					},
					"implementation": {
						startState:    StateImplPlanning,
						supportsTasks: true,
					},
				},
			},
			currentState: StateImplPlanning,
			expected:     "implementation", // state match, not alphabetical
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDefaultTaskPhase(tt.currentState)
			if result != tt.expected {
				t.Errorf("GetDefaultTaskPhase(%v) = %s, expected %s", tt.currentState, result, tt.expected)
			}
		})
	}
}

func TestGetAvailableTransitions(t *testing.T) {
	t.Run("returns transitions from branching state", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("BranchState"),
				BranchOn(func(_ *state.Project) string { return "val1" }),
				When("val1", sdkstate.Event("E1"), sdkstate.State("State2"),
					WithDescription("Branch to State2")),
				When("val2", sdkstate.Event("E2"), sdkstate.State("State3"),
					WithDescription("Branch to State3"),
					WithGuard("test guard", func(_ *state.Project) bool { return true })),
			).
			Build()

		transitions := config.GetAvailableTransitions(sdkstate.State("BranchState"))

		if len(transitions) != 2 {
			t.Fatalf("expected 2 transitions, got %d", len(transitions))
		}
		// Verify sorted by event name
		if transitions[0].Event != sdkstate.Event("E1") {
			t.Errorf("expected first event E1, got %s", transitions[0].Event)
		}
		if transitions[1].Event != sdkstate.Event("E2") {
			t.Errorf("expected second event E2, got %s", transitions[1].Event)
		}
		// Verify first transition
		if transitions[0].From != sdkstate.State("BranchState") {
			t.Errorf("expected from BranchState, got %s", transitions[0].From)
		}
		if transitions[0].To != sdkstate.State("State2") {
			t.Errorf("expected to State2, got %s", transitions[0].To)
		}
		if transitions[0].Description != "Branch to State2" {
			t.Errorf("expected description 'Branch to State2', got %s", transitions[0].Description)
		}
		if transitions[0].GuardDesc != "" {
			t.Errorf("expected empty guard desc, got %s", transitions[0].GuardDesc)
		}
		// Verify second transition
		if transitions[1].To != sdkstate.State("State3") {
			t.Errorf("expected to State3, got %s", transitions[1].To)
		}
		if transitions[1].Description != "Branch to State3" {
			t.Errorf("expected description 'Branch to State3', got %s", transitions[1].Description)
		}
		if transitions[1].GuardDesc != "test guard" {
			t.Errorf("expected guard desc 'test guard', got %s", transitions[1].GuardDesc)
		}
	})

	t.Run("returns transitions from non-branching state", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("State1"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
				WithDescription("Direct transition"),
			).
			Build()

		transitions := config.GetAvailableTransitions(sdkstate.State("State1"))

		if len(transitions) != 1 {
			t.Fatalf("expected 1 transition, got %d", len(transitions))
		}
		if transitions[0].Event != sdkstate.Event("E1") {
			t.Errorf("expected event E1, got %s", transitions[0].Event)
		}
		if transitions[0].From != sdkstate.State("State1") {
			t.Errorf("expected from State1, got %s", transitions[0].From)
		}
		if transitions[0].To != sdkstate.State("State2") {
			t.Errorf("expected to State2, got %s", transitions[0].To)
		}
		if transitions[0].Description != "Direct transition" {
			t.Errorf("expected description 'Direct transition', got %s", transitions[0].Description)
		}
	})

	t.Run("returns empty slice for state with no transitions", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").Build()

		transitions := config.GetAvailableTransitions(sdkstate.State("NoTransitions"))

		if len(transitions) != 0 {
			t.Errorf("expected empty slice, got %d transitions", len(transitions))
		}
	})

	t.Run("combines branch and direct transitions", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("MixedState"),
				BranchOn(func(_ *state.Project) string { return "val" }),
				When("val", sdkstate.Event("E1"), sdkstate.State("S1")),
			).
			AddTransition(
				sdkstate.State("MixedState"),
				sdkstate.State("S2"),
				sdkstate.Event("E2"),
			).
			Build()

		transitions := config.GetAvailableTransitions(sdkstate.State("MixedState"))

		if len(transitions) != 2 {
			t.Fatalf("expected 2 transitions, got %d", len(transitions))
		}
		// Verify events are present (sorted)
		if transitions[0].Event != sdkstate.Event("E1") {
			t.Errorf("expected first event E1, got %s", transitions[0].Event)
		}
		if transitions[1].Event != sdkstate.Event("E2") {
			t.Errorf("expected second event E2, got %s", transitions[1].Event)
		}
	})

	t.Run("returns transitions in sorted order by event", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(sdkstate.State("MultiState"), sdkstate.State("S1"), sdkstate.Event("EventZ")).
			AddTransition(sdkstate.State("MultiState"), sdkstate.State("S2"), sdkstate.Event("EventA")).
			AddTransition(sdkstate.State("MultiState"), sdkstate.State("S3"), sdkstate.Event("EventM")).
			Build()

		transitions := config.GetAvailableTransitions(sdkstate.State("MultiState"))

		if len(transitions) != 3 {
			t.Fatalf("expected 3 transitions, got %d", len(transitions))
		}
		// Verify sorted by event name
		if transitions[0].Event != sdkstate.Event("EventA") {
			t.Errorf("expected first event EventA, got %s", transitions[0].Event)
		}
		if transitions[1].Event != sdkstate.Event("EventM") {
			t.Errorf("expected second event EventM, got %s", transitions[1].Event)
		}
		if transitions[2].Event != sdkstate.Event("EventZ") {
			t.Errorf("expected third event EventZ, got %s", transitions[2].Event)
		}
	})
}

func TestGetTransitionDescription(t *testing.T) {
	t.Run("returns description for branch transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("State1"),
				BranchOn(func(_ *state.Project) string { return "val" }),
				When("val", sdkstate.Event("E1"), sdkstate.State("State2"),
					WithDescription("Branch description")),
			).
			Build()

		desc := config.GetTransitionDescription(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if desc != "Branch description" {
			t.Errorf("expected 'Branch description', got %s", desc)
		}
	})

	t.Run("returns description for direct transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("State1"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
				WithDescription("Direct description"),
			).
			Build()

		desc := config.GetTransitionDescription(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if desc != "Direct description" {
			t.Errorf("expected 'Direct description', got %s", desc)
		}
	})

	t.Run("returns empty string for non-existent transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").Build()

		desc := config.GetTransitionDescription(
			sdkstate.State("NoState"),
			sdkstate.Event("NoEvent"))

		if desc != "" {
			t.Errorf("expected empty string, got %s", desc)
		}
	})

	t.Run("returns empty string when description not provided", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("State1"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
			).
			Build()

		desc := config.GetTransitionDescription(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if desc != "" {
			t.Errorf("expected empty string, got %s", desc)
		}
	})
}

func TestGetTargetState(t *testing.T) {
	t.Run("returns target for branch transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("State1"),
				BranchOn(func(_ *state.Project) string { return "val" }),
				When("val", sdkstate.Event("E1"), sdkstate.State("State2")),
			).
			Build()

		target := config.GetTargetState(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if target != sdkstate.State("State2") {
			t.Errorf("expected State2, got %s", target)
		}
	})

	t.Run("returns target for direct transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("State1"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
			).
			Build()

		target := config.GetTargetState(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if target != sdkstate.State("State2") {
			t.Errorf("expected State2, got %s", target)
		}
	})

	t.Run("returns empty state for non-existent transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").Build()

		target := config.GetTargetState(
			sdkstate.State("NoState"),
			sdkstate.Event("NoEvent"))

		if target != "" {
			t.Errorf("expected empty state, got %s", target)
		}
	})
}

func TestGetGuardDescription(t *testing.T) {
	t.Run("returns guard description for branch transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("State1"),
				BranchOn(func(_ *state.Project) string { return "val" }),
				When("val", sdkstate.Event("E1"), sdkstate.State("State2"),
					WithGuard("test guard", func(_ *state.Project) bool { return true })),
			).
			Build()

		desc := config.GetGuardDescription(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if desc != "test guard" {
			t.Errorf("expected 'test guard', got %s", desc)
		}
	})

	t.Run("returns guard description for direct transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("State1"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
				WithGuard("test guard", func(_ *state.Project) bool { return true }),
			).
			Build()

		desc := config.GetGuardDescription(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if desc != "test guard" {
			t.Errorf("expected 'test guard', got %s", desc)
		}
	})

	t.Run("returns empty string for transition without guard", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("State1"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
			).
			Build()

		desc := config.GetGuardDescription(
			sdkstate.State("State1"),
			sdkstate.Event("E1"))

		if desc != "" {
			t.Errorf("expected empty string, got %s", desc)
		}
	})

	t.Run("returns empty string for non-existent transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").Build()

		desc := config.GetGuardDescription(
			sdkstate.State("NoState"),
			sdkstate.Event("NoEvent"))

		if desc != "" {
			t.Errorf("expected empty string, got %s", desc)
		}
	})
}

func TestIsBranchingState(t *testing.T) {
	t.Run("returns true for state with AddBranch", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("BranchState"),
				BranchOn(func(_ *state.Project) string { return "val" }),
				When("val", sdkstate.Event("E1"), sdkstate.State("State2")),
			).
			Build()

		if !config.IsBranchingState(sdkstate.State("BranchState")) {
			t.Error("expected true for branching state")
		}
	})

	t.Run("returns false for state with only AddTransition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(
				sdkstate.State("LinearState"),
				sdkstate.State("State2"),
				sdkstate.Event("E1"),
			).
			Build()

		if config.IsBranchingState(sdkstate.State("LinearState")) {
			t.Error("expected false for non-branching state")
		}
	})

	t.Run("returns false for state with no transitions", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").Build()

		if config.IsBranchingState(sdkstate.State("NoState")) {
			t.Error("expected false for state with no transitions")
		}
	})

	t.Run("returns false for state with multiple AddTransition but no AddBranch", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddTransition(sdkstate.State("MultiState"), sdkstate.State("S1"), sdkstate.Event("E1")).
			AddTransition(sdkstate.State("MultiState"), sdkstate.State("S2"), sdkstate.Event("E2")).
			Build()

		if config.IsBranchingState(sdkstate.State("MultiState")) {
			t.Error("expected false for state with multiple transitions but no AddBranch")
		}
	})
}
