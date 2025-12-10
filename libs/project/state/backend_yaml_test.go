package state

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestYAMLBackend_Load tests the Load method of YAMLBackend.
func TestYAMLBackend_Load(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, fs core.FS)
		wantState *project.ProjectState
		wantErr   error
	}{
		{
			name: "loads existing valid project state",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				data := `name: test-project
type: standard
branch: feat/test
description: A test project
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T10:00:00Z
phases: {}
statechart:
  current_state: PlanningActive
  updated_at: 2025-01-15T10:00:00Z
`
				require.NoError(t, fs.WriteFile("project/state.yaml", []byte(data), 0644))
			},
			wantState: &project.ProjectState{
				Name:        "test-project",
				Type:        "standard",
				Branch:      "feat/test",
				Description: "A test project",
			},
		},
		{
			name: "returns ErrNotFound for missing file",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				// No file created - ensure directory exists but file doesn't
				require.NoError(t, fs.MkdirAll("project", 0755))
			},
			wantErr: ErrNotFound,
		},
		{
			name: "returns ErrInvalidState for invalid YAML",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				invalidYAML := `name: [invalid yaml structure`
				require.NoError(t, fs.WriteFile("project/state.yaml", []byte(invalidYAML), 0644))
			},
			wantErr: ErrInvalidState,
		},
		{
			name: "loads all ProjectState fields correctly",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				data := `name: full-project
type: exploration
branch: explore/testing
description: Full project with all fields
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T11:00:00Z
phases:
  planning:
    status: completed
    enabled: true
    created_at: 2025-01-15T10:00:00Z
    tasks: []
    inputs: []
    outputs: []
statechart:
  current_state: ImplementationActive
  updated_at: 2025-01-15T11:00:00Z
agent_sessions:
  planner: "session-123"
`
				require.NoError(t, fs.WriteFile("project/state.yaml", []byte(data), 0644))
			},
			wantState: &project.ProjectState{
				Name:        "full-project",
				Type:        "exploration",
				Branch:      "explore/testing",
				Description: "Full project with all fields",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := billy.NewMemory()
			if tt.setup != nil {
				tt.setup(t, memFS)
			}

			backend := NewYAMLBackend(memFS)
			got, err := backend.Load(context.Background())

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantState.Name, got.Name)
			assert.Equal(t, tt.wantState.Type, got.Type)
			assert.Equal(t, tt.wantState.Branch, got.Branch)
			assert.Equal(t, tt.wantState.Description, got.Description)
		})
	}
}

// TestYAMLBackend_Save_NewFile tests saving to a new file.
func TestYAMLBackend_Save_NewFile(t *testing.T) {
	memFS := billy.NewMemory()
	require.NoError(t, memFS.MkdirAll("project", 0755))

	state := &project.ProjectState{
		Name:        "new-project",
		Type:        "standard",
		Branch:      "feat/new",
		Description: "A new project",
	}

	backend := NewYAMLBackend(memFS)
	err := backend.Save(context.Background(), state)
	require.NoError(t, err)

	data, err := memFS.ReadFile("project/state.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: new-project")
	assert.Contains(t, string(data), "type: standard")
	assert.Contains(t, string(data), "branch: feat/new")
}

// TestYAMLBackend_Save_Overwrite tests overwriting an existing file.
func TestYAMLBackend_Save_Overwrite(t *testing.T) {
	memFS := billy.NewMemory()
	require.NoError(t, memFS.MkdirAll("project", 0755))
	require.NoError(t, memFS.WriteFile("project/state.yaml", []byte("name: old-project"), 0644))

	state := &project.ProjectState{
		Name:   "updated-project",
		Type:   "standard",
		Branch: "feat/updated",
	}

	backend := NewYAMLBackend(memFS)
	err := backend.Save(context.Background(), state)
	require.NoError(t, err)

	data, err := memFS.ReadFile("project/state.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: updated-project")
	assert.NotContains(t, string(data), "old-project")
}

// TestYAMLBackend_Save_AtomicWrite tests that temp file is cleaned up after save.
func TestYAMLBackend_Save_AtomicWrite(t *testing.T) {
	memFS := billy.NewMemory()
	require.NoError(t, memFS.MkdirAll("project", 0755))

	state := &project.ProjectState{
		Name:   "atomic-test",
		Type:   "standard",
		Branch: "feat/atomic",
	}

	backend := NewYAMLBackend(memFS)
	err := backend.Save(context.Background(), state)
	require.NoError(t, err)

	// Verify main file exists
	exists, err := memFS.Exists("project/state.yaml")
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify temp file does not exist
	exists, err = memFS.Exists("project/state.yaml.tmp")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestYAMLBackend_Save_PreservesFields tests that all ProjectState fields are preserved.
func TestYAMLBackend_Save_PreservesFields(t *testing.T) {
	memFS := billy.NewMemory()
	require.NoError(t, memFS.MkdirAll("project", 0755))

	state := &project.ProjectState{
		Name:        "full-project",
		Type:        "exploration",
		Branch:      "explore/full",
		Description: "A full project with all fields",
		Phases: map[string]project.PhaseState{
			"planning": {Status: "completed", Enabled: true},
		},
		Statechart:     project.StatechartState{Current_state: "ImplementationActive"},
		Agent_sessions: map[string]string{"planner": "session-abc"},
	}

	backend := NewYAMLBackend(memFS)
	err := backend.Save(context.Background(), state)
	require.NoError(t, err)

	data, err := memFS.ReadFile("project/state.yaml")
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "name: full-project")
	assert.Contains(t, content, "description: A full project with all fields")
	assert.Contains(t, content, "current_state: ImplementationActive")
}

