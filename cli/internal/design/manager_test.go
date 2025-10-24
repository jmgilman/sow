package design

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// setupTestContext creates a temporary test directory and sow context.
func setupTestContext(t *testing.T) (*sow.Context, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "sow-design-test-*")
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

func TestInitDesign(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Initialize design
	if err := InitDesign(ctx, "test-topic", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// Verify design exists
	if !Exists(ctx) {
		t.Error("Exists() = false, want true after initialization")
	}

	// Verify index can be loaded
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if index.Design.Topic != "test-topic" {
		t.Errorf("Topic = %q, want %q", index.Design.Topic, "test-topic")
	}

	if index.Design.Branch != "design/test" {
		t.Errorf("Branch = %q, want %q", index.Design.Branch, "design/test")
	}

	if index.Design.Status != "active" {
		t.Errorf("Status = %q, want %q", index.Design.Status, "active")
	}

	if len(index.Inputs) != 0 {
		t.Errorf("Inputs length = %d, want 0", len(index.Inputs))
	}

	if len(index.Outputs) != 0 {
		t.Errorf("Outputs length = %d, want 0", len(index.Outputs))
	}

	// Test duplicate initialization
	if err := InitDesign(ctx, "another-topic", "design/another"); !errors.Is(err, ErrDesignExists) {
		t.Errorf("InitDesign() duplicate = %v, want ErrDesignExists", err)
	}
}

func TestExists(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Should not exist initially
	if Exists(ctx) {
		t.Error("Exists() = true, want false before initialization")
	}

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// Should exist now
	if !Exists(ctx) {
		t.Error("Exists() = false, want true after initialization")
	}
}

func TestAddInput(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// Add input
	path := ".sow/exploration/oauth.md"
	description := "OAuth research findings"
	tags := []string{"oauth", "research"}

	if err := AddInput(ctx, "exploration", path, description, tags); err != nil {
		t.Fatalf("AddInput() failed: %v", err)
	}

	// Verify input was added
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Inputs) != 1 {
		t.Fatalf("Inputs length = %d, want 1", len(index.Inputs))
	}

	input := index.Inputs[0]
	if input.Type != "exploration" {
		t.Errorf("Type = %q, want %q", input.Type, "exploration")
	}

	if input.Path != path {
		t.Errorf("Path = %q, want %q", input.Path, path)
	}

	if input.Description != description {
		t.Errorf("Description = %q, want %q", input.Description, description)
	}

	if len(input.Tags) != 2 {
		t.Fatalf("Tags length = %d, want 2", len(input.Tags))
	}

	// Test adding duplicate input
	if err := AddInput(ctx, "exploration", path, "Different description", tags); !errors.Is(err, ErrInputExists) {
		t.Errorf("AddInput() duplicate = %v, want ErrInputExists", err)
	}
}

func TestRemoveInput(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design and add input
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	path := "test.md"
	if err := AddInput(ctx, "file", path, "Test file", []string{"tag"}); err != nil {
		t.Fatalf("AddInput() failed: %v", err)
	}

	// Remove input
	if err := RemoveInput(ctx, path); err != nil {
		t.Fatalf("RemoveInput() failed: %v", err)
	}

	// Verify removal
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Inputs) != 0 {
		t.Errorf("Inputs length after removal = %d, want 0", len(index.Inputs))
	}

	// Test removing non-existent input
	if err := RemoveInput(ctx, "nonexistent.md"); !errors.Is(err, ErrInputNotFound) {
		t.Errorf("RemoveInput() non-existent = %v, want ErrInputNotFound", err)
	}
}

func TestGetInput(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design and add input
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	path := "test.md"
	description := "Test description"
	if err := AddInput(ctx, "file", path, description, nil); err != nil {
		t.Fatalf("AddInput() failed: %v", err)
	}

	// Get input
	input, err := GetInput(ctx, path)
	if err != nil {
		t.Fatalf("GetInput() failed: %v", err)
	}

	if input.Path != path {
		t.Errorf("Path = %q, want %q", input.Path, path)
	}

	if input.Description != description {
		t.Errorf("Description = %q, want %q", input.Description, description)
	}

	// Test getting non-existent input
	if _, err := GetInput(ctx, "nonexistent.md"); !errors.Is(err, ErrInputNotFound) {
		t.Errorf("GetInput() non-existent = %v, want ErrInputNotFound", err)
	}
}

