package sow

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jmgilman/go/git"
)

// TestWorktreePath_PreservesSlashes tests that WorktreePath preserves forward slashes
// in branch names to maintain git's semantic branch grouping
func TestWorktreePath_PreservesSlashes(t *testing.T) {
	repoRoot := "/Users/test/repo"
	branch := "feat/auth"

	result := WorktreePath(repoRoot, branch)

	expected := filepath.Join(repoRoot, ".sow", "worktrees", "feat", "auth")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestWorktreePath_SimpleBranch tests path generation for simple branch names
func TestWorktreePath_SimpleBranch(t *testing.T) {
	repoRoot := "/Users/test/repo"
	branch := "main"

	result := WorktreePath(repoRoot, branch)

	expected := filepath.Join(repoRoot, ".sow", "worktrees", "main")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestWorktreePath_NestedSlashes tests path generation with multiple slashes
func TestWorktreePath_NestedSlashes(t *testing.T) {
	repoRoot := "/Users/test/repo"
	branch := "feature/epic/task"

	result := WorktreePath(repoRoot, branch)

	expected := filepath.Join(repoRoot, ".sow", "worktrees", "feature", "epic", "task")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestEnsureWorktree_CreatesWhenMissing tests that ensureWorktree creates a worktree
// when the path doesn't exist
func TestEnsureWorktree_CreatesWhenMissing(t *testing.T) {
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

	// Create a test branch
	headRef, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get HEAD: %v", err)
	}

	branchName := "test-branch"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Open repository with git wrapper
	gitRepo, err := git.Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open repo with wrapper: %v", err)
	}

	// Create context
	ctx := &Context{
		repo:     &Git{repo: gitRepo},
		repoRoot: tempDir,
	}

	// Test worktree path
	worktreePath := filepath.Join(tempDir, ".sow", "worktrees", branchName)

	// Ensure worktree directory parent exists
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		t.Fatalf("failed to create worktree parent dir: %v", err)
	}

	// Call EnsureWorktree
	err = EnsureWorktree(ctx, worktreePath, branchName)
	if err != nil {
		t.Fatalf("EnsureWorktree failed: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Errorf("worktree path does not exist after ensureWorktree: %s", worktreePath)
	}
}

// TestEnsureWorktree_SucceedsWhenExists tests that ensureWorktree is idempotent
// and returns nil when the worktree already exists
func TestEnsureWorktree_SucceedsWhenExists(t *testing.T) {
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

	// Create a test branch
	headRef, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get HEAD: %v", err)
	}

	branchName := "test-branch"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Open repository with git wrapper
	gitRepo, err := git.Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open repo with wrapper: %v", err)
	}

	// Create context
	ctx := &Context{
		repo:     &Git{repo: gitRepo},
		repoRoot: tempDir,
	}

	// Test worktree path
	worktreePath := filepath.Join(tempDir, ".sow", "worktrees", branchName)

	// Ensure worktree directory parent exists
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		t.Fatalf("failed to create worktree parent dir: %v", err)
	}

	// Create worktree first time
	err = EnsureWorktree(ctx, worktreePath, branchName)
	if err != nil {
		t.Fatalf("first EnsureWorktree failed: %v", err)
	}

	// Call EnsureWorktree again (should be idempotent)
	err = EnsureWorktree(ctx, worktreePath, branchName)
	if err != nil {
		t.Errorf("EnsureWorktree should be idempotent but returned error: %v", err)
	}
}

// TestCheckUncommittedChanges_CleanRepo tests that checkUncommittedChanges returns nil
// for a clean repository
func TestCheckUncommittedChanges_CleanRepo(t *testing.T) {
	// Create a temporary directory for test repository
	tempDir := t.TempDir()

	// Initialize a real git repository
	repo, err := gogit.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create an initial commit
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

	// Open repository with git wrapper
	gitRepo, err := git.Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open repo with wrapper: %v", err)
	}

	// Create context
	ctx := &Context{
		repo:     &Git{repo: gitRepo},
		repoRoot: tempDir,
	}

	// Check for uncommitted changes (should return nil for clean repo)
	err = CheckUncommittedChanges(ctx)
	if err != nil {
		t.Errorf("expected nil for clean repo, got error: %v", err)
	}
}

// TestCheckUncommittedChanges_UncommittedFiles tests that checkUncommittedChanges
// returns an error when there are uncommitted changes
func TestCheckUncommittedChanges_UncommittedFiles(t *testing.T) {
	// Create a temporary directory for test repository
	tempDir := t.TempDir()

	// Initialize a real git repository
	repo, err := gogit.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create an initial commit
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

	// Now create an uncommitted file
	uncommittedFile := filepath.Join(tempDir, "uncommitted.txt")
	if err := os.WriteFile(uncommittedFile, []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("failed to write uncommitted file: %v", err)
	}

	// Open repository with git wrapper
	gitRepo, err := git.Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open repo with wrapper: %v", err)
	}

	// Create context
	ctx := &Context{
		repo:     &Git{repo: gitRepo},
		repoRoot: tempDir,
	}

	// Check for uncommitted changes (should return error)
	err = CheckUncommittedChanges(ctx)
	if err == nil {
		t.Errorf("expected error for uncommitted changes, got nil")
	}

	// Verify error message is user-friendly
	if err != nil && err.Error() == "" {
		t.Errorf("error message should not be empty")
	}
}

// TestCheckUncommittedChanges_ModifiedFiles tests detection of modified tracked files
func TestCheckUncommittedChanges_ModifiedFiles(t *testing.T) {
	// Create a temporary directory for test repository
	tempDir := t.TempDir()

	// Initialize a real git repository
	repo, err := gogit.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create an initial commit
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

	// Now modify the committed file
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Open repository with git wrapper
	gitRepo, err := git.Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open repo with wrapper: %v", err)
	}

	// Create context
	ctx := &Context{
		repo:     &Git{repo: gitRepo},
		repoRoot: tempDir,
	}

	// Check for uncommitted changes (should return error)
	err = CheckUncommittedChanges(ctx)
	if err == nil {
		t.Errorf("expected error for modified files, got nil")
	}
}
