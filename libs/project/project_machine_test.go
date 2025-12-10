package project

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test states and events for project machine tests.
const (
	pmTestStatePlanningActive State = "PlanningActive"
	pmTestStateImplPlanning   State = "ImplementationPlanning"
	pmTestStateImplExecuting  State = "ImplementationExecuting"
	pmTestStateReviewActive   State = "ReviewActive"

	pmTestEventAdvancePlanning Event = "AdvancePlanning"
	pmTestEventStartExecution  Event = "StartExecution"
	pmTestEventStartReview     Event = "StartReview"
)

// createTestProject creates a test project with phases for testing.
func createTestProject() *state.Project {
	proj := &state.Project{}
	proj.Name = "test-project"
	proj.Statechart.Current_state = string(pmTestStatePlanningActive)
	proj.Phases = map[string]project.PhaseState{
		"planning": {
			Status: "in_progress",
		},
		"implementation": {
			Status: "pending",
		},
		"review": {
			Status: "pending",
		},
	}
	return proj
}

//nolint:funlen // Test function with many subtests
func TestProjectTypeConfig_BuildProjectMachine(t *testing.T) {
	t.Parallel()

	t.Run("builds machine with initial state", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		require.NotNil(t, machine)
		assert.Equal(t, pmTestStatePlanningActive, machine.State())
	})

	t.Run("built machine can fire transitions", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		err := machine.Fire(pmTestEventAdvancePlanning)

		require.NoError(t, err)
		assert.Equal(t, pmTestStateImplPlanning, machine.State())
	})

	t.Run("guards are bound to project instance", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
				WithProjectGuard("all tasks complete", func(p *state.Project) bool {
					return p.AllTasksComplete()
				}),
			).
			Build()

		// Project with no tasks (vacuously true)
		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		err := machine.Fire(pmTestEventAdvancePlanning)
		require.NoError(t, err)
		assert.Equal(t, pmTestStateImplPlanning, machine.State())
	})

	t.Run("guards block transition when returning false", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
				WithProjectGuard("plan approved", func(p *state.Project) bool {
					// Check for approved artifact
					return p.PhaseOutputApproved("planning", "plan")
				}),
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		err := machine.Fire(pmTestEventAdvancePlanning)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "plan approved")
	})

	t.Run("onEntry actions are bound to project", func(t *testing.T) {
		t.Parallel()

		var enteredProject *state.Project
		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
				WithProjectOnEntry(func(p *state.Project) error {
					enteredProject = p
					return nil
				}),
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		err := machine.Fire(pmTestEventAdvancePlanning)

		require.NoError(t, err)
		assert.Same(t, proj, enteredProject)
	})

	t.Run("onExit actions are bound to project", func(t *testing.T) {
		t.Parallel()

		var exitedProject *state.Project
		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
				WithProjectOnExit(func(p *state.Project) error {
					exitedProject = p
					return nil
				}),
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		err := machine.Fire(pmTestEventAdvancePlanning)

		require.NoError(t, err)
		assert.Same(t, proj, exitedProject)
	})

	t.Run("prompts are bound to project", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			WithPrompt(pmTestStatePlanningActive, func(p *state.Project) string {
				return "Working on: " + p.Name
			}).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		prompt := machine.Prompt()
		assert.Equal(t, "Working on: test-project", prompt)
	})

	t.Run("prompts update with state changes", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			WithPrompt(pmTestStatePlanningActive, func(*state.Project) string {
				return "Planning"
			}).
			WithPrompt(pmTestStateImplPlanning, func(*state.Project) string {
				return "Implementing"
			}).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		assert.Equal(t, "Planning", machine.Prompt())

		require.NoError(t, machine.Fire(pmTestEventAdvancePlanning))

		assert.Equal(t, "Implementing", machine.Prompt())
	})

	t.Run("no prompt returns empty string", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		assert.Empty(t, machine.Prompt())
	})

	t.Run("multiple transitions work correctly", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
			).
			AddTransition(
				pmTestStateImplPlanning,
				pmTestStateImplExecuting,
				pmTestEventStartExecution,
			).
			AddTransition(
				pmTestStateImplExecuting,
				pmTestStateReviewActive,
				pmTestEventStartReview,
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		require.NoError(t, machine.Fire(pmTestEventAdvancePlanning))
		assert.Equal(t, pmTestStateImplPlanning, machine.State())

		require.NoError(t, machine.Fire(pmTestEventStartExecution))
		assert.Equal(t, pmTestStateImplExecuting, machine.State())

		require.NoError(t, machine.Fire(pmTestEventStartReview))
		assert.Equal(t, pmTestStateReviewActive, machine.State())
	})
}

func TestProjectTypeConfig_CanFire(t *testing.T) {
	t.Parallel()

	config := NewProjectTypeConfigBuilder("test").
		SetInitialState(pmTestStatePlanningActive).
		AddTransition(
			pmTestStatePlanningActive,
			pmTestStateImplPlanning,
			pmTestEventAdvancePlanning,
		).
		Build()

	t.Run("returns true for valid event", func(t *testing.T) {
		t.Parallel()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		assert.True(t, machine.CanFire(pmTestEventAdvancePlanning))
	})

	t.Run("returns false for invalid event", func(t *testing.T) {
		t.Parallel()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		assert.False(t, machine.CanFire(pmTestEventStartExecution))
	})
}

func TestProjectTypeConfig_PermittedTriggers(t *testing.T) {
	t.Parallel()

	t.Run("returns available events", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(pmTestStatePlanningActive).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateImplPlanning,
				pmTestEventAdvancePlanning,
			).
			AddTransition(
				pmTestStatePlanningActive,
				pmTestStateReviewActive,
				pmTestEventStartReview,
			).
			Build()

		proj := createTestProject()
		machine := config.BuildProjectMachine(proj, pmTestStatePlanningActive)

		events := machine.PermittedTriggers()

		assert.Len(t, events, 2)
		assert.Contains(t, events, pmTestEventAdvancePlanning)
		assert.Contains(t, events, pmTestEventStartReview)
	})
}
