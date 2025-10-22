package refs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/schemas"
)

// CacheManager orchestrates ref caching and workspace symlinking.
// This is a low-level manager that handles the physical caching operations.
type CacheManager struct {
	cacheDir string // Base cache directory (e.g., ~/.cache/sow/refs)
	sowDir   string // .sow directory path
}

// NewCacheManager creates a new refs cache manager using the default cache directory.
// The default cache directory is ~/.cache/sow/refs.
func NewCacheManager(sowDir string) (*CacheManager, error) {
	cacheDir, err := DefaultCacheDir()
	if err != nil {
		return nil, err
	}
	return &CacheManager{
		cacheDir: cacheDir,
		sowDir:   sowDir,
	}, nil
}

// NewCacheManagerWithCache creates a new refs cache manager with a custom cache directory.
// This is primarily for testing; production code should use NewCacheManager.
func NewCacheManagerWithCache(cacheDir, sowDir string) *CacheManager {
	return &CacheManager{
		cacheDir: cacheDir,
		sowDir:   sowDir,
	}
}

// DefaultCacheDir returns the default cache directory for refs.
// This is ~/.cache/sow/refs on all platforms.
func DefaultCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cache", "sow", "refs"), nil
}

// Install installs a ref by caching it and creating a workspace symlink.
// Returns the workspace symlink path.
func (m *CacheManager) Install(ctx context.Context, ref *schemas.Ref) (string, error) {
	// Infer type from URL
	typeName, err := InferTypeFromURL(ref.Source)
	if err != nil {
		return "", fmt.Errorf("failed to infer type from URL: %w", err)
	}

	// Get the ref type implementation
	refType, err := TypeForScheme(ctx, typeName)
	if err != nil {
		return "", fmt.Errorf("failed to get ref type: %w", err)
	}

	// Validate config for this type
	if err := refType.ValidateConfig(ref.Config); err != nil {
		return "", fmt.Errorf("invalid config for type %s: %w", typeName, err)
	}

	// Cache the ref (downloads/clones to cache directory)
	cachePath, err := refType.Cache(ctx, m.cacheDir, ref)
	if err != nil {
		return "", fmt.Errorf("failed to cache ref: %w", err)
	}

	// Apply subpath if specified (for git refs)
	if ref.Config.Path != "" {
		cachePath = filepath.Join(cachePath, ref.Config.Path)
	}

	// Determine workspace path based on semantic type
	workspacePath := m.workspacePath(ref)

	// Create workspace symlink to cache
	if err := m.createWorkspaceSymlink(cachePath, workspacePath); err != nil {
		return "", fmt.Errorf("failed to create workspace symlink: %w", err)
	}

	return workspacePath, nil
}

// Update updates a ref by refreshing its cache and verifying the symlink.
func (m *CacheManager) Update(ctx context.Context, ref *schemas.Ref) error {
	// Infer type from URL
	typeName, err := InferTypeFromURL(ref.Source)
	if err != nil {
		return fmt.Errorf("failed to infer type from URL: %w", err)
	}

	// Get the ref type implementation
	refType, err := TypeForScheme(ctx, typeName)
	if err != nil {
		return fmt.Errorf("failed to get ref type: %w", err)
	}

	// Update the cache
	// Note: We don't have the CachedRef here yet (that would come from cache index)
	// For now, we'll pass nil and let Update handle it
	if err := refType.Update(ctx, m.cacheDir, ref, nil); err != nil {
		return fmt.Errorf("failed to update ref cache: %w", err)
	}

	// Verify workspace symlink still exists and is correct
	workspacePath := m.workspacePath(ref)
	cachePath := refType.CachePath(m.cacheDir, ref)

	// Apply subpath if specified
	if ref.Config.Path != "" {
		cachePath = filepath.Join(cachePath, ref.Config.Path)
	}

	// Verify and fix workspace symlink
	if err := m.verifyWorkspaceSymlink(cachePath, workspacePath); err != nil {
		return fmt.Errorf("failed to verify workspace symlink: %w", err)
	}

	return nil
}

// verifyWorkspaceSymlink checks if workspace symlink exists and points to correct cache path.
// If it doesn't exist or points to wrong location, it will be created/fixed.
func (m *CacheManager) verifyWorkspaceSymlink(cachePath, workspacePath string) error {
	linkInfo, err := os.Lstat(workspacePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Symlink doesn't exist, create it
			return m.createWorkspaceSymlink(cachePath, workspacePath)
		}
		return fmt.Errorf("failed to check workspace symlink: %w", err)
	}

	// Symlink exists, verify it's actually a symlink
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("workspace path exists but is not a symlink: %s", workspacePath)
	}

	// Check if it points to correct location
	target, err := os.Readlink(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to read workspace symlink: %w", err)
	}

	if target == cachePath {
		// Already correct
		return nil
	}

	// Points to wrong location, recreate it
	if err := os.Remove(workspacePath); err != nil {
		return fmt.Errorf("failed to remove incorrect symlink: %w", err)
	}

	return m.createWorkspaceSymlink(cachePath, workspacePath)
}

// Remove removes a ref by cleaning up its cache and workspace symlink.
func (m *CacheManager) Remove(ctx context.Context, ref *schemas.Ref) error {
	// Infer type from URL
	typeName, err := InferTypeFromURL(ref.Source)
	if err != nil {
		return fmt.Errorf("failed to infer type from URL: %w", err)
	}

	// Get the ref type implementation
	refType, err := TypeForScheme(ctx, typeName)
	if err != nil {
		return fmt.Errorf("failed to get ref type: %w", err)
	}

	// Remove workspace symlink
	workspacePath := m.workspacePath(ref)
	if err := os.Remove(workspacePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove workspace symlink: %w", err)
	}

	// Cleanup cache
	if err := refType.Cleanup(ctx, m.cacheDir, ref); err != nil {
		return fmt.Errorf("failed to cleanup ref cache: %w", err)
	}

	return nil
}

// workspacePath determines the workspace symlink path.
// All refs go to .sow/refs/{link} regardless of semantic type.
func (m *CacheManager) workspacePath(ref *schemas.Ref) string {
	return filepath.Join(m.sowDir, "refs", ref.Link)
}

// createWorkspaceSymlink creates a symlink from workspace to cache.
func (m *CacheManager) createWorkspaceSymlink(cachePath, workspacePath string) error {
	// Create parent directory if needed
	parentDir := filepath.Dir(workspacePath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Check if symlink already exists
	if linkInfo, err := os.Lstat(workspacePath); err == nil {
		// Symlink exists, check if it's correct
		if linkInfo.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("workspace path exists but is not a symlink: %s", workspacePath)
		}

		target, err := os.Readlink(workspacePath)
		if err != nil {
			return fmt.Errorf("failed to read existing symlink: %w", err)
		}

		if target == cachePath {
			// Already correct, nothing to do
			return nil
		}

		// Points to wrong location, remove it
		if err := os.Remove(workspacePath); err != nil {
			return fmt.Errorf("failed to remove incorrect symlink: %w", err)
		}
	}

	// Create the symlink
	if err := os.Symlink(cachePath, workspacePath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}
