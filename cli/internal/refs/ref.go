package refs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/libs/schemas"
)

// Ref represents a reference with operations.
// All operations automatically persist state changes.
type Ref struct {
	manager *Manager
	id      string
}

// ID returns the ref ID.
func (r *Ref) ID() string {
	return r.id
}

// Source returns the ref source URL.
func (r *Ref) Source() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}
	return ref.Source, nil
}

// Link returns the workspace symlink name.
func (r *Ref) Link() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}
	return ref.Link, nil
}

// Semantic returns the semantic type (knowledge or code).
func (r *Ref) Semantic() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}
	return ref.Semantic, nil
}

// Tags returns the topic tags.
func (r *Ref) Tags() ([]string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return nil, err
	}
	return ref.Tags, nil
}

// Description returns the ref description.
func (r *Ref) Description() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}
	return ref.Description, nil
}

// Config returns the ref configuration.
func (r *Ref) Config() (schemas.RefConfig, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return schemas.RefConfig{}, err
	}
	return ref.Config, nil
}

// Type returns the ref type (git, file, etc.).
func (r *Ref) Type() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}
	return InferTypeFromURL(ref.Source)
}

// IsLocal returns true if the ref is in the local index (not shared).
func (r *Ref) IsLocal() (bool, error) {
	_, isLocal, err := r.manager.findRefInIndexes(r.id)
	return isLocal, err
}

// Update updates the ref by refreshing its cache.
func (r *Ref) Update(ctx context.Context) error {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return err
	}

	// Create cache manager
	sowDir := filepath.Join(r.manager.ctx.RepoRoot(), ".sow")
	cacheManager, err := NewCacheManager(sowDir)
	if err != nil {
		return fmt.Errorf("failed to create refs cache manager: %w", err)
	}

	// Update the ref
	return cacheManager.Update(ctx, ref)
}

// Remove removes the ref and optionally prunes the cache.
func (r *Ref) Remove(ctx context.Context, pruneCache bool) error {
	return r.manager.Remove(ctx, r.id, pruneCache)
}

// Status checks if the ref is stale (behind remote).
// Returns true if stale, false if current.
func (r *Ref) Status(ctx context.Context) (bool, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return false, err
	}

	// Infer type
	typeName, err := InferTypeFromURL(ref.Source)
	if err != nil {
		return false, fmt.Errorf("failed to infer type: %w", err)
	}

	// Get type implementation
	refType, err := GetType(typeName)
	if err != nil {
		return false, fmt.Errorf("unknown reference type: %s", typeName)
	}

	// Check if enabled
	enabled, err := refType.IsEnabled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if type enabled: %w", err)
	}
	if !enabled {
		return false, fmt.Errorf("reference type %s is not enabled", typeName)
	}

	// Get cache directory
	cacheDir, err := DefaultCacheDir()
	if err != nil {
		return false, fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Load cache index (if it exists)
	// For now, we'll pass nil for cached ref and let IsStale handle it
	isStale, err := refType.IsStale(ctx, cacheDir, ref, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check staleness: %w", err)
	}

	return isStale, nil
}

// WorkspacePath returns the filesystem path where the ref is symlinked.
func (r *Ref) WorkspacePath() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}
	return filepath.Join(r.manager.ctx.RepoRoot(), ".sow", "refs", ref.Link), nil
}

// CachePath returns the filesystem path where the ref is cached.
func (r *Ref) CachePath() (string, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return "", err
	}

	// Infer type
	typeName, err := InferTypeFromURL(ref.Source)
	if err != nil {
		return "", fmt.Errorf("failed to infer type: %w", err)
	}

	// Get type implementation
	refType, err := GetType(typeName)
	if err != nil {
		return "", fmt.Errorf("unknown reference type: %s", typeName)
	}

	// Get cache directory
	cacheDir, err := DefaultCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Get cache path from type
	cachePath := refType.CachePath(cacheDir, ref)

	// Apply subpath if specified
	if ref.Config.Path != "" {
		cachePath = filepath.Join(cachePath, ref.Config.Path)
	}

	return cachePath, nil
}

// Exists checks if the ref's workspace symlink exists.
func (r *Ref) Exists() (bool, error) {
	workspacePath, err := r.WorkspacePath()
	if err != nil {
		return false, err
	}

	_, err = os.Lstat(workspacePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check workspace path: %w", err)
	}

	return true, nil
}

// Schema returns the full schemas.Ref structure for this ref.
// This is useful for operations that need all ref details at once.
func (r *Ref) Schema() (*schemas.Ref, error) {
	ref, _, err := r.manager.findRefInIndexes(r.id)
	if err != nil {
		return nil, err
	}
	return ref, nil
}
