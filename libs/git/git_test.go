package git

import (
	"errors"
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

// initTestRepo creates an initialized git repository in a temp directory.
// Returns the temp directory path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err, "failed to init git repo")

	// Create initial commit (required for branch operations)
	wt, err := repo.Worktree()
	require.NoError(t, err, "failed to get worktree")

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

	return tempDir
}

// TestNewGit_OpensValidRepository tests that NewGit successfully opens a valid git repository.
func TestNewGit_OpensValidRepository(t *testing.T) {
	tempDir := initTestRepo(t)

	g, err := NewGit(tempDir)

	require.NoError(t, err)
	assert.NotNil(t, g)
}

// TestNewGit_ReturnsErrNotGitRepository tests that NewGit returns ErrNotGitRepository
// for a directory that is not a git repository.
func TestNewGit_ReturnsErrNotGitRepository(t *testing.T) {
	tempDir := t.TempDir() // Empty directory, not a git repo

	g, err := NewGit(tempDir)

	require.Error(t, err)
	assert.Nil(t, g)

	var notRepoErr ErrNotGitRepository
	assert.True(t, errors.As(err, &notRepoErr), "error should be ErrNotGitRepository")
	assert.Equal(t, tempDir, notRepoErr.Path)
}

// TestGit_Repository_ReturnsNonNil tests that Repository() returns a non-nil value
// after successful open.
func TestGit_Repository_ReturnsNonNil(t *testing.T) {
	tempDir := initTestRepo(t)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	repo := g.Repository()

	assert.NotNil(t, repo)
}

// TestGit_RepoRoot_ReturnsPath tests that RepoRoot() returns the path passed to NewGit.
func TestGit_RepoRoot_ReturnsPath(t *testing.T) {
	tempDir := initTestRepo(t)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	repoRoot := g.RepoRoot()

	assert.Equal(t, tempDir, repoRoot)
}

// TestGit_CurrentBranch_ReturnsCorrectBranch tests that CurrentBranch returns
// the correct branch name.
func TestGit_CurrentBranch_ReturnsCorrectBranch(t *testing.T) {
	tempDir := initTestRepo(t)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	branch, err := g.CurrentBranch()

	require.NoError(t, err)
	// go-git defaults to "master" on init
	assert.Equal(t, "master", branch)
}

// TestGit_CurrentBranch_ReturnsFeatureBranch tests that CurrentBranch returns
// correct name when on a feature branch.
func TestGit_CurrentBranch_ReturnsFeatureBranch(t *testing.T) {
	tempDir := initTestRepo(t)

	// Create and checkout a feature branch
	repo, err := gogit.PlainOpen(tempDir)
	require.NoError(t, err)

	headRef, err := repo.Head()
	require.NoError(t, err)

	branchName := "feature/test"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	err = wt.Checkout(&gogit.CheckoutOptions{
		Branch: branchRef,
	})
	require.NoError(t, err)

	// Now test
	g, err := NewGit(tempDir)
	require.NoError(t, err)

	branch, err := g.CurrentBranch()

	require.NoError(t, err)
	assert.Equal(t, branchName, branch)
}

// TestGit_IsProtectedBranch_ReturnsTrueForMain tests that main is protected.
func TestGit_IsProtectedBranch_ReturnsTrueForMain(t *testing.T) {
	tempDir := initTestRepo(t)
	g, err := NewGit(tempDir)
	require.NoError(t, err)

	assert.True(t, g.IsProtectedBranch("main"))
}

// TestGit_IsProtectedBranch_ReturnsTrueForMaster tests that master is protected.
func TestGit_IsProtectedBranch_ReturnsTrueForMaster(t *testing.T) {
	tempDir := initTestRepo(t)
	g, err := NewGit(tempDir)
	require.NoError(t, err)

	assert.True(t, g.IsProtectedBranch("master"))
}

// TestGit_IsProtectedBranch_ReturnsFalseForOtherBranches tests that non-protected
// branches return false.
func TestGit_IsProtectedBranch_ReturnsFalseForOtherBranches(t *testing.T) {
	tempDir := initTestRepo(t)
	g, err := NewGit(tempDir)
	require.NoError(t, err)

	tests := []string{
		"develop",
		"feature/x",
		"feat/auth",
		"bugfix/123",
		"release/1.0",
	}

	for _, branch := range tests {
		t.Run(branch, func(t *testing.T) {
			assert.False(t, g.IsProtectedBranch(branch))
		})
	}
}

// TestGit_HasUncommittedChanges_ReturnsFalseForCleanRepo tests that a clean
// repository returns false.
func TestGit_HasUncommittedChanges_ReturnsFalseForCleanRepo(t *testing.T) {
	tempDir := initTestRepo(t)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	hasChanges, err := g.HasUncommittedChanges()

	require.NoError(t, err)
	assert.False(t, hasChanges)
}

