package refs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/schemas"
)

// Manager is the high-level manager for refs operations.
// It uses sow.Context for filesystem access and manages both committed and local indexes.
type Manager struct {
	ctx *sow.Context
}

// NewManager creates a new refs manager using the given context.
func NewManager(ctx *sow.Context) *Manager {
	return &Manager{ctx: ctx}
}

// Add adds a new reference to the repository.
// The reference type is automatically inferred from the URL scheme.
// Returns the created Ref for further operations.
func (m *Manager) Add(ctx context.Context, url string, opts ...RefOption) (*Ref, error) {
	// Apply options
	cfg := &refConfig{
		semantic: "knowledge", // default
		local:    false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate required fields
	if cfg.link == "" {
		return nil, fmt.Errorf("ref link is required (use WithRefLink option)")
	}
	if cfg.description == "" {
		return nil, fmt.Errorf("ref description is required (use WithRefDescription option)")
	}

	// Validate semantic type
	if cfg.semantic != "knowledge" && cfg.semantic != "code" {
		return nil, fmt.Errorf("semantic must be 'knowledge' or 'code', got: %s", cfg.semantic)
	}

	// Infer type from URL
	typeName, err := InferTypeFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to infer type from URL: %w", err)
	}

	// Get type implementation
	refType, err := GetType(typeName)
	if err != nil {
		return nil, fmt.Errorf("unknown reference type: %s", typeName)
	}

	// Check if type is enabled
	enabled, err := refType.IsEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if type enabled: %w", err)
	}
	if !enabled {
		return nil, fmt.Errorf("reference type %s is not enabled (required tools missing)", typeName)
	}

	// Normalize URL for type
	normalizedURL, isLocal, err := m.normalizeURLForType(url, typeName, cfg)
	if err != nil {
		return nil, err
	}
	cfg.local = isLocal

	// Generate ID if not specified
	if cfg.id == "" {
		cfg.id = m.generateRefID(normalizedURL, typeName)
	}

	// Create ref structure
	ref := &schemas.Ref{
		Id:          cfg.id,
		Source:      normalizedURL,
		Semantic:    cfg.semantic,
		Link:        cfg.link,
		Tags:        cfg.tags,
		Description: cfg.description,
		Config: schemas.RefConfig{
			Branch: cfg.branch,
			Path:   cfg.path,
		},
	}

	// Validate config for this type
	if err := refType.ValidateConfig(ref.Config); err != nil {
		return nil, fmt.Errorf("invalid config for type %s: %w", typeName, err)
	}

	// Load appropriate index
	index, isLocal, err := m.loadRefIndex(cfg.local)
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	// Check for duplicate ID
	for _, existingRef := range index.Refs {
		if existingRef.Id == ref.Id {
			indexType := "committed"
			if isLocal {
				indexType = "local"
			}
			return nil, fmt.Errorf("ref with ID %q already exists in %s index", ref.Id, indexType)
		}
	}

	// Create cache manager and install ref
	sowDir := filepath.Join(m.ctx.RepoRoot(), ".sow")
	cacheManager, err := NewCacheManager(sowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create refs cache manager: %w", err)
	}

	_, err = cacheManager.Install(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to install ref: %w", err)
	}

	// Add to index and save
	index.Refs = append(index.Refs, *ref)
	if err := m.saveRefIndex(index, isLocal); err != nil {
		return nil, fmt.Errorf("failed to save index: %w", err)
	}

	return &Ref{
		manager: m,
		id:      cfg.id,
	}, nil
}

// Get retrieves a ref by ID from either committed or local index.
func (m *Manager) Get(id string) (*Ref, error) {
	// Try committed index first
	committedIndex, err := m.loadCommittedRefIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load committed index: %w", err)
	}

	for _, ref := range committedIndex.Refs {
		if ref.Id == id {
			return &Ref{manager: m, id: id}, nil
		}
	}

	// Try local index
	localIndex, err := m.loadLocalRefIndex()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load local index: %w", err)
	}

	if localIndex != nil {
		for _, ref := range localIndex.Refs {
			if ref.Id == id {
				return &Ref{manager: m, id: id}, nil
			}
		}
	}

	return nil, fmt.Errorf("ref %q not found", id)
}

