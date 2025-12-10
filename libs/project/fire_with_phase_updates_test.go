package project

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test states and events for fire with phase updates tests.
const (
	fwpuTestStatePlanningActive State = "PlanningActive"
	fwpuTestStateImplPlanning   State = "ImplementationPlanning"
	fwpuTestStateImplExecuting  State = "ImplementationExecuting"
	fwpuTestStateReviewActive   State = "ReviewActive"
	fwpuTestStateFinalize       State = "FinalizeChecks"

	fwpuTestEventAdvancePlanning Event = "AdvancePlanning"
	fwpuTestEventStartExecution  Event = "StartExecution"
	fwpuTestEventReviewPass      Event = "ReviewPass"
	fwpuTestEventReviewFail      Event = "ReviewFail"
)

// createFwpuTestProject creates a test project with phases for testing.
func createFwpuTestProject() *state.Project {
	proj := &state.Project{}
	proj.Name = "test-project"
	proj.Statechart.Current_state = string(fwpuTestStatePlanningActive)
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
func TestProjectTypeConfig_FireWithPhaseUpdates(t *testing.T) {
	t.Parallel()

	t.Run("fires event and transitions state", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStatePlanningActive).
			WithPhase("planning",
				WithStartState(fwpuTestStatePlanningActive),
				WithEndState(fwpuTestStatePlanningActive),
			).
			WithPhase("implementation",
				WithStartState(fwpuTestStateImplPlanning),
				WithEndState(fwpuTestStateImplExecuting),
			).
			AddTransition(
				fwpuTestStatePlanningActive,
				fwpuTestStateImplPlanning,
				fwpuTestEventAdvancePlanning,
			).
			Build()

		proj := createFwpuTestProject()
		machine := config.BuildProjectMachine(proj, fwpuTestStatePlanningActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventAdvancePlanning, proj)

		require.NoError(t, err)
		assert.Equal(t, fwpuTestStateImplPlanning, machine.State())
	})

	t.Run("marks exiting phase as completed", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStatePlanningActive).
			WithPhase("planning",
				WithStartState(fwpuTestStatePlanningActive),
				WithEndState(fwpuTestStatePlanningActive),
			).
			WithPhase("implementation",
				WithStartState(fwpuTestStateImplPlanning),
				WithEndState(fwpuTestStateImplExecuting),
			).
			AddTransition(
				fwpuTestStatePlanningActive,
				fwpuTestStateImplPlanning,
				fwpuTestEventAdvancePlanning,
			).
			Build()

		proj := createFwpuTestProject()
		machine := config.BuildProjectMachine(proj, fwpuTestStatePlanningActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventAdvancePlanning, proj)

		require.NoError(t, err)
		assert.Equal(t, "completed", proj.Phases["planning"].Status)
		assert.False(t, proj.Phases["planning"].Completed_at.IsZero())
	})

	t.Run("marks entering phase as in_progress", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStatePlanningActive).
			WithPhase("planning",
				WithStartState(fwpuTestStatePlanningActive),
				WithEndState(fwpuTestStatePlanningActive),
			).
			WithPhase("implementation",
				WithStartState(fwpuTestStateImplPlanning),
				WithEndState(fwpuTestStateImplExecuting),
			).
			AddTransition(
				fwpuTestStatePlanningActive,
				fwpuTestStateImplPlanning,
				fwpuTestEventAdvancePlanning,
			).
			Build()

		proj := createFwpuTestProject()
		machine := config.BuildProjectMachine(proj, fwpuTestStatePlanningActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventAdvancePlanning, proj)

		require.NoError(t, err)
		assert.Equal(t, "in_progress", proj.Phases["implementation"].Status)
		assert.False(t, proj.Phases["implementation"].Started_at.IsZero())
	})

	t.Run("does not update phase if not entering start state", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStateImplPlanning).
			WithPhase("implementation",
				WithStartState(fwpuTestStateImplPlanning),
				WithEndState(fwpuTestStateImplExecuting),
			).
			AddTransition(
				fwpuTestStateImplPlanning,
				fwpuTestStateImplExecuting,
				fwpuTestEventStartExecution,
			).
			Build()

		proj := createFwpuTestProject()
		proj.Phases["implementation"] = project.PhaseState{
			Status: "in_progress",
		}
		machine := config.BuildProjectMachine(proj, fwpuTestStateImplPlanning)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventStartExecution, proj)

		require.NoError(t, err)
		// Implementation is still in_progress - we're not entering a start state
		assert.Equal(t, "in_progress", proj.Phases["implementation"].Status)
	})

	t.Run("marks phase as failed when WithProjectFailedPhase is used", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStateReviewActive).
			WithPhase("review",
				WithStartState(fwpuTestStateReviewActive),
				WithEndState(fwpuTestStateReviewActive),
			).
			WithPhase("implementation",
				WithStartState(fwpuTestStateImplPlanning),
				WithEndState(fwpuTestStateImplExecuting),
			).
			AddTransition(
				fwpuTestStateReviewActive,
				fwpuTestStateImplPlanning,
				fwpuTestEventReviewFail,
				WithProjectFailedPhase("review"),
			).
			Build()

		proj := createFwpuTestProject()
		proj.Phases["review"] = project.PhaseState{
			Status: "in_progress",
		}
		machine := config.BuildProjectMachine(proj, fwpuTestStateReviewActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventReviewFail, proj)

		require.NoError(t, err)
		assert.Equal(t, "failed", proj.Phases["review"].Status)
		assert.False(t, proj.Phases["review"].Failed_at.IsZero())
	})

	t.Run("marks phase as completed (not failed) for success path", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStateReviewActive).
			WithPhase("review",
				WithStartState(fwpuTestStateReviewActive),
				WithEndState(fwpuTestStateReviewActive),
			).
			WithPhase("finalize",
				WithStartState(fwpuTestStateFinalize),
				WithEndState(fwpuTestStateFinalize),
			).
			AddTransition(
				fwpuTestStateReviewActive,
				fwpuTestStateFinalize,
				fwpuTestEventReviewPass,
			).
			Build()

		proj := createFwpuTestProject()
		proj.Phases["review"] = project.PhaseState{
			Status: "in_progress",
		}
		proj.Phases["finalize"] = project.PhaseState{
			Status: "pending",
		}
		machine := config.BuildProjectMachine(proj, fwpuTestStateReviewActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventReviewPass, proj)

		require.NoError(t, err)
		assert.Equal(t, "completed", proj.Phases["review"].Status)
		assert.False(t, proj.Phases["review"].Completed_at.IsZero())
	})

	t.Run("returns error for invalid event", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStatePlanningActive).
			AddTransition(
				fwpuTestStatePlanningActive,
				fwpuTestStateImplPlanning,
				fwpuTestEventAdvancePlanning,
			).
			Build()

		proj := createFwpuTestProject()
		machine := config.BuildProjectMachine(proj, fwpuTestStatePlanningActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventStartExecution, proj)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "transition failed")
	})

	t.Run("does not modify phase status if already not pending when entering", func(t *testing.T) {
		t.Parallel()

		config := NewProjectTypeConfigBuilder("test").
			SetInitialState(fwpuTestStateReviewActive).
			WithPhase("review",
				WithStartState(fwpuTestStateReviewActive),
				WithEndState(fwpuTestStateReviewActive),
			).
			WithPhase("implementation",
				WithStartState(fwpuTestStateImplPlanning),
				WithEndState(fwpuTestStateImplExecuting),
			).
			AddTransition(
				fwpuTestStateReviewActive,
				fwpuTestStateImplPlanning,
				fwpuTestEventReviewFail,
				WithProjectFailedPhase("review"),
			).
			Build()

		proj := createFwpuTestProject()
		proj.Phases["review"] = project.PhaseState{
			Status: "in_progress",
		}
		// Implementation already in_progress (rework scenario)
		proj.Phases["implementation"] = project.PhaseState{
			Status: "in_progress",
		}
		machine := config.BuildProjectMachine(proj, fwpuTestStateReviewActive)

		err := config.FireWithPhaseUpdates(machine, fwpuTestEventReviewFail, proj)

		require.NoError(t, err)
		// Implementation stays in_progress (MarkPhaseInProgress only updates from "pending")
		assert.Equal(t, "in_progress", proj.Phases["implementation"].Status)
	})
}
