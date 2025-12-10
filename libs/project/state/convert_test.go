package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
)

func TestConvertArtifacts(t *testing.T) {
	t.Run("converts empty slice", func(t *testing.T) {
		result := convertArtifacts(nil)

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("converts populated slice preserving all fields", func(t *testing.T) {
		now := time.Now()
		stateArtifacts := []project.ArtifactState{
			{
				Type:       "review",
				Path:       "/path/to/review",
				Approved:   true,
				Created_at: now,
				Metadata:   map[string]any{"assessment": "pass"},
			},
			{
				Type:       "design_doc",
				Path:       "/path/to/design",
				Approved:   false,
				Created_at: now.Add(-time.Hour),
				Metadata:   nil,
			},
		}

		result := convertArtifacts(stateArtifacts)

		assert.Len(t, result, 2)
		assert.Equal(t, "review", result[0].Type)
		assert.Equal(t, "/path/to/review", result[0].Path)
		assert.True(t, result[0].Approved)
		assert.Equal(t, now, result[0].Created_at)
		assert.Equal(t, "pass", result[0].Metadata["assessment"])
		assert.Equal(t, "design_doc", result[1].Type)
		assert.Nil(t, result[1].Metadata)
	})
}

func TestConvertArtifactsToState(t *testing.T) {
	t.Run("converts empty collection", func(t *testing.T) {
		result := convertArtifactsToState(ArtifactCollection{})

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("converts populated collection preserving all fields", func(t *testing.T) {
		now := time.Now()
		artifacts := ArtifactCollection{
			{ArtifactState: project.ArtifactState{
				Type:       "task_list",
				Path:       "/path/to/tasks",
				Approved:   true,
				Created_at: now,
			}},
		}

		result := convertArtifactsToState(artifacts)

		assert.Len(t, result, 1)
		assert.Equal(t, "task_list", result[0].Type)
		assert.Equal(t, "/path/to/tasks", result[0].Path)
		assert.True(t, result[0].Approved)
		assert.Equal(t, now, result[0].Created_at)
	})
}

func TestConvertTasks(t *testing.T) {
	t.Run("converts empty slice", func(t *testing.T) {
		result := convertTasks(nil)

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("converts populated slice preserving all fields", func(t *testing.T) {
		now := time.Now()
		stateTasks := []project.TaskState{
			{
				Id:             "010",
				Name:           "Implement feature",
				Phase:          "implementation",
				Status:         "completed",
				Created_at:     now,
				Iteration:      2,
				Assigned_agent: "implementer",
				Inputs: []project.ArtifactState{
					{Type: "reference", Path: "/input"},
				},
				Outputs: []project.ArtifactState{
					{Type: "modified", Path: "/output"},
				},
			},
		}

		result := convertTasks(stateTasks)

		assert.Len(t, result, 1)
		assert.Equal(t, "010", result[0].Id)
		assert.Equal(t, "Implement feature", result[0].Name)
		assert.Equal(t, "implementation", result[0].Phase)
		assert.Equal(t, "completed", result[0].Status)
		assert.Equal(t, int64(2), result[0].Iteration)
		assert.Len(t, result[0].Inputs, 1)
		assert.Len(t, result[0].Outputs, 1)
	})
}

func TestConvertTasksToState(t *testing.T) {
	t.Run("converts empty collection", func(t *testing.T) {
		result := convertTasksToState(TaskCollection{})

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("converts populated collection preserving all fields", func(t *testing.T) {
		tasks := TaskCollection{
			{TaskState: project.TaskState{
				Id:     "020",
				Name:   "Write tests",
				Status: "pending",
			}},
		}

		result := convertTasksToState(tasks)

		assert.Len(t, result, 1)
		assert.Equal(t, "020", result[0].Id)
		assert.Equal(t, "Write tests", result[0].Name)
		assert.Equal(t, "pending", result[0].Status)
	})
}

func TestConvertPhases(t *testing.T) {
	t.Run("converts empty map", func(t *testing.T) {
		result := convertPhases(nil)

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("converts populated map preserving all fields", func(t *testing.T) {
		now := time.Now()
		statePhases := map[string]project.PhaseState{
			"planning": {
				Status:     "completed",
				Enabled:    true,
				Created_at: now,
				Iteration:  1,
				Metadata:   map[string]any{"approved": true},
				Inputs: []project.ArtifactState{
					{Type: "reference", Path: "/input"},
				},
				Outputs: []project.ArtifactState{
					{Type: "task_list", Path: "/output"},
				},
				Tasks: []project.TaskState{
					{Id: "010", Name: "Design API"},
				},
			},
		}

		result := convertPhases(statePhases)

		assert.Len(t, result, 1)
		phase, exists := result["planning"]
		assert.True(t, exists)
		assert.Equal(t, "completed", phase.Status)
		assert.True(t, phase.Enabled)
		assert.Equal(t, now, phase.Created_at)
		assert.Equal(t, int64(1), phase.Iteration)
		assert.Equal(t, true, phase.Metadata["approved"])
		assert.Len(t, phase.Inputs, 1)
		assert.Len(t, phase.Outputs, 1)
		assert.Len(t, phase.Tasks, 1)
	})
}

func TestConvertPhasesToState(t *testing.T) {
	t.Run("converts empty collection", func(t *testing.T) {
		result := convertPhasesToState(PhaseCollection{})

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("converts populated collection preserving all fields", func(t *testing.T) {
		phases := PhaseCollection{
			"implementation": &Phase{PhaseState: project.PhaseState{
				Status:    "in_progress",
				Enabled:   true,
				Iteration: 2,
			}},
		}

		result := convertPhasesToState(phases)

		assert.Len(t, result, 1)
		phase, exists := result["implementation"]
		assert.True(t, exists)
		assert.Equal(t, "in_progress", phase.Status)
		assert.True(t, phase.Enabled)
		assert.Equal(t, int64(2), phase.Iteration)
	})
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("artifacts round-trip preserves all fields", func(t *testing.T) {
		now := time.Now()
		original := []project.ArtifactState{
			{
				Type:       "review",
				Path:       "/path",
				Approved:   true,
				Created_at: now,
				Metadata:   map[string]any{"key": "value"},
			},
		}

		converted := convertArtifacts(original)
		back := convertArtifactsToState(converted)

		assert.Equal(t, original, back)
	})

	t.Run("tasks round-trip preserves all fields", func(t *testing.T) {
		original := []project.TaskState{
			{
				Id:             "010",
				Name:           "Test task",
				Phase:          "planning",
				Status:         "pending",
				Iteration:      1,
				Assigned_agent: "implementer",
			},
		}

		converted := convertTasks(original)
		back := convertTasksToState(converted)

		assert.Equal(t, original, back)
	})

	t.Run("phases round-trip preserves all fields", func(t *testing.T) {
		original := map[string]project.PhaseState{
			"planning": {
				Status:    "completed",
				Enabled:   true,
				Iteration: 1,
			},
		}

		converted := convertPhases(original)
		back := convertPhasesToState(converted)

		assert.Equal(t, original, back)
	})
}