// List returns all refs matching the given filters.
func (m *Manager) List(opts ...RefListOption) ([]*Ref, error) {
	// Apply options
	cfg := &refListConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var allRefs []schemas.Ref

	// Load committed refs unless localOnly
	if !cfg.localOnly {
		committedIndex, err := m.loadCommittedRefIndex()
		if err != nil {
			return nil, fmt.Errorf("failed to load committed index: %w", err)
		}
		allRefs = append(allRefs, committedIndex.Refs...)
	}

	// Load local refs unless committedOnly
	if !cfg.committedOnly {
		localIndex, err := m.loadLocalRefIndex()
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load local index: %w", err)
		}
		if localIndex != nil {
			allRefs = append(allRefs, localIndex.Refs...)
		}
	}

	// Filter refs
	var filtered []*Ref
	for _, ref := range allRefs {
		if m.matchesRefFilters(ref, cfg) {
			filtered = append(filtered, &Ref{manager: m, id: ref.Id})
		}
	}

	return filtered, nil
}

// Remove removes a ref by ID.
func (m *Manager) Remove(ctx context.Context, id string, pruneCache bool) error {
	// Find the ref in either index
	ref, isLocal, err := m.findRefInIndexes(id)
	if err != nil {
		return err
	}

	// Create cache manager
	sowDir := filepath.Join(m.ctx.RepoRoot(), ".sow")
	cacheManager, err := NewCacheManager(sowDir)
	if err != nil {
		return fmt.Errorf("failed to create refs cache manager: %w", err)
	}

	// Remove via cache manager (handles cache cleanup if pruneCache is true)
	if pruneCache {
		if err := cacheManager.Remove(ctx, ref); err != nil {
			return fmt.Errorf("failed to remove ref: %w", err)
		}
	} else {
		// Just remove the symlink, keep cache
		workspacePath := filepath.Join(m.ctx.RepoRoot(), ".sow", "refs", ref.Link)
		if err := os.Remove(workspacePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove workspace symlink: %w", err)
		}
	}

	// Remove from index
	index, _, err := m.loadRefIndex(isLocal)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Filter out the removed ref
	var newRefs []schemas.Ref
	for _, r := range index.Refs {
		if r.Id != id {
			newRefs = append(newRefs, r)
		}
	}
	index.Refs = newRefs

	// Save index
	if err := m.saveRefIndex(index, isLocal); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// InitRefs initializes all refs after cloning a repository.
// This installs refs from the committed index.
func (m *Manager) InitRefs(ctx context.Context) error {
	// Load committed index
	committedIndex, err := m.loadCommittedRefIndex()
	if err != nil {
		return fmt.Errorf("failed to load committed index: %w", err)
	}

	// Create cache manager
	sowDir := filepath.Join(m.ctx.RepoRoot(), ".sow")
	cacheManager, err := NewCacheManager(sowDir)
	if err != nil {
		return fmt.Errorf("failed to create refs cache manager: %w", err)
	}

	// Install each ref
	for _, ref := range committedIndex.Refs {
		// Check if type is enabled
		typeName, err := InferTypeFromURL(ref.Source)
		if err != nil {
			return fmt.Errorf("failed to infer type for ref %s: %w", ref.Id, err)
		}

		refType, err := GetType(typeName)
		if err != nil {
			return fmt.Errorf("unknown type for ref %s: %s", ref.Id, typeName)
		}

		enabled, err := refType.IsEnabled(ctx)
		if err != nil {
			return fmt.Errorf("failed to check if type enabled for ref %s: %w", ref.Id, err)
		}

		if !enabled {
			// Skip disabled types with warning (caller should report this)
			continue
		}

		// Install ref
		if _, err := cacheManager.Install(ctx, &ref); err != nil {
			return fmt.Errorf("failed to install ref %s: %w", ref.Id, err)
		}
	}

	return nil
}

// Helper methods

// normalizeURLForType normalizes a URL based on its type and validates type-specific flags.
func (m *Manager) normalizeURLForType(rawURL, typeName string, cfg *refConfig) (string, bool, error) {
	normalizedURL := rawURL
	local := cfg.local

	switch typeName {
	case "git":
		normalized, _, err := ParseGitURL(rawURL)
		if err != nil {
			return "", local, fmt.Errorf("failed to parse git URL: %w", err)
		}
		normalizedURL = normalized

	case "file":
		// Convert path to file URL if needed
		if len(rawURL) < 7 || rawURL[:7] != "file://" {
			fileURL, err := PathToFileURL(rawURL)
			if err != nil {
				return "", local, fmt.Errorf("failed to convert path to file URL: %w", err)
			}
			normalizedURL = fileURL
		}

		// Validate file URL
		if err := ValidateFileURL(normalizedURL); err != nil {
			return "", local, fmt.Errorf("invalid file URL: %w", err)
		}

		// File refs are always local
		local = true

		// File refs don't support branch/path
		if cfg.branch != "" {
			return "", local, fmt.Errorf("--branch flag only valid for git URLs")
		}
		if cfg.path != "" {
			return "", local, fmt.Errorf("--path flag only valid for git URLs")
		}

	default:
		// For other types, validate they don't use git-specific flags
		if cfg.branch != "" {
			return "", local, fmt.Errorf("--branch flag only valid for git URLs")
		}
		if cfg.path != "" {
			return "", local, fmt.Errorf("--path flag only valid for git URLs")
		}
	}

	return normalizedURL, local, nil
}

// generateRefID generates a ref ID from URL and type.
func (m *Manager) generateRefID(url, typeName string) string {
	// This logic mirrors the original generateIDFromURL in refs/refs.go
	switch typeName {
	case "git":
		// Remove scheme prefix
		for _, prefix := range []string{"git+https://", "git+ssh://", "git+http://", "git@"} {
			if len(url) > len(prefix) && url[:len(prefix)] == prefix {
				url = url[len(prefix):]
				break
			}
		}

		// Remove domain (take last 2 path components)
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			url = parts[len(parts)-2] + "-" + parts[len(parts)-1]
		}

		// Remove .git suffix
		if len(url) > 4 && url[len(url)-4:] == ".git" {
			url = url[:len(url)-4]
		}

	case "file":
		// Get base directory name
		if len(url) > 7 && url[:7] == "file://" {
			url = url[7:]
		}
		if len(url) > 0 && url[len(url)-1] == '/' {
			url = url[:len(url)-1]
		}
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			url = parts[len(parts)-1]
		}
	}

	// Convert to kebab-case
	url = strings.ToLower(url)
	url = strings.ReplaceAll(url, "/", "-")
	url = strings.ReplaceAll(url, "_", "-")
	url = strings.ReplaceAll(url, ":", "-")

	return url
}

