package state

import (
	"context"
	"testing"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants - use string values to avoid import cycles.
const (
	testStatePlanning     = State("PlanningActive")
	testStateImplementing = State("ImplementationPlanning")
	testEventInit         = Event("project_init")
	testEventComplete     = Event("complete_planning")
)

// mockPromptGenerator is a test implementation of PromptGenerator.
type mockPromptGenerator struct {
	generateFunc func(state State, projectState *schemas.ProjectState) (string, error)
}

func (m *mockPromptGenerator) GeneratePrompt(state State, projectState *schemas.ProjectState) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(state, projectState)
	}
	return "test prompt", nil
}

func TestNewBuilder(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	builder := NewBuilder(testStatePlanning, projectState, generator)

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.sm)
	assert.Equal(t, projectState, builder.projectState)
	assert.Equal(t, generator, builder.promptGenerator)
	assert.False(t, builder.suppressPrompts)
}

func TestMachineBuilder_AddTransition_Unconditional(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	builder := NewBuilder(NoProject, projectState, generator)
	builder.AddTransition(NoProject, testStatePlanning, testEventInit)

	machine := builder.Build()

	assert.NotNil(t, machine)
	assert.Equal(t, NoProject, machine.State())

	// Fire the event and verify transition
	err := machine.Fire(testEventInit)
	require.NoError(t, err)
	assert.Equal(t, testStatePlanning, machine.State())
}

func TestMachineBuilder_AddTransition_WithGuard(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	guardCalled := false
	guardResult := false

	builder := NewBuilder(testStatePlanning, projectState, generator)
	builder.AddTransition(
		testStatePlanning,
		testStateImplementing,
		testEventComplete,
		WithGuard(func() bool {
			guardCalled = true
			return guardResult
		}),
	)

	machine := builder.Build()

	// Guard returns false, transition should fail
	err := machine.Fire(testEventComplete)
	assert.Error(t, err)
	assert.True(t, guardCalled)
	assert.Equal(t, testStatePlanning, machine.State())

	// Guard returns true, transition should succeed
	guardResult = true
	guardCalled = false
	err = machine.Fire(testEventComplete)
	require.NoError(t, err)
	assert.True(t, guardCalled)
	assert.Equal(t, testStateImplementing, machine.State())
}

func TestMachineBuilder_SuppressPrompts(t *testing.T) {
	projectState := &schemas.ProjectState{}
	promptCalled := false
	generator := &mockPromptGenerator{
		generateFunc: func(_ State, _ *schemas.ProjectState) (string, error) {
			promptCalled = true
			return "test prompt", nil
		},
	}

	builder := NewBuilder(NoProject, projectState, generator)
	builder.SuppressPrompts(true).AddTransition(NoProject, testStatePlanning, testEventInit)

	machine := builder.Build()

	// Fire event, prompt should not be generated
	err := machine.Fire(testEventInit)
	require.NoError(t, err)
	assert.False(t, promptCalled, "Prompt should not be generated when suppressed")
}

func TestMachineBuilder_Build(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	builder := NewBuilder(NoProject, projectState, generator)
	builder.AddTransition(NoProject, testStatePlanning, testEventInit)

	machine := builder.Build()

	assert.NotNil(t, machine)
	assert.NotNil(t, machine.sm)
	assert.Equal(t, projectState, machine.projectState)
	assert.Equal(t, NoProject, machine.State())
}

func TestMachineBuilder_ConfigureState(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	builder := NewBuilder(NoProject, projectState, generator)

	// ConfigureState should return a valid StateConfiguration
	cfg := builder.ConfigureState(NoProject)
	assert.NotNil(t, cfg)

	// Should be able to use it to configure transitions directly
	cfg.Permit(testEventInit, testStatePlanning)

	machine := builder.Build()

	// Verify the transition works
	err := machine.Fire(testEventInit)
	require.NoError(t, err)
	assert.Equal(t, testStatePlanning, machine.State())
}

func TestWithGuard(t *testing.T) {
	called := false
	guard := func() bool {
		called = true
		return true
	}

	config := &transitionConfig{}
	option := WithGuard(guard)
	option(config)

	assert.NotNil(t, config.guard)

	// Call the guard and verify it works
	result := config.guard()
	assert.True(t, result)
	assert.True(t, called)
}

// TestMachineBuilder_ChainedTransitions verifies that multiple transitions can be added fluently.
func TestMachineBuilder_ChainedTransitions(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	eventAB := Event("a_to_b")
	eventBC := Event("b_to_c")

	// Chain multiple AddTransition calls
	machine := NewBuilder(stateA, projectState, generator).
		AddTransition(stateA, stateB, eventAB).
		AddTransition(stateB, stateC, eventBC).
		SuppressPrompts(true).
		Build()

	assert.Equal(t, stateA, machine.State())

	err := machine.Fire(eventAB)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())

	err = machine.Fire(eventBC)
	require.NoError(t, err)
	assert.Equal(t, stateC, machine.State())
}

