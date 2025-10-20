package refs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmgilman/sow/cli/schemas"
)

func TestNewManagerWithCache(t *testing.T) {
	m := NewManagerWithCache("/cache", "/sow")
	if m.cacheDir != "/cache" {
		t.Errorf("Manager.cacheDir = %q, want %q", m.cacheDir, "/cache")
	}
	if m.sowDir != "/sow" {
		t.Errorf("Manager.sowDir = %q, want %q", m.sowDir, "/sow")
	}
}

func TestManager_Install_FileType_Knowledge(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create manager
	m := NewManagerWithCache(cacheDir, sowDir)

	// Create ref
	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:       "style-guide",
		Source:   sourceURL,
		Semantic: "knowledge",
		Link:     "style-guide",
	}

	// Install the ref
	workspacePath, err := m.Install(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Install() error = %v", err)
	}

	expectedWorkspacePath := filepath.Join(sowDir, "refs", "style-guide")
	if workspacePath != expectedWorkspacePath {
		t.Errorf("Manager.Install() workspace path = %q, want %q", workspacePath, expectedWorkspacePath)
	}

	// Verify workspace symlink was created
	linkInfo, err := os.Lstat(workspacePath)
	if err != nil {
		t.Fatalf("Workspace symlink does not exist: %v", err)
	}

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Workspace path is not a symlink")
	}

	// Verify symlink points to cache
	target, err := os.Readlink(workspacePath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	expectedCachePath := filepath.Join(cacheDir, "file", ref.Id)
	if target != expectedCachePath {
		t.Errorf("Symlink target = %q, want %q", target, expectedCachePath)
	}

	// Verify cache symlink exists and points to source
	cacheInfo, err := os.Lstat(expectedCachePath)
	if err != nil {
		t.Fatalf("Cache symlink does not exist: %v", err)
	}

	if cacheInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Cache path is not a symlink")
	}

	cacheTarget, err := os.Readlink(expectedCachePath)
	if err != nil {
		t.Fatalf("Failed to read cache symlink: %v", err)
	}

	if cacheTarget != sourceDir {
		t.Errorf("Cache symlink target = %q, want %q", cacheTarget, sourceDir)
	}
}

func TestManager_Install_FileType_Code(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create manager
	m := NewManagerWithCache(cacheDir, sowDir)

	// Create ref with code semantic
	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:       "example-code",
		Source:   sourceURL,
		Semantic: "code",
		Link:     "examples",
	}

	// Install the ref
	workspacePath, err := m.Install(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Install() error = %v", err)
	}

	expectedWorkspacePath := filepath.Join(sowDir, "refs", "examples")
	if workspacePath != expectedWorkspacePath {
		t.Errorf("Manager.Install() workspace path = %q, want %q", workspacePath, expectedWorkspacePath)
	}

	// Verify workspace symlink was created in refs/
	if _, err := os.Lstat(workspacePath); err != nil {
		t.Fatalf("Workspace symlink does not exist: %v", err)
	}
}

func TestManager_Install_InvalidSource(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	m := NewManagerWithCache(cacheDir, sowDir)

	// Create ref with non-existent source
	ref := &schemas.Ref{
		Id:       "missing",
		Source:   "file:///nonexistent/path",
		Semantic: "knowledge",
		Link:     "missing",
	}

	_, err := m.Install(ctx, ref)
	if err == nil {
		t.Error("Manager.Install() expected error for non-existent source, got nil")
	}
}

func TestManager_Update(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create manager
	m := NewManagerWithCache(cacheDir, sowDir)

	// Create and install ref
	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:       "test-update",
		Source:   sourceURL,
		Semantic: "knowledge",
		Link:     "test-update",
	}

	workspacePath, err := m.Install(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Install() error = %v", err)
	}

	// Verify initial symlink
	if _, err := os.Lstat(workspacePath); err != nil {
		t.Fatalf("Workspace symlink does not exist after install: %v", err)
	}

	// Update the ref
	err = m.Update(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Update() error = %v", err)
	}

	// Verify symlink still exists and is correct
	linkInfo, err := os.Lstat(workspacePath)
	if err != nil {
		t.Fatalf("Workspace symlink does not exist after update: %v", err)
	}

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Workspace path is not a symlink after update")
	}

	target, err := os.Readlink(workspacePath)
	if err != nil {
		t.Fatalf("Failed to read symlink after update: %v", err)
	}

	expectedCachePath := filepath.Join(cacheDir, "file", ref.Id)
	if target != expectedCachePath {
		t.Errorf("Symlink target after update = %q, want %q", target, expectedCachePath)
	}
}

