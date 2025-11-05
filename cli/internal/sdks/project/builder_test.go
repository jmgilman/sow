package project

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for use across tests.
const (
	testProjectName = "test-project"
	testPhaseName   = "planning"
)

var (
	testState1 = sdkstate.State("state1")
	testState2 = sdkstate.State("state2")
	testEvent1 = sdkstate.Event("event1")
)

// TestNewProjectTypeConfigBuilder tests the constructor.
func TestNewProjectTypeConfigBuilder(t *testing.T) {
	t.Run("creates builder with given name", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		require.NotNil(t, builder)
		assert.Equal(t, testProjectName, builder.name)
	})

	t.Run("initializes empty collections", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		assert.NotNil(t, builder.phaseConfigs)
		assert.Empty(t, builder.phaseConfigs)

		assert.NotNil(t, builder.transitions)
		assert.Empty(t, builder.transitions)

		assert.NotNil(t, builder.onAdvance)
		assert.Empty(t, builder.onAdvance)

		assert.NotNil(t, builder.prompts)
		assert.Empty(t, builder.prompts)
	})
}

// TestWithPhase tests phase configuration.
func TestWithPhase(t *testing.T) {
	t.Run("adds phase config", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		result := builder.WithPhase(testPhaseName)

		assert.NotNil(t, builder.phaseConfigs[testPhaseName])
		assert.Equal(t, testPhaseName, builder.phaseConfigs[testPhaseName].name)
		assert.Same(t, builder, result) // chainable
	})

	t.Run("applies options correctly", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		builder.WithPhase(testPhaseName,
			WithStartState(testState1),
			WithEndState(testState2),
			WithInputs("type1", "type2"),
			WithOutputs("type3"),
			WithTasks(),
			WithMetadataSchema("test schema"),
		)

		pc := builder.phaseConfigs[testPhaseName]
		require.NotNil(t, pc)
		assert.Equal(t, testState1, pc.startState)
		assert.Equal(t, testState2, pc.endState)
		assert.Equal(t, []string{"type1", "type2"}, pc.allowedInputTypes)
		assert.Equal(t, []string{"type3"}, pc.allowedOutputTypes)
		assert.True(t, pc.supportsTasks)
		assert.Equal(t, "test schema", pc.metadataSchema)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		result := builder.WithPhase(testPhaseName)

		assert.Same(t, builder, result)
	})

	t.Run("multiple phases can be added", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		builder.WithPhase("phase1")
		builder.WithPhase("phase2")
		builder.WithPhase("phase3")

		assert.Len(t, builder.phaseConfigs, 3)
		assert.NotNil(t, builder.phaseConfigs["phase1"])
		assert.NotNil(t, builder.phaseConfigs["phase2"])
		assert.NotNil(t, builder.phaseConfigs["phase3"])
	})
}

// TestSetInitialState tests initial state configuration.
func TestSetInitialState(t *testing.T) {
	t.Run("sets initial state", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		builder.SetInitialState(testState1)

		assert.Equal(t, testState1, builder.initialState)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		result := builder.SetInitialState(testState1)

		assert.Same(t, builder, result)
	})
}

// TestAddTransition tests transition configuration.
func TestAddTransition(t *testing.T) {
	t.Run("adds transition with from/to/event", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		builder.AddTransition(testState1, testState2, testEvent1)

		require.Len(t, builder.transitions, 1)
		assert.Equal(t, testState1, builder.transitions[0].From)
		assert.Equal(t, testState2, builder.transitions[0].To)
		assert.Equal(t, testEvent1, builder.transitions[0].Event)
	})

	t.Run("applies options correctly", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		guardFunc := func(_ *state.Project) bool { return true }
		entryAction := func(_ *state.Project) error { return nil }
		exitAction := func(_ *state.Project) error { return nil }

		builder.AddTransition(testState1, testState2, testEvent1,
			WithGuard("test guard", guardFunc),
			WithOnEntry(entryAction),
			WithOnExit(exitAction),
		)

		require.Len(t, builder.transitions, 1)
		tc := builder.transitions[0]
		assert.NotNil(t, tc.guardTemplate)
		assert.NotNil(t, tc.onEntry)
		assert.NotNil(t, tc.onExit)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		result := builder.AddTransition(testState1, testState2, testEvent1)

		assert.Same(t, builder, result)
	})

	t.Run("multiple transitions can be added", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		state3 := sdkstate.State("state3")
		event2 := sdkstate.Event("event2")

		builder.AddTransition(testState1, testState2, testEvent1)
		builder.AddTransition(testState2, state3, event2)

		assert.Len(t, builder.transitions, 2)
	})
}

// TestOnAdvance tests event determiner configuration.
func TestOnAdvance(t *testing.T) {
	t.Run("stores event determiner for state", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		determiner := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}

		builder.OnAdvance(testState1, determiner)

		assert.NotNil(t, builder.onAdvance[testState1])
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		determiner := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}

		result := builder.OnAdvance(testState1, determiner)

		assert.Same(t, builder, result)
	})

	t.Run("multiple states can have determiners", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		determiner1 := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}
		determiner2 := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}

		builder.OnAdvance(testState1, determiner1)
		builder.OnAdvance(testState2, determiner2)

		assert.Len(t, builder.onAdvance, 2)
		assert.NotNil(t, builder.onAdvance[testState1])
		assert.NotNil(t, builder.onAdvance[testState2])
	})
}