func TestListInputs(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// List should be empty initially
	inputs, err := ListInputs(ctx)
	if err != nil {
		t.Fatalf("ListInputs() failed: %v", err)
	}

	if len(inputs) != 0 {
		t.Errorf("Initial inputs length = %d, want 0", len(inputs))
	}

	// Add multiple inputs
	if err := AddInput(ctx, "file", "test1.md", "First", nil); err != nil {
		t.Fatalf("AddInput() failed: %v", err)
	}

	if err := AddInput(ctx, "file", "test2.md", "Second", nil); err != nil {
		t.Fatalf("AddInput() failed: %v", err)
	}

	// List should have 2 inputs
	inputs, err = ListInputs(ctx)
	if err != nil {
		t.Fatalf("ListInputs() after adding = %v", err)
	}

	if len(inputs) != 2 {
		t.Errorf("Inputs length = %d, want 2", len(inputs))
	}
}

func TestAddOutput(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// Add output
	path := "adr-001.md"
	description := "Test ADR"
	target := ".sow/knowledge/adrs/"
	docType := "adr"
	tags := []string{"decision"}

	if err := AddOutput(ctx, path, description, target, docType, tags); err != nil {
		t.Fatalf("AddOutput() failed: %v", err)
	}

	// Verify output was added
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Outputs) != 1 {
		t.Fatalf("Outputs length = %d, want 1", len(index.Outputs))
	}

	output := index.Outputs[0]
	if output.Path != path {
		t.Errorf("Path = %q, want %q", output.Path, path)
	}

	if output.Description != description {
		t.Errorf("Description = %q, want %q", output.Description, description)
	}

	if output.Target_location != target {
		t.Errorf("Target_location = %q, want %q", output.Target_location, target)
	}

	if output.Type != docType {
		t.Errorf("Type = %q, want %q", output.Type, docType)
	}

	if len(output.Tags) != 1 {
		t.Fatalf("Tags length = %d, want 1", len(output.Tags))
	}

	// Test adding duplicate output
	if err := AddOutput(ctx, path, "Different", target, docType, tags); !errors.Is(err, ErrOutputExists) {
		t.Errorf("AddOutput() duplicate = %v, want ErrOutputExists", err)
	}
}

func TestRemoveOutput(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design and add output
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	path := "adr-001.md"
	if err := AddOutput(ctx, path, "Test ADR", ".sow/knowledge/adrs/", "adr", nil); err != nil {
		t.Fatalf("AddOutput() failed: %v", err)
	}

	// Remove output
	if err := RemoveOutput(ctx, path); err != nil {
		t.Fatalf("RemoveOutput() failed: %v", err)
	}

	// Verify removal
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Outputs) != 0 {
		t.Errorf("Outputs length after removal = %d, want 0", len(index.Outputs))
	}

	// Test removing non-existent output
	if err := RemoveOutput(ctx, "nonexistent.md"); !errors.Is(err, ErrOutputNotFound) {
		t.Errorf("RemoveOutput() non-existent = %v, want ErrOutputNotFound", err)
	}
}

func TestUpdateOutputTarget(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design and add output
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	path := "adr-001.md"
	originalTarget := ".sow/knowledge/adrs/"
	if err := AddOutput(ctx, path, "Test ADR", originalTarget, "adr", nil); err != nil {
		t.Fatalf("AddOutput() failed: %v", err)
	}

	// Update target
	newTarget := ".sow/knowledge/decisions/"
	if err := UpdateOutputTarget(ctx, path, newTarget); err != nil {
		t.Fatalf("UpdateOutputTarget() failed: %v", err)
	}

	// Verify update
	output, err := GetOutput(ctx, path)
	if err != nil {
		t.Fatalf("GetOutput() failed: %v", err)
	}

	if output.Target_location != newTarget {
		t.Errorf("Target_location after update = %q, want %q", output.Target_location, newTarget)
	}

	// Test updating non-existent output
	if err := UpdateOutputTarget(ctx, "nonexistent.md", ".sow/"); !errors.Is(err, ErrOutputNotFound) {
		t.Errorf("UpdateOutputTarget() non-existent = %v, want ErrOutputNotFound", err)
	}
}

