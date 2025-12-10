package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase_EmbeddedState(t *testing.T) {
	t.Run("embeds PhaseState fields", func(t *testing.T) {
		now := time.Now()
		phase := Phase{
			PhaseState: project.PhaseState{
				Status:     "in_progress",
				Enabled:    true,
				Created_at: now,
				Iteration:  2,
			},
		}

		assert.Equal(t, "in_progress", phase.Status)
		assert.True(t, phase.Enabled)
		assert.Equal(t, now, phase.Created_at)
		assert.Equal(t, int64(2), phase.Iteration)
	})
}

func TestIncrementPhaseIteration(t *testing.T) {
	tests := []struct {
		name          string
		phases        map[string]project.PhaseState
		phaseName     string
		wantIteration int64
		wantErr       bool
	}{
		{
			name:      "returns error when phase not found",
			phases:    map[string]project.PhaseState{},
			phaseName: "nonexistent",
			wantErr:   true,
		},
		{
			name: "increments from 0 to 1",
			phases: map[string]project.PhaseState{
				"planning": {Status: "pending", Iteration: 0},
			},
			phaseName:     "planning",
			wantIteration: 1,
			wantErr:       false,
		},
		{
			name: "increments from 1 to 2",
			phases: map[string]project.PhaseState{
				"planning": {Status: "in_progress", Iteration: 1},
			},
			phaseName:     "planning",
			wantIteration: 2,
			wantErr:       false,
		},
		{
			name: "increments from 5 to 6",
			phases: map[string]project.PhaseState{
				"implementation": {Status: "in_progress", Iteration: 5},
			},
			phaseName:     "implementation",
			wantIteration: 6,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := NewProject(project.ProjectState{
				Name:   "test",
				Phases: tt.phases,
			}, NewMemoryBackend())

			err := IncrementPhaseIteration(proj, tt.phaseName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantIteration, proj.Phases[tt.phaseName].Iteration)
			}
		})
	}
}

func TestMarkPhaseFailed(t *testing.T) {
	t.Run("returns error when phase not found", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name:   "test",
			Phases: map[string]project.PhaseState{},
		}, NewMemoryBackend())

		err := MarkPhaseFailed(proj, "nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("sets status to failed and records timestamp", func(t *testing.T) {
		before := time.Now()
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"review": {Status: "in_progress"},
			},
		}, NewMemoryBackend())

		err := MarkPhaseFailed(proj, "review")
		after := time.Now()

		require.NoError(t, err)
		phase := proj.Phases["review"]
		assert.Equal(t, "failed", phase.Status)
		assert.False(t, phase.Failed_at.IsZero())
		assert.True(t, phase.Failed_at.After(before) || phase.Failed_at.Equal(before))
		assert.True(t, phase.Failed_at.Before(after) || phase.Failed_at.Equal(after))
	})
}

func TestMarkPhaseInProgress(t *testing.T) {
	t.Run("returns error when phase not found", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name:   "test",
			Phases: map[string]project.PhaseState{},
		}, NewMemoryBackend())

		err := MarkPhaseInProgress(proj, "nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("sets status and timestamp when status is pending", func(t *testing.T) {
		before := time.Now()
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"planning": {Status: "pending"},
			},
		}, NewMemoryBackend())

		err := MarkPhaseInProgress(proj, "planning")
		after := time.Now()

		require.NoError(t, err)
		phase := proj.Phases["planning"]
		assert.Equal(t, "in_progress", phase.Status)
		assert.False(t, phase.Started_at.IsZero())
		assert.True(t, phase.Started_at.After(before) || phase.Started_at.Equal(before))
		assert.True(t, phase.Started_at.Before(after) || phase.Started_at.Equal(after))
	})

	t.Run("does not change status when already in_progress", func(t *testing.T) {
		originalTime := time.Now().Add(-time.Hour)
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"planning": {Status: "in_progress", Started_at: originalTime},
			},
		}, NewMemoryBackend())

		err := MarkPhaseInProgress(proj, "planning")

		require.NoError(t, err)
		phase := proj.Phases["planning"]
		assert.Equal(t, "in_progress", phase.Status)
		assert.Equal(t, originalTime, phase.Started_at)
	})

	t.Run("does not change status when completed", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"planning": {Status: "completed"},
			},
		}, NewMemoryBackend())

		err := MarkPhaseInProgress(proj, "planning")

		require.NoError(t, err)
		assert.Equal(t, "completed", proj.Phases["planning"].Status)
	})
}