func TestManager_Update_RecreatesSymlink(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create manager
	m := NewManagerWithCache(cacheDir, sowDir)

	// Create and install ref
	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:       "test-recreate",
		Source:   sourceURL,
		Semantic: "knowledge",
		Link:     "test-recreate",
	}

	workspacePath, err := m.Install(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Install() error = %v", err)
	}

	// Remove workspace symlink to simulate it being deleted
	if err := os.Remove(workspacePath); err != nil {
		t.Fatalf("Failed to remove workspace symlink: %v", err)
	}

	// Update should recreate the symlink
	err = m.Update(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Update() error = %v", err)
	}

	// Verify symlink was recreated
	if _, err := os.Lstat(workspacePath); err != nil {
		t.Fatalf("Workspace symlink was not recreated: %v", err)
	}
}

func TestManager_Remove(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create manager
	m := NewManagerWithCache(cacheDir, sowDir)

	// Create and install ref
	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:       "test-remove",
		Source:   sourceURL,
		Semantic: "knowledge",
		Link:     "test-remove",
	}

	workspacePath, err := m.Install(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Install() error = %v", err)
	}

	cachePath := filepath.Join(cacheDir, "file", ref.Id)

	// Verify both symlinks exist
	if _, err := os.Lstat(workspacePath); err != nil {
		t.Fatalf("Workspace symlink does not exist: %v", err)
	}
	if _, err := os.Lstat(cachePath); err != nil {
		t.Fatalf("Cache symlink does not exist: %v", err)
	}

	// Remove the ref
	err = m.Remove(ctx, ref)
	if err != nil {
		t.Fatalf("Manager.Remove() error = %v", err)
	}

	// Verify workspace symlink was removed
	if _, err := os.Lstat(workspacePath); err == nil {
		t.Error("Workspace symlink still exists after remove")
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking workspace symlink: %v", err)
	}

	// Verify cache symlink was removed
	if _, err := os.Lstat(cachePath); err == nil {
		t.Error("Cache symlink still exists after remove")
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking cache symlink: %v", err)
	}
}

func TestManager_Remove_AlreadyRemoved(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	sowDir := filepath.Join(tmpDir, ".sow")

	m := NewManagerWithCache(cacheDir, sowDir)

	// Create ref (not installed)
	ref := &schemas.Ref{
		Id:       "not-installed",
		Source:   "file:///some/path",
		Semantic: "knowledge",
		Link:     "not-installed",
	}

	// Remove should not error even if ref is not installed
	err := m.Remove(ctx, ref)
	if err != nil {
		t.Errorf("Manager.Remove() error = %v, want nil (idempotent)", err)
	}
}

func TestManager_workspacePath(t *testing.T) {
	m := NewManagerWithCache("/cache", "/sow")

	tests := []struct {
		name     string
		ref      *schemas.Ref
		expected string
	}{
		{
			name: "knowledge semantic",
			ref: &schemas.Ref{
				Semantic: "knowledge",
				Link:     "style-guide",
			},
			expected: "/sow/refs/style-guide",
		},
		{
			name: "code semantic",
			ref: &schemas.Ref{
				Semantic: "code",
				Link:     "examples",
			},
			expected: "/sow/refs/examples",
		},
		{
			name: "any semantic goes to refs",
			ref: &schemas.Ref{
				Semantic: "other",
				Link:     "test",
			},
			expected: "/sow/refs/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.workspacePath(tt.ref)
			if result != tt.expected {
				t.Errorf("Manager.workspacePath() = %q, want %q", result, tt.expected)
			}
		})
	}
}
