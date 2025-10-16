package sowfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmgilman/go/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSowFSWithFS tests the testing constructor with in-memory filesystem.
func TestNewSowFSWithFS(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *billy.MemoryFS
		repoRoot  string
		wantErr   error
		checkFunc func(t *testing.T, sowFS *SowFSImpl)
	}{
		{
			name: "success - .sow directory exists",
			setup: func() *billy.MemoryFS {
				fs := billy.NewMemory()
				_ = fs.MkdirAll(".sow/knowledge", 0755)
				_ = fs.MkdirAll(".sow/refs", 0755)
				return fs
			},
			repoRoot: "/test/repo",
			wantErr:  nil,
			checkFunc: func(t *testing.T, sowFS *SowFSImpl) {
				assert.NotNil(t, sowFS)
				assert.Equal(t, "/test/repo", sowFS.RepoRoot())

				// Verify filesystem is chrooted to .sow
				// Should be able to access knowledge/ directly (not .sow/knowledge/)
				exists, err := sowFS.fs.Exists("knowledge")
				require.NoError(t, err)
				assert.True(t, exists)
			},
		},
		{
			name: "error - .sow directory missing",
			setup: func() *billy.MemoryFS {
				fs := billy.NewMemory()
				// Don't create .sow directory
				return fs
			},
			repoRoot: "/test/repo",
			wantErr:  ErrSowNotInitialized,
		},
		{
			name: "error - .sow is a file not directory",
			setup: func() *billy.MemoryFS {
				fs := billy.NewMemory()
				_ = fs.WriteFile(".sow", []byte("not a directory"), 0644)
				return fs
			},
			repoRoot: "/test/repo",
			wantErr:  ErrSowNotInitialized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setup()

			sowFS, err := NewSowFSWithFS(fs, tt.repoRoot)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, sowFS)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sowFS)
				if tt.checkFunc != nil {
					tt.checkFunc(t, sowFS)
				}
			}
		})
	}
}

// TestFindGitRepoRoot tests git repository root detection.
func TestFindGitRepoRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create git repo structure:
	// tmpDir/
	//   .git/
	//   src/
	//     subdir/
	gitDir := filepath.Join(tmpDir, ".git")
	srcDir := filepath.Join(tmpDir, "src")
	subDir := filepath.Join(srcDir, "subdir")

	require.NoError(t, os.MkdirAll(gitDir, 0755))
	require.NoError(t, os.MkdirAll(subDir, 0755))

	tests := []struct {
		name      string
		startPath string
		wantRoot  string
		wantErr   error
	}{
		{
			name:      "from repo root",
			startPath: tmpDir,
			wantRoot:  tmpDir,
			wantErr:   nil,
		},
		{
			name:      "from src directory",
			startPath: srcDir,
			wantRoot:  tmpDir,
			wantErr:   nil,
		},
		{
			name:      "from deeply nested directory",
			startPath: subDir,
			wantRoot:  tmpDir,
			wantErr:   nil,
		},
		{
			name:      "not in git repo",
			startPath: t.TempDir(), // Different temp dir without .git
			wantErr:   ErrNotInGitRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := findGitRepoRoot(tt.startPath)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantRoot, root)
			}
		})
	}
}

// TestFindGitRepoRoot_Worktree tests git worktree case where .git is a file.
func TestFindGitRepoRoot_Worktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git file (worktree case)
	gitFile := filepath.Join(tmpDir, ".git")
	err := os.WriteFile(gitFile, []byte("gitdir: /path/to/main/repo/.git/worktrees/feature"), 0644)
	require.NoError(t, err)

	// Should still detect this as repo root
	root, err := findGitRepoRoot(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, tmpDir, root)
}

// TestNewSowFSFromPath tests creating SowFS from a path.
func TestNewSowFSFromPath(t *testing.T) {
	// Create temporary git repo with .sow
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	sowDir := filepath.Join(tmpDir, ".sow")
	srcDir := filepath.Join(tmpDir, "src")

	require.NoError(t, os.MkdirAll(gitDir, 0755))
	require.NoError(t, os.MkdirAll(sowDir, 0755))
	require.NoError(t, os.MkdirAll(srcDir, 0755))

	tests := []struct {
		name      string
		path      string
		wantErr   error
		checkFunc func(t *testing.T, sowFS *SowFSImpl)
	}{
		{
			name:    "success - from repo root",
			path:    tmpDir,
			wantErr: nil,
			checkFunc: func(t *testing.T, sowFS *SowFSImpl) {
				assert.Equal(t, tmpDir, sowFS.RepoRoot())
			},
		},
		{
			name:    "success - from subdirectory",
			path:    srcDir,
			wantErr: nil,
			checkFunc: func(t *testing.T, sowFS *SowFSImpl) {
				// Should still find repo root
				assert.Equal(t, tmpDir, sowFS.RepoRoot())
			},
		},
		{
			name:    "error - not in git repo",
			path:    t.TempDir(), // Different temp dir
			wantErr: ErrNotInGitRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sowFS, err := NewSowFSFromPath(tt.path)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, sowFS)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sowFS)
				if tt.checkFunc != nil {
					tt.checkFunc(t, sowFS)
				}

				// Cleanup
				_ = sowFS.Close()
			}
		})
	}
}

// TestNewSowFSFromPath_NoSowDirectory tests error when .sow missing.
func TestNewSowFSFromPath_NoSowDirectory(t *testing.T) {
	// Create git repo WITHOUT .sow
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	sowFS, err := NewSowFSFromPath(tmpDir)

	assert.ErrorIs(t, err, ErrSowNotInitialized)
	assert.Nil(t, sowFS)
}

// TestSowFSImpl_Close tests Close method.
func TestSowFSImpl_Close(t *testing.T) {
	fs := billy.NewMemory()
	_ = fs.MkdirAll(".sow", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test")
	require.NoError(t, err)

	err = sowFS.Close()
	assert.NoError(t, err)
}
