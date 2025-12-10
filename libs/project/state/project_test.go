package state

import (
	"context"
	"testing"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/qmuntal/stateless"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProject(t *testing.T) {
	t.Run("creates wrapper with embedded state", func(t *testing.T) {
		state := project.ProjectState{
			Name:   "test-project",
			Type:   "standard",
			Branch: "feat/test",
			Phases: map[string]project.PhaseState{
				"planning": {Status: "pending"},
			},
		}
		backend := NewMemoryBackend()

		proj := NewProject(state, backend)

		require.NotNil(t, proj)
		assert.Equal(t, "test-project", proj.Name)
		assert.Equal(t, "standard", proj.Type)
		assert.Equal(t, "feat/test", proj.Branch)
		assert.Len(t, proj.Phases, 1)
	})

	t.Run("stores backend reference", func(t *testing.T) {
		state := project.ProjectState{Name: "test"}
		backend := NewMemoryBackend()

		proj := NewProject(state, backend)

		assert.Equal(t, backend, proj.Backend())
	})

	t.Run("initializes with nil config and machine", func(t *testing.T) {
		state := project.ProjectState{Name: "test"}
		backend := NewMemoryBackend()

		proj := NewProject(state, backend)

		assert.Nil(t, proj.Config())
		assert.Nil(t, proj.Machine())
	})
}

func TestProject_Config(t *testing.T) {
	t.Run("returns nil when not set", func(t *testing.T) {
		proj := NewProject(project.ProjectState{}, NewMemoryBackend())

		assert.Nil(t, proj.Config())
	})

	t.Run("returns config after SetConfig", func(t *testing.T) {
		proj := NewProject(project.ProjectState{}, NewMemoryBackend())
		config := &mockProjectTypeConfig{name: "standard"}

		proj.SetConfig(config)

		require.NotNil(t, proj.Config())
		assert.Equal(t, "standard", proj.Config().Name())
	})
}

func TestProject_Machine(t *testing.T) {
	t.Run("returns nil when not set", func(t *testing.T) {
		proj := NewProject(project.ProjectState{}, NewMemoryBackend())

		assert.Nil(t, proj.Machine())
	})

	t.Run("returns machine after SetMachine", func(t *testing.T) {
		proj := NewProject(project.ProjectState{}, NewMemoryBackend())
		machine := stateless.NewStateMachine("initial")

		proj.SetMachine(machine)

		assert.Equal(t, machine, proj.Machine())
	})
}

func TestProject_Backend(t *testing.T) {
	t.Run("returns the backend passed to constructor", func(t *testing.T) {
		backend := NewMemoryBackend()
		proj := NewProject(project.ProjectState{}, backend)

		assert.Equal(t, backend, proj.Backend())
	})
}

