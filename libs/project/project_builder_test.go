package project

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test states and events for builder tests.
const (
	builderTestStatePlanningActive State = "PlanningActive"
	builderTestStateImplPlanning   State = "ImplementationPlanning"
	builderTestStateImplExecuting  State = "ImplementationExecuting"
	builderTestStateReviewActive   State = "ReviewActive"
	builderTestStateFinalize       State = "FinalizeChecks"

	builderTestEventAdvancePlanning Event = "AdvancePlanning"
	builderTestEventStartImpl       Event = "StartImplementation"
	builderTestEventReviewPass      Event = "ReviewPass"
	builderTestEventReviewFail      Event = "ReviewFail"
)

func TestNewProjectTypeConfigBuilder(t *testing.T) {
	t.Parallel()

	t.Run("creates builder with name", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("standard")

		require.NotNil(t, builder)
		config := builder.Build()
		assert.Equal(t, "standard", config.Name())
	})

	t.Run("creates builder with initialized maps", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		require.NotNil(t, builder)
		config := builder.Build()
		assert.NotNil(t, config.phaseConfigs)
		assert.NotNil(t, config.onAdvance)
		assert.NotNil(t, config.prompts)
		assert.NotNil(t, config.branches)
	})
}

func TestProjectTypeConfigBuilder_SetInitialState(t *testing.T) {
	t.Parallel()

	t.Run("sets initial state", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive).
			Build()

		assert.Equal(t, builderTestStatePlanningActive, config.InitialState())
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.SetInitialState(builderTestStatePlanningActive)

		assert.Same(t, builder, result)
	})
}

func TestProjectTypeConfigBuilder_WithPhase(t *testing.T) {
	t.Parallel()

	t.Run("adds phase with name", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			WithPhase("planning").
			Build()

		phases := config.Phases()
		require.NotNil(t, phases["planning"])
		assert.Equal(t, "planning", phases["planning"].Name())
	})

	t.Run("adds phase with options", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			WithPhase("implementation",
				WithStartState(builderTestStateImplPlanning),
				WithEndState(builderTestStateImplExecuting),
				WithInputs("task_list"),
				WithOutputs("code", "tests"),
				WithTasks(),
				WithMetadataSchema(`{ field: string }`),
			).
			Build()

		phase := config.Phases()["implementation"]
		require.NotNil(t, phase)
		assert.Equal(t, builderTestStateImplPlanning, phase.StartState())
		assert.Equal(t, builderTestStateImplExecuting, phase.EndState())
		assert.Equal(t, []string{"task_list"}, phase.AllowedInputTypes())
		assert.Equal(t, []string{"code", "tests"}, phase.AllowedOutputTypes())
		assert.True(t, phase.SupportsTasks())
		assert.Equal(t, `{ field: string }`, phase.MetadataSchema())
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.WithPhase("planning")

		assert.Same(t, builder, result)
	})

	t.Run("adds multiple phases", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			WithPhase("planning").
			WithPhase("implementation").
			WithPhase("review").
			Build()

		phases := config.Phases()
		assert.Len(t, phases, 3)
		assert.NotNil(t, phases["planning"])
		assert.NotNil(t, phases["implementation"])
		assert.NotNil(t, phases["review"])
	})
}

