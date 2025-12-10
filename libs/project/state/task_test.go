package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
)

func TestTask_EmbeddedState(t *testing.T) {
	t.Run("embeds TaskState fields", func(t *testing.T) {
		now := time.Now()
		task := Task{
			TaskState: project.TaskState{
				Id:             "010",
				Name:           "Implement feature",
				Phase:          "implementation",
				Status:         "in_progress",
				Created_at:     now,
				Iteration:      1,
				Assigned_agent: "implementer",
			},
		}

		assert.Equal(t, "010", task.Id)
		assert.Equal(t, "Implement feature", task.Name)
		assert.Equal(t, "implementation", task.Phase)
		assert.Equal(t, "in_progress", task.Status)
		assert.Equal(t, now, task.Created_at)
		assert.Equal(t, int64(1), task.Iteration)
		assert.Equal(t, "implementer", task.Assigned_agent)
	})

	t.Run("accesses nested inputs and outputs", func(t *testing.T) {
		task := Task{
			TaskState: project.TaskState{
				Id:   "020",
				Name: "Test task",
				Inputs: []project.ArtifactState{
					{Type: "reference", Path: "/path/to/input"},
				},
				Outputs: []project.ArtifactState{
					{Type: "modified", Path: "/path/to/output"},
				},
			},
		}

		assert.Len(t, task.Inputs, 1)
		assert.Equal(t, "reference", task.Inputs[0].Type)
		assert.Len(t, task.Outputs, 1)
		assert.Equal(t, "modified", task.Outputs[0].Type)
	})
}
