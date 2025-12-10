package git

import (
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jmgilman/go/git"
)

// Git provides git repository operations.
//
// This type wraps github.com/jmgilman/go/git.Repository with sow-specific
// conveniences and protected branch checking.
type Git struct {
	repo     *git.Repository
	repoRoot string
}

// NewGit creates a new Git instance for the repository.
//
// The repoRoot should be the absolute path to the git repository root
// (the directory containing .git/).
//
// Returns ErrNotGitRepository if the directory is not a git repository.
func NewGit(repoRoot string) (*Git, error) {
	repo, err := git.Open(repoRoot)
	if err != nil {
		return nil, ErrNotGitRepository{Path: repoRoot}
	}

	return &Git{
		repo:     repo,
		repoRoot: repoRoot,
	}, nil
}

// Repository returns the underlying git.Repository for advanced operations.
func (g *Git) Repository() *git.Repository {
	return g.repo
}

// RepoRoot returns the absolute path to the repository root.
func (g *Git) RepoRoot() string {
	return g.repoRoot
}

// CurrentBranch returns the name of the current git branch.
// Returns an empty string if HEAD is in detached state.
func (g *Git) CurrentBranch() (string, error) {
	branch, err := g.repo.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return branch, nil
}

// IsProtectedBranch checks if the given branch name is protected (main/master).
func (g *Git) IsProtectedBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

// HasUncommittedChanges checks if the repository has uncommitted changes.
// Returns true if there are modified, staged, or deleted files.
// Untracked files are NOT considered uncommitted changes.
func (g *Git) HasUncommittedChanges() (bool, error) {
	wt, err := g.repo.Underlying().Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get repository status: %w", err)
	}

	// Check each file status - only count non-untracked files as uncommitted changes
	for _, fileStatus := range status {
		// Untracked files have Worktree status of Untracked and Staging status of Untracked
		if fileStatus.Worktree == gogit.Untracked && fileStatus.Staging == gogit.Untracked {
			continue
		}
		// Any other status means there are uncommitted changes
		return true, nil
	}

	return false, nil
}

// Branches returns a list of all local branch names.
func (g *Git) Branches() ([]string, error) {
	branches, err := g.repo.ListBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var names []string
	for _, branch := range branches {
		if !branch.IsRemote {
			names = append(names, branch.Name)
		}
	}

	return names, nil
}

// CheckoutBranch checks out the specified branch.
func (g *Git) CheckoutBranch(branchName string) error {
	wt, err := g.repo.Underlying().Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	return nil
}