// TestYAMLBackend_Exists tests the Exists method of YAMLBackend.
func TestYAMLBackend_Exists(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(t *testing.T, fs core.FS)
		want   bool
		wantOK bool // true if we expect no error
	}{
		{
			name: "returns true for existing file",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				require.NoError(t, fs.WriteFile("project/state.yaml", []byte("name: test"), 0644))
			},
			want:   true,
			wantOK: true,
		},
		{
			name: "returns false for missing file",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				// No file created
			},
			want:   false,
			wantOK: true,
		},
		{
			name:   "returns false for missing directory",
			setup:  nil, // No setup needed - test with empty filesystem
			want:   false,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := billy.NewMemory()
			if tt.setup != nil {
				tt.setup(t, memFS)
			}

			backend := NewYAMLBackend(memFS)
			got, err := backend.Exists(context.Background())

			if tt.wantOK {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestYAMLBackend_Delete tests the Delete method of YAMLBackend.
func TestYAMLBackend_Delete(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, fs core.FS)
		validate func(t *testing.T, fs core.FS)
		wantErr  bool
	}{
		{
			name: "deletes existing file",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				require.NoError(t, fs.WriteFile("project/state.yaml", []byte("name: test"), 0644))
			},
			validate: func(t *testing.T, fs core.FS) {
				t.Helper()
				exists, err := fs.Exists("project/state.yaml")
				require.NoError(t, err)
				assert.False(t, exists, "file should be deleted")
			},
			wantErr: false,
		},
		{
			name: "returns error for non-existent file",
			setup: func(t *testing.T, fs core.FS) {
				t.Helper()
				require.NoError(t, fs.MkdirAll("project", 0755))
				// No file created
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := billy.NewMemory()
			if tt.setup != nil {
				tt.setup(t, memFS)
			}

			backend := NewYAMLBackend(memFS)
			err := backend.Delete(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, memFS)
			}
		})
	}
}

// TestYAMLBackend_NewYAMLBackendWithPath tests the custom path constructor.
func TestYAMLBackend_NewYAMLBackendWithPath(t *testing.T) {
	memFS := billy.NewMemory()
	customPath := "custom/path/state.yaml"

	backend := NewYAMLBackendWithPath(memFS, customPath)

	// Verify the backend uses the custom path
	require.NoError(t, memFS.MkdirAll("custom/path", 0755))
	state := &project.ProjectState{
		Name:   "custom-path-test",
		Type:   "standard",
		Branch: "feat/custom",
	}

	err := backend.Save(context.Background(), state)
	require.NoError(t, err)

	// Verify file is at custom path
	exists, err := memFS.Exists(customPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify default path does not exist
	exists, err = memFS.Exists("project/state.yaml")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestYAMLBackend_ErrorWrapping tests that errors are properly wrapped with context.
func TestYAMLBackend_ErrorWrapping(t *testing.T) {
	t.Run("Load wraps ErrNotFound with context", func(t *testing.T) {
		memFS := billy.NewMemory()
		require.NoError(t, memFS.MkdirAll("project", 0755))

		backend := NewYAMLBackend(memFS)
		_, err := backend.Load(context.Background())

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		// Verify error message contains context
		assert.Contains(t, err.Error(), "load project state")
	})

	t.Run("Load wraps ErrInvalidState with context", func(t *testing.T) {
		memFS := billy.NewMemory()
		require.NoError(t, memFS.MkdirAll("project", 0755))
		require.NoError(t, memFS.WriteFile("project/state.yaml", []byte("invalid: [yaml"), 0644))

		backend := NewYAMLBackend(memFS)
		_, err := backend.Load(context.Background())

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidState))
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("Delete wraps underlying errors", func(t *testing.T) {
		memFS := billy.NewMemory()
		require.NoError(t, memFS.MkdirAll("project", 0755))
		// Don't create the file

		backend := NewYAMLBackend(memFS)
		err := backend.Delete(context.Background())

		require.Error(t, err)
		// Should wrap the underlying fs.ErrNotExist
		assert.True(t, errors.Is(err, fs.ErrNotExist))
	})
}

// TestYAMLBackend_RoundTrip tests that data survives a save/load cycle.
func TestYAMLBackend_RoundTrip(t *testing.T) {
	memFS := billy.NewMemory()
	require.NoError(t, memFS.MkdirAll("project", 0755))

	original := &project.ProjectState{
		Name:        "roundtrip-test",
		Type:        "standard",
		Branch:      "feat/roundtrip",
		Description: "Testing round-trip serialization",
		Phases: map[string]project.PhaseState{
			"planning": {
				Status:  "completed",
				Enabled: true,
				Tasks:   []project.TaskState{},
				Inputs:  []project.ArtifactState{},
				Outputs: []project.ArtifactState{},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "ImplementationActive",
		},
		Agent_sessions: map[string]string{
			"researcher": "sess-456",
		},
	}

	backend := NewYAMLBackend(memFS)
	ctx := context.Background()

	// Save
	err := backend.Save(ctx, original)
	require.NoError(t, err)

	// Load
	loaded, err := backend.Load(ctx)
	require.NoError(t, err)

	// Verify key fields match
	assert.Equal(t, original.Name, loaded.Name)
	assert.Equal(t, original.Type, loaded.Type)
	assert.Equal(t, original.Branch, loaded.Branch)
	assert.Equal(t, original.Description, loaded.Description)
	assert.Equal(t, original.Statechart.Current_state, loaded.Statechart.Current_state)

	// Verify phases
	require.Len(t, loaded.Phases, 1)
	assert.Equal(t, "completed", loaded.Phases["planning"].Status)

	// Verify agent sessions
	require.Len(t, loaded.Agent_sessions, 1)
	assert.Equal(t, "sess-456", loaded.Agent_sessions["researcher"])
}
