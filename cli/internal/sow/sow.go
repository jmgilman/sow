// Package sow provides the core foundation for the sow CLI.
//
// This package contains:
//   - Context: Unified access to filesystem, git, and GitHub
//   - Core initialization: Init(), DetectContext()
//   - Domain errors used across the system
//   - Option patterns for project operations
//   - Formatting utilities for command output
//   - Validation infrastructure
//
// Business logic for projects and tasks lives in the internal/project package.
// External reference management lives in the internal/refs package.
package sow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// StructureVersion is the current .sow structure version.
	// This version is written to .sow/.version during initialization.
	StructureVersion = "1.0.0"
)

// isInitialized checks if sow has been initialized in the repository.
// This is an internal helper. Commands should use context.IsInitialized() instead.
func isInitialized(repoRoot string) bool {
	sowPath := filepath.Join(repoRoot, ".sow")
	_, err := os.Stat(sowPath)
	return err == nil
}

// Init initializes the sow structure in the repository.
// Creates .sow/ directory with knowledge/ and refs/ subdirectories.
func Init(repoRoot string) error {
	// Check if in a git repository
	gitPath := filepath.Join(repoRoot, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		return fmt.Errorf("not in git repository")
	}

	// Check if already initialized
	if isInitialized(repoRoot) {
		return fmt.Errorf(".sow directory already exists - repository already initialized")
	}

	// Create base .sow directory
	sowPath := filepath.Join(repoRoot, ".sow")
	if err := os.MkdirAll(sowPath, 0755); err != nil {
		return fmt.Errorf("failed to create .sow directory: %w", err)
	}

	// Create knowledge directory
	knowledgePath := filepath.Join(sowPath, "knowledge")
	if err := os.MkdirAll(knowledgePath, 0755); err != nil {
		return fmt.Errorf("failed to create knowledge directory: %w", err)
	}

	// Create knowledge/adrs directory
	adrsPath := filepath.Join(knowledgePath, "adrs")
	if err := os.MkdirAll(adrsPath, 0755); err != nil {
		return fmt.Errorf("failed to create adrs directory: %w", err)
	}

	// Create refs directory
	refsPath := filepath.Join(sowPath, "refs")
	if err := os.MkdirAll(refsPath, 0755); err != nil {
		return fmt.Errorf("failed to create refs directory: %w", err)
	}

	// Create .gitignore for refs
	gitignorePath := filepath.Join(refsPath, ".gitignore")
	gitignoreContent := []byte("# Ignore all symlinks and local refs\n*\n!.gitignore\n!index.json\n!index.local.json\n")
	if err := os.WriteFile(gitignorePath, gitignoreContent, 0644); err != nil {
		return fmt.Errorf("failed to create refs .gitignore: %w", err)
	}

	// Create version file
	versionPath := filepath.Join(sowPath, ".version")
	versionContent := []byte(StructureVersion + "\n")
	if err := os.WriteFile(versionPath, versionContent, 0644); err != nil {
		return fmt.Errorf("failed to create version file: %w", err)
	}

	// Create refs index with version
	indexPath := filepath.Join(refsPath, "index.json")
	indexContent := []byte(`{
  "version": "1.0.0",
  "refs": []
}
`)
	if err := os.WriteFile(indexPath, indexContent, 0644); err != nil {
		return fmt.Errorf("failed to create refs index: %w", err)
	}

	return nil
}

// DetectContext detects the current workspace context.
// Returns the context type ("none", "project", or "task") and task ID if applicable.
func DetectContext(repoRoot string) (string, string) {
	// Check if we're in a task directory
	// Task directories are at .sow/project/phases/{phase}/tasks/{task-id}/
	cwd, err := os.Getwd()
	if err != nil {
		return "none", ""
	}

	// Check if current dir is under .sow/project/phases/*/tasks/*
	relPath, err := filepath.Rel(repoRoot, cwd)
	if err != nil {
		return "none", ""
	}

	// Normalize to forward slashes for parsing
	relPath = filepath.ToSlash(relPath)

	// Split path into components
	parts := strings.Split(relPath, "/")

	if len(parts) >= 6 {
		if parts[0] == ".sow" && parts[1] == "project" && parts[2] == "phases" && parts[4] == "tasks" {
			return "task", parts[5]
		}
	}

	// Check if we're anywhere under .sow/project
	if len(parts) >= 2 && parts[0] == ".sow" && parts[1] == "project" {
		return "project", ""
	}

	return "none", ""
}

