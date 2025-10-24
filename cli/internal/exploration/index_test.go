package exploration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

// setupTestContext creates a temporary test directory and sow context.
func setupTestContext(t *testing.T) (*sow.Context, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "sow-exploration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo with go-git
	gogitRepo, err := gogit.PlainInit(tmpDir, false)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create initial commit so we have a proper git repository
	wt, err := gogitRepo.Worktree()
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Create README
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create README: %v", err)
	}

	// Stage and commit
	if _, err := wt.Add("README.md"); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to add README: %v", err)
	}

	if _, err := wt.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to commit: %v", err)
	}

	// Initialize sow structure
	if err := sow.Init(tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize sow: %v", err)
	}

	// Create context
	ctx, err := sow.NewContext(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create context: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return ctx, cleanup
}

func TestExists(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Should not exist initially
	if Exists(ctx) {
		t.Error("Exists() = true, want false before initialization")
	}

	// Create exploration
	if err := InitExploration(ctx, "test-topic", "explore/test-topic"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Should exist now
	if !Exists(ctx) {
		t.Error("Exists() = false, want true after initialization")
	}
}

func TestInitExploration(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	topic := "authentication-research"
	branch := "explore/authentication-research"

	// Initialize exploration
	if err := InitExploration(ctx, topic, branch); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Verify directory exists
	explorationDir := filepath.Join(ctx.RepoRoot(), ".sow", "exploration")
	if _, err := os.Stat(explorationDir); os.IsNotExist(err) {
		t.Error("Exploration directory was not created")
	}

	// Verify index was created
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if index.Exploration.Topic != topic {
		t.Errorf("Topic = %q, want %q", index.Exploration.Topic, topic)
	}

	if index.Exploration.Branch != branch {
		t.Errorf("Branch = %q, want %q", index.Exploration.Branch, branch)
	}

	if index.Exploration.Status != "active" {
		t.Errorf("Status = %q, want %q", index.Exploration.Status, "active")
	}

	if len(index.Files) != 0 {
		t.Errorf("Files length = %d, want 0", len(index.Files))
	}

	// Test that re-initializing fails
	if err := InitExploration(ctx, topic, branch); !errors.Is(err, ErrExplorationExists) {
		t.Errorf("InitExploration() second time = %v, want ErrExplorationExists", err)
	}
}

func TestLoadIndex(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Should fail when no exploration exists
	_, err := LoadIndex(ctx)
	if !errors.Is(err, ErrNoExploration) {
		t.Errorf("LoadIndex() with no exploration = %v, want ErrNoExploration", err)
	}

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Should succeed now
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if index == nil {
		t.Fatal("LoadIndex() returned nil index")
	}
}

func TestSaveIndex(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Load index
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	// Modify index
	index.Exploration.Status = "completed"
	index.Files = append(index.Files, schemas.ExplorationFile{
		Path:        "test.md",
		Description: "Test file",
		Tags:        []string{"test"},
		Created_at:  time.Now(),
	})

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		t.Fatalf("SaveIndex() failed: %v", err)
	}

	// Load again and verify
	reloaded, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() after save failed: %v", err)
	}

	if reloaded.Exploration.Status != "completed" {
		t.Errorf("Status after reload = %q, want %q", reloaded.Exploration.Status, "completed")
	}

	if len(reloaded.Files) != 1 {
		t.Fatalf("Files length after reload = %d, want 1", len(reloaded.Files))
	}

	if reloaded.Files[0].Path != "test.md" {
		t.Errorf("File path = %q, want %q", reloaded.Files[0].Path, "test.md")
	}
}

func TestDelete(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Verify it exists
	if !Exists(ctx) {
		t.Fatal("Exploration should exist before delete")
	}

	// Delete
	if err := Delete(ctx); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify it no longer exists
	if Exists(ctx) {
		t.Error("Exploration should not exist after delete")
	}

	// Delete again should fail
	if err := Delete(ctx); !errors.Is(err, ErrNoExploration) {
		t.Errorf("Delete() when not exists = %v, want ErrNoExploration", err)
	}
}

func TestGetFilePath(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Get file path
	path := GetFilePath(ctx, "test.md")
	expected := filepath.Join(ctx.RepoRoot(), ".sow", "exploration", "test.md")

	if path != expected {
		t.Errorf("GetFilePath() = %q, want %q", path, expected)
	}
}
