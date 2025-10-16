package sowfs

import (
	"testing"
	"time"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectFS_State tests reading project state
func TestProjectFS_State(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*billy.MemoryFS)
		wantErr bool
		errType error
	}{
		{
			name: "read existing state",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
				stateYAML := `project:
  name: test-project
  branch: feature/test
  description: Test project
  created_at: 2024-01-01T00:00:00Z
  updated_at: 2024-01-02T00:00:00Z
phases:
  discovery:
    status: pending
    created_at: 2024-01-01T00:00:00Z
    started_at: null
    completed_at: null
    enabled: false
    discovery_type: null
    artifacts: []
  design:
    status: pending
    created_at: 2024-01-01T00:00:00Z
    started_at: null
    completed_at: null
    enabled: false
    architect_used: null
    artifacts: []
  implementation:
    status: in_progress
    created_at: 2024-01-01T00:00:00Z
    started_at: "2024-01-01T00:00:00Z"
    completed_at: null
    enabled: true
    planner_used: null
    tasks: []
    pending_task_additions: null
  review:
    status: pending
    created_at: 2024-01-01T00:00:00Z
    started_at: null
    completed_at: null
    enabled: true
    iteration: 1
    reports: []
  finalize:
    status: pending
    created_at: 2024-01-01T00:00:00Z
    started_at: null
    completed_at: null
    enabled: true
    documentation_updates: null
    artifacts_moved: null
    project_deleted: false
    pr_url: null
`
				fs.WriteFile(".sow/project/state.yaml", []byte(stateYAML), 0644)
			},
			wantErr: false,
		},
		{
			name: "project not found",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
				// Don't create state.yaml
			},
			wantErr: true,
			errType: ErrProjectNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			projectFS := sowFS.project
			if projectFS == nil {
				projectFS = NewProjectFS(sowFS, sowFS.validator)
			}

			state, err := projectFS.State()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, state)
				assert.Equal(t, "test-project", state.Project.Name)
			}
		})
	}
}

// TestProjectFS_WriteState tests writing project state
func TestProjectFS_WriteState(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	projectFS := NewProjectFS(sowFS, sowFS.validator)

	// Create test state
	now := time.Now()
	state := &schemas.ProjectState{}
	state.Project.Name = "test-project"
	state.Project.Branch = "feature/test"
	state.Project.Description = "Test project"
	state.Project.Created_at = now
	state.Project.Updated_at = now

	// Initialize all phases with valid data
	state.Phases.Discovery.Status = "pending"
	state.Phases.Discovery.Created_at = now
	state.Phases.Discovery.Started_at = nil
	state.Phases.Discovery.Completed_at = nil
	state.Phases.Discovery.Enabled = false
	state.Phases.Discovery.Discovery_type = nil
	state.Phases.Discovery.Artifacts = []schemas.Artifact{}

	state.Phases.Design.Status = "pending"
	state.Phases.Design.Created_at = now
	state.Phases.Design.Started_at = nil
	state.Phases.Design.Completed_at = nil
	state.Phases.Design.Enabled = false
	state.Phases.Design.Architect_used = nil
	state.Phases.Design.Artifacts = []schemas.Artifact{}

	state.Phases.Implementation.Status = "in_progress"
	state.Phases.Implementation.Created_at = now
	state.Phases.Implementation.Started_at = now.Format(time.RFC3339)
	state.Phases.Implementation.Completed_at = nil
	state.Phases.Implementation.Enabled = true
	state.Phases.Implementation.Planner_used = nil
	state.Phases.Implementation.Tasks = []schemas.Task{}
	state.Phases.Implementation.Pending_task_additions = nil

	state.Phases.Review.Status = "pending"
	state.Phases.Review.Created_at = now
	state.Phases.Review.Started_at = nil
	state.Phases.Review.Completed_at = nil
	state.Phases.Review.Enabled = true
	state.Phases.Review.Iteration = 1
	state.Phases.Review.Reports = []schemas.ReviewReport{}

	state.Phases.Finalize.Status = "pending"
	state.Phases.Finalize.Created_at = now
	state.Phases.Finalize.Started_at = nil
	state.Phases.Finalize.Completed_at = nil
	state.Phases.Finalize.Enabled = true
	state.Phases.Finalize.Documentation_updates = nil
	state.Phases.Finalize.Artifacts_moved = nil
	state.Phases.Finalize.Project_deleted = false
	state.Phases.Finalize.Pr_url = nil

	// Write state
	err = projectFS.WriteState(state)
	require.NoError(t, err)

	// Read back and verify
	readState, err := projectFS.State()
	require.NoError(t, err)
	assert.Equal(t, "test-project", readState.Project.Name)
	assert.Equal(t, "feature/test", readState.Project.Branch)
}