func TestMarkPhaseCompleted(t *testing.T) {
	t.Run("returns error when phase not found", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name:   "test",
			Phases: map[string]project.PhaseState{},
		}, NewMemoryBackend())

		err := MarkPhaseCompleted(proj, "nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("sets status to completed and records timestamp", func(t *testing.T) {
		before := time.Now()
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"planning": {Status: "in_progress"},
			},
		}, NewMemoryBackend())

		err := MarkPhaseCompleted(proj, "planning")
		after := time.Now()

		require.NoError(t, err)
		phase := proj.Phases["planning"]
		assert.Equal(t, "completed", phase.Status)
		assert.False(t, phase.Completed_at.IsZero())
		assert.True(t, phase.Completed_at.After(before) || phase.Completed_at.Equal(before))
		assert.True(t, phase.Completed_at.Before(after) || phase.Completed_at.Equal(after))
	})
}

func TestAddPhaseInputFromOutput_Errors(t *testing.T) {
	t.Run("returns error when source phase not found", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name:   "test",
			Phases: map[string]project.PhaseState{"target": {}},
		}, NewMemoryBackend())

		err := AddPhaseInputFromOutput(proj, "nonexistent", "target", "review", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source phase")
	})

	t.Run("returns error when target phase not found", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name:   "test",
			Phases: map[string]project.PhaseState{"source": {}},
		}, NewMemoryBackend())

		err := AddPhaseInputFromOutput(proj, "source", "nonexistent", "review", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "target phase")
	})

	t.Run("returns error when no matching artifact found", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"source": {Outputs: []project.ArtifactState{{Type: "design_doc", Path: "/design"}}},
				"target": {},
			},
		}, NewMemoryBackend())

		err := AddPhaseInputFromOutput(proj, "source", "target", "review", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no matching artifact")
	})
}

func TestAddPhaseInputFromOutput_Success(t *testing.T) {
	t.Run("copies matching artifact without filter", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"source": {Outputs: []project.ArtifactState{{Type: "review", Path: "/path/to/review"}}},
				"target": {Inputs: []project.ArtifactState{}},
			},
		}, NewMemoryBackend())

		err := AddPhaseInputFromOutput(proj, "source", "target", "review", nil)

		require.NoError(t, err)
		require.Len(t, proj.Phases["target"].Inputs, 1)
		assert.Equal(t, "review", proj.Phases["target"].Inputs[0].Type)
	})

	t.Run("returns latest matching artifact when multiple exist", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"source": {Outputs: []project.ArtifactState{
					{Type: "review", Path: "/old/review"},
					{Type: "review", Path: "/new/review"},
				}},
				"target": {},
			},
		}, NewMemoryBackend())

		err := AddPhaseInputFromOutput(proj, "source", "target", "review", nil)

		require.NoError(t, err)
		require.Len(t, proj.Phases["target"].Inputs, 1)
		assert.Equal(t, "/new/review", proj.Phases["target"].Inputs[0].Path)
	})

	t.Run("applies filter to select artifact", func(t *testing.T) {
		proj := NewProject(project.ProjectState{
			Name: "test",
			Phases: map[string]project.PhaseState{
				"source": {Outputs: []project.ArtifactState{
					{Type: "review", Path: "/failed/review", Metadata: map[string]any{"assessment": "fail"}},
					{Type: "review", Path: "/passed/review", Metadata: map[string]any{"assessment": "pass"}},
				}},
				"target": {},
			},
		}, NewMemoryBackend())

		filter := func(a *project.ArtifactState) bool {
			assessment, ok := a.Metadata["assessment"].(string)
			return ok && assessment == "fail"
		}

		err := AddPhaseInputFromOutput(proj, "source", "target", "review", filter)

		require.NoError(t, err)
		require.Len(t, proj.Phases["target"].Inputs, 1)
		assert.Equal(t, "/failed/review", proj.Phases["target"].Inputs[0].Path)
	})
}
