package modes

import (
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// CreateBranch creates a new git branch and checks it out.
// This is shared across all modes that need to create branches.
func CreateBranch(git *sow.Git, branchName string) error {
	// Use underlying go-git to create branch
	wt, err := git.Repository().Underlying().Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get current HEAD
	head, err := git.Repository().Underlying().Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create branch reference
	branchRef := "refs/heads/" + branchName
	if err := git.Repository().Underlying().Storer.SetReference(
		plumbing.NewHashReference(plumbing.ReferenceName(branchRef), head.Hash()),
	); err != nil {
		return fmt.Errorf("failed to create branch reference: %w", err)
	}

	// Checkout the new branch
	if err := wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRef),
	}); err != nil {
		return fmt.Errorf("failed to checkout new branch: %w", err)
	}

	return nil
}

// HandleBranchScenario handles the --branch flag scenario for any mode.
// Returns: (branchName, topic, shouldCreateNew, error).
func HandleBranchScenario(ctx *sow.Context, mode Mode, branchName string, existsFunc ExistsFunc) (string, string, bool, error) {
	git := ctx.Git()

	// Check if branch exists locally
	branches, err := git.Branches()
	if err != nil {
		return "", "", false, fmt.Errorf("failed to list branches: %w", err)
	}

	branchExists := branchExistsInList(branches, branchName)

	if branchExists {
		// Checkout existing branch
		if err := git.CheckoutBranch(branchName); err != nil {
			return "", "", false, fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}
	} else {
		// Create new branch
		if err := ensureAndCreateBranch(git, branchName); err != nil {
			return "", "", false, err
		}
	}

	// Extract topic from branch name
	topic := ExtractTopicFromBranch(mode, branchName)

	// Check if mode session exists in this branch
	modeExists := existsFunc(ctx)

	return branchName, topic, !modeExists, nil
}

// branchExistsInList checks if a branch name exists in the list of branches.
func branchExistsInList(branches []string, branchName string) bool {
	for _, b := range branches {
		if b == branchName {
			return true
		}
	}
	return false
}

// ensureAndCreateBranch ensures we're on a protected branch, then creates and checks out a new branch.
func ensureAndCreateBranch(git *sow.Git, branchName string) error {
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if !git.IsProtectedBranch(currentBranch) {
		return fmt.Errorf("cannot create branch %s from %s - please checkout main/master first", branchName, currentBranch)
	}

	if err := CreateBranch(git, branchName); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return nil
}

// HandleCurrentBranchScenario handles the no-flags scenario (current branch) for any mode.
// Returns: (branchName, topic, shouldCreateNew, error).
func HandleCurrentBranchScenario(ctx *sow.Context, mode Mode, existsFunc ExistsFunc) (string, string, bool, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", "", false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if mode session exists
	modeExists := existsFunc(ctx)

	if !modeExists {
		// Validate we're not on a protected branch before creating new
		if git.IsProtectedBranch(currentBranch) {
			return "", "", false, fmt.Errorf("cannot create %s on protected branch '%s' - create a branch first", mode.Name(), currentBranch)
		}
	}

	// Extract topic from branch name
	topic := ExtractTopicFromBranch(mode, currentBranch)

	return currentBranch, topic, !modeExists, nil
}
