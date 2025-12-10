package project

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("creates builder with initial state", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(testStatePending, nil)

		require.NotNil(t, builder)
		machine := builder.Build()
		assert.Equal(t, testStatePending, machine.State())
	})

	t.Run("creates builder with prompt function", func(t *testing.T) {
		t.Parallel()

		promptFunc := func(_ State) string {
			return "test prompt"
		}

		builder := NewBuilder(testStatePending, promptFunc)
		machine := builder.Build()

		assert.Equal(t, "test prompt", machine.Prompt())
	})

	t.Run("creates builder with nil prompt function", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(testStatePending, nil)
		machine := builder.Build()

		assert.Empty(t, machine.Prompt())
	})
}

func TestMachineBuilder_AddTransition(t *testing.T) {
	t.Parallel()

	t.Run("adds transition correctly", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		assert.True(t, machine.CanFire(testEventStart))
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(testStatePending, nil)

		result := builder.AddTransition(testStatePending, testStateActive, testEventStart)

		assert.Same(t, builder, result)
	})

	t.Run("multiple transitions from same state work", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			AddTransition(testStatePending, testStateCompleted, testEventComplete).
			Build()

		assert.True(t, machine.CanFire(testEventStart))
		assert.True(t, machine.CanFire(testEventComplete))
	})

	t.Run("chained transitions work correctly", func(t *testing.T) {
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

func TestMachineBuilder_Build(t *testing.T) {
	t.Parallel()

	t.Run("creates working machine", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart).
			Build()

		require.NotNil(t, machine)
		assert.Equal(t, testStatePending, machine.State())
	})

	t.Run("machine without transitions works", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).Build()

		require.NotNil(t, machine)
		assert.Equal(t, testStatePending, machine.State())
		assert.Empty(t, machine.PermittedTriggers())
	})
}

func TestMachineBuilder_Guards(t *testing.T) {
	t.Parallel()

	t.Run("guards block transitions when false", func(t *testing.T) {
		t.Parallel()

		guardCalled := false
		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuard(func() bool {
					guardCalled = true
					return false
				})).
			Build()

		err := machine.Fire(testEventStart)

		require.Error(t, err)
		assert.True(t, guardCalled, "guard should have been called")
		assert.Equal(t, testStatePending, machine.State()) // State unchanged
	})

	t.Run("guards allow transitions when true", func(t *testing.T) {
		t.Parallel()

		guardCalled := false
		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuard(func() bool {
					guardCalled = true
					return true
				})).
			Build()

		err := machine.Fire(testEventStart)

		require.NoError(t, err)
		assert.True(t, guardCalled, "guard should have been called")
		assert.Equal(t, testStateActive, machine.State())
	})

	t.Run("guard description appears in error messages", func(t *testing.T) {
		t.Parallel()

		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuardDescription("all tasks must be complete", func() bool {
					return false
				})).
			Build()

		err := machine.Fire(testEventStart)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "all tasks must be complete")
	})
}

//nolint:funlen // test functions with many subtests are naturally long
func TestMachineBuilder_Actions(t *testing.T) {
	t.Parallel()

	t.Run("OnEntry actions run on transition", func(t *testing.T) {
		t.Parallel()

		entryCalled := false
		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithOnEntry(func(_ context.Context, _ ...any) error {
					entryCalled = true
					return nil
				})).
			Build()

		require.NoError(t, machine.Fire(testEventStart))

		assert.True(t, entryCalled, "entry action should have been called")
	})

	t.Run("OnExit actions run on transition", func(t *testing.T) {
		t.Parallel()

		exitCalled := false
		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithOnExit(func(_ context.Context, _ ...any) error {
					exitCalled = true
					return nil
				})).
			Build()

		require.NoError(t, machine.Fire(testEventStart))

		assert.True(t, exitCalled, "exit action should have been called")
	})

	t.Run("both entry and exit actions run", func(t *testing.T) {
		t.Parallel()

		var callOrder []string
		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithOnExit(func(_ context.Context, _ ...any) error {
					callOrder = append(callOrder, "exit")
					return nil
				}),
				WithOnEntry(func(_ context.Context, _ ...any) error {
					callOrder = append(callOrder, "entry")
					return nil
				})).
			Build()

		require.NoError(t, machine.Fire(testEventStart))

		assert.Contains(t, callOrder, "exit")
		assert.Contains(t, callOrder, "entry")
	})

	t.Run("multiple OnEntry actions for same target state are composed", func(t *testing.T) {
		t.Parallel()

		var callOrder []string
		machine := NewBuilder(testStatePending, nil).
			// First transition to Active has OnEntry
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithOnEntry(func(_ context.Context, _ ...any) error {
					callOrder = append(callOrder, "entry1")
					return nil
				})).
			// Second transition to Active also has OnEntry
			AddTransition(testStateCompleted, testStateActive, Event("restart"),
				WithOnEntry(func(_ context.Context, _ ...any) error {
					callOrder = append(callOrder, "entry2")
					return nil
				})).
			Build()

		require.NoError(t, machine.Fire(testEventStart))

		// Both entry actions should be composed and run
		assert.Contains(t, callOrder, "entry1")
		assert.Contains(t, callOrder, "entry2")
	})

	t.Run("multiple OnExit actions for same source state are composed", func(t *testing.T) {
		t.Parallel()

		var callOrder []string
		machine := NewBuilder(testStatePending, nil).
			// First transition from Pending has OnExit
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithOnExit(func(_ context.Context, _ ...any) error {
					callOrder = append(callOrder, "exit1")
					return nil
				})).
			// Second transition from Pending also has OnExit
			AddTransition(testStatePending, testStateCompleted, testEventComplete,
				WithOnExit(func(_ context.Context, _ ...any) error {
					callOrder = append(callOrder, "exit2")
					return nil
				})).
			Build()

		require.NoError(t, machine.Fire(testEventStart))

		// Both exit actions should be composed and run
		assert.Contains(t, callOrder, "exit1")
		assert.Contains(t, callOrder, "exit2")
	})

	t.Run("composed actions stop on first error", func(t *testing.T) {
		t.Parallel()

		action2Called := false
		machine := NewBuilder(testStatePending, nil).
			// First transition has failing OnEntry
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithOnEntry(func(_ context.Context, _ ...any) error {
					return errors.New("action1 failed")
				})).
			// Second transition to same state has another OnEntry
			AddTransition(testStateCompleted, testStateActive, Event("restart"),
				WithOnEntry(func(_ context.Context, _ ...any) error {
					action2Called = true
					return nil
				})).
			Build()

		err := machine.Fire(testEventStart)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "action1 failed")
		// Second action should NOT have been called (error stops composition)
		assert.False(t, action2Called)
	})
}

