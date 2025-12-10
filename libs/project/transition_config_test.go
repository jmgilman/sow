package project

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/stretchr/testify/assert"
)

// Test states and events for TransitionConfig tests.
const (
	transitionTestStateFrom    State = "StateFrom"
	transitionTestStateTo      State = "StateTo"
	transitionTestEvent        Event = "TestEvent"
	transitionTestFailedPhase        = "review"
)

func TestTransitionConfig_GuardDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		guardTemplate GuardTemplate
		want          string
	}{
		{
			name: "returns guard description",
			guardTemplate: GuardTemplate{
				Description: "all tasks complete",
				Func:        func(*state.Project) bool { return true },
			},
			want: "all tasks complete",
		},
		{
			name:          "returns empty when no guard",
			guardTemplate: GuardTemplate{},
			want:          "",
		},
		{
			name: "returns empty when guard has no description",
			guardTemplate: GuardTemplate{
				Func: func(*state.Project) bool { return true },
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc := &TransitionConfig{
				From:          transitionTestStateFrom,
				To:            transitionTestStateTo,
				Event:         transitionTestEvent,
				guardTemplate: tt.guardTemplate,
			}

			got := tc.GuardDescription()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTransitionConfig_Description(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "returns description",
			description: "Complete planning and begin implementation",
			want:        "Complete planning and begin implementation",
		},
		{
			name:        "returns empty description",
			description: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc := &TransitionConfig{
				From:        transitionTestStateFrom,
				To:          transitionTestStateTo,
				Event:       transitionTestEvent,
				description: tt.description,
			}

			got := tc.Description()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTransitionConfig_FailedPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		failedPhase string
		want        string
	}{
		{
			name:        "returns failed phase name",
			failedPhase: "review",
			want:        "review",
		},
		{
			name:        "returns empty when no failed phase",
			failedPhase: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc := &TransitionConfig{
				From:        transitionTestStateFrom,
				To:          transitionTestStateTo,
				Event:       transitionTestEvent,
				failedPhase: tt.failedPhase,
			}

			got := tc.FailedPhase()

			assert.Equal(t, tt.want, got)
		})
	}
}
