package refs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/libs/schemas"
)

// FileType implements RefType for local file references.
type FileType struct{}

// Ensure FileType implements RefType.
var _ RefType = (*FileType)(nil)

// Register file type on package init.
func init() {
	Register(&FileType{})
}

// Name returns the type name.
func (f *FileType) Name() string {
	return "file"
}

// IsEnabled checks if file type is available.
// File type is always enabled (native filesystem).
func (f *FileType) IsEnabled(_ context.Context) (bool, error) {
	return true, nil
}

// Init initializes file type resources.
func (f *FileType) Init(_ context.Context, cacheDir string) error {
	// Create file cache directory
	fileCacheDir := filepath.Join(cacheDir, "file")

	if err := os.MkdirAll(fileCacheDir, 0o755); err != nil {
		return fmt.Errorf("failed to create file cache directory: %w", err)
	}

	return nil
}

// Cache creates a symlink to the local file path.
func (f *FileType) Cache(_ context.Context, cacheDir string, ref *schemas.Ref) (string, error) {
	// Parse file URL to get source path
	sourcePath, err := FileURLToPath(ref.Source)
	if err != nil {
		return "", fmt.Errorf("failed to parse file URL: %w", err)
	}

	// Validate source path exists
	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source path does not exist: %s", sourcePath)
		}
		return "", fmt.Errorf("failed to stat source path: %w", err)
	}

	// Source must be a directory
	if !info.IsDir() {
		return "", fmt.Errorf("source path must be a directory: %s", sourcePath)
	}

	// Get cache path
	cachePath := f.CachePath(cacheDir, ref)

	// Check if symlink already exists
	linkInfo, err := os.Lstat(cachePath)
	if err == nil {
		// Symlink exists, verify it's actually a symlink
		if linkInfo.Mode()&os.ModeSymlink == 0 {
			return "", fmt.Errorf("cache path exists but is not a symlink: %s", cachePath)
		}

		// Check if it points to the correct location
		target, err := os.Readlink(cachePath)
		if err != nil {
			return "", fmt.Errorf("failed to read existing symlink: %w", err)
		}

		if target == sourcePath {
			// Already cached correctly
			return cachePath, nil
		}

		// Points to wrong location, remove it
		if err := os.Remove(cachePath); err != nil {
			return "", fmt.Errorf("failed to remove incorrect symlink: %w", err)
		}
	}

	// Create parent directory for cache path
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create symlink
	if err := os.Symlink(sourcePath, cachePath); err != nil {
		return "", fmt.Errorf("failed to create symlink: %w", err)
	}

	return cachePath, nil
}

// Update is a no-op for file refs (they point directly to source).
func (f *FileType) Update(_ context.Context, _ string, _ *schemas.Ref, _ *schemas.CachedRef) error {
	// File refs don't need updates - they're symlinks to live directories
	// Changes in source directory are automatically reflected
	return nil
}

// IsStale always returns false for file refs (never stale).
func (f *FileType) IsStale(_ context.Context, _ string, _ *schemas.Ref, _ *schemas.CachedRef) (bool, error) {
	// File refs are never stale - they point directly to source
	return false, nil
}

// CachePath returns the cache path for a file ref.
func (f *FileType) CachePath(cacheDir string, ref *schemas.Ref) string {
	return filepath.Join(cacheDir, "file", ref.Id)
}

// Cleanup removes the cached file symlink.
func (f *FileType) Cleanup(_ context.Context, cacheDir string, ref *schemas.Ref) error {
	// Get cache path
	cachePath := f.CachePath(cacheDir, ref)

	// Check if it exists
	if _, err := os.Lstat(cachePath); err != nil {
		if os.IsNotExist(err) {
			// Already removed, nothing to do
			return nil
		}
		return fmt.Errorf("failed to stat cache path: %w", err)
	}

	// Remove the symlink
	if err := os.Remove(cachePath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	return nil
}

// ValidateConfig validates file-specific configuration.
func (f *FileType) ValidateConfig(config schemas.RefConfig) error {
	// File type doesn't use additional config
	// Path is in the source URL itself

	// Validate that no git-specific config is set
	if config.Branch != "" {
		return fmt.Errorf("file refs do not support branch config")
	}

	if config.Path != "" {
		return fmt.Errorf("file refs do not support path config (path is in URL)")
	}

	return nil
}
