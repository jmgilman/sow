package sow

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// setupTestRepo creates a test git repository with an initial commit.
func setupTestRepo(t *testing.T) (string, *gogit.Repository) {
	t.Helper()

	// Create a temporary directory for test repository
	tempDir := t.TempDir()

	// Initialize a real git repository
	repo, err := gogit.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create an initial commit (required for worktrees)
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create a test file and commit it
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = wt.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	return tempDir, repo
}

// createWorktree creates a worktree for testing by simulating git's worktree structure.
func createWorktree(t *testing.T, repo *gogit.Repository, mainRepoPath, branchName string) string {
	t.Helper()

	// Create a test branch
	headRef, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get HEAD: %v", err)
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Create worktree path - use simpler path without .sow prefix for testing
	worktreePath := t.TempDir()

	// Create the worktree metadata directory in main repo
	worktreeMetaPath := filepath.Join(mainRepoPath, ".git", "worktrees", branchName)
	if err := os.MkdirAll(worktreeMetaPath, 0755); err != nil {
		t.Fatalf("failed to create worktree metadata dir: %v", err)
	}

	// Create .git file in worktree pointing to main repo's worktree metadata
	gitFilePath := filepath.Join(worktreePath, ".git")
	gitFileContent := "gitdir: " + worktreeMetaPath + "\n"
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	return worktreePath
}

// TestNewContext_MainRepo tests that NewContext correctly detects a main repository
// and sets worktree fields appropriately.
func TestNewContext_MainRepo(t *testing.T) {
	// Set up a test repository (main repo, not worktree)
	testRepoPath, _ := setupTestRepo(t)

	// Create Context
	ctx, err := NewContext(testRepoPath)

	// Verify no errors
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's not detected as a worktree
	if ctx.IsWorktree() {
		t.Error("expected main repo, got worktree")
	}

	// Verify MainRepoRoot returns the repo path itself
	if ctx.MainRepoRoot() != testRepoPath {
		t.Errorf("expected main repo root %s, got %s", testRepoPath, ctx.MainRepoRoot())
	}
}

// TestNewContext_Worktree tests that NewContext correctly detects a worktree environment
// and extracts the main repository path.
func TestNewContext_Worktree(t *testing.T) {
	// Set up test main repo
	mainRepoPath, repo := setupTestRepo(t)

	// Create worktree using git commands
	branchName := "test-branch"
	worktreePath := createWorktree(t, repo, mainRepoPath, branchName)

	// Create Context from worktree path
	ctx, err := NewContext(worktreePath)

	// Verify no errors
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's detected as a worktree
	if !ctx.IsWorktree() {
		t.Error("expected worktree, got main repo")
	}

	// Verify MainRepoRoot returns the main repo path (not the worktree path)
	if ctx.MainRepoRoot() != mainRepoPath {
		t.Errorf("expected main repo root %s, got %s", mainRepoPath, ctx.MainRepoRoot())
	}
}

