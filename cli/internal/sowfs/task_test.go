package sowfs

import (
	"testing"
	"time"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskFS_ID tests getting task ID
func TestTaskFS_ID(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	assert.Equal(t, "010", taskFS.ID())
}

// TestTaskFS_State tests reading task state
func TestTaskFS_State(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*billy.MemoryFS)
		wantErr bool
		errType error
	}{
		{
			name: "read existing state",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
				stateYAML := `task:
  id: "010"
  name: test-task
  phase: implementation
  status: in_progress
  created_at: 2024-01-01T00:00:00Z
  started_at: "2024-01-01T00:00:00Z"
  updated_at: 2024-01-02T00:00:00Z
  completed_at: null
  iteration: 1
  references: []
  feedback: []
  files_modified: []
`
				fs.WriteFile(".sow/project/phases/implementation/tasks/010/state.yaml", []byte(stateYAML), 0644)
			},
			wantErr: false,
		},
		{
			name: "task state not found",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
			},
			wantErr: true,
			errType: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

			state, err := taskFS.State()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, state)
				assert.Equal(t, "010", state.Task.Id)
				assert.Equal(t, "test-task", state.Task.Name)
			}
		})
	}
}

// TestTaskFS_WriteState tests writing task state
func TestTaskFS_WriteState(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	// Create test state
	state := &schemas.TaskState{}
	state.Task.Id = "010"
	state.Task.Name = "test-task"
	state.Task.Phase = "implementation"
	state.Task.Status = "in_progress"
	state.Task.Created_at = time.Now()
	state.Task.Updated_at = time.Now()
	state.Task.Iteration = 1
	state.Task.References = []string{}
	state.Task.Feedback = []schemas.Feedback{}
	state.Task.Files_modified = []string{}

	// Write state
	err = taskFS.WriteState(state)
	require.NoError(t, err)

	// Read back and verify
	readState, err := taskFS.State()
	require.NoError(t, err)
	assert.Equal(t, "010", readState.Task.Id)
	assert.Equal(t, "test-task", readState.Task.Name)
}

// TestTaskFS_AppendLog tests appending to task log
func TestTaskFS_AppendLog(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	// Append first entry
	err = taskFS.AppendLog("First task log entry")
	require.NoError(t, err)

	// Append second entry
	err = taskFS.AppendLog("Second task log entry")
	require.NoError(t, err)

	// Read log
	log, err := taskFS.ReadLog()
	require.NoError(t, err)
	assert.Contains(t, log, "First task log entry")
	assert.Contains(t, log, "Second task log entry")
}

// TestTaskFS_ReadLog tests reading task log
func TestTaskFS_ReadLog(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*billy.MemoryFS)
		want    string
		wantErr bool
	}{
		{
			name: "read existing log",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
				fs.WriteFile(".sow/project/phases/implementation/tasks/010/log.md", []byte("Task log content"), 0644)
			},
			want: "Task log content",
		},
		{
			name: "log doesn't exist",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
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

			taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

			log, err := taskFS.ReadLog()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, log)
			}
		})
	}
}

// TestTaskFS_Description tests description operations
func TestTaskFS_Description(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	// Write description
	err = taskFS.WriteDescription("# Task Description\n\nImplement feature X")
	require.NoError(t, err)

	// Read description
	desc, err := taskFS.ReadDescription()
	require.NoError(t, err)
	assert.Contains(t, desc, "Task Description")
	assert.Contains(t, desc, "Implement feature X")
}

// TestTaskFS_Feedback tests feedback file operations
func TestTaskFS_Feedback(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	// Write feedback files
	err = taskFS.WriteFeedback("001.md", "First feedback")
	require.NoError(t, err)

	err = taskFS.WriteFeedback("002.md", "Second feedback")
	require.NoError(t, err)

	// List feedback
	files, err := taskFS.ListFeedback()
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "001.md")
	assert.Contains(t, files, "002.md")

	// Read feedback
	content, err := taskFS.ReadFeedback("001.md")
	require.NoError(t, err)
	assert.Equal(t, "First feedback", content)
}

// TestTaskFS_ListFeedback_Empty tests listing when no feedback
func TestTaskFS_ListFeedback_Empty(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	files, err := taskFS.ListFeedback()
	require.NoError(t, err)
	assert.Empty(t, files)
}

// TestTaskFS_WriteFeedback_PathTraversal tests path traversal protection
func TestTaskFS_WriteFeedback_PathTraversal(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	// Try path traversal in filename
	err = taskFS.WriteFeedback("../../../etc/passwd", "malicious content")
	require.NoError(t, err)

	// Verify file was written to safe location (sanitized to "passwd")
	files, err := taskFS.ListFeedback()
	require.NoError(t, err)
	assert.Contains(t, files, "passwd")

	// Verify content
	content, err := taskFS.ReadFeedback("passwd")
	require.NoError(t, err)
	assert.Equal(t, "malicious content", content)
}

// TestTaskFS_Path tests getting task path
func TestTaskFS_Path(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	path := taskFS.Path()
	assert.Equal(t, "/test/repo/.sow/project/phases/implementation/tasks/010", path)
}

// TestTaskFS_Integration tests full workflow
func TestTaskFS_Integration(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	taskFS := NewTaskFS(sowFS, "010", sowFS.validator)

	// Write state
	state := &schemas.TaskState{}
	state.Task.Id = "010"
	state.Task.Name = "integration-test"
	state.Task.Phase = "implementation"
	state.Task.Status = "in_progress"
	state.Task.Created_at = time.Now()
	state.Task.Updated_at = time.Now()
	state.Task.Iteration = 1
	state.Task.References = []string{"refs/style-guide"}
	state.Task.Feedback = []schemas.Feedback{}
	state.Task.Files_modified = []string{"src/main.go"}

	err = taskFS.WriteState(state)
	require.NoError(t, err)

	// Write description
	err = taskFS.WriteDescription("# Integration Test Task\n\nTest all features")
	require.NoError(t, err)

	// Append log entries
	err = taskFS.AppendLog("Started task")
	require.NoError(t, err)
	err = taskFS.AppendLog("Implemented feature")
	require.NoError(t, err)

	// Write feedback
	err = taskFS.WriteFeedback("001.md", "Please add tests")
	require.NoError(t, err)

	// Verify all data
	readState, err := taskFS.State()
	require.NoError(t, err)
	assert.Equal(t, "010", readState.Task.Id)
	assert.Equal(t, "integration-test", readState.Task.Name)

	desc, err := taskFS.ReadDescription()
	require.NoError(t, err)
	assert.Contains(t, desc, "Integration Test Task")

	log, err := taskFS.ReadLog()
	require.NoError(t, err)
	assert.Contains(t, log, "Started task")
	assert.Contains(t, log, "Implemented feature")

	feedbackFiles, err := taskFS.ListFeedback()
	require.NoError(t, err)
	assert.Len(t, feedbackFiles, 1)

	feedback, err := taskFS.ReadFeedback("001.md")
	require.NoError(t, err)
	assert.Equal(t, "Please add tests", feedback)
}