func TestProjectTypeConfigBuilder_AddTransition(t *testing.T) {
	t.Parallel()

	t.Run("adds transition", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive).
			AddTransition(
				builderTestStatePlanningActive,
				builderTestStateImplPlanning,
				builderTestEventAdvancePlanning,
			).
			Build()

		tc := config.GetTransition(
			builderTestStatePlanningActive,
			builderTestStateImplPlanning,
			builderTestEventAdvancePlanning,
		)
		require.NotNil(t, tc)
		assert.Equal(t, builderTestStatePlanningActive, tc.From)
		assert.Equal(t, builderTestStateImplPlanning, tc.To)
		assert.Equal(t, builderTestEventAdvancePlanning, tc.Event)
	})

	t.Run("adds transition with options", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive).
			AddTransition(
				builderTestStatePlanningActive,
				builderTestStateImplPlanning,
				builderTestEventAdvancePlanning,
				WithProjectGuard("all artifacts approved", func(*state.Project) bool { return true }),
				WithProjectDescription("Complete planning phase"),
				WithProjectFailedPhase("planning"),
			).
			Build()

		tc := config.GetTransition(
			builderTestStatePlanningActive,
			builderTestStateImplPlanning,
			builderTestEventAdvancePlanning,
		)
		require.NotNil(t, tc)
		assert.Equal(t, "all artifacts approved", tc.GuardDescription())
		assert.Equal(t, "Complete planning phase", tc.Description())
		assert.Equal(t, "planning", tc.FailedPhase())
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.AddTransition(
			builderTestStatePlanningActive,
			builderTestStateImplPlanning,
			builderTestEventAdvancePlanning,
		)

		assert.Same(t, builder, result)
	})

	t.Run("adds multiple transitions", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive).
			AddTransition(
				builderTestStatePlanningActive,
				builderTestStateImplPlanning,
				builderTestEventAdvancePlanning,
			).
			AddTransition(
				builderTestStateImplPlanning,
				builderTestStateImplExecuting,
				builderTestEventStartImpl,
			).
			Build()

		tc1 := config.GetTransition(
			builderTestStatePlanningActive,
			builderTestStateImplPlanning,
			builderTestEventAdvancePlanning,
		)
		tc2 := config.GetTransition(
			builderTestStateImplPlanning,
			builderTestStateImplExecuting,
			builderTestEventStartImpl,
		)
		assert.NotNil(t, tc1)
		assert.NotNil(t, tc2)
	})
}

func TestProjectTypeConfigBuilder_OnAdvance(t *testing.T) {
	t.Parallel()

	t.Run("sets event determiner", func(t *testing.T) {
		t.Parallel()

		determiner := func(*state.Project) (Event, error) {
			return builderTestEventAdvancePlanning, nil
		}

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive).
			OnAdvance(builderTestStatePlanningActive, determiner).
			Build()

		// Create a minimal project with the initial state
		proj := &state.Project{}
		proj.Statechart.Current_state = string(builderTestStatePlanningActive)

		event, err := config.DetermineEvent(proj)
		require.NoError(t, err)
		assert.Equal(t, builderTestEventAdvancePlanning, event)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.OnAdvance(builderTestStatePlanningActive, func(*state.Project) (Event, error) {
			return builderTestEventAdvancePlanning, nil
		})

		assert.Same(t, builder, result)
	})
}

func TestProjectTypeConfigBuilder_AddBranch(t *testing.T) {
	t.Parallel()

	t.Run("adds branch with discriminator and paths", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStateReviewActive).
			AddBranch(builderTestStateReviewActive,
				BranchOn(func(*state.Project) string { return "pass" }),
				When("pass", builderTestEventReviewPass, builderTestStateFinalize),
				When("fail", builderTestEventReviewFail, builderTestStateImplPlanning),
			).
			Build()

		// Verify it's a branching state
		assert.True(t, config.IsBranchingState(builderTestStateReviewActive))

		// Verify transitions were generated
		transitions := config.GetAvailableTransitions(builderTestStateReviewActive)
		assert.Len(t, transitions, 2)
	})

	t.Run("adds branch with path options", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStateReviewActive).
			AddBranch(builderTestStateReviewActive,
				BranchOn(func(*state.Project) string { return "fail" }),
				When("pass", builderTestEventReviewPass, builderTestStateFinalize,
					WithProjectDescription("Review approved"),
				),
				When("fail", builderTestEventReviewFail, builderTestStateImplPlanning,
					WithProjectDescription("Review rejected"),
					WithProjectFailedPhase("review"),
				),
			).
			Build()

		// Verify branch info
		tc := config.GetTransition(
			builderTestStateReviewActive,
			builderTestStateImplPlanning,
			builderTestEventReviewFail,
		)
		require.NotNil(t, tc)
		assert.Equal(t, "Review rejected", tc.Description())
		assert.Equal(t, "review", tc.FailedPhase())
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.AddBranch(builderTestStateReviewActive,
			BranchOn(func(*state.Project) string { return "pass" }),
			When("pass", builderTestEventReviewPass, builderTestStateFinalize),
		)

		assert.Same(t, builder, result)
	})
}