func TestGetOutput(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design and add output
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	path := "adr-001.md"
	description := "Test ADR"
	target := ".sow/knowledge/adrs/"
	if err := AddOutput(ctx, path, description, target, "adr", nil); err != nil {
		t.Fatalf("AddOutput() failed: %v", err)
	}

	// Get output
	output, err := GetOutput(ctx, path)
	if err != nil {
		t.Fatalf("GetOutput() failed: %v", err)
	}

	if output.Path != path {
		t.Errorf("Path = %q, want %q", output.Path, path)
	}

	if output.Description != description {
		t.Errorf("Description = %q, want %q", output.Description, description)
	}

	// Test getting non-existent output
	if _, err := GetOutput(ctx, "nonexistent.md"); !errors.Is(err, ErrOutputNotFound) {
		t.Errorf("GetOutput() non-existent = %v, want ErrOutputNotFound", err)
	}
}

func TestListOutputs(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// List should be empty initially
	outputs, err := ListOutputs(ctx)
	if err != nil {
		t.Fatalf("ListOutputs() failed: %v", err)
	}

	if len(outputs) != 0 {
		t.Errorf("Initial outputs length = %d, want 0", len(outputs))
	}

	// Add multiple outputs
	if err := AddOutput(ctx, "adr-001.md", "First ADR", ".sow/knowledge/adrs/", "adr", nil); err != nil {
		t.Fatalf("AddOutput() failed: %v", err)
	}

	if err := AddOutput(ctx, "architecture.md", "Architecture doc", ".sow/knowledge/architecture/", "architecture", nil); err != nil {
		t.Fatalf("AddOutput() failed: %v", err)
	}

	// List should have 2 outputs
	outputs, err = ListOutputs(ctx)
	if err != nil {
		t.Fatalf("ListOutputs() after adding = %v", err)
	}

	if len(outputs) != 2 {
		t.Errorf("Outputs length = %d, want 2", len(outputs))
	}
}

func TestUpdateStatus(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// Update to in_review
	if err := UpdateStatus(ctx, "in_review"); err != nil {
		t.Fatalf("UpdateStatus() to in_review failed: %v", err)
	}

	// Verify status
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if index.Design.Status != "in_review" {
		t.Errorf("Status = %q, want %q", index.Design.Status, "in_review")
	}

	// Update to completed
	if err := UpdateStatus(ctx, "completed"); err != nil {
		t.Fatalf("UpdateStatus() to completed failed: %v", err)
	}

	index, err = LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if index.Design.Status != "completed" {
		t.Errorf("Status = %q, want %q", index.Design.Status, "completed")
	}

	// Test invalid status
	if err := UpdateStatus(ctx, "invalid"); err == nil {
		t.Error("UpdateStatus() with invalid status succeeded, want error")
	}
}

func TestDelete(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create design
	if err := InitDesign(ctx, "test", "design/test"); err != nil {
		t.Fatalf("InitDesign() failed: %v", err)
	}

	// Verify it exists
	if !Exists(ctx) {
		t.Fatal("Exists() = false after initialization")
	}

	// Delete it
	if err := Delete(ctx); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify it no longer exists
	if Exists(ctx) {
		t.Error("Exists() = true after deletion, want false")
	}

	// Test deleting non-existent design
	if err := Delete(ctx); !errors.Is(err, ErrNoDesign) {
		t.Errorf("Delete() non-existent = %v, want ErrNoDesign", err)
	}
}
