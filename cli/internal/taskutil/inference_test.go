package taskutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestResolveTaskID_ProvidedID tests resolution when user provides an ID explicitly.
func TestResolveTaskID_ProvidedID(t *testing.T) {
	// Setup doesn't matter much since we're providing an ID
	tmpDir := t.TempDir()
	createMinimalSowStructure(t, tmpDir)

	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)

	sowFS, err := sowfs.NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Test with provided ID
	taskID, err := ResolveTaskID(sowFS, "010")
	require.NoError(t, err)
	assert.Equal(t, "010", taskID)

	// Test with another provided ID
	taskID, err = ResolveTaskID(sowFS, "020")
	require.NoError(t, err)
	assert.Equal(t, "020", taskID)
}

// TestResolveTaskID_InferredFromDirectory tests inference from current directory.
func TestResolveTaskID_InferredFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create task directory
	taskDir := filepath.Join(tmpDir, ".sow/project/phases/implementation/tasks/010")
	require.NoError(t, os.MkdirAll(taskDir, 0755))

	// Create chrooted filesystem and project state
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTask(t, chrootFS, "010", "pending")
	createTaskState(t, chrootFS, "010")

	// Change to task directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(taskDir))

	sowFS, err := sowfs.NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Resolve with empty string - should infer from directory
	taskID, err := ResolveTaskID(sowFS, "")
	require.NoError(t, err)
	assert.Equal(t, "010", taskID)
}

// TestResolveTaskID_InferredFromActiveTask tests inference from active task in state.
func TestResolveTaskID_InferredFromActiveTask(t *testing.T) {
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

	sowFS, err := sowfs.NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Resolve with empty string - should infer from active task
	taskID, err := ResolveTaskID(sowFS, "")
	require.NoError(t, err)
	assert.Equal(t, "010", taskID)
}

// TestResolveTaskID_InferenceFails tests error handling when inference fails.
func TestResolveTaskID_InferenceFails(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create chrooted filesystem with project but no in_progress tasks
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTasks(t, chrootFS, []string{"010"}, []string{"pending"})

	// Change to repo root
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(tmpDir))

	sowFS, err := sowfs.NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Resolve with empty string - should fail
	_, err = ResolveTaskID(sowFS, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to infer task ID")
	assert.Contains(t, err.Error(), "no task currently in progress")
}

// TestResolveTaskIDFromArgs_WithArgs tests the args wrapper with provided ID.
func TestResolveTaskIDFromArgs_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	createMinimalSowStructure(t, tmpDir)

	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)

	sowFS, err := sowfs.NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Test with args containing task ID
	args := []string{"020"}
	taskID, err := ResolveTaskIDFromArgs(sowFS, args)
	require.NoError(t, err)
	assert.Equal(t, "020", taskID)
}

// TestResolveTaskIDFromArgs_WithoutArgs tests the args wrapper with inference.
func TestResolveTaskIDFromArgs_WithoutArgs(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	createMinimalSowStructure(t, tmpDir)

	// Create chrooted filesystem with active task
	localFS := billy.NewLocal()
	chrootFS, err := localFS.Chroot(tmpDir)
	require.NoError(t, err)
	createProjectWithTask(t, chrootFS, "030", "in_progress")

	// Change to repo root
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	require.NoError(t, os.Chdir(tmpDir))

	sowFS, err := sowfs.NewSowFSWithFS(chrootFS, tmpDir)
	require.NoError(t, err)
	defer func() { _ = sowFS.Close() }()

	// Test with empty args - should infer
	args := []string{}
	taskID, err := ResolveTaskIDFromArgs(sowFS, args)
	require.NoError(t, err)
	assert.Equal(t, "030", taskID)
}

// Helper functions

func createMinimalSowStructure(t *testing.T, tmpDir string) {
	t.Helper()
	sowDir := filepath.Join(tmpDir, ".sow")
	require.NoError(t, os.MkdirAll(filepath.Join(sowDir, "knowledge"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sowDir, "refs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sowDir, ".version"), []byte("1.0.0"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sowDir, "refs/index.json"), []byte(`{"version":"1.0.0","refs":[]}`), 0644))
}

func createProjectWithTask(t *testing.T, fs core.FS, taskID string, status string) {
	t.Helper()
	createProjectWithTasks(t, fs, []string{taskID}, []string{status})
}

func createProjectWithTasks(t *testing.T, fs core.FS, taskIDs []string, statuses []string) {
	t.Helper()
	require.Equal(t, len(taskIDs), len(statuses), "taskIDs and statuses must have same length")

	require.NoError(t, fs.MkdirAll(".sow/project", 0755))

	now := time.Now()
	state := &schemas.ProjectState{}

	state.Project.Name = "test-project"
	state.Project.Branch = "feat/test"
	state.Project.Description = "Test project"
	state.Project.Created_at = now
	state.Project.Updated_at = now

	state.Phases.Discovery.Status = "pending"
	state.Phases.Discovery.Created_at = now
	state.Phases.Discovery.Enabled = false
	state.Phases.Discovery.Artifacts = []schemas.Artifact{}

	state.Phases.Design.Status = "pending"
	state.Phases.Design.Created_at = now
	state.Phases.Design.Enabled = false
	state.Phases.Design.Artifacts = []schemas.Artifact{}

	state.Phases.Implementation.Status = "in_progress"
	state.Phases.Implementation.Created_at = now
	state.Phases.Implementation.Started_at = now.Format(time.RFC3339)
	state.Phases.Implementation.Enabled = true
	state.Phases.Implementation.Tasks = make([]schemas.Task, len(taskIDs))
	for i, id := range taskIDs {
		state.Phases.Implementation.Tasks[i] = schemas.Task{
			Id:       id,
			Name:     "Task " + id,
			Status:   statuses[i],
			Parallel: false,
		}
	}

	state.Phases.Review.Status = "pending"
	state.Phases.Review.Created_at = now
	state.Phases.Review.Enabled = true
	state.Phases.Review.Iteration = 1
	state.Phases.Review.Reports = []schemas.ReviewReport{}

	state.Phases.Finalize.Status = "pending"
	state.Phases.Finalize.Created_at = now
	state.Phases.Finalize.Enabled = true
	state.Phases.Finalize.Project_deleted = false

	data, err := yaml.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, fs.WriteFile(".sow/project/state.yaml", data, 0644))
}

func createTaskState(t *testing.T, fs core.FS, taskID string) {
	t.Helper()

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

	data, err := yaml.Marshal(state)
	require.NoError(t, err)

	taskStatePath := ".sow/project/phases/implementation/tasks/" + taskID + "/state.yaml"
	require.NoError(t, fs.WriteFile(taskStatePath, data, 0644))
}
