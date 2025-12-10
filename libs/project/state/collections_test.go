package state

import (
	"testing"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhaseCollection_Get(t *testing.T) {
	tests := []struct {
		name      string
		phases    PhaseCollection
		phaseName string
		wantErr   bool
	}{
		{
			name:      "returns error for empty collection",
			phases:    PhaseCollection{},
			phaseName: "planning",
			wantErr:   true,
		},
		{
			name: "returns error when phase not found",
			phases: PhaseCollection{
				"implementation": &Phase{PhaseState: project.PhaseState{Status: "pending"}},
			},
			phaseName: "planning",
			wantErr:   true,
		},
		{
			name: "returns phase when found",
			phases: PhaseCollection{
				"planning": &Phase{PhaseState: project.PhaseState{Status: "completed"}},
			},
			phaseName: "planning",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.phases.Get(tt.phaseName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "completed", got.Status)
			}
		})
	}
}

func TestArtifactCollection_Get(t *testing.T) {
	tests := []struct {
		name      string
		artifacts ArtifactCollection
		index     int
		wantErr   bool
		wantType  string
	}{
		{
			name:      "returns error for empty collection",
			artifacts: ArtifactCollection{},
			index:     0,
			wantErr:   true,
		},
		{
			name: "returns error for negative index",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "review"}},
			},
			index:   -1,
			wantErr: true,
		},
		{
			name: "returns error for index equal to length",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "review"}},
			},
			index:   1,
			wantErr: true,
		},
		{
			name: "returns error for index greater than length",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "review"}},
			},
			index:   5,
			wantErr: true,
		},
		{
			name: "returns artifact at valid index",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "design_doc"}},
				{ArtifactState: project.ArtifactState{Type: "review"}},
			},
			index:    1,
			wantErr:  false,
			wantType: "review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.artifacts.Get(tt.index)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "out of range")
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantType, got.Type)
			}
		})
	}
}

func TestArtifactCollection_Add(t *testing.T) {
	t.Run("appends artifact to empty collection", func(t *testing.T) {
		var coll ArtifactCollection

		err := coll.Add(Artifact{ArtifactState: project.ArtifactState{Type: "review"}})

		require.NoError(t, err)
		assert.Len(t, coll, 1)
		assert.Equal(t, "review", coll[0].Type)
	})

	t.Run("appends artifact to existing collection", func(t *testing.T) {
		coll := ArtifactCollection{
			{ArtifactState: project.ArtifactState{Type: "design_doc"}},
		}

		err := coll.Add(Artifact{ArtifactState: project.ArtifactState{Type: "review"}})

		require.NoError(t, err)
		assert.Len(t, coll, 2)
		assert.Equal(t, "design_doc", coll[0].Type)
		assert.Equal(t, "review", coll[1].Type)
	})
}

func TestArtifactCollection_Remove(t *testing.T) {
	tests := []struct {
		name      string
		artifacts ArtifactCollection
		index     int
		wantErr   bool
		wantLen   int
	}{
		{
			name:      "returns error for empty collection",
			artifacts: ArtifactCollection{},
			index:     0,
			wantErr:   true,
		},
		{
			name: "returns error for negative index",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "review"}},
			},
			index:   -1,
			wantErr: true,
		},
		{
			name: "returns error for index equal to length",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "review"}},
			},
			index:   1,
			wantErr: true,
		},
		{
			name: "removes artifact at valid index",
			artifacts: ArtifactCollection{
				{ArtifactState: project.ArtifactState{Type: "design_doc"}},
				{ArtifactState: project.ArtifactState{Type: "review"}},
				{ArtifactState: project.ArtifactState{Type: "task_list"}},
			},
			index:   1,
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.artifacts.Remove(tt.index)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "out of range")
			} else {
				require.NoError(t, err)
				assert.Len(t, tt.artifacts, tt.wantLen)
				// Verify the middle element was removed
				assert.Equal(t, "design_doc", tt.artifacts[0].Type)
				assert.Equal(t, "task_list", tt.artifacts[1].Type)
			}
		})
	}
}

func TestTaskCollection_Get(t *testing.T) {
	tests := []struct {
		name    string
		tasks   TaskCollection
		id      string
		wantErr bool
	}{
		{
			name:    "returns error for empty collection",
			tasks:   TaskCollection{},
			id:      "010",
			wantErr: true,
		},
		{
			name: "returns error when task not found",
			tasks: TaskCollection{
				{TaskState: project.TaskState{Id: "010", Name: "Task 1"}},
				{TaskState: project.TaskState{Id: "020", Name: "Task 2"}},
			},
			id:      "030",
			wantErr: true,
		},
		{
			name: "returns task when found",
			tasks: TaskCollection{
				{TaskState: project.TaskState{Id: "010", Name: "Task 1"}},
				{TaskState: project.TaskState{Id: "020", Name: "Task 2"}},
			},
			id:      "020",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tasks.Get(tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "Task 2", got.Name)
			}
		})
	}
}

func TestTaskCollection_Add(t *testing.T) {
	t.Run("appends task to empty collection", func(t *testing.T) {
		var coll TaskCollection

		err := coll.Add(Task{TaskState: project.TaskState{Id: "010", Name: "Task 1"}})

		require.NoError(t, err)
		assert.Len(t, coll, 1)
		assert.Equal(t, "010", coll[0].Id)
	})

	t.Run("appends task to existing collection", func(t *testing.T) {
		coll := TaskCollection{
			{TaskState: project.TaskState{Id: "010", Name: "Task 1"}},
		}

		err := coll.Add(Task{TaskState: project.TaskState{Id: "020", Name: "Task 2"}})

		require.NoError(t, err)
		assert.Len(t, coll, 2)
		assert.Equal(t, "010", coll[0].Id)
		assert.Equal(t, "020", coll[1].Id)
	})
}

func TestTaskCollection_Remove(t *testing.T) {
	tests := []struct {
		name    string
		tasks   TaskCollection
		id      string
		wantErr bool
		wantLen int
	}{
		{
			name:    "returns error for empty collection",
			tasks:   TaskCollection{},
			id:      "010",
			wantErr: true,
		},
		{
			name: "returns error when task not found",
			tasks: TaskCollection{
				{TaskState: project.TaskState{Id: "010", Name: "Task 1"}},
			},
			id:      "020",
			wantErr: true,
		},
		{
			name: "removes task by id",
			tasks: TaskCollection{
				{TaskState: project.TaskState{Id: "010", Name: "Task 1"}},
				{TaskState: project.TaskState{Id: "020", Name: "Task 2"}},
				{TaskState: project.TaskState{Id: "030", Name: "Task 3"}},
			},
			id:      "020",
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tasks.Remove(tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Len(t, tt.tasks, tt.wantLen)
				// Verify the middle task was removed
				assert.Equal(t, "010", tt.tasks[0].Id)
				assert.Equal(t, "030", tt.tasks[1].Id)
			}
		})
	}
}