func TestMachineBuilder_ComplexScenarios(t *testing.T) {
	t.Parallel()

	t.Run("complex machine with multiple states and transitions", func(t *testing.T) {
		t.Parallel()

		// Define a workflow-like state machine
		const (
			stateIdle       State = "Idle"
			stateProcessing State = "Processing"
			stateReviewing  State = "Reviewing"
			stateDone       State = "Done"

			eventProcess Event = "Process"
			eventReview  Event = "Review"
			eventApprove Event = "Approve"
			eventReject  Event = "Reject"
		)

		machine := NewBuilder(stateIdle, nil).
			AddTransition(stateIdle, stateProcessing, eventProcess).
			AddTransition(stateProcessing, stateReviewing, eventReview).
			AddTransition(stateReviewing, stateDone, eventApprove).
			AddTransition(stateReviewing, stateProcessing, eventReject).
			Build()

		// Start idle
		assert.Equal(t, stateIdle, machine.State())

		// Process
		require.NoError(t, machine.Fire(eventProcess))
		assert.Equal(t, stateProcessing, machine.State())

		// Review
		require.NoError(t, machine.Fire(eventReview))
		assert.Equal(t, stateReviewing, machine.State())

		// Reject goes back to processing
		require.NoError(t, machine.Fire(eventReject))
		assert.Equal(t, stateProcessing, machine.State())

		// Review again
		require.NoError(t, machine.Fire(eventReview))
		assert.Equal(t, stateReviewing, machine.State())

		// Approve completes
		require.NoError(t, machine.Fire(eventApprove))
		assert.Equal(t, stateDone, machine.State())
	})

	t.Run("dynamic guard evaluation", func(t *testing.T) {
		t.Parallel()

		// Guard that changes based on external state
		taskCount := 0
		machine := NewBuilder(testStatePending, nil).
			AddTransition(testStatePending, testStateActive, testEventStart,
				WithGuard(func() bool {
					return taskCount > 0
				})).
			Build()

		// Should fail initially
		assert.False(t, machine.CanFire(testEventStart))

		// Update external state
		taskCount = 5

		// Should pass now
		assert.True(t, machine.CanFire(testEventStart))
		require.NoError(t, machine.Fire(testEventStart))
	})

	t.Run("prompt generation works correctly", func(t *testing.T) {
		t.Parallel()

		promptFunc := func(s State) string {
			switch s {
			case testStatePending:
				return "Waiting to start"
			case testStateActive:
				return "Work in progress"
			case testStateCompleted:
				return "All done!"
			default:
				return ""
			}
		}

		machine := NewBuilder(testStatePending, promptFunc).
			AddTransition(testStatePending, testStateActive, testEventStart).
			AddTransition(testStateActive, testStateCompleted, testEventComplete).
			Build()

		assert.Equal(t, "Waiting to start", machine.Prompt())

		require.NoError(t, machine.Fire(testEventStart))
		assert.Equal(t, "Work in progress", machine.Prompt())

		require.NoError(t, machine.Fire(testEventComplete))
		assert.Equal(t, "All done!", machine.Prompt())
	})
}
