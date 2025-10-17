package sowfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestContextFS_Detect_TaskContext tests context detection from within a task directory.
func TestContextFS_Detect_TaskContext(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()

	// Resolve tmpDir to canonical path (handles macOS /private symlinks)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	taskDir := filepath.Join(tmpDir, ".sow/project/phases/implementation/tasks/010")
	require.NoError(t, os.MkdirAll(taskDir, 0755))

	// Create minimal .sow structure for SowFS
	createMinimalSowStructure(t, tmpDir)

	// Change to task directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(taskDir))

	// Create SowFS - chroot local filesystem to tmpDir first
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Detect context
	ctx, err := sowFS.Context().Detect()
	require.NoError(t, err)
	assert.Equal(t, ContextTask, ctx.Type)
	assert.Equal(t, "010", ctx.TaskID)
}

// TestContextFS_Detect_ProjectContext tests context detection from project root.
func TestContextFS_Detect_ProjectContext(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()

	// Resolve tmpDir to canonical path (handles macOS /private symlinks)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Change to repo root
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(tmpDir))

	// Create SowFS - chroot local filesystem to tmpDir first
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Detect context
	ctx, err := sowFS.Context().Detect()
	require.NoError(t, err)
	assert.Equal(t, ContextProject, ctx.Type)
	assert.Equal(t, "", ctx.TaskID)
}

// TestContextFS_Detect_MultipleTaskIDs tests various valid gap-numbered task IDs.
func TestContextFS_Detect_MultipleTaskIDs(t *testing.T) {
	validIDs := []string{"010", "020", "100", "1000", "030"}

	for _, taskID := range validIDs {
		t.Run(taskID, func(t *testing.T) {
			// Create temporary test structure
			tmpDir := t.TempDir()

			// Resolve tmpDir to canonical path (handles macOS /private symlinks)
			tmpDir, err := filepath.EvalSymlinks(tmpDir)
			require.NoError(t, err)

			taskDir := filepath.Join(tmpDir, ".sow/project/phases/implementation/tasks", taskID)
			require.NoError(t, os.MkdirAll(taskDir, 0755))
			createMinimalSowStructure(t, tmpDir)

			// Change to task directory
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalWd) }()
			require.NoError(t, os.Chdir(taskDir))

			// Create SowFS - chroot local filesystem to tmpDir first
			localFS := billy.NewLocal()
			chrootFS, err := localFS.Chroot(tmpDir)
			require.NoError(t, err)
			sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
			require.NoError(t, err)
			defer func() { _ = sowFS.Close() }()

			// Detect context
			ctx, err := sowFS.Context().Detect()
			require.NoError(t, err)
			assert.Equal(t, ContextTask, ctx.Type)
			assert.Equal(t, taskID, ctx.TaskID)
		})
	}
}

// TestContextFS_Detect_TaskSubdirectory tests detection from within a task subdirectory.
func TestContextFS_Detect_TaskSubdirectory(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()

	// Resolve tmpDir to canonical path (handles macOS /private symlinks)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	feedbackDir := filepath.Join(tmpDir, ".sow/project/phases/implementation/tasks/020/feedback")
	require.NoError(t, os.MkdirAll(feedbackDir, 0755))
	createMinimalSowStructure(t, tmpDir)

	// Change to feedback subdirectory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(feedbackDir))

	// Create SowFS - chroot local filesystem to tmpDir first
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Detect context
	ctx, err := sowFS.Context().Detect()
	require.NoError(t, err)
	assert.Equal(t, ContextTask, ctx.Type)
	assert.Equal(t, "020", ctx.TaskID)
}

// TestContextType_String tests the String() method.
func TestContextType_String(t *testing.T) {
	tests := []struct {
		ct   ContextType
		want string
	}{
		{ContextNone, "none"},
		{ContextProject, "project"},
		{ContextTask, "task"},
		{ContextType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ct.String())
		})
	}
}

// createMinimalSowStructure creates the minimum .sow structure for tests.
func createMinimalSowStructure(t *testing.T, tmpDir string) {
	t.Helper()

	sowDir := filepath.Join(tmpDir, ".sow")
	require.NoError(t, os.MkdirAll(filepath.Join(sowDir, "knowledge"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sowDir, "refs"), 0755))

	// Create version file
	require.NoError(t, os.WriteFile(filepath.Join(sowDir, ".version"), []byte("1.0.0"), 0644))

	// Create refs index
	require.NoError(t, os.WriteFile(filepath.Join(sowDir, "refs/index.json"), []byte(`{"version":"1.0.0","refs":[]}`), 0644))
}

// TestContextFS_InferTaskID_FromDirectory tests inference when in a task directory.
func TestContextFS_InferTaskID_FromDirectory(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	// Create minimal .sow structure
	createMinimalSowStructure(t, tmpDir)

	// Create task directory on real filesystem for directory detection
	taskDir := filepath.Join(tmpDir, ".sow/project/phases/implementation/tasks/010")
	require.NoError(t, os.MkdirAll(taskDir, 0755))

	// Create chrooted filesystem and project state
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTasks(t, chrootFS, []string{"010"}, []string{"pending"})

	// Create task state file in the real filesystem directory (for validation)
	createTaskState(t, chrootFS, "010")

	// Change to task directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(taskDir))

	// Create SowFS with the already created chrootFS
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Infer task ID
	taskID, err := sowFS.Context().InferTaskID()
	require.NoError(t, err)
	assert.Equal(t, "010", taskID)
}