// TestWithPrompt tests prompt generator configuration.
func TestWithPrompt(t *testing.T) {
	t.Run("stores prompt generator for state", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		generator := func(_ *state.Project) string {
			return "test prompt"
		}

		builder.WithPrompt(testState1, generator)

		assert.NotNil(t, builder.prompts[testState1])
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		generator := func(_ *state.Project) string {
			return "test prompt"
		}

		result := builder.WithPrompt(testState1, generator)

		assert.Same(t, builder, result)
	})

	t.Run("multiple states can have prompts", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		generator1 := func(_ *state.Project) string {
			return "prompt1"
		}
		generator2 := func(_ *state.Project) string {
			return "prompt2"
		}

		builder.WithPrompt(testState1, generator1)
		builder.WithPrompt(testState2, generator2)

		assert.Len(t, builder.prompts, 2)
		assert.NotNil(t, builder.prompts[testState1])
		assert.NotNil(t, builder.prompts[testState2])
	})
}

// TestBuild tests the Build method.
func TestBuild(t *testing.T) {
	t.Run("creates ProjectTypeConfig with all data", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)

		// Add all types of configuration
		builder.WithPhase(testPhaseName, WithStartState(testState1))
		builder.SetInitialState(testState1)
		builder.AddTransition(testState1, testState2, testEvent1)

		determiner := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}
		builder.OnAdvance(testState1, determiner)

		generator := func(_ *state.Project) string {
			return "test prompt"
		}
		builder.WithPrompt(testState1, generator)

		config := builder.Build()

		require.NotNil(t, config)
		assert.Equal(t, testProjectName, config.name)
		assert.Len(t, config.phaseConfigs, 1)
		assert.Equal(t, testState1, config.initialState)
		assert.Len(t, config.transitions, 1)
		assert.Len(t, config.onAdvance, 1)
		assert.Len(t, config.prompts, 1)
	})

	t.Run("copies name from builder", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		config := builder.Build()

		assert.Equal(t, testProjectName, config.name)
	})

	t.Run("copies phase configs", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		builder.WithPhase("phase1")
		builder.WithPhase("phase2")

		config := builder.Build()

		assert.Len(t, config.phaseConfigs, 2)
		assert.NotNil(t, config.phaseConfigs["phase1"])
		assert.NotNil(t, config.phaseConfigs["phase2"])
	})

	t.Run("copies initial state", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		builder.SetInitialState(testState1)

		config := builder.Build()

		assert.Equal(t, testState1, config.initialState)
	})

	t.Run("copies transitions", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		builder.AddTransition(testState1, testState2, testEvent1)

		config := builder.Build()

		require.Len(t, config.transitions, 1)
		assert.Equal(t, testState1, config.transitions[0].From)
	})

	t.Run("copies onAdvance map", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		determiner := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}
		builder.OnAdvance(testState1, determiner)

		config := builder.Build()

		assert.Len(t, config.onAdvance, 1)
		assert.NotNil(t, config.onAdvance[testState1])
	})

	t.Run("copies prompts map", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		generator := func(_ *state.Project) string {
			return "test"
		}
		builder.WithPrompt(testState1, generator)

		config := builder.Build()

		assert.Len(t, config.prompts, 1)
		assert.NotNil(t, config.prompts[testState1])
	})

	t.Run("builder is reusable - can call Build multiple times", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		builder.WithPhase(testPhaseName)
		builder.SetInitialState(testState1)

		config1 := builder.Build()
		config2 := builder.Build()

		// Both configs should exist and be separate instances
		require.NotNil(t, config1)
		require.NotNil(t, config2)
		assert.NotSame(t, config1, config2)

		// But should have same content
		assert.Equal(t, config1.name, config2.name)
		assert.Len(t, config2.phaseConfigs, 1)
		assert.Equal(t, testState1, config2.initialState)
	})

	t.Run("builder is reusable - can modify after Build", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder(testProjectName)
		builder.WithPhase("phase1")

		config1 := builder.Build()
		assert.Len(t, config1.phaseConfigs, 1)

		// Modify builder after first build
		builder.WithPhase("phase2")

		config2 := builder.Build()
		assert.Len(t, config2.phaseConfigs, 2)

		// First config should be unchanged
		assert.Len(t, config1.phaseConfigs, 1)
	})
}

// TestChaining tests that all methods can be chained.
func TestChaining(t *testing.T) {
	t.Run("all methods can be chained in single expression", func(t *testing.T) {
		determiner := func(_ *state.Project) (sdkstate.Event, error) {
			return testEvent1, nil
		}
		generator := func(_ *state.Project) string {
			return "test"
		}

		config := NewProjectTypeConfigBuilder(testProjectName).
			WithPhase("phase1", WithStartState(testState1)).
			WithPhase("phase2", WithEndState(testState2)).
			SetInitialState(testState1).
			AddTransition(testState1, testState2, testEvent1).
			OnAdvance(testState1, determiner).
			WithPrompt(testState1, generator).
			Build()

		require.NotNil(t, config)
		assert.Equal(t, testProjectName, config.name)
		assert.Len(t, config.phaseConfigs, 2)
		assert.Equal(t, testState1, config.initialState)
		assert.Len(t, config.transitions, 1)
		assert.Len(t, config.onAdvance, 1)
		assert.Len(t, config.prompts, 1)
	})
}