// TestProjectFS_AppendLog tests appending to project log
func TestProjectFS_AppendLog(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	projectFS := NewProjectFS(sowFS, sowFS.validator)

	// Append first entry
	err = projectFS.AppendLog("First log entry")
	require.NoError(t, err)

	// Append second entry
	err = projectFS.AppendLog("Second log entry")
	require.NoError(t, err)

	// Read log
	log, err := projectFS.ReadLog()
	require.NoError(t, err)
	assert.Contains(t, log, "First log entry")
	assert.Contains(t, log, "Second log entry")
}

// TestProjectFS_ReadLog tests reading project log
func TestProjectFS_ReadLog(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*billy.MemoryFS)
		want    string
		wantErr bool
	}{
		{
			name: "read existing log",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
				fs.WriteFile(".sow/project/log.md", []byte("Log content"), 0644)
			},
			want: "Log content",
		},
		{
			name: "log doesn't exist",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			projectFS := NewProjectFS(sowFS, sowFS.validator)

			log, err := projectFS.ReadLog()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, log)
			}
		})
	}
}

// TestProjectFS_Task tests getting a specific task
func TestProjectFS_Task(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*billy.MemoryFS)
		taskID  string
		wantErr bool
		errType error
	}{
		{
			name: "valid task exists",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
			},
			taskID:  "010",
			wantErr: false,
		},
		{
			name: "task not found",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks", 0755)
			},
			taskID:  "020",
			wantErr: true,
			errType: ErrTaskNotFound,
		},
		{
			name: "invalid task ID format",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks", 0755)
			},
			taskID:  "abc",
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "task ID not gap-numbered",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks", 0755)
			},
			taskID:  "011",
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			projectFS := NewProjectFS(sowFS, sowFS.validator)

			taskFS, err := projectFS.Task(tt.taskID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, taskFS)
				assert.Equal(t, tt.taskID, taskFS.ID())
			}
		})
	}
}

// TestProjectFS_Tasks tests listing all tasks
func TestProjectFS_Tasks(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*billy.MemoryFS)
		wantCount int
	}{
		{
			name: "multiple tasks",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
				fs.MkdirAll(".sow/project/phases/implementation/tasks/020", 0755)
				fs.MkdirAll(".sow/project/phases/implementation/tasks/030", 0755)
			},
			wantCount: 3,
		},
		{
			name: "no tasks directory",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
			},
			wantCount: 0,
		},
		{
			name: "empty tasks directory",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks", 0755)
			},
			wantCount: 0,
		},
		{
			name: "filters non-gap-numbered directories",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
				fs.MkdirAll(".sow/project/phases/implementation/tasks/020", 0755)
				fs.MkdirAll(".sow/project/phases/implementation/tasks/abc", 0755)
				fs.WriteFile(".sow/project/phases/implementation/tasks/file.txt", []byte("test"), 0644)
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			projectFS := NewProjectFS(sowFS, sowFS.validator)

			tasks, err := projectFS.Tasks()
			require.NoError(t, err)
			assert.Len(t, tasks, tt.wantCount)
		})
	}
}

// TestProjectFS_Exists tests checking if project exists
func TestProjectFS_Exists(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*billy.MemoryFS)
		exists bool
	}{
		{
			name: "project exists",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
				// Write valid (minimal) project state
				validState := `project:
  name: test-project
  branch: feat/test
  description: Test
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
phases:
  discovery:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  design:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  implementation:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    tasks: []
  review:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    iteration: 1
    reports: []
  finalize:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    project_deleted: false
`
				fs.WriteFile(".sow/project/state.yaml", []byte(validState), 0644)
			},
			exists: true,
		},
		{
			name: "project doesn't exist",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project", 0755)
			},
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			projectFS := NewProjectFS(sowFS, sowFS.validator)

			exists, err := projectFS.Exists()
			require.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

// TestProjectFS_Context tests context file operations
func TestProjectFS_Context(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	projectFS := NewProjectFS(sowFS, sowFS.validator)

	// Write context file
	err = projectFS.WriteContext("test.md", []byte("Context content"))
	require.NoError(t, err)

	// Write nested context file
	err = projectFS.WriteContext("dir/nested.md", []byte("Nested content"))
	require.NoError(t, err)

	// Read context file
	data, err := projectFS.ReadContext("test.md")
	require.NoError(t, err)
	assert.Equal(t, "Context content", string(data))

	// Read nested context file
	data, err = projectFS.ReadContext("dir/nested.md")
	require.NoError(t, err)
	assert.Equal(t, "Nested content", string(data))

	// List context files
	files, err := projectFS.ListContextFiles()
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "test.md")
	assert.Contains(t, files, "dir/nested.md")
}

// TestProjectFS_ListContextFiles_Empty tests listing when no context files
func TestProjectFS_ListContextFiles_Empty(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	projectFS := NewProjectFS(sowFS, sowFS.validator)

	files, err := projectFS.ListContextFiles()
	require.NoError(t, err)
	assert.Empty(t, files)
}