// TestContextFS_InferTaskID_FromActiveTask tests inference from active task in state.
func TestContextFS_InferTaskID_FromActiveTask(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create chrooted filesystem and project state
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTasks(t, chrootFS, []string{"010", "020"}, []string{"in_progress", "pending"})

	// Change to repo root (not in task directory)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(tmpDir))
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Infer task ID - should find the in_progress task
	taskID, err := sowFS.Context().InferTaskID()
	require.NoError(t, err)
	assert.Equal(t, "010", taskID)
}

// TestContextFS_InferTaskID_NoActiveTask tests error when no task is in progress.
func TestContextFS_InferTaskID_NoActiveTask(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create chrooted filesystem and project state
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTasks(t, chrootFS, []string{"010", "020"}, []string{"pending", "pending"})

	// Change to repo root
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(tmpDir))
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Infer task ID - should fail
	_, err = sowFS.Context().InferTaskID()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no task currently in progress")
}

// TestContextFS_InferTaskID_MultipleActiveTasks tests error when multiple tasks are in progress.
func TestContextFS_InferTaskID_MultipleActiveTasks(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create chrooted filesystem and project state
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTasks(t, chrootFS, []string{"010", "020", "030"}, []string{"in_progress", "in_progress", "pending"})

	// Change to repo root
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(tmpDir))
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Infer task ID - should fail
	_, err = sowFS.Context().InferTaskID()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple tasks in progress")
	assert.Contains(t, err.Error(), "010")
	assert.Contains(t, err.Error(), "020")
}

// TestContextFS_InferTaskID_DirectoryTakesPrecedence tests that directory detection takes precedence.
func TestContextFS_InferTaskID_DirectoryTakesPrecedence(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create task directory on real filesystem for directory detection
	task010Dir := filepath.Join(tmpDir, ".sow/project/phases/implementation/tasks/010")
	require.NoError(t, os.MkdirAll(task010Dir, 0755))

	// Create chrooted filesystem and project state
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTasks(t, chrootFS, []string{"010", "020"}, []string{"pending", "in_progress"})

	// Create task state file in the real filesystem directory (for validation)
	createTaskState(t, chrootFS, "010")

	// Change to task 010 directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(task010Dir))
	sowFS, err := NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Infer task ID - should return 010 (directory takes precedence)
	taskID, err := sowFS.Context().InferTaskID()
	require.NoError(t, err)
	assert.Equal(t, "010", taskID)
}

// createProjectWithTasks creates a project with specified tasks and statuses.
// fs parameter should be the chrooted billy filesystem that will be used for SowFS.
func createProjectWithTasks(t *testing.T, fs core.FS, taskIDs []string, statuses []string) {
	t.Helper()
	require.Equal(t, len(taskIDs), len(statuses), "taskIDs and statuses must have same length")

	// Create project directory in the billy filesystem
	require.NoError(t, fs.MkdirAll(".sow/project", 0755))

	// Create ProjectState struct (following pattern from project_test.go)
	now := time.Now()
	state := &schemas.ProjectState{}

	// Project metadata
	state.Project.Name = "test-project"
	state.Project.Branch = "feat/test"
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
	state.Phases.Implementation.Pending_task_additions = nil

	// Build tasks array from parameters
	state.Phases.Implementation.Tasks = make([]schemas.Task, len(taskIDs))
	for i, id := range taskIDs {
		state.Phases.Implementation.Tasks[i] = schemas.Task{
			Id:           id,
			Name:         "Task " + id,
			Status:       statuses[i],
			Parallel:     false,
			Dependencies: nil,
		}
	}

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

	// Marshal to YAML
	data, err := yaml.Marshal(state)
	require.NoError(t, err)

	// Write to the billy filesystem
	require.NoError(t, fs.WriteFile(".sow/project/state.yaml", data, 0644))
}

// createTaskState creates a minimal task state file for the given task ID.
func createTaskState(t *testing.T, fs core.FS, taskID string) {
	t.Helper()

	// Create task state struct
	now := time.Now()
	state := &schemas.TaskState{}

	state.Task.Id = taskID
	state.Task.Name = "Task " + taskID
	state.Task.Phase = "implementation"
	state.Task.Status = "pending"
	state.Task.Created_at = now
	state.Task.Updated_at = now
	state.Task.Iteration = 1
	state.Task.Assigned_agent = "implementer"
	state.Task.References = []string{}
	state.Task.Feedback = []schemas.Feedback{}
	state.Task.Files_modified = []string{}

	// Marshal to YAML
	data, err := yaml.Marshal(state)
	require.NoError(t, err)

	// Write to the billy filesystem
	taskStatePath := ".sow/project/phases/implementation/tasks/" + taskID + "/state.yaml"
	require.NoError(t, fs.WriteFile(taskStatePath, data, 0644))
}
