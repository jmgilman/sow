package refs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmgilman/sow/libs/schemas"
)

func TestFileType_Name(t *testing.T) {
	f := &FileType{}
	if f.Name() != "file" {
		t.Errorf("FileType.Name() = %q, want %q", f.Name(), "file")
	}
}

func TestFileType_IsEnabled(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	enabled, err := f.IsEnabled(ctx)
	if err != nil {
		t.Fatalf("FileType.IsEnabled() error = %v", err)
	}

	// File type should always be enabled
	if !enabled {
		t.Error("FileType.IsEnabled() = false, want true (always enabled)")
	}
}

func TestFileType_IsStale(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	ref := &schemas.Ref{
		Id:     "test-file",
		Source: "file:///some/path",
	}

	stale, err := f.IsStale(ctx, "/cache", ref, nil)
	if err != nil {
		t.Fatalf("FileType.IsStale() error = %v", err)
	}

	// File refs should never be stale
	if stale {
		t.Error("FileType.IsStale() = true, want false (file refs never stale)")
	}
}

func TestFileType_Update(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	ref := &schemas.Ref{
		Id:     "test-file",
		Source: "file:///some/path",
	}

	// Update should be a no-op (no error)
	err := f.Update(ctx, "/cache", ref, nil)
	if err != nil {
		t.Errorf("FileType.Update() error = %v, want nil (no-op)", err)
	}
}

func TestFileType_ValidateConfig(t *testing.T) {
	f := &FileType{}

	tests := []struct {
		name      string
		config    schemas.RefConfig
		wantError bool
	}{
		{
			name: "valid empty config",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "",
			},
			wantError: false,
		},
		{
			name: "invalid with branch",
			config: schemas.RefConfig{
				Branch: "main",
				Path:   "",
			},
			wantError: true,
		},
		{
			name: "invalid with path",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "docs",
			},
			wantError: true,
		},
		{
			name: "invalid with both",
			config: schemas.RefConfig{
				Branch: "main",
				Path:   "docs",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.ValidateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("FileType.ValidateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestFileType_CachePath(t *testing.T) {
	f := &FileType{}

	ref := &schemas.Ref{
		Id:     "test-file",
		Source: "file:///Users/josh/docs",
	}

	cacheDir := "/cache"
	path := f.CachePath(cacheDir, ref)

	expected := filepath.Join("/cache", "file", ref.Id)
	if path != expected {
		t.Errorf("FileType.CachePath() = %q, want %q", path, expected)
	}
}

func TestFileType_Interface(_ *testing.T) {
	// Verify FileType implements RefType interface
	var _ RefType = (*FileType)(nil)
}

func TestFileType_Init(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	err := f.Init(ctx, cacheDir)
	if err != nil {
		t.Fatalf("FileType.Init() error = %v", err)
	}

	// Verify file cache directory was created
	fileCacheDir := filepath.Join(cacheDir, "file")
	info, err := os.Stat(fileCacheDir)
	if err != nil {
		t.Fatalf("Failed to stat file cache directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("File cache path is not a directory")
	}
}

func TestFileType_Cache(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	// Create temp directories for testing
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create a ref with file URL
	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:     "test-file",
		Source: sourceURL,
	}

	// Test caching
	cachePath, err := f.Cache(ctx, cacheDir, ref)
	if err != nil {
		t.Fatalf("FileType.Cache() error = %v", err)
	}

	expectedPath := filepath.Join(cacheDir, "file", ref.Id)
	if cachePath != expectedPath {
		t.Errorf("FileType.Cache() path = %q, want %q", cachePath, expectedPath)
	}

	// Verify symlink was created
	linkInfo, err := os.Lstat(cachePath)
	if err != nil {
		t.Fatalf("Failed to stat cache path: %v", err)
	}

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Cache path is not a symlink")
	}

	// Verify symlink target
	target, err := os.Readlink(cachePath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if target != sourceDir {
		t.Errorf("Symlink target = %q, want %q", target, sourceDir)
	}

	// Test caching again (should succeed with existing symlink)
	cachePath2, err := f.Cache(ctx, cacheDir, ref)
	if err != nil {
		t.Fatalf("FileType.Cache() second call error = %v", err)
	}

	if cachePath2 != cachePath {
		t.Errorf("FileType.Cache() second call path = %q, want %q", cachePath2, cachePath)
	}
}

func TestFileType_Cache_SourceNotExists(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create a ref with non-existent source
	ref := &schemas.Ref{
		Id:     "test-missing",
		Source: "file:///nonexistent/path",
	}

	_, err := f.Cache(ctx, cacheDir, ref)
	if err == nil {
		t.Error("FileType.Cache() expected error for non-existent source, got nil")
	}
}

func TestFileType_Cache_SourceIsFile(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create a file (not directory)
	sourceFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(sourceFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	sourceURL, err := PathToFileURL(sourceFile)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:     "test-file",
		Source: sourceURL,
	}

	_, err = f.Cache(ctx, cacheDir, ref)
	if err == nil {
		t.Error("FileType.Cache() expected error for file source, got nil")
	}
}

func TestFileType_Cleanup(t *testing.T) {
	f := &FileType{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	sourceURL, err := PathToFileURL(sourceDir)
	if err != nil {
		t.Fatalf("Failed to convert path to URL: %v", err)
	}

	ref := &schemas.Ref{
		Id:     "test-cleanup",
		Source: sourceURL,
	}

	// Cache the ref first
	cachePath, err := f.Cache(ctx, cacheDir, ref)
	if err != nil {
		t.Fatalf("FileType.Cache() error = %v", err)
	}

	// Verify symlink exists
	if _, err := os.Lstat(cachePath); err != nil {
		t.Fatalf("Symlink does not exist: %v", err)
	}

	// Cleanup
	err = f.Cleanup(ctx, cacheDir, ref)
	if err != nil {
		t.Fatalf("FileType.Cleanup() error = %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(cachePath); err == nil {
		t.Error("Symlink still exists after cleanup")
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking symlink after cleanup: %v", err)
	}

	// Cleanup again should not error
	err = f.Cleanup(ctx, cacheDir, ref)
	if err != nil {
		t.Errorf("FileType.Cleanup() second call error = %v, want nil", err)
	}
}
