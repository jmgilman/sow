package sow

import (
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jmgilman/go/git"
)

// Git provides access to git repository operations.
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
// Returns an error if the directory is not a git repository.
func NewGit(repoRoot string) (*Git, error) {
	// Open the repository using the git package
	repo, err := git.Open(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &Git{
		repo:     repo,
		repoRoot: repoRoot,
	}, nil
}

// Repository returns the underlying git.Repository for advanced operations.
//
// This provides an escape hatch for operations not exposed by the Git type.
// Callers can use the full power of github.com/jmgilman/go/git when needed.
func (g *Git) Repository() *git.Repository {
	return g.repo
}

// RepoRoot returns the absolute path to the repository root.
func (g *Git) RepoRoot() string {
	return g.repoRoot
}

// CurrentBranch returns the name of the current git branch.
//
// Returns an error if:
//   - HEAD cannot be read
//   - Repository is in an unexpected state
//
// Returns an empty string if HEAD is in detached state.
//
// Example:
//
//	branch, err := git.CurrentBranch()
//	// branch = "main" or "feat/my-feature"
func (g *Git) CurrentBranch() (string, error) {
	branch, err := g.repo.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return branch, nil
}

// IsProtectedBranch checks if the given branch name is a protected branch.
//
// Protected branches are:
//   - main
//   - master
//
// Returns true if the branch parameter matches a protected branch name,
// false otherwise.
//
// This is used to prevent creating sow projects on protected branches
// to avoid accidental commits to the main development line.
//
// Example:
//
//	if git.IsProtectedBranch("main") {
//	    return errors.New("cannot create project on main branch")
//	}
func (g *Git) IsProtectedBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

// HasUncommittedChanges checks if the repository has uncommitted changes.
//
// Returns true if there are:
//   - Modified files (staged or unstaged)
//   - Untracked files
//   - Deleted files
//
// Returns an error if the working tree cannot be accessed.
func (g *Git) HasUncommittedChanges() (bool, error) {
	// Use underlying go-git for status (not exposed by wrapper)
	wt, err := g.repo.Underlying().Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get repository status: %w", err)
	}

	// If status is not empty, there are uncommitted changes
	return !status.IsClean(), nil
}

// Branches returns a list of all local branch names.
//
// Returns an error if branches cannot be listed.
func (g *Git) Branches() ([]string, error) {
	branches, err := g.repo.ListBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// Filter to only local branches
	var names []string
	for _, branch := range branches {
		if !branch.IsRemote {
			names = append(names, branch.Name)
		}
	}

	return names, nil
}

// CheckoutBranch checks out the specified branch.
//
// Parameters:
//   - branchName: The name of the branch to checkout (e.g., "feat/auth", "123-add-auth")
//
// Returns an error if:
//   - The branch doesn't exist
//   - The checkout operation fails
//
// Example:
//
//	err := git.CheckoutBranch("feat/auth")
func (g *Git) CheckoutBranch(branchName string) error {
	// Get the worktree
	wt, err := g.repo.Underlying().Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout the branch using the underlying go-git
	err = wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	return nil
}
