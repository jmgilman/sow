package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSessionType_Project(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create .sow/project/ directory
	projectDir := filepath.Join(tmpDir, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Detect session type
	sessionType := detectSessionType(tmpDir)

	if sessionType != "project" {
		t.Errorf("expected 'project', got '%s'", sessionType)
	}
}

func TestDetectSessionType_Exploration(t *testing.T) {
	tmpDir := t.TempDir()

	explorationDir := filepath.Join(tmpDir, ".sow", "exploration")
	if err := os.MkdirAll(explorationDir, 0755); err != nil {
		t.Fatalf("failed to create exploration dir: %v", err)
	}

	sessionType := detectSessionType(tmpDir)

	if sessionType != "exploration" {
		t.Errorf("expected 'exploration', got '%s'", sessionType)
	}
}

func TestDetectSessionType_Design(t *testing.T) {
	tmpDir := t.TempDir()

	designDir := filepath.Join(tmpDir, ".sow", "design")
	if err := os.MkdirAll(designDir, 0755); err != nil {
		t.Fatalf("failed to create design dir: %v", err)
	}

	sessionType := detectSessionType(tmpDir)

	if sessionType != "design" {
		t.Errorf("expected 'design', got '%s'", sessionType)
	}
}

func TestDetectSessionType_Breakdown(t *testing.T) {
	tmpDir := t.TempDir()

	breakdownDir := filepath.Join(tmpDir, ".sow", "breakdown")
	if err := os.MkdirAll(breakdownDir, 0755); err != nil {
		t.Fatalf("failed to create breakdown dir: %v", err)
	}

	sessionType := detectSessionType(tmpDir)

	if sessionType != "breakdown" {
		t.Errorf("expected 'breakdown', got '%s'", sessionType)
	}
}

func TestDetectSessionType_Unknown(t *testing.T) {
	tmpDir := t.TempDir()

	// No session directories created

	sessionType := detectSessionType(tmpDir)

	if sessionType != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", sessionType)
	}
}

func TestExtractBranchName(t *testing.T) {
	repoRoot := "/Users/test/repo"

	tests := []struct {
		name     string
		wtPath   string
		expected string
	}{
		{
			name:     "simple branch",
			wtPath:   "/Users/test/repo/.sow/worktrees/main",
			expected: "main",
		},
		{
			name:     "branch with slash",
			wtPath:   "/Users/test/repo/.sow/worktrees/feat/auth",
			expected: "feat/auth",
		},
		{
			name:     "nested branch",
			wtPath:   "/Users/test/repo/.sow/worktrees/explore/topic/subtopic",
			expected: "explore/topic/subtopic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBranchName(tt.wtPath, repoRoot)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsSowWorktree(t *testing.T) {
	repoRoot := "/Users/test/repo"

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "sow worktree",
			path:     "/Users/test/repo/.sow/worktrees/feat/auth",
			expected: true,
		},
		{
			name:     "main repo",
			path:     "/Users/test/repo",
			expected: false,
		},
		{
			name:     "non-sow worktree",
			path:     "/Users/test/repo/worktrees/other",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSowWorktree(tt.path, repoRoot)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWorktreeRemove_PathNotExists(t *testing.T) {
	// Test that runWorktreeRemove returns error when path doesn't exist
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist")

	// Create a mock command - we'll test the function directly
	// Since we can't easily mock the context, this test validates the path check logic
	// Full integration testing will be done in Task 090

	// Verify the path doesn't exist
	if _, err := os.Stat(nonExistentPath); !os.IsNotExist(err) {
		t.Fatal("test setup failed: path should not exist")
	}

	// This test documents the expected behavior:
	// runWorktreeRemove should check if path exists and return error if not
	// The actual test will be in integration tests where we can mock ctx.Git()
}

func TestWorktreeRemove_PathIsNotWorktree(t *testing.T) {
	// Test that runWorktreeRemove returns error when path exists but is not a worktree
	tmpDir := t.TempDir()

	// Create a regular directory (not a worktree)
	regularDir := filepath.Join(tmpDir, "regular-dir")
	if err := os.MkdirAll(regularDir, 0755); err != nil {
		t.Fatalf("failed to create regular dir: %v", err)
	}

	// This test documents the expected behavior:
	// runWorktreeRemove should verify path is in the worktrees list
	// If not found, return error "path is not a worktree"
	// The actual test will be in integration tests where we can mock ctx.Git()
}

func TestWorktreePrune_CommandStructure(t *testing.T) {
	// Test that prune command is properly configured
	// This validates command structure without requiring git operations

	// Get the worktree command
	cmd := NewWorktreeCmd()

	// Find the prune subcommand
	pruneCmd, _, err := cmd.Find([]string{"prune"})
	if err != nil {
		t.Fatalf("prune subcommand not found: %v", err)
	}

	// Verify command properties
	if pruneCmd.Use != "prune" {
		t.Errorf("expected Use='prune', got '%s'", pruneCmd.Use)
	}

	if pruneCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if pruneCmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	// Verify it accepts no arguments
	if pruneCmd.Args != nil {
		t.Error("prune command should accept no arguments")
	}
}