// loadRefIndex loads the appropriate index (committed or local).
func (m *Manager) loadRefIndex(isLocal bool) (*schemas.RefsCommittedIndex, bool, error) {
	if isLocal {
		localIndex, err := m.loadLocalRefIndex()
		if err != nil && !os.IsNotExist(err) {
			return nil, true, err
		}
		if localIndex == nil {
			localIndex = &schemas.RefsLocalIndex{
				Version: "1.0.0",
				Refs:    []schemas.Ref{},
			}
		}
		// Convert to committed index structure
		return &schemas.RefsCommittedIndex{
			Version: localIndex.Version,
			Refs:    localIndex.Refs,
		}, true, nil
	}

	committedIndex, err := m.loadCommittedRefIndex()
	return committedIndex, false, err
}

// saveRefIndex saves the index (committed or local).
func (m *Manager) saveRefIndex(index *schemas.RefsCommittedIndex, isLocal bool) error {
	fs := m.ctx.FS()

	if isLocal {
		localIndex := &schemas.RefsLocalIndex{
			Version: "1.0.0",
			Refs:    index.Refs,
		}
		return m.saveLocalRefIndex(fs, localIndex)
	}

	index.Version = "1.0.0"
	return m.saveCommittedRefIndex(fs, index)
}

// loadCommittedRefIndex loads the committed refs index.
func (m *Manager) loadCommittedRefIndex() (*schemas.RefsCommittedIndex, error) {
	fs := m.ctx.FS()
	path := "refs/index.json"

	// Check if file exists
	if _, err := fs.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// Return empty index
			return &schemas.RefsCommittedIndex{
				Version: "1.0.0",
				Refs:    []schemas.Ref{},
			}, nil
		}
		return nil, fmt.Errorf("failed to stat committed refs index: %w", err)
	}

	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read committed refs index: %w", err)
	}

	var index schemas.RefsCommittedIndex
	if err := unmarshalJSON(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// loadLocalRefIndex loads the local refs index.
func (m *Manager) loadLocalRefIndex() (*schemas.RefsLocalIndex, error) {
	fs := m.ctx.FS()
	path := "refs/index.local.json"

	// Check if file exists
	if _, err := fs.Stat(path); err != nil {
		// Return unwrapped so callers can check os.IsNotExist()
		return nil, err //nolint:wrapcheck // Intentionally unwrapped for os.IsNotExist() checks
	}

	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read local refs index: %w", err)
	}

	var index schemas.RefsLocalIndex
	if err := unmarshalJSON(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// saveCommittedRefIndex saves the committed refs index.
func (m *Manager) saveCommittedRefIndex(fs sow.FS, index *schemas.RefsCommittedIndex) error {
	path := "refs/index.json"
	data, err := marshalJSON(index)
	if err != nil {
		return err
	}
	if err := fs.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write committed refs index: %w", err)
	}
	return nil
}

// saveLocalRefIndex saves the local refs index.
func (m *Manager) saveLocalRefIndex(fs sow.FS, index *schemas.RefsLocalIndex) error {
	path := "refs/index.local.json"
	data, err := marshalJSON(index)
	if err != nil {
		return err
	}
	if err := fs.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write local refs index: %w", err)
	}
	return nil
}

