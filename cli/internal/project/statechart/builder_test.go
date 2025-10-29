package statechart

import (
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