// TestMachineBuilder_MultipleTransitionsFromSameState verifies that multiple transitions
// can originate from the same state.
func TestMachineBuilder_MultipleTransitionsFromSameState(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	eventToB := Event("to_b")
	eventToC := Event("to_c")

	builder := NewBuilder(stateA, projectState, generator)
	builder.AddTransition(stateA, stateB, eventToB)
	builder.AddTransition(stateA, stateC, eventToC)
	machine := builder.SuppressPrompts(true).Build()

	// Should be able to fire either event
	can, err := machine.CanFire(eventToB)
	require.NoError(t, err)
	assert.True(t, can)

	can, err = machine.CanFire(eventToC)
	require.NoError(t, err)
	assert.True(t, can)

	// Fire one event
	err = machine.Fire(eventToB)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())
}

// TestWithOnEntry verifies that entry actions are configured correctly.
func TestWithOnEntry(t *testing.T) {
	entryCalled := false
	entryAction := func(_ context.Context, _ ...any) error {
		entryCalled = true
		return nil
	}

	config := &transitionConfig{}
	option := WithOnEntry(entryAction)
	option(config)

	assert.Len(t, config.entryActions, 1)

	// Call the entry action and verify it works
	err := config.entryActions[0](context.Background())
	assert.NoError(t, err)
	assert.True(t, entryCalled)
}

// TestWithOnExit verifies that exit actions are configured correctly.
func TestWithOnExit(t *testing.T) {
	exitCalled := false
	exitAction := func(_ context.Context, _ ...any) error {
		exitCalled = true
		return nil
	}

	config := &transitionConfig{}
	option := WithOnExit(exitAction)
	option(config)

	assert.Len(t, config.exitActions, 1)

	// Call the exit action and verify it works
	err := config.exitActions[0](context.Background())
	assert.NoError(t, err)
	assert.True(t, exitCalled)
}

// TestMachineBuilder_WithEntryAndExitActions verifies that entry and exit actions execute.
func TestMachineBuilder_WithEntryAndExitActions(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	exitCalled := false
	entryCalled := false

	builder := NewBuilder(stateA, projectState, generator)
	builder.AddTransition(
		stateA,
		stateB,
		eventGo,
		WithOnExit(func(_ context.Context, _ ...any) error {
			exitCalled = true
			return nil
		}),
		WithOnEntry(func(_ context.Context, _ ...any) error {
			entryCalled = true
			return nil
		}),
	)

	machine := builder.SuppressPrompts(true).Build()

	err := machine.Fire(eventGo)
	require.NoError(t, err)

	assert.True(t, exitCalled, "Exit action should have been called")
	assert.True(t, entryCalled, "Entry action should have been called")
	assert.Equal(t, stateB, machine.State())
}

// TestMachineBuilder_MultipleOnEntryActions verifies that multiple entry actions execute in order.
func TestMachineBuilder_MultipleOnEntryActions(t *testing.T) {
	projectState := &schemas.ProjectState{}
	generator := &mockPromptGenerator{}

	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	var executionOrder []int

	builder := NewBuilder(stateA, projectState, generator)
	builder.AddTransition(
		stateA,
		stateB,
		eventGo,
		WithOnEntry(func(_ context.Context, _ ...any) error {
			executionOrder = append(executionOrder, 1)
			return nil
		}),
		WithOnEntry(func(_ context.Context, _ ...any) error {
			executionOrder = append(executionOrder, 2)
			return nil
		}),
	)

	machine := builder.SuppressPrompts(true).Build()

	err := machine.Fire(eventGo)
	require.NoError(t, err)

	assert.Equal(t, []int{1, 2}, executionOrder, "Entry actions should execute in order")
}

// TestMachineBuilder_BuildReturnsCorrectProjectState verifies that the built machine
// retains the project state.
func TestMachineBuilder_BuildReturnsCorrectProjectState(t *testing.T) {
	projectState := &schemas.ProjectState{}
	projectState.Project.Name = "test-project"
	projectState.Project.Type = "feature"

	generator := &mockPromptGenerator{}

	builder := NewBuilder(NoProject, projectState, generator)
	machine := builder.Build()

	assert.Equal(t, projectState, machine.ProjectState())
	assert.Equal(t, "test-project", machine.ProjectState().Project.Name)
	assert.Equal(t, "feature", machine.ProjectState().Project.Type)
}
