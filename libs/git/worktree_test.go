package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorktreePath tests path generation for various branch name formats.
func TestWorktreePath(t *testing.T) {
	repoRoot := "/Users/test/repo"

	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{
			name:   "simple branch name",
			branch: "main",
			want:   filepath.Join(repoRoot, ".sow", "worktrees", "main"),
		},
		{
			name:   "branch with single slash creates nested path",
			branch: "feat/auth",
			want:   filepath.Join(repoRoot, ".sow", "worktrees", "feat", "auth"),
		},
		{
			name:   "branch with multiple slashes creates deeply nested path",
			branch: "feature/epic/task",
			want:   filepath.Join(repoRoot, ".sow", "worktrees", "feature", "epic", "task"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WorktreePath(repoRoot, tt.branch)
			assert.Equal(t, tt.want, got)
		})
	}
}

// setupTestRepo creates a git repo in a temp directory with an initial commit.
// Returns the repo, tempDir, and a cleanup function.
func setupTestRepo(t *testing.T) (*gogit.Repository, string) {
	t.Helper()

	tempDir := t.TempDir()

	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err, "failed to init git repo")

	wt, err := repo.Worktree()
	require.NoError(t, err, "failed to get worktree")

	// Create a test file and commit it
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err, "failed to write test file")

	_, err = wt.Add("test.txt")
	require.NoError(t, err, "failed to add file")

	_, err = wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err, "failed to commit")

	return repo, tempDir
}

// TestEnsureWorktree_CreatesWhenMissing tests that EnsureWorktree creates a worktree
// when the path doesn't exist.
func TestEnsureWorktree_CreatesWhenMissing(t *testing.T) {
	repo, tempDir := setupTestRepo(t)

	// Create a test branch
	headRef, err := repo.Head()
	require.NoError(t, err, "failed to get HEAD")

	branchName := "test-branch"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	require.NoError(t, err, "failed to create branch")

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Test worktree path
	worktreePath := filepath.Join(tempDir, ".sow", "worktrees", branchName)

	// Call EnsureWorktree
	err = EnsureWorktree(g, tempDir, worktreePath, branchName)
	require.NoError(t, err, "EnsureWorktree failed")

	// Verify worktree was created
	_, err = os.Stat(worktreePath)
	assert.NoError(t, err, "worktree path should exist after EnsureWorktree")
}

// TestEnsureWorktree_SucceedsWhenExists tests that EnsureWorktree is idempotent
// and returns nil when the worktree already exists.
func TestEnsureWorktree_SucceedsWhenExists(t *testing.T) {
	repo, tempDir := setupTestRepo(t)

	// Create a test branch
	headRef, err := repo.Head()
	require.NoError(t, err, "failed to get HEAD")

	branchName := "test-branch"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	require.NoError(t, err, "failed to create branch")

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Test worktree path
	worktreePath := filepath.Join(tempDir, ".sow", "worktrees", branchName)

	// Create worktree first time
	err = EnsureWorktree(g, tempDir, worktreePath, branchName)
	require.NoError(t, err, "first EnsureWorktree failed")

	// Call EnsureWorktree again (should be idempotent)
	err = EnsureWorktree(g, tempDir, worktreePath, branchName)
	assert.NoError(t, err, "EnsureWorktree should be idempotent")
}

// TestEnsureWorktree_CreatesBranchIfMissing tests that EnsureWorktree creates
// the branch if it doesn't exist.
func TestEnsureWorktree_CreatesBranchIfMissing(t *testing.T) {
	_, tempDir := setupTestRepo(t)

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Note: We're NOT creating the branch ahead of time
	branchName := "new-feature-branch"
	worktreePath := filepath.Join(tempDir, ".sow", "worktrees", branchName)

	// Call EnsureWorktree - it should create the branch
	err = EnsureWorktree(g, tempDir, worktreePath, branchName)
	require.NoError(t, err, "EnsureWorktree failed")

	// Verify worktree was created
	_, err = os.Stat(worktreePath)
	assert.NoError(t, err, "worktree path should exist")
}

// TestCheckUncommittedChanges_CleanRepo tests that CheckUncommittedChanges returns nil
// for a clean repository.
func TestCheckUncommittedChanges_CleanRepo(t *testing.T) {
	_, tempDir := setupTestRepo(t)

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Check for uncommitted changes (should return nil for clean repo)
	err = CheckUncommittedChanges(g)
	assert.NoError(t, err, "expected nil for clean repo")
}

// TestCheckUncommittedChanges_UntrackedFiles tests that CheckUncommittedChanges
// allows untracked files (matches git status behavior).
func TestCheckUncommittedChanges_UntrackedFiles(t *testing.T) {
	_, tempDir := setupTestRepo(t)

	// Create an untracked file (should be allowed - doesn't block worktree creation)
	untrackedFile := filepath.Join(tempDir, "untracked.txt")
	err := os.WriteFile(untrackedFile, []byte("untracked"), 0644)
	require.NoError(t, err, "failed to write untracked file")

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Check for uncommitted changes (should pass - untracked files are allowed)
	err = CheckUncommittedChanges(g)
	assert.NoError(t, err, "untracked files should not block worktree creation")
}

// TestCheckUncommittedChanges_ModifiedFiles tests detection of modified tracked files.
func TestCheckUncommittedChanges_ModifiedFiles(t *testing.T) {
	_, tempDir := setupTestRepo(t)

	// Modify the committed file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("modified"), 0644)
	require.NoError(t, err, "failed to modify file")

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Check for uncommitted changes (should return error)
	err = CheckUncommittedChanges(g)
	assert.Error(t, err, "expected error for modified files")
	assert.Contains(t, err.Error(), "uncommitted changes")
}

// TestCheckUncommittedChanges_StagedChanges tests detection of staged changes.
func TestCheckUncommittedChanges_StagedChanges(t *testing.T) {
	repo, tempDir := setupTestRepo(t)

	// Modify and stage a file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("staged changes"), 0644)
	require.NoError(t, err, "failed to modify file")

	wt, err := repo.Worktree()
	require.NoError(t, err, "failed to get worktree")

	_, err = wt.Add("test.txt")
	require.NoError(t, err, "failed to stage file")

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Check for uncommitted changes (should return error)
	err = CheckUncommittedChanges(g)
	assert.Error(t, err, "expected error for staged changes")
	assert.Contains(t, err.Error(), "uncommitted changes")
}

// TestCheckUncommittedChanges_SkipEnvVar tests that SOW_SKIP_UNCOMMITTED_CHECK=1
// skips the check even with uncommitted changes.
func TestCheckUncommittedChanges_SkipEnvVar(t *testing.T) {
	_, tempDir := setupTestRepo(t)

	// Modify the committed file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("modified"), 0644)
	require.NoError(t, err, "failed to modify file")

	// Set the environment variable
	t.Setenv("SOW_SKIP_UNCOMMITTED_CHECK", "1")

	// Create Git instance
	g, err := NewGit(tempDir)
	require.NoError(t, err, "failed to create Git instance")

	// Check for uncommitted changes (should return nil because env var is set)
	err = CheckUncommittedChanges(g)
	assert.NoError(t, err, "should skip check when SOW_SKIP_UNCOMMITTED_CHECK=1")
}