// TestGit_HasUncommittedChanges_ReturnsTrueForModifiedFiles tests that
// modified tracked files return true.
func TestGit_HasUncommittedChanges_ReturnsTrueForModifiedFiles(t *testing.T) {
	tempDir := initTestRepo(t)

	// Modify the tracked file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("modified content"), 0644)
	require.NoError(t, err)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	hasChanges, err := g.HasUncommittedChanges()

	require.NoError(t, err)
	assert.True(t, hasChanges)
}

// TestGit_HasUncommittedChanges_ReturnsTrueForStagedChanges tests that
// staged changes return true.
func TestGit_HasUncommittedChanges_ReturnsTrueForStagedChanges(t *testing.T) {
	tempDir := initTestRepo(t)

	// Create a new file and stage it
	newFile := filepath.Join(tempDir, "new.txt")
	err := os.WriteFile(newFile, []byte("new content"), 0644)
	require.NoError(t, err)

	repo, err := gogit.PlainOpen(tempDir)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	_, err = wt.Add("new.txt")
	require.NoError(t, err)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	hasChanges, err := g.HasUncommittedChanges()

	require.NoError(t, err)
	assert.True(t, hasChanges)
}

// TestGit_HasUncommittedChanges_ReturnsFalseForOnlyUntrackedFiles tests that
// untracked files do NOT count as uncommitted changes.
func TestGit_HasUncommittedChanges_ReturnsFalseForOnlyUntrackedFiles(t *testing.T) {
	tempDir := initTestRepo(t)

	// Create an untracked file (don't stage it)
	untrackedFile := filepath.Join(tempDir, "untracked.txt")
	err := os.WriteFile(untrackedFile, []byte("untracked content"), 0644)
	require.NoError(t, err)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	hasChanges, err := g.HasUncommittedChanges()

	require.NoError(t, err)
	assert.False(t, hasChanges, "untracked files should NOT count as uncommitted changes")
}

// TestGit_Branches_ReturnsLocalBranches tests that Branches() returns
// a list of local branches.
func TestGit_Branches_ReturnsLocalBranches(t *testing.T) {
	tempDir := initTestRepo(t)

	// Create some additional branches
	repo, err := gogit.PlainOpen(tempDir)
	require.NoError(t, err)

	headRef, err := repo.Head()
	require.NoError(t, err)

	// Create feature/a branch
	branchA := plumbing.NewBranchReferenceName("feature/a")
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchA, headRef.Hash()))
	require.NoError(t, err)

	// Create feature/b branch
	branchB := plumbing.NewBranchReferenceName("feature/b")
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchB, headRef.Hash()))
	require.NoError(t, err)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	branches, err := g.Branches()

	require.NoError(t, err)
	assert.Contains(t, branches, "master")
	assert.Contains(t, branches, "feature/a")
	assert.Contains(t, branches, "feature/b")
}

// TestGit_Branches_DoesNotIncludeRemoteBranches tests that remote tracking
// branches are NOT included in the list.
func TestGit_Branches_DoesNotIncludeRemoteBranches(t *testing.T) {
	tempDir := initTestRepo(t)

	// Create a remote tracking branch reference directly
	repo, err := gogit.PlainOpen(tempDir)
	require.NoError(t, err)

	headRef, err := repo.Head()
	require.NoError(t, err)

	// Create a remote tracking branch reference
	remoteRef := plumbing.NewRemoteReferenceName("origin", "main")
	err = repo.Storer.SetReference(plumbing.NewHashReference(remoteRef, headRef.Hash()))
	require.NoError(t, err)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	branches, err := g.Branches()

	require.NoError(t, err)
	// Should only have local master branch
	assert.Contains(t, branches, "master")
	// Should NOT contain remote tracking branches
	assert.NotContains(t, branches, "origin/main")
}

// TestGit_CheckoutBranch_SuccessfullyCheckoutExistingBranch tests that
// CheckoutBranch successfully checks out an existing branch.
func TestGit_CheckoutBranch_SuccessfullyCheckoutExistingBranch(t *testing.T) {
	tempDir := initTestRepo(t)

	// Create a feature branch
	repo, err := gogit.PlainOpen(tempDir)
	require.NoError(t, err)

	headRef, err := repo.Head()
	require.NoError(t, err)

	branchName := "feature/checkout-test"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	require.NoError(t, err)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	// Checkout the branch
	err = g.CheckoutBranch(branchName)

	require.NoError(t, err)

	// Verify we're now on the feature branch
	currentBranch, err := g.CurrentBranch()
	require.NoError(t, err)
	assert.Equal(t, branchName, currentBranch)
}

// TestGit_CheckoutBranch_ReturnsErrorForNonExistentBranch tests that
// CheckoutBranch returns an error for a non-existent branch.
func TestGit_CheckoutBranch_ReturnsErrorForNonExistentBranch(t *testing.T) {
	tempDir := initTestRepo(t)

	g, err := NewGit(tempDir)
	require.NoError(t, err)

	err = g.CheckoutBranch("nonexistent-branch")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-branch")
}
