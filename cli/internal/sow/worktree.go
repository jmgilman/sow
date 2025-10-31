package sow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// WorktreePath returns the path where a worktree for the given branch should be created.
// Preserves forward slashes in branch names to maintain git's branch namespacing.
// Example: branch "feat/auth" → "<repoRoot>/.sow/worktrees/feat/auth/".
func WorktreePath(repoRoot, branch string) string {
	return filepath.Join(repoRoot, ".sow", "worktrees", branch)
}

// EnsureWorktree creates a git worktree at the specified path for the given branch.
// If the worktree already exists, returns nil (idempotent operation).
// Creates the branch if it doesn't exist.
func EnsureWorktree(ctx *Context, path, branch string) error {
	// Check if worktree already exists
	if _, err := os.Stat(path); err == nil {
		// Path exists - assume it's a valid worktree
		return nil
	}

	// Create parent directories for the worktree path
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	// Check current branch in main repo
	currentBranchCmd := exec.Command("git", "branch", "--show-current")
	currentBranchCmd.Dir = ctx.RepoRoot()
	currentBranchOutput, err := currentBranchCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := string(currentBranchOutput)
	currentBranch = currentBranch[:len(currentBranch)-1] // Remove trailing newline

	// If we're currently on the branch we want to create a worktree for,
	// switch to a different branch first (git worktree add fails if branch is checked out)
	if currentBranch == branch {
		switchCmd := exec.Command("git", "checkout", "master")
		switchCmd.Dir = ctx.RepoRoot()
		if err := switchCmd.Run(); err != nil {
			// If master doesn't exist, try main
			switchCmd = exec.Command("git", "checkout", "main")
			switchCmd.Dir = ctx.RepoRoot()
			_ = switchCmd.Run() // Ignore error - we'll fail later if needed
		}
	}

	// Check if branch exists using git CLI
	checkCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	checkCmd.Dir = ctx.RepoRoot()
	branchExists := checkCmd.Run() == nil

	// If branch doesn't exist, create it using git CLI (so both go-git and git CLI can see it)
	if !branchExists {
		// Get current HEAD
		repo := ctx.Git().Repository().Underlying()
		head, err := repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}

		// Create branch using git CLI
		createBranchCmd := exec.Command("git", "branch", branch, head.Hash().String())
		createBranchCmd.Dir = ctx.RepoRoot()
		if output, err := createBranchCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create branch via git CLI: %w (output: %s)", err, output)
		}
	}

	// Create worktree using git CLI (more reliable than go-git for worktrees)
	addCmd := exec.Command("git", "worktree", "add", path, branch)
	addCmd.Dir = ctx.RepoRoot()
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add worktree: command %v failed with exit code %d: %s", addCmd.Args, addCmd.ProcessState.ExitCode(), string(output))
	}

	return nil
}

// CheckUncommittedChanges verifies the main repository has no uncommitted changes.
// Returns an error if uncommitted changes exist.
// Can be skipped in test environments by setting SOW_SKIP_UNCOMMITTED_CHECK=1.
func CheckUncommittedChanges(ctx *Context) error {
	// Allow tests to skip this check if needed (for testscript compatibility)
	if os.Getenv("SOW_SKIP_UNCOMMITTED_CHECK") == "1" {
		return nil
	}

	worktree, err := ctx.Git().Repository().Underlying().Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	if !status.IsClean() {
		return fmt.Errorf("repository has uncommitted changes - commit or stash them before creating worktree")
	}

	return nil
}