// TestContext_IsWorktree tests the IsWorktree accessor method.
func TestContext_IsWorktree(t *testing.T) {
	// Test main repo scenario
	t.Run("MainRepo", func(t *testing.T) {
		testRepoPath, _ := setupTestRepo(t)
		ctx, err := NewContext(testRepoPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.IsWorktree() {
			t.Error("IsWorktree() should return false for main repo")
		}
	})

	// Test worktree scenario
	t.Run("Worktree", func(t *testing.T) {
		mainRepoPath, repo := setupTestRepo(t)
		worktreePath := createWorktree(t, repo, mainRepoPath, "test-branch")

		ctx, err := NewContext(worktreePath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ctx.IsWorktree() {
			t.Error("IsWorktree() should return true for worktree")
		}
	})
}

// TestContext_MainRepoRoot tests the MainRepoRoot accessor method.
func TestContext_MainRepoRoot(t *testing.T) {
	// Test main repo scenario
	t.Run("MainRepo", func(t *testing.T) {
		testRepoPath, _ := setupTestRepo(t)
		ctx, err := NewContext(testRepoPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.MainRepoRoot() != testRepoPath {
			t.Errorf("expected MainRepoRoot() to return %s, got %s", testRepoPath, ctx.MainRepoRoot())
		}
	})

	// Test worktree scenario
	t.Run("Worktree", func(t *testing.T) {
		mainRepoPath, repo := setupTestRepo(t)
		worktreePath := createWorktree(t, repo, mainRepoPath, "test-branch")

		ctx, err := NewContext(worktreePath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.MainRepoRoot() != mainRepoPath {
			t.Errorf("expected MainRepoRoot() to return %s, got %s", mainRepoPath, ctx.MainRepoRoot())
		}
	})
}

// TestNewContext_WorktreeWithSlashesInBranch tests worktree detection with branch names
// containing slashes (e.g., feat/auth).
func TestNewContext_WorktreeWithSlashesInBranch(t *testing.T) {
	// Set up test main repo
	mainRepoPath, repo := setupTestRepo(t)

	// Create worktree with slash in branch name
	branchName := "feat/auth"
	worktreePath := createWorktree(t, repo, mainRepoPath, branchName)

	// Create Context from worktree path
	ctx, err := NewContext(worktreePath)

	// Verify no errors
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's detected as a worktree
	if !ctx.IsWorktree() {
		t.Error("expected worktree, got main repo")
	}

	// Verify MainRepoRoot correctly extracts the main repo path
	if ctx.MainRepoRoot() != mainRepoPath {
		t.Errorf("expected main repo root %s, got %s", mainRepoPath, ctx.MainRepoRoot())
	}
}

// TestContext_SinksPath_MainRepo tests that SinksPath returns the local path
// when running in the main repository (not a worktree).
func TestContext_SinksPath_MainRepo(t *testing.T) {
	// Set up test repo (main, not worktree)
	testRepoPath, _ := setupTestRepo(t)

	// Create Context
	ctx, err := NewContext(testRepoPath)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Verify SinksPath returns local .sow/sinks path
	expected := filepath.Join(testRepoPath, ".sow", "sinks")
	if ctx.SinksPath() != expected {
		t.Errorf("expected %s, got %s", expected, ctx.SinksPath())
	}
}

// TestContext_SinksPath_Worktree tests that SinksPath returns the main repo path
// when running in a worktree.
func TestContext_SinksPath_Worktree(t *testing.T) {
	// Set up test main repo
	mainRepoPath, repo := setupTestRepo(t)

	// Create worktree
	branchName := "test-branch"
	worktreePath := createWorktree(t, repo, mainRepoPath, branchName)

	// Create context from worktree path
	ctx, err := NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Should return main repo path, not worktree path
	expected := filepath.Join(mainRepoPath, ".sow", "sinks")
	if ctx.SinksPath() != expected {
		t.Errorf("expected %s, got %s", expected, ctx.SinksPath())
	}
}

// TestContext_ReposPath_MainRepo tests that ReposPath returns the local path
// when running in the main repository (not a worktree).
func TestContext_ReposPath_MainRepo(t *testing.T) {
	// Set up test repo (main, not worktree)
	testRepoPath, _ := setupTestRepo(t)

	// Create Context
	ctx, err := NewContext(testRepoPath)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Verify ReposPath returns local .sow/repos path
	expected := filepath.Join(testRepoPath, ".sow", "repos")
	if ctx.ReposPath() != expected {
		t.Errorf("expected %s, got %s", expected, ctx.ReposPath())
	}
}

// TestContext_ReposPath_Worktree tests that ReposPath returns the main repo path
// when running in a worktree.
func TestContext_ReposPath_Worktree(t *testing.T) {
	// Set up test main repo
	mainRepoPath, repo := setupTestRepo(t)

	// Create worktree
	branchName := "test-branch"
	worktreePath := createWorktree(t, repo, mainRepoPath, branchName)

	// Create context from worktree path
	ctx, err := NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Should return main repo path, not worktree path
	expected := filepath.Join(mainRepoPath, ".sow", "repos")
	if ctx.ReposPath() != expected {
		t.Errorf("expected %s, got %s", expected, ctx.ReposPath())
	}
}
