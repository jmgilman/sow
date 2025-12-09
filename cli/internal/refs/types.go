// Package refs provides the reference type system for managing external resources.
package refs

import (
	"context"

	"github.com/jmgilman/sow/libs/schemas"
)

// RefType defines the interface that all reference types must implement.
//
// Each type (git, file, web, etc.) provides its own implementation of how to:
// - Check system requirements (IsEnabled)
// - Initialize type-specific resources (Init)
// - Cache content locally (Cache)
// - Update cached content (Update)
// - Check staleness (IsStale)
// - Clean up cache (Cleanup)
// - Validate configuration (ValidateConfig)
//
// All types enforce local filesystem caching to ensure AI agents can
// always access content via standard file operations.
type RefType interface {
	// Name returns the type name (e.g., "git", "file").
	Name() string

	// IsEnabled checks if this type is available on the local system.
	// Returns false if required tools or dependencies are missing.
	//
	// Example: git type checks for git binary in PATH.
	IsEnabled(ctx context.Context) (bool, error)

	// Init performs one-time initialization for the type.
	// Called during `sow refs init` for enabled types only.
	//
	// Example: git type creates cache directory structure.
	Init(ctx context.Context, cacheDir string) error

	// Cache fetches and caches a reference to the local filesystem.
	// Returns the absolute cache path where content is stored.
	//
	// The cache path should be: {cacheDir}/{type}/{ref.Id}/
	//
	// Example: git type clones repo to cache directory.
	Cache(ctx context.Context, cacheDir string, ref *schemas.Ref) (string, error)

	// Update updates an existing cached reference.
	// Called when user requests refresh from remote source.
	//
	// Example: git type performs git pull in cache directory.
	Update(ctx context.Context, cacheDir string, ref *schemas.Ref, cached *schemas.CachedRef) error

	// IsStale checks if the cached version is outdated compared to source.
	// Returns true if cache is behind, false if current.
	//
	// Example: git type fetches from remote and compares SHAs.
	IsStale(ctx context.Context, cacheDir string, ref *schemas.Ref, cached *schemas.CachedRef) (bool, error)

	// CachePath returns the expected filesystem path for a cached ref.
	// This is the path where Cache() will/did store content.
	//
	// Path format: {cacheDir}/{type}/{ref.Id}/
	CachePath(cacheDir string, ref *schemas.Ref) string

	// Cleanup removes cached content for a specific ref.
	// Called when ref is removed and cache should be pruned.
	//
	// Example: git type deletes the cache directory.
	Cleanup(ctx context.Context, cacheDir string, ref *schemas.Ref) error

	// ValidateConfig validates type-specific config before adding ref.
	// Returns descriptive errors for invalid configurations.
	//
	// The config parameter is the parsed ref.Config field.
	// Each type validates its own config fields.
	//
	// Example: git type validates branch name format, path validity.
	ValidateConfig(config schemas.RefConfig) error
}

// RefTypeInfo provides metadata about a reference type.
type RefTypeInfo struct {
	// Name is the type identifier (e.g., "git", "file")
	Name string

	// Description is a human-readable description
	Description string

	// URLSchemes are the URL schemes this type handles
	// Example: git handles "git+https", "git+ssh"
	URLSchemes []string

	// RequiresTools lists external tools required by this type
	// Example: git requires ["git"]
	RequiresTools []string

	// Enabled indicates if this type is currently enabled
	Enabled bool
}
