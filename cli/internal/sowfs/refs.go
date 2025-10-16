package sowfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	gobilly "github.com/go-git/go-billy/v5"
	"github.com/jmgilman/sow/cli/schemas"
)

// RefsFS provides access to the .sow/refs/ directory and index files.
//
// This domain handles external knowledge and code references including:
//   - index.json - Committed refs index (team-shared)
//   - index.local.json - Local refs index (git-ignored)
//   - Symlinks to cached repositories
//
// All paths are relative to .sow/refs/.
type RefsFS interface {
	// CommittedIndex reads and validates the committed refs index.
	// Returns the parsed and validated RefsCommittedIndex struct.
	// File: .sow/refs/index.json
	CommittedIndex() (*schemas.RefsCommittedIndex, error)

	// LocalIndex reads and validates the local refs index.
	// Returns the parsed and validated RefsLocalIndex struct.
	// File: .sow/refs/index.local.json
	// Returns empty index if file doesn't exist (local index is optional)
	LocalIndex() (*schemas.RefsLocalIndex, error)

	// WriteCommittedIndex validates and writes the committed refs index.
	// Validates against CUE schema before writing.
	// File: .sow/refs/index.json
	WriteCommittedIndex(index *schemas.RefsCommittedIndex) error

	// WriteLocalIndex validates and writes the local refs index.
	// Validates against CUE schema before writing.
	// File: .sow/refs/index.local.json
	WriteLocalIndex(index *schemas.RefsLocalIndex) error

	// SymlinkExists checks if a symlink exists in .sow/refs/
	// Name is just the symlink name (e.g., "go-style-guide")
	SymlinkExists(name string) (bool, error)

	// CreateSymlink creates a symlink in .sow/refs/
	// Target is the absolute path to the cached repository or directory
	// Name is the symlink name (e.g., "go-style-guide")
	CreateSymlink(target, name string) error

	// RemoveSymlink removes a symlink from .sow/refs/
	// Name is the symlink name (e.g., "go-style-guide")
	RemoveSymlink(name string) error

	// ListSymlinks returns all symlink names in .sow/refs/
	// Returns just the names, not full paths
	ListSymlinks() ([]string, error)

	// ReadSymlink reads the target of a symlink in .sow/refs/
	// Name is the symlink name (e.g., "go-style-guide")
	// Returns the absolute path the symlink points to
	ReadSymlink(name string) (string, error)
}

// RefsFSImpl is the concrete implementation of RefsFS.
type RefsFSImpl struct {
	// sowFS is the parent SowFS (for accessing chrooted .sow filesystem)
	sowFS *SowFSImpl

	// validator provides CUE schema validation
	validator *schemas.CUEValidator
}

// Ensure RefsFSImpl implements RefsFS.
var _ RefsFS = (*RefsFSImpl)(nil)

// NewRefsFS creates a new RefsFS instance.
func NewRefsFS(sowFS *SowFSImpl, validator *schemas.CUEValidator) *RefsFSImpl {
	return &RefsFSImpl{
		sowFS:     sowFS,
		validator: validator,
	}
}

// billyFS returns the underlying billy filesystem for symlink operations.
func (r *RefsFSImpl) billyFS() (gobilly.Filesystem, error) {
	// Try to unwrap if it's a billy wrapper type
	type unwrapper interface {
		Unwrap() gobilly.Filesystem
	}

	if u, ok := r.sowFS.fs.(unwrapper); ok {
		return u.Unwrap(), nil
	}

	return nil, fmt.Errorf("filesystem does not support symlink operations")
}

// CommittedIndex reads the committed refs index.
func (r *RefsFSImpl) CommittedIndex() (*schemas.RefsCommittedIndex, error) {
	indexPath := "refs/index.json"

	// Read index file
	data, err := r.sowFS.fs.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index if file doesn't exist
			return &schemas.RefsCommittedIndex{
				Refs: []schemas.Ref{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read committed index: %w", err)
	}

	// Validate against CUE schema
	if err := r.validator.ValidateRefsCommittedIndex(data); err != nil {
		return nil, fmt.Errorf("committed index validation failed: %w", err)
	}

	// Parse JSON
	var index schemas.RefsCommittedIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse committed index JSON: %w", err)
	}

	return &index, nil
}

// LocalIndex reads the local refs index.
func (r *RefsFSImpl) LocalIndex() (*schemas.RefsLocalIndex, error) {
	indexPath := "refs/index.local.json"

	// Check if file exists
	exists, err := r.sowFS.fs.Exists(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check local index: %w", err)
	}

	// Return empty index if file doesn't exist (local index is optional)
	if !exists {
		return &schemas.RefsLocalIndex{
			Refs: []schemas.Ref{},
		}, nil
	}

	// Read index file
	data, err := r.sowFS.fs.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local index: %w", err)
	}

	// Validate against CUE schema
	if err := r.validator.ValidateRefsLocalIndex(data); err != nil {
		return nil, fmt.Errorf("local index validation failed: %w", err)
	}

	// Parse JSON
	var index schemas.RefsLocalIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse local index JSON: %w", err)
	}

	return &index, nil
}

