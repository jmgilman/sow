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

// TestGit_IsProtectedBranch tests branch protection detection.
func TestGit_IsProtectedBranch(t *testing.T) {
	tempDir := initTestRepo(t)
	g, err := NewGit(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name   string
		branch string
		want   bool
	}{
		{name: "main is protected", branch: "main", want: true},
		{name: "master is protected", branch: "master", want: true},
		{name: "develop is not protected", branch: "develop", want: false},
		{name: "feature branch is not protected", branch: "feature/x", want: false},
		{name: "feat branch is not protected", branch: "feat/auth", want: false},
		{name: "bugfix branch is not protected", branch: "bugfix/123", want: false},
		{name: "release branch is not protected", branch: "release/1.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.IsProtectedBranch(tt.branch)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGit_HasUncommittedChanges tests detection of uncommitted changes.
func TestGit_HasUncommittedChanges(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, tempDir string)
		want  bool
	}{
		{
			name:  "clean repo returns false",
			setup: func(_ *testing.T, _ string) {},
			want:  false,
		},
		{
			name: "modified tracked file returns true",
			setup: func(t *testing.T, tempDir string) {
				t.Helper()
				testFile := filepath.Join(tempDir, "test.txt")
				err := os.WriteFile(testFile, []byte("modified content"), 0644)
				require.NoError(t, err)
			},
			want: true,
		},
		{
			name: "staged changes return true",
			setup: func(t *testing.T, tempDir string) {
				t.Helper()
				newFile := filepath.Join(tempDir, "new.txt")
				err := os.WriteFile(newFile, []byte("new content"), 0644)
				require.NoError(t, err)

				repo, err := gogit.PlainOpen(tempDir)
				require.NoError(t, err)

				wt, err := repo.Worktree()
				require.NoError(t, err)

				_, err = wt.Add("new.txt")
				require.NoError(t, err)
			},
			want: true,
		},
		{
			name: "untracked files only returns false",
			setup: func(t *testing.T, tempDir string) {
				t.Helper()
				untrackedFile := filepath.Join(tempDir, "untracked.txt")
				err := os.WriteFile(untrackedFile, []byte("untracked content"), 0644)
				require.NoError(t, err)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := initTestRepo(t)
			tt.setup(t, tempDir)

			g, err := NewGit(tempDir)
			require.NoError(t, err)

			got, err := g.HasUncommittedChanges()

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
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
