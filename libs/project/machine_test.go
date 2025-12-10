package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test states and events for machine tests.
const (
	testStatePending   State = "Pending"
	testStateActive    State = "Active"
	testStateCompleted State = "Completed"

	testEventStart    Event = "Start"
	testEventComplete Event = "Complete"
	testEventReset    Event = "Reset"
)

func TestMachine_State(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialState State
		want         State
	}{
		{
			name:         "returns initial state",
			initialState: testStatePending,
			want:         testStatePending,
		},
		{
			name:         "returns NoProject state",
			initialState: NoProject,
			want:         NoProject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			machine := NewBuilder(tt.initialState, nil).Build()

			got := machine.State()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMachine_Fire(t *testing.T) {
	t.Parallel()

	t.Run("transitions to new state", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		err := machine.Fire(testEventStart)

		require.NoError(t, err)
		assert.Equal(t, testStateActive, machine.State())
	})

	t.Run("returns error for invalid transition", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		err := machine.Fire(testEventComplete)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Complete")
		assert.Equal(t, testStatePending, machine.State()) // State unchanged
	})

	t.Run("allows multiple transitions in sequence", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			AddTransition(testStateActive, testStateCompleted, testEventComplete).
			Build()

		require.NoError(t, machine.Fire(testEventStart))
		assert.Equal(t, testStateActive, machine.State())

		require.NoError(t, machine.Fire(testEventComplete))
		assert.Equal(t, testStateCompleted, machine.State())
	})
}

func TestMachine_CanFire(t *testing.T) {
	t.Parallel()

	t.Run("returns true for valid event", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		canFire := machine.CanFire(testEventStart)

		assert.True(t, canFire)
	})

	t.Run("returns false for invalid event", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		canFire := machine.CanFire(testEventComplete)

		assert.False(t, canFire)
	})

	t.Run("returns false when guard fails", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuard(func() bool { return false })).
			Build()

		canFire := machine.CanFire(testEventStart)

		assert.False(t, canFire)
	})

	t.Run("returns true when guard passes", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuard(func() bool { return true })).
			Build()

		canFire := machine.CanFire(testEventStart)

		assert.True(t, canFire)
	})
}

func TestMachine_PermittedTriggers(t *testing.T) {
	t.Parallel()

	t.Run("returns available events from current state", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			AddTransition(testStateActive, testStateCompleted, testEventComplete).
			Build()

		events := machine.PermittedTriggers()

		assert.Len(t, events, 1)
		assert.Contains(t, events, testEventStart)
	})

	t.Run("returns multiple events when available", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			AddTransition(testStatePending, NoProject, testEventReset).
			Build()

		events := machine.PermittedTriggers()

		assert.Len(t, events, 2)
		assert.Contains(t, events, testEventStart)
		assert.Contains(t, events, testEventReset)
	})

	t.Run("returns empty slice when no events available", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStateCompleted, nil).Build()

		events := machine.PermittedTriggers()

		assert.Empty(t, events)
	})

	t.Run("excludes events blocked by guards", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuard(func() bool { return false })).
			Build()

		events := machine.PermittedTriggers()

		assert.Empty(t, events)
	})
}

func TestMachine_Prompt(t *testing.T) {
	t.Parallel()

	t.Run("returns prompt for current state", func(t *testing.T) {
		t.Parallel()

		promptFunc := func(s State) string {
			if s == testStatePending {
				return "Start your work"
			}
			return ""
		}

		machine := NewBuilder(testStatePending, promptFunc).Build()

		prompt := machine.Prompt()

		assert.Equal(t, "Start your work", prompt)
	})

	t.Run("returns empty string when no prompt configured", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).Build()

		prompt := machine.Prompt()

		assert.Empty(t, prompt)
	})

	t.Run("returns empty string when state has no prompt", func(t *testing.T) {
		t.Parallel()

		promptFunc := func(s State) string {
			if s == testStateActive {
				return "Work in progress"
			}
			return ""
		}

		machine := NewBuilder(testStatePending, promptFunc).Build()

		prompt := machine.Prompt()

		assert.Empty(t, prompt)
	})

	t.Run("returns updated prompt after state change", func(t *testing.T) {
		t.Parallel()

		promptFunc := func(s State) string {
			switch s {
			case testStatePending:
				return "Start your work"
			case testStateActive:
				return "Continue working"
			default:
				return ""
			}
		}

		machine := NewBuilder(testStatePending, promptFunc).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		assert.Equal(t, "Start your work", machine.Prompt())

		require.NoError(t, machine.Fire(testEventStart))

		assert.Equal(t, "Continue working", machine.Prompt())
	})
}
