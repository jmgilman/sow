package sowfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmgilman/go/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