// findRefInIndexes finds a ref by ID in either index.
// Returns the ref, whether it's local, and any error.
func (m *Manager) findRefInIndexes(id string) (*schemas.Ref, bool, error) {
	// Try committed index first
	committedIndex, err := m.loadCommittedRefIndex()
	if err != nil {
		return nil, false, fmt.Errorf("failed to load committed index: %w", err)
	}

	for _, ref := range committedIndex.Refs {
		if ref.Id == id {
			return &ref, false, nil
		}
	}

	// Try local index
	localIndex, err := m.loadLocalRefIndex()
	if err != nil && !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("failed to load local index: %w", err)
	}

	if localIndex != nil {
		for _, ref := range localIndex.Refs {
			if ref.Id == id {
				return &ref, true, nil
			}
		}
	}

	return nil, false, fmt.Errorf("ref %q not found", id)
}

// matchesRefFilters checks if a ref matches the given filters.
func (m *Manager) matchesRefFilters(ref schemas.Ref, cfg *refListConfig) bool {
	// Filter by type
	if cfg.typeFilter != "" {
		refType, err := InferTypeFromURL(ref.Source)
		if err != nil || refType != cfg.typeFilter {
			return false
		}
	}

	// Filter by semantic
	if cfg.semanticFilter != "" && ref.Semantic != cfg.semanticFilter {
		return false
	}

	// Filter by tags
	if len(cfg.tagsFilter) > 0 {
		for _, filterTag := range cfg.tagsFilter {
			found := false
			for _, refTag := range ref.Tags {
				if refTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}