// WriteCommittedIndex writes the committed refs index.
func (r *RefsFSImpl) WriteCommittedIndex(index *schemas.RefsCommittedIndex) error {
	indexPath := "refs/index.json"

	// Encode to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal committed index: %w", err)
	}

	// Validate against CUE schema
	if err := r.validator.ValidateRefsCommittedIndex(data); err != nil {
		return fmt.Errorf("committed index validation failed: %w", err)
	}

	// Ensure refs directory exists
	if err := r.sowFS.fs.MkdirAll("refs", 0755); err != nil {
		return fmt.Errorf("failed to create refs directory: %w", err)
	}

	// Write to file
	if err := r.sowFS.fs.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write committed index: %w", err)
	}

	return nil
}

// WriteLocalIndex writes the local refs index.
func (r *RefsFSImpl) WriteLocalIndex(index *schemas.RefsLocalIndex) error {
	indexPath := "refs/index.local.json"

	// Encode to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal local index: %w", err)
	}

	// Validate against CUE schema
	if err := r.validator.ValidateRefsLocalIndex(data); err != nil {
		return fmt.Errorf("local index validation failed: %w", err)
	}

	// Ensure refs directory exists
	if err := r.sowFS.fs.MkdirAll("refs", 0755); err != nil {
		return fmt.Errorf("failed to create refs directory: %w", err)
	}

	// Write to file
	if err := r.sowFS.fs.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write local index: %w", err)
	}

	return nil
}

// SymlinkExists checks if a symlink exists.
func (r *RefsFSImpl) SymlinkExists(name string) (bool, error) {
	// Sanitize name to prevent path traversal
	name = filepath.Base(name)
	symlinkPath := filepath.Join("refs", name)

	// Check if path exists
	exists, err := r.sowFS.fs.Exists(symlinkPath)
	if err != nil {
		return false, fmt.Errorf("failed to check symlink: %w", err)
	}
	if !exists {
		return false, nil
	}

	// Get billy filesystem for symlink operations
	billyFS, err := r.billyFS()
	if err != nil {
		return false, err
	}

	// Check if it's a symlink
	info, err := billyFS.Lstat(symlinkPath)
	if err != nil {
		return false, fmt.Errorf("failed to stat symlink: %w", err)
	}

	return info.Mode()&os.ModeSymlink != 0, nil
}

// CreateSymlink creates a symlink.
func (r *RefsFSImpl) CreateSymlink(target, name string) error {
	// Sanitize name to prevent path traversal
	name = filepath.Base(name)
	symlinkPath := filepath.Join("refs", name)

	// Ensure refs directory exists
	if err := r.sowFS.fs.MkdirAll("refs", 0755); err != nil {
		return fmt.Errorf("failed to create refs directory: %w", err)
	}

	// Get billy filesystem for symlink operations
	billyFS, err := r.billyFS()
	if err != nil {
		return err
	}

	// Create symlink
	if err := billyFS.Symlink(target, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// RemoveSymlink removes a symlink.
func (r *RefsFSImpl) RemoveSymlink(name string) error {
	// Sanitize name to prevent path traversal
	name = filepath.Base(name)
	symlinkPath := filepath.Join("refs", name)

	// Remove symlink
	if err := r.sowFS.fs.Remove(symlinkPath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	return nil
}

// ListSymlinks lists all symlinks.
func (r *RefsFSImpl) ListSymlinks() ([]string, error) {
	refsPath := "refs"

	// Check if refs directory exists
	exists, err := r.sowFS.fs.Exists(refsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check refs directory: %w", err)
	}
	if !exists {
		// No refs directory means no symlinks
		return []string{}, nil
	}

	// Read directory entries
	entries, err := r.sowFS.fs.ReadDir(refsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read refs directory: %w", err)
	}

	// Get billy filesystem for symlink operations
	billyFS, err := r.billyFS()
	if err != nil {
		return nil, err
	}

	// Filter for symlinks only
	var symlinks []string
	for _, entry := range entries {
		// Skip index files
		if entry.Name() == "index.json" || entry.Name() == "index.local.json" {
			continue
		}

		// Check if it's a symlink
		entryPath := filepath.Join(refsPath, entry.Name())
		info, err := billyFS.Lstat(entryPath)
		if err != nil {
			continue // Skip entries we can't stat
		}

		if info.Mode()&os.ModeSymlink != 0 {
			symlinks = append(symlinks, entry.Name())
		}
	}

	return symlinks, nil
}

// ReadSymlink reads a symlink target.
func (r *RefsFSImpl) ReadSymlink(name string) (string, error) {
	// Sanitize name to prevent path traversal
	name = filepath.Base(name)
	symlinkPath := filepath.Join("refs", name)

	// Get billy filesystem for symlink operations
	billyFS, err := r.billyFS()
	if err != nil {
		return "", err
	}

	// Read symlink target
	target, err := billyFS.Readlink(symlinkPath)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink: %w", err)
	}

	return target, nil
}