func TestProjectTypeConfigBuilder_WithPrompt(t *testing.T) {
	t.Parallel()

	t.Run("sets prompt generator", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive).
			WithPrompt(builderTestStatePlanningActive, func(*state.Project) string {
				return "Create your plan"
			}).
			Build()

		prompt := config.GetStatePrompt(builderTestStatePlanningActive, &state.Project{})
		assert.Equal(t, "Create your plan", prompt)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.WithPrompt(builderTestStatePlanningActive, func(*state.Project) string {
			return "Test"
		})

		assert.Same(t, builder, result)
	})
}

func TestProjectTypeConfigBuilder_WithOrchestratorPrompt(t *testing.T) {
	t.Parallel()

	t.Run("sets orchestrator prompt", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			WithOrchestratorPrompt(func(*state.Project) string {
				return "Orchestrator guidance"
			}).
			Build()

		prompt := config.OrchestratorPrompt(&state.Project{})
		assert.Equal(t, "Orchestrator guidance", prompt)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.WithOrchestratorPrompt(func(*state.Project) string {
			return "Test"
		})

		assert.Same(t, builder, result)
	})
}

func TestProjectTypeConfigBuilder_WithInitializer(t *testing.T) {
	t.Parallel()

	t.Run("sets initializer", func(t *testing.T) {
		t.Parallel()

		called := false
		config := NewProjectTypeConfigBuilder("test").
			WithInitializer(func(*state.Project, map[string][]project.ArtifactState) error {
				called = true
				return nil
			}).
			Build()

		err := config.Initialize(&state.Project{}, nil)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test")

		result := builder.WithInitializer(func(*state.Project, map[string][]project.ArtifactState) error {
			return nil
		})

		assert.Same(t, builder, result)
	})
}

func TestProjectTypeConfigBuilder_Build(t *testing.T) {
	t.Parallel()

	t.Run("builds complete config", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("standard").
			SetInitialState(builderTestStatePlanningActive).
			WithPhase("planning",
				WithStartState(builderTestStatePlanningActive),
				WithEndState(builderTestStatePlanningActive),
			).
			WithPhase("implementation",
				WithStartState(builderTestStateImplPlanning),
				WithEndState(builderTestStateImplExecuting),
				WithTasks(),
			).
			AddTransition(
				builderTestStatePlanningActive,
				builderTestStateImplPlanning,
				builderTestEventAdvancePlanning,
			).
			WithPrompt(builderTestStatePlanningActive, func(*state.Project) string {
				return "Plan"
			}).
			Build()

		assert.Equal(t, "standard", config.Name())
		assert.Equal(t, builderTestStatePlanningActive, config.InitialState())
		assert.Len(t, config.Phases(), 2)
		assert.NotNil(t, config.GetTransition(
			builderTestStatePlanningActive,
			builderTestStateImplPlanning,
			builderTestEventAdvancePlanning,
		))
		assert.Equal(t, "Plan", config.GetStatePrompt(builderTestStatePlanningActive, &state.Project{}))
	})

	t.Run("builder can be reused", func(t *testing.T) {
		t.Parallel()

		builder := NewProjectTypeConfigBuilder("test").
			SetInitialState(builderTestStatePlanningActive)

		config1 := builder.Build()
		config2 := builder.Build()

		// Both configs should have the same values
		assert.Equal(t, config1.Name(), config2.Name())
		assert.Equal(t, config1.InitialState(), config2.InitialState())

		// But should be different instances
		assert.NotSame(t, config1, config2)
	})
}
