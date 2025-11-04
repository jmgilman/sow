package project

import (
	"testing"

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
		StateStart           = sdkstate.State("StateStart")
		StateImplPlanning    = sdkstate.State("StateImplPlanning")
		StateImplExecuting   = sdkstate.State("StateImplExecuting")
		StateReview          = sdkstate.State("StateReview")
		StateUnknown         = sdkstate.State("StateUnknown")
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