func TestProject_PhaseOutputApproved(t *testing.T) {
	tests := []struct {
		name       string
		phases     map[string]project.PhaseState
		phaseName  string
		outputType string
		want       bool
	}{
		{
			name:       "returns false when phase not found",
			phases:     map[string]project.PhaseState{},
			phaseName:  "nonexistent",
			outputType: "review",
			want:       false,
		},
		{
			name: "returns false when no outputs",
			phases: map[string]project.PhaseState{
				"planning": {Status: "completed", Outputs: []project.ArtifactState{}},
			},
			phaseName:  "planning",
			outputType: "review",
			want:       false,
		},
		{
			name: "returns false when output exists but not approved",
			phases: map[string]project.PhaseState{
				"planning": {
					Status: "completed",
					Outputs: []project.ArtifactState{
						{Type: "review", Approved: false},
					},
				},
			},
			phaseName:  "planning",
			outputType: "review",
			want:       false,
		},
		{
			name: "returns false when approved output has different type",
			phases: map[string]project.PhaseState{
				"planning": {
					Status: "completed",
					Outputs: []project.ArtifactState{
						{Type: "design_doc", Approved: true},
					},
				},
			},
			phaseName:  "planning",
			outputType: "review",
			want:       false,
		},
		{
			name: "returns true when matching output is approved",
			phases: map[string]project.PhaseState{
				"planning": {
					Status: "completed",
					Outputs: []project.ArtifactState{
						{Type: "review", Approved: true},
					},
				},
			},
			phaseName:  "planning",
			outputType: "review",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := project.ProjectState{
				Name:   "test",
				Phases: tt.phases,
			}
			proj := NewProject(state, NewMemoryBackend())

			got := proj.PhaseOutputApproved(tt.phaseName, tt.outputType)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProject_PhaseMetadataBool(t *testing.T) {
	tests := []struct {
		name      string
		phases    map[string]project.PhaseState
		phaseName string
		key       string
		want      bool
	}{
		{
			name:      "returns false when phase not found",
			phases:    map[string]project.PhaseState{},
			phaseName: "nonexistent",
			key:       "approved",
			want:      false,
		},
		{
			name: "returns false when metadata is nil",
			phases: map[string]project.PhaseState{
				"planning": {Status: "pending", Metadata: nil},
			},
			phaseName: "planning",
			key:       "approved",
			want:      false,
		},
		{
			name: "returns false when key not found",
			phases: map[string]project.PhaseState{
				"planning": {
					Status:   "pending",
					Metadata: map[string]any{"other": true},
				},
			},
			phaseName: "planning",
			key:       "approved",
			want:      false,
		},
		{
			name: "returns false when value is not bool",
			phases: map[string]project.PhaseState{
				"planning": {
					Status:   "pending",
					Metadata: map[string]any{"approved": "yes"},
				},
			},
			phaseName: "planning",
			key:       "approved",
			want:      false,
		},
		{
			name: "returns false when bool value is false",
			phases: map[string]project.PhaseState{
				"planning": {
					Status:   "pending",
					Metadata: map[string]any{"approved": false},
				},
			},
			phaseName: "planning",
			key:       "approved",
			want:      false,
		},
		{
			name: "returns true when bool value is true",
			phases: map[string]project.PhaseState{
				"planning": {
					Status:   "pending",
					Metadata: map[string]any{"approved": true},
				},
			},
			phaseName: "planning",
			key:       "approved",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := project.ProjectState{
				Name:   "test",
				Phases: tt.phases,
			}
			proj := NewProject(state, NewMemoryBackend())

			got := proj.PhaseMetadataBool(tt.phaseName, tt.key)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProject_AllTasksComplete(t *testing.T) {
	tests := []struct {
		name   string
		phases map[string]project.PhaseState
		want   bool
	}{
		{
			name:   "returns true when no phases",
			phases: map[string]project.PhaseState{},
			want:   true,
		},
		{
			name: "returns true when phases have no tasks",
			phases: map[string]project.PhaseState{
				"planning": {Status: "completed", Tasks: []project.TaskState{}},
			},
			want: true,
		},
		{
			name: "returns false when any task is not completed",
			phases: map[string]project.PhaseState{
				"planning": {
					Status: "completed",
					Tasks: []project.TaskState{
						{Id: "010", Status: "completed"},
						{Id: "020", Status: "in_progress"},
					},
				},
			},
			want: false,
		},
		{
			name: "returns true when all tasks across all phases are completed",
			phases: map[string]project.PhaseState{
				"planning": {
					Tasks: []project.TaskState{
						{Id: "010", Status: "completed"},
					},
				},
				"implementation": {
					Tasks: []project.TaskState{
						{Id: "020", Status: "completed"},
						{Id: "030", Status: "completed"},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := project.ProjectState{
				Name:   "test",
				Phases: tt.phases,
			}
			proj := NewProject(state, NewMemoryBackend())

			got := proj.AllTasksComplete()

			assert.Equal(t, tt.want, got)
		})
	}
}

// mockProjectTypeConfig is a minimal implementation for testing.
type mockProjectTypeConfig struct {
	name string
}

func (m *mockProjectTypeConfig) Name() string {
	return m.name
}

// Mock method to verify type assertion - test that our interface is minimal but functional.
func TestProjectTypeConfig_Interface(t *testing.T) {
	// This test verifies that our interface works correctly with mocks.
	var config ProjectTypeConfig = &mockProjectTypeConfig{name: "test"}
	assert.Equal(t, "test", config.Name())
}

// TestProject_Save verifies that Save persists the project state through the backend.
func TestProject_Save(t *testing.T) {
	t.Run("saves project state to backend", func(t *testing.T) {
		backend := NewMemoryBackend()
		state := project.ProjectState{
			Name:   "test-project",
			Type:   "standard",
			Branch: "feat/test",
		}
		proj := NewProject(state, backend)

		err := proj.Save(context.Background())
		require.NoError(t, err)

		// Verify it was saved
		loaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "test-project", loaded.Name)
		assert.Equal(t, "standard", loaded.Type)
	})
}
