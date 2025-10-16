package refs

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jmgilman/go/git/cache"
	"github.com/jmgilman/sow/cli/schemas"
)

// GitType implements RefType for git repositories.
type GitType struct {
	cache   *cache.RepositoryCache
	cacheMu sync.Mutex
}

// Ensure GitType implements RefType.
var _ RefType = (*GitType)(nil)

// Register git type on package init.
func init() {
	Register(&GitType{})
}

// Name returns the type name.
func (g *GitType) Name() string {
	return "git"
}

// IsEnabled checks if git is available on the system.
func (g *GitType) IsEnabled(_ context.Context) (bool, error) {
	// Check if git binary exists in PATH
	_, err := exec.LookPath("git")
	return err == nil, nil
}

// Init initializes git type resources.
func (g *GitType) Init(_ context.Context, cacheDir string) error {
	// Initialize the repository cache
	gitCacheDir := filepath.Join(cacheDir, "git")

	cache, err := cache.NewRepositoryCache(gitCacheDir)
	if err != nil {
		return fmt.Errorf("failed to create git cache: %w", err)
	}

	g.cacheMu.Lock()
	g.cache = cache
	g.cacheMu.Unlock()

	return nil
}

// ensureCache ensures the cache is initialized (lazy initialization).
func (g *GitType) ensureCache(cacheDir string) error {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	if g.cache != nil {
		return nil
	}

	gitCacheDir := filepath.Join(cacheDir, "git")
	cache, err := cache.NewRepositoryCache(gitCacheDir)
	if err != nil {
		return fmt.Errorf("failed to create git cache: %w", err)
	}

	g.cache = cache
	return nil
}

// Cache clones a git repository to the cache.
func (g *GitType) Cache(ctx context.Context, cacheDir string, ref *schemas.Ref) (string, error) {
	// Ensure cache is initialized
	if err := g.ensureCache(cacheDir); err != nil {
		return "", err
	}

	// Get git-compatible URL (strips git+ prefix for actual git operations)
	gitURL := toGitURL(ref.Source)

	// Build cache options
	opts := []cache.CacheOption{}

	// Add branch if specified
	if ref.Config.Branch != "" {
		opts = append(opts, cache.WithRef(ref.Config.Branch))
	}

	// Get checkout using ref.Id as stable cache key
	// This creates or reuses existing checkout
	checkoutPath, err := g.cache.GetCheckout(ctx, gitURL, ref.Id, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to get checkout: %w", err)
	}

	// If a subpath is specified, we'll return the checkout path
	// The caller (symlink creation) will handle pointing to the subpath
	return checkoutPath, nil
}

// Update pulls latest changes from remote.
func (g *GitType) Update(ctx context.Context, cacheDir string, ref *schemas.Ref, _ *schemas.CachedRef) error {
	// Ensure cache is initialized
	if err := g.ensureCache(cacheDir); err != nil {
		return err
	}

	// Get git-compatible URL
	gitURL := toGitURL(ref.Source)

	// Build cache options with update flag
	opts := []cache.CacheOption{
		cache.WithUpdate(), // Force refresh from remote
	}

	// Add branch if specified
	if ref.Config.Branch != "" {
		opts = append(opts, cache.WithRef(ref.Config.Branch))
	}

	// Update the checkout
	_, err := g.cache.GetCheckout(ctx, gitURL, ref.Id, opts...)
	if err != nil {
		return fmt.Errorf("failed to update checkout: %w", err)
	}

	// Note: The cache handles updating cached.Metadata internally
	// We don't need to manually update the CachedRef here
	return nil
}

// IsStale checks if cache is behind remote.
func (g *GitType) IsStale(_ context.Context, cacheDir string, _ *schemas.Ref, _ *schemas.CachedRef) (bool, error) {
	// Ensure cache is initialized
	if err := g.ensureCache(cacheDir); err != nil {
		return false, err
	}

	// For now, we'll implement a simple check:
	// Try to update and see if anything changed
	// A more sophisticated implementation would fetch and compare SHAs without updating

	// TODO: Implement proper staleness check
	// This could involve:
	// 1. Fetching from remote (git fetch)
	// 2. Comparing local HEAD with remote HEAD
	// 3. Counting commits behind (git rev-list --count HEAD..@{u})
	//
	// For now, we return false and rely on explicit Update() calls
	return false, nil
}

// CachePath returns the cache path for a git ref.
func (g *GitType) CachePath(cacheDir string, ref *schemas.Ref) string {
	// The cache uses this structure:
	// {gitCacheDir}/checkouts/{normalized_url}/{branch}/{ref.Id}/
	//
	// Since we don't have the cache instance to normalize the URL,
	// we'll construct a simplified path
	gitCacheDir := filepath.Join(cacheDir, "git")
	return filepath.Join(gitCacheDir, "checkouts", ref.Id)
}

// Cleanup removes the cached git repository.
func (g *GitType) Cleanup(_ context.Context, cacheDir string, ref *schemas.Ref) error {
	// Ensure cache is initialized
	if err := g.ensureCache(cacheDir); err != nil {
		return err
	}

	// Get git-compatible URL
	gitURL := toGitURL(ref.Source)

	// Remove the checkout using ref.Id as cache key
	if err := g.cache.RemoveCheckout(gitURL, ref.Id); err != nil {
		return fmt.Errorf("failed to remove checkout: %w", err)
	}

	return nil
}

// toGitURL converts a normalized git URL (with git+ prefix) to a URL that git can use.
// Examples:
//   - git+https://github.com/org/repo -> https://github.com/org/repo
//   - git+ssh://git@github.com/org/repo -> ssh://git@github.com/org/repo
func toGitURL(normalizedURL string) string {
	// Strip git+ prefix if present
	return strings.TrimPrefix(normalizedURL, "git+")
}

// ValidateConfig validates git-specific configuration.
func (g *GitType) ValidateConfig(config schemas.RefConfig) error {
	// Validate branch name if specified
	if config.Branch != "" {
		// Basic validation: no spaces, no special chars that would break git
		if len(config.Branch) == 0 {
			return fmt.Errorf("branch cannot be empty string")
		}

		// TODO: More comprehensive branch name validation
		// Git branch names can't contain: .., @{, \, ^, ~, :, ?, *, [, spaces at start/end
	}

	// Validate path if specified
	if config.Path != "" {
		// Ensure no path traversal attempts
		cleaned := filepath.Clean(config.Path)
		if filepath.IsAbs(cleaned) {
			return fmt.Errorf("path must be relative, not absolute: %s", config.Path)
		}

		// Check for .. traversal (cleaned path shouldn't start with .. or contain ../)
		if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "..\\") {
			return fmt.Errorf("path contains invalid traversal: %s", config.Path)
		}
	}

	return nil
}
